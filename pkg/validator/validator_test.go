package validator

import "testing"

func TestIPv4(t *testing.T) {
	tests := []struct {
		input string
		valid bool
	}{
		{"192.168.1.1", true},
		{"10.0.0.1", true},
		{"255.255.255.255", true},
		{"0.0.0.0", true},
		{"999.1.1.1", false},
		{"abc", false},
		{"", false},
		{"::1", false}, // IPv6
	}
	for _, tt := range tests {
		err := IPv4(tt.input)
		if tt.valid && err != nil {
			t.Errorf("IPv4(%q) should be valid, got error: %v", tt.input, err)
		}
		if !tt.valid && err == nil {
			t.Errorf("IPv4(%q) should be invalid", tt.input)
		}
	}
}

func TestCIDR(t *testing.T) {
	tests := []struct {
		input string
		valid bool
	}{
		{"192.168.1.0/24", true},
		{"10.0.0.0/8", true},
		{"0.0.0.0/0", true},
		{"fe80::/10", true},
		{"192.168.1.1", false},
		{"abc/24", false},
	}
	for _, tt := range tests {
		err := CIDR(tt.input)
		if tt.valid && err != nil {
			t.Errorf("CIDR(%q) should be valid, got error: %v", tt.input, err)
		}
		if !tt.valid && err == nil {
			t.Errorf("CIDR(%q) should be invalid", tt.input)
		}
	}
}

func TestPort(t *testing.T) {
	if err := Port(80); err != nil {
		t.Errorf("Port(80) should be valid")
	}
	if err := Port(443); err != nil {
		t.Errorf("Port(443) should be valid")
	}
	if err := Port(0); err == nil {
		t.Errorf("Port(0) should be invalid")
	}
	if err := Port(65536); err == nil {
		t.Errorf("Port(65536) should be invalid")
	}
	if err := Port(-1); err == nil {
		t.Errorf("Port(-1) should be invalid")
	}
}

func TestPortRange(t *testing.T) {
	if err := PortRange("80-443"); err != nil {
		t.Errorf("PortRange(80-443) should be valid: %v", err)
	}
	if err := PortRange("443-80"); err == nil {
		t.Errorf("PortRange(443-80) should be invalid (start > end)")
	}
	if err := PortRange("80"); err == nil {
		t.Errorf("PortRange(80) should be invalid (no range)")
	}
}

func TestMAC(t *testing.T) {
	if err := MAC("00:1A:2B:3C:4D:5E"); err != nil {
		t.Errorf("valid MAC should pass: %v", err)
	}
	if err := MAC("00:1a:2b:3c:4d:5e"); err != nil {
		t.Errorf("lowercase MAC should pass: %v", err)
	}
	if err := MAC("invalid"); err == nil {
		t.Errorf("invalid MAC should fail")
	}
}

func TestHostname(t *testing.T) {
	if err := Hostname("firewall01"); err != nil {
		t.Errorf("simple hostname should pass: %v", err)
	}
	if err := Hostname("fw.example.com"); err != nil {
		t.Errorf("FQDN should pass: %v", err)
	}
	if err := Hostname("-invalid"); err == nil {
		t.Errorf("hostname starting with dash should fail")
	}
}

func TestProtocol(t *testing.T) {
	for _, p := range []string{"tcp", "udp", "icmp", "any", "esp", "gre"} {
		if err := Protocol(p); err != nil {
			t.Errorf("Protocol(%q) should be valid: %v", p, err)
		}
	}
	if err := Protocol("invalid"); err == nil {
		t.Errorf("invalid protocol should fail")
	}
}

func TestFirewallAction(t *testing.T) {
	for _, a := range []string{"pass", "block", "reject", "drop"} {
		if err := FirewallAction(a); err != nil {
			t.Errorf("FirewallAction(%q) should be valid: %v", a, err)
		}
	}
	if err := FirewallAction("allow"); err == nil {
		t.Errorf("'allow' should be invalid firewall action")
	}
}

func TestInterfaceName(t *testing.T) {
	valid := []string{"eth0", "ens3", "br0", "bond0", "vlan100", "wg0"}
	for _, n := range valid {
		if err := InterfaceName(n); err != nil {
			t.Errorf("InterfaceName(%q) should be valid: %v", n, err)
		}
	}
	if err := InterfaceName("0eth"); err == nil {
		t.Errorf("interface name starting with digit should fail")
	}
}
