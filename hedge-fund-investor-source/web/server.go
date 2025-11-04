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

	"dsl-ob-poc/hedge-fund-investor-source/web/internal/datastore"
	registry "dsl-ob-poc/internal/domain-registry"
	hedgefundinvestor "dsl-ob-poc/internal/domains/hedge-fund-investor"
	"dsl-ob-poc/internal/domains/onboarding"
	"dsl-ob-poc/internal/shared-dsl/session"
)

// Server represents the multi-domain web server for DSL agent chat
type Server struct {
	router       *mux.Router
	registry     *registry.Registry
	domainRouter *registry.Router
	sessionMgr   *session.Manager
	dictionary   interface{} // TODO: Use proper dictionary service
	store        datastore.DataStore
	sessions     map[string]*ChatSession
	mu           sync.RWMutex
	upgrader     websocket.Upgrader
}

// ChatSession represents a user's chat session with multi-domain support
type ChatSession struct {
	SessionID     string
	CurrentDomain string
	Context       map[string]interface{}
	BuiltDSL      string // Accumulated DSL throughout conversation
	History       []ChatMessage
	CreatedAt     time.Time
	LastUsed      time.Time
}

// ChatMessage represents a single message in the chat
type ChatMessage struct {
	Role      string                       `json:"role"` // "user" or "agent"
	Content   string                       `json:"content"`
	DSL       string                       `json:"dsl,omitempty"`
	Fragment  string                       `json:"fragment,omitempty"`
	Domain    string                       `json:"domain,omitempty"`
	Response  *registry.GenerationResponse `json:"response,omitempty"`
	Timestamp time.Time                    `json:"timestamp"`
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
	Domain    string                 `json:"domain,omitempty"`
	Context   map[string]interface{} `json:"context,omitempty"`
}

// ChatResponse represents a response to the client
type ChatResponse struct {
	SessionID string                       `json:"session_id"`
	Message   string                       `json:"message"`
	DSL       string                       `json:"dsl,omitempty"`
	Fragment  string                       `json:"fragment,omitempty"`
	Domain    string                       `json:"domain,omitempty"`
	Response  *registry.GenerationResponse `json:"response,omitempty"`
	Error     string                       `json:"error,omitempty"`
}

// DomainInfo represents domain information for the client
type DomainInfo struct {
	Name        string         `json:"name"`
	Version     string         `json:"version"`
	Description string         `json:"description"`
	IsHealthy   bool           `json:"is_healthy"`
	VerbCount   int            `json:"verb_count"`
	Categories  map[string]int `json:"categories"`
}

// NewServer creates a new multi-domain web server
func NewServer(dictService interface{}, store datastore.DataStore, apiKey string) (*Server, error) {
	// Create domain registry
	reg := registry.NewRegistry()

	// Register hedge fund domain
	hfDomain := hedgefundinvestor.NewDomain()
	if err := reg.Register(hfDomain); err != nil {
		return nil, fmt.Errorf("failed to register hedge fund domain: %w", err)
	}

	// Register onboarding domain
	obDomain := onboarding.NewDomain()
	if err := reg.Register(obDomain); err != nil {
		return nil, fmt.Errorf("failed to register onboarding domain: %w", err)
	}

	s := &Server{
		router:       mux.NewRouter(),
		registry:     reg,
		domainRouter: registry.NewRouter(reg),
		sessionMgr:   session.NewManager(),
		dictionary:   dictService,
		store:        store,
		sessions:     make(map[string]*ChatSession),
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				return true // Allow all origins in development
			},
		},
	}

	s.setupRoutes()
	return s, nil
}

