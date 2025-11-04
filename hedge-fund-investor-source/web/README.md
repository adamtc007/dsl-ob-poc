# Hedge Fund DSL Agent - Web Interface

A modern, TypeScript-based web interface for the Hedge Fund Investor DSL Generation Agent with real-time chat, DSL preview, validation, and execution capabilities.

## Overview

This web application provides a conversational interface to interact with the AI-powered DSL generation agent. Users can describe what they want to do in natural language, and the agent generates validated, parseable DSL operations that can be executed against the hedge fund investor register system.

## Features

- 🤖 **AI-Powered Chat Interface**: Natural language conversation with DSL agent
- 📝 **Real-Time DSL Generation**: Instant DSL code generation from instructions
- ✅ **Syntax Validation**: Live validation of generated DSL with attribute UUID checking
- 🎯 **State Machine Tracking**: Visual representation of investor lifecycle states
- 📊 **Execution Dashboard**: Execute and monitor DSL operations
- 🔍 **Attribute Browser**: Browse and search the data dictionary
- 📚 **DSL Vocabulary Reference**: Complete documentation of all 17 verbs
- 💾 **Session Persistence**: Save and restore conversation history
- 🌐 **WebSocket Support**: Real-time bidirectional communication
- 🎨 **Modern UI**: Built with React, TypeScript, and Tailwind CSS

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                     Web Browser                             │
│  ┌───────────────────────────────────────────────────────┐  │
│  │          TypeScript React Application                 │  │
│  │  ┌─────────────────┐  ┌─────────────────────────┐   │  │
│  │  │  Chat Interface │  │  DSL Viewer/Editor      │   │  │
│  │  │  - Message UI   │  │  - Syntax Highlighting  │   │  │
│  │  │  - User Input   │  │  - Validation           │   │  │
│  │  │  - History      │  │  - Execution Control    │   │  │
│  │  └─────────────────┘  └─────────────────────────┘   │  │
│  │  ┌─────────────────┐  ┌─────────────────────────┐   │  │
│  │  │  State Machine  │  │  Attribute Dictionary   │   │  │
│  │  │  Visualization  │  │  Browser                │   │  │
│  │  └─────────────────┘  └─────────────────────────┘   │  │
│  └───────────────────────────────────────────────────────┘  │
└────────────────────────┬────────────────────────────────────┘
                         │
                         │ HTTP/REST API + WebSocket
                         │
┌────────────────────────┴────────────────────────────────────┐
│                   Go Web Server                             │
│  ┌───────────────────────────────────────────────────────┐  │
│  │  HTTP Handlers           │  WebSocket Handler        │  │
│  │  - /api/chat             │  - /ws                    │  │
│  │  - /api/dsl/generate     │  - Real-time messaging    │  │
│  │  - /api/dsl/validate     │  - Session management     │  │
│  │  - /api/dsl/execute      │                           │  │
│  └───────────────────────────────────────────────────────┘  │
│  ┌───────────────────────────────────────────────────────┐  │
│  │           Hedge Fund DSL Agent (RAG)                  │  │
│  │  - Semantic search on dictionary                      │  │
│  │  - Gemini AI integration                              │  │
│  │  - DSL generation with attribute UUIDs                │  │
│  └───────────────────────────────────────────────────────┘  │
└────────────────────────┬────────────────────────────────────┘
                         │
                         │ Database Connection
                         │
