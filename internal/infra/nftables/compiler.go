// Package nftables — compiler.go contains the nftables ruleset compiler.
// Translates NixGuard domain rules into nft syntax with full support for:
// - Per-interface filtering (input/output/forward)
// - Floating rules (multi-interface)
// - NAT (DNAT, SNAT, masquerade, 1:1)
// - Aliases (sets) with all types (host, network, port, GeoIP)
// - Logging with prefix
// - Connection tracking state
// - Protocol-specific matching (TCP flags, ICMP types)
package nftables

import (
	"fmt"
	"net"
	"strings"

	"github.com/nixguard/nixguard/internal/domain/firewall"
)

const (
	aliasFamilyV4    = "v4"
	aliasFamilyV6    = "v6"
	aliasFamilyMixed = "mixed"
	aliasFamilyPort  = "port"
)

// PredefinedFilters controls built-in network filtering.
type PredefinedFilters struct {
	BogonFilter   bool     // Drop traffic from bogon networks on WAN interfaces
	RFC1918Filter bool     // Drop RFC1918 (private) source addresses on WAN interfaces
	WANInterfaces []string // Interfaces where predefined filters apply
}

// CompileRuleset generates a complete nft-compatible ruleset string.
func CompileRuleset(rules []firewall.Rule, aliases []firewall.Alias, filters *PredefinedFilters) string {
	var b strings.Builder
	aliasFamilies := buildAliasFamilies(aliases)

	b.WriteString("#!/usr/sbin/nft -f\n")
	b.WriteString("# NixGuard nftables ruleset — auto-generated\n")
	b.WriteString("# DO NOT EDIT — managed by nixguard-agent\n\n")

	// Flush existing nixguard tables
	b.WriteString("table inet nixguard\ndelete table inet nixguard\n\n")

	b.WriteString("table inet nixguard {\n")

	// ── Compile aliases into nft sets ──────────────────────
	for _, alias := range aliases {
		if !alias.Enabled || len(alias.Entries) == 0 {
			continue
		}
		compileSet(&b, alias, aliasFamilies[alias.Name])
	}

	// ── Compile predefined filter sets ────────────────────
	if filters != nil {
		compilePredefinedSets(&b, filters)
	}

	// ── Counter for each rule (for stats) ──────────────────
	for _, rule := range rules {
		if !rule.Enabled {
			continue
		}
		b.WriteString(fmt.Sprintf("\tcounter cnt_%s {\n\t\tpackets 0 bytes 0\n\t}\n\n", rule.ID))
	}

	// ── Input chain ────────────────────────────────────────
	// Check if any user rules target the loopback interface
	hasLoopbackRules := false
	for _, rule := range rules {
		if rule.Enabled && rule.Interface == "lo" {
			hasLoopbackRules = true
			break
		}
	}

	b.WriteString("\tchain input {\n")
	b.WriteString("\t\ttype filter hook input priority filter; policy drop;\n\n")
	b.WriteString("\t\t# Connection tracking\n")
	b.WriteString("\t\tct state established,related accept\n")
	b.WriteString("\t\tct state invalid drop\n\n")
	b.WriteString("\t\t# Fragment reassembly (nf_defrag auto-loaded by ct)\n")
	b.WriteString("\t\tip frag-off & 0x1fff != 0 counter drop\n\n")
	b.WriteString("\t\t# Reverse path filtering (anti-spoofing)\n")
	b.WriteString("\t\tfib saddr . iif oif missing drop\n\n")

	// Predefined filter rules (bogon/RFC1918) on WAN interfaces
	if filters != nil {
		compilePredefinedFilterRules(&b, filters)
	}

	if hasLoopbackRules {
		// Emit user rules for loopback BEFORE the blanket accept,
		// so block/reject rules on lo are actually evaluated.
		for _, rule := range rules {
			if rule.Enabled && rule.Direction == firewall.DirectionIn && !rule.IsFloating && rule.Interface == "lo" {
				compileRule(&b, rule, "input", aliasFamilies)
			}
		}
	}

	b.WriteString("\t\t# Loopback\n")
	b.WriteString("\t\tiif \"lo\" accept\n\n")
	b.WriteString("\t\t# ICMP/ICMPv6 (rate limited)\n")
	b.WriteString("\t\tip protocol icmp limit rate 25/second accept\n")
	b.WriteString("\t\tip6 nexthdr icmpv6 limit rate 25/second accept\n\n")

	for _, rule := range rules {
		if rule.Enabled && rule.Direction == firewall.DirectionIn && !rule.IsFloating && rule.Interface != "lo" {
			compileRule(&b, rule, "input", aliasFamilies)
		}
	}
	b.WriteString("\t}\n\n")

	// ── Forward chain ──────────────────────────────────────
	b.WriteString("\tchain forward {\n")
	b.WriteString("\t\ttype filter hook forward priority filter; policy drop;\n\n")
	b.WriteString("\t\tct state established,related accept\n")
	b.WriteString("\t\tct state invalid drop\n\n")
	b.WriteString("\t\t# Fragment reassembly (nf_defrag auto-loaded by ct)\n")
	b.WriteString("\t\tip frag-off & 0x1fff != 0 counter drop\n\n")
	b.WriteString("\t\t# Reverse path filtering (anti-spoofing)\n")
	b.WriteString("\t\tfib saddr . iif oif missing drop\n\n")

	// Predefined filter rules (bogon/RFC1918) on WAN interfaces
	if filters != nil {
		compilePredefinedFilterRules(&b, filters)
	}

	// Floating rules apply to forward chain
	for _, rule := range rules {
		if rule.Enabled && rule.IsFloating {
			compileRule(&b, rule, "forward", aliasFamilies)
		}
	}
	// Non-floating forward rules
	for _, rule := range rules {
		if rule.Enabled && !rule.IsFloating && rule.Direction == firewall.DirectionIn {
			// Generate forward rules for inter-interface traffic
			if rule.Interface != "" {
				compileForwardRule(&b, rule, aliasFamilies)
			}
		}
	}
	b.WriteString("\t}\n\n")

	// ── Output chain ───────────────────────────────────────
	b.WriteString("\tchain output {\n")
	b.WriteString("\t\ttype filter hook output priority filter; policy accept;\n\n")

	for _, rule := range rules {
		if rule.Enabled && rule.Direction == firewall.DirectionOut {
			compileRule(&b, rule, "output", aliasFamilies)
		}
	}
	b.WriteString("\t}\n")

	b.WriteString("}\n")

	return b.String()
}

