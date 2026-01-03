// Package auth provides OAuth 2.1 authentication for the MCP server.
package auth

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jwt"
)

// Common errors.
var (
	ErrMissingToken       = errors.New("missing access token")
	ErrInvalidToken       = errors.New("invalid access token")
	ErrTokenExpired       = errors.New("token expired")
	ErrInsufficientScope  = errors.New("insufficient scope")
	ErrInvalidAudience    = errors.New("invalid audience")
	ErrInvalidIssuer      = errors.New("invalid issuer")
)

// Claims represents validated JWT claims.
type Claims struct {
	Subject   string
	Issuer    string
	Audience  []string
	Scopes    []string
	ExpiresAt time.Time
	IssuedAt  time.Time
	Extra     map[string]interface{}
}

// HasScope checks if the claims include a specific scope.
func (c *Claims) HasScope(scope string) bool {
	for _, s := range c.Scopes {
		if s == scope || s == "mcp:*" {
			return true
		}
	}
	return false
}

// Validator validates OAuth tokens.
type Validator interface {
	Validate(ctx context.Context, token string) (*Claims, error)
}

// JWTValidator validates JWT access tokens.
type JWTValidator struct {
	issuer     string
	audience   string
	jwksURL    string
	algorithms []string
	keySet     jwk.Set
	mu         sync.RWMutex
	lastFetch  time.Time
	fetchTTL   time.Duration
}

// JWTValidatorConfig configures the JWT validator.
type JWTValidatorConfig struct {
	Issuer     string
	Audience   string
	JWKSURL    string
	Algorithms []string
}

// NewJWTValidator creates a new JWT validator.
func NewJWTValidator(config JWTValidatorConfig) *JWTValidator {
	if len(config.Algorithms) == 0 {
		config.Algorithms = []string{"RS256", "ES256"}
	}

	jwksURL := config.JWKSURL
	if jwksURL == "" && config.Issuer != "" {
		// Default JWKS URL
		jwksURL = strings.TrimSuffix(config.Issuer, "/") + "/.well-known/jwks.json"
	}

	return &JWTValidator{
		issuer:     config.Issuer,
		audience:   config.Audience,
		jwksURL:    jwksURL,
		algorithms: config.Algorithms,
		fetchTTL:   5 * time.Minute,
	}
}

// getKeySet fetches or returns cached JWKS.
func (v *JWTValidator) getKeySet(ctx context.Context) (jwk.Set, error) {
	v.mu.RLock()
	if v.keySet != nil && time.Since(v.lastFetch) < v.fetchTTL {
		defer v.mu.RUnlock()
		return v.keySet, nil
	}
	v.mu.RUnlock()

	v.mu.Lock()
	defer v.mu.Unlock()

	// Double-check after acquiring write lock
	if v.keySet != nil && time.Since(v.lastFetch) < v.fetchTTL {
		return v.keySet, nil
	}

	keySet, err := jwk.Fetch(ctx, v.jwksURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch JWKS: %w", err)
	}

	v.keySet = keySet
	v.lastFetch = time.Now()
	return keySet, nil
}

// Validate validates a JWT access token.
func (v *JWTValidator) Validate(ctx context.Context, token string) (*Claims, error) {
	keySet, err := v.getKeySet(ctx)
	if err != nil {
		return nil, err
	}

	// Parse and verify the token
	parsed, err := jwt.Parse(
		[]byte(token),
		jwt.WithKeySet(keySet),
		jwt.WithValidate(true),
	)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidToken, err)
	}

	// Validate issuer
	if v.issuer != "" && parsed.Issuer() != v.issuer {
		return nil, ErrInvalidIssuer
	}

	// Validate audience
	if v.audience != "" {
		audiences := parsed.Audience()
		found := false
		for _, aud := range audiences {
			if aud == v.audience {
				found = true
				break
			}
		}
		if !found {
			return nil, ErrInvalidAudience
		}
	}

	// Extract scopes
	var scopes []string
	if scopeClaim, ok := parsed.Get("scope"); ok {
		if scopeStr, ok := scopeClaim.(string); ok {
			scopes = strings.Split(scopeStr, " ")
		}
	}

	return &Claims{
		Subject:   parsed.Subject(),
		Issuer:    parsed.Issuer(),
		Audience:  parsed.Audience(),
		Scopes:    scopes,
		ExpiresAt: parsed.Expiration(),
		IssuedAt:  parsed.IssuedAt(),
		Extra:     make(map[string]interface{}),
	}, nil
}

