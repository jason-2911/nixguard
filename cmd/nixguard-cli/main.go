// Package main is the entry point for the NixGuard CLI tool.
// Provides command-line management of NixGuard via the REST/gRPC API.
package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/nixguard/nixguard/pkg/version"
)

var rootCmd = &cobra.Command{
	Use:   "nixguard",
	Short: "NixGuard - Linux Network Security Framework",
	Long: `NixGuard is a complete firewall and network security framework
for Linux with 100% OPNsense feature parity.`,
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("NixGuard %s (built %s)\n", version.Version, version.BuildTime)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)

	// Module commands — each module registers its subcommands
	rootCmd.AddCommand(
		newFirewallCmd(),
		newNetworkCmd(),
		newVPNCmd(),
		newDNSCmd(),
		newDHCPCmd(),
		newIDSCmd(),
		newSystemCmd(),
		newMonitorCmd(),
		newBackupCmd(),
		newDiagCmd(),
		newServiceCmd(),
	)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// Stub command constructors — each will be in its own file
func newFirewallCmd() *cobra.Command {
	return &cobra.Command{Use: "firewall", Short: "Manage firewall rules and aliases", Aliases: []string{"fw"}}
}
func newNetworkCmd() *cobra.Command {
	return &cobra.Command{Use: "network", Short: "Manage network interfaces, routing, and gateways", Aliases: []string{"net"}}
}
func newVPNCmd() *cobra.Command {
	return &cobra.Command{Use: "vpn", Short: "Manage VPN tunnels (IPsec, OpenVPN, WireGuard)"}
}
func newDNSCmd() *cobra.Command {
	return &cobra.Command{Use: "dns", Short: "Manage DNS resolver and filtering"}
}
func newDHCPCmd() *cobra.Command {
	return &cobra.Command{Use: "dhcp", Short: "Manage DHCP server and leases"}
}
func newIDSCmd() *cobra.Command {
	return &cobra.Command{Use: "ids", Short: "Manage intrusion detection/prevention (Suricata)"}
}
func newSystemCmd() *cobra.Command {
	return &cobra.Command{Use: "system", Short: "System settings, updates, and diagnostics", Aliases: []string{"sys"}}
}
func newMonitorCmd() *cobra.Command {
	return &cobra.Command{Use: "monitor", Short: "View monitoring data, graphs, and alerts", Aliases: []string{"mon"}}
}
func newBackupCmd() *cobra.Command {
	return &cobra.Command{Use: "backup", Short: "Backup and restore configuration"}
}
func newDiagCmd() *cobra.Command {
	return &cobra.Command{Use: "diag", Short: "Run diagnostics (ping, traceroute, capture)"}
}
func newServiceCmd() *cobra.Command {
	return &cobra.Command{Use: "service", Short: "Control NixGuard services", Aliases: []string{"svc"}}
}