// CompileNATRuleset generates nft NAT rules.
func CompileNATRuleset(rules []firewall.NATRule) string {
	var b strings.Builder

	b.WriteString("#!/usr/sbin/nft -f\n")
	b.WriteString("# NixGuard NAT ruleset — auto-generated\n\n")
	b.WriteString("table inet nixguard_nat\ndelete table inet nixguard_nat\n\n")
	b.WriteString("table inet nixguard_nat {\n")

	// ── DNAT / Port Forwarding ─────────────────────────────
	b.WriteString("\tchain prerouting {\n")
	b.WriteString("\t\ttype nat hook prerouting priority dstnat; policy accept;\n\n")

	for _, r := range rules {
		if !r.Enabled {
			continue
		}
		switch r.Type {
		case firewall.NATPortForward:
			compilePortForward(&b, r)
		case firewall.NATOneToOne:
			compileOneToOneNAT(&b, r)
		}
	}

	for _, r := range rules {
		if !r.Enabled || r.Type != firewall.NATPortForward || !r.NATReflection {
			continue
		}
		compileNATReflectionDNAT(&b, r)
	}
	b.WriteString("\t}\n\n")

	// ── SNAT / Masquerade ──────────────────────────────────
	b.WriteString("\tchain postrouting {\n")
	b.WriteString("\t\ttype nat hook postrouting priority srcnat; policy accept;\n\n")

	for _, r := range rules {
		if !r.Enabled {
			continue
		}
		switch r.Type {
		case firewall.NATOutbound:
			compileOutboundNAT(&b, r)
		case firewall.NATOneToOne:
			compileOneToOneSNAT(&b, r)
		}
	}

	for _, r := range rules {
		if !r.Enabled || r.Type != firewall.NATPortForward || !r.NATReflection {
			continue
		}
		compileNATReflectionSNAT(&b, r)
	}
	b.WriteString("\t}\n")

	b.WriteString("}\n")
	return b.String()
}

