#!/bin/bash
#
# Integration test script for Agent Team system
# Run: ./scripts/test_agent_team.sh

set -e

echo "🧪 PicoClaw Agent Team Test Suite"
echo "=================================="

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test counter
TESTS_PASSED=0
TESTS_FAILED=0

# Function to run test
run_test() {
    local name=$1
    local command=$2
    
    echo -n "Testing $name... "
    if eval "$command" > /tmp/test_output.log 2>&1; then
        echo -e "${GREEN}✓ PASSED${NC}"
        ((TESTS_PASSED++))
    else
        echo -e "${RED}✗ FAILED${NC}"
        echo "  Output: $(cat /tmp/test_output.log | head -5)"
        ((TESTS_FAILED++))
    fi
}

echo ""
echo "📦 Backend Tests (Go)"
echo "---------------------"

# Go unit tests
run_test "Mailbox System" "go test -v ./pkg/mailbox/... -run TestMailbox"
run_test "Mailbox Hub" "go test -v ./pkg/mailbox/... -run TestHub"
run_test "Agent Loop" "go test -v ./pkg/agents/... -run TestAgentLoop_StartStop"
run_test "Task Analysis" "go test -v ./pkg/agents/... -run TestAnalyzeTask"
run_test "Task Queue" "go test -v ./pkg/agents/... -run TestTaskQueue"
run_test "Coordinator" "go test -v ./pkg/agents/... -run TestJarvisCoordinator_StartStop"

echo ""
echo "🔧 Build Tests"
echo "--------------"

run_test "Build Agents Package" "go build ./pkg/agents/..."
run_test "Build Mailbox Package" "go build ./pkg/mailbox/..."
run_test "Build Memory Package" "go build ./pkg/memory/..."
run_test "Build API/UI Package" "go build ./pkg/api/ui/..."
run_test "Build Full Project" "go build ./..."

echo ""
echo "🌐 Frontend Tests"
echo "-----------------"

cd ui

run_test "TypeScript Type Check" "npm run typecheck"
run_test "Build UI" "npm run build"

cd ..

echo ""
echo "📊 Test Summary"
echo "---------------"
echo -e "Tests Passed: ${GREEN}$TESTS_PASSED${NC}"
echo -e "Tests Failed: ${RED}$TESTS_FAILED${NC}"

if [ $TESTS_FAILED -eq 0 ]; then
    echo ""
    echo -e "${GREEN}🎉 All tests passed!${NC}"
    exit 0
else
    echo ""
    echo -e "${RED}⚠️  Some tests failed!${NC}"
    exit 1
fi
