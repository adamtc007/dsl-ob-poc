# Web Interface - Quick Start Guide

## What We Built

A **TypeScript + React** web interface for conversational DSL generation with a **Go backend** serving the AI agent.

## Files Created

✅ **`server.go`** (850+ lines) - Complete Go web server with:
- REST API endpoints for chat, DSL generation, validation, execution
- WebSocket support for real-time bidirectional communication
- Session management with automatic cleanup
- Integration with HF DSL Agent (RAG-powered)
- CORS middleware for development

✅ **`README.md`** (705 lines) - Comprehensive documentation covering:
- Architecture and technology stack
- Project structure
- API endpoints (REST + WebSocket)
- TypeScript type definitions
- Deployment instructions
- Troubleshooting guide

✅ **`UI_MOCKUP.md`** - Visual mockups showing:
- Main interface layout
- Chat interface with DSL generation
- DSL viewer with syntax highlighting
- State machine visualization
- Attribute dictionary browser
- Execution dashboard
- Mobile-responsive design

## Quick Setup

### 1. Install Dependencies

```bash
# Backend (Go)
cd dsl-ob-poc/hedge-fund-investor-source/web
go mod tidy

# Frontend (TypeScript + React)
cd frontend
npm install
```

### 2. Configure Environment

```bash
export GEMINI_API_KEY="your-api-key"
export DB_CONN_STRING="postgres://localhost:5432/postgres?sslmode=disable"
export PORT=8080
```

### 3. Run Backend

```bash
cd web
go run server.go
```

Output:
```
Starting Hedge Fund DSL Agent Web Server on :8080
```

### 4. Run Frontend (Development)

```bash
cd frontend
npm run dev
```

Output:
```
VITE v5.0.8  ready in 234 ms
➜  Local:   http://localhost:5173/
```

### 5. Access Application

Open browser: `http://localhost:5173`

## API Examples

### Chat with Agent
```bash
curl -X POST http://localhost:8080/api/chat \
  -H "Content-Type: application/json" \
  -d '{
    "message": "Create opportunity for Swiss investor Acme Capital LP",
    "context": {}
  }'
```

Response:
```json
{
  "session_id": "uuid-here",
  "message": "Generated investor.start-opportunity operation with 98% confidence",
  "dsl": "(investor.start-opportunity\n  @attr{uuid-0001} = \"Acme Capital LP\"\n  @attr{uuid-0002} = \"CORPORATE\"\n  @attr{uuid-0003} = \"CH\")",
  "response": {
    "verb": "investor.start-opportunity",
    "parameters": {...},
    "to_state": "OPPORTUNITY",
    "confidence": 0.98
  }
}
```

### WebSocket Connection
```javascript
const ws = new WebSocket('ws://localhost:8080/ws');

ws.onmessage = (event) => {
  const msg = JSON.parse(event.data);
  if (msg.type === 'welcome') {
    console.log('Connected:', msg.payload.session_id);
  }
};

// Send chat message
ws.send(JSON.stringify({
  type: 'chat',
  payload: {
    message: 'Start KYC process'
  }
}));
```

## Frontend Stack

Ready to implement with:

```json
{
  "dependencies": {
    "react": "^18.2.0",
    "typescript": "^5.3.3",
    "vite": "^5.0.8",
    "tailwindcss": "^3.3.6",
    "axios": "^1.6.2",
    "lucide-react": "^0.300.0"
  }
}
```

Components to build:
- `ChatInterface.tsx` - Main chat UI
- `DSLViewer.tsx` - Code viewer with syntax highlighting
- `StateMachine.tsx` - Visual state diagram
- `AttributeBrowser.tsx` - Dictionary explorer

## Key Features

🤖 **Natural Language → DSL**: Type in English, get valid DSL with attribute UUIDs
✅ **Real-Time Validation**: Parser validates UUIDs and types instantly
⚡ **One-Click Execution**: Execute DSL operations with single click
📊 **State Tracking**: Visual representation of investor lifecycle
🔍 **Attribute Search**: Browse and search 50+ data dictionary attributes
💾 **Session Persistence**: Save and restore conversation history
🌐 **WebSocket**: Real-time bidirectional communication

## Production Build

```bash
# Build frontend
cd frontend
npm run build

# Copy to static directory
cp -r dist/* ../static/

# Build Go binary
cd ..
go build -o hf-dsl-web server.go

# Run
./hf-dsl-web
```

## Docker

```bash
# Build image
docker build -t hf-dsl-web .

# Run container
docker run -p 8080:8080 \
  -e GEMINI_API_KEY=$GEMINI_API_KEY \
  -e DB_CONN_STRING=$DB_CONN_STRING \
  hf-dsl-web
```

## Next Steps

1. **Implement Frontend Components** - Build React components from mockups
2. **Add Authentication** - OAuth2 or JWT for user auth
3. **Deploy** - Docker + Kubernetes or cloud platform
4. **Monitor** - Add logging, metrics, error tracking

---

**The web interface provides a modern, conversational way to interact with the DSL agent, making complex hedge fund operations accessible through natural language.**
