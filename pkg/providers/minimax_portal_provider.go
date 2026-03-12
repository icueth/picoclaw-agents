package providers

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"picoclaw/agent/pkg/auth"
	"picoclaw/agent/pkg/logger"
	anthropicprovider "picoclaw/agent/pkg/providers/anthropic"
)

const (
	minimaxPortalBaseURLGlobal = "https://api.minimax.io/anthropic"
	minimaxPortalBaseURLCN     = "https://api.minimaxi.com/anthropic"
	minimaxPortalDefaultModel  = "MiniMax-M2.5"
	minimaxPortalClientID      = "78257093-7e40-4613-99e0-527b14b39113"
	minimaxPortalGrantType     = "urn:ietf:params:oauth:grant-type:user_code"
	minimaxPortalScope         = "group_id profile model.completion"
)

// MiniMaxPortalProvider implements LLMProvider using MiniMax Portal API (Anthropic-compatible).
// It authenticates via MiniMax's custom OAuth device-code flow.
type MiniMaxPortalProvider struct {
	delegate *anthropicprovider.Provider
}

// NewMiniMaxPortalProvider creates a new MiniMaxPortal provider using stored OAuth credentials.
func NewMiniMaxPortalProvider() (*MiniMaxPortalProvider, error) {
	tokenSource, apiBase, err := createMiniMaxTokenSource()
	if err != nil {
		return nil, err
	}

	cred, _ := auth.GetCredential("minimax-portal")
	initialToken := ""
	if cred != nil {
		initialToken = cred.AccessToken
	}

	delegate := anthropicprovider.NewProviderWithTokenSourceAndBaseURL(initialToken, tokenSource, apiBase)
	return &MiniMaxPortalProvider{delegate: delegate}, nil
}

