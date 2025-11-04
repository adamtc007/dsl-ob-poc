package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"

	"dsl-ob-poc/internal/datastore"
	hfagent "dsl-ob-poc/internal/hf-agent"
)

// Server represents the web server for DSL agent chat
type Server struct {
	router   *mux.Router
	agent    *hfagent.HedgeFundDSLAgent
	store    datastore.DataStore
	sessions map[string]*ChatSession
	mu       sync.RWMutex
	upgrader websocket.Upgrader
}

// ChatSession represents a user's chat session with context
type ChatSession struct {
	SessionID string
	Context   hfagent.DSLGenerationRequest
	History   []ChatMessage
	CreatedAt time.Time
	LastUsed  time.Time
}

// ChatMessage represents a single message in the chat
type ChatMessage struct {
	Role      string                         `json:"role"` // "user" or "agent"
	Content   string                         `json:"content"`
	DSL       string                         `json:"dsl,omitempty"`
	Response  *hfagent.DSLGenerationResponse `json:"response,omitempty"`
	Timestamp time.Time                      `json:"timestamp"`
}

// WebSocketMessage represents messages sent over WebSocket
type WebSocketMessage struct {
	Type    string          `json:"type"` // "chat", "dsl_generate", "dsl_validate", "dsl_execute"
	Payload json.RawMessage `json:"payload"`
}

// ChatRequest represents a chat message from the client
type ChatRequest struct {
	SessionID string                 `json:"session_id,omitempty"`
	Message   string                 `json:"message"`
	Context   map[string]interface{} `json:"context,omitempty"`
}

// ChatResponse represents a response to the client
type ChatResponse struct {
	SessionID string                         `json:"session_id"`
	Message   string                         `json:"message"`
	DSL       string                         `json:"dsl,omitempty"`
	Response  *hfagent.DSLGenerationResponse `json:"response,omitempty"`
	Error     string                         `json:"error,omitempty"`
}

// NewServer creates a new web server
func NewServer(agent *hfagent.HedgeFundDSLAgent, store datastore.DataStore) *Server {
	s := &Server{
		router:   mux.NewRouter(),
		agent:    agent,
		store:    store,
		sessions: make(map[string]*ChatSession),
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				return true // Allow all origins in development
			},
		},
	}

	s.setupRoutes()
	return s
}

// setupRoutes configures all HTTP routes
func (s *Server) setupRoutes() {
	// Static files (will serve the React app)
	s.router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./web/static"))))

	// API routes
	api := s.router.PathPrefix("/api").Subrouter()
	api.HandleFunc("/health", s.handleHealth).Methods("GET")
	api.HandleFunc("/chat", s.handleChat).Methods("POST")
	api.HandleFunc("/dsl/generate", s.handleGenerateDSL).Methods("POST")
	api.HandleFunc("/dsl/validate", s.handleValidateDSL).Methods("POST")
	api.HandleFunc("/dsl/execute", s.handleExecuteDSL).Methods("POST")
	api.HandleFunc("/session/{id}", s.handleGetSession).Methods("GET")
	api.HandleFunc("/session/{id}/history", s.handleGetHistory).Methods("GET")
	api.HandleFunc("/vocabulary", s.handleGetVocabulary).Methods("GET")
	api.HandleFunc("/attributes", s.handleGetAttributes).Methods("GET")

	// WebSocket endpoint
	s.router.HandleFunc("/ws", s.handleWebSocket)

	// Serve index.html for all other routes (SPA)
	s.router.PathPrefix("/").HandlerFunc(s.handleIndex)
}

// handleHealth returns server health status
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "healthy",
		"service": "hedge-fund-dsl-agent",
		"time":    time.Now().UTC(),
	})
}

