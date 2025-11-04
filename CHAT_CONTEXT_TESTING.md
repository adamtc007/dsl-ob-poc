# Chat Context Testing Guide

This guide demonstrates how the DSL Agent's conversational context tracking works, allowing natural multi-turn conversations without repeating entity references.

## Overview

The chat agent maintains **session context** that includes:
- `investor_id` - UUID of the current investor
- `investor_name` - Legal name of the investor
- `investor_type` - Type (INDIVIDUAL, CORPORATE, etc.)
- `domicile` - Jurisdiction
- `fund_id` - Current fund reference
- `class_id` - Current share class reference
- `series_id` - Current series reference
- `current_state` - Lifecycle state (OPPORTUNITY, KYC_PENDING, etc.)

Once an entity is mentioned, the agent remembers it for the rest of the session.

---

## Test Scenario 1: New Investor Flow

### Conversation Flow

**Turn 1: Create Opportunity**
```
User: "Create opportunity for adam cearns into Absolute Alpha Fund"

Agent Response:
- Generates DSL: (investor.start-opportunity :legal-name "adam cearns" :type "INDIVIDUAL")
- Server generates UUID: "a1b2c3d4-..."
- Context stored:
  * investor_id: "a1b2c3d4-..."
  * investor_name: "adam cearns"
  * investor_type: "INDIVIDUAL"
  * current_state: "OPPORTUNITY"
```

**Turn 2: Amend Details (Implicit Reference)**
```
User: "Set domicile to UK"
OR
User: "Set their domicile to UK"
OR
User: "Update domicile to UK for this investor"

Agent Response:
- Uses investor_id from context automatically
- Generates DSL: (investor.amend-details :investor "a1b2c3d4-..." :domicile "UK")
- Context updated:
  * domicile: "UK"
```

**Turn 3: Start KYC (No Explicit Investor)**
```
User: "Start KYC"
OR
User: "Begin KYC for this investor"
OR
User: "Start KYC process"

Agent Response:
- Uses investor_id from context
- Uses investor_name for explanation
- Generates DSL: (kyc.begin :investor "a1b2c3d4-..." :tier "STANDARD")
- Context updated:
  * current_state: "KYC_PENDING"
```

**Turn 4: Collect Document**
```
User: "Collect passport document"

Agent Response:
- Uses investor_id from context
- Generates DSL: (kyc.collect-doc :investor "a1b2c3d4-..." :doc-type "PASSPORT")
- Explanation: "Collecting passport document for adam cearns"
```

---

## Test Scenario 2: Multiple Entities

### Conversation Flow

**Turn 1: Create Corporate Investor**
```
User: "Create opportunity for Alpine Capital AG, a Swiss corporation"

Context After:
- investor_id: "xyz-789-..."
- investor_name: "Alpine Capital AG"
- investor_type: "CORPORATE"
- domicile: "CH"
- current_state: "OPPORTUNITY"
```

**Turn 2: Record Indication with Fund**
```
User: "Record indication of $5M for Global Equity Fund Class A"

Context After:
- investor_id: "xyz-789-..." (unchanged)
- fund_id: "<fund-uuid>" (new)
- class_id: "<class-uuid>" (new)
- current_state: "PRECHECKS"
```

**Turn 3: Start KYC (All Context Preserved)**
```
User: "Start enhanced KYC"

Agent Response:
- Uses investor_id: "xyz-789-..."
- Uses fund_id: "<fund-uuid>" (if needed)
- Generates: (kyc.begin :investor "xyz-789-..." :tier "ENHANCED")
```

---

## Test Scenario 3: Context Switch Detection

### Creating a New Investor Mid-Session

**Turn 1: First Investor**
```
User: "Create opportunity for John Smith"

Context:
- investor_id: "aaa-111-..."
- investor_name: "John Smith"
```

**Turn 2: Operations on John**
```
User: "Start KYC"
→ Uses investor_id "aaa-111-..." (John Smith)
```

**Turn 3: NEW Investor**
```
User: "Create opportunity for Jane Doe"

Context REPLACED:
- investor_id: "bbb-222-..." (NEW UUID)
- investor_name: "Jane Doe"
- (John Smith's context is replaced)
```

**Turn 4: Operations Apply to Jane**
```
User: "Start KYC"
→ Uses investor_id "bbb-222-..." (Jane Doe, current context)
```

**Note**: To work with John again, you'd need to explicitly mention "John Smith" or start a new session.

---

## Test Scenario 4: Pronouns and Implicit References

The agent understands these implicit references:

| User Input | Agent Interpretation |
|------------|---------------------|
| "Start KYC" | Uses `investor_id` from context |
| "Their domicile" | Refers to investor in context |
| "This investor" | Uses `investor_id` from context |
| "The investor" | Uses `investor_id` from context |
| "Them" | Uses `investor_id` from context |
| "This fund" | Uses `fund_id` from context |
| "The fund" | Uses `fund_id` from context |

