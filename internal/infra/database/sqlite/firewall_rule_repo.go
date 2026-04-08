package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/nixguard/nixguard/internal/domain/firewall"
	"github.com/nixguard/nixguard/internal/infra/database"
)

// RuleRepo implements firewall.RuleRepository using SQLite.
type RuleRepo struct {
	db *database.DB
}

func NewRuleRepo(db *database.DB) *RuleRepo {
	return &RuleRepo{db: db}
}

func (r *RuleRepo) List(ctx context.Context, f firewall.RuleFilter) ([]firewall.Rule, error) {
	query := `SELECT id, interface_name, direction, action, protocol,
		source_type, source_value, source_port, source_not,
		dest_type, dest_value, dest_port, dest_not,
		log_enabled, description, enabled, rule_order, category,
		is_floating, floating_ifaces, gateway, state_type, max_states,
		tag, tagged, schedule_name, schedule_start, schedule_end, schedule_days,
		created_at, updated_at
		FROM firewall_rules WHERE 1=1`

	var args []interface{}

	if f.Interface != "" {
		query += " AND interface_name = ?"
		args = append(args, f.Interface)
	}
	if f.Action != "" {
		query += " AND action = ?"
		args = append(args, string(f.Action))
	}
	if f.Protocol != "" {
		query += " AND protocol = ?"
		args = append(args, string(f.Protocol))
	}
	if f.Enabled != nil {
		query += " AND enabled = ?"
		if *f.Enabled {
			args = append(args, 1)
		} else {
			args = append(args, 0)
		}
	}
	if f.IsFloating != nil {
		query += " AND is_floating = ?"
		if *f.IsFloating {
			args = append(args, 1)
		} else {
			args = append(args, 0)
		}
	}
	if f.Category != "" {
		query += " AND category = ?"
		args = append(args, f.Category)
	}
	if f.Search != "" {
		query += " AND (description LIKE ? OR interface_name LIKE ?)"
		s := "%" + f.Search + "%"
		args = append(args, s, s)
	}

	query += " ORDER BY rule_order ASC"

	if f.Limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", f.Limit)
		if f.Offset > 0 {
			query += fmt.Sprintf(" OFFSET %d", f.Offset)
		}
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query rules: %w", err)
	}
	defer rows.Close()

	return scanRules(rows)
}

func (r *RuleRepo) GetByID(ctx context.Context, id string) (*firewall.Rule, error) {
	rules, err := r.queryRules(ctx, "WHERE id = ?", id)
	if err != nil {
		return nil, err
	}
	if len(rules) == 0 {
		return nil, fmt.Errorf("rule %s not found", id)
	}
	return &rules[0], nil
}

func (r *RuleRepo) Create(ctx context.Context, rule *firewall.Rule) error {
	floatingIfaces, _ := json.Marshal(rule.Interfaces)
	scheduleDays := "[]"
	if rule.Schedule != nil {
		d, _ := json.Marshal(rule.Schedule.Weekdays)
		scheduleDays = string(d)
	}

	schedName, schedStart, schedEnd := "", "", ""
	if rule.Schedule != nil {
		schedName = rule.Schedule.Name
		schedStart = rule.Schedule.StartTime
		schedEnd = rule.Schedule.EndTime
	}

	_, err := r.db.ExecContext(ctx, `INSERT INTO firewall_rules
		(id, interface_name, direction, action, protocol,
		source_type, source_value, source_port, source_not,
		dest_type, dest_value, dest_port, dest_not,
		log_enabled, description, enabled, rule_order, category,
		is_floating, floating_ifaces, gateway, state_type, max_states,
		tag, tagged, schedule_name, schedule_start, schedule_end, schedule_days,
		created_at, updated_at)
		VALUES (?,?,?,?,?, ?,?,?,?, ?,?,?,?, ?,?,?,?,?, ?,?,?,?,?, ?,?,?,?,?,?, ?,?)`,
		rule.ID, rule.Interface, string(rule.Direction), string(rule.Action), string(rule.Protocol),
		string(rule.Source.Type), rule.Source.Value, rule.Source.Port, boolToInt(rule.Source.Not),
		string(rule.Destination.Type), rule.Destination.Value, rule.Destination.Port, boolToInt(rule.Destination.Not),
		boolToInt(rule.Log), rule.Description, boolToInt(rule.Enabled), rule.Order, rule.Category,
		boolToInt(rule.IsFloating), string(floatingIfaces), rule.Gateway, rule.StateType, rule.MaxStates,
		rule.Tag, rule.Tagged, schedName, schedStart, schedEnd, scheduleDays,
		rule.CreatedAt, rule.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert rule: %w", err)
	}
	return nil
}

