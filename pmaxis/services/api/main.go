package main

import (
	"context"
	_ "embed"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	clickhouse "github.com/pmaxis/pmaxis/libs/clickhouse-client"
	"github.com/pmaxis/pmaxis/libs/config"
	postgres "github.com/pmaxis/pmaxis/libs/postgres-client"
	redisclient "github.com/pmaxis/pmaxis/libs/redis-client"
	"github.com/pmaxis/pmaxis/libs/service"
	"github.com/pmaxis/pmaxis/services/api/handlers"
	"github.com/pmaxis/pmaxis/services/api/middleware"
	"github.com/pmaxis/pmaxis/services/api/ws"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

//go:embed openapi.json
var openapiSpec []byte

//go:embed admin.html
var adminHTML []byte

//go:embed register.html
var registerHTML []byte

func main() {
	svc := service.NewBaseService("api")

	svc.Run(context.Background(), func(ctx context.Context) error {
		// Config
		port := config.GetEnv("API_PORT", "8088")
		kafkaBrokers := config.GetEnv("KAFKA_BROKERS", "localhost:9092")
		redisAddr := config.GetEnv("REDIS_ADDR", "localhost:6379")
		chAddr := config.GetEnv("CLICKHOUSE_ADDR", "localhost:9000")
		chUser := config.GetEnv("CLICKHOUSE_USER", "default")
		chPassword := config.GetEnv("CLICKHOUSE_PASSWORD", "")
		chDatabase := config.GetEnv("CLICKHOUSE_DB", "default")
		pgURL := config.GetEnv("POSTGRES_URL", "postgres://postgres:postgres@localhost:5432/pmaxis")
		gammaBaseURL := config.GetEnv("GAMMA_API_BASE_URL", "https://gamma-api.polymarket.com")
		dataAPIURL := config.GetEnv("DATA_API_URL", "https://data-api.polymarket.com")
		clobAPIURL := config.GetEnv("CLOB_API_URL", "https://clob.polymarket.com")

		// Clients
		rdb := redisclient.NewClient(redisclient.Options{Addr: redisAddr}, svc.Logger)
		ch, err := clickhouse.NewClient(clickhouse.Options{
			Addr:     chAddr,
			User:     chUser,
			Password: chPassword,
			Database: chDatabase,
		}, svc.Logger)
		if err != nil {
			return fmt.Errorf("failed to connect to clickhouse: %w", err)
		}
		pg, err := postgres.NewClient(ctx, pgURL, 5, svc.Logger)
		if err != nil {
			return fmt.Errorf("failed to connect to postgres: %w", err)
		}
		defer pg.Close()

		h := &handlers.APIHandler{
			Redis:      rdb,
			ClickHouse: ch,
			Postgres:   pg,
			Logger:     svc.Logger,
			GammaURL:   gammaBaseURL,
			DataURL:    dataAPIURL,
			ClobURL:    clobAPIURL,
		}

		// Auto-migrate api_keys table
		if err := h.MigrateAPIKeys(ctx); err != nil {
			return fmt.Errorf("failed to migrate api_keys table: %w", err)
		}

		// Router
		r := mux.NewRouter()
		r.Use(middleware.CORS)
		r.Use(middleware.SecurityHeaders)

		// Public Routes
		r.HandleFunc("/health", h.GetHealth).Methods("GET")

		// Metrics — gated behind DEBUG_TOKEN
		r.Handle("/metrics", middleware.DebugAuth(promhttp.Handler()))

		// WebSocket Gateway
		hub := ws.NewHub(svc.Logger)
		go hub.Run(ctx)
		r.Handle("/stream", middleware.AuthWS(rdb, svc.Logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ws.ServeWS(hub, w, r, svc.Logger)
		})))

		// Real-time Kafka → WebSocket broadcaster
		go ws.StartKafkaBroadcaster(ctx, hub, kafkaBrokers, svc.Logger)

		// API Docs
		r.HandleFunc("/openapi.json", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Write(openapiSpec)
		}).Methods("GET")
		r.HandleFunc("/docs", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Security-Policy", "default-src 'none'; script-src https://cdn.jsdelivr.net 'unsafe-eval' 'unsafe-inline'; style-src https://cdn.jsdelivr.net 'unsafe-inline'; img-src 'self' data: https:; connect-src *; font-src https://cdn.jsdelivr.net data:; worker-src blob:")
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.Write([]byte(`<!DOCTYPE html>
<html>
<head>
  <title>PMAxis API Docs</title>
  <meta charset="utf-8"/>
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <style>
    *, *::before, *::after { box-sizing: border-box; }
    body { margin: 0; }
    #theme-bar {
      position: fixed;
      top: 12px;
      right: 16px;
      z-index: 99999;
      display: flex;
      align-items: center;
      gap: 8px;
      background: rgba(15,15,25,0.75);
      border: 1px solid rgba(255,255,255,0.12);
      border-radius: 10px;
      padding: 6px 12px;
      backdrop-filter: blur(10px);
      font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif;
      font-size: 13px;
      color: rgba(255,255,255,0.8);
      box-shadow: 0 4px 16px rgba(0,0,0,0.4);
    }
    #theme-bar select {
      background: transparent;
      border: none;
      color: rgba(255,255,255,0.9);
      font-size: 13px;
      cursor: pointer;
      outline: none;
      padding: 0 4px;
    }
    #theme-bar select option { background: #111; color: #fff; }
  </style>
</head>
<body>
  <div id="theme-bar">
    <span>🎨 Theme</span>
    <select id="theme-select" onchange="applyTheme(this.value)">
      <option value="deepSpace">Deep Space</option>
      <option value="purple">Purple</option>
      <option value="moon">Moon</option>
      <option value="bluePlanet">Blue Planet</option>
      <option value="saturn">Saturn</option>
      <option value="solarized">Solarized</option>
      <option value="default">Default</option>
    </select>
  </div>
  <script id="api-reference" data-url="/openapi.json"></script>
  <script>
    var ROUNDED_CSS = [
      '.scalar-card { border-radius: 14px !important; }',
      '.scalar-button, button { border-radius: 8px !important; }',
      'input, textarea, select { border-radius: 7px !important; }',
      '.endpoint-details-card { border-radius: 12px !important; }',
      '.scalar-api-client__send-button { border-radius: 8px !important; }',
      'code, pre { border-radius: 8px !important; }',
      '.tag-section { border-radius: 14px !important; }',
    ].join('\n');
    function buildConfig(theme) {
      return JSON.stringify({ theme: theme, layout: 'modern', customCss: ROUNDED_CSS });
    }
    function applyTheme(theme) {
      localStorage.setItem('pmaxis-docs-theme', theme);
      location.reload();
    }
    var saved = localStorage.getItem('pmaxis-docs-theme') || 'deepSpace';
    document.getElementById('theme-select').value = saved;
    document.getElementById('api-reference').setAttribute('data-configuration', buildConfig(saved));
  </script>
  <script src="https://cdn.jsdelivr.net/npm/@scalar/api-reference"></script>
</body>
</html>`))
		}).Methods("GET")

		// Debug Endpoints — gated by X-Debug-Token header
		debug := r.PathPrefix("/debug").Subrouter()
		debug.Use(middleware.DebugAuth)
		debug.HandleFunc("/markets", h.GetDebugMarkets).Methods("GET")
		debug.HandleFunc("/markets/{id}", h.GetDebugMarket).Methods("GET")
		debug.HandleFunc("/trades", h.GetDebugTrades).Methods("GET")
		debug.HandleFunc("/orderbook/{id}", h.GetDebugOrderbook).Methods("GET")
		debug.HandleFunc("/signals", h.GetDebugSignals).Methods("GET")
		debug.HandleFunc("/signals/{id}", h.GetDebugSignalsByMarket).Methods("GET")

		// Admin Panel — gated by DEBUG_TOKEN (same as debug routes)
		admin := r.PathPrefix("/admin").Subrouter()
		admin.Use(middleware.DebugAuth)
		admin.HandleFunc("", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'; connect-src 'self'")
			w.Write(adminHTML)
		}).Methods("GET")
		admin.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, "/admin", http.StatusMovedPermanently)
		}).Methods("GET")
		admin.HandleFunc("/keys", h.AdminListKeys).Methods("GET")
		admin.HandleFunc("/keys", h.AdminCreateKey).Methods("POST")
		admin.HandleFunc("/keys/stats", h.AdminGetStats).Methods("GET")
		admin.HandleFunc("/keys/{key}", h.AdminGetKey).Methods("GET")
		admin.HandleFunc("/keys/{key}", h.AdminUpdateKey).Methods("PUT")
		admin.HandleFunc("/keys/{key}/revoke", h.AdminRevokeKey).Methods("POST")
		admin.HandleFunc("/keys/{key}/activate", h.AdminActivateKey).Methods("POST")
		admin.HandleFunc("/keys/{key}/reset", h.AdminResetUsage).Methods("POST")
		admin.HandleFunc("/keys/{key}/usage", h.AdminGetUsage).Methods("GET")

		// Developer Registration Page — public
		r.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'; connect-src 'self'")
			w.Write(registerHTML)
		}).Methods("GET")

		// Public key registration endpoint (no API key required)
		r.HandleFunc("/v1/keys/register", h.RegisterKey).Methods("POST")

		// Protected API Routes — require valid API key + rate limited per tier
		api := r.PathPrefix("/v1").Subrouter()
		api.Use(middleware.Auth(rdb, svc.Logger))
		api.Use(middleware.RateLimit(rdb, svc.Logger))

		api.HandleFunc("/keys/rotate", h.RotateKey).Methods("POST")
		api.HandleFunc("/markets", h.GetMarkets).Methods("GET")
		api.HandleFunc("/markets/search", h.SearchMarkets).Methods("GET")
		api.HandleFunc("/markets/top", h.GetTopMarkets).Methods("GET")
		api.HandleFunc("/prices", h.GetBatchPrices).Methods("GET")
		api.HandleFunc("/events", h.GetEvents).Methods("GET")
		api.HandleFunc("/categories", h.GetCategories).Methods("GET")
		api.HandleFunc("/tags", h.GetTags).Methods("GET")
		api.HandleFunc("/series", h.GetSeries).Methods("GET")
		api.HandleFunc("/markets/{id}", h.GetMarket).Methods("GET")
		api.HandleFunc("/markets/{id}/orderbook", h.GetOrderbook).Methods("GET")
		api.HandleFunc("/markets/{id}/price", h.GetPrice).Methods("GET")
		api.HandleFunc("/markets/{id}/price-history", h.GetMarketPriceHistory).Methods("GET")
		api.HandleFunc("/markets/{id}/trades", h.GetMarketTrades).Methods("GET")
		api.HandleFunc("/markets/{id}/summary", h.GetMarketSummary).Methods("GET")
		api.HandleFunc("/markets/{id}/stats", h.GetMarketStats).Methods("GET")
		api.HandleFunc("/markets/{id}/candles", h.GetCandles).Methods("GET")
		api.HandleFunc("/markets/{id}/positions", h.GetMarketPositions).Methods("GET")
		api.HandleFunc("/markets/{id}/related", h.GetRelatedMarkets).Methods("GET")
		api.HandleFunc("/positions", h.GetPositions).Methods("GET")
		api.HandleFunc("/positions/closed", h.GetClosedPositions).Methods("GET")
		api.HandleFunc("/markets/{id}/signals", h.GetSignalsByMarket).Methods("GET")

		// Wallet Activity Routes
		api.HandleFunc("/wallets/watch", h.WatchWallet).Methods("POST")
		api.HandleFunc("/wallets/watched", h.GetWatchedWallets).Methods("GET")
		api.HandleFunc("/wallets/{address}/watch", h.UnwatchWallet).Methods("DELETE")
		api.HandleFunc("/wallets/{address}/activity", h.GetWalletActivity).Methods("GET")
		api.HandleFunc("/wallets/{address}/onchain", h.GetWalletOnchain).Methods("GET")
		api.HandleFunc("/wallets/{address}/summary", h.GetWalletSummary).Methods("GET")

		// GraphQL Placeholder
		r.HandleFunc("/graphql", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(`{"data": {"message": "GraphQL API placeholder"}}`))
		}).Methods("POST")

		// Log registered routes
		r.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
			path, _ := route.GetPathTemplate()
			svc.Logger.Info("Registered route", "path", path)
			return nil
		})

		srv := &http.Server{
			Handler:        r,
			Addr:           ":" + port,
			WriteTimeout:   15 * time.Second,
			ReadTimeout:    15 * time.Second,
			IdleTimeout:    60 * time.Second,
			MaxHeaderBytes: 1 << 16, // 64 KB
		}

		svc.Logger.Info("API server starting", "port", port)

		go func() {
			if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				svc.Logger.Error("listen error", "error", err)
			}
		}()

		<-ctx.Done()

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return srv.Shutdown(shutdownCtx)
	})
}
