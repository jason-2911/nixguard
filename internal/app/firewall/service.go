// Package firewall implements the firewall application service.
// It orchestrates domain logic, persistence, and infrastructure.
package firewall

import (
	"bufio"
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/nixguard/nixguard/internal/domain/firewall"
	"github.com/nixguard/nixguard/internal/event"
	"github.com/nixguard/nixguard/pkg/crypto"
)

// Service orchestrates firewall operations.
type Service struct {
	rules   firewall.RuleRepository
	aliases firewall.AliasRepository
	nat     firewall.NATRepository
	engine  firewall.FirewallEngine
	geoip   firewall.GeoIPProvider
	bus     *event.Bus
	log     *slog.Logger
}

// NewService creates a firewall application service.
func NewService(
	rules firewall.RuleRepository,
	aliases firewall.AliasRepository,
	nat firewall.NATRepository,
	engine firewall.FirewallEngine,
	geoip firewall.GeoIPProvider,
	bus *event.Bus,
	log *slog.Logger,
) *Service {
	return &Service{
		rules:   rules,
		aliases: aliases,
		nat:     nat,
		engine:  engine,
		geoip:   geoip,
		bus:     bus,
		log:     log,
	}
}

// ─── Rule Operations ───────────────────────────────────────────

// ListRules returns firewall rules matching the filter.
func (s *Service) ListRules(ctx context.Context, filter firewall.RuleFilter) ([]firewall.Rule, error) {
	rules, err := s.rules.List(ctx, filter)
	if err != nil {
		return nil, err
	}
	s.attachRuleStats(ctx, rules)
	return rules, nil
}

// GetRule returns a single rule by ID.
func (s *Service) GetRule(ctx context.Context, id string) (*firewall.Rule, error) {
	rule, err := s.rules.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	stats, err := s.engine.GetRuleStats(ctx)
	if err == nil {
		if stat, ok := stats[rule.ID]; ok {
			rule.Stats = stat
		}
	}
	return rule, nil
}

// CreateRule creates a new firewall rule and applies the ruleset.
func (s *Service) CreateRule(ctx context.Context, input CreateRuleInput) (*firewall.Rule, error) {
	rule := &firewall.Rule{
		ID:          crypto.RandomID(8),
		Interface:   input.Interface,
		Direction:   input.Direction,
		Action:      input.Action,
		Protocol:    input.Protocol,
		Source:      input.Source,
		Destination: input.Destination,
		Log:         input.Log,
		Description: input.Description,
		Enabled:     true,
		Order:       input.Order,
		IsFloating:  input.IsFloating,
		Interfaces:  input.Interfaces,
		Gateway:     input.Gateway,
		Category:    input.Category,
		StateType:   input.StateType,
		MaxStates:   input.MaxStates,
		Tag:         input.Tag,
		Tagged:      input.Tagged,
		Schedule:    input.Schedule,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.rules.Create(ctx, rule); err != nil {
		return nil, fmt.Errorf("create rule: %w", err)
	}

	// Apply updated ruleset
	if err := s.applyRuleset(ctx); err != nil {
		s.log.Error("failed to apply ruleset after create", slog.String("error", err.Error()))
		// Rule is saved but not applied — log but don't rollback
	}

	s.bus.Publish(ctx, event.Event{
		Type:    event.FirewallRuleCreated,
		Source:  "firewall",
		Payload: rule,
	})

	return rule, nil
}

// UpdateRule modifies an existing rule and reapplies.
func (s *Service) UpdateRule(ctx context.Context, id string, input UpdateRuleInput) (*firewall.Rule, error) {
	rule, err := s.rules.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if input.Interface != nil {
		rule.Interface = *input.Interface
	}
	if input.Direction != nil {
		rule.Direction = *input.Direction
	}
	if input.Action != nil {
		rule.Action = *input.Action
	}
	if input.Protocol != nil {
		rule.Protocol = *input.Protocol
	}
	if input.Source != nil {
		rule.Source = *input.Source
	}
	if input.Destination != nil {
		rule.Destination = *input.Destination
	}
	if input.Enabled != nil {
		rule.Enabled = *input.Enabled
	}
	if input.Description != nil {
		rule.Description = *input.Description
	}
	if input.Log != nil {
		rule.Log = *input.Log
	}
	if input.Order != nil {
		rule.Order = *input.Order
	}
	if input.IsFloating != nil {
		rule.IsFloating = *input.IsFloating
	}
	if input.Interfaces != nil {
		rule.Interfaces = append([]string(nil), (*input.Interfaces)...)
	}
	if input.Gateway != nil {
		rule.Gateway = *input.Gateway
	}
	if input.Category != nil {
		rule.Category = *input.Category
	}
	if input.StateType != nil {
		rule.StateType = *input.StateType
	}
	if input.MaxStates != nil {
		rule.MaxStates = *input.MaxStates
	}
	if input.Tag != nil {
		rule.Tag = *input.Tag
	}
	if input.Tagged != nil {
		rule.Tagged = *input.Tagged
	}
	if input.Schedule != nil {
		rule.Schedule = input.Schedule
	}
	rule.UpdatedAt = time.Now()

	if err := s.rules.Update(ctx, rule); err != nil {
		return nil, fmt.Errorf("update rule: %w", err)
	}

	if err := s.applyRuleset(ctx); err != nil {
		s.log.Error("failed to apply ruleset after update", slog.String("error", err.Error()))
	}

	s.bus.Publish(ctx, event.Event{
		Type:    event.FirewallRuleUpdated,
		Source:  "firewall",
		Payload: rule,
	})

	return rule, nil
}

// DeleteRule removes a rule and reapplies.
func (s *Service) DeleteRule(ctx context.Context, id string) error {
	if err := s.rules.Delete(ctx, id); err != nil {
		return fmt.Errorf("delete rule: %w", err)
	}

	if err := s.applyRuleset(ctx); err != nil {
		s.log.Error("failed to apply ruleset after delete", slog.String("error", err.Error()))
	}

	s.bus.Publish(ctx, event.Event{
		Type:    event.FirewallRuleDeleted,
		Source:  "firewall",
		Payload: map[string]string{"id": id},
	})

	return nil
}

// ReorderRules changes the evaluation order of rules.
func (s *Service) ReorderRules(ctx context.Context, ruleIDs []string) error {
	if err := s.rules.Reorder(ctx, ruleIDs); err != nil {
		return fmt.Errorf("reorder rules: %w", err)
	}
	return s.applyRuleset(ctx)
}

// ApplyRules forces a full ruleset recompile and reload.
func (s *Service) ApplyRules(ctx context.Context) error {
	return s.applyRuleset(ctx)
}

// GetRuleStats exposes kernel counters for the GUI.
func (s *Service) GetRuleStats(ctx context.Context) (map[string]firewall.RuleStats, error) {
	return s.engine.GetRuleStats(ctx)
}

// ─── Alias Operations ──────────────────────────────────────────

func (s *Service) ListAliases(ctx context.Context) ([]firewall.Alias, error) {
	return s.aliases.List(ctx)
}

func (s *Service) GetAlias(ctx context.Context, id string) (*firewall.Alias, error) {
	return s.aliases.GetByID(ctx, id)
}

func (s *Service) CreateAlias(ctx context.Context, input CreateAliasInput) (*firewall.Alias, error) {
	alias := &firewall.Alias{
		ID:          crypto.RandomID(8),
		Name:        input.Name,
		Type:        input.Type,
		Description: input.Description,
		Entries:     input.Entries,
		UpdateFreq:  input.UpdateFreq,
		Enabled:     true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.aliases.Create(ctx, alias); err != nil {
		return nil, fmt.Errorf("create alias: %w", err)
	}

	if err := s.applyRuleset(ctx); err != nil {
		s.log.Error("failed to apply ruleset after alias create", slog.String("error", err.Error()))
	}

	return alias, nil
}

func (s *Service) UpdateAlias(ctx context.Context, id string, input UpdateAliasInput) (*firewall.Alias, error) {
	alias, err := s.aliases.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if input.Name != nil {
		alias.Name = *input.Name
	}
	if input.Type != nil {
		alias.Type = *input.Type
	}
	if input.Description != nil {
		alias.Description = *input.Description
	}
	if input.Entries != nil {
		alias.Entries = append([]string(nil), (*input.Entries)...)
	}
	if input.UpdateFreq != nil {
		alias.UpdateFreq = *input.UpdateFreq
	}
	if input.Enabled != nil {
		alias.Enabled = *input.Enabled
	}
	alias.UpdatedAt = time.Now()

	if err := s.aliases.Update(ctx, alias); err != nil {
		return nil, fmt.Errorf("update alias: %w", err)
	}
	if err := s.applyRuleset(ctx); err != nil {
		s.log.Error("failed to apply ruleset after alias update", slog.String("error", err.Error()))
	}
	return alias, nil
}

func (s *Service) DeleteAlias(ctx context.Context, id string) error {
	if err := s.aliases.Delete(ctx, id); err != nil {
		return fmt.Errorf("delete alias: %w", err)
	}
	if err := s.applyRuleset(ctx); err != nil {
		s.log.Error("failed to apply ruleset after alias delete", slog.String("error", err.Error()))
	}
	return nil
}

// ─── NAT Operations ────────────────────────────────────────────

func (s *Service) ListNATRules(ctx context.Context, natType firewall.NATType) ([]firewall.NATRule, error) {
	return s.nat.List(ctx, natType)
}

func (s *Service) GetNATRule(ctx context.Context, id string) (*firewall.NATRule, error) {
	return s.nat.GetByID(ctx, id)
}

func (s *Service) CreateNATRule(ctx context.Context, input CreateNATInput) (*firewall.NATRule, error) {
	rule := &firewall.NATRule{
		ID:             crypto.RandomID(8),
		Type:           input.Type,
		Interface:      input.Interface,
		Protocol:       input.Protocol,
		Source:         input.Source,
		Destination:    input.Destination,
		RedirectTarget: input.RedirectTarget,
		RedirectPort:   input.RedirectPort,
		Description:    input.Description,
		Enabled:        true,
		NATReflection:  input.NATReflection,
		CreatedAt:      time.Now(),
	}

	if err := s.nat.Create(ctx, rule); err != nil {
		return nil, fmt.Errorf("create NAT rule: %w", err)
	}

	if err := s.applyNATRules(ctx); err != nil {
		s.log.Error("failed to apply NAT rules", slog.String("error", err.Error()))
	}

	return rule, nil
}

func (s *Service) UpdateNATRule(ctx context.Context, id string, input UpdateNATInput) (*firewall.NATRule, error) {
	rule, err := s.nat.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if input.Type != nil {
		rule.Type = *input.Type
	}
	if input.Interface != nil {
		rule.Interface = *input.Interface
	}
	if input.Protocol != nil {
		rule.Protocol = *input.Protocol
	}
	if input.Source != nil {
		rule.Source = *input.Source
	}
	if input.Destination != nil {
		rule.Destination = *input.Destination
	}
	if input.RedirectTarget != nil {
		rule.RedirectTarget = *input.RedirectTarget
	}
	if input.RedirectPort != nil {
		rule.RedirectPort = *input.RedirectPort
	}
	if input.Description != nil {
		rule.Description = *input.Description
	}
	if input.NATReflection != nil {
		rule.NATReflection = *input.NATReflection
	}
	if input.Enabled != nil {
		rule.Enabled = *input.Enabled
	}

	if err := s.nat.Update(ctx, rule); err != nil {
		return nil, fmt.Errorf("update NAT rule: %w", err)
	}
	if err := s.applyNATRules(ctx); err != nil {
		s.log.Error("failed to apply NAT rules after update", slog.String("error", err.Error()))
	}
	return rule, nil
}

func (s *Service) DeleteNATRule(ctx context.Context, id string) error {
	if err := s.nat.Delete(ctx, id); err != nil {
		return fmt.Errorf("delete NAT rule: %w", err)
	}
	if err := s.applyNATRules(ctx); err != nil {
		s.log.Error("failed to apply NAT rules after delete", slog.String("error", err.Error()))
	}
	return nil
}

// ─── State Table ───────────────────────────────────────────────

func (s *Service) GetStates(ctx context.Context, filter firewall.StateFilter) ([]firewall.State, error) {
	return s.engine.GetStates(ctx, filter)
}

func (s *Service) FlushStates(ctx context.Context, filter firewall.StateFilter) error {
	return s.engine.FlushStates(ctx, filter)
}

// ─── Live Traffic ──────────────────────────────────────────────

func (s *Service) CaptureTraffic(ctx context.Context, input TrafficCaptureInput) ([]firewall.CapturedPacket, error) {
	return s.engine.CaptureTraffic(ctx, firewall.TrafficFilter{
		Interface: input.Interface,
		SourceIP:  input.SourceIP,
		DestIP:    input.DestIP,
		Protocol:  input.Protocol,
		Count:     input.Count,
		SnapLen:   input.SnapLen,
	})
}

func (s *Service) ExportPCAP(ctx context.Context, input TrafficCaptureInput) (*firewall.PCAPExport, error) {
	return s.engine.ExportPCAP(ctx, firewall.TrafficFilter{
		Interface: input.Interface,
		SourceIP:  input.SourceIP,
		DestIP:    input.DestIP,
		Protocol:  input.Protocol,
		Count:     input.Count,
		SnapLen:   input.SnapLen,
	})
}

// ─── Internal ──────────────────────────────────────────────────

// applyRuleset fetches all rules and aliases, then applies atomically.
func (s *Service) applyRuleset(ctx context.Context) error {
	rules, err := s.rules.List(ctx, firewall.RuleFilter{})
	if err != nil {
		return fmt.Errorf("fetch rules: %w", err)
	}

	aliases, err := s.aliases.List(ctx)
	if err != nil {
		return fmt.Errorf("fetch aliases: %w", err)
	}

	effectiveAliases, err := s.effectiveAliases(ctx, aliases)
	if err != nil {
		return fmt.Errorf("resolve aliases: %w", err)
	}

	if err := s.engine.ApplyRules(ctx, rules, effectiveAliases); err != nil {
		return fmt.Errorf("apply rules: %w", err)
	}

	s.bus.Publish(ctx, event.Event{
		Type:   event.FirewallRulesApplied,
		Source: "firewall",
	})

	return nil
}

func (s *Service) applyNATRules(ctx context.Context) error {
	rules, err := s.nat.List(ctx, "")
	if err != nil {
		return fmt.Errorf("fetch NAT rules: %w", err)
	}
	if err := s.engine.ApplyNAT(ctx, rules); err != nil {
		return fmt.Errorf("apply NAT: %w", err)
	}
	return nil
}

func (s *Service) attachRuleStats(ctx context.Context, rules []firewall.Rule) {
	if len(rules) == 0 {
		return
	}

	stats, err := s.engine.GetRuleStats(ctx)
	if err != nil {
		return
	}

	for i := range rules {
		if stat, ok := stats[rules[i].ID]; ok {
			rules[i].Stats = stat
		}
	}
}

func (s *Service) effectiveAliases(ctx context.Context, aliases []firewall.Alias) ([]firewall.Alias, error) {
	byName := make(map[string]firewall.Alias, len(aliases))
	for _, alias := range aliases {
		byName[alias.Name] = alias
	}

	var out []firewall.Alias
	for _, alias := range aliases {
		resolved, err := s.resolveAliasEntries(ctx, alias, byName, map[string]bool{})
		if err != nil {
			return nil, err
		}
		alias.Entries = resolved
		out = append(out, alias)
	}
	return out, nil
}

func (s *Service) resolveAliasEntries(
	ctx context.Context,
	alias firewall.Alias,
	byName map[string]firewall.Alias,
	seen map[string]bool,
) ([]string, error) {
	if seen[alias.Name] {
		return nil, fmt.Errorf("alias cycle detected at %s", alias.Name)
	}

	seen[alias.Name] = true
	defer delete(seen, alias.Name)

	switch alias.Type {
	case firewall.AliasNested:
		var resolved []string
		for _, name := range alias.Entries {
			ref, ok := byName[name]
			if !ok {
				continue
			}
			entries, err := s.resolveAliasEntries(ctx, ref, byName, seen)
			if err != nil {
				return nil, err
			}
			resolved = append(resolved, entries...)
		}
		return dedupeStrings(resolved), nil
	case firewall.AliasURL, firewall.AliasURLTable:
		entries, err := fetchURLAliasEntries(ctx, alias.Entries)
		if err != nil {
			s.log.Warn("failed to refresh URL alias, using persisted entries",
				slog.String("alias", alias.Name),
				slog.String("error", err.Error()),
			)
			return dedupeStrings(alias.Entries), nil
		}
		return dedupeStrings(entries), nil
	case firewall.AliasGeoIP:
		if s.geoip == nil {
			return dedupeStrings(alias.Entries), nil
		}

		var resolved []string
		for _, countryCode := range alias.Entries {
			entries, err := s.geoip.Resolve(ctx, countryCode)
			if err != nil {
				return nil, fmt.Errorf("resolve geoip %s: %w", countryCode, err)
			}
			resolved = append(resolved, entries...)
		}
		return dedupeStrings(resolved), nil
	default:
		return dedupeStrings(alias.Entries), nil
	}
}

func fetchURLAliasEntries(ctx context.Context, urls []string) ([]string, error) {
	client := &http.Client{Timeout: 10 * time.Second}

	var entries []string
	for _, rawURL := range urls {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
		if err != nil {
			return nil, err
		}

		resp, err := client.Do(req)
		if err != nil {
			return nil, err
		}

		func() {
			defer resp.Body.Close()
			scanner := bufio.NewScanner(resp.Body)
			for scanner.Scan() {
				line := strings.TrimSpace(scanner.Text())
				if line == "" || strings.HasPrefix(line, "#") {
					continue
				}
				entries = append(entries, line)
			}
		}()
	}

	return dedupeStrings(entries), nil
}

// StartURLRefresher launches a background goroutine that periodically
// re-fetches URL-based alias entries and reapplies the ruleset.
func (s *Service) StartURLRefresher(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				s.log.Info("URL alias refresher stopped")
				return
			case <-ticker.C:
				s.refreshURLAliases(ctx)
			}
		}
	}()
	s.log.Info("URL alias refresher started", slog.Duration("interval", 5*time.Minute))
}

func (s *Service) refreshURLAliases(ctx context.Context) {
	aliases, err := s.aliases.List(ctx)
	if err != nil {
		s.log.Error("URL refresh: failed to list aliases", slog.String("error", err.Error()))
		return
	}

	updated := false
	for _, alias := range aliases {
		if !alias.Enabled {
			continue
		}
		if alias.Type != firewall.AliasURL && alias.Type != firewall.AliasURLTable {
			continue
		}

		entries, err := fetchURLAliasEntries(ctx, alias.Entries)
		if err != nil {
			s.log.Warn("URL refresh: fetch failed",
				slog.String("alias", alias.Name),
				slog.String("error", err.Error()),
			)
			continue
		}

		if len(entries) > 0 {
			alias.Entries = entries
			if err := s.aliases.Update(ctx, &alias); err != nil {
				s.log.Error("URL refresh: update failed",
					slog.String("alias", alias.Name),
					slog.String("error", err.Error()),
				)
				continue
			}
			updated = true
			s.log.Info("URL alias refreshed",
				slog.String("alias", alias.Name),
				slog.Int("entries", len(entries)),
			)
		}
	}

	if updated {
		if err := s.applyRuleset(ctx); err != nil {
			s.log.Error("URL refresh: failed to reapply ruleset", slog.String("error", err.Error()))
		}
	}
}

func dedupeStrings(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	out := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	return out
}