func (r *RuleRepo) Update(ctx context.Context, rule *firewall.Rule) error {
	floatingIfaces, _ := json.Marshal(rule.Interfaces)
	scheduleDays := "[]"
	if rule.Schedule != nil {
		d, _ := json.Marshal(rule.Schedule.Weekdays)
		scheduleDays = string(d)
	}
	schedName, schedStart, schedEnd := "", "", ""
	if rule.Schedule != nil {
		schedName = rule.Schedule.Name
		schedStart = rule.Schedule.StartTime
		schedEnd = rule.Schedule.EndTime
	}

	_, err := r.db.ExecContext(ctx, `UPDATE firewall_rules SET
		interface_name=?, direction=?, action=?, protocol=?,
		source_type=?, source_value=?, source_port=?, source_not=?,
		dest_type=?, dest_value=?, dest_port=?, dest_not=?,
		log_enabled=?, description=?, enabled=?, rule_order=?, category=?,
		is_floating=?, floating_ifaces=?, gateway=?, state_type=?, max_states=?,
		tag=?, tagged=?, schedule_name=?, schedule_start=?, schedule_end=?, schedule_days=?,
		updated_at=?
		WHERE id=?`,
		rule.Interface, string(rule.Direction), string(rule.Action), string(rule.Protocol),
		string(rule.Source.Type), rule.Source.Value, rule.Source.Port, boolToInt(rule.Source.Not),
		string(rule.Destination.Type), rule.Destination.Value, rule.Destination.Port, boolToInt(rule.Destination.Not),
		boolToInt(rule.Log), rule.Description, boolToInt(rule.Enabled), rule.Order, rule.Category,
		boolToInt(rule.IsFloating), string(floatingIfaces), rule.Gateway, rule.StateType, rule.MaxStates,
		rule.Tag, rule.Tagged, schedName, schedStart, schedEnd, scheduleDays,
		rule.UpdatedAt, rule.ID,
	)
	return err
}

func (r *RuleRepo) Delete(ctx context.Context, id string) error {
	res, err := r.db.ExecContext(ctx, "DELETE FROM firewall_rules WHERE id = ?", id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("rule %s not found", id)
	}
	return nil
}

func (r *RuleRepo) Reorder(ctx context.Context, ruleIDs []string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, "UPDATE firewall_rules SET rule_order = ? WHERE id = ?")
	if err != nil {
		return err
	}
	defer stmt.Close()

	for i, id := range ruleIDs {
		if _, err := stmt.ExecContext(ctx, i, id); err != nil {
			return fmt.Errorf("reorder rule %s: %w", id, err)
		}
	}
	return tx.Commit()
}

func (r *RuleRepo) GetByInterface(ctx context.Context, iface string) ([]firewall.Rule, error) {
	return r.queryRules(ctx, "WHERE interface_name = ? AND enabled = 1 ORDER BY rule_order ASC", iface)
}

// ── internal helpers ───────────────────────────────────────────

func (r *RuleRepo) queryRules(ctx context.Context, where string, args ...interface{}) ([]firewall.Rule, error) {
	query := `SELECT id, interface_name, direction, action, protocol,
		source_type, source_value, source_port, source_not,
		dest_type, dest_value, dest_port, dest_not,
		log_enabled, description, enabled, rule_order, category,
		is_floating, floating_ifaces, gateway, state_type, max_states,
		tag, tagged, schedule_name, schedule_start, schedule_end, schedule_days,
		created_at, updated_at
		FROM firewall_rules ` + where

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanRules(rows)
}