// Chat implements LLMProvider.Chat.
func (p *MiniMaxPortalProvider) Chat(
	ctx context.Context,
	messages []Message,
	tools []ToolDefinition,
	model string,
	options map[string]any,
) (*LLMResponse, error) {
	resp, err := p.delegate.Chat(ctx, messages, tools, model, options)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// GetDefaultModel returns the default MiniMax model.
func (p *MiniMaxPortalProvider) GetDefaultModel() string {
	return minimaxPortalDefaultModel
}

// createMiniMaxTokenSource creates a token source function that auto-refreshes MiniMax OAuth tokens.
// Returns the token source function, the API base URL, and any error.
func createMiniMaxTokenSource() (func() (string, error), string, error) {
	cred, err := auth.GetCredential("minimax-portal")
	if err != nil {
		return nil, "", fmt.Errorf("loading minimax-portal credentials: %w", err)
	}
	if cred == nil {
		return nil, "", fmt.Errorf(
			"no credentials for minimax-portal. Run: picoclaw auth login --provider minimax-portal",
		)
	}

	// Determine API base URL from stored credential
	apiBase := minimaxPortalBaseURLGlobal
	if cred.ProjectID == "cn" { // We store region in ProjectID field
		apiBase = minimaxPortalBaseURLCN
	}

	tokenSource := func() (string, error) {
		cred, err := auth.GetCredential("minimax-portal")
		if err != nil {
			return "", fmt.Errorf("loading minimax-portal credentials: %w", err)
		}
		if cred == nil {
			return "", fmt.Errorf(
				"no credentials for minimax-portal. Run: picoclaw auth login --provider minimax-portal",
			)
		}

		// Refresh if nearing expiry
		if cred.NeedsRefresh() && cred.RefreshToken != "" {
			region := auth.MiniMaxPortalRegionGlobal
			if cred.ProjectID == "cn" {
				region = auth.MiniMaxPortalRegionCN
			}
			cfg := auth.MiniMaxPortalOAuthConfig(region)
			refreshed, err := refreshMiniMaxToken(cred.RefreshToken, cfg.Issuer)
			if err != nil {
				logger.WarnCF("provider.minimax-portal", "Failed to refresh token, using existing", map[string]any{
					"error": err.Error(),
				})
				// Fall through to use existing token if not expired
			} else {
				refreshed.Provider = "minimax-portal"
				refreshed.ProjectID = cred.ProjectID
				if err := auth.SetCredential("minimax-portal", refreshed); err != nil {
					logger.WarnCF("provider.minimax-portal", "Failed to save refreshed token", map[string]any{
						"error": err.Error(),
					})
				}
				cred = refreshed
			}
		}

		if cred.IsExpired() {
			return "", fmt.Errorf(
				"minimax-portal credentials expired. Run: picoclaw auth login --provider minimax-portal",
			)
		}

		return cred.AccessToken, nil
	}

	return tokenSource, apiBase, nil
}

// MiniMaxDeviceCodeInfo holds device code authorization info from MiniMax.
type MiniMaxDeviceCodeInfo struct {
	UserCode        string `json:"user_code"`
	VerificationURI string `json:"verification_uri"`
	ExpiredIn       int64  `json:"expired_in"`
	Interval        int    `json:"interval"`
	State           string `json:"state"`
}

// RequestMiniMaxDeviceCode requests a device code from MiniMax OAuth endpoint.
func RequestMiniMaxDeviceCode(cfg auth.OAuthProviderConfig) (*MiniMaxDeviceCodeInfo, error) {
	// Generate PKCE
	verifierBytes := make([]byte, 32)
	if _, err := rand.Read(verifierBytes); err != nil {
		return nil, fmt.Errorf("generating PKCE verifier: %w", err)
	}
	verifier := base64.RawURLEncoding.EncodeToString(verifierBytes)
	challengeHash := sha256.Sum256([]byte(verifier))
	challenge := base64.RawURLEncoding.EncodeToString(challengeHash[:])

	// Generate state
	stateBytes := make([]byte, 16)
	if _, err := rand.Read(stateBytes); err != nil {
		return nil, fmt.Errorf("generating state: %w", err)
	}
	state := base64.RawURLEncoding.EncodeToString(stateBytes)

	formData := url.Values{
		"response_type":         {"code"},
		"client_id":             {cfg.ClientID},
		"scope":                 {cfg.Scopes},
		"code_challenge":        {challenge},
		"code_challenge_method": {"S256"},
		"state":                 {state},
	}

	codeEndpoint := cfg.Issuer + "/oauth/code"
	resp, err := http.Post(codeEndpoint, "application/x-www-form-urlencoded", strings.NewReader(formData.Encode()))
	if err != nil {
		return nil, fmt.Errorf("requesting device code: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("device code request failed (HTTP %d): %s", resp.StatusCode, string(body))
	}

	var result MiniMaxDeviceCodeInfo
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parsing device code response: %w", err)
	}

	if result.UserCode == "" || result.VerificationURI == "" {
		return nil, fmt.Errorf("incomplete device code response from MiniMax: %s", string(body))
	}

	if result.State != state {
		return nil, fmt.Errorf("state mismatch in MiniMax OAuth response: possible CSRF")
	}

	// Store verifier in the info so the caller can use it when polling
	result.State = verifier // reuse State field to carry verifier

	return &result, nil
}

// PollMiniMaxToken polls MiniMax OAuth token endpoint for authorization.
// Returns (credential, nil) on success, (nil, nil) if still pending, or (nil, err) on failure.
func PollMiniMaxToken(cfg auth.OAuthProviderConfig, userCode, codeVerifier string) (*auth.AuthCredential, error) {
	formData := url.Values{
		"grant_type":    {minimaxPortalGrantType},
		"client_id":     {cfg.ClientID},
		"user_code":     {userCode},
		"code_verifier": {codeVerifier},
	}

	tokenEndpoint := cfg.Issuer + "/oauth/token"
	resp, err := http.Post(tokenEndpoint, "application/x-www-form-urlencoded", strings.NewReader(formData.Encode()))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var payload struct {
		Status              string  `json:"status"`
		AccessToken         string  `json:"access_token"`
		RefreshToken        string  `json:"refresh_token"`
		ExpiredIn           float64 `json:"expired_in"`
		ResourceURL         string  `json:"resource_url"`
		NotificationMessage string  `json:"notification_message"`
		BaseResp            *struct {
			StatusCode int    `json:"status_code"`
			StatusMsg  string `json:"status_msg"`
		} `json:"base_resp"`
	}

	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, fmt.Errorf("parsing token response: %w", err)
	}

	if resp.StatusCode/100 != 2 {
		msg := string(body)
		if payload.BaseResp != nil && payload.BaseResp.StatusMsg != "" {
			msg = payload.BaseResp.StatusMsg
		}
		return nil, fmt.Errorf("token request failed: %s", msg)
	}

	if payload.Status == "error" {
		return nil, fmt.Errorf("MiniMax OAuth error: please try again")
	}

	if payload.Status != "success" {
		// Still pending
		return nil, nil
	}

	if payload.AccessToken == "" || payload.RefreshToken == "" {
		return nil, fmt.Errorf("incomplete token response from MiniMax")
	}

	var expiresAt time.Time
	if payload.ExpiredIn > 0 {
		if payload.ExpiredIn > 1e11 { // 1e11 is year 5138 in seconds, so it's definitely milliseconds if greater
			expiresAt = time.UnixMilli(int64(payload.ExpiredIn))
		} else {
			expiresAt = time.Unix(int64(payload.ExpiredIn), 0)
		}
	}

	cred := &auth.AuthCredential{
		AccessToken:  payload.AccessToken,
		RefreshToken: payload.RefreshToken,
		ExpiresAt:    expiresAt,
		Provider:     "minimax-portal",
		AuthMethod:   "oauth",
	}

	return cred, nil
}

