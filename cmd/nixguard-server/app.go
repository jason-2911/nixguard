package main

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/nixguard/nixguard/internal/adapter/http/handler/firewall"
	"github.com/nixguard/nixguard/internal/adapter/http/handler/network"
	"github.com/nixguard/nixguard/internal/adapter/http/middleware"
	fwApp "github.com/nixguard/nixguard/internal/app/firewall"
	netApp "github.com/nixguard/nixguard/internal/app/network"
	"github.com/nixguard/nixguard/internal/config"
	"github.com/nixguard/nixguard/internal/event"
	"github.com/nixguard/nixguard/internal/infra/database"
	"github.com/nixguard/nixguard/internal/infra/database/sqlite"
	"github.com/nixguard/nixguard/internal/infra/iproute2"
	"github.com/nixguard/nixguard/internal/infra/nftables"
	"github.com/nixguard/nixguard/pkg/executor"
	"github.com/nixguard/nixguard/pkg/version"
)

// Application holds all initialized modules and provides the HTTP router.
type Application struct {
	cfg    *config.Config
	bus    *event.Bus
	log    *slog.Logger
	db     *database.DB
	router http.Handler
}

func initApp(cfg *config.Config, bus *event.Bus, log *slog.Logger) (*Application, error) {
	app := &Application{
		cfg: cfg,
		bus: bus,
		log: log,
	}

	// ── Database ───────────────────────────────────────────────
	db, err := database.Open(cfg.Database.Driver, cfg.Database.DSN, log)
	if err != nil {
		return nil, err
	}
	app.db = db

	// Run migrations
	if err := db.Migrate("internal/infra/database/migrations"); err != nil {
		return nil, err
	}
	log.Info("database ready")

	// ── Command Executor (for local dev, runs commands directly) ─
	exec := executor.NewSafe(log, cfg.Agent.AllowedCommands)

	// ── Repositories ───────────────────────────────────────────
	ruleRepo := sqlite.NewRuleRepo(db)
	aliasRepo := sqlite.NewAliasRepo(db)
	natRepo := sqlite.NewNATRepo(db)
	ifaceRepo := sqlite.NewInterfaceRepo(db)
	routeRepo := sqlite.NewRouteRepo(db)
	gwRepo := sqlite.NewGatewayRepo(db)
	gwGroupRepo := sqlite.NewGatewayGroupRepo(db)

	// ── Infrastructure Adapters ────────────────────────────────
	nftEngine := nftables.NewAdapter(exec, log)
	netEngine := iproute2.NewAdapter(exec, log)

	// ── Application Services ───────────────────────────────────
	fwService := fwApp.NewService(ruleRepo, aliasRepo, natRepo, nftEngine, nil, bus, log)
	netService := netApp.NewService(ifaceRepo, routeRepo, gwRepo, gwGroupRepo, netEngine, bus, log)

	// Start background URL alias refresher
	fwService.StartURLRefresher(context.Background())

	// ── HTTP Handlers ──────────────────────────────────────────
	fwHandler := firewall.NewHandler(fwService)
	netHandler := network.NewHandler(netService)

	// ── Build Router ───────────────────────────────────────────
	mux := http.NewServeMux()

	// System routes
	mux.HandleFunc("GET /api/v1/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"healthy","timestamp":"` + r.Header.Get("Date") + `"}`))
	})
	mux.HandleFunc("GET /api/v1/version", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"version":"` + version.Version + `","build_time":"` + version.BuildTime + `"}`))
	})

	// Register module routes
	fwHandler.RegisterRoutes(mux)
	netHandler.RegisterRoutes(mux)

	// Middleware chain
	mw := middleware.NewChain(
		middleware.Recovery(log),
		middleware.RequestID(),
		middleware.Logger(log),
		middleware.CORS(cfg.Server.CORSOrigins),
	)

	app.router = mw(mux)

	log.Info("application initialized",
		slog.Bool("firewall", true),
		slog.Bool("network", true),
	)

	return app, nil
}

// Router returns the configured HTTP handler.
func (a *Application) Router() http.Handler {
	return a.router
}

// Shutdown gracefully stops all modules.
func (a *Application) Shutdown(ctx context.Context) error {
	a.log.Info("shutting down modules...")
	a.bus.Close()
	if a.db != nil {
		a.db.Close()
	}
	return nil
}
