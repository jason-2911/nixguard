package nftables

import (
	"strings"
	"testing"

	"github.com/nixguard/nixguard/internal/domain/firewall"
)

func TestCompileRuleset_Empty(t *testing.T) {
	result := CompileRuleset(nil, nil, nil)
	if !strings.Contains(result, "table inet nixguard") {
		t.Error("expected nixguard table declaration")
	}
	if !strings.Contains(result, "chain input") {
		t.Error("expected input chain")
	}
	if !strings.Contains(result, "chain forward") {
		t.Error("expected forward chain")
	}
	if !strings.Contains(result, "chain output") {
		t.Error("expected output chain")
	}
	if !strings.Contains(result, "ct state established,related accept") {
		t.Error("expected conntrack established rule")
	}
}

func TestCompileRuleset_SinglePassRule(t *testing.T) {
	rules := []firewall.Rule{
		{
			ID:        "test001",
			Interface: "eth0",
			Direction: firewall.DirectionIn,
			Action:    firewall.ActionPass,
			Protocol:  firewall.ProtoTCP,
			Source:    firewall.Address{Type: firewall.AddrAny},
			Destination: firewall.Address{
				Type:  firewall.AddrSingle,
				Value: "192.168.1.100",
				Port:  "443",
			},
			Enabled:     true,
			Description: "Allow HTTPS to server",
		},
	}

	result := CompileRuleset(rules, nil, nil)

	// Should contain the rule
	if !strings.Contains(result, "test001") {
		t.Error("expected rule ID in output")
	}
	if !strings.Contains(result, "iifname \"eth0\"") {
		t.Error("expected interface match")
	}
	if !strings.Contains(result, "ip daddr 192.168.1.100") {
		t.Error("expected destination IP match")
	}
	if !strings.Contains(result, "tcp dport 443") {
		t.Error("expected destination port match")
	}
	if !strings.Contains(result, "accept") {
		t.Error("expected accept action")
	}
	if !strings.Contains(result, "counter name cnt_test001") {
		t.Error("expected counter reference")
	}
}

func TestCompileRuleset_BlockRule(t *testing.T) {
	rules := []firewall.Rule{
		{
			ID:        "block001",
			Interface: "eth1",
			Direction: firewall.DirectionIn,
			Action:    firewall.ActionBlock,
			Protocol:  firewall.ProtoAny,
			Source: firewall.Address{
				Type:  firewall.AddrNetwork,
				Value: "10.0.0.0/8",
			},
			Destination: firewall.Address{Type: firewall.AddrAny},
			Enabled:     true,
			Log:         true,
			Description: "Block private range",
		},
	}

	result := CompileRuleset(rules, nil, nil)

	if !strings.Contains(result, "ip saddr 10.0.0.0/8") {
		t.Error("expected source network match")
	}
	if !strings.Contains(result, "drop") {
		t.Error("expected drop action")
	}
	if !strings.Contains(result, "log prefix") {
		t.Error("expected log statement")
	}
}

func TestCompileRuleset_DisabledRuleSkipped(t *testing.T) {
	rules := []firewall.Rule{
		{
			ID:          "disabled01",
			Direction:   firewall.DirectionIn,
			Action:      firewall.ActionPass,
			Protocol:    firewall.ProtoTCP,
			Source:      firewall.Address{Type: firewall.AddrAny},
			Destination: firewall.Address{Type: firewall.AddrAny},
			Enabled:     false,
		},
	}

	result := CompileRuleset(rules, nil, nil)

	if strings.Contains(result, "disabled01") {
		t.Error("disabled rule should not appear in compiled output")
	}
}

func TestCompileRuleset_Alias(t *testing.T) {
	aliases := []firewall.Alias{
		{
			Name:    "webservers",
			Type:    firewall.AliasHost,
			Entries: []string{"192.168.1.10", "192.168.1.11", "192.168.1.12"},
			Enabled: true,
		},
	}

	result := CompileRuleset(nil, aliases, nil)

	if !strings.Contains(result, "set webservers") {
		t.Error("expected set declaration for alias")
	}
	if !strings.Contains(result, "type ipv4_addr") {
		t.Error("expected ipv4_addr type for host alias")
	}
	if !strings.Contains(result, "192.168.1.10, 192.168.1.11, 192.168.1.12") {
		t.Error("expected alias entries in set elements")
	}
}

