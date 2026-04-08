// Package middleware provides HTTP middleware for the NixGuard API.
package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"github.com/nixguard/nixguard/internal/config"
	"github.com/nixguard/nixguard/pkg/crypto"
)

// NewChain composes multiple middleware into a single wrapper.
func NewChain(middlewares ...func(http.Handler) http.Handler) func(http.Handler) http.Handler {
	return func(final http.Handler) http.Handler {
		for i := len(middlewares) - 1; i >= 0; i-- {
			final = middlewares[i](final)
		}
		return final
	}
}

// Recovery catches panics and returns 500.
func Recovery(log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					log.Error("panic recovered",
						slog.Any("error", err),
						slog.String("stack", string(debug.Stack())),
					)
					http.Error(w, "internal server error", http.StatusInternalServerError)
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}

// RequestID adds a unique request ID to each request context.
func RequestID() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			id := r.Header.Get("X-Request-ID")
			if id == "" {
				id = crypto.RandomID(16)
			}
			ctx := context.WithValue(r.Context(), ctxKeyRequestID, id)
			w.Header().Set("X-Request-ID", id)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// Logger logs HTTP requests with timing.
func Logger(log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			sw := &statusWriter{ResponseWriter: w, status: 200}

			next.ServeHTTP(sw, r)

			log.Info("http request",
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.Int("status", sw.status),
				slog.Duration("duration", time.Since(start)),
				slog.String("remote", r.RemoteAddr),
				slog.String("request_id", GetRequestID(r.Context())),
			)
		})
	}
}

// CORS handles Cross-Origin Resource Sharing.
func CORS(origins []string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			allowed := false
			for _, o := range origins {
				if o == "*" || o == origin {
					allowed = true
					break
				}
			}
			if allowed {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, PATCH, OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Request-ID")
				w.Header().Set("Access-Control-Max-Age", "86400")
			}
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// RateLimit implements a simple token-bucket rate limiter per IP.
func RateLimit(requestsPerSecond int) func(http.Handler) http.Handler {
	type client struct {
		tokens    float64
		lastCheck time.Time
	}
	var (
		mu      sync.Mutex
		clients = make(map[string]*client)
	)
	rate := float64(requestsPerSecond)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := strings.Split(r.RemoteAddr, ":")[0]

			mu.Lock()
			c, exists := clients[ip]
			if !exists {
				c = &client{tokens: rate, lastCheck: time.Now()}
				clients[ip] = c
			}
			elapsed := time.Since(c.lastCheck).Seconds()
			c.tokens += elapsed * rate
			if c.tokens > rate {
				c.tokens = rate
			}
			c.lastCheck = time.Now()
			if c.tokens < 1 {
				mu.Unlock()
				http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
				return
			}
			c.tokens--
			mu.Unlock()

			next.ServeHTTP(w, r)
		})
	}
}

// Auth validates JWT tokens and populates the request context with user info.
func Auth(cfg config.AuthConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip auth for health/version endpoints
			if r.URL.Path == "/api/v1/health" || r.URL.Path == "/api/v1/version" {
				next.ServeHTTP(w, r)
				return
			}
			// Skip auth for login endpoint
			if r.URL.Path == "/api/v1/auth/login" && r.Method == http.MethodPost {
				next.ServeHTTP(w, r)
				return
			}

			authHeader := r.Header.Get("Authorization")
			if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			// TODO: Validate JWT token and extract claims
			// token := strings.TrimPrefix(authHeader, "Bearer ")
			// claims, err := validateToken(token, cfg.JWTSecret)

			next.ServeHTTP(w, r)
		})
	}
}

// ── Context helpers ────────────────────────────────────────────

type contextKey string

const ctxKeyRequestID contextKey = "request_id"

// GetRequestID extracts the request ID from context.
func GetRequestID(ctx context.Context) string {
	if id, ok := ctx.Value(ctxKeyRequestID).(string); ok {
		return id
	}
	return ""
}

// statusWriter wraps ResponseWriter to capture status code.
type statusWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}
