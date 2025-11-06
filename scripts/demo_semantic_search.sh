#!/bin/bash
set -e

# ============================================================================
# Semantic Search Demonstration Script
# ============================================================================
# This script demonstrates the value of vector embeddings and semantic search
# for the dictionary attribute system. It shows how AI agents can discover
# relevant attributes using natural language queries.
#
# Prerequisites:
# - OpenAI API key set: export OPENAI_API_KEY="sk-..."
# - Database initialized with seed data
# - ./dsl-poc binary built
# ============================================================================

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

echo -e "${BLUE}╔═══════════════════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║         Semantic Search Demonstration for Dictionary         ║${NC}"
echo -e "${BLUE}║              AttributeID-as-Type + Vector RAG                 ║${NC}"
echo -e "${BLUE}╚═══════════════════════════════════════════════════════════════╝${NC}"
echo

# Check prerequisites
if [ ! -f "./dsl-poc" ]; then
    echo -e "${YELLOW}⚠️  ./dsl-poc binary not found. Building...${NC}"
    make build-greenteagc || { echo -e "${YELLOW}❌ Build failed${NC}"; exit 1; }
fi

if [ -z "$OPENAI_API_KEY" ]; then
    echo -e "${YELLOW}❌ OPENAI_API_KEY environment variable not set${NC}"
    echo -e "${CYAN}💡 Set it with: export OPENAI_API_KEY=\"sk-...\"${NC}"
    exit 1
fi

if [ -z "$DB_CONN_STRING" ]; then
    echo -e "${YELLOW}⚠️  DB_CONN_STRING not set. Using default: postgres://localhost:5432/postgres?sslmode=disable${NC}"
    export DB_CONN_STRING="postgres://localhost:5432/postgres?sslmode=disable"
fi

echo -e "${GREEN}✅ Prerequisites check passed${NC}"
echo

# Step 1: Check if embeddings exist
echo -e "${CYAN}════════════════════════════════════════════════════════════════${NC}"
echo -e "${CYAN}Step 1: Checking if dictionary has vector embeddings...${NC}"
echo -e "${CYAN}════════════════════════════════════════════════════════════════${NC}"
echo

# Try a quick semantic search to see if embeddings exist
echo -e "${YELLOW}Testing if attributes have embeddings...${NC}"
./dsl-poc semantic-search --query="test" --top=1 2>&1 | grep -q "no attributes have embeddings" && NEED_EMBEDDINGS=true || NEED_EMBEDDINGS=false

if [ "$NEED_EMBEDDINGS" = true ]; then
    echo -e "${YELLOW}⚠️  No embeddings found. Generating them now...${NC}"
    echo -e "${YELLOW}⏱️  This will take 2-3 minutes for 69 attributes (OpenAI API rate limits)${NC}"
    echo

    ./dsl-poc generate-embeddings --model=text-embedding-3-small --batch-size=5

    echo
    echo -e "${GREEN}✅ Embeddings generated successfully!${NC}"
else
    echo -e "${GREEN}✅ Embeddings already exist!${NC}"
fi

echo
echo -e "${CYAN}════════════════════════════════════════════════════════════════${NC}"
echo -e "${CYAN}Step 2: Semantic Search Demonstrations${NC}"
echo -e "${CYAN}════════════════════════════════════════════════════════════════${NC}"
echo

# ============================================================================
# Query 1: Wealth-related attributes
# ============================================================================
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${BLUE}Query 1: \"What attributes track someone's wealth?\"${NC}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${YELLOW}Use Case: AI agent needs to collect financial information for suitability${NC}"
echo

./dsl-poc semantic-search --query="What attributes track someone's wealth?" --top=5

echo
read -p "Press Enter to continue..."
echo

# ============================================================================
# Query 2: Identity documents
# ============================================================================
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${BLUE}Query 2: \"identity documents for individuals\"${NC}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${YELLOW}Use Case: AI agent building KYC document checklist for retail client${NC}"
echo

