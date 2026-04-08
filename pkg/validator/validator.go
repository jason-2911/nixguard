// Package validator provides input validation utilities for NixGuard.
// Validates IPs, CIDRs, ports, MAC addresses, hostnames, and firewall-specific inputs.
package validator

import (
	"fmt"
	"net"
	"regexp"
	"strconv"
	"strings"
)

var (
	hostnameRe = regexp.MustCompile(`^[a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?(\.[a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?)*$`)
	macRe      = regexp.MustCompile(`^([0-9A-Fa-f]{2}:){5}[0-9A-Fa-f]{2}$`)
	ifnameRe   = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9._\-]{0,14}$`)
)

// IPv4 validates an IPv4 address.
func IPv4(s string) error {
	ip := net.ParseIP(s)
	if ip == nil || ip.To4() == nil {
		return fmt.Errorf("invalid IPv4 address: %s", s)
	}
	return nil
}

// IPv6 validates an IPv6 address.
func IPv6(s string) error {
	ip := net.ParseIP(s)
	if ip == nil || ip.To4() != nil {
		return fmt.Errorf("invalid IPv6 address: %s", s)
	}
	return nil
}

// IP validates an IPv4 or IPv6 address.
func IP(s string) error {
	if net.ParseIP(s) == nil {
		return fmt.Errorf("invalid IP address: %s", s)
	}
	return nil
}

// CIDR validates a CIDR notation (e.g., 192.168.1.0/24).
func CIDR(s string) error {
	_, _, err := net.ParseCIDR(s)
	if err != nil {
		return fmt.Errorf("invalid CIDR: %s", s)
	}
	return nil
}

// Port validates a port number (1-65535).
func Port(p int) error {
	if p < 1 || p > 65535 {
		return fmt.Errorf("invalid port: %d (must be 1-65535)", p)
	}
	return nil
}

// PortRange validates a port range string (e.g., "80-443").
func PortRange(s string) error {
	parts := strings.SplitN(s, "-", 2)
	if len(parts) != 2 {
		return fmt.Errorf("invalid port range: %s (expected start-end)", s)
	}
	start, err := strconv.Atoi(parts[0])
	if err != nil {
		return fmt.Errorf("invalid port range start: %s", parts[0])
	}
	end, err := strconv.Atoi(parts[1])
	if err != nil {
		return fmt.Errorf("invalid port range end: %s", parts[1])
	}
	if err := Port(start); err != nil {
		return err
	}
	if err := Port(end); err != nil {
		return err
	}
	if start > end {
		return fmt.Errorf("port range start (%d) > end (%d)", start, end)
	}
	return nil
}

// MAC validates a MAC address (XX:XX:XX:XX:XX:XX).
func MAC(s string) error {
	if !macRe.MatchString(s) {
		return fmt.Errorf("invalid MAC address: %s", s)
	}
	return nil
}

// Hostname validates a hostname per RFC 1123.
func Hostname(s string) error {
	if len(s) > 253 || !hostnameRe.MatchString(s) {
		return fmt.Errorf("invalid hostname: %s", s)
	}
	return nil
}

// InterfaceName validates a Linux network interface name.
func InterfaceName(s string) error {
	if !ifnameRe.MatchString(s) {
		return fmt.Errorf("invalid interface name: %s", s)
	}
	return nil
}

// Protocol validates a firewall protocol string.
func Protocol(s string) error {
	valid := map[string]bool{
		"tcp": true, "udp": true, "icmp": true, "icmpv6": true,
		"any": true, "esp": true, "ah": true, "gre": true,
		"sctp": true, "ipv6-icmp": true,
	}
	if !valid[strings.ToLower(s)] {
		return fmt.Errorf("invalid protocol: %s", s)
	}
	return nil
}

// FirewallAction validates a firewall action.
func FirewallAction(s string) error {
	valid := map[string]bool{
		"pass": true, "block": true, "reject": true, "drop": true,
	}
	if !valid[strings.ToLower(s)] {
		return fmt.Errorf("invalid action: %s (must be pass/block/reject/drop)", s)
	}
	return nil
}
