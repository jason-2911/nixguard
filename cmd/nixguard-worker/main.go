// Package main is the entry point for the NixGuard background worker.
// Handles scheduled tasks: log rotation, GeoIP updates, certificate renewal,
// backup scheduling, rule updates, health checks, and metrics collection.
package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/nixguard/nixguard/internal/config"
	"github.com/nixguard/nixguard/pkg/logger"
	"github.com/nixguard/nixguard/pkg/version"
)

func main() {
	cfg, err := config.Load("configs/defaults/server.yaml")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		os.Exit(1)
	}

	log := logger.New(cfg.Log.Level, cfg.Log.Format)
	log.Info("starting nixguard-worker",
		slog.String("version", version.Version),
	)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// ── Register Jobs ──────────────────────────────────────────
	scheduler := NewScheduler(log)

	// Core jobs
	scheduler.Register("geoip_update", "0 3 * * 0", NewGeoIPUpdateJob(cfg, log))       // Weekly Sunday 3am
	scheduler.Register("cert_renewal", "0 2 * * *", NewCertRenewalJob(cfg, log))        // Daily 2am
	scheduler.Register("backup_auto", "0 1 * * *", NewAutoBackupJob(cfg, log))          // Daily 1am
	scheduler.Register("log_rotation", "0 0 * * *", NewLogRotationJob(cfg, log))        // Daily midnight
	scheduler.Register("ids_rule_update", "0 4 * * *", NewIDSRuleUpdateJob(cfg, log))   // Daily 4am
	scheduler.Register("health_check", "*/1 * * * *", NewHealthCheckJob(cfg, log))      // Every minute
	scheduler.Register("metrics_collect", "*/30 * * * * *", NewMetricsJob(cfg, log))    // Every 30s
	scheduler.Register("dns_blocklist", "0 5 * * 0", NewDNSBlocklistJob(cfg, log))      // Weekly Sunday 5am
	scheduler.Register("session_cleanup", "*/5 * * * *", NewSessionCleanupJob(cfg, log)) // Every 5min

	scheduler.Start(ctx)

	<-ctx.Done()
	log.Info("shutting down worker...")
	scheduler.Stop()
	log.Info("worker stopped")
}

// Scheduler, Job interface, and job constructors are stubs for now.
// They will be implemented as the module system matures.

type Job interface {
	Run(ctx context.Context) error
}

type Scheduler struct {
	log  *slog.Logger
	jobs map[string]Job
}

func NewScheduler(log *slog.Logger) *Scheduler {
	return &Scheduler{log: log, jobs: make(map[string]Job)}
}

func (s *Scheduler) Register(name, cron string, job Job) {
	s.log.Info("registered job", slog.String("name", name), slog.String("schedule", cron))
	s.jobs[name] = job
}

func (s *Scheduler) Start(ctx context.Context) {
	s.log.Info("scheduler started", slog.Int("job_count", len(s.jobs)))
}

func (s *Scheduler) Stop() {
	s.log.Info("scheduler stopped")
}

// Stub job constructors
type stubJob struct{}

func (j *stubJob) Run(ctx context.Context) error { return nil }

func NewGeoIPUpdateJob(cfg *config.Config, log *slog.Logger) Job    { return &stubJob{} }
func NewCertRenewalJob(cfg *config.Config, log *slog.Logger) Job    { return &stubJob{} }
func NewAutoBackupJob(cfg *config.Config, log *slog.Logger) Job     { return &stubJob{} }
func NewLogRotationJob(cfg *config.Config, log *slog.Logger) Job    { return &stubJob{} }
func NewIDSRuleUpdateJob(cfg *config.Config, log *slog.Logger) Job  { return &stubJob{} }
func NewHealthCheckJob(cfg *config.Config, log *slog.Logger) Job    { return &stubJob{} }
func NewMetricsJob(cfg *config.Config, log *slog.Logger) Job        { return &stubJob{} }
func NewDNSBlocklistJob(cfg *config.Config, log *slog.Logger) Job   { return &stubJob{} }
func NewSessionCleanupJob(cfg *config.Config, log *slog.Logger) Job { return &stubJob{} }
