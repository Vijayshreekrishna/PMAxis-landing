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

//go:embed viz.html
var vizHTML []byte

//go:embed status.html
var statusHTML []byte

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
		r.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'; connect-src 'self'")
			w.Write(statusHTML)
		}).Methods("GET")

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

		// Background uptime recorder — runs every 5 minutes, stores to Redis
		go func() {
			uptimeTick := time.NewTicker(5 * time.Minute)
			defer uptimeTick.Stop()
			h.RecordUptimeSnapshot(ctx)
			for {
				select {
				case <-ctx.Done():
					return
				case <-uptimeTick.C:
					h.RecordUptimeSnapshot(ctx)
				}
			}
		}()

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
  <title>PMAxis — API Docs</title>
  <meta charset="utf-8"/>
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <style>
    *, *::before, *::after { box-sizing: border-box; margin: 0; padding: 0; }
    body { background: #0A0A0A; font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif; -webkit-font-smoothing: antialiased; }
    .docs-nav {
      display: flex; align-items: center; justify-content: space-between;
      padding: 0 32px; height: 68px;
      background: #0A0A0A; border-bottom: 1px solid #1E1E1E;
      position: sticky; top: 0; z-index: 99999;
    }
    .docs-brand { display: flex; align-items: center; gap: 10px; text-decoration: none; }
    .docs-brand-name { font-size: 16px; font-weight: 700; color: #fff; letter-spacing: -0.02em; }
    .docs-brand-divider { width: 1px; height: 16px; background: #1E1E1E; margin: 0 4px; }
    .docs-brand-sub { font-size: 13px; color: #A1A1AA; font-weight: 400; }
    .docs-nav-links { display: flex; align-items: center; gap: 28px; }
    .docs-nav-link { font-size: 13px; color: #A1A1AA; text-decoration: none; font-weight: 500; transition: color 0.15s; }
    .docs-nav-link:hover { color: #fff; }
    .docs-nav-btn {
      font-size: 13px; font-weight: 600; color: #0A0A0A;
      background: #00E676; border: none; border-radius: 8px;
      padding: 9px 18px; cursor: pointer; text-decoration: none;
      transition: opacity 0.15s;
    }
    .docs-nav-btn:hover { opacity: 0.88; }
  </style>
</head>
<body>
  <nav class="docs-nav">
    <a class="docs-brand" href="/">
      <svg width="38" height="38" viewBox="0 0 803 795" fill="none" xmlns="http://www.w3.org/2000/svg">
        <path fill="#FFFFFF" d="M719.962 114.503C724.439 116.738 746.095 136.885 751.202 141.275C743.13 152.558 727.925 169.756 718.52 180.14C667.116 237.986 604.881 285.207 535.329 319.136C528.842 322.253 501.635 334.541 495.686 335.719C493.671 334.794 493.692 334.098 492.165 332.003C481.767 318.552 471.393 311.209 457.07 302.886C532.805 280.515 608.565 231.922 664.325 176.745C684.543 156.739 701.958 136.391 719.962 114.503Z"/>
        <path fill="#FFFFFF" d="M103.731 114.306C106.532 116.771 116.373 129.166 119.413 132.747C128.996 144.095 139.01 155.071 149.433 165.651C213.595 230.118 280.396 274.662 366.923 302.87C352.429 310.952 342.084 318.858 331.657 332.058L328.841 336.043C319.39 333.204 296.981 322.854 288.065 318.523C216.559 283.785 152.851 232.607 99.9718 173.495C90.9908 163.455 80.8168 152.054 72.5908 141.466C80.8318 133.892 95.1408 120.907 103.731 114.306Z"/>
        <path fill="#FFFFFF" d="M500.639 448.854C510.914 451.537 533.17 462.09 542.691 466.637C603.572 495.713 657.966 537.098 703.779 586.511C719.932 603.934 737.715 624.431 751.412 643.924C743.385 651.277 729.34 662.024 720.529 669.194C717.837 667.836 711.304 658.002 708.901 655.026C699.702 643.633 690.268 632.508 680.354 621.743C622.903 558.485 550.473 510.682 469.721 482.728C482.46 472.687 492.344 462.846 500.639 448.854Z"/>
        <path fill="#FFFFFF" d="M322.57 449.109C324.527 450.34 331.226 460.882 333.985 463.96C341.099 471.897 346.509 476.457 354.784 482.779C251.686 517.317 170.299 583.28 104.517 668.566L103.202 668.709C97.76 665.303 78.849 649.187 72.52 644.052C81.2 630.876 98.635 610.478 109.165 598.717C157.332 544.918 214.867 498.845 280.012 467.285C293.255 460.869 308.655 453.904 322.57 449.109Z"/>
        <path fill="#00E676" d="M404.129 336.369C437.402 331.991 467.935 355.383 472.368 388.649C476.801 421.915 453.459 452.487 420.201 456.975C386.865 461.473 356.206 438.064 351.762 404.721C347.319 371.378 370.778 340.757 404.129 336.369Z"/>
      </svg>
      <span class="docs-brand-name">PMAxis</span>
      <div class="docs-brand-divider"></div>
      <span class="docs-brand-sub">API Reference</span>
    </a>
    <div class="docs-nav-links">
      <a class="docs-nav-link" href="/viz">Explorer</a>
      <a class="docs-nav-link" href="/status">Status</a>
      <a class="docs-nav-btn" href="/register">Get API Key</a>
    </div>
  </nav>
  <script id="api-reference" data-url="/openapi.json" data-configuration='{"theme":"deepSpace","layout":"modern"}'></script>
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
		admin.HandleFunc("/keys/{key}", h.AdminDeleteKey).Methods("DELETE")

			// Data Explorer — public, no auth
			r.HandleFunc("/viz", func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "text/html; charset=utf-8")
				w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline' https://cdn.jsdelivr.net https://unpkg.com; style-src 'self' 'unsafe-inline'; connect-src 'self'; img-src 'self' data:")
				w.Write(vizHTML)
			}).Methods("GET")

			// Public read-only data routes for /viz page (no API key required)
			vizData := r.PathPrefix("/viz/data").Subrouter()
			vizData.Use(middleware.CORS)
			vizData.HandleFunc("/stats", h.GetPlatformStats).Methods("GET")
			vizData.HandleFunc("/trades/recent", h.GetRecentTrades).Methods("GET")
			vizData.HandleFunc("/markets/top", h.GetTopMarkets).Methods("GET")
			vizData.HandleFunc("/markets/{id}/candles", h.GetCandles).Methods("GET")
			vizData.HandleFunc("/markets/{id}/orderbook", h.GetOrderbook).Methods("GET")
			vizData.HandleFunc("/orderbooks/available", h.GetAvailableOrderbooks).Methods("GET")
			vizData.HandleFunc("/uptime", h.GetUptimeHistory).Methods("GET")
			vizData.HandleFunc("/categories", h.GetCategories).Methods("GET")
			vizData.HandleFunc("/tags", h.GetTags).Methods("GET")

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
		api.HandleFunc("/stats", h.GetPlatformStats).Methods("GET")
		api.HandleFunc("/trades/recent", h.GetRecentTrades).Methods("GET")
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