// NoOpValidator is a validator that allows all requests (for development).
type NoOpValidator struct{}

// Validate always returns valid claims.
func (v *NoOpValidator) Validate(ctx context.Context, token string) (*Claims, error) {
	return &Claims{
		Subject:   "anonymous",
		Scopes:    []string{"mcp:*"},
		ExpiresAt: time.Now().Add(24 * time.Hour),
		IssuedAt:  time.Now(),
	}, nil
}

// APIKeyValidator validates static API keys.
type APIKeyValidator struct {
	apiKey string
}

// NewAPIKeyValidator creates a new API key validator.
func NewAPIKeyValidator(apiKey string) *APIKeyValidator {
	return &APIKeyValidator{apiKey: apiKey}
}

// Validate checks if the token matches the configured API key.
func (v *APIKeyValidator) Validate(ctx context.Context, token string) (*Claims, error) {
	if token != v.apiKey {
		return nil, ErrInvalidToken
	}
	return &Claims{
		Subject:   "api-key-user",
		Scopes:    []string{"mcp:*"},
		ExpiresAt: time.Now().Add(24 * time.Hour),
		IssuedAt:  time.Now(),
	}, nil
}

// contextKey is a custom type for context keys.
type contextKey string

const claimsKey contextKey = "auth_claims"

// ClaimsFromContext extracts claims from context.
func ClaimsFromContext(ctx context.Context) (*Claims, bool) {
	claims, ok := ctx.Value(claimsKey).(*Claims)
	return claims, ok
}

// Middleware returns HTTP middleware for OAuth authentication.
func Middleware(validator Validator, requiredScopes []string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := extractBearerToken(r)
			if token == "" {
				unauthorized(w, "missing token")
				return
			}

			claims, err := validator.Validate(r.Context(), token)
			if err != nil {
				if errors.Is(err, ErrTokenExpired) {
					unauthorized(w, "token expired")
				} else if errors.Is(err, ErrInvalidAudience) {
					forbidden(w, "invalid audience")
				} else if errors.Is(err, ErrInvalidIssuer) {
					forbidden(w, "invalid issuer")
				} else {
					unauthorized(w, "invalid token")
				}
				return
			}

			// Check required scopes
			for _, scope := range requiredScopes {
				if !claims.HasScope(scope) {
					forbidden(w, fmt.Sprintf("missing scope: %s", scope))
					return
				}
			}

			// Add claims to context
			ctx := context.WithValue(r.Context(), claimsKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// extractBearerToken extracts the bearer token from the Authorization header.
func extractBearerToken(r *http.Request) string {
	auth := r.Header.Get("Authorization")
	if auth == "" {
		return ""
	}

	parts := strings.SplitN(auth, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
		return ""
	}

	return parts[1]
}

// unauthorized sends a 401 response.
func unauthorized(w http.ResponseWriter, message string) {
	w.Header().Set("WWW-Authenticate", `Bearer realm="mcp-calculator"`)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	fmt.Fprintf(w, `{"jsonrpc":"2.0","id":null,"error":{"code":-32001,"message":"Unauthorized: %s"}}`, message)
}

// forbidden sends a 403 response.
func forbidden(w http.ResponseWriter, message string) {
	w.Header().Set("WWW-Authenticate", `Bearer realm="mcp-calculator", error="insufficient_scope"`)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusForbidden)
	fmt.Fprintf(w, `{"jsonrpc":"2.0","id":null,"error":{"code":-32002,"message":"Forbidden: %s"}}`, message)
}

// Config holds OAuth configuration.
type Config struct {
	Enabled   bool
	Issuer    string
	Audience  string
	JWKSURL   string
	Scopes    []string
	Algorithms []string
}

// NewValidatorFromConfig creates a validator from configuration.
func NewValidatorFromConfig(cfg Config) Validator {
	if !cfg.Enabled {
		return &NoOpValidator{}
	}

	return NewJWTValidator(JWTValidatorConfig{
		Issuer:     cfg.Issuer,
		Audience:   cfg.Audience,
		JWKSURL:    cfg.JWKSURL,
		Algorithms: cfg.Algorithms,
	})
}