// ─── Set Compilation ───────────────────────────────────────────

func compileSet(b *strings.Builder, alias firewall.Alias, family string) {
	if alias.Type == firewall.AliasPort {
		writeSet(b, alias.Name, "inet_service", false, alias.Entries)
		return
	}

	v4Entries, v6Entries := splitEntriesByFamily(alias.Entries)
	v4Interval := entriesUseInterval(v4Entries)
	v6Interval := entriesUseInterval(v6Entries)

	switch family {
	case aliasFamilyMixed:
		if len(v4Entries) > 0 {
			writeSet(b, alias.Name+"_v4", "ipv4_addr", v4Interval, v4Entries)
		}
		if len(v6Entries) > 0 {
			writeSet(b, alias.Name+"_v6", "ipv6_addr", v6Interval, v6Entries)
		}
	case aliasFamilyV6:
		writeSet(b, alias.Name, "ipv6_addr", v6Interval, v6Entries)
	default:
		writeSet(b, alias.Name, "ipv4_addr", v4Interval, v4Entries)
	}
}

// ─── Rule Compilation ──────────────────────────────────────────

func compileRule(b *strings.Builder, rule firewall.Rule, chain string, aliasFamilies map[string]string) {
	b.WriteString(fmt.Sprintf("\t\t# [%s] %s\n", rule.ID, rule.Description))
	b.WriteString("\t\t")

	var parts []string

	// Interface match
	if ifaceMatch := compileInterfaceMatch(rule, chain); ifaceMatch != "" {
		parts = append(parts, ifaceMatch)
	}

	// ── L3: Address matches first (nftables requires L3 before L4) ──

	// Source address
	srcMatch := compileAddress(rule.Source, "saddr", aliasFamilies)
	if srcMatch != "" {
		parts = append(parts, srcMatch)
	}

	// Destination address
	dstMatch := compileAddress(rule.Destination, "daddr", aliasFamilies)
	if dstMatch != "" {
		parts = append(parts, dstMatch)
	}

	// ── L4: Protocol and port matches ──

	// For TCP/UDP with ports, compilePort() emits "tcp dport 9999" as a
	// single expression, so we skip the standalone protocol keyword.
	// For TCP/UDP without ports we still need the protocol match.
	hasPorts := rule.Source.Port != "" || rule.Destination.Port != ""
	if rule.Protocol != firewall.ProtoAny {
		switch rule.Protocol {
		case firewall.ProtoTCP:
			if !hasPorts {
				parts = append(parts, "meta l4proto tcp")
			}
		case firewall.ProtoUDP:
			if !hasPorts {
				parts = append(parts, "meta l4proto udp")
			}
		case firewall.ProtoICMP:
			parts = append(parts, "ip protocol icmp")
		case firewall.ProtoICMPv6:
			parts = append(parts, "ip6 nexthdr icmpv6")
		case firewall.ProtoESP:
			parts = append(parts, "ip protocol esp")
		case firewall.ProtoAH:
			parts = append(parts, "ip protocol ah")
		case firewall.ProtoGRE:
			parts = append(parts, "ip protocol gre")
		}
	}

	// Source port
	srcPort := compilePort(rule.Source, "sport", rule.Protocol)
	if srcPort != "" {
		parts = append(parts, srcPort)
	}

	// Destination port
	dstPort := compilePort(rule.Destination, "dport", rule.Protocol)
	if dstPort != "" {
		parts = append(parts, dstPort)
	}

	// Policy routing: mark packets for gateway selection
	if rule.Gateway != "" && rule.Action == firewall.ActionPass {
		mark := gatewayMark(rule.Gateway)
		parts = append(parts, fmt.Sprintf("meta mark set 0x%x", mark))
	}

	// Counter
	parts = append(parts, fmt.Sprintf("counter name cnt_%s", rule.ID))

	// Log
	if rule.Log {
		prefix := fmt.Sprintf("nixguard_%s_%s", rule.Action, rule.ID[:8])
		parts = append(parts, fmt.Sprintf("log prefix \"%s\" group 1", prefix))
	}

	// Action
	switch rule.Action {
	case firewall.ActionPass:
		parts = append(parts, "accept")
	case firewall.ActionBlock:
		parts = append(parts, "drop")
	case firewall.ActionReject:
		parts = append(parts, "reject")
	}

	b.WriteString(strings.Join(parts, " "))
	b.WriteString("\n")
}