// handleChat processes a chat message and generates DSL
func (s *Server) handleChat(w http.ResponseWriter, r *http.Request) {
	var req ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Get or create session
	session := s.getOrCreateSession(req.SessionID)

	// Add user message to history
	userMsg := ChatMessage{
		Role:      "user",
		Content:   req.Message,
		Timestamp: time.Now().UTC(),
	}
	session.History = append(session.History, userMsg)

	// Update context with any provided context
	if req.Context != nil {
		if investorID, ok := req.Context["investor_id"].(string); ok {
			session.Context.InvestorID = investorID
		}
		if state, ok := req.Context["current_state"].(string); ok {
			session.Context.CurrentState = state
		}
	}

	// Generate DSL using the agent
	session.Context.Instruction = req.Message
	ctx := r.Context()

	response, err := s.agent.GenerateDSL(ctx, session.Context)
	if err != nil {
		s.respondError(w, fmt.Sprintf("Failed to generate DSL: %v", err), http.StatusInternalServerError)
		return
	}

	// Update session context with state transition
	session.Context.CurrentState = response.ToState
	if response.Parameters["investor"] != nil {
		if investorID, ok := response.Parameters["investor"].(string); ok {
			session.Context.InvestorID = investorID
		}
	}

	// Add agent response to history
	agentMsg := ChatMessage{
		Role:      "agent",
		Content:   response.Explanation,
		DSL:       response.DSL,
		Response:  response,
		Timestamp: time.Now().UTC(),
	}
	session.History = append(session.History, agentMsg)
	session.LastUsed = time.Now().UTC()

	// Send response
	chatResp := ChatResponse{
		SessionID: session.SessionID,
		Message:   response.Explanation,
		DSL:       response.DSL,
		Response:  response,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(chatResp)
}

// handleGenerateDSL generates DSL without chat context
func (s *Server) handleGenerateDSL(w http.ResponseWriter, r *http.Request) {
	var req hfagent.DSLGenerationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	response, err := s.agent.GenerateDSL(r.Context(), req)
	if err != nil {
		s.respondError(w, fmt.Sprintf("DSL generation failed: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleValidateDSL validates a DSL string
func (s *Server) handleValidateDSL(w http.ResponseWriter, r *http.Request) {
	var req struct {
		DSL string `json:"dsl"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// TODO: Implement actual DSL validation with parser
	// For now, basic validation
	valid := len(req.DSL) > 0

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"valid":  valid,
		"errors": []string{},
	})
}

// handleExecuteDSL executes a validated DSL operation
func (s *Server) handleExecuteDSL(w http.ResponseWriter, r *http.Request) {
	var req struct {
		DSL       string `json:"dsl"`
		SessionID string `json:"session_id,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// TODO: Implement actual DSL execution
	// For now, mock response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "success",
		"message": "DSL execution pending implementation",
	})
}

// handleGetSession retrieves session information
func (s *Server) handleGetSession(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sessionID := vars["id"]

	s.mu.RLock()
	session, exists := s.sessions[sessionID]
	s.mu.RUnlock()

	if !exists {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(session)
}

// handleGetHistory retrieves chat history for a session
func (s *Server) handleGetHistory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sessionID := vars["id"]

	s.mu.RLock()
	session, exists := s.sessions[sessionID]
	s.mu.RUnlock()

	if !exists {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(session.History)
}

// handleGetVocabulary returns the DSL vocabulary
func (s *Server) handleGetVocabulary(w http.ResponseWriter, r *http.Request) {
	vocab := hfagent.GetHedgeFundDSLVocabulary()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(vocab)
}

// handleGetAttributes returns available attributes from dictionary
func (s *Server) handleGetAttributes(w http.ResponseWriter, r *http.Request) {
	// TODO: Query dictionary table for attributes
	// For now, mock response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"attributes": []map[string]string{
			{"id": "uuid-0001", "name": "hf.investor.legal-name", "type": "string"},
			{"id": "uuid-0002", "name": "hf.investor.type", "type": "enum"},
			{"id": "uuid-0003", "name": "hf.investor.domicile", "type": "country-code"},
		},
	})
}

// handleWebSocket handles WebSocket connections for real-time chat
func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}
	defer conn.Close()

	sessionID := uuid.New().String()
	session := s.getOrCreateSession(sessionID)

	log.Printf("WebSocket connection established: %s", sessionID)

	// Send welcome message
	welcomeMsg := map[string]interface{}{
		"type": "welcome",
		"payload": map[string]string{
			"session_id": sessionID,
			"message":    "Connected to Hedge Fund DSL Agent. How can I help you?",
		},
	}
	conn.WriteJSON(welcomeMsg)

	// Handle messages
	for {
		var msg WebSocketMessage
		err := conn.ReadJSON(&msg)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		s.handleWebSocketMessage(conn, session, msg)
	}

	log.Printf("WebSocket connection closed: %s", sessionID)
}

