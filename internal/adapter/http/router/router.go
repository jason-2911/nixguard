// Package router configures the HTTP routing for NixGuard API.
// Follows REST conventions with versioned API paths.
//
// Route structure:
//   /api/v1/firewall/rules
//   /api/v1/network/interfaces
//   /api/v1/vpn/ipsec/tunnels
//   /api/v1/system/health
//   /ws/v1/events          (WebSocket for real-time)
package router

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/nixguard/nixguard/pkg/version"
)

// New creates the main HTTP router with all module routes registered.
func New(mw func(http.Handler) http.Handler, log *slog.Logger) http.Handler {
	mux := http.NewServeMux()

	// ── System routes (always available) ───────────────────────
	mux.HandleFunc("GET /api/v1/health", healthHandler)
	mux.HandleFunc("GET /api/v1/version", versionHandler)

	// ── Module route groups ────────────────────────────────────
	// Each module handler will be registered here during app init.
	// Example:
	//   mux.Handle("/api/v1/firewall/", firewallHandler)
	//   mux.Handle("/api/v1/network/", networkHandler)
	//   mux.Handle("/api/v1/vpn/", vpnHandler)
	//   mux.Handle("/api/v1/dns/", dnsHandler)
	//   mux.Handle("/api/v1/dhcp/", dhcpHandler)
	//   mux.Handle("/api/v1/ids/", idsHandler)
	//   mux.Handle("/api/v1/proxy/", proxyHandler)
	//   mux.Handle("/api/v1/loadbalancer/", lbHandler)
	//   mux.Handle("/api/v1/ha/", haHandler)
	//   mux.Handle("/api/v1/monitor/", monitorHandler)
	//   mux.Handle("/api/v1/auth/", authHandler)
	//   mux.Handle("/api/v1/system/", systemHandler)

	// ── Static files (embedded React build) ────────────────────
	// mux.Handle("/", http.FileServer(http.FS(webFS)))

	// Apply middleware chain
	return mw(mux)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().UTC(),
	})
}

func versionHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{
		"version":    version.Version,
		"build_time": version.BuildTime,
	})
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
