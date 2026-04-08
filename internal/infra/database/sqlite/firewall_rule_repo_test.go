package sqlite

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/nixguard/nixguard/internal/domain/firewall"
	"github.com/nixguard/nixguard/internal/infra/database"
	"github.com/nixguard/nixguard/pkg/logger"
)

func setupTestDB(t *testing.T) *database.DB {
	t.Helper()
	log := logger.New("error", "text")

	tmpFile, err := os.CreateTemp("", "nixguard_test_*.db")
	if err != nil {
		t.Fatalf("create temp db: %v", err)
	}
	tmpFile.Close()
	t.Cleanup(func() { os.Remove(tmpFile.Name()) })

	db, err := database.Open("sqlite", tmpFile.Name(), log)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	// Run migrations
	err = db.Migrate("../../../../internal/infra/database/migrations")
	if err != nil {
		// Try relative from test location
		err = db.Migrate("../migrations")
		if err != nil {
			t.Fatalf("migrate: %v", err)
		}
	}

	return db
}

func TestRuleRepo_CRUD(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRuleRepo(db)
	ctx := context.Background()

	// Create
	rule := &firewall.Rule{
		ID:          "rule_test_001",
		Interface:   "eth0",
		Direction:   firewall.DirectionIn,
		Action:      firewall.ActionPass,
		Protocol:    firewall.ProtoTCP,
		Source:      firewall.Address{Type: firewall.AddrAny},
		Destination: firewall.Address{Type: firewall.AddrSingle, Value: "192.168.1.100", Port: "443"},
		Log:         true,
		Description: "Test HTTPS rule",
		Enabled:     true,
		Order:       0,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	err := repo.Create(ctx, rule)
	if err != nil {
		t.Fatalf("create rule: %v", err)
	}

	// Read
	got, err := repo.GetByID(ctx, "rule_test_001")
	if err != nil {
		t.Fatalf("get rule: %v", err)
	}
	if got.Interface != "eth0" {
		t.Errorf("expected eth0, got %s", got.Interface)
	}
	if got.Action != firewall.ActionPass {
		t.Errorf("expected pass, got %s", got.Action)
	}
	if got.Destination.Value != "192.168.1.100" {
		t.Errorf("expected 192.168.1.100, got %s", got.Destination.Value)
	}
	if got.Destination.Port != "443" {
		t.Errorf("expected port 443, got %s", got.Destination.Port)
	}
	if !got.Log {
		t.Error("expected log=true")
	}
	if !got.Enabled {
		t.Error("expected enabled=true")
	}

	// List
	rules, err := repo.List(ctx, firewall.RuleFilter{})
	if err != nil {
		t.Fatalf("list rules: %v", err)
	}
	if len(rules) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(rules))
	}

	// List with filter
	rules, err = repo.List(ctx, firewall.RuleFilter{Interface: "eth0"})
	if err != nil {
		t.Fatalf("list rules filtered: %v", err)
	}
	if len(rules) != 1 {
		t.Errorf("expected 1 rule for eth0, got %d", len(rules))
	}
	rules, err = repo.List(ctx, firewall.RuleFilter{Interface: "eth1"})
	if err != nil {
		t.Fatalf("list rules filtered: %v", err)
	}
	if len(rules) != 0 {
		t.Errorf("expected 0 rules for eth1, got %d", len(rules))
	}

	// Update
	got.Description = "Updated description"
	got.Action = firewall.ActionBlock
	got.UpdatedAt = time.Now()
	err = repo.Update(ctx, got)
	if err != nil {
		t.Fatalf("update rule: %v", err)
	}
	got2, _ := repo.GetByID(ctx, "rule_test_001")
	if got2.Description != "Updated description" {
		t.Errorf("expected updated description")
	}
	if got2.Action != firewall.ActionBlock {
		t.Errorf("expected block action after update")
	}

	// Delete
	err = repo.Delete(ctx, "rule_test_001")
	if err != nil {
		t.Fatalf("delete rule: %v", err)
	}
	_, err = repo.GetByID(ctx, "rule_test_001")
	if err == nil {
		t.Error("expected error after delete")
	}
}

func TestRuleRepo_Reorder(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRuleRepo(db)
	ctx := context.Background()

	now := time.Now()
	for i, id := range []string{"r1", "r2", "r3"} {
		repo.Create(ctx, &firewall.Rule{
			ID: id, Direction: firewall.DirectionIn, Action: firewall.ActionPass,
			Protocol: firewall.ProtoAny, Source: firewall.Address{Type: firewall.AddrAny},
			Destination: firewall.Address{Type: firewall.AddrAny},
			Order: i, Enabled: true, CreatedAt: now, UpdatedAt: now,
		})
	}

	// Reorder: r3, r1, r2
	err := repo.Reorder(ctx, []string{"r3", "r1", "r2"})
	if err != nil {
		t.Fatalf("reorder: %v", err)
	}

	rules, _ := repo.List(ctx, firewall.RuleFilter{})
	if len(rules) != 3 {
		t.Fatalf("expected 3 rules")
	}
	if rules[0].ID != "r3" {
		t.Errorf("expected r3 first, got %s", rules[0].ID)
	}
	if rules[1].ID != "r1" {
		t.Errorf("expected r1 second, got %s", rules[1].ID)
	}
	if rules[2].ID != "r2" {
		t.Errorf("expected r2 third, got %s", rules[2].ID)
	}
}

func TestAliasRepo_CRUD(t *testing.T) {
	db := setupTestDB(t)
	repo := NewAliasRepo(db)
	ctx := context.Background()

	alias := &firewall.Alias{
		ID:          "alias_test_001",
		Name:        "webservers",
		Type:        firewall.AliasHost,
		Description: "Web server IPs",
		Entries:     []string{"10.0.0.1", "10.0.0.2", "10.0.0.3"},
		Enabled:     true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	err := repo.Create(ctx, alias)
	if err != nil {
		t.Fatalf("create alias: %v", err)
	}

	got, err := repo.GetByName(ctx, "webservers")
	if err != nil {
		t.Fatalf("get alias by name: %v", err)
	}
	if len(got.Entries) != 3 {
		t.Errorf("expected 3 entries, got %d", len(got.Entries))
	}
	if got.Entries[0] != "10.0.0.1" {
		t.Errorf("expected 10.0.0.1, got %s", got.Entries[0])
	}

	aliases, err := repo.List(ctx)
	if err != nil {
		t.Fatalf("list aliases: %v", err)
	}
	if len(aliases) != 1 {
		t.Errorf("expected 1 alias, got %d", len(aliases))
	}
}

func TestNATRepo_CRUD(t *testing.T) {
	db := setupTestDB(t)
	repo := NewNATRepo(db)
	ctx := context.Background()

	rule := &firewall.NATRule{
		ID:             "nat_test_001",
		Type:           firewall.NATPortForward,
		Interface:      "eth0",
		Protocol:       firewall.ProtoTCP,
		Source:         firewall.Address{Type: firewall.AddrAny},
		Destination:    firewall.Address{Type: firewall.AddrSingle, Value: "1.2.3.4", Port: "443"},
		RedirectTarget: "192.168.1.100",
		RedirectPort:   "443",
		Description:    "Forward HTTPS",
		Enabled:        true,
		NATReflection:  true,
		CreatedAt:      time.Now(),
	}

	err := repo.Create(ctx, rule)
	if err != nil {
		t.Fatalf("create NAT rule: %v", err)
	}

	rules, err := repo.List(ctx, firewall.NATPortForward)
	if err != nil {
		t.Fatalf("list NAT rules: %v", err)
	}
	if len(rules) != 1 {
		t.Fatalf("expected 1 NAT rule, got %d", len(rules))
	}
	if rules[0].RedirectTarget != "192.168.1.100" {
		t.Errorf("expected redirect 192.168.1.100, got %s", rules[0].RedirectTarget)
	}
	if !rules[0].NATReflection {
		t.Error("expected NAT reflection enabled")
	}

	// Filter by type
	outbound, _ := repo.List(ctx, firewall.NATOutbound)
	if len(outbound) != 0 {
		t.Errorf("expected 0 outbound rules, got %d", len(outbound))
	}
}