// handleWebSocketMessage processes a WebSocket message
func (s *Server) handleWebSocketMessage(conn *websocket.Conn, session *ChatSession, msg WebSocketMessage) {
	switch msg.Type {
	case "chat":
		var chatReq ChatRequest
		if err := json.Unmarshal(msg.Payload, &chatReq); err != nil {
			s.sendWSError(conn, "Invalid chat request")
			return
		}

		// Generate DSL
		session.Context.Instruction = chatReq.Message
		ctx := context.Background()

		response, err := s.agent.GenerateDSL(ctx, session.Context)
		if err != nil {
			s.sendWSError(conn, fmt.Sprintf("Failed to generate DSL: %v", err))
			return
		}

		// Update session
		session.Context.CurrentState = response.ToState
		session.LastUsed = time.Now().UTC()

		// Send response
		conn.WriteJSON(map[string]interface{}{
			"type": "chat_response",
			"payload": map[string]interface{}{
				"message":  response.Explanation,
				"dsl":      response.DSL,
				"verb":     response.Verb,
				"state":    response.ToState,
				"response": response,
			},
		})

	case "ping":
		conn.WriteJSON(map[string]interface{}{
			"type":    "pong",
			"payload": map[string]string{"time": time.Now().UTC().Format(time.RFC3339)},
		})

	default:
		s.sendWSError(conn, fmt.Sprintf("Unknown message type: %s", msg.Type))
	}
}

// sendWSError sends an error message over WebSocket
func (s *Server) sendWSError(conn *websocket.Conn, message string) {
	conn.WriteJSON(map[string]interface{}{
		"type":    "error",
		"payload": map[string]string{"error": message},
	})
}

// handleIndex serves the React app
func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./web/static/index.html")
}

// getOrCreateSession gets existing session or creates new one
func (s *Server) getOrCreateSession(sessionID string) *ChatSession {
	s.mu.Lock()
	defer s.mu.Unlock()

	if sessionID == "" {
		sessionID = uuid.New().String()
	}

	session, exists := s.sessions[sessionID]
	if !exists {
		session = &ChatSession{
			SessionID: sessionID,
			Context:   hfagent.DSLGenerationRequest{},
			History:   []ChatMessage{},
			CreatedAt: time.Now().UTC(),
			LastUsed:  time.Now().UTC(),
		}
		s.sessions[sessionID] = session
	}

	return session
}

// respondError sends an error response
func (s *Server) respondError(w http.ResponseWriter, message string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}

// Start starts the web server
func (s *Server) Start(addr string) error {
	log.Printf("Starting Hedge Fund DSL Agent Web Server on %s", addr)
	return http.ListenAndServe(addr, s.router)
}

// CleanupSessions removes inactive sessions
func (s *Server) CleanupSessions(maxAge time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now().UTC()
	for id, session := range s.sessions {
		if now.Sub(session.LastUsed) > maxAge {
			delete(s.sessions, id)
			log.Printf("Cleaned up inactive session: %s", id)
		}
	}
}

// StartCleanupRoutine starts a goroutine to periodically clean up sessions
func (s *Server) StartCleanupRoutine(interval, maxAge time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		for range ticker.C {
			s.CleanupSessions(maxAge)
		}
	}()
}

func main() {
	// Get API key
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		apiKey = os.Getenv("GOOGLE_API_KEY")
	}
	if apiKey == "" {
		log.Fatal("GEMINI_API_KEY or GOOGLE_API_KEY environment variable required")
	}

	// Initialize agent
	ctx := context.Background()
	agent, err := hfagent.NewHedgeFundDSLAgent(ctx, apiKey)
	if err != nil {
		log.Fatalf("Failed to initialize DSL agent: %v", err)
	}
	defer agent.Close()

	// Initialize datastore (can be nil for now)
	var store datastore.DataStore

	// Create server
	server := NewServer(agent, store)

	// Start session cleanup routine (clean up after 1 hour of inactivity)
	server.StartCleanupRoutine(15*time.Minute, 1*time.Hour)

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	addr := fmt.Sprintf(":%s", port)
	log.Fatal(server.Start(addr))
}
