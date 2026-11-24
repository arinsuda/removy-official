package auth_module

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"

	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type GoogleAuthorizationParams struct {
	ClientID      string
	RedirectURI   string
	State         string
	Nonce         string
	CodeChallenge string
	Scope         string
}

type GoogleTokenResult struct {
	AccessToken string `json:"access_token"`
	IDToken     string `json:"id_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

type GoogleIdentityClaims struct {
	Subject       string `json:"sub"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
	HostedDomain  string `json:"hd"`
	Nonce         string `json:"nonce"`
	Audience      string `json:"aud"`
	Issuer        string `json:"iss"`
	ExpiresAt     int64  `json:"exp"`
	IssuedAt      int64  `json:"iat"`
}

type GoogleOAuthProvider interface {
	AuthorizationURL(params GoogleAuthorizationParams) string
	ExchangeCode(ctx context.Context, code string, codeVerifier string, clientID string, clientSecret string, redirectURI string) (*GoogleTokenResult, error)
	VerifyIDToken(ctx context.Context, rawIDToken string, clientID string, expectedNonce string) (*GoogleIdentityClaims, error)
}

type RealGoogleOAuthProvider struct {
	httpClient *http.Client
}

func NewRealGoogleOAuthProvider() *RealGoogleOAuthProvider {
	return &RealGoogleOAuthProvider{
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

func (p *RealGoogleOAuthProvider) AuthorizationURL(params GoogleAuthorizationParams) string {
	scope := params.Scope
	if scope == "" {
		scope = "openid email profile"
	}

	query := url.Values{}
	query.Set("client_id", params.ClientID)
	query.Set("redirect_uri", params.RedirectURI)
	query.Set("response_type", "code")
	query.Set("scope", scope)
	query.Set("state", params.State)
	query.Set("nonce", params.Nonce)
	query.Set("code_challenge", params.CodeChallenge)
	query.Set("code_challenge_method", "S256")
	query.Set("prompt", "select_account")

	return "https://accounts.google.com/o/oauth2/v2/auth?" + query.Encode()
}

func (p *RealGoogleOAuthProvider) ExchangeCode(ctx context.Context, code string, codeVerifier string, clientID string, clientSecret string, redirectURI string) (*GoogleTokenResult, error) {
	data := url.Values{}
	data.Set("code", code)
	data.Set("client_id", clientID)
	data.Set("client_secret", clientSecret)
	data.Set("redirect_uri", redirectURI)
	data.Set("grant_type", "authorization_code")
	data.Set("code_verifier", codeVerifier)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://oauth2.googleapis.com/token", strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("google token endpoint request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp struct {
			Error            string `json:"error"`
			ErrorDescription string `json:"error_description"`
		}
		_ = json.NewDecoder(resp.Body).Decode(&errResp)
		return nil, fmt.Errorf("google token endpoint returned status %d: %s (%s)", resp.StatusCode, errResp.Error, errResp.ErrorDescription)
	}

	var tokenResult GoogleTokenResult
	if err := json.NewDecoder(resp.Body).Decode(&tokenResult); err != nil {
		return nil, fmt.Errorf("failed to decode google token response: %w", err)
	}

	if tokenResult.IDToken == "" {
		return nil, errors.New("google response missing id_token")
	}

	return &tokenResult, nil
}

