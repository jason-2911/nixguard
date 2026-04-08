// Package network provides HTTP handlers for the network API.
// All endpoints are under /api/v1/network/
package network

import (
	"encoding/json"
	"net/http"

	netApp "github.com/nixguard/nixguard/internal/app/network"
)

type Handler struct {
	svc *netApp.Service
}

func NewHandler(svc *netApp.Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	// Interfaces
	mux.HandleFunc("GET /api/v1/network/interfaces", h.listInterfaces)
	mux.HandleFunc("GET /api/v1/network/interfaces/{id}", h.getInterface)
	mux.HandleFunc("POST /api/v1/network/interfaces", h.createInterface)
	mux.HandleFunc("PUT /api/v1/network/interfaces/{id}", h.updateInterface)
	mux.HandleFunc("DELETE /api/v1/network/interfaces/{id}", h.deleteInterface)

	// Routes
	mux.HandleFunc("GET /api/v1/network/routes", h.listRoutes)
	mux.HandleFunc("POST /api/v1/network/routes", h.createRoute)
	mux.HandleFunc("DELETE /api/v1/network/routes/{id}", h.deleteRoute)
	mux.HandleFunc("GET /api/v1/network/routes/system", h.getSystemRoutes)

	// Gateways
	mux.HandleFunc("GET /api/v1/network/gateways", h.listGateways)
	mux.HandleFunc("POST /api/v1/network/gateways", h.createGateway)
	mux.HandleFunc("DELETE /api/v1/network/gateways/{id}", h.deleteGateway)
	mux.HandleFunc("GET /api/v1/network/gateway-groups", h.listGatewayGroups)
	mux.HandleFunc("POST /api/v1/network/gateway-groups", h.createGatewayGroup)
	mux.HandleFunc("DELETE /api/v1/network/gateway-groups/{id}", h.deleteGatewayGroup)
}

// ── Interfaces ─────────────────────────────────────────────────

func (h *Handler) listInterfaces(w http.ResponseWriter, r *http.Request) {
	ifaces, err := h.svc.ListInterfaces(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, ifaces)
}

func (h *Handler) getInterface(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	iface, err := h.svc.GetInterface(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, iface)
}

func (h *Handler) createInterface(w http.ResponseWriter, r *http.Request) {
	var input netApp.CreateInterfaceInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	iface, err := h.svc.CreateInterface(r.Context(), input)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, iface)
}

func (h *Handler) updateInterface(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var input netApp.UpdateInterfaceInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	iface, err := h.svc.UpdateInterface(r.Context(), id, input)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, iface)
}

func (h *Handler) deleteInterface(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := h.svc.DeleteInterface(r.Context(), id); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ── Routes ─────────────────────────────────────────────────────

func (h *Handler) listRoutes(w http.ResponseWriter, r *http.Request) {
	routes, err := h.svc.ListRoutes(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, routes)
}

func (h *Handler) createRoute(w http.ResponseWriter, r *http.Request) {
	var input netApp.CreateRouteInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	route, err := h.svc.CreateRoute(r.Context(), input)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, route)
}

func (h *Handler) deleteRoute(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := h.svc.DeleteRoute(r.Context(), id); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) getSystemRoutes(w http.ResponseWriter, r *http.Request) {
	routes, err := h.svc.GetSystemRoutes(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, routes)
}

// ── Gateways ───────────────────────────────────────────────────

func (h *Handler) listGateways(w http.ResponseWriter, r *http.Request) {
	gateways, err := h.svc.ListGateways(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, gateways)
}

func (h *Handler) createGateway(w http.ResponseWriter, r *http.Request) {
	var input netApp.CreateGatewayInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	gw, err := h.svc.CreateGateway(r.Context(), input)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, gw)
}

func (h *Handler) deleteGateway(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := h.svc.DeleteGateway(r.Context(), id); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) listGatewayGroups(w http.ResponseWriter, r *http.Request) {
	groups, err := h.svc.ListGatewayGroups(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, groups)
}

func (h *Handler) createGatewayGroup(w http.ResponseWriter, r *http.Request) {
	var input netApp.CreateGatewayGroupInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	group, err := h.svc.CreateGatewayGroup(r.Context(), input)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, group)
}

func (h *Handler) deleteGatewayGroup(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := h.svc.DeleteGatewayGroup(r.Context(), id); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