func TestCompileRuleset_NetworkAlias(t *testing.T) {
	aliases := []firewall.Alias{
		{
			Name:    "internal_nets",
			Type:    firewall.AliasNetwork,
			Entries: []string{"10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16"},
			Enabled: true,
		},
	}

	result := CompileRuleset(nil, aliases, nil)

	if !strings.Contains(result, "set internal_nets") {
		t.Error("expected set declaration")
	}
	if !strings.Contains(result, "flags interval") {
		t.Error("expected interval flag for network alias")
	}
}

func TestCompileRuleset_AliasReference(t *testing.T) {
	rules := []firewall.Rule{
		{
			ID:        "alias01",
			Direction: firewall.DirectionIn,
			Action:    firewall.ActionPass,
			Protocol:  firewall.ProtoTCP,
			Source: firewall.Address{
				Type:  firewall.AddrAlias,
				Value: "trusted_hosts",
			},
			Destination: firewall.Address{
				Type: firewall.AddrAny,
				Port: "22",
			},
			Enabled: true,
		},
	}

	result := CompileRuleset(rules, nil, nil)

	if !strings.Contains(result, "@trusted_hosts") {
		t.Error("expected alias reference with @ notation")
	}
}

func TestCompileRuleset_IPv6Rule(t *testing.T) {
	rules := []firewall.Rule{
		{
			ID:        "v6rule01",
			Interface: "wan0",
			Direction: firewall.DirectionIn,
			Action:    firewall.ActionPass,
			Protocol:  firewall.ProtoTCP,
			Source:    firewall.Address{Type: firewall.AddrAny},
			Destination: firewall.Address{
				Type:  firewall.AddrSingle,
				Value: "2001:db8::10",
				Port:  "443",
			},
			Enabled: true,
		},
	}

	result := CompileRuleset(rules, nil, nil)

	if !strings.Contains(result, "ip6 daddr 2001:db8::10") {
		t.Error("expected IPv6 destination match")
	}
	if !strings.Contains(result, "tcp dport 443") {
		t.Error("expected destination port match")
	}
}

func TestCompileRuleset_MixedAlias(t *testing.T) {
	aliases := []firewall.Alias{
		{
			Name:    "trusted_hosts",
			Type:    firewall.AliasHost,
			Entries: []string{"192.0.2.10", "2001:db8::20"},
			Enabled: true,
		},
	}
	rules := []firewall.Rule{
		{
			ID:        "mixed01",
			Direction: firewall.DirectionIn,
			Action:    firewall.ActionPass,
			Protocol:  firewall.ProtoTCP,
			Source: firewall.Address{
				Type:  firewall.AddrAlias,
				Value: "trusted_hosts",
			},
			Destination: firewall.Address{Type: firewall.AddrAny, Port: "22"},
			Enabled:     true,
		},
	}

	result := CompileRuleset(rules, aliases, nil)

	if !strings.Contains(result, "set trusted_hosts_v4") {
		t.Error("expected IPv4 alias set")
	}
	if !strings.Contains(result, "set trusted_hosts_v6") {
		t.Error("expected IPv6 alias set")
	}
	if !strings.Contains(result, "ip saddr @trusted_hosts_v4 or ip6 saddr @trusted_hosts_v6") {
		t.Error("expected mixed-family alias reference")
	}
}