./dsl-poc semantic-search --query="identity documents for individuals" --top=5

echo
read -p "Press Enter to continue..."
echo

# ============================================================================
# Query 3: Corporate entity information
# ============================================================================
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${BLUE}Query 3: \"corporate entity registration information\"${NC}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${YELLOW}Use Case: AI agent onboarding institutional client${NC}"
echo

./dsl-poc semantic-search --query="corporate entity registration information" --top=5

echo
read -p "Press Enter to continue..."
echo

# ============================================================================
# Query 4: Risk assessment
# ============================================================================
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${BLUE}Query 4: \"risk assessment and compliance screening\"${NC}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${YELLOW}Use Case: AI agent determining compliance requirements${NC}"
echo

./dsl-poc semantic-search --query="risk assessment and compliance screening" --top=5

echo
read -p "Press Enter to continue..."
echo

# ============================================================================
# Query 5: Trust structures
# ============================================================================
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${BLUE}Query 5: \"trust settlor beneficiary information\"${NC}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${YELLOW}Use Case: AI agent handling trust UBO identification${NC}"
echo

./dsl-poc semantic-search --query="trust settlor beneficiary information" --top=5

echo
read -p "Press Enter to continue..."
echo

# ============================================================================
# Query 6: Domain-filtered search (KYC only)
# ============================================================================
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${BLUE}Query 6: \"person name address\" (filtered to KYC domain)${NC}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${YELLOW}Use Case: AI agent focused only on KYC-related attributes${NC}"
echo

./dsl-poc semantic-search --query="person name address" --domain=KYC --top=5

echo
read -p "Press Enter to continue..."
echo

# ============================================================================
# Summary
# ============================================================================
echo -e "${GREEN}╔═══════════════════════════════════════════════════════════════╗${NC}"
echo -e "${GREEN}║                    Demonstration Complete!                    ║${NC}"
echo -e "${GREEN}╚═══════════════════════════════════════════════════════════════╝${NC}"
echo
echo -e "${CYAN}✨ Key Takeaways:${NC}"
echo -e "${CYAN}─────────────────────────────────────────────────────────────────${NC}"
echo -e "1️⃣  ${GREEN}Natural Language Queries${NC}: AI agents can ask in plain English"
echo -e "   \"What tracks wealth?\" → finds net_worth, source_of_wealth, etc."
echo
echo -e "2️⃣  ${GREEN}Semantic Understanding${NC}: Matches meaning, not just keywords"
echo -e "   \"corporate registration\" → finds entity.legal_name, registration_number, etc."
echo
echo -e "3️⃣  ${GREEN}Context-Aware${NC}: Rich descriptions provide full context"
echo -e "   Each attribute includes when/where/why/how to use it"
echo
echo -e "4️⃣  ${GREEN}Domain Filtering${NC}: Can scope search to specific domains"
echo -e "   --domain=KYC limits to KYC attributes only"
echo
echo -e "5️⃣  ${GREEN}RAG for DSL Generation${NC}: AI agents retrieve relevant attributes"
echo -e "   Then generate valid DSL using actual UUIDs from dictionary"
echo
echo -e "${CYAN}─────────────────────────────────────────────────────────────────${NC}"
echo -e "${CYAN}💡 How This Enables AI Agents:${NC}"
echo -e "${CYAN}─────────────────────────────────────────────────────────────────${NC}"
echo -e "• Agent receives: \"Start KYC for US retail investor\""
echo -e "• Agent searches: \"individual identity documents US tax\""
echo -e "• Dictionary returns: passport, SSN, tax ID, address proof"
echo -e "• Agent generates DSL with correct attribute UUIDs"
echo -e "• No hallucination - only approved attributes from dictionary"
echo
echo -e "${GREEN}🎯 This is the AttributeID-as-Type pattern + Vector RAG in action!${NC}"
echo