// LoginMiniMaxPortal performs the full MiniMax device-code OAuth flow.
func LoginMiniMaxPortal(cfg auth.OAuthProviderConfig) (*auth.AuthCredential, error) {
	info, err := RequestMiniMaxDeviceCode(cfg)
	if err != nil {
		return nil, fmt.Errorf("requesting device code: %w", err)
	}

	codeVerifier := info.State // We stored verifier in State field

	fmt.Printf("\nTo authenticate with MiniMax Portal:\n")
	fmt.Printf("  1. Open: %s\n", info.VerificationURI)
	fmt.Printf("  2. Enter code: %s\n\n", info.UserCode)
	fmt.Printf("Waiting for authorization...\n")

	interval := time.Duration(info.Interval) * time.Millisecond
	if interval < 2*time.Second {
		interval = 2 * time.Second
	}

	var deadline time.Time
	if info.ExpiredIn > 0 {
		if info.ExpiredIn > 1e11 {
			deadline = time.UnixMilli(info.ExpiredIn)
		} else {
			deadline = time.Unix(info.ExpiredIn, 0)
		}
	} else {
		deadline = time.Now().Add(15 * time.Minute)
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		if time.Now().After(deadline) {
			return nil, fmt.Errorf("MiniMax OAuth timed out waiting for authorization")
		}
		cred, err := PollMiniMaxToken(cfg, info.UserCode, codeVerifier)
		if err != nil {
			return nil, err
		}
		if cred != nil {
			return cred, nil
		}
		// Gradually back off
		if interval < 10*time.Second {
			interval = time.Duration(float64(interval) * 1.5)
			ticker.Reset(interval)
		}
	}

	return nil, fmt.Errorf("MiniMax OAuth ticker closed unexpectedly")
}

// refreshMiniMaxToken refreshes a MiniMax OAuth access token using the refresh token.
func refreshMiniMaxToken(refreshToken, baseURL string) (*auth.AuthCredential, error) {
	formData := url.Values{
		"grant_type":    {"refresh_token"},
		"client_id":     {minimaxPortalClientID},
		"refresh_token": {refreshToken},
	}

	tokenEndpoint := baseURL + "/oauth/token"
	resp, err := http.Post(tokenEndpoint, "application/x-www-form-urlencoded", strings.NewReader(formData.Encode()))
	if err != nil {
		return nil, fmt.Errorf("refreshing token: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token refresh failed (HTTP %d): %s", resp.StatusCode, string(body))
	}

	var payload struct {
		Status       string  `json:"status"`
		AccessToken  string  `json:"access_token"`
		RefreshToken string  `json:"refresh_token"`
		ExpiredIn    float64 `json:"expired_in"`
	}

	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, fmt.Errorf("parsing refresh response: %w", err)
	}

	if payload.Status != "success" || payload.AccessToken == "" {
		return nil, fmt.Errorf("token refresh failed: status=%s", payload.Status)
	}

	var expiresAt time.Time
	if payload.ExpiredIn > 0 {
		if payload.ExpiredIn > 1e11 {
			expiresAt = time.UnixMilli(int64(payload.ExpiredIn))
		} else {
			expiresAt = time.Unix(int64(payload.ExpiredIn), 0)
		}
	}

	newRefresh := payload.RefreshToken
	if newRefresh == "" {
		newRefresh = refreshToken // keep old refresh if not rotated
	}

	return &auth.AuthCredential{
		AccessToken:  payload.AccessToken,
		RefreshToken: newRefresh,
		ExpiresAt:    expiresAt,
		Provider:     "minimax-portal",
		AuthMethod:   "oauth",
	}, nil
}