// setupRoutes configures all HTTP routes
func (s *Server) setupRoutes() {
	// Static files (will serve the React app)
	s.router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))

	// API routes
	api := s.router.PathPrefix("/api").Subrouter()
	api.HandleFunc("/health", s.handleHealth).Methods("GET")
	api.HandleFunc("/chat", s.handleChat).Methods("POST")
	api.HandleFunc("/dsl/generate", s.handleGenerateDSL).Methods("POST")
	api.HandleFunc("/dsl/validate", s.handleValidateDSL).Methods("POST")
	api.HandleFunc("/dsl/execute", s.handleExecuteDSL).Methods("POST")
	api.HandleFunc("/session/{id}", s.handleGetSession).Methods("GET")
	api.HandleFunc("/session/{id}/history", s.handleGetHistory).Methods("GET")
	api.HandleFunc("/domains", s.handleGetDomains).Methods("GET")
	api.HandleFunc("/domains/{domain}/vocabulary", s.handleGetVocabulary).Methods("GET")
	api.HandleFunc("/vocabulary", s.handleGetAllVocabularies).Methods("GET")
	api.HandleFunc("/attributes", s.handleGetAttributes).Methods("GET")
	api.HandleFunc("/routing/metrics", s.handleGetRoutingMetrics).Methods("GET")

	// WebSocket endpoint
	s.router.HandleFunc("/ws", s.handleWebSocket)

	// Serve index.html for all other routes (SPA)
	s.router.PathPrefix("/").HandlerFunc(s.handleIndex)
}

// handleHealth returns server health status
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	registryHealthy := s.registry.IsHealthy()
	domains := s.registry.ListWithMetadata()

	healthStatus := map[string]interface{}{
		"status":           "healthy",
		"service":          "multi-domain-dsl-agent",
		"registry_healthy": registryHealthy,
		"domains":          len(domains),
		"time":             time.Now().UTC(),
	}

	if !registryHealthy {
		healthStatus["status"] = "degraded"
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(healthStatus)
}

