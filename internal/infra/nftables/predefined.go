package nftables

import (
	"fmt"
	"strings"

	"github.com/nixguard/nixguard/pkg/netutil"
)

func compilePredefinedSets(b *strings.Builder, filters *PredefinedFilters) {
	if filters.BogonFilter {
		writeSet(b, "bogons_v4", "ipv4_addr", true, netutil.BogonRangesV4)
		writeSet(b, "bogons_v6", "ipv6_addr", true, netutil.BogonRangesV6)
	}
	if filters.RFC1918Filter {
		writeSet(b, "rfc1918", "ipv4_addr", true, netutil.RFC1918Ranges)
	}
}

func compilePredefinedFilterRules(b *strings.Builder, filters *PredefinedFilters) {
	for _, iface := range filters.WANInterfaces {
		if filters.BogonFilter {
			b.WriteString(fmt.Sprintf("\t\t# Bogon filter on %s\n", iface))
			b.WriteString(fmt.Sprintf("\t\tiifname \"%s\" ip saddr @bogons_v4 counter drop\n", iface))
			b.WriteString(fmt.Sprintf("\t\tiifname \"%s\" ip6 saddr @bogons_v6 counter drop\n", iface))
		}
		if filters.RFC1918Filter {
			b.WriteString(fmt.Sprintf("\t\t# RFC1918 filter on %s\n", iface))
			b.WriteString(fmt.Sprintf("\t\tiifname \"%s\" ip saddr @rfc1918 counter drop\n", iface))
		}
	}
}