func compileForwardRule(b *strings.Builder, rule firewall.Rule, aliasFamilies map[string]string) {
	b.WriteString(fmt.Sprintf("\t\t# [%s] forward: %s\n", rule.ID, rule.Description))
	b.WriteString("\t\t")

	var parts []string
	if ifaceMatch := compileInterfaceMatch(rule, "forward"); ifaceMatch != "" {
		parts = append(parts, ifaceMatch)
	}

	// L3 address matches first
	srcMatch := compileAddress(rule.Source, "saddr", aliasFamilies)
	if srcMatch != "" {
		parts = append(parts, srcMatch)
	}
	dstMatch := compileAddress(rule.Destination, "daddr", aliasFamilies)
	if dstMatch != "" {
		parts = append(parts, dstMatch)
	}

	// L4 protocol and port
	hasPorts := rule.Source.Port != "" || rule.Destination.Port != ""
	if rule.Protocol != firewall.ProtoAny {
		if !hasPorts || (rule.Protocol != firewall.ProtoTCP && rule.Protocol != firewall.ProtoUDP) {
			parts = append(parts, protocolToNFT(rule.Protocol))
		}
	}
	dstPort := compilePort(rule.Destination, "dport", rule.Protocol)
	if dstPort != "" {
		parts = append(parts, dstPort)
	}

	parts = append(parts, fmt.Sprintf("counter name cnt_%s", rule.ID))

	switch rule.Action {
	case firewall.ActionPass:
		parts = append(parts, "accept")
	case firewall.ActionBlock:
		parts = append(parts, "drop")
	case firewall.ActionReject:
		parts = append(parts, "reject")
	}

	b.WriteString(strings.Join(parts, " "))
	b.WriteString("\n")
}

func compileAddress(addr firewall.Address, field string, aliasFamilies map[string]string) string {
	if addr.Type == firewall.AddrAny || addr.Value == "" {
		return ""
	}

	switch addr.Type {
	case firewall.AddrSingle, firewall.AddrAddress, firewall.AddrNetwork:
		return buildFamilyMatch(detectAddressFamily(addr.Value), field, addr.Value, addr.Not)
	case firewall.AddrAlias:
		return compileAliasAddress(addr.Value, field, addr.Not, aliasFamilies[addr.Value])
	case firewall.AddrGeoIP:
		return buildFamilyMatch(aliasFamilyV4, field, "@geoip_"+strings.ToLower(addr.Value), addr.Not)
	}
	return ""
}

func compilePort(addr firewall.Address, field string, proto firewall.Protocol) string {
	if addr.Port == "" {
		return ""
	}
	if proto != firewall.ProtoTCP && proto != firewall.ProtoUDP {
		return ""
	}

	protoStr := strings.ToLower(string(proto))

	// Port range
	if strings.Contains(addr.Port, "-") {
		return fmt.Sprintf("%s %s %s", protoStr, field, addr.Port)
	}
	// Multiple ports
	if strings.Contains(addr.Port, ",") {
		return fmt.Sprintf("%s %s { %s }", protoStr, field, addr.Port)
	}
	// Alias reference (contains non-digit chars)
	if strings.IndexFunc(addr.Port, func(r rune) bool { return (r < '0' || r > '9') && r != '_' && r != '-' }) != -1 {
		return fmt.Sprintf("%s %s @%s", protoStr, field, addr.Port)
	}
	// Single port
	return fmt.Sprintf("%s %s %s", protoStr, field, addr.Port)
}