func scanRules(rows *sql.Rows) ([]firewall.Rule, error) {
	var rules []firewall.Rule
	for rows.Next() {
		var rule firewall.Rule
		var srcNot, dstNot, logE, enabled, isFloating int
		var floatingIfaces, scheduleDays string
		var schedName, schedStart, schedEnd string
		var createdAt, updatedAt string

		err := rows.Scan(
			&rule.ID, &rule.Interface, &rule.Direction, &rule.Action, &rule.Protocol,
			&rule.Source.Type, &rule.Source.Value, &rule.Source.Port, &srcNot,
			&rule.Destination.Type, &rule.Destination.Value, &rule.Destination.Port, &dstNot,
			&logE, &rule.Description, &enabled, &rule.Order, &rule.Category,
			&isFloating, &floatingIfaces, &rule.Gateway, &rule.StateType, &rule.MaxStates,
			&rule.Tag, &rule.Tagged, &schedName, &schedStart, &schedEnd, &scheduleDays,
			&createdAt, &updatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan rule: %w", err)
		}

		rule.Source.Not = srcNot == 1
		rule.Destination.Not = dstNot == 1
		rule.Log = logE == 1
		rule.Enabled = enabled == 1
		rule.IsFloating = isFloating == 1
		rule.CreatedAt = parseTime(createdAt)
		rule.UpdatedAt = parseTime(updatedAt)

		json.Unmarshal([]byte(floatingIfaces), &rule.Interfaces)

		if schedName != "" {
			rule.Schedule = &firewall.Schedule{
				Name:      schedName,
				StartTime: schedStart,
				EndTime:   schedEnd,
			}
			json.Unmarshal([]byte(scheduleDays), &rule.Schedule.Weekdays)
		}

		rules = append(rules, rule)
	}
	return rules, rows.Err()
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// parseTime parses SQLite datetime strings into time.Time.
func parseTime(s string) time.Time {
	formats := []string{
		"2006-01-02T15:04:05Z",
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05.999999999Z07:00",
		"2006-01-02T15:04:05-07:00",
		time.RFC3339,
		time.RFC3339Nano,
	}
	for _, f := range formats {
		if t, err := time.Parse(f, s); err == nil {
			return t
		}
	}
	return time.Time{}
}

// Ensure interface compliance
var _ firewall.RuleRepository = (*RuleRepo)(nil)

// ── Alias Repository ───────────────────────────────────────────

type AliasRepo struct {
	db *database.DB
}

func NewAliasRepo(db *database.DB) *AliasRepo {
	return &AliasRepo{db: db}
}

func (r *AliasRepo) List(ctx context.Context) ([]firewall.Alias, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, name, alias_type, description, entries, update_freq, enabled, created_at, updated_at
		FROM firewall_aliases ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanAliases(rows)
}

func (r *AliasRepo) GetByID(ctx context.Context, id string) (*firewall.Alias, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, name, alias_type, description, entries, update_freq, enabled, created_at, updated_at
		FROM firewall_aliases WHERE id = ?`, id)
	return scanAlias(row)
}

func (r *AliasRepo) GetByName(ctx context.Context, name string) (*firewall.Alias, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, name, alias_type, description, entries, update_freq, enabled, created_at, updated_at
		FROM firewall_aliases WHERE name = ?`, name)
	return scanAlias(row)
}