func TestCompileRuleset_NestedAliasWithCIDRUsesInterval(t *testing.T) {
	aliases := []firewall.Alias{
		{
			Name:    "nested_all",
			Type:    firewall.AliasNested,
			Entries: []string{"192.168.1.10", "10.0.0.0/8", "2001:db8::10", "fd00::/8"},
			Enabled: true,
		},
	}

	result := CompileRuleset(nil, aliases, nil)

	if !strings.Contains(result, "set nested_all_v4") {
		t.Error("expected IPv4 set for mixed nested alias")
	}
	if !strings.Contains(result, "set nested_all_v6") {
		t.Error("expected IPv6 set for mixed nested alias")
	}
	if !strings.Contains(result, "set nested_all_v4 {\n\t\ttype ipv4_addr\n\t\tflags interval") {
		t.Error("expected interval flag for IPv4 set containing CIDR")
	}
	if !strings.Contains(result, "set nested_all_v6 {\n\t\ttype ipv6_addr\n\t\tflags interval") {
		t.Error("expected interval flag for IPv6 set containing CIDR")
	}
}

func TestCompileNATRuleset_PortForward(t *testing.T) {
	rules := []firewall.NATRule{
		{
			ID:        "nat001",
			Type:      firewall.NATPortForward,
			Interface: "eth0",
			Protocol:  firewall.ProtoTCP,
			Destination: firewall.Address{
				Type:  firewall.AddrSingle,
				Value: "203.0.113.1",
				Port:  "443",
			},
			RedirectTarget: "192.168.1.100",
			RedirectPort:   "443",
			Enabled:        true,
			Description:    "HTTPS to webserver",
		},
	}

	result := CompileNATRuleset(rules)

	if !strings.Contains(result, "table inet nixguard_nat") {
		t.Error("expected NAT table")
	}
	if !strings.Contains(result, "chain prerouting") {
		t.Error("expected prerouting chain")
	}
	if !strings.Contains(result, "dnat to 192.168.1.100:443") {
		t.Error("expected DNAT target")
	}
}

func TestCompileNATRuleset_Masquerade(t *testing.T) {
	rules := []firewall.NATRule{
		{
			ID:          "nat002",
			Type:        firewall.NATOutbound,
			Interface:   "eth0",
			Protocol:    firewall.ProtoAny,
			Source:      firewall.Address{Type: firewall.AddrAny},
			Enabled:     true,
			Description: "Outbound NAT on WAN",
		},
	}

	result := CompileNATRuleset(rules)

	if !strings.Contains(result, "chain postrouting") {
		t.Error("expected postrouting chain")
	}
	if !strings.Contains(result, "masquerade") {
		t.Error("expected masquerade action")
	}
	if !strings.Contains(result, "oifname \"eth0\"") {
		t.Error("expected outgoing interface match")
	}
}

func TestCompileNATRuleset_NAT66(t *testing.T) {
	rules := []firewall.NATRule{
		{
			ID:             "nat66",
			Type:           firewall.NATPortForward,
			Interface:      "wan0",
			Protocol:       firewall.ProtoTCP,
			Destination:    firewall.Address{Type: firewall.AddrSingle, Value: "2001:db8::1", Port: "443"},
			RedirectTarget: "fd00::10",
			RedirectPort:   "8443",
			Enabled:        true,
		},
	}

	result := CompileNATRuleset(rules)

	if !strings.Contains(result, "table inet nixguard_nat") {
		t.Error("expected inet NAT table for dual-stack NAT")
	}
	if !strings.Contains(result, "ip6 daddr 2001:db8::1") {
		t.Error("expected IPv6 destination match")
	}
	if !strings.Contains(result, "dnat to [fd00::10]:8443") {
		t.Error("expected IPv6 DNAT target with brackets")
	}
}