// handleChat processes a chat message using multi-domain routing
func (s *Server) handleChat(w http.ResponseWriter, r *http.Request) {
	var req ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.respondError(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Get or create session
	session := s.getOrCreateSession(req.SessionID, req.Domain)

	// Add user message to history
	userMsg := ChatMessage{
		Role:      "user",
		Content:   req.Message,
		Domain:    session.CurrentDomain,
		Timestamp: time.Now().UTC(),
	}
	session.History = append(session.History, userMsg)

	ctx := r.Context()

	// Route to appropriate domain
	routingReq := &registry.RoutingRequest{
		Message:       req.Message,
		SessionID:     session.SessionID,
		CurrentDomain: session.CurrentDomain,
		Context:       session.Context,
		ExistingDSL:   session.BuiltDSL,
		Timestamp:     time.Now(),
	}

	routingResp, err := s.domainRouter.Route(ctx, routingReq)
	if err != nil {
		s.respondError(w, fmt.Sprintf("Failed to route message: %v", err), http.StatusInternalServerError)
		return
	}

	// Update session domain if it changed
	if routingResp.DomainName != session.CurrentDomain {
		session.CurrentDomain = routingResp.DomainName
		log.Printf("Session %s switched to domain: %s (reason: %s)", session.SessionID, routingResp.DomainName, routingResp.Reason)
	}

	// Generate DSL using the selected domain
	genReq := &registry.GenerationRequest{
		Instruction:   req.Message,
		SessionID:     session.SessionID,
		CurrentDomain: session.CurrentDomain,
		Context:       session.Context,
		ExistingDSL:   session.BuiltDSL,
		Timestamp:     time.Now(),
	}

	genResp, err := routingResp.Domain.GenerateDSL(ctx, genReq)
	if err != nil {
		s.respondError(w, fmt.Sprintf("Failed to generate DSL: %v", err), http.StatusInternalServerError)
		return
	}

	// Validate verbs
	if err := routingResp.Domain.ValidateVerbs(genResp.DSL); err != nil {
		s.respondError(w, fmt.Sprintf("DSL validation failed: %v", err), http.StatusBadRequest)
		return
	}

	// Accumulate DSL through DSL State Manager (single source of truth)
	err = s.sessionMgr.AccumulateDSL(session.SessionID, genResp.DSL)
	if err != nil {
		s.respondError(w, fmt.Sprintf("Failed to accumulate DSL: %v", err), http.StatusInternalServerError)
		return
	}

	// Get updated DSL from state manager
	updatedSession, err := s.sessionMgr.Get(session.SessionID)
	if err != nil {
		s.respondError(w, fmt.Sprintf("Failed to get updated session: %v", err), http.StatusInternalServerError)
		return
	}
	session.BuiltDSL = updatedSession.GetDSL()

	// Update session context with any new values
	if genResp.ContextUpdates != nil {
		if session.Context == nil {
			session.Context = make(map[string]interface{})
		}
		for k, v := range genResp.ContextUpdates {
			session.Context[k] = v
		}
	}

	// Add agent response to history
	agentMsg := ChatMessage{
		Role:      "agent",
		Content:   genResp.Explanation,
		DSL:       session.BuiltDSL, // Complete accumulated DSL
		Fragment:  genResp.DSL,      // Individual fragment for this operation
		Domain:    session.CurrentDomain,
		Response:  genResp,
		Timestamp: time.Now().UTC(),
	}
	session.History = append(session.History, agentMsg)
	session.LastUsed = time.Now().UTC()

	// Send response
	chatResp := ChatResponse{
		SessionID: session.SessionID,
		Message:   genResp.Explanation,
		DSL:       session.BuiltDSL, // Complete accumulated DSL
		Fragment:  genResp.DSL,      // Individual operation
		Domain:    session.CurrentDomain,
		Response:  genResp,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(chatResp)
}

// handleGenerateDSL generates DSL using domain routing
func (s *Server) handleGenerateDSL(w http.ResponseWriter, r *http.Request) {
	var req registry.GenerationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Route to appropriate domain if not specified
	if req.CurrentDomain == "" {
		routingReq := &registry.RoutingRequest{
			Message:   req.Instruction,
			SessionID: req.SessionID,
			Context:   req.Context,
			Timestamp: time.Now(),
		}

		routingResp, err := s.domainRouter.Route(r.Context(), routingReq)
		if err != nil {
			s.respondError(w, fmt.Sprintf("Failed to route request: %v", err), http.StatusInternalServerError)
			return
		}
		req.CurrentDomain = routingResp.DomainName
	}

	// Get domain
	domain, err := s.registry.Get(req.CurrentDomain)
	if err != nil {
		s.respondError(w, fmt.Sprintf("Domain not found: %v", err), http.StatusNotFound)
		return
	}

	// Generate DSL
	response, err := domain.GenerateDSL(r.Context(), &req)
	if err != nil {
		s.respondError(w, fmt.Sprintf("DSL generation failed: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleValidateDSL validates a DSL string using domain validation
func (s *Server) handleValidateDSL(w http.ResponseWriter, r *http.Request) {
	var req struct {
		DSL    string `json:"dsl"`
		Domain string `json:"domain,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// If no domain specified, try to route based on DSL verbs
	domainName := req.Domain
	if domainName == "" {
		routingReq := &registry.RoutingRequest{
			Message:     req.DSL,
			ExistingDSL: req.DSL,
			Timestamp:   time.Now(),
		}

		routingResp, err := s.domainRouter.Route(r.Context(), routingReq)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to determine domain: %v", err), http.StatusBadRequest)
			return
		}
		domainName = routingResp.DomainName
	}

	// Get domain
	domain, err := s.registry.Get(domainName)
	if err != nil {
		http.Error(w, fmt.Sprintf("Domain not found: %v", err), http.StatusNotFound)
		return
	}

	// Validate DSL
	err = domain.ValidateVerbs(req.DSL)
	valid := err == nil
	errors := []string{}
	if err != nil {
		errors = append(errors, err.Error())
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"valid":  valid,
		"domain": domainName,
		"errors": errors,
	})
}

// handleExecuteDSL executes a validated DSL operation
func (s *Server) handleExecuteDSL(w http.ResponseWriter, r *http.Request) {
	var req struct {
		DSL       string `json:"dsl"`
		Domain    string `json:"domain,omitempty"`
		SessionID string `json:"session_id,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// TODO: Implement actual DSL execution through domains
	// This would involve parsing the DSL and executing the operations
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "success",
		"message": "DSL execution pending implementation",
		"domain":  req.Domain,
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

// handleGetDomains returns all available domains
func (s *Server) handleGetDomains(w http.ResponseWriter, r *http.Request) {
	domainsInfo := s.registry.ListWithMetadata()

	// Convert to client-friendly format
	domains := make(map[string]DomainInfo)
	for name, info := range domainsInfo {
		domain, _ := s.registry.Get(name)
		domainVocab := domain.GetVocabulary()

		// Count verbs by category
		categories := make(map[string]int)
		for _, verbDef := range domainVocab.Verbs {
			categories[verbDef.Category]++
		}

		domains[name] = DomainInfo{
			Name:        info.Name,
			Version:     info.Version,
			Description: info.Description,
			IsHealthy:   info.IsHealthy,
			VerbCount:   len(domainVocab.Verbs),
			Categories:  categories,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"domains": domains,
		"total":   len(domains),
	})
}

// handleGetVocabulary returns the vocabulary for a specific domain
func (s *Server) handleGetVocabulary(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	domainName := vars["domain"]

	vocab, err := s.registry.GetVocabulary(domainName)
	if err != nil {
		http.Error(w, fmt.Sprintf("Domain not found: %v", err), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(vocab)
}

// handleGetAllVocabularies returns vocabularies for all domains
func (s *Server) handleGetAllVocabularies(w http.ResponseWriter, r *http.Request) {
	vocabularies := s.registry.GetAllVocabularies()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(vocabularies)
}

// handleGetAttributes returns available attributes from dictionary
func (s *Server) handleGetAttributes(w http.ResponseWriter, r *http.Request) {
	// TODO: Query dictionary service for attributes
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

// handleGetRoutingMetrics returns routing statistics
func (s *Server) handleGetRoutingMetrics(w http.ResponseWriter, r *http.Request) {
	metrics := s.domainRouter.GetRoutingMetrics()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metrics)
}

// handleWebSocket handles WebSocket connections for real-time multi-domain chat
func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}
	defer conn.Close()

	sessionID := uuid.New().String()
	session := s.getOrCreateSession(sessionID, "hedge-fund-investor") // Default to hedge fund for backward compatibility

	log.Printf("WebSocket connection established: %s", sessionID)

	// Send welcome message
	welcomeMsg := map[string]interface{}{
		"type": "welcome",
		"payload": map[string]interface{}{
			"session_id":        sessionID,
			"current_domain":    session.CurrentDomain,
			"message":           "Connected to Multi-Domain DSL Agent. How can I help you?",
			"available_domains": s.registry.List(),
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

		ctx := context.Background()

		// Route to appropriate domain
		routingReq := &registry.RoutingRequest{
			Message:       chatReq.Message,
			SessionID:     session.SessionID,
			CurrentDomain: session.CurrentDomain,
			Context:       session.Context,
			ExistingDSL:   session.BuiltDSL,
			Timestamp:     time.Now(),
		}

		routingResp, err := s.domainRouter.Route(ctx, routingReq)
		if err != nil {
			s.sendWSError(conn, fmt.Sprintf("Failed to route message: %v", err))
			return
		}

		// Update session domain if it changed
		if routingResp.DomainName != session.CurrentDomain {
			session.CurrentDomain = routingResp.DomainName
		}

		// Generate DSL
		genReq := &registry.GenerationRequest{
			Instruction:   chatReq.Message,
			SessionID:     session.SessionID,
			CurrentDomain: session.CurrentDomain,
			Context:       session.Context,
			ExistingDSL:   session.BuiltDSL,
			Timestamp:     time.Now(),
		}

		genResp, err := routingResp.Domain.GenerateDSL(ctx, genReq)
		if err != nil {
			s.sendWSError(conn, fmt.Sprintf("Failed to generate DSL: %v", err))
			return
		}

		// Accumulate DSL through DSL State Manager (single source of truth)
		err = s.sessionMgr.AccumulateDSL(session.SessionID, genResp.DSL)
		if err != nil {
			s.sendWSError(conn, fmt.Sprintf("Failed to accumulate DSL: %v", err))
			return
		}

		// Get updated DSL from state manager
		updatedSession, err := s.sessionMgr.Get(session.SessionID)
		if err != nil {
			s.sendWSError(conn, fmt.Sprintf("Failed to get updated session: %v", err))
			return
		}
		session.BuiltDSL = updatedSession.GetDSL()

		// Update session context
		if genResp.ContextUpdates != nil {
			if session.Context == nil {
				session.Context = make(map[string]interface{})
			}
			for k, v := range genResp.ContextUpdates {
				session.Context[k] = v
			}
		}

		session.LastUsed = time.Now().UTC()

		// Send response
		conn.WriteJSON(map[string]interface{}{
			"type": "chat_response",
			"payload": map[string]interface{}{
				"message":        genResp.Explanation,
				"dsl":            session.BuiltDSL, // Complete accumulated DSL
				"fragment":       genResp.DSL,      // Individual operation
				"domain":         session.CurrentDomain,
				"verb":           genResp.Verb,
				"from_state":     genResp.FromState,
				"to_state":       genResp.ToState,
				"confidence":     genResp.Confidence,
				"routing_reason": routingResp.Reason,
				"response":       genResp,
			},
		})

	case "switch_domain":
		var switchReq struct {
			Domain string `json:"domain"`
		}
		if err := json.Unmarshal(msg.Payload, &switchReq); err != nil {
			s.sendWSError(conn, "Invalid domain switch request")
			return
		}

		// Validate domain exists
		if _, err := s.registry.Get(switchReq.Domain); err != nil {
			s.sendWSError(conn, fmt.Sprintf("Domain not found: %s", switchReq.Domain))
			return
		}

		// Update session
		session.CurrentDomain = switchReq.Domain
		session.LastUsed = time.Now().UTC()

		conn.WriteJSON(map[string]interface{}{
			"type": "domain_switched",
			"payload": map[string]interface{}{
				"domain":  switchReq.Domain,
				"message": fmt.Sprintf("Switched to %s domain", switchReq.Domain),
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
	http.ServeFile(w, r, "./static/index.html")
}

// getOrCreateSession gets existing session or creates new one with domain support
func (s *Server) getOrCreateSession(sessionID, defaultDomain string) *ChatSession {
	s.mu.Lock()
	defer s.mu.Unlock()

	session, exists := s.sessions[sessionID]
	if !exists {
		// Use provided default domain or fallback to hedge-fund-investor
		if defaultDomain == "" {
			defaultDomain = "hedge-fund-investor"
		}

		// Validate domain exists
		if _, err := s.registry.Get(defaultDomain); err != nil {
			log.Printf("Warning: Default domain '%s' not found, using hedge-fund-investor", defaultDomain)
			defaultDomain = "hedge-fund-investor"
		}

		session = &ChatSession{
			SessionID:     sessionID,
			CurrentDomain: defaultDomain,
			Context:       make(map[string]interface{}),
			BuiltDSL:      "",
			History:       []ChatMessage{},
			CreatedAt:     time.Now().UTC(),
			LastUsed:      time.Now().UTC(),
		}
		s.sessions[sessionID] = session
		log.Printf("Created new session: %s (domain: %s)", sessionID, defaultDomain)
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
	log.Printf("Starting Multi-Domain DSL Agent Web Server on %s", addr)
	log.Printf("Registered domains: %v", s.registry.List())
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

	// Initialize dictionary service (mock for now)
	var dictService interface{}

	// Initialize datastore (can be nil for now)
	var store datastore.DataStore

	// Create server
	server, err := NewServer(dictService, store, apiKey)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

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