---

## Testing in the UI

### Access the Chat Interface
```bash
# Server should be running on:
http://localhost:8080
```

### Test Sequence

1. **Create adam cearns**:
   ```
   Create opportunity for adam cearns into Absolute Alpha Fund
   ```
   - Verify UUID is generated and shown
   - Check that context is stored

2. **Add domicile without repeating name**:
   ```
   Set domicile to UK
   ```
   - Should use adam cearns' UUID automatically
   - Explanation should mention "adam cearns"

3. **Start KYC implicitly**:
   ```
   Start KYC
   ```
   - Should use adam cearns' UUID
   - Should transition state to KYC_PENDING

4. **Continue operations**:
   ```
   Collect passport
   Screen against WorldCheck
   Approve KYC with low risk
   ```
   - All should use adam cearns' UUID
   - No need to repeat the investor name

---

## Expected DSL Output

For the adam cearns flow above, the complete DSL should look like:

```lisp
;; Turn 1: Create opportunity
(investor.start-opportunity
  :legal-name "adam cearns"
  :type "INDIVIDUAL")
;; Server assigns UUID: a1b2c3d4-...

;; Turn 2: Add domicile
(investor.amend-details
  :investor "a1b2c3d4-..."
  :domicile "UK")

;; Turn 3: Start KYC
(kyc.begin
  :investor "a1b2c3d4-..."
  :tier "STANDARD")

;; Turn 4: Collect document
(kyc.collect-doc
  :investor "a1b2c3d4-..."
  :doc-type "PASSPORT")
```

Notice how the UUID appears in every operation after the first, even though the user never typed it!

---

## Implementation Details

### Server-Side Context Management

**Location**: `hedge-fund-investor-source/web/server.go`

```go
// Context updated after each DSL generation
session.Context.InvestorID = uuid.New().String()
session.Context.InvestorName = "adam cearns"
session.Context.CurrentState = "OPPORTUNITY"
```

### Agent Prompt Engineering

**Location**: `hedge-fund-investor-source/web/internal/hf-agent/hf_dsl_agent.go`

System prompt includes:
```
## CONTEXT AWARENESS
When context is provided (investor_id, investor_name, fund_id, etc.), USE IT automatically:
- "this investor", "the investor", "them" → use investor_id from context
- "start KYC" (without specifying who) → use investor_id from context
```

### Context Persistence

Contexts are stored per session ID:
- Session created on first message
- Session ID returned to client
- Client includes session ID in subsequent requests
- Context persists until server restart or session timeout

---

## Troubleshooting

### Problem: Agent doesn't use context

**Check**:
1. Is `session_id` being sent by client?
2. Is context being printed in server logs?
3. Is the agent receiving the context in the prompt?

**Debug**:
```bash
# Check server logs
tail -f hedge-fund-investor-source/web/server.log
```

### Problem: Context gets overwritten

**Expected behavior**: Creating a new investor replaces the context.

**Solution**: If you need to work with multiple investors in one conversation, you'll need to explicitly name them each time, or implement multi-entity context tracking (future enhancement).

---

## Future Enhancements

1. **Multi-Entity Context**: Track multiple investors, funds in same session
2. **Context Switching**: "Switch to John Smith" to change active entity
3. **Context History**: "Go back to previous investor"
4. **Context Export**: Download session context as JSON
5. **Context Persistence**: Save context to database for long-running sessions

---

## API Testing with curl

```bash
# Create session with adam cearns
curl -X POST http://localhost:8080/api/chat \
  -H "Content-Type: application/json" \
  -d '{
    "message": "Create opportunity for adam cearns"
  }'

# Response includes session_id, save it
SESSION_ID="<from-response>"

# Continue conversation with context
curl -X POST http://localhost:8080/api/chat \
  -H "Content-Type: application/json" \
  -d "{
    \"session_id\": \"$SESSION_ID\",
    \"message\": \"Set domicile to UK\"
  }"

# Agent automatically uses adam cearns' UUID from context!
```

---

## Success Criteria

✅ **Pass**: User creates investor, then says "Start KYC" and agent uses the investor's UUID

✅ **Pass**: User creates investor, then says "their domicile is UK" and agent updates correct investor

✅ **Pass**: Agent explanations mention investor name even when user just says "this investor"

✅ **Pass**: DSL contains correct UUIDs without user ever typing them

❌ **Fail**: Agent uses placeholder `<investor_id>` when context exists

❌ **Fail**: Agent asks "which investor?" when context is available

---

**The goal**: Natural conversation where entities are remembered, just like talking to a human assistant! 🎯