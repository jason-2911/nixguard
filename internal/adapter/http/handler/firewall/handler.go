// Package firewall provides HTTP handlers for the firewall API.
// All endpoints are under /api/v1/firewall/
package firewall

import (
	"encoding/json"
	"net/http"
	"path/filepath"
	"strconv"

	fwApp "github.com/nixguard/nixguard/internal/app/firewall"
	fwDomain "github.com/nixguard/nixguard/internal/domain/firewall"
)

// Handler handles firewall HTTP requests.
type Handler struct {
	svc *fwApp.Service
}

func NewHandler(svc *fwApp.Service) *Handler {
	return &Handler{svc: svc}
}

// RegisterRoutes registers all firewall routes on the given mux.
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	// Rules
	mux.HandleFunc("GET /api/v1/firewall/rules", h.listRules)
	mux.HandleFunc("GET /api/v1/firewall/rules/{id}", h.getRule)
	mux.HandleFunc("POST /api/v1/firewall/rules", h.createRule)
	mux.HandleFunc("PUT /api/v1/firewall/rules/{id}", h.updateRule)
	mux.HandleFunc("DELETE /api/v1/firewall/rules/{id}", h.deleteRule)
	mux.HandleFunc("POST /api/v1/firewall/rules/reorder", h.reorderRules)

	// Aliases
	mux.HandleFunc("GET /api/v1/firewall/aliases", h.listAliases)
	mux.HandleFunc("GET /api/v1/firewall/aliases/{id}", h.getAlias)
	mux.HandleFunc("POST /api/v1/firewall/aliases", h.createAlias)
	mux.HandleFunc("PUT /api/v1/firewall/aliases/{id}", h.updateAlias)
	mux.HandleFunc("DELETE /api/v1/firewall/aliases/{id}", h.deleteAlias)

	// NAT
	mux.HandleFunc("GET /api/v1/firewall/nat", h.listNATRules)
	mux.HandleFunc("GET /api/v1/firewall/nat/{id}", h.getNATRule)
	mux.HandleFunc("POST /api/v1/firewall/nat", h.createNATRule)
	mux.HandleFunc("PUT /api/v1/firewall/nat/{id}", h.updateNATRule)
	mux.HandleFunc("DELETE /api/v1/firewall/nat/{id}", h.deleteNATRule)

	// States
	mux.HandleFunc("GET /api/v1/firewall/states", h.listStates)
	mux.HandleFunc("DELETE /api/v1/firewall/states", h.flushStates)

	// Live traffic
	mux.HandleFunc("GET /api/v1/firewall/traffic", h.captureTraffic)
	mux.HandleFunc("POST /api/v1/firewall/traffic/export", h.exportTraffic)
	mux.HandleFunc("GET /api/v1/firewall/traffic/export/{name}", h.downloadTrafficExport)

	// Apply
	mux.HandleFunc("POST /api/v1/firewall/apply", h.applyRules)
}

// ── Rules ──────────────────────────────────────────────────────

func (h *Handler) listRules(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	filter := fwDomain.RuleFilter{
		Interface: q.Get("interface"),
		Category:  q.Get("category"),
		Search:    q.Get("search"),
	}
	if q.Get("action") != "" {
		filter.Action = fwDomain.Action(q.Get("action"))
	}
	if q.Get("protocol") != "" {
		filter.Protocol = fwDomain.Protocol(q.Get("protocol"))
	}
	if limit, err := strconv.Atoi(q.Get("limit")); err == nil {
		filter.Limit = limit
	}
	if offset, err := strconv.Atoi(q.Get("offset")); err == nil {
		filter.Offset = offset
	}

	rules, err := h.svc.ListRules(r.Context(), filter)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if rules == nil {
		rules = []fwDomain.Rule{}
	}
	writeJSON(w, http.StatusOK, rules)
}

func (h *Handler) getRule(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	rule, err := h.svc.GetRule(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, rule)
}

func (h *Handler) createRule(w http.ResponseWriter, r *http.Request) {
	var input fwApp.CreateRuleInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	rule, err := h.svc.CreateRule(r.Context(), input)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, rule)
}

func (h *Handler) updateRule(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var input fwApp.UpdateRuleInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	rule, err := h.svc.UpdateRule(r.Context(), id, input)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, rule)
}

func (h *Handler) deleteRule(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := h.svc.DeleteRule(r.Context(), id); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) reorderRules(w http.ResponseWriter, r *http.Request) {
	var body struct {
		RuleIDs []string `json:"rule_ids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := h.svc.ReorderRules(r.Context(), body.RuleIDs); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// ── Aliases ────────────────────────────────────────────────────

func (h *Handler) listAliases(w http.ResponseWriter, r *http.Request) {
	aliases, err := h.svc.ListAliases(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if aliases == nil {
		aliases = []fwDomain.Alias{}
	}
	writeJSON(w, http.StatusOK, aliases)
}

func (h *Handler) createAlias(w http.ResponseWriter, r *http.Request) {
	var input fwApp.CreateAliasInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	alias, err := h.svc.CreateAlias(r.Context(), input)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, alias)
}

func (h *Handler) getAlias(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	alias, err := h.svc.GetAlias(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, alias)
}

func (h *Handler) updateAlias(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var input fwApp.UpdateAliasInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	alias, err := h.svc.UpdateAlias(r.Context(), id, input)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, alias)
}

func (h *Handler) deleteAlias(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := h.svc.DeleteAlias(r.Context(), id); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ── NAT ────────────────────────────────────────────────────────

func (h *Handler) listNATRules(w http.ResponseWriter, r *http.Request) {
	natType := fwDomain.NATType(r.URL.Query().Get("type"))
	rules, err := h.svc.ListNATRules(r.Context(), natType)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if rules == nil {
		rules = []fwDomain.NATRule{}
	}
	writeJSON(w, http.StatusOK, rules)
}

func (h *Handler) createNATRule(w http.ResponseWriter, r *http.Request) {
	var input fwApp.CreateNATInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	rule, err := h.svc.CreateNATRule(r.Context(), input)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, rule)
}

func (h *Handler) getNATRule(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	rule, err := h.svc.GetNATRule(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, rule)
}

func (h *Handler) updateNATRule(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var input fwApp.UpdateNATInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	rule, err := h.svc.UpdateNATRule(r.Context(), id, input)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, rule)
}

func (h *Handler) deleteNATRule(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := h.svc.DeleteNATRule(r.Context(), id); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ── States ─────────────────────────────────────────────────────

func (h *Handler) listStates(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	filter := fwDomain.StateFilter{
		Protocol: q.Get("protocol"),
		SourceIP: q.Get("src"),
		DestIP:   q.Get("dst"),
	}
	if limit, err := strconv.Atoi(q.Get("limit")); err == nil {
		filter.Limit = limit
	}

	states, err := h.svc.GetStates(r.Context(), filter)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if states == nil {
		states = []fwDomain.State{}
	}
	writeJSON(w, http.StatusOK, states)
}

func (h *Handler) flushStates(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	filter := fwDomain.StateFilter{
		Protocol: q.Get("protocol"),
		SourceIP: q.Get("src"),
		DestIP:   q.Get("dst"),
	}
	if err := h.svc.FlushStates(r.Context(), filter); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "flushed"})
}

func (h *Handler) captureTraffic(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	input := fwApp.TrafficCaptureInput{
		Interface: q.Get("interface"),
		SourceIP:  q.Get("src"),
		DestIP:    q.Get("dst"),
		Protocol:  q.Get("protocol"),
	}
	if count, err := strconv.Atoi(q.Get("count")); err == nil {
		input.Count = count
	}
	if snapLen, err := strconv.Atoi(q.Get("snap_len")); err == nil {
		input.SnapLen = snapLen
	}

	packets, err := h.svc.CaptureTraffic(r.Context(), input)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if packets == nil {
		packets = []fwDomain.CapturedPacket{}
	}
	writeJSON(w, http.StatusOK, packets)
}

func (h *Handler) exportTraffic(w http.ResponseWriter, r *http.Request) {
	var input fwApp.TrafficCaptureInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	exported, err := h.svc.ExportPCAP(r.Context(), input)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, exported)
}

func (h *Handler) downloadTrafficExport(w http.ResponseWriter, r *http.Request) {
	name := filepath.Base(r.PathValue("name"))
	if name == "." || name == "/" || name == "" {
		writeError(w, http.StatusBadRequest, "invalid export name")
		return
	}

	w.Header().Set("Content-Disposition", "attachment; filename=\""+name+"\"")
	http.ServeFile(w, r, filepath.Join("data", "pcap", name))
}

func (h *Handler) applyRules(w http.ResponseWriter, r *http.Request) {
	if err := h.svc.ApplyRules(r.Context()); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "applied"})
}

// ── Helpers ────────────────────────────────────────────────────

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
