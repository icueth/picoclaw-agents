// Test script for gateway meeting API
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const baseURL = "http://localhost:18790"

func main() {
	fmt.Println("🧪 Testing Gateway Meeting API")
	fmt.Println("================================")
	
	// Wait for server to be ready
	fmt.Println("\n⏳ Waiting for server...")
	if !waitForServer() {
		fmt.Println("❌ Server not available")
		fmt.Println("   Please start the gateway first: ./picoclaw gateway")
		return
	}
	fmt.Println("✅ Server is ready")
	
	// Test 1: Health check
	fmt.Println("\n📋 Test 1: Health Check")
	testHealth()
	
	// Test 2: List agents
	fmt.Println("\n📋 Test 2: List Agents")
	testListAgents()
	
	// Test 3: Get agent detail
	fmt.Println("\n📋 Test 3: Get Agent Detail (Jarvis)")
	testGetAgent("jarvis")
	
	fmt.Println("\n================================")
	fmt.Println("✅ API Tests Complete!")
	fmt.Println("\n📚 Available Endpoints:")
	fmt.Println("   GET  /health              - Health check")
	fmt.Println("   GET  /ready               - Readiness check")
	fmt.Println("   GET  /api/agents          - List all agents")
	fmt.Println("   GET  /api/agents/{id}     - Get agent details")
	fmt.Println("   GET  /api/meetings        - List meetings")
	fmt.Println("   POST /api/meetings        - Create meeting")
	fmt.Println("   GET  /api/meetings/{id}   - Get meeting details")
	fmt.Println("   GET  /api/schedule        - List schedules")
	fmt.Println("   POST /api/schedule        - Schedule meeting")
	fmt.Println("   GET  /api/schedule/upcoming - Get upcoming meetings")
	fmt.Println("   DEL  /api/schedule/{id}   - Cancel schedule")
}

func waitForServer() bool {
	for i := 0; i < 10; i++ {
		resp, err := http.Get(baseURL + "/ready")
		if err == nil && resp.StatusCode == http.StatusOK {
			resp.Body.Close()
			return true
		}
		if resp != nil {
			resp.Body.Close()
		}
		time.Sleep(500 * time.Millisecond)
	}
	return false
}

func testHealth() {
	resp, err := http.Get(baseURL + "/health")
	if err != nil {
		fmt.Printf("   ❌ Error: %v\n", err)
		return
	}
	defer resp.Body.Close()
	
	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("   ✅ Status: %d\n", resp.StatusCode)
	fmt.Printf("   📄 Response: %s\n", string(body))
}

func testListAgents() {
	resp, err := http.Get(baseURL + "/api/agents")
	if err != nil {
		fmt.Printf("   ❌ Error: %v\n", err)
		return
	}
	defer resp.Body.Close()
	
	var agents []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&agents); err != nil {
		fmt.Printf("   ❌ Decode error: %v\n", err)
		return
	}
	
	fmt.Printf("   ✅ Found %d agents\n", len(agents))
	for _, agent := range agents {
		fmt.Printf("   • %s %s (%s)\n", agent["avatar"], agent["name"], agent["role"])
	}
}

func testGetAgent(agentID string) {
	resp, err := http.Get(fmt.Sprintf("%s/api/agents/%s", baseURL, agentID))
	if err != nil {
		fmt.Printf("   ❌ Error: %v\n", err)
		return
	}
	defer resp.Body.Close()
	
	var agent map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&agent); err != nil {
		fmt.Printf("   ❌ Decode error: %v\n", err)
		return
	}
	
	fmt.Printf("   ✅ Agent: %s %s\n", agent["avatar"], agent["name"])
	fmt.Printf("   📋 Role: %s, Dept: %s\n", agent["role"], agent["department"])
}
