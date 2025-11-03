#!/bin/bash

# refresh_zips.sh - Refresh the two context zip files for the DSL Onboarding POC project
# Usage: ./refresh_zips.sh

set -e  # Exit on any error

echo "🔄 Refreshing DSL Onboarding POC context zip files..."
echo ""

# Remove existing zip files if they exist
for zip_file in internal_core.zip internal_cli.zip root_files.zip; do
    if [ -f "$zip_file" ]; then
        echo "🗑️  Removing existing $zip_file"
        rm "$zip_file"
    fi
done

echo ""

# Create Zip 1: Core internal packages (≤10 files, no tests)
echo "📦 Creating internal_core.zip (Core Packages)..."
echo "   Contains: store.go, dsl.go, agent.go, dictionary/attribute.go"
zip -r internal_core.zip internal/store/store.go internal/dsl/dsl.go internal/agent/agent.go internal/dictionary/attribute.go > /dev/null
echo "   ✅ internal_core.zip created successfully ($(unzip -l internal_core.zip | grep -E '\.(go|sql)$' | wc -l | tr -d ' ') source files)"

echo ""

# Create Zip 2: CLI and mocks (≤10 files, no tests)
echo "📦 Creating internal_cli.zip (CLI & Mocks)..."
echo "   Contains: CLI commands + data.go (no test files)"
zip -r internal_cli.zip internal/cli/create.go internal/cli/add_products.go internal/cli/discover_kyc.go internal/cli/discover_services.go internal/cli/discover_resources.go internal/cli/history.go internal/mocks/data.go > /dev/null
echo "   ✅ internal_cli.zip created successfully ($(unzip -l internal_cli.zip | grep -E '\.(go|sql)$' | wc -l | tr -d ' ') source files)"

echo ""

# Create Zip 3: Root files (≤10 files)
echo "📦 Creating root_files.zip (Root & Config)..."
echo "   Contains: main.go, go.mod, go.sum, sql/init.sql, README.md"
zip -r root_files.zip main.go go.mod go.sum sql/init.sql README.md > /dev/null
echo "   ✅ root_files.zip created successfully ($(unzip -l root_files.zip | grep -E '\.(go|mod|sql)$' | wc -l | tr -d ' ') source files)"

echo ""
echo "🎉 Context zip files refreshed successfully!"
echo ""
echo "Files created (all ≤10 source files each):"
ls -lh *_core.zip *_cli.zip root_files.zip
echo ""
echo "📊 File counts:"
echo "   internal_core.zip: $(unzip -l internal_core.zip | grep -E '\.(go|sql)$' | wc -l | tr -d ' ') source files (store, dsl, agent, dictionary)"
echo "   internal_cli.zip:  $(unzip -l internal_cli.zip | grep -E '\.(go|sql)$' | wc -l | tr -d ' ') source files (cli, mocks)"
echo "   root_files.zip:    $(unzip -l root_files.zip | grep -E '\.(go|mod|sql)$' | wc -l | tr -d ' ') source files (main, config, schema)"
echo ""
echo "These zips contain the complete, correct version of the DSL Onboarding POC project"
echo "and serve as shared context for Claude Code instances."