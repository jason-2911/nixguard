package sqlite

import (
	"context"
	"testing"
	"time"

	"github.com/nixguard/nixguard/internal/domain/network"
)

func TestInterfaceRepo_Update(t *testing.T) {
	db := setupTestDB(t)
	repo := NewInterfaceRepo(db)
	ctx := context.Background()

	iface := &network.Interface{
		ID:          "if_test_001",
		Name:        "eth10",
		Alias:       "LAN",
		Type:        network.IfTypePhysical,
		Enabled:     true,
		Description: "initial",
		MTU:         1500,
		IPv4Config:  &network.IPv4Config{Mode: "static", Address: "192.168.10.1/24", Gateway: ""},
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := repo.Create(ctx, iface); err != nil {
		t.Fatalf("create interface: %v", err)
	}

	iface.Description = "updated"
	iface.MTU = 9000
	iface.Enabled = false
	iface.IPv6Config = &network.IPv6Config{Mode: "static", Address: "2001:db8::1/64", Gateway: "2001:db8::ff"}
	iface.UpdatedAt = time.Now()

	if err := repo.Update(ctx, iface); err != nil {
		t.Fatalf("update interface: %v", err)
	}

	got, err := repo.GetByID(ctx, iface.ID)
	if err != nil {
		t.Fatalf("get interface: %v", err)
	}
	if got.Description != "updated" {
		t.Fatalf("expected updated description, got %q", got.Description)
	}
	if got.MTU != 9000 {
		t.Fatalf("expected mtu 9000, got %d", got.MTU)
	}
	if got.Enabled {
		t.Fatal("expected interface disabled after update")
	}
	if got.IPv6Config == nil || got.IPv6Config.Address != "2001:db8::1/64" {
		t.Fatalf("expected ipv6 config preserved, got %#v", got.IPv6Config)
	}
}

func TestRouteRepo_CRUD(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRouteRepo(db)
	ctx := context.Background()

	route := &network.Route{
		ID:          "route_test_001",
		Destination: "2001:db8:100::/64",
		Gateway:     "2001:db8::1",
		Interface:   "wan0",
		Metric:      20,
		Table:       100,
		Type:        network.RouteStatic,
		Enabled:     true,
		Description: "IPv6 static route",
	}

	if err := repo.Create(ctx, route); err != nil {
		t.Fatalf("create route: %v", err)
	}

	got, err := repo.GetByID(ctx, route.ID)
	if err != nil {
		t.Fatalf("get route: %v", err)
	}
	if got.Table != 100 {
		t.Fatalf("expected table 100, got %d", got.Table)
	}

	got.Metric = 5
	if err := repo.Update(ctx, got); err != nil {
		t.Fatalf("update route: %v", err)
	}

	list, err := repo.List(ctx)
	if err != nil {
		t.Fatalf("list routes: %v", err)
	}
	if len(list) != 1 || list[0].Metric != 5 {
		t.Fatalf("expected updated route metric 5, got %#v", list)
	}

	if err := repo.Delete(ctx, route.ID); err != nil {
		t.Fatalf("delete route: %v", err)
	}
}

func TestGatewayRepo_CRUD(t *testing.T) {
	db := setupTestDB(t)
	repo := NewGatewayRepo(db)
	ctx := context.Background()

	gateway := &network.Gateway{
		ID:          "gw_test_001",
		Name:        "wan-primary",
		Interface:   "wan0",
		Address:     "192.0.2.1",
		Protocol:    "inet",
		MonitorIP:   "1.1.1.1",
		Weight:      1,
		Priority:    10,
		IsDefault:   true,
		Enabled:     true,
		Description: "primary upstream",
		MonitorConfig: network.MonitorConfig{
			Interval:         5,
			LossThreshold:    20,
			LatencyThreshold: 500,
			DownCount:        3,
			Method:           "icmp",
		},
	}

	if err := repo.Create(ctx, gateway); err != nil {
		t.Fatalf("create gateway: %v", err)
	}

	got, err := repo.GetByID(ctx, gateway.ID)
	if err != nil {
		t.Fatalf("get gateway: %v", err)
	}
	if got.Name != gateway.Name || !got.IsDefault {
		t.Fatalf("gateway mismatch: %#v", got)
	}

	got.Priority = 30
	if err := repo.Update(ctx, got); err != nil {
		t.Fatalf("update gateway: %v", err)
	}

	list, err := repo.List(ctx)
	if err != nil {
		t.Fatalf("list gateways: %v", err)
	}
	if len(list) != 1 || list[0].Priority != 30 {
		t.Fatalf("expected updated priority 30, got %#v", list)
	}

	if err := repo.Delete(ctx, gateway.ID); err != nil {
		t.Fatalf("delete gateway: %v", err)
	}
}

func TestGatewayGroupRepo_CRUD(t *testing.T) {
	db := setupTestDB(t)
	repo := NewGatewayGroupRepo(db)
	ctx := context.Background()

	group := &network.GatewayGroup{
		ID:   "gwgrp_test_001",
		Name: "wan-failover",
		Members: []network.GatewayGroupMember{
			{GatewayID: "gw1", Tier: 1, Weight: 1},
			{GatewayID: "gw2", Tier: 2, Weight: 1},
		},
		Trigger:     "member_down",
		Description: "primary/failover",
	}

	if err := repo.Create(ctx, group); err != nil {
		t.Fatalf("create gateway group: %v", err)
	}

	got, err := repo.GetByID(ctx, group.ID)
	if err != nil {
		t.Fatalf("get gateway group: %v", err)
	}
	if got.Name != group.Name || len(got.Members) != 2 {
		t.Fatalf("gateway group mismatch: %#v", got)
	}

	got.Trigger = "packet_loss"
	if err := repo.Update(ctx, got); err != nil {
		t.Fatalf("update gateway group: %v", err)
	}

	list, err := repo.List(ctx)
	if err != nil {
		t.Fatalf("list gateway groups: %v", err)
	}
	if len(list) != 1 || list[0].Trigger != "packet_loss" {
		t.Fatalf("expected updated trigger packet_loss, got %#v", list)
	}

	if err := repo.Delete(ctx, group.ID); err != nil {
		t.Fatalf("delete gateway group: %v", err)
	}
}