┌────────────────────────┴────────────────────────────────────┐
│                    PostgreSQL                               │
│  - dictionary table (attributes with RAG metadata)          │
│  - hf_dsl_executions (DSL execution history)                │
│  - hf_investors, hf_trades, etc. (domain tables)            │
└─────────────────────────────────────────────────────────────┘
```

## Technology Stack

### Backend (Go)
- **gorilla/mux**: HTTP routing
- **gorilla/websocket**: WebSocket support
- **Gemini AI SDK**: RAG-powered DSL generation
- **PostgreSQL**: Data storage

### Frontend (TypeScript + React)
- **React 18**: UI framework
- **TypeScript**: Type safety
- **Vite**: Build tool and dev server
- **Tailwind CSS**: Styling
- **Axios**: HTTP client
- **react-markdown**: Markdown rendering
- **prism-react-renderer**: Code syntax highlighting
- **lucide-react**: Icons

## Project Structure

```
web/
├── server.go                    # Go web server
├── frontend/
│   ├── package.json            # Node dependencies
│   ├── tsconfig.json           # TypeScript config
│   ├── vite.config.ts          # Vite config
│   ├── tailwind.config.js      # Tailwind CSS config
│   ├── postcss.config.js       # PostCSS config
│   ├── index.html              # HTML entry point
│   └── src/
│       ├── main.tsx            # React entry point
│       ├── App.tsx             # Main application component
│       ├── types/              # TypeScript type definitions
│       │   ├── api.ts          # API types
│       │   └── dsl.ts          # DSL types
│       ├── components/         # React components
│       │   ├── ChatInterface.tsx
│       │   ├── DSLViewer.tsx
│       │   ├── StateMachine.tsx
│       │   ├── AttributeBrowser.tsx
│       │   ├── VocabularyReference.tsx
│       │   └── ExecutionDashboard.tsx
│       ├── hooks/              # Custom React hooks
│       │   ├── useWebSocket.ts
│       │   ├── useChat.ts
│       │   └── useDSL.ts
│       ├── services/           # API services
│       │   ├── api.ts          # HTTP API client
│       │   └── websocket.ts   # WebSocket client
│       └── styles/             # CSS files
│           └── globals.css
└── static/                     # Built static files (generated)
```

## Quick Start

### Prerequisites

- Go 1.21+ installed
- Node.js 18+ and npm installed
- PostgreSQL database with schema applied
- `GEMINI_API_KEY` or `GOOGLE_API_KEY` environment variable set

### Backend Setup

```bash
# From the web directory
cd dsl-ob-poc/hedge-fund-investor-source/web

# Install Go dependencies
go mod download

# Set environment variables
export GEMINI_API_KEY="your-api-key"
export DB_CONN_STRING="postgres://localhost:5432/postgres?sslmode=disable"
export PORT=8080

# Run the server
go run server.go
```

Server will start on `http://localhost:8080`

### Frontend Setup

```bash
# Navigate to frontend directory
cd frontend

# Install dependencies
npm install

# Run development server
npm run dev
```

Development server will start on `http://localhost:5173` (Vite default)

### Production Build

```bash
# Build frontend
cd frontend
npm run build

# Output goes to: dist/
# Copy to Go server's static directory:
cp -r dist/* ../static/

# Run Go server (serves static files from /static/)
cd ..
go run server.go
```

Access application at `http://localhost:8080`

## API Endpoints

### REST API

#### Health Check
```
GET /api/health
Response: {"status": "healthy", "service": "hedge-fund-dsl-agent", "time": "2024-01-15T10:00:00Z"}
```

#### Chat (Generate DSL from Natural Language)
```
POST /api/chat
Content-Type: application/json

Request:
{
  "session_id": "optional-session-uuid",
  "message": "Create an opportunity for Acme Capital LP, Swiss corporate investor",
  "context": {
    "investor_id": "optional-uuid",
    "current_state": "optional-state"
  }
}

Response:
{
  "session_id": "uuid",
  "message": "Generated investor.start-opportunity operation with 98% confidence",
  "dsl": "(investor.start-opportunity\n  @attr{uuid-0001} = \"Acme Capital LP\"\n  @attr{uuid-0002} = \"CORPORATE\"\n  @attr{uuid-0003} = \"CH\")",
  "response": {
    "verb": "investor.start-opportunity",
    "parameters": {...},
    "from_state": "",
    "to_state": "OPPORTUNITY",
    "guard_conditions": [],
    "explanation": "Creates initial investor opportunity record...",
    "confidence": 0.98,
    "warnings": []
  }
}
```

#### Generate DSL (Direct)
```
POST /api/dsl/generate
Content-Type: application/json

Request:
{
  "instruction": "Submit subscription for $5M",
  "investor_id": "uuid",
  "current_state": "KYC_APPROVED",
  "fund_id": "uuid",
  "class_id": "uuid"
}

Response: (same as DSLGenerationResponse)
```

#### Validate DSL
```
POST /api/dsl/validate
Content-Type: application/json

Request:
{
  "dsl": "(investor.start-opportunity @attr{uuid-0001} = \"Acme\" ...)"
}

Response:
{
  "valid": true,
  "errors": []
}
```

#### Execute DSL
```
POST /api/dsl/execute
Content-Type: application/json

Request:
{
  "dsl": "(investor.start-opportunity ...)",
  "session_id": "optional-uuid"
}

Response:
{
  "status": "success",
  "message": "DSL executed successfully",
  "result": {...}
}
```