func TestParseConntrack(t *testing.T) {
	input := `tcp      6 117 TIME_WAIT src=192.168.1.100 dst=93.184.216.34 sport=54321 dport=80 packets=10 bytes=1500 src=93.184.216.34 dst=192.168.1.100 sport=80 dport=54321 packets=8 bytes=12000 [ASSURED] mark=0 use=1
udp      17 29 src=192.168.1.50 dst=8.8.8.8 sport=12345 dport=53 packets=1 bytes=64 src=8.8.8.8 dst=192.168.1.50 sport=53 dport=12345 packets=1 bytes=128 mark=0 use=1`

	states := ParseConntrack(input)

	if len(states) != 2 {
		t.Fatalf("expected 2 states, got %d", len(states))
	}

	// First entry — TCP
	if states[0].Protocol != "tcp" {
		t.Errorf("expected tcp, got %s", states[0].Protocol)
	}
	if states[0].SourceIP != "192.168.1.100" {
		t.Errorf("expected src 192.168.1.100, got %s", states[0].SourceIP)
	}
	if states[0].DestIP != "93.184.216.34" {
		t.Errorf("expected dst 93.184.216.34, got %s", states[0].DestIP)
	}
	if states[0].SourcePort != 54321 {
		t.Errorf("expected sport 54321, got %d", states[0].SourcePort)
	}
	if states[0].DestPort != 80 {
		t.Errorf("expected dport 80, got %d", states[0].DestPort)
	}
	if states[0].State != "TIME_WAIT" {
		t.Errorf("expected TIME_WAIT state, got %s", states[0].State)
	}

	// Second entry — UDP
	if states[1].Protocol != "udp" {
		t.Errorf("expected udp, got %s", states[1].Protocol)
	}
	if states[1].DestPort != 53 {
		t.Errorf("expected dport 53, got %d", states[1].DestPort)
	}
}

func TestCompileRuleset_FloatingRule(t *testing.T) {
	rules := []firewall.Rule{
		{
			ID:          "float01",
			Direction:   firewall.DirectionIn,
			Action:      firewall.ActionPass,
			Protocol:    firewall.ProtoTCP,
			IsFloating:  true,
			Source:      firewall.Address{Type: firewall.AddrAny},
			Destination: firewall.Address{Type: firewall.AddrAny, Port: "80"},
			Enabled:     true,
		},
	}

	result := CompileRuleset(rules, nil, nil)

	// Floating rules go to forward chain
	lines := strings.Split(result, "\n")
	inForward := false
	foundRule := false
	for _, line := range lines {
		if strings.Contains(line, "chain forward") {
			inForward = true
		}
		if strings.Contains(line, "chain output") {
			inForward = false
		}
		if inForward && strings.Contains(line, "float01") {
			foundRule = true
		}
	}
	if !foundRule {
		t.Error("floating rule should appear in forward chain")
	}
}

func TestCompileRuleset_RejectAction(t *testing.T) {
	rules := []firewall.Rule{
		{
			ID:          "rej01",
			Direction:   firewall.DirectionIn,
			Action:      firewall.ActionReject,
			Protocol:    firewall.ProtoTCP,
			Source:      firewall.Address{Type: firewall.AddrAny},
			Destination: firewall.Address{Type: firewall.AddrAny, Port: "23"},
			Enabled:     true,
		},
	}

	result := CompileRuleset(rules, nil, nil)

	if !strings.Contains(result, "reject") {
		t.Error("expected reject action")
	}
}

func TestCompileNATRuleset_Reflection(t *testing.T) {
	rules := []firewall.NATRule{{
		ID:          "ref001",
		Type:        firewall.NATPortForward,
		Interface:   "wan0",
		Protocol:    firewall.ProtoTCP,
		Destination: firewall.Address{Type: firewall.AddrSingle, Value: "203.0.113.1", Port: "443"},
		RedirectTarget:      "192.168.1.100",
		RedirectPort:        "8443",
		NATReflection:       true,
		ReflectionInterface: "lan0",
		ReflectionNetwork:   "192.168.1.0/24",
		Enabled:             true,
		Description:         "HTTPS with reflection",
	}}
	result := CompileNATRuleset(rules)
	if !strings.Contains(result, "NAT Reflection (DNAT)") {
		t.Error("missing reflection DNAT rule")
	}
	if !strings.Contains(result, "NAT Reflection (SNAT)") {
		t.Error("missing reflection SNAT rule")
	}
	if !strings.Contains(result, `iifname "lan0"`) {
		t.Error("reflection should use LAN interface")
	}
	if !strings.Contains(result, "masquerade") {
		t.Error("reflection SNAT should masquerade")
	}
}

