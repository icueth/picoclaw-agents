// Test OAuth Authentication
// ทดสอบการ authenticate กับ providers ต่างๆ

package main

import (
	"fmt"
	"os"

	"picoclaw/agent/pkg/auth"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: test-auth <provider>")
		fmt.Println("Providers: google-antigravity, minimax-portal")
		os.Exit(1)
	}

	provider := os.Args[1]

	switch provider {
	case "google-antigravity":
		fmt.Println("🔐 Authenticating with Google Cloud Code Assist (Antigravity)...")
		fmt.Println("Scopes: cloud-platform, userinfo, cclog, experimentsandconfigs, generative-language")
		fmt.Println("\n🌐 Opening browser for authentication...")
		
		cfg := auth.GoogleAntigravityOAuthConfig()
		cred, err := auth.LoginBrowser(cfg)
		if err != nil {
			fmt.Printf("❌ Authentication failed: %v\n", err)
			os.Exit(1)
		}
		
		// Save credential
		if err := auth.SetCredential("google-antigravity", cred); err != nil {
			fmt.Printf("❌ Failed to save credential: %v\n", err)
			os.Exit(1)
		}
		
		fmt.Printf("✅ Authentication successful!\n")
		fmt.Printf("   Email: %s\n", cred.Email)
		fmt.Printf("   Token: %s...\n", cred.AccessToken[:30])
		fmt.Printf("   Project ID: %s\n", cred.ProjectID)
		fmt.Printf("   Expires: %s\n", cred.ExpiresAt)
		
	case "minimax-portal":
		fmt.Println("🔐 Authenticating with MiniMax Portal...")
		cfg := auth.MiniMaxPortalOAuthConfig(auth.MiniMaxPortalRegionGlobal)
		cred, err := auth.LoginDeviceCode(cfg)
		if err != nil {
			fmt.Printf("❌ Authentication failed: %v\n", err)
			os.Exit(1)
		}
		
		// Save credential
		if err := auth.SetCredential("minimax-portal", cred); err != nil {
			fmt.Printf("❌ Failed to save credential: %v\n", err)
			os.Exit(1)
		}
		
		fmt.Printf("✅ Authentication successful!\n")
		fmt.Printf("   Token: %s...\n", cred.AccessToken[:30])
		
	default:
		fmt.Printf("❌ Unknown provider: %s\n", provider)
		os.Exit(1)
	}
}