#### Get Session
```
GET /api/session/{session_id}
Response: Full session object with context and history
```

#### Get History
```
GET /api/session/{session_id}/history
Response: Array of chat messages
```

#### Get Vocabulary
```
GET /api/vocabulary
Response: Complete DSL vocabulary with all 17 verbs
```

#### Get Attributes
```
GET /api/attributes
Response: Array of dictionary attributes with metadata
```

### WebSocket API

#### Connect
```
ws://localhost:8080/ws
```

#### Message Format
```json
{
  "type": "message_type",
  "payload": {...}
}
```

#### Message Types

**Client → Server:**

1. `chat` - Send chat message
```json
{
  "type": "chat",
  "payload": {
    "message": "Start KYC process",
    "context": {}
  }
}
```

2. `ping` - Heartbeat
```json
{
  "type": "ping",
  "payload": {}
}
```

**Server → Client:**

1. `welcome` - Connection established
```json
{
  "type": "welcome",
  "payload": {
    "session_id": "uuid",
    "message": "Connected to Hedge Fund DSL Agent"
  }
}
```

2. `chat_response` - DSL generated
```json
{
  "type": "chat_response",
  "payload": {
    "message": "Generated KYC operation",
    "dsl": "(kyc.begin ...)",
    "verb": "kyc.begin",
    "state": "KYC_PENDING",
    "response": {...}
  }
}
```

3. `error` - Error occurred
```json
{
  "type": "error",
  "payload": {
    "error": "Error message"
  }
}
```

4. `pong` - Heartbeat response
```json
{
  "type": "pong",
  "payload": {
    "time": "2024-01-15T10:00:00Z"
  }
}
```

## Frontend TypeScript Types

```typescript
// types/api.ts
export interface ChatRequest {
  session_id?: string;
  message: string;
  context?: Record<string, any>;
}

export interface ChatResponse {
  session_id: string;
  message: string;
  dsl?: string;
  response?: DSLGenerationResponse;
  error?: string;
}

export interface DSLGenerationResponse {
  dsl: string;
  verb: string;
  parameters: Record<string, any>;
  from_state?: string;
  to_state?: string;
  guard_conditions?: string[];
  explanation: string;
  confidence: number;
  warnings?: string[];
}

export interface ChatMessage {
  role: "user" | "agent";
  content: string;
  dsl?: string;
  response?: DSLGenerationResponse;
  timestamp: Date;
}

export interface Session {
  session_id: string;
  context: {
    investor_id?: string;
    current_state?: string;
    investor_type?: string;
    investor_name?: string;
  };
  history: ChatMessage[];
  created_at: Date;
  last_used: Date;
}

export interface DictionaryAttribute {
  attribute_id: string;
  name: string;
  long_description: string;
  group_id: string;
  mask: string;
  domain: string;
  vector: string;
}
```

## Example Usage

### Example 1: Complete Lifecycle Chat Session

```typescript
// User sends messages, agent responds with DSL

User: "Create opportunity for Acme Capital LP, Swiss corporate investor"
Agent: Generated investor.start-opportunity operation
DSL:
  (investor.start-opportunity
    @attr{a1b2c3d4-0001} = "Acme Capital LP"
    @attr{a1b2c3d4-0002} = "CORPORATE"
    @attr{a1b2c3d4-0003} = "CH")
State: → OPPORTUNITY

User: "They want to invest $5M in Global Opportunities Fund"
Agent: Generated investor.record-indication operation
DSL:
  (investor.record-indication
    @attr{a1b2c3d4-0004} = <investor_id>
    @attr{a1b2c3d4-0005} = <fund_id>
    @attr{a1b2c3d4-0006} = 5000000.00
    @attr{a1b2c3d4-0007} = "USD")
State: OPPORTUNITY → PRECHECKS

User: "Start standard KYC"
Agent: Generated kyc.begin operation
DSL:
  (kyc.begin
    @attr{a1b2c3d4-0004} = <investor_id>
    @attr{a1b2c3d4-0008} = "STANDARD")
State: PRECHECKS → KYC_PENDING

... continues through lifecycle
```

### Example 2: Direct API Usage