func (r *AliasRepo) Create(ctx context.Context, alias *firewall.Alias) error {
	entries, _ := json.Marshal(alias.Entries)
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO firewall_aliases (id, name, alias_type, description, entries, update_freq, enabled, created_at, updated_at)
		VALUES (?,?,?,?,?,?,?,?,?)`,
		alias.ID, alias.Name, string(alias.Type), alias.Description,
		string(entries), alias.UpdateFreq, boolToInt(alias.Enabled),
		alias.CreatedAt, alias.UpdatedAt)
	return err
}

func (r *AliasRepo) Update(ctx context.Context, alias *firewall.Alias) error {
	entries, _ := json.Marshal(alias.Entries)
	_, err := r.db.ExecContext(ctx,
		`UPDATE firewall_aliases SET name=?, alias_type=?, description=?, entries=?, update_freq=?, enabled=?, updated_at=?
		WHERE id=?`,
		alias.Name, string(alias.Type), alias.Description,
		string(entries), alias.UpdateFreq, boolToInt(alias.Enabled),
		alias.UpdatedAt, alias.ID)
	return err
}

func (r *AliasRepo) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM firewall_aliases WHERE id = ?", id)
	return err
}

func scanAliases(rows *sql.Rows) ([]firewall.Alias, error) {
	var aliases []firewall.Alias
	for rows.Next() {
		var a firewall.Alias
		var entries string
		var enabled int
		var createdAt, updatedAt string
		err := rows.Scan(&a.ID, &a.Name, &a.Type, &a.Description, &entries, &a.UpdateFreq, &enabled, &createdAt, &updatedAt)
		if err != nil {
			return nil, err
		}
		a.Enabled = enabled == 1
		a.CreatedAt = parseTime(createdAt)
		a.UpdatedAt = parseTime(updatedAt)
		json.Unmarshal([]byte(entries), &a.Entries)
		aliases = append(aliases, a)
	}
	return aliases, rows.Err()
}

func scanAlias(row *sql.Row) (*firewall.Alias, error) {
	var a firewall.Alias
	var entries string
	var enabled int
	var createdAt, updatedAt string
	err := row.Scan(&a.ID, &a.Name, &a.Type, &a.Description, &entries, &a.UpdateFreq, &enabled, &createdAt, &updatedAt)
	if err != nil {
		return nil, err
	}
	a.Enabled = enabled == 1
	a.CreatedAt = parseTime(createdAt)
	a.UpdatedAt = parseTime(updatedAt)
	json.Unmarshal([]byte(entries), &a.Entries)
	return &a, nil
}

var _ firewall.AliasRepository = (*AliasRepo)(nil)

// ── NAT Repository ─────────────────────────────────────────────

type NATRepo struct {
	db *database.DB
}

func NewNATRepo(db *database.DB) *NATRepo {
	return &NATRepo{db: db}
}

func (r *NATRepo) List(ctx context.Context, natType firewall.NATType) ([]firewall.NATRule, error) {
	query := `SELECT id, nat_type, interface_name, protocol,
		source_type, source_value, source_port, source_not,
		dest_type, dest_value, dest_port, dest_not,
		redirect_target, redirect_port, description, enabled, nat_reflection, rule_order, created_at
		FROM firewall_nat_rules`
	var args []interface{}
	if natType != "" {
		query += " WHERE nat_type = ?"
		args = append(args, string(natType))
	}
	query += " ORDER BY rule_order ASC"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rules []firewall.NATRule
	for rows.Next() {
		var n firewall.NATRule
		var srcNot, dstNot, enabled, natRefl, order int
		var createdAt string
		err := rows.Scan(&n.ID, &n.Type, &n.Interface, &n.Protocol,
			&n.Source.Type, &n.Source.Value, &n.Source.Port, &srcNot,
			&n.Destination.Type, &n.Destination.Value, &n.Destination.Port, &dstNot,
			&n.RedirectTarget, &n.RedirectPort, &n.Description, &enabled, &natRefl, &order, &createdAt)
		if err != nil {
			return nil, err
		}
		n.Source.Not = srcNot == 1
		n.Destination.Not = dstNot == 1
		n.Enabled = enabled == 1
		n.NATReflection = natRefl == 1
		n.CreatedAt = parseTime(createdAt)
		rules = append(rules, n)
	}
	return rules, rows.Err()
}

func (r *NATRepo) GetByID(ctx context.Context, id string) (*firewall.NATRule, error) {
	rules, err := r.List(ctx, "")
	if err != nil {
		return nil, err
	}
	for _, rule := range rules {
		if rule.ID == id {
			return &rule, nil
		}
	}
	return nil, fmt.Errorf("NAT rule %s not found", id)
}

func (r *NATRepo) Create(ctx context.Context, rule *firewall.NATRule) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO firewall_nat_rules (id, nat_type, interface_name, protocol,
		source_type, source_value, source_port, source_not,
		dest_type, dest_value, dest_port, dest_not,
		redirect_target, redirect_port, description, enabled, nat_reflection, created_at)
		VALUES (?,?,?,?, ?,?,?,?, ?,?,?,?, ?,?,?,?,?,?)`,
		rule.ID, string(rule.Type), rule.Interface, string(rule.Protocol),
		string(rule.Source.Type), rule.Source.Value, rule.Source.Port, boolToInt(rule.Source.Not),
		string(rule.Destination.Type), rule.Destination.Value, rule.Destination.Port, boolToInt(rule.Destination.Not),
		rule.RedirectTarget, rule.RedirectPort, rule.Description, boolToInt(rule.Enabled), boolToInt(rule.NATReflection),
		rule.CreatedAt)
	return err
}

func (r *NATRepo) Update(ctx context.Context, rule *firewall.NATRule) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE firewall_nat_rules SET nat_type=?, interface_name=?, protocol=?,
		source_type=?, source_value=?, source_port=?, source_not=?,
		dest_type=?, dest_value=?, dest_port=?, dest_not=?,
		redirect_target=?, redirect_port=?, description=?, enabled=?, nat_reflection=?,
		updated_at=datetime('now')
		WHERE id=?`,
		string(rule.Type), rule.Interface, string(rule.Protocol),
		string(rule.Source.Type), rule.Source.Value, rule.Source.Port, boolToInt(rule.Source.Not),
		string(rule.Destination.Type), rule.Destination.Value, rule.Destination.Port, boolToInt(rule.Destination.Not),
		rule.RedirectTarget, rule.RedirectPort, rule.Description, boolToInt(rule.Enabled), boolToInt(rule.NATReflection),
		rule.ID)
	return err
}

func (r *NATRepo) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM firewall_nat_rules WHERE id = ?", id)
	return err
}

// Helper used elsewhere
func JoinStrings(s []string) string {
	return strings.Join(s, ",")
}

var _ firewall.NATRepository = (*NATRepo)(nil)
