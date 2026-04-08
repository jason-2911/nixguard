// Package network implements the network application service.
// Orchestrates interface management, routing, and gateway monitoring.
package network

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/nixguard/nixguard/internal/domain/network"
	"github.com/nixguard/nixguard/internal/event"
	"github.com/nixguard/nixguard/pkg/crypto"
)

// Service orchestrates network operations.
type Service struct {
	ifaces   network.InterfaceRepository
	routes   network.RouteRepository
	gateways network.GatewayRepository
	groups   network.GatewayGroupRepository
	engine   network.NetworkEngine
	bus      *event.Bus
	log      *slog.Logger

	// Gateway status cache
	gwMu     sync.RWMutex
	gwStatus map[string]*network.GatewayStatus
}

func NewService(
	ifaces network.InterfaceRepository,
	routes network.RouteRepository,
	gateways network.GatewayRepository,
	groups network.GatewayGroupRepository,
	engine network.NetworkEngine,
	bus *event.Bus,
	log *slog.Logger,
) *Service {
	return &Service{
		ifaces:   ifaces,
		routes:   routes,
		gateways: gateways,
		groups:   groups,
		engine:   engine,
		bus:      bus,
		log:      log,
		gwStatus: make(map[string]*network.GatewayStatus),
	}
}

// ═══════════════════════════════════════════════════════════════
// Interface Operations
// ═══════════════════════════════════════════════════════════════

func (s *Service) ListInterfaces(ctx context.Context) ([]network.Interface, error) {
	// Get configured interfaces from DB
	dbIfaces, err := s.ifaces.List(ctx)
	if err != nil {
		return nil, err
	}

	// Discover all system interfaces with live status
	sysIfaces, err := s.engine.DiscoverInterfaces(ctx)
	if err != nil {
		s.log.Warn("interface discovery failed, using DB only", slog.String("error", err.Error()))
		// Fallback: enrich DB records with individual status calls
		for i := range dbIfaces {
			if status, err := s.engine.GetInterfaceStatus(ctx, dbIfaces[i].Name); err == nil {
				dbIfaces[i].Status = *status
			}
		}
		return dbIfaces, nil
	}

	// Build lookup of DB interfaces by name
	dbByName := make(map[string]*network.Interface, len(dbIfaces))
	for i := range dbIfaces {
		dbByName[dbIfaces[i].Name] = &dbIfaces[i]
	}

	// Merge: for each system interface, overlay DB config if it exists
	result := make([]network.Interface, 0, len(sysIfaces))
	for _, sys := range sysIfaces {
		if db, ok := dbByName[sys.Name]; ok {
			// DB record exists — use DB config + live status
			db.Status = sys.Status
			db.MAC = sys.MAC
			db.MTU = sys.MTU
			result = append(result, *db)
			delete(dbByName, sys.Name)
		} else {
			// No DB record — show discovered interface as-is
			result = append(result, sys)
		}
	}

	// Append any DB-only interfaces (e.g., not yet created on system)
	for _, db := range dbByName {
		result = append(result, *db)
	}

	return result, nil
}

func (s *Service) GetInterface(ctx context.Context, id string) (*network.Interface, error) {
	iface, err := s.ifaces.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	status, err := s.engine.GetInterfaceStatus(ctx, iface.Name)
	if err == nil {
		iface.Status = *status
	}
	return iface, nil
}

func (s *Service) CreateInterface(ctx context.Context, input CreateInterfaceInput) (*network.Interface, error) {
	iface := &network.Interface{
		ID:           crypto.RandomID(8),
		Name:         input.Name,
		Alias:        input.Alias,
		Type:         input.Type,
		Enabled:      true,
		Description:  input.Description,
		MTU:          input.MTU,
		IPv4Config:   input.IPv4Config,
		IPv6Config:   input.IPv6Config,
		VLANConfig:   input.VLANConfig,
		BondConfig:   input.BondConfig,
		BridgeConfig: input.BridgeConfig,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if iface.MTU == 0 {
		iface.MTU = 1500
	}

	// Create in system
	switch iface.Type {
	case network.IfTypeVLAN:
		if iface.VLANConfig == nil {
			return nil, fmt.Errorf("VLAN config required")
		}
		name, err := s.engine.CreateVLAN(ctx, iface.VLANConfig.Parent, iface.VLANConfig.Tag)
		if err != nil {
			return nil, fmt.Errorf("create VLAN: %w", err)
		}
		iface.Name = name
	case network.IfTypeBond:
		if iface.BondConfig == nil {
			return nil, fmt.Errorf("bond config required")
		}
		if err := s.engine.CreateBond(ctx, iface.Name, *iface.BondConfig); err != nil {
			return nil, fmt.Errorf("create bond: %w", err)
		}
	case network.IfTypeBridge:
		if iface.BridgeConfig == nil {
			return nil, fmt.Errorf("bridge config required")
		}
		if err := s.engine.CreateBridge(ctx, iface.Name, *iface.BridgeConfig); err != nil {
			return nil, fmt.Errorf("create bridge: %w", err)
		}
	}

	// Apply IP address
	if iface.IPv4Config != nil && iface.IPv4Config.Mode == "static" && iface.IPv4Config.Address != "" {
		if err := s.engine.SetInterfaceAddress(ctx, iface.Name, iface.IPv4Config.Address); err != nil {
			s.log.Error("failed to set address", slog.String("error", err.Error()))
		}
	}

	// Set MTU
	if iface.MTU > 0 && iface.MTU != 1500 {
		s.engine.SetInterfaceMTU(ctx, iface.Name, iface.MTU)
	}

	// Bring up
	if iface.Enabled {
		s.engine.SetInterfaceUp(ctx, iface.Name)
	}

	// Persist
	if err := s.ifaces.Create(ctx, iface); err != nil {
		return nil, fmt.Errorf("save interface: %w", err)
	}

	s.bus.Publish(ctx, event.Event{Type: event.InterfaceUp, Source: "network", Payload: iface})
	return iface, nil
}

func (s *Service) UpdateInterface(ctx context.Context, id string, input UpdateInterfaceInput) (*network.Interface, error) {
	iface, err := s.ifaces.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if input.Alias != nil {
		iface.Alias = *input.Alias
	}
	if input.Description != nil {
		iface.Description = *input.Description
	}
	if input.MTU != nil {
		iface.MTU = *input.MTU
		s.engine.SetInterfaceMTU(ctx, iface.Name, iface.MTU)
	}
	if input.Enabled != nil {
		iface.Enabled = *input.Enabled
		if iface.Enabled {
			s.engine.SetInterfaceUp(ctx, iface.Name)
		} else {
			s.engine.SetInterfaceDown(ctx, iface.Name)
		}
	}
	if input.IPv4Config != nil {
		iface.IPv4Config = input.IPv4Config
		if iface.IPv4Config.Mode == "static" && iface.IPv4Config.Address != "" {
			s.engine.SetInterfaceAddress(ctx, iface.Name, iface.IPv4Config.Address)
		}
	}

	iface.UpdatedAt = time.Now()
	if err := s.ifaces.Update(ctx, iface); err != nil {
		return nil, err
	}

	return iface, nil
}

func (s *Service) DeleteInterface(ctx context.Context, id string) error {
	iface, err := s.ifaces.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Only allow deleting virtual interfaces
	switch iface.Type {
	case network.IfTypeVLAN:
		s.engine.DeleteVLAN(ctx, iface.Name)
	case network.IfTypeBond:
		s.engine.DeleteBond(ctx, iface.Name)
	case network.IfTypeBridge:
		s.engine.DeleteBridge(ctx, iface.Name)
	default:
		return fmt.Errorf("cannot delete physical interface %s", iface.Name)
	}

	return s.ifaces.Delete(ctx, id)
}

// ═══════════════════════════════════════════════════════════════
// Route Operations
// ═══════════════════════════════════════════════════════════════

func (s *Service) ListRoutes(ctx context.Context) ([]network.Route, error) {
	return s.routes.List(ctx)
}

func (s *Service) CreateRoute(ctx context.Context, input CreateRouteInput) (*network.Route, error) {
	route := &network.Route{
		ID:          crypto.RandomID(8),
		Destination: input.Destination,
		Gateway:     input.Gateway,
		Interface:   input.Interface,
		Metric:      input.Metric,
		Table:       input.Table,
		Type:        network.RouteStatic,
		Enabled:     true,
		Description: input.Description,
	}

	if route.Table == 0 {
		route.Table = 254 // main table
	}

	// Apply to kernel
	if err := s.engine.AddRoute(ctx, *route); err != nil {
		return nil, fmt.Errorf("add route: %w", err)
	}

	if err := s.routes.Create(ctx, route); err != nil {
		return nil, fmt.Errorf("save route: %w", err)
	}

	s.bus.Publish(ctx, event.Event{Type: event.RouteChanged, Source: "network", Payload: route})
	return route, nil
}

func (s *Service) DeleteRoute(ctx context.Context, id string) error {
	route, err := s.routes.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if err := s.engine.DeleteRoute(ctx, *route); err != nil {
		s.log.Warn("failed to delete kernel route", slog.String("error", err.Error()))
	}

	return s.routes.Delete(ctx, id)
}

// GetSystemRoutes returns routes from the kernel routing table.
func (s *Service) GetSystemRoutes(ctx context.Context) ([]network.Route, error) {
	return s.engine.ListRoutes(ctx, 0)
}

// ═══════════════════════════════════════════════════════════════
// Gateway Operations
// ═══════════════════════════════════════════════════════════════

func (s *Service) ListGateways(ctx context.Context) ([]network.Gateway, error) {
	gateways, err := s.gateways.List(ctx)
	if err != nil {
		return nil, err
	}

	s.gwMu.RLock()
	defer s.gwMu.RUnlock()
	for i := range gateways {
		if status, ok := s.gwStatus[gateways[i].ID]; ok {
			gateways[i].Status = *status
		}
	}

	return gateways, nil
}

func (s *Service) CreateGateway(ctx context.Context, input CreateGatewayInput) (*network.Gateway, error) {
	gw := &network.Gateway{
		ID:          crypto.RandomID(8),
		Name:        input.Name,
		Interface:   input.Interface,
		Address:     input.Address,
		Protocol:    input.Protocol,
		MonitorIP:   input.MonitorIP,
		Weight:      input.Weight,
		Priority:    input.Priority,
		IsDefault:   input.IsDefault,
		Enabled:     true,
		Description: input.Description,
		MonitorConfig: network.MonitorConfig{
			Interval:         input.MonitorInterval,
			LossThreshold:    input.LossThreshold,
			LatencyThreshold: input.LatencyThreshold,
			DownCount:        input.DownCount,
			Method:           input.MonitorMethod,
		},
	}

	if gw.Protocol == "" {
		gw.Protocol = "inet"
	}
	if gw.MonitorIP == "" {
		gw.MonitorIP = gw.Address
	}
	if gw.MonitorConfig.Interval == 0 {
		gw.MonitorConfig.Interval = 5
	}
	if gw.MonitorConfig.LossThreshold == 0 {
		gw.MonitorConfig.LossThreshold = 20
	}
	if gw.MonitorConfig.LatencyThreshold == 0 {
		gw.MonitorConfig.LatencyThreshold = 500
	}
	if gw.MonitorConfig.DownCount == 0 {
		gw.MonitorConfig.DownCount = 3
	}
	if gw.MonitorConfig.Method == "" {
		gw.MonitorConfig.Method = "icmp"
	}

	if err := s.gateways.Create(ctx, gw); err != nil {
		return nil, fmt.Errorf("save gateway: %w", err)
	}

	// Add default route if marked default
	if gw.IsDefault {
		s.engine.AddRoute(ctx, network.Route{
			Destination: "default",
			Gateway:     gw.Address,
			Interface:   gw.Interface,
			Metric:      gw.Priority,
		})
	}

	return gw, nil
}

func (s *Service) DeleteGateway(ctx context.Context, id string) error {
	return s.gateways.Delete(ctx, id)
}

func (s *Service) ListGatewayGroups(ctx context.Context) ([]network.GatewayGroup, error) {
	if s.groups == nil {
		return []network.GatewayGroup{}, nil
	}
	return s.groups.List(ctx)
}

func (s *Service) CreateGatewayGroup(ctx context.Context, input CreateGatewayGroupInput) (*network.GatewayGroup, error) {
	if s.groups == nil {
		return nil, fmt.Errorf("gateway groups repository not configured")
	}

	group := &network.GatewayGroup{
		ID:          crypto.RandomID(8),
		Name:        input.Name,
		Members:     append([]network.GatewayGroupMember(nil), input.Members...),
		Trigger:     input.Trigger,
		Description: input.Description,
	}
	if group.Trigger == "" {
		group.Trigger = "member_down"
	}

	if err := s.groups.Create(ctx, group); err != nil {
		return nil, fmt.Errorf("save gateway group: %w", err)
	}
	return group, nil
}

func (s *Service) DeleteGatewayGroup(ctx context.Context, id string) error {
	if s.groups == nil {
		return fmt.Errorf("gateway groups repository not configured")
	}
	return s.groups.Delete(ctx, id)
}

// MonitorGateways runs a single monitoring pass for all gateways.
func (s *Service) MonitorGateways(ctx context.Context) {
	gateways, err := s.gateways.List(ctx)
	if err != nil {
		s.log.Error("failed to list gateways for monitoring", slog.String("error", err.Error()))
		return
	}

	for _, gw := range gateways {
		if !gw.Enabled {
			continue
		}
		target := gw.MonitorIP
		if target == "" {
			target = gw.Address
		}

		var status *network.GatewayStatus
		var monErr error
		switch gw.MonitorConfig.Method {
		case "tcp":
			status, monErr = s.engine.CheckGatewayTCP(ctx, target, gw.MonitorConfig.Port)
		case "http":
			status, monErr = s.engine.CheckGatewayHTTP(ctx, gw.MonitorConfig.URL)
		default:
			status, monErr = s.engine.PingGateway(ctx, target)
		}
		err = monErr
		if err != nil {
			s.log.Warn("gateway monitor error", slog.String("gateway", gw.Name), slog.String("error", err.Error()))
			continue
		}
		status.LastCheck = time.Now()

		s.gwMu.Lock()
		prevStatus := s.gwStatus[gw.ID]
		s.gwStatus[gw.ID] = status
		s.gwMu.Unlock()

		// Detect state changes
		if prevStatus != nil && prevStatus.State != status.State {
			evtType := event.GatewayUp
			if status.State != "online" {
				evtType = event.GatewayDown
			}
			s.bus.Publish(ctx, event.Event{
				Type:    evtType,
				Source:  "network",
				Payload: map[string]interface{}{"gateway": gw.Name, "state": status.State},
			})
			s.log.Warn("gateway state changed",
				slog.String("gateway", gw.Name),
				slog.String("from", prevStatus.State),
				slog.String("to", status.State),
			)
		}
	}
}

// ═══════════════════════════════════════════════════════════════
// DTOs
// ═══════════════════════════════════════════════════════════════

type CreateInterfaceInput struct {
	Name         string                `json:"name" validate:"required"`
	Alias        string                `json:"alias"`
	Type         network.InterfaceType `json:"type" validate:"required"`
	Description  string                `json:"description"`
	MTU          int                   `json:"mtu"`
	IPv4Config   *network.IPv4Config   `json:"ipv4_config"`
	IPv6Config   *network.IPv6Config   `json:"ipv6_config"`
	VLANConfig   *network.VLANConfig   `json:"vlan_config"`
	BondConfig   *network.BondConfig   `json:"bond_config"`
	BridgeConfig *network.BridgeConfig `json:"bridge_config"`
}

type UpdateInterfaceInput struct {
	Alias       *string             `json:"alias,omitempty"`
	Description *string             `json:"description,omitempty"`
	MTU         *int                `json:"mtu,omitempty"`
	Enabled     *bool               `json:"enabled,omitempty"`
	IPv4Config  *network.IPv4Config `json:"ipv4_config,omitempty"`
	IPv6Config  *network.IPv6Config `json:"ipv6_config,omitempty"`
}

type CreateRouteInput struct {
	Destination string `json:"destination" validate:"required"`
	Gateway     string `json:"gateway"`
	Interface   string `json:"interface"`
	Metric      int    `json:"metric"`
	Table       int    `json:"table"`
	Description string `json:"description"`
}

type CreateGatewayInput struct {
	Name             string `json:"name" validate:"required"`
	Interface        string `json:"interface" validate:"required"`
	Address          string `json:"address" validate:"required"`
	Protocol         string `json:"protocol"`
	MonitorIP        string `json:"monitor_ip"`
	Weight           int    `json:"weight"`
	Priority         int    `json:"priority"`
	IsDefault        bool   `json:"is_default"`
	Description      string `json:"description"`
	MonitorInterval  int    `json:"monitor_interval"`
	LossThreshold    int    `json:"loss_threshold"`
	LatencyThreshold int    `json:"latency_threshold"`
	DownCount        int    `json:"down_count"`
	MonitorMethod    string `json:"monitor_method"`
}

type CreateGatewayGroupInput struct {
	Name        string                       `json:"name" validate:"required"`
	Members     []network.GatewayGroupMember `json:"members"`
	Trigger     string                       `json:"trigger"`
	Description string                       `json:"description"`
}
