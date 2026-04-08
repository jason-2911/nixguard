// Package main is the entry point for the NixGuard privileged agent.
// The agent runs as root and exposes a gRPC service over Unix domain socket.
// It executes system-level operations: nftables, iproute2, systemd, tcpdump.
//
// Security model:
// - Listens ONLY on Unix domain socket (no network exposure)
// - Socket permissions restricted to nixguard group
// - Command whitelist — only pre-approved operations
// - Full audit logging of all executed commands
package main

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/nixguard/nixguard/internal/config"
	"github.com/nixguard/nixguard/pkg/executor"
	"github.com/nixguard/nixguard/pkg/logger"
	"github.com/nixguard/nixguard/pkg/version"
	"google.golang.org/grpc"
)

const defaultSocketPath = "/var/run/nixguard/agent.sock"

func main() {
	cfg, err := config.LoadAgent("configs/defaults/agent.yaml")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		os.Exit(1)
	}

	log := logger.New(cfg.Log.Level, cfg.Log.Format)
	log.Info("starting nixguard-agent",
		slog.String("version", version.Version),
		slog.String("socket", cfg.Agent.SocketPath),
	)

	// ── Verify Root ────────────────────────────────────────────
	if os.Getuid() != 0 {
		log.Error("nixguard-agent must run as root")
		os.Exit(1)
	}

	// ── Command Executor (whitelisted) ─────────────────────────
	exec := executor.NewSafe(log, cfg.Agent.AllowedCommands)

	// ── gRPC Server ────────────────────────────────────────────
	socketPath := cfg.Agent.SocketPath
	if socketPath == "" {
		socketPath = defaultSocketPath
	}

	// Clean up stale socket
	os.Remove(socketPath)

	lis, err := net.Listen("unix", socketPath)
	if err != nil {
		log.Error("failed to listen", slog.String("error", err.Error()))
		os.Exit(1)
	}

	// Restrict socket permissions: owner + group only
	if err := os.Chmod(socketPath, 0660); err != nil {
		log.Error("failed to set socket permissions", slog.String("error", err.Error()))
		os.Exit(1)
	}

	srv := grpc.NewServer(
		grpc.UnaryInterceptor(auditInterceptor(log)),
	)

	registerServices(srv, exec, log)

	// ── Graceful Shutdown ──────────────────────────────────────
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		log.Info("agent listening", slog.String("socket", socketPath))
		if err := srv.Serve(lis); err != nil {
			log.Error("agent server error", slog.String("error", err.Error()))
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	log.Info("shutting down agent...")
	srv.GracefulStop()
	os.Remove(socketPath)
	log.Info("agent stopped")
}

func registerServices(srv *grpc.Server, exec *executor.Safe, log *slog.Logger) {
	// Register all agent service implementations:
	// - FirewallAgent (nftables/iptables operations)
	// - NetworkAgent (iproute2, interface config)
	// - ServiceAgent (systemd service management)
	// - CaptureAgent (tcpdump, packet capture)
	// - SystemAgent (sysctl, kernel params)
	_ = exec // Will be used by service implementations
}

func auditInterceptor(log *slog.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		log.Info("agent rpc",
			slog.String("method", info.FullMethod),
		)
		resp, err := handler(ctx, req)
		if err != nil {
			log.Error("agent rpc failed",
				slog.String("method", info.FullMethod),
				slog.String("error", err.Error()),
			)
		}
		return resp, err
	}
}