func (p *RealGoogleOAuthProvider) VerifyIDToken(ctx context.Context, rawIDToken string, clientID string, expectedNonce string) (*GoogleIdentityClaims, error) {
	parser := jwt.NewParser(jwt.WithoutClaimsValidation())
	token, _, err := parser.ParseUnverified(rawIDToken, jwt.MapClaims{})
	if err != nil {
		return nil, fmt.Errorf("failed to parse id_token: %w", err)
	}

	claimsMap, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid id_token claims format")
	}

	claims := &GoogleIdentityClaims{}
	if sub, ok := claimsMap["sub"].(string); ok {
		claims.Subject = sub
	}
	if email, ok := claimsMap["email"].(string); ok {
		claims.Email = email
	}
	if ev, ok := claimsMap["email_verified"].(bool); ok {
		claims.EmailVerified = ev
	} else if evStr, ok := claimsMap["email_verified"].(string); ok {
		claims.EmailVerified = (evStr == "true")
	}
	if name, ok := claimsMap["name"].(string); ok {
		claims.Name = name
	}
	if picture, ok := claimsMap["picture"].(string); ok {
		claims.Picture = picture
	}
	if hd, ok := claimsMap["hd"].(string); ok {
		claims.HostedDomain = hd
	}
	if nonce, ok := claimsMap["nonce"].(string); ok {
		claims.Nonce = nonce
	}

	if aud, ok := claimsMap["aud"].(string); ok {
		claims.Audience = aud
	}
	if iss, ok := claimsMap["iss"].(string); ok {
		claims.Issuer = iss
	}
	if exp, ok := claimsMap["exp"].(float64); ok {
		claims.ExpiresAt = int64(exp)
	}
	if iat, ok := claimsMap["iat"].(float64); ok {
		claims.IssuedAt = int64(iat)
	}

	now := time.Now().Unix()

	if claims.Subject == "" {
		return nil, errors.New("missing sub claim in id_token")
	}
	if claims.Email == "" {
		return nil, errors.New("missing email claim in id_token")
	}
	if clientID != "" && claims.Audience != clientID {
		return nil, fmt.Errorf("invalid audience: expected %s, got %s", clientID, claims.Audience)
	}
	if claims.Issuer != "https://accounts.google.com" && claims.Issuer != "accounts.google.com" {
		return nil, fmt.Errorf("invalid issuer: %s", claims.Issuer)
	}
	if claims.ExpiresAt > 0 && now > claims.ExpiresAt {
		return nil, errors.New("id_token is expired")
	}
	if claims.IssuedAt > 0 && now+300 < claims.IssuedAt {
		return nil, errors.New("id_token issued in the future")
	}
	if expectedNonce != "" && claims.Nonce != expectedNonce {
		return nil, fmt.Errorf("nonce mismatch: expected %s, got %s", expectedNonce, claims.Nonce)
	}

	return claims, nil
}

func IsAuthoritativeGoogleEmail(claims *GoogleIdentityClaims) bool {
	normalized := strings.ToLower(strings.TrimSpace(claims.Email))
	if strings.HasSuffix(normalized, "@gmail.com") || strings.HasSuffix(normalized, "@googlemail.com") {
		return true
	}
	return claims.EmailVerified && claims.HostedDomain != ""
}

func SanitizeUsernameBase(input string) string {
	reg := regexp.MustCompile(`[^a-zA-Z0-9_]`)
	cleaned := reg.ReplaceAllString(input, "")
	cleaned = strings.Trim(cleaned, "_")
	if len(cleaned) < 3 {
		cleaned = "user_" + cleaned
	}
	if len(cleaned) > 30 {
		cleaned = cleaned[:30]
	}
	return strings.ToLower(cleaned)
}

func GenerateRandomSuffix(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, length)
	rand.Read(b)
	for i := range b {
		b[i] = charset[int(b[i])%len(charset)]
	}
	return string(b)
}

// Mock implementation for unit tests
type MockGoogleOAuthProvider struct {
	AuthURLFunc     func(params GoogleAuthorizationParams) string
	ExchangeFunc    func(ctx context.Context, code, codeVerifier, clientID, clientSecret, redirectURI string) (*GoogleTokenResult, error)
	VerifyTokenFunc func(ctx context.Context, rawIDToken, clientID, expectedNonce string) (*GoogleIdentityClaims, error)
}

func (m *MockGoogleOAuthProvider) AuthorizationURL(params GoogleAuthorizationParams) string {
	if m.AuthURLFunc != nil {
		return m.AuthURLFunc(params)
	}
	return "https://accounts.google.com/o/oauth2/v2/auth?state=" + params.State
}

func (m *MockGoogleOAuthProvider) ExchangeCode(ctx context.Context, code string, codeVerifier string, clientID string, clientSecret string, redirectURI string) (*GoogleTokenResult, error) {
	if m.ExchangeFunc != nil {
		return m.ExchangeFunc(ctx, code, codeVerifier, clientID, clientSecret, redirectURI)
	}
	return &GoogleTokenResult{
		AccessToken: "mock_access_token",
		IDToken:     "mock_id_token",
	}, nil
}

func (m *MockGoogleOAuthProvider) VerifyIDToken(ctx context.Context, rawIDToken string, clientID string, expectedNonce string) (*GoogleIdentityClaims, error) {
	if m.VerifyTokenFunc != nil {
		return m.VerifyTokenFunc(ctx, rawIDToken, clientID, expectedNonce)
	}
	return &GoogleIdentityClaims{
		Subject:       "mock_google_sub_12345",
		Email:         "mockuser@gmail.com",
		EmailVerified: true,
		Name:          "Mock User",
		Nonce:         expectedNonce,
		Audience:      clientID,
		Issuer:        "https://accounts.google.com",
		ExpiresAt:     time.Now().Add(1 * time.Hour).Unix(),
	}, nil
}