func protocolToNFT(p firewall.Protocol) string {
	switch p {
	case firewall.ProtoTCP:
		return "meta l4proto tcp"
	case firewall.ProtoUDP:
		return "meta l4proto udp"
	case firewall.ProtoICMP:
		return "ip protocol icmp"
	case firewall.ProtoICMPv6:
		return "ip6 nexthdr icmpv6"
	default:
		return ""
	}
}

// ─── NAT Compilation ───────────────────────────────────────────

func compilePortForward(b *strings.Builder, r firewall.NATRule) {
	b.WriteString(fmt.Sprintf("\t\t# [%s] Port Forward: %s\n", r.ID, r.Description))
	b.WriteString("\t\t")

	var parts []string
	parts = append(parts, fmt.Sprintf("iifname \"%s\"", r.Interface))

	proto := strings.ToLower(string(r.Protocol))
	if proto != "any" && proto != "" {
		parts = append(parts, proto)
	}

	// Destination match (external IP/port)
	if r.Destination.Value != "" && r.Destination.Type != firewall.AddrAny {
		parts = append(parts, buildFamilyMatch(detectAddressFamily(r.Destination.Value), "daddr", r.Destination.Value, false))
	}
	if r.Destination.Port != "" {
		parts = append(parts, fmt.Sprintf("%s dport %s", proto, r.Destination.Port))
	}

	// DNAT target
	target := formatNATTarget(r.RedirectTarget, r.RedirectPort)
	parts = append(parts, fmt.Sprintf("dnat to %s", target))

	b.WriteString(strings.Join(parts, " "))
	b.WriteString("\n")
}

func compileOneToOneNAT(b *strings.Builder, r firewall.NATRule) {
	b.WriteString(fmt.Sprintf("\t\t# [%s] 1:1 NAT (DNAT): %s\n", r.ID, r.Description))
	b.WriteString(fmt.Sprintf("\t\tiifname \"%s\" %s dnat to %s\n",
		r.Interface, buildFamilyMatch(detectAddressFamily(r.Destination.Value), "daddr", r.Destination.Value, false), formatNATTarget(r.RedirectTarget, "")))
}

func compileOneToOneSNAT(b *strings.Builder, r firewall.NATRule) {
	b.WriteString(fmt.Sprintf("\t\t# [%s] 1:1 NAT (SNAT): %s\n", r.ID, r.Description))
	b.WriteString(fmt.Sprintf("\t\toifname \"%s\" %s snat to %s\n",
		r.Interface, buildFamilyMatch(detectAddressFamily(r.RedirectTarget), "saddr", r.RedirectTarget, false), formatNATTarget(r.Destination.Value, "")))
}

func compileOutboundNAT(b *strings.Builder, r firewall.NATRule) {
	b.WriteString(fmt.Sprintf("\t\t# [%s] Outbound NAT: %s\n", r.ID, r.Description))

	if r.Source.Value != "" && r.Source.Type != firewall.AddrAny {
		b.WriteString(fmt.Sprintf("\t\toifname \"%s\" %s masquerade\n",
			r.Interface, buildFamilyMatch(detectAddressFamily(r.Source.Value), "saddr", r.Source.Value, false)))
	} else {
		b.WriteString(fmt.Sprintf("\t\toifname \"%s\" masquerade\n", r.Interface))
	}
}

func compileNATReflectionDNAT(b *strings.Builder, r firewall.NATRule) {
	if r.ReflectionInterface == "" || r.Destination.Value == "" {
		return
	}
	b.WriteString(fmt.Sprintf("\t\t# [%s] NAT Reflection (DNAT): %s\n", r.ID, r.Description))

	proto := strings.ToLower(string(r.Protocol))
	var parts []string
	parts = append(parts, fmt.Sprintf("iifname \"%s\"", r.ReflectionInterface))
	if r.Destination.Value != "" {
		parts = append(parts, buildFamilyMatch(detectAddressFamily(r.Destination.Value), "daddr", r.Destination.Value, false))
	}
	if proto != "any" && proto != "" && r.Destination.Port != "" {
		parts = append(parts, fmt.Sprintf("%s dport %s", proto, r.Destination.Port))
	}
	target := formatNATTarget(r.RedirectTarget, r.RedirectPort)
	parts = append(parts, fmt.Sprintf("dnat to %s", target))

	b.WriteString("\t\t")
	b.WriteString(strings.Join(parts, " "))
	b.WriteString("\n")
}

func compileNATReflectionSNAT(b *strings.Builder, r firewall.NATRule) {
	if r.ReflectionNetwork == "" || r.RedirectTarget == "" {
		return
	}
	b.WriteString(fmt.Sprintf("\t\t# [%s] NAT Reflection (SNAT): %s\n", r.ID, r.Description))

	family := detectAddressFamily(r.ReflectionNetwork)
	proto := strings.ToLower(string(r.Protocol))
	var parts []string
	parts = append(parts, buildFamilyMatch(family, "saddr", r.ReflectionNetwork, false))
	parts = append(parts, buildFamilyMatch(detectAddressFamily(r.RedirectTarget), "daddr", r.RedirectTarget, false))
	if proto != "any" && proto != "" && r.RedirectPort != "" {
		parts = append(parts, fmt.Sprintf("%s dport %s", proto, r.RedirectPort))
	}
	parts = append(parts, "masquerade")

	b.WriteString("\t\t")
	b.WriteString(strings.Join(parts, " "))
	b.WriteString("\n")
}

func buildAliasFamilies(aliases []firewall.Alias) map[string]string {
	families := make(map[string]string, len(aliases))
	for _, alias := range aliases {
		if alias.Type == firewall.AliasPort {
			families[alias.Name] = aliasFamilyPort
			continue
		}

		v4Entries, v6Entries := splitEntriesByFamily(alias.Entries)
		switch {
		case len(v4Entries) > 0 && len(v6Entries) > 0:
			families[alias.Name] = aliasFamilyMixed
		case len(v6Entries) > 0:
			families[alias.Name] = aliasFamilyV6
		default:
			families[alias.Name] = aliasFamilyV4
		}
	}
	return families
}

func splitEntriesByFamily(entries []string) (v4Entries []string, v6Entries []string) {
	for _, entry := range entries {
		if detectAddressFamily(entry) == aliasFamilyV6 {
			v6Entries = append(v6Entries, entry)
		} else {
			v4Entries = append(v4Entries, entry)
		}
	}
	return v4Entries, v6Entries
}

func entriesUseInterval(entries []string) bool {
	for _, entry := range entries {
		if entryUsesInterval(entry) {
			return true
		}
	}
	return false
}

func entryUsesInterval(entry string) bool {
	entry = strings.TrimSpace(entry)
	if entry == "" {
		return false
	}

	if _, _, err := net.ParseCIDR(entry); err == nil {
		return true
	}

	parts := strings.Split(entry, "-")
	if len(parts) != 2 {
		return false
	}

	left := strings.TrimSpace(parts[0])
	right := strings.TrimSpace(parts[1])
	return net.ParseIP(left) != nil && net.ParseIP(right) != nil
}

func writeSet(b *strings.Builder, name, nftType string, interval bool, entries []string) {
	if len(entries) == 0 {
		return
	}

	b.WriteString(fmt.Sprintf("\tset %s {\n", name))
	b.WriteString(fmt.Sprintf("\t\ttype %s\n", nftType))
	if interval {
		b.WriteString("\t\tflags interval\n")
		b.WriteString("\t\tauto-merge\n")
	}
	b.WriteString(fmt.Sprintf("\t\telements = { %s }\n", strings.Join(entries, ", ")))
	b.WriteString("\t}\n\n")
}

func compileInterfaceMatch(rule firewall.Rule, chain string) string {
	ifaces := rule.Interfaces
	if len(ifaces) == 0 && rule.Interface != "" {
		ifaces = []string{rule.Interface}
	}
	if len(ifaces) == 0 {
		return ""
	}

	quoted := make([]string, 0, len(ifaces))
	for _, iface := range ifaces {
		quoted = append(quoted, fmt.Sprintf("\"%s\"", iface))
	}

	field := "iifname"
	if chain == "output" {
		field = "oifname"
	}
	if len(quoted) == 1 {
		return fmt.Sprintf("%s %s", field, quoted[0])
	}
	return fmt.Sprintf("%s { %s }", field, strings.Join(quoted, ", "))
}

func detectAddressFamily(value string) string {
	value = strings.TrimSpace(strings.Trim(value, "[]"))
	if value == "" {
		return aliasFamilyV4
	}

	if strings.Contains(value, "/") {
		if _, ipNet, err := net.ParseCIDR(value); err == nil && ipNet != nil {
			if ipNet.IP.To4() == nil {
				return aliasFamilyV6
			}
			return aliasFamilyV4
		}
	}

	ip := net.ParseIP(value)
	if ip != nil && ip.To4() == nil {
		return aliasFamilyV6
	}
	return aliasFamilyV4
}

func buildFamilyMatch(family, field, value string, negate bool) string {
	operator := ""
	if negate {
		operator = "!= "
	}

	prefix := "ip"
	if family == aliasFamilyV6 {
		prefix = "ip6"
	}
	return fmt.Sprintf("%s %s %s%s", prefix, field, operator, value)
}

func compileAliasAddress(name, field string, negate bool, family string) string {
	switch family {
	case aliasFamilyV6:
		return buildFamilyMatch(aliasFamilyV6, field, "@"+name, negate)
	case aliasFamilyMixed:
		if negate {
			return fmt.Sprintf("not (ip %s @%s_v4 or ip6 %s @%s_v6)", field, name, field, name)
		}
		return fmt.Sprintf("(ip %s @%s_v4 or ip6 %s @%s_v6)", field, name, field, name)
	default:
		return buildFamilyMatch(aliasFamilyV4, field, "@"+name, negate)
	}
}

// gatewayMark generates a deterministic nftables mark from a gateway name.
func gatewayMark(gateway string) uint32 {
	var h uint32 = 2166136261 // FNV offset basis
	for _, c := range gateway {
		h ^= uint32(c)
		h *= 16777619
	}
	return (h % 1000) + 100 // range 100-1099 to avoid conflicts
}

func formatNATTarget(addr, port string) string {
	target := addr
	if detectAddressFamily(addr) == aliasFamilyV6 && port != "" {
		target = "[" + addr + "]"
	}
	if port != "" {
		target += ":" + port
	}
	return target
}

// ─── Conntrack Parser ──────────────────────────────────────────

// ParseConntrack parses conntrack -L output into State structs.
func ParseConntrack(output string) []firewall.State {
	var states []firewall.State
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		state := firewall.State{}
		fields := strings.Fields(line)
		if len(fields) < 4 {
			continue
		}

		// Format: protocol proto_num state src=x dst=x sport=x dport=x ...
		state.Protocol = fields[0]

		sportSet, dportSet := false, false
		for _, f := range fields {
			if strings.HasPrefix(f, "src=") {
				val := strings.TrimPrefix(f, "src=")
				if state.SourceIP == "" {
					state.SourceIP = val
				}
			} else if strings.HasPrefix(f, "dst=") {
				val := strings.TrimPrefix(f, "dst=")
				if state.DestIP == "" {
					state.DestIP = val
				}
			} else if strings.HasPrefix(f, "sport=") && !sportSet {
				fmt.Sscanf(strings.TrimPrefix(f, "sport="), "%d", &state.SourcePort)
				sportSet = true
			} else if strings.HasPrefix(f, "dport=") && !dportSet {
				fmt.Sscanf(strings.TrimPrefix(f, "dport="), "%d", &state.DestPort)
				dportSet = true
			} else if strings.HasPrefix(f, "packets=") {
				fmt.Sscanf(strings.TrimPrefix(f, "packets="), "%d", &state.Packets)
			} else if strings.HasPrefix(f, "bytes=") {
				fmt.Sscanf(strings.TrimPrefix(f, "bytes="), "%d", &state.Bytes)
			}
			// Check for state keywords
			switch f {
			case "ESTABLISHED", "SYN_SENT", "SYN_RECV", "FIN_WAIT", "CLOSE_WAIT",
				"LAST_ACK", "TIME_WAIT", "CLOSE", "LISTEN":
				state.State = f
			}
		}

		if state.SourceIP != "" {
			states = append(states, state)
		}
	}
	return states
}
