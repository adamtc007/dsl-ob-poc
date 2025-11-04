/**
 * Types for the Hedge Fund DSL Agent UI
 */

// Chat Message Types
export interface ChatMessage {
  role: "user" | "agent";
  content: string;
  dsl?: string;
  response?: DSLGenerationResponse;
  timestamp: Date;
}

// API Response Types
export interface ChatResponse {
  sessionId: string;
  message: string;
  dsl?: string;
  response?: DSLGenerationResponse;
  error?: string;
}

export interface DSLGenerationResponse {
  verb: string;
  parameters: Record<string, unknown>;
  explanation: string;
  dsl: string;
  from_state: string;
  to_state: string;
  confidence: number;
  generated_at: string;
}

// Session Management
export interface Session {
  sessionId: string;
  context: DSLGenerationRequest;
  history: ChatMessage[];
  createdAt: Date;
  lastUsed: Date;
}

// Request Types
export interface ChatRequest {
  sessionId?: string;
  message: string;
  context?: Record<string, unknown>;
}

export interface DSLGenerationRequest {
  instruction: string;
  currentState?: string;
  investorId?: string;
  parameters?: Record<string, unknown>;
}

// Dictionary and Attributes
export interface DictionaryAttribute {
  id: string;
  name: string;
  type: string;
  description?: string;
  default_value?: unknown;
  required?: boolean;
  group?: string;
}

// WebSocket Message Types
export interface WebSocketMessage {
  type: string;
  payload: Record<string, unknown>;
}

export interface WebSocketChatRequest {
  message: string;
  context?: Record<string, unknown>;
}

// DSL Vocabulary Types
export interface DSLVocabulary {
  verbs: Record<string, DSLVerb>;
  states: string[];
  transitionRules: TransitionRule[];
}

export interface DSLVerb {
  name: string;
  description: string;
  parameters: Record<string, DSLParameter>;
  requiredParameters: string[];
  allowedStates: string[];
  transitionsTo?: string;
}

export interface DSLParameter {
  name: string;
  type: string;
  description: string;
  required: boolean;
  allowedValues?: string[];
}

export interface TransitionRule {
  fromState: string;
  toState: string;
  allowedVerbs: string[];
}