func TestCompileNATRuleset_IPv6Masquerade(t *testing.T) {
	rules := []firewall.NATRule{
		{
			ID:        "nat_v6_01",
			Type:      firewall.NATOutbound,
			Interface: "wan0",
			Protocol:  firewall.ProtoAny,
			Source: firewall.Address{
				Type:  firewall.AddrNetwork,
				Value: "2001:db8:1::/48",
			},
			Enabled:     true,
			Description: "IPv6 outbound NAT",
		},
	}

	result := CompileNATRuleset(rules)

	if !strings.Contains(result, "ip6 saddr 2001:db8:1::/48") {
		t.Errorf("expected ip6 saddr match, got:\n%s", result)
	}
	if !strings.Contains(result, "masquerade") {
		t.Error("expected masquerade action")
	}
	if !strings.Contains(result, `oifname "wan0"`) {
		t.Error("expected outgoing interface match")
	}
}

func TestCompileRuleset_BogonFilter(t *testing.T) {
	filters := &PredefinedFilters{
		BogonFilter:   true,
		WANInterfaces: []string{"eth0"},
	}

	result := CompileRuleset(nil, nil, filters)

	if !strings.Contains(result, "set bogons_v4") {
		t.Error("expected bogons_v4 set")
	}
	if !strings.Contains(result, "set bogons_v6") {
		t.Error("expected bogons_v6 set")
	}
	if !strings.Contains(result, `iifname "eth0" ip saddr @bogons_v4 counter drop`) {
		t.Error("expected bogon v4 filter rule on eth0")
	}
	if !strings.Contains(result, `iifname "eth0" ip6 saddr @bogons_v6 counter drop`) {
		t.Error("expected bogon v6 filter rule on eth0")
	}
	// Should not have rfc1918 set when not enabled
	if strings.Contains(result, "set rfc1918") {
		t.Error("rfc1918 set should not be present when not enabled")
	}
}

func TestCompileRuleset_RFC1918Filter(t *testing.T) {
	filters := &PredefinedFilters{
		RFC1918Filter: true,
		WANInterfaces: []string{"wan0"},
	}

	result := CompileRuleset(nil, nil, filters)

	if !strings.Contains(result, "set rfc1918") {
		t.Error("expected rfc1918 set")
	}
	if !strings.Contains(result, `iifname "wan0" ip saddr @rfc1918 counter drop`) {
		t.Error("expected RFC1918 filter rule on wan0")
	}
	// Should not have bogon sets when not enabled
	if strings.Contains(result, "set bogons_v4") {
		t.Error("bogon sets should not be present when not enabled")
	}
}

func TestCompileRuleset_PolicyRouting(t *testing.T) {
	rules := []firewall.Rule{{
		ID:        "pbr001",
		Interface: "eth0",
		Direction: firewall.DirectionIn,
		Action:    firewall.ActionPass,
		Protocol:  firewall.ProtoTCP,
		Source:    firewall.Address{Type: firewall.AddrNetwork, Value: "10.0.0.0/8"},
		Destination: firewall.Address{Type: firewall.AddrAny},
		Gateway:     "wan-secondary",
		Enabled:     true,
		Description: "Route 10.x via secondary WAN",
	}}
	result := CompileRuleset(rules, nil, nil)
	if !strings.Contains(result, "meta mark set") {
		t.Error("expected meta mark set for policy routing")
	}
	if !strings.Contains(result, "accept") {
		t.Error("expected accept after mark")
	}
}

func TestCompileRuleset_BogonAndRFC1918Combined(t *testing.T) {
	filters := &PredefinedFilters{
		BogonFilter:   true,
		RFC1918Filter: true,
		WANInterfaces: []string{"eth0", "eth1"},
	}

	result := CompileRuleset(nil, nil, filters)

	// Both sets should be present
	if !strings.Contains(result, "set bogons_v4") {
		t.Error("expected bogons_v4 set")
	}
	if !strings.Contains(result, "set rfc1918") {
		t.Error("expected rfc1918 set")
	}
	// Filter rules for both interfaces
	if !strings.Contains(result, `iifname "eth0" ip saddr @bogons_v4 counter drop`) {
		t.Error("expected bogon filter on eth0")
	}
	if !strings.Contains(result, `iifname "eth1" ip saddr @rfc1918 counter drop`) {
		t.Error("expected RFC1918 filter on eth1")
	}
}