```typescript
import axios from 'axios';

const client = axios.create({
  baseURL: 'http://localhost:8080/api'
});

// Generate DSL
const response = await client.post('/chat', {
  message: 'Create opportunity for Swiss investor',
  context: {}
});

console.log(response.data.dsl);
// (investor.start-opportunity ...)

// Validate DSL
const validation = await client.post('/dsl/validate', {
  dsl: response.data.dsl
});

console.log(validation.data.valid); // true

// Execute DSL
const execution = await client.post('/dsl/execute', {
  dsl: response.data.dsl
});

console.log(execution.data.status); // "success"
```

## Development Workflow

### 1. Run Backend in Dev Mode
```bash
cd web
go run server.go
```

### 2. Run Frontend in Dev Mode
```bash
cd frontend
npm run dev
```

Frontend will proxy API requests to backend via Vite config:
```typescript
// vite.config.ts
export default defineConfig({
  server: {
    proxy: {
      '/api': 'http://localhost:8080',
      '/ws': {
        target: 'ws://localhost:8080',
        ws: true
      }
    }
  }
});
```

### 3. Make Changes
- Edit TypeScript files in `frontend/src/`
- Changes auto-reload in browser (hot module replacement)
- Backend changes require restart

### 4. Test
```bash
# Type checking
npm run type-check

# Linting
npm run lint

# Build test
npm run build
```

## Environment Variables

### Backend (Go)
```bash
GEMINI_API_KEY=your-gemini-api-key
DB_CONN_STRING=postgres://localhost:5432/postgres?sslmode=disable
PORT=8080
```

### Frontend (Vite)
```bash
# .env.development
VITE_API_URL=http://localhost:8080/api
VITE_WS_URL=ws://localhost:8080/ws

# .env.production
VITE_API_URL=/api
VITE_WS_URL=/ws
```

## Deployment

### Docker Deployment

```dockerfile
# Dockerfile
FROM node:18 AS frontend-build
WORKDIR /app/frontend
COPY frontend/package*.json ./
RUN npm install
COPY frontend/ ./
RUN npm run build

FROM golang:1.21 AS backend-build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o server ./web/server.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=backend-build /app/server .
COPY --from=frontend-build /app/frontend/dist ./static
EXPOSE 8080
CMD ["./server"]
```

Build and run:
```bash
docker build -t hf-dsl-agent-web .
docker run -p 8080:8080 \
  -e GEMINI_API_KEY=$GEMINI_API_KEY \
  -e DB_CONN_STRING=$DB_CONN_STRING \
  hf-dsl-agent-web
```

### Production Considerations

1. **Security**:
   - Use HTTPS (TLS certificates)
   - Add authentication/authorization
   - Implement rate limiting
   - Sanitize user inputs

2. **Performance**:
   - Enable gzip compression
   - Add caching headers for static assets
   - Use CDN for static files
   - Connection pooling for database

3. **Monitoring**:
   - Add structured logging
   - Metrics (Prometheus)
   - Error tracking (Sentry)
   - Health checks and readiness probes

4. **Scalability**:
   - Horizontal scaling with load balancer
   - Session store (Redis) for multi-instance deployments
   - WebSocket sticky sessions

## Troubleshooting

### WebSocket Connection Fails
- Check CORS settings in backend
- Verify WebSocket URL (ws:// vs wss://)
- Check firewall/proxy settings

### API Returns 404
- Ensure backend is running
- Check API base URL in frontend config
- Verify route registration in Go server

### DSL Generation Slow
- Check AI API key validity
- Monitor AI API rate limits
- Consider caching common operations

### Build Fails
- Clear node_modules: `rm -rf node_modules && npm install`
- Check Node.js version: `node --version`
- Verify TypeScript config

## Future Enhancements

- [ ] Multi-user authentication (OAuth2)
- [ ] Role-based access control
- [ ] DSL execution history with filtering
- [ ] Visual state machine editor
- [ ] Attribute dictionary editor
- [ ] DSL diff viewer for changes
- [ ] Export DSL as files
- [ ] Import/execute DSL files
- [ ] Real-time collaboration
- [ ] Mobile-responsive design improvements

## License

Internal use only - Part of DSL Onboarding POC

## Support

For questions or issues:
- Review documentation in `hedge-fund-investor-source/documentation/`
- Check `HEDGE_FUND_INVESTOR.md` for system overview
- Examine `ATTRIBUTE_BASED_DSL_ARCHITECTURE.md` for DSL details

---

**Version**: 1.0.0  
**Last Updated**: December 2024  
**Status**: Ready for development