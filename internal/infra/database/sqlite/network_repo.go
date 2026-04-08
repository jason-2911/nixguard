package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/nixguard/nixguard/internal/domain/network"
	"github.com/nixguard/nixguard/internal/infra/database"
)

// ═══════════════════════════════════════════════════════════════
// Interface Repository
// ═══════════════════════════════════════════════════════════════

type InterfaceRepo struct {
	db *database.DB
}

func NewInterfaceRepo(db *database.DB) *InterfaceRepo {
	return &InterfaceRepo{db: db}
}

func (r *InterfaceRepo) List(ctx context.Context) ([]network.Interface, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, name, alias_name, if_type, enabled, description, mtu, mac_address,
		ipv4_mode, ipv4_address, ipv4_gateway, ipv6_mode, ipv6_address, ipv6_gateway,
		vlan_parent, vlan_tag, bond_members, bond_mode, bridge_members, bridge_stp,
		pppoe_parent, pppoe_user, pppoe_pass, created_at, updated_at
		FROM network_interfaces ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ifaces []network.Interface
	for rows.Next() {
		iface, err := scanInterface(rows)
		if err != nil {
			return nil, err
		}
		ifaces = append(ifaces, *iface)
	}
	return ifaces, rows.Err()
}

func (r *InterfaceRepo) GetByID(ctx context.Context, id string) (*network.Interface, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, name, alias_name, if_type, enabled, description, mtu, mac_address,
		ipv4_mode, ipv4_address, ipv4_gateway, ipv6_mode, ipv6_address, ipv6_gateway,
		vlan_parent, vlan_tag, bond_members, bond_mode, bridge_members, bridge_stp,
		pppoe_parent, pppoe_user, pppoe_pass, created_at, updated_at
		FROM network_interfaces WHERE id = ?`, id)
	return scanInterfaceRow(row)
}

func (r *InterfaceRepo) GetByName(ctx context.Context, name string) (*network.Interface, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, name, alias_name, if_type, enabled, description, mtu, mac_address,
		ipv4_mode, ipv4_address, ipv4_gateway, ipv6_mode, ipv6_address, ipv6_gateway,
		vlan_parent, vlan_tag, bond_members, bond_mode, bridge_members, bridge_stp,
		pppoe_parent, pppoe_user, pppoe_pass, created_at, updated_at
		FROM network_interfaces WHERE name = ?`, name)
	return scanInterfaceRow(row)
}

func (r *InterfaceRepo) Create(ctx context.Context, iface *network.Interface) error {
	bondMembers, _ := json.Marshal(ifaceSlice(iface.BondConfig))
	bridgeMembers, _ := json.Marshal(bridgeSlice(iface.BridgeConfig))

	ipv4Mode, ipv4Addr, ipv4Gw := "", "", ""
	if iface.IPv4Config != nil {
		ipv4Mode = iface.IPv4Config.Mode
		ipv4Addr = iface.IPv4Config.Address
		ipv4Gw = iface.IPv4Config.Gateway
	}
	ipv6Mode, ipv6Addr, ipv6Gw := "", "", ""
	if iface.IPv6Config != nil {
		ipv6Mode = iface.IPv6Config.Mode
		ipv6Addr = iface.IPv6Config.Address
		ipv6Gw = iface.IPv6Config.Gateway
	}
	vlanParent, vlanTag := "", 0
	if iface.VLANConfig != nil {
		vlanParent = iface.VLANConfig.Parent
		vlanTag = iface.VLANConfig.Tag
	}
	bondMode := ""
	if iface.BondConfig != nil {
		bondMode = iface.BondConfig.Mode
	}
	bridgeSTP := 0
	if iface.BridgeConfig != nil && iface.BridgeConfig.STP {
		bridgeSTP = 1
	}
	pppoeParent, pppoeUser, pppoePass := "", "", ""
	if iface.PPPoEConfig != nil {
		pppoeParent = iface.PPPoEConfig.Parent
		pppoeUser = iface.PPPoEConfig.Username
		pppoePass = iface.PPPoEConfig.Password
	}

	_, err := r.db.ExecContext(ctx,
		`INSERT INTO network_interfaces
		(id, name, alias_name, if_type, enabled, description, mtu, mac_address,
		ipv4_mode, ipv4_address, ipv4_gateway, ipv6_mode, ipv6_address, ipv6_gateway,
		vlan_parent, vlan_tag, bond_members, bond_mode, bridge_members, bridge_stp,
		pppoe_parent, pppoe_user, pppoe_pass, created_at, updated_at)
		VALUES (?,?,?,?,?,?,?,?, ?,?,?,?,?,?, ?,?,?,?,?,?, ?,?,?,?,?)`,
		iface.ID, iface.Name, iface.Alias, string(iface.Type), boolToInt(iface.Enabled),
		iface.Description, iface.MTU, iface.MAC,
		ipv4Mode, ipv4Addr, ipv4Gw, ipv6Mode, ipv6Addr, ipv6Gw,
		vlanParent, vlanTag, string(bondMembers), bondMode, string(bridgeMembers), bridgeSTP,
		pppoeParent, pppoeUser, pppoePass, iface.CreatedAt, iface.UpdatedAt)
	return err
}

func (r *InterfaceRepo) Update(ctx context.Context, iface *network.Interface) error {
	bondMembers, _ := json.Marshal(ifaceSlice(iface.BondConfig))
	bridgeMembers, _ := json.Marshal(bridgeSlice(iface.BridgeConfig))

	ipv4Mode, ipv4Addr, ipv4Gw := "", "", ""
	if iface.IPv4Config != nil {
		ipv4Mode = iface.IPv4Config.Mode
		ipv4Addr = iface.IPv4Config.Address
		ipv4Gw = iface.IPv4Config.Gateway
	}
	ipv6Mode, ipv6Addr, ipv6Gw := "", "", ""
	if iface.IPv6Config != nil {
		ipv6Mode = iface.IPv6Config.Mode
		ipv6Addr = iface.IPv6Config.Address
		ipv6Gw = iface.IPv6Config.Gateway
	}
	vlanParent, vlanTag := "", 0
	if iface.VLANConfig != nil {
		vlanParent = iface.VLANConfig.Parent
		vlanTag = iface.VLANConfig.Tag
	}
	bondMode := ""
	if iface.BondConfig != nil {
		bondMode = iface.BondConfig.Mode
	}
	bridgeSTP := 0
	if iface.BridgeConfig != nil && iface.BridgeConfig.STP {
		bridgeSTP = 1
	}
	pppoeParent, pppoeUser, pppoePass := "", "", ""
	if iface.PPPoEConfig != nil {
		pppoeParent = iface.PPPoEConfig.Parent
		pppoeUser = iface.PPPoEConfig.Username
		pppoePass = iface.PPPoEConfig.Password
	}

	_, err := r.db.ExecContext(ctx,
		`UPDATE network_interfaces SET
		name=?, alias_name=?, if_type=?, enabled=?, description=?, mtu=?, mac_address=?,
		ipv4_mode=?, ipv4_address=?, ipv4_gateway=?, ipv6_mode=?, ipv6_address=?, ipv6_gateway=?,
		vlan_parent=?, vlan_tag=?, bond_members=?, bond_mode=?, bridge_members=?, bridge_stp=?,
		pppoe_parent=?, pppoe_user=?, pppoe_pass=?, updated_at=?
		WHERE id=?`,
		iface.Name, iface.Alias, string(iface.Type), boolToInt(iface.Enabled), iface.Description, iface.MTU, iface.MAC,
		ipv4Mode, ipv4Addr, ipv4Gw, ipv6Mode, ipv6Addr, ipv6Gw,
		vlanParent, vlanTag, string(bondMembers), bondMode, string(bridgeMembers), bridgeSTP,
		pppoeParent, pppoeUser, pppoePass, iface.UpdatedAt, iface.ID)
	return err
}

func (r *InterfaceRepo) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM network_interfaces WHERE id = ?", id)
	return err
}

func scanInterface(rows *sql.Rows) (*network.Interface, error) {
	var i network.Interface
	var enabled, vlanTag, bridgeSTP int
	var bondMembers, bridgeMembers string
	var ipv4Mode, ipv4Addr, ipv4Gw, ipv6Mode, ipv6Addr, ipv6Gw string
	var vlanParent, bondMode, pppoeParent, pppoeUser, pppoePass string
	var createdAt, updatedAt string

	err := rows.Scan(&i.ID, &i.Name, &i.Alias, &i.Type, &enabled, &i.Description, &i.MTU, &i.MAC,
		&ipv4Mode, &ipv4Addr, &ipv4Gw, &ipv6Mode, &ipv6Addr, &ipv6Gw,
		&vlanParent, &vlanTag, &bondMembers, &bondMode, &bridgeMembers, &bridgeSTP,
		&pppoeParent, &pppoeUser, &pppoePass, &createdAt, &updatedAt)
	if err != nil {
		return nil, err
	}

	i.Enabled = enabled == 1
	i.CreatedAt = parseTime(createdAt)
	i.UpdatedAt = parseTime(updatedAt)
	if ipv4Mode != "" {
		i.IPv4Config = &network.IPv4Config{Mode: ipv4Mode, Address: ipv4Addr, Gateway: ipv4Gw}
	}
	if ipv6Mode != "" {
		i.IPv6Config = &network.IPv6Config{Mode: ipv6Mode, Address: ipv6Addr, Gateway: ipv6Gw}
	}
	if vlanParent != "" {
		i.VLANConfig = &network.VLANConfig{Parent: vlanParent, Tag: vlanTag}
	}
	if bondMode != "" {
		var members []string
		json.Unmarshal([]byte(bondMembers), &members)
		i.BondConfig = &network.BondConfig{Members: members, Mode: bondMode}
	}
	if bridgeMembers != "[]" && bridgeMembers != "" {
		var members []string
		json.Unmarshal([]byte(bridgeMembers), &members)
		if len(members) > 0 {
			i.BridgeConfig = &network.BridgeConfig{Members: members, STP: bridgeSTP == 1}
		}
	}
	if pppoeParent != "" {
		i.PPPoEConfig = &network.PPPoEConfig{Parent: pppoeParent, Username: pppoeUser, Password: pppoePass}
	}

	return &i, nil
}

func scanInterfaceRow(row *sql.Row) (*network.Interface, error) {
	var i network.Interface
	var enabled, vlanTag, bridgeSTP int
	var bondMembers, bridgeMembers string
	var ipv4Mode, ipv4Addr, ipv4Gw, ipv6Mode, ipv6Addr, ipv6Gw string
	var vlanParent, bondMode, pppoeParent, pppoeUser, pppoePass string
	var createdAt, updatedAt string

	err := row.Scan(&i.ID, &i.Name, &i.Alias, &i.Type, &enabled, &i.Description, &i.MTU, &i.MAC,
		&ipv4Mode, &ipv4Addr, &ipv4Gw, &ipv6Mode, &ipv6Addr, &ipv6Gw,
		&vlanParent, &vlanTag, &bondMembers, &bondMode, &bridgeMembers, &bridgeSTP,
		&pppoeParent, &pppoeUser, &pppoePass, &createdAt, &updatedAt)
	if err != nil {
		return nil, err
	}
	i.Enabled = enabled == 1
	i.CreatedAt = parseTime(createdAt)
	i.UpdatedAt = parseTime(updatedAt)
	if ipv4Mode != "" {
		i.IPv4Config = &network.IPv4Config{Mode: ipv4Mode, Address: ipv4Addr, Gateway: ipv4Gw}
	}
	if ipv6Mode != "" {
		i.IPv6Config = &network.IPv6Config{Mode: ipv6Mode, Address: ipv6Addr, Gateway: ipv6Gw}
	}
	if vlanParent != "" {
		i.VLANConfig = &network.VLANConfig{Parent: vlanParent, Tag: vlanTag}
	}
	if bondMode != "" {
		var members []string
		json.Unmarshal([]byte(bondMembers), &members)
		i.BondConfig = &network.BondConfig{Members: members, Mode: bondMode}
	}
	if bridgeMembers != "[]" && bridgeMembers != "" {
		var members []string
		json.Unmarshal([]byte(bridgeMembers), &members)
		if len(members) > 0 {
			i.BridgeConfig = &network.BridgeConfig{Members: members, STP: bridgeSTP == 1}
		}
	}
	if pppoeParent != "" {
		i.PPPoEConfig = &network.PPPoEConfig{Parent: pppoeParent, Username: pppoeUser, Password: pppoePass}
	}
	return &i, nil
}

func ifaceSlice(cfg *network.BondConfig) []string {
	if cfg == nil {
		return []string{}
	}
	return cfg.Members
}
func bridgeSlice(cfg *network.BridgeConfig) []string {
	if cfg == nil {
		return []string{}
	}
	return cfg.Members
}

var _ network.InterfaceRepository = (*InterfaceRepo)(nil)

// ═══════════════════════════════════════════════════════════════
// Route Repository
// ═══════════════════════════════════════════════════════════════

type RouteRepo struct {
	db *database.DB
}

func NewRouteRepo(db *database.DB) *RouteRepo {
	return &RouteRepo{db: db}
}

func (r *RouteRepo) List(ctx context.Context) ([]network.Route, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, destination, gateway, interface_name, metric, route_table, route_type, enabled, description, created_at, updated_at
		FROM network_routes ORDER BY metric ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var routes []network.Route
	for rows.Next() {
		var rt network.Route
		var enabled int
		var createdAt, updatedAt string
		err := rows.Scan(&rt.ID, &rt.Destination, &rt.Gateway, &rt.Interface, &rt.Metric,
			&rt.Table, &rt.Type, &enabled, &rt.Description, &createdAt, &updatedAt)
		if err != nil {
			return nil, err
		}
		rt.Enabled = enabled == 1
		rt.CreatedAt = parseTime(createdAt)
		rt.UpdatedAt = parseTime(updatedAt)
		routes = append(routes, rt)
	}
	return routes, rows.Err()
}

func (r *RouteRepo) GetByID(ctx context.Context, id string) (*network.Route, error) {
	var rt network.Route
	var enabled int
	var createdAt, updatedAt string
	err := r.db.QueryRowContext(ctx,
		`SELECT id, destination, gateway, interface_name, metric, route_table, route_type, enabled, description, created_at, updated_at
		FROM network_routes WHERE id = ?`, id).
		Scan(&rt.ID, &rt.Destination, &rt.Gateway, &rt.Interface, &rt.Metric,
			&rt.Table, &rt.Type, &enabled, &rt.Description, &createdAt, &updatedAt)
	if err != nil {
		return nil, err
	}
	rt.Enabled = enabled == 1
	rt.CreatedAt = parseTime(createdAt)
	rt.UpdatedAt = parseTime(updatedAt)
	return &rt, nil
}

func (r *RouteRepo) Create(ctx context.Context, route *network.Route) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO network_routes (id, destination, gateway, interface_name, metric, route_table, route_type, enabled, description)
		VALUES (?,?,?,?,?,?,?,?,?)`,
		route.ID, route.Destination, route.Gateway, route.Interface, route.Metric,
		route.Table, string(route.Type), boolToInt(route.Enabled), route.Description)
	return err
}

func (r *RouteRepo) Update(ctx context.Context, route *network.Route) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE network_routes SET destination=?, gateway=?, interface_name=?, metric=?,
		route_table=?, route_type=?, enabled=?, description=?, updated_at=datetime('now') WHERE id=?`,
		route.Destination, route.Gateway, route.Interface, route.Metric,
		route.Table, string(route.Type), boolToInt(route.Enabled), route.Description, route.ID)
	return err
}

func (r *RouteRepo) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM network_routes WHERE id = ?", id)
	return err
}

var _ network.RouteRepository = (*RouteRepo)(nil)

// ═══════════════════════════════════════════════════════════════
// Gateway Repository
// ═══════════════════════════════════════════════════════════════

type GatewayRepo struct {
	db *database.DB
}

func NewGatewayRepo(db *database.DB) *GatewayRepo {
	return &GatewayRepo{db: db}
}

func (r *GatewayRepo) List(ctx context.Context) ([]network.Gateway, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, name, interface_name, address, protocol, monitor_ip, weight, priority,
		is_default, enabled, description, monitor_interval, loss_threshold, latency_threshold,
		down_count, monitor_method, created_at, updated_at
		FROM network_gateways ORDER BY priority ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var gateways []network.Gateway
	for rows.Next() {
		var gw network.Gateway
		var isDefault, enabled int
		var createdAt, updatedAt string
		err := rows.Scan(&gw.ID, &gw.Name, &gw.Interface, &gw.Address, &gw.Protocol,
			&gw.MonitorIP, &gw.Weight, &gw.Priority, &isDefault, &enabled, &gw.Description,
			&gw.MonitorConfig.Interval, &gw.MonitorConfig.LossThreshold, &gw.MonitorConfig.LatencyThreshold,
			&gw.MonitorConfig.DownCount, &gw.MonitorConfig.Method, &createdAt, &updatedAt)
		if err != nil {
			return nil, err
		}
		gw.IsDefault = isDefault == 1
		gw.Enabled = enabled == 1
		gw.CreatedAt = parseTime(createdAt)
		gw.UpdatedAt = parseTime(updatedAt)
		gateways = append(gateways, gw)
	}
	return gateways, rows.Err()
}

func (r *GatewayRepo) GetByID(ctx context.Context, id string) (*network.Gateway, error) {
	var gw network.Gateway
	var isDefault, enabled int
	var createdAt, updatedAt string
	err := r.db.QueryRowContext(ctx,
		`SELECT id, name, interface_name, address, protocol, monitor_ip, weight, priority,
		is_default, enabled, description, monitor_interval, loss_threshold, latency_threshold,
		down_count, monitor_method, created_at, updated_at
		FROM network_gateways WHERE id = ?`, id).
		Scan(&gw.ID, &gw.Name, &gw.Interface, &gw.Address, &gw.Protocol,
			&gw.MonitorIP, &gw.Weight, &gw.Priority, &isDefault, &enabled, &gw.Description,
			&gw.MonitorConfig.Interval, &gw.MonitorConfig.LossThreshold, &gw.MonitorConfig.LatencyThreshold,
			&gw.MonitorConfig.DownCount, &gw.MonitorConfig.Method, &createdAt, &updatedAt)
	if err != nil {
		return nil, err
	}
	gw.IsDefault = isDefault == 1
	gw.Enabled = enabled == 1
	gw.CreatedAt = parseTime(createdAt)
	gw.UpdatedAt = parseTime(updatedAt)
	return &gw, nil
}

func (r *GatewayRepo) Create(ctx context.Context, gw *network.Gateway) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO network_gateways (id, name, interface_name, address, protocol, monitor_ip,
		weight, priority, is_default, enabled, description,
		monitor_interval, loss_threshold, latency_threshold, down_count, monitor_method)
		VALUES (?,?,?,?,?,?, ?,?,?,?,?, ?,?,?,?,?)`,
		gw.ID, gw.Name, gw.Interface, gw.Address, gw.Protocol, gw.MonitorIP,
		gw.Weight, gw.Priority, boolToInt(gw.IsDefault), boolToInt(gw.Enabled), gw.Description,
		gw.MonitorConfig.Interval, gw.MonitorConfig.LossThreshold, gw.MonitorConfig.LatencyThreshold,
		gw.MonitorConfig.DownCount, gw.MonitorConfig.Method)
	return err
}

func (r *GatewayRepo) Update(ctx context.Context, gw *network.Gateway) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE network_gateways SET name=?, interface_name=?, address=?, protocol=?, monitor_ip=?,
		weight=?, priority=?, is_default=?, enabled=?, description=?,
		monitor_interval=?, loss_threshold=?, latency_threshold=?, down_count=?, monitor_method=?,
		updated_at=datetime('now') WHERE id=?`,
		gw.Name, gw.Interface, gw.Address, gw.Protocol, gw.MonitorIP,
		gw.Weight, gw.Priority, boolToInt(gw.IsDefault), boolToInt(gw.Enabled), gw.Description,
		gw.MonitorConfig.Interval, gw.MonitorConfig.LossThreshold, gw.MonitorConfig.LatencyThreshold,
		gw.MonitorConfig.DownCount, gw.MonitorConfig.Method, gw.ID)
	return err
}

func (r *GatewayRepo) Delete(ctx context.Context, id string) error {
	res, err := r.db.ExecContext(ctx, "DELETE FROM network_gateways WHERE id = ?", id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("gateway %s not found", id)
	}
	return nil
}

var _ network.GatewayRepository = (*GatewayRepo)(nil)

// ═══════════════════════════════════════════════════════════════
// Gateway Group Repository
// ═══════════════════════════════════════════════════════════════

type GatewayGroupRepo struct {
	db *database.DB
}

func NewGatewayGroupRepo(db *database.DB) *GatewayGroupRepo {
	return &GatewayGroupRepo{db: db}
}

func (r *GatewayGroupRepo) List(ctx context.Context) ([]network.GatewayGroup, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, name, members, trigger_level, description FROM network_gateway_groups ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var groups []network.GatewayGroup
	for rows.Next() {
		group, err := scanGatewayGroup(rows)
		if err != nil {
			return nil, err
		}
		groups = append(groups, *group)
	}
	return groups, rows.Err()
}

func (r *GatewayGroupRepo) GetByID(ctx context.Context, id string) (*network.GatewayGroup, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, name, members, trigger_level, description FROM network_gateway_groups WHERE id = ?`, id)
	return scanGatewayGroupRow(row)
}

func (r *GatewayGroupRepo) Create(ctx context.Context, group *network.GatewayGroup) error {
	members, _ := json.Marshal(group.Members)
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO network_gateway_groups (id, name, members, trigger_level, description)
		VALUES (?,?,?,?,?)`,
		group.ID, group.Name, string(members), group.Trigger, group.Description)
	return err
}

func (r *GatewayGroupRepo) Update(ctx context.Context, group *network.GatewayGroup) error {
	members, _ := json.Marshal(group.Members)
	_, err := r.db.ExecContext(ctx,
		`UPDATE network_gateway_groups SET name=?, members=?, trigger_level=?, description=? WHERE id=?`,
		group.Name, string(members), group.Trigger, group.Description, group.ID)
	return err
}

func (r *GatewayGroupRepo) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM network_gateway_groups WHERE id = ?", id)
	return err
}

func scanGatewayGroup(rows *sql.Rows) (*network.GatewayGroup, error) {
	var group network.GatewayGroup
	var members string
	if err := rows.Scan(&group.ID, &group.Name, &members, &group.Trigger, &group.Description); err != nil {
		return nil, err
	}
	json.Unmarshal([]byte(members), &group.Members)
	return &group, nil
}

func scanGatewayGroupRow(row *sql.Row) (*network.GatewayGroup, error) {
	var group network.GatewayGroup
	var members string
	if err := row.Scan(&group.ID, &group.Name, &members, &group.Trigger, &group.Description); err != nil {
		return nil, err
	}
	json.Unmarshal([]byte(members), &group.Members)
	return &group, nil
}

var _ network.GatewayGroupRepository = (*GatewayGroupRepo)(nil)
