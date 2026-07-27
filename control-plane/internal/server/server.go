package server

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/hanzoai/agents/control-plane/internal/config"
	"github.com/hanzoai/agents/control-plane/internal/core/interfaces"
	coreservices "github.com/hanzoai/agents/control-plane/internal/core/services" // Core services
	"github.com/hanzoai/agents/control-plane/internal/events"                     // Event system
	"github.com/hanzoai/agents/control-plane/internal/handlers"                   // Agent handlers
	"github.com/hanzoai/agents/control-plane/internal/handlers/ui"                // UI handlers
	"github.com/hanzoai/agents/control-plane/internal/infrastructure/communication"
	"github.com/hanzoai/agents/control-plane/internal/infrastructure/process"
	infrastorage "github.com/hanzoai/agents/control-plane/internal/infrastructure/storage"
	"github.com/hanzoai/agents/control-plane/internal/logger"
	"github.com/hanzoai/agents/control-plane/internal/server/middleware"
	"github.com/hanzoai/agents/control-plane/internal/services" // Services
	"github.com/hanzoai/agents/control-plane/internal/storage"
	"github.com/hanzoai/agents/control-plane/internal/utils"
	client "github.com/hanzoai/agents/control-plane/web/client"

	"github.com/gin-contrib/cors" // CORS middleware
	"github.com/gin-gonic/gin"
	metric "github.com/luxfi/metric"
	"github.com/zap-proto/go/transport"
	"github.com/zap-proto/zip"
)

// HanzoAgentsServer represents the core HanzoAgents orchestration service.
type HanzoAgentsServer struct {
	storage               storage.StorageProvider
	cache                 storage.CacheProvider
	App                   *zip.App
	uiService             *services.UIService           // Add UIService
	executionsUIService   *services.ExecutionsUIService // Add ExecutionsUIService
	healthMonitor         *services.HealthMonitor
	presenceManager       *services.PresenceManager
	statusManager         *services.StatusManager // Add StatusManager for unified status management
	agentService          interfaces.AgentService // Add AgentService for lifecycle management
	agentClient           interfaces.AgentClient  // Add AgentClient for MCP communication
	config                *config.Config
	storageHealthOverride func(context.Context) gin.H
	cacheHealthOverride   func(context.Context) gin.H
	// DID Services
	keystoreService *services.KeystoreService
	didService      *services.DIDService
	vcService       *services.VCService
	didRegistry     *services.DIDRegistry
	hanzoAgentsHome string
	// Cleanup service
	cleanupService         *handlers.ExecutionCleanupService
	payloadStore           services.PayloadStore
	registryWatcherCancel  context.CancelFunc
	adminServer            *transport.Server
	adminPort              int
	webhookDispatcher      services.WebhookDispatcher
	observabilityForwarder services.ObservabilityForwarder
}

// NewHanzoAgentsServer creates a new instance of the HanzoAgentsServer.
func NewHanzoAgentsServer(cfg *config.Config) (*HanzoAgentsServer, error) {
	// Define hanzoAgentsHome at the very top
	hanzoAgentsHome := os.Getenv("HANZO_AGENTS_HOME")
	if hanzoAgentsHome == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, err
		}
		hanzoAgentsHome = filepath.Join(homeDir, ".hanzo-agents")
	}

	dirs, err := utils.EnsureDataDirectories()
	if err != nil {
		return nil, fmt.Errorf("failed to ensure data directories: %w", err)
	}

	factory := &storage.StorageFactory{}
	storageProvider, cacheProvider, err := factory.CreateStorage(cfg.Storage)
	if err != nil {
		return nil, err
	}

	app := zip.New(zip.Config{AppName: "hanzo-agents-control-plane"})

	// Sync installed.yaml to database for package visibility
	_ = SyncPackagesFromRegistry(hanzoAgentsHome, storageProvider)

	// Initialize agent client for communication with agent nodes
	agentClient := communication.NewHTTPAgentClient(storageProvider, 5*time.Second)

	// Create infrastructure components for AgentService
	fileSystem := infrastorage.NewFileSystemAdapter()
	registryPath := filepath.Join(hanzoAgentsHome, "installed.json")
	registryStorage := infrastorage.NewLocalRegistryStorage(fileSystem, registryPath)
	processManager := process.NewProcessManager()
	portManager := process.NewPortManager()

	// Create AgentService
	agentService := coreservices.NewAgentService(processManager, portManager, registryStorage, agentClient, hanzoAgentsHome)

	// Initialize StatusManager for unified status management
	statusManagerConfig := services.StatusManagerConfig{
		ReconcileInterval:       30 * time.Second,
		StatusCacheTTL:          5 * time.Minute,
		MaxTransitionTime:       2 * time.Minute,
		HeartbeatStaleThreshold: cfg.HanzoAgents.NodeHealth.HeartbeatStaleThreshold,
	}

	// Create UIService first (without StatusManager)
	uiService := services.NewUIService(storageProvider, agentClient, agentService, nil)

	// Create StatusManager with UIService and AgentClient
	statusManager := services.NewStatusManager(storageProvider, statusManagerConfig, uiService, agentClient)

	// Update UIService with StatusManager reference
	uiService = services.NewUIService(storageProvider, agentClient, agentService, statusManager)

	// Presence manager tracks node leases so stale nodes age out quickly
	presenceConfig := services.PresenceManagerConfig{
		HeartbeatTTL:  5 * time.Minute,
		SweepInterval: 30 * time.Second,
		HardEvictTTL:  30 * time.Minute,
	}
	presenceManager := services.NewPresenceManager(statusManager, presenceConfig)

	executionsUIService := services.NewExecutionsUIService(storageProvider) // Initialize ExecutionsUIService

	// Initialize health monitor with configurable settings
	healthMonitorConfig := services.HealthMonitorConfig{
		CheckInterval:       cfg.HanzoAgents.NodeHealth.CheckInterval,
		CheckTimeout:        cfg.HanzoAgents.NodeHealth.CheckTimeout,
		ConsecutiveFailures: cfg.HanzoAgents.NodeHealth.ConsecutiveFailures,
		RecoveryDebounce:    cfg.HanzoAgents.NodeHealth.RecoveryDebounce,
	}
	healthMonitor := services.NewHealthMonitor(storageProvider, healthMonitorConfig, uiService, agentClient, statusManager, presenceManager)
	presenceManager.SetExpireCallback(healthMonitor.UnregisterAgent)

	// Initialize DID services if enabled
	var keystoreService *services.KeystoreService
	var didService *services.DIDService
	var vcService *services.VCService
	var didRegistry *services.DIDRegistry

	if cfg.Features.DID.Enabled {
		fmt.Println("🔐 Initializing DID and VC services...")

		// Use universal path management for DID directories
		dirs, err := utils.EnsureDataDirectories()
		if err != nil {
			return nil, fmt.Errorf("failed to create DID directories: %w", err)
		}

		// Update keystore path to use universal paths
		if cfg.Features.DID.Keystore.Path == "./data/keys" {
			cfg.Features.DID.Keystore.Path = dirs.KeysDir
		}

		fmt.Printf("🔑 Creating keystore service at: %s\n", cfg.Features.DID.Keystore.Path)
		// Instantiate services in dependency order: Keystore → DID → VC, Registry
		keystoreService, err = services.NewKeystoreService(&cfg.Features.DID.Keystore)
		if err != nil {
			return nil, fmt.Errorf("failed to create keystore service: %w", err)
		}

		fmt.Println("📋 Creating DID registry...")
		didRegistry = services.NewDIDRegistryWithStorage(storageProvider)

		fmt.Println("🆔 Creating DID service...")
		didService = services.NewDIDService(&cfg.Features.DID, keystoreService, didRegistry)

		fmt.Println("📜 Creating VC service...")
		vcService = services.NewVCService(&cfg.Features.DID, didService, storageProvider)

		// Initialize services
		fmt.Println("🔧 Initializing DID registry...")
		if err = didRegistry.Initialize(); err != nil {
			return nil, fmt.Errorf("failed to initialize DID registry: %w", err)
		}

		fmt.Println("🔧 Initializing VC service...")
		if err = vcService.Initialize(); err != nil {
			return nil, fmt.Errorf("failed to initialize VC service: %w", err)
		}

		// Generate af server ID based on hanzo-agents home directory
		hanzoAgentsServerID := generateHanzoAgentsServerID(hanzoAgentsHome)

		// Initialize af server DID with dynamic ID
		fmt.Printf("🧠 Initializing af server DID (ID: %s)...\n", hanzoAgentsServerID)
		if err := didService.Initialize(hanzoAgentsServerID); err != nil {
			return nil, fmt.Errorf("failed to initialize af server DID: %w", err)
		}

		// Validate that af server DID was successfully created
		registry, err := didService.GetRegistry(hanzoAgentsServerID)
		if err != nil {
			return nil, fmt.Errorf("failed to validate af server DID creation: %w", err)
		}
		if registry == nil || registry.RootDID == "" {
			return nil, fmt.Errorf("af server DID validation failed: registry or root DID is empty")
		}

		fmt.Printf("✅ HanzoAgents server DID created successfully: %s\n", registry.RootDID)

		// Backfill existing nodes with DIDs
		fmt.Println("🔄 Starting DID backfill for existing nodes...")
		ctx := context.Background()
		if err := didService.BackfillExistingNodes(ctx, storageProvider); err != nil {
			fmt.Printf("⚠️ DID backfill failed: %v\n", err)
		}

		fmt.Println("✅ DID and VC services initialized successfully!")
	} else {
		fmt.Println("⚠️ DID and VC services are DISABLED in configuration")
	}

	payloadStore := services.NewFilePayloadStore(dirs.PayloadsDir)

	webhookDispatcher := services.NewWebhookDispatcher(storageProvider, services.WebhookDispatcherConfig{
		Timeout:         cfg.HanzoAgents.ExecutionQueue.WebhookTimeout,
		MaxAttempts:     cfg.HanzoAgents.ExecutionQueue.WebhookMaxAttempts,
		RetryBackoff:    cfg.HanzoAgents.ExecutionQueue.WebhookRetryBackoff,
		MaxRetryBackoff: cfg.HanzoAgents.ExecutionQueue.WebhookMaxRetryBackoff,
	})
	if err := webhookDispatcher.Start(context.Background()); err != nil {
		logger.Logger.Warn().Err(err).Msg("failed to start webhook dispatcher")
	}

	// Initialize observability forwarder for external webhook integration
	observabilityForwarder := services.NewObservabilityForwarder(storageProvider, services.ObservabilityForwarderConfig{
		BatchSize:       10,
		BatchTimeout:    time.Second,
		HTTPTimeout:     10 * time.Second,
		MaxAttempts:     3,
		RetryBackoff:    time.Second,
		MaxRetryBackoff: 30 * time.Second,
		WorkerCount:     2,
		QueueSize:       1000,
	})
	if err := observabilityForwarder.Start(context.Background()); err != nil {
		logger.Logger.Warn().Err(err).Msg("failed to start observability forwarder")
	}

	// Initialize execution cleanup service
	cleanupService := handlers.NewExecutionCleanupService(storageProvider, cfg.HanzoAgents.ExecutionCleanup)

	adminPort := cfg.HanzoAgents.Port + 100
	if envPort := os.Getenv("HANZO_AGENTS_ADMIN_GRPC_PORT"); envPort != "" {
		if parsedPort, parseErr := strconv.Atoi(envPort); parseErr == nil {
			adminPort = parsedPort
		} else {
			logger.Logger.Warn().Err(parseErr).Str("value", envPort).Msg("invalid HANZO_AGENTS_ADMIN_GRPC_PORT, using default offset")
		}
	}

	return &HanzoAgentsServer{
		storage:                storageProvider,
		cache:                  cacheProvider,
		App:                    app,
		uiService:              uiService,
		executionsUIService:    executionsUIService,
		healthMonitor:          healthMonitor,
		presenceManager:        presenceManager,
		statusManager:          statusManager,
		agentService:           agentService,
		agentClient:            agentClient,
		config:                 cfg,
		keystoreService:        keystoreService,
		didService:             didService,
		vcService:              vcService,
		didRegistry:            didRegistry,
		hanzoAgentsHome:        hanzoAgentsHome,
		cleanupService:         cleanupService,
		payloadStore:           payloadStore,
		webhookDispatcher:      webhookDispatcher,
		observabilityForwarder: observabilityForwarder,
		registryWatcherCancel:  nil,
		adminPort:              adminPort,
	}, nil
}

// Start initializes and starts the HanzoAgentsServer.
func (s *HanzoAgentsServer) Start() error {
	// Setup routes
	s.setupRoutes()

	// Start status manager service in background
	go s.statusManager.Start()

	if s.presenceManager != nil {
		go s.presenceManager.Start()

		// Recover presence leases from database
		go func() {
			ctx := context.Background()
			if err := s.presenceManager.RecoverFromDatabase(ctx, s.storage); err != nil {
				logger.Logger.Error().Err(err).Msg("Failed to recover presence leases from database")
			}
		}()
	}

	// Start health monitor service in background
	go s.healthMonitor.Start()

	// Recover previously registered nodes and check their health
	go func() {
		ctx := context.Background()
		if err := s.healthMonitor.RecoverFromDatabase(ctx); err != nil {
			logger.Logger.Error().Err(err).Msg("Failed to recover nodes from database")
		}
	}()

	// Start execution cleanup service in background
	ctx := context.Background()
	if err := s.cleanupService.Start(ctx); err != nil {
		logger.Logger.Error().Err(err).Msg("Failed to start execution cleanup service")
		// Don't fail server startup if cleanup service fails to start
	}

	// Start reasoner event heartbeat (30 second intervals)
	events.StartHeartbeat(30 * time.Second)

	// Start node event heartbeat (30 second intervals)
	events.StartNodeHeartbeat(30 * time.Second)

	if s.registryWatcherCancel == nil {
		cancel, err := StartPackageRegistryWatcher(context.Background(), s.hanzoAgentsHome, s.storage)
		if err != nil {
			logger.Logger.Error().Err(err).Msg("failed to start package registry watcher")
		} else {
			s.registryWatcherCancel = cancel
		}
	}

	if err := s.startAdminServer(); err != nil {
		return fmt.Errorf("failed to start admin server: %w", err)
	}

	// Start HTTP server. The scheme is explicit: a bare address would bind
	// zip's default ZAP transport, and this is the service's HTTP REST
	// surface. ZAP is served separately by the admin server.
	return s.App.Listen("http://:" + strconv.Itoa(s.config.HanzoAgents.Port))
}

// Stop gracefully shuts down the HanzoAgentsServer.
func (s *HanzoAgentsServer) Stop() error {
	if s.adminServer != nil {
		_ = s.adminServer.Close()
	}

	if s.App != nil {
		_ = s.App.Shutdown()
	}

	// Stop status manager service
	if s.statusManager != nil {
		s.statusManager.Stop()
	}

	if s.presenceManager != nil {
		s.presenceManager.Stop()
	}

	// Stop health monitor service
	s.healthMonitor.Stop()

	// Stop execution cleanup service
	if s.cleanupService != nil {
		if err := s.cleanupService.Stop(); err != nil {
			logger.Logger.Error().Err(err).Msg("Failed to stop execution cleanup service")
		}
	}

	if s.registryWatcherCancel != nil {
		s.registryWatcherCancel()
		s.registryWatcherCancel = nil
	}

	// Stop UI service heartbeat
	if s.uiService != nil {
		s.uiService.StopHeartbeat()
	}

	// Stop observability forwarder
	if s.observabilityForwarder != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := s.observabilityForwarder.Stop(ctx); err != nil {
			logger.Logger.Error().Err(err).Msg("Failed to stop observability forwarder")
		}
	}

	// TODO: Implement graceful shutdown for HTTP, WebSocket, gRPC
	return nil
}

// unregisterAgentFromMonitoring removes an agent from health monitoring
func (s *HanzoAgentsServer) unregisterAgentFromMonitoring(c *gin.Context) {
	nodeID := c.Param("node_id")
	if nodeID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "node_id is required"})
		return
	}

	if s.healthMonitor != nil {
		s.healthMonitor.UnregisterAgent(nodeID)
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": fmt.Sprintf("Agent %s unregistered from health monitoring", nodeID),
		})
	} else {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "health monitor not available"})
	}
}

// healthCheckHandler provides comprehensive health check for container orchestration
func (s *HanzoAgentsServer) healthCheckHandler(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	healthStatus := gin.H{
		"status":    "healthy",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"version":   "1.0.0", // TODO: Get from build info
		"checks":    gin.H{},
	}

	allHealthy := true
	checks := healthStatus["checks"].(gin.H)

	// Storage health check
	if s.storage != nil || s.storageHealthOverride != nil {
		storageHealth := s.checkStorageHealth(ctx)
		checks["storage"] = storageHealth
		if storageHealth["status"] != "healthy" {
			allHealthy = false
		}
	} else {
		checks["storage"] = gin.H{
			"status":  "unhealthy",
			"message": "storage not initialized",
		}
		allHealthy = false
	}

	// Cache health check
	if s.cache != nil || s.cacheHealthOverride != nil {
		cacheHealth := s.checkCacheHealth(ctx)
		checks["cache"] = cacheHealth
		if cacheHealth["status"] != "healthy" {
			allHealthy = false
		}
	} else {
		checks["cache"] = gin.H{
			"status":  "healthy",
			"message": "cache not configured (optional)",
		}
	}

	// Overall status
	if !allHealthy {
		healthStatus["status"] = "unhealthy"
		c.JSON(http.StatusServiceUnavailable, healthStatus)
		return
	}

	c.JSON(http.StatusOK, healthStatus)
}

// checkStorageHealth performs storage-specific health checks
func (s *HanzoAgentsServer) checkStorageHealth(ctx context.Context) gin.H {
	if s.storageHealthOverride != nil {
		return s.storageHealthOverride(ctx)
	}

	startTime := time.Now()

	// For local storage, try a basic operation
	if err := ctx.Err(); err != nil {
		return gin.H{
			"status":  "unhealthy",
			"message": "context timeout during storage check",
		}
	}

	return gin.H{
		"status":        "healthy",
		"message":       "storage is responsive",
		"response_time": time.Since(startTime).Milliseconds(),
	}
}

// checkCacheHealth performs cache-specific health checks
func (s *HanzoAgentsServer) checkCacheHealth(ctx context.Context) gin.H {
	if s.cacheHealthOverride != nil {
		return s.cacheHealthOverride(ctx)
	}

	startTime := time.Now()

	// Try a simple cache operation
	testKey := "health_check_" + fmt.Sprintf("%d", time.Now().Unix())
	testValue := "ok"

	// Set a test value
	if err := s.cache.Set(testKey, testValue, time.Minute); err != nil {
		return gin.H{
			"status":        "unhealthy",
			"message":       fmt.Sprintf("cache set operation failed: %v", err),
			"response_time": time.Since(startTime).Milliseconds(),
		}
	}

	// Get the test value
	var retrieved string
	if err := s.cache.Get(testKey, &retrieved); err != nil {
		return gin.H{
			"status":        "unhealthy",
			"message":       fmt.Sprintf("cache get operation failed: %v", err),
			"response_time": time.Since(startTime).Milliseconds(),
		}
	}

	// Clean up
	if err := s.cache.Delete(testKey); err != nil {
		return gin.H{
			"status":        "unhealthy",
			"message":       fmt.Sprintf("cache delete operation failed: %v", err),
			"response_time": time.Since(startTime).Milliseconds(),
		}
	}

	return gin.H{
		"status":        "healthy",
		"message":       "cache is responsive",
		"response_time": time.Since(startTime).Milliseconds(),
	}
}

func (s *HanzoAgentsServer) setupRoutes() {
	// Configure CORS from configuration
	corsConfig := cors.Config{
		AllowOrigins:     s.config.API.CORS.AllowedOrigins,
		AllowMethods:     s.config.API.CORS.AllowedMethods,
		AllowHeaders:     s.config.API.CORS.AllowedHeaders,
		ExposeHeaders:    s.config.API.CORS.ExposedHeaders,
		AllowCredentials: s.config.API.CORS.AllowCredentials,
	}

	// Fallback to defaults if not configured
	if len(corsConfig.AllowOrigins) == 0 {
		corsConfig.AllowOrigins = []string{"http://localhost:3000", "http://localhost:5173"}
	}
	if len(corsConfig.AllowMethods) == 0 {
		corsConfig.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	}
	if len(corsConfig.AllowHeaders) == 0 {
		corsConfig.AllowHeaders = []string{"Origin", "Content-Type", "Accept", "Authorization", "X-API-Key"}
	}

	// The gin middleware stack runs unmodified behind one zip middleware, so
	// CORS origin echoing, the access-log format and the auth abort bodies
	// stay byte-for-byte identical. Order is preserved.
	s.App.Use(ginChain(
		cors.New(corsConfig),

		// Request logging middleware
		gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
			return fmt.Sprintf("%s - [%s] \"%s %s %s %d %s \"%s\" %s\"\n",
				param.ClientIP,
				param.TimeStamp.Format(time.RFC1123),
				param.Method,
				param.Path,
				param.Request.Proto,
				param.StatusCode,
				param.Latency,
				param.Request.UserAgent(),
				param.ErrorMessage,
			)
		}),

		// Timeout middleware for all routes (1 hour for long-running executions)
		func(c *gin.Context) {
			// Set a timeout for the request
			ctx := c.Request.Context()
			timeoutCtx, cancel := context.WithTimeout(ctx, 3600*time.Second)
			defer cancel()

			c.Request = c.Request.WithContext(timeoutCtx)
			c.Next()
		},

		// API key authentication (supports headers + api_key query param)
		middleware.APIKeyAuth(middleware.AuthConfig{
			APIKey:    s.config.API.Auth.APIKey,
			SkipPaths: s.config.API.Auth.SkipPaths,
		}),
	))
	if s.config.API.Auth.APIKey != "" {
		logger.Logger.Info().Msg("🔐 API key authentication enabled")
	}

	// Expose Prometheus metrics
	s.App.Get("/metrics", zip.AdaptNetHTTP(metric.NewHTTPHandler(metric.DefaultRegistry, metric.HandlerOpts{})))

	// Public health check endpoint for load balancers and container orchestration (e.g., Railway, K8s)
	s.App.Get("/health", ginHandler(s.healthCheckHandler))

	// Serve UI files - embedded or filesystem based on availability
	if s.config.UI.Enabled {
		// Check if UI is embedded in the binary
		if s.config.UI.Mode == "embedded" && client.IsUIEmbedded() {
			// Use embedded UI
			client.RegisterUIRoutes(s.App)
			fmt.Println("Using embedded UI files")
		} else {
			// Use filesystem UI
			distPath := s.config.UI.DistPath
			if distPath == "" {
				// Get the executable path and find UI dist relative to it
				execPath, err := os.Executable()
				if err != nil {
					distPath = filepath.Join("apps", "platform", "hanzo-agents", "web", "client", "dist")
					if _, statErr := os.Stat(distPath); os.IsNotExist(statErr) {
						distPath = filepath.Join("web", "client", "dist")
					}
				} else {
					execDir := filepath.Dir(execPath)
					// Look for web/client/dist relative to the executable directory
					distPath = filepath.Join(execDir, "web", "client", "dist")

					// If that doesn't exist, try going up one level (if binary is in apps/platform/hanzo-agents/)
					if _, err := os.Stat(distPath); os.IsNotExist(err) {
						distPath = filepath.Join(filepath.Dir(execDir), "apps", "platform", "hanzo-agents", "web", "client", "dist")
					}

					// Final fallback to current working directory
					if _, err := os.Stat(distPath); os.IsNotExist(err) {
						altPath := filepath.Join("apps", "platform", "hanzo-agents", "web", "client", "dist")
						if _, altErr := os.Stat(altPath); altErr == nil {
							distPath = altPath
						} else {
							distPath = filepath.Join("web", "client", "dist")
						}
					}
				}
			}

			// Serve static files from filesystem. gin's StaticFS registered
			// GET+HEAD over the same http.FileServer; keep both verbs.
			staticUI := zip.AdaptNetHTTP(http.StripPrefix("/ui", http.FileServer(http.Dir(distPath))))
			s.App.Get("/ui/*", staticUI)
			s.App.Head("/ui/*", staticUI)

			// Root redirect
			s.App.Get("/", func(c *zip.Ctx) error {
				return c.Redirect(http.StatusMovedPermanently, "/ui/")
			})

			fmt.Printf("Using filesystem UI files from: %s\n", distPath)
		}
	}

	// UI API routes - Moved before API routes to prevent route conflicts
	if s.config.UI.Enabled { // Only add UI API routes if UI is generally enabled
		uiAPI := s.App.Group("/v1/ui")
		{
			// Agents management group - All agent-related operations
			agents := uiAPI.Group("/agents")
			{
				// Package API endpoints
				packagesHandler := ui.NewPackageHandler(s.storage)
				agents.Get("/packages", ginHandler(packagesHandler.ListPackagesHandler))
				agents.Get("/packages/:packageId/details", ginHandler(packagesHandler.GetPackageDetailsHandler))

				// Agent lifecycle management endpoints
				lifecycleHandler := ui.NewLifecycleHandler(s.storage, s.agentService)
				agents.Get("/running", ginHandler(lifecycleHandler.ListRunningAgentsHandler))

				// Individual agent operations
				agents.Get("/:agentId/details", ginHandler(func(c *gin.Context) {
					// TODO: Implement agent details
					c.JSON(http.StatusOK, gin.H{"message": "Agent details endpoint"})
				}))
				agents.Get("/:agentId/status", ginHandler(lifecycleHandler.GetAgentStatusHandler))
				agents.Post("/:agentId/start", ginHandler(lifecycleHandler.StartAgentHandler))
				agents.Post("/:agentId/stop", ginHandler(lifecycleHandler.StopAgentHandler))
				agents.Post("/:agentId/reconcile", ginHandler(lifecycleHandler.ReconcileAgentHandler))

				// Configuration endpoints
				configHandler := ui.NewConfigHandler(s.storage)
				agents.Get("/:agentId/config/schema", ginHandler(configHandler.GetConfigSchemaHandler))
				agents.Get("/:agentId/config", ginHandler(configHandler.GetConfigHandler))
				agents.Post("/:agentId/config", ginHandler(configHandler.SetConfigHandler))

				// Environment file endpoints
				envHandler := ui.NewEnvHandler(s.storage, s.agentService, s.hanzoAgentsHome)
				agents.Get("/:agentId/env", ginHandler(envHandler.GetEnvHandler))
				agents.Put("/:agentId/env", ginHandler(envHandler.PutEnvHandler))
				agents.Patch("/:agentId/env", ginHandler(envHandler.PatchEnvHandler))
				agents.Delete("/:agentId/env/:key", ginHandler(envHandler.DeleteEnvVarHandler))

				// Agent execution history endpoints
				agentExecutionHandler := ui.NewExecutionHandler(s.storage, s.payloadStore, s.webhookDispatcher)
				agents.Get("/:agentId/executions", ginHandler(agentExecutionHandler.ListExecutionsHandler))
				agents.Get("/:agentId/executions/:executionId", ginHandler(agentExecutionHandler.GetExecutionDetailsHandler))
			}

			// Nodes management group - All node-related operations
			nodes := uiAPI.Group("/nodes")
			{
				// Nodes UI endpoints
				uiNodesHandler := ui.NewNodesHandler(s.uiService)
				nodes.Get("/summary", ginHandler(uiNodesHandler.GetNodesSummaryHandler))
				nodes.Get("/events", ginHandler(uiNodesHandler.StreamNodeEventsHandler))

				// Unified status endpoints
				nodes.Get("/:nodeId/status", ginHandler(uiNodesHandler.GetNodeStatusHandler))
				nodes.Post("/:nodeId/status/refresh", ginHandler(uiNodesHandler.RefreshNodeStatusHandler))
				nodes.Post("/status/bulk", ginHandler(uiNodesHandler.BulkNodeStatusHandler))
				nodes.Post("/status/refresh", ginHandler(uiNodesHandler.RefreshAllNodeStatusHandler))

				// Individual node operations
				nodes.Get("/:nodeId/details", ginHandler(uiNodesHandler.GetNodeDetailsHandler))

				// DID and VC management endpoints for nodes
				didHandler := ui.NewDIDHandler(s.storage, s.didService, s.vcService)
				nodes.Get("/:nodeId/did", ginHandler(didHandler.GetNodeDIDHandler))
				nodes.Get("/:nodeId/vc-status", ginHandler(didHandler.GetNodeVCStatusHandler))

				// MCP management endpoints for nodes
				mcpHandler := ui.NewMCPHandler(s.uiService, s.agentClient)
				nodes.Get("/:nodeId/mcp/health", ginHandler(mcpHandler.GetMCPHealthHandler))
				nodes.Get("/:nodeId/mcp/events", ginHandler(mcpHandler.GetMCPEventsHandler))
				nodes.Get("/:nodeId/mcp/metrics", ginHandler(mcpHandler.GetMCPMetricsHandler))
				nodes.Post("/:nodeId/mcp/servers/:alias/restart", ginHandler(mcpHandler.RestartMCPServerHandler))
				nodes.Get("/:nodeId/mcp/servers/:alias/tools", ginHandler(mcpHandler.GetMCPToolsHandler))
			}

			// Executions management group
			executions := uiAPI.Group("/executions")
			{
				// Executions UI endpoints
				uiExecutionsHandler := ui.NewExecutionHandler(s.storage, s.payloadStore, s.webhookDispatcher)
				executions.Get("/summary", ginHandler(uiExecutionsHandler.GetExecutionsSummaryHandler))
				executions.Get("/stats", ginHandler(uiExecutionsHandler.GetExecutionStatsHandler))
				executions.Get("/enhanced", ginHandler(uiExecutionsHandler.GetEnhancedExecutionsHandler))
				executions.Get("/events", ginHandler(uiExecutionsHandler.StreamExecutionEventsHandler))

				// Timeline endpoint for hourly aggregated data
				timelineHandler := ui.NewExecutionTimelineHandler(s.storage)
				executions.Get("/timeline", ginHandler(timelineHandler.GetExecutionTimelineHandler))

				// Recent activity endpoint
				recentActivityHandler := ui.NewRecentActivityHandler(s.storage)
				executions.Get("/recent", ginHandler(recentActivityHandler.GetRecentActivityHandler))

				// Individual execution operations
				executions.Get("/:execution_id/details", ginHandler(uiExecutionsHandler.GetExecutionDetailsGlobalHandler))
				executions.Post("/:execution_id/webhook/retry", ginHandler(uiExecutionsHandler.RetryExecutionWebhookHandler))

				// Execution notes endpoints for UI
				executions.Post("/note", ginHandler(handlers.AddExecutionNoteHandler(s.storage)))
				executions.Get("/:execution_id/notes", ginHandler(handlers.GetExecutionNotesHandler(s.storage)))

				// DID and VC management endpoints for executions
				didHandler := ui.NewDIDHandler(s.storage, s.didService, s.vcService)
				executions.Get("/:execution_id/vc", ginHandler(didHandler.GetExecutionVCHandler))
				executions.Get("/:execution_id/vc-status", ginHandler(didHandler.GetExecutionVCStatusHandler))
				executions.Post("/:execution_id/verify-vc", ginHandler(didHandler.VerifyExecutionVCComprehensiveHandler))
			}

			// Workflows management group
			workflows := uiAPI.Group("/workflows")
			{
				workflows.Get("/:workflowId/dag", ginHandler(handlers.GetWorkflowDAGHandler(s.storage)))
				workflows.Delete("/:workflowId/cleanup", ginHandler(handlers.CleanupWorkflowHandler(s.storage)))
				didHandler := ui.NewDIDHandler(s.storage, s.didService, s.vcService)
				workflows.Post("/vc-status", ginHandler(didHandler.GetWorkflowVCStatusBatchHandler))
				workflows.Get("/:workflowId/vc-chain", ginHandler(didHandler.GetWorkflowVCChainHandler))
				workflows.Post("/:workflowId/verify-vc", ginHandler(didHandler.VerifyWorkflowVCComprehensiveHandler))

				// Workflow notes SSE streaming
				workflowNotesHandler := ui.NewExecutionHandler(s.storage, s.payloadStore, s.webhookDispatcher)
				workflows.Get("/:workflowId/notes/events", ginHandler(workflowNotesHandler.StreamWorkflowNodeNotesHandler))
			}

			// Reasoners management group
			reasoners := uiAPI.Group("/reasoners")
			{
				reasonersHandler := ui.NewReasonersHandler(s.storage)
				reasoners.Get("/all", ginHandler(reasonersHandler.GetAllReasonersHandler))
				reasoners.Get("/events", ginHandler(reasonersHandler.StreamReasonerEventsHandler))
				reasoners.Get("/:reasonerId/details", ginHandler(reasonersHandler.GetReasonerDetailsHandler))
				reasoners.Get("/:reasonerId/metrics", ginHandler(reasonersHandler.GetPerformanceMetricsHandler))
				reasoners.Get("/:reasonerId/executions", ginHandler(reasonersHandler.GetExecutionHistoryHandler))
				reasoners.Get("/:reasonerId/templates", ginHandler(reasonersHandler.GetExecutionTemplatesHandler))
				reasoners.Post("/:reasonerId/templates", ginHandler(reasonersHandler.SaveExecutionTemplateHandler))
			}

			// MCP system-wide endpoints
			mcp := uiAPI.Group("/mcp")
			{
				mcpHandler := ui.NewMCPHandler(s.uiService, s.agentClient)
				mcp.Get("/status", ginHandler(mcpHandler.GetMCPStatusHandler))
			}

			// Dashboard endpoints
			dashboard := uiAPI.Group("/dashboard")
			{
				dashboardHandler := ui.NewDashboardHandler(s.storage, s.agentService)
				dashboard.Get("/summary", ginHandler(dashboardHandler.GetDashboardSummaryHandler))
				dashboard.Get("/enhanced", ginHandler(dashboardHandler.GetEnhancedDashboardSummaryHandler))
			}

			// DID system-wide endpoints
			did := uiAPI.Group("/did")
			{
				didHandler := ui.NewDIDHandler(s.storage, s.didService, s.vcService)
				did.Get("/status", ginHandler(didHandler.GetDIDSystemStatusHandler))
				did.Get("/export/vcs", ginHandler(didHandler.ExportVCsHandler))
				did.Get("/:did/resolution-bundle", ginHandler(didHandler.GetDIDResolutionBundleHandler))
				did.Get("/:did/resolution-bundle/download", ginHandler(didHandler.DownloadDIDResolutionBundleHandler))
			}

			// VC system-wide endpoints
			vc := uiAPI.Group("/vc")
			{
				didHandler := ui.NewDIDHandler(s.storage, s.didService, s.vcService)
				vc.Get("/:vcId/download", ginHandler(didHandler.DownloadVCHandler))
				vc.Post("/verify", ginHandler(didHandler.VerifyVCHandler))
			}

			// Identity & Trust endpoints (DID Explorer and Credentials)
			identityHandler := ui.NewIdentityHandlers(s.storage)
			identityHandler.RegisterRoutes(uiAPI, ginHandler)
		}

		uiAPIV2 := s.App.Group("/v2/ui")
		{
			workflowRunsHandler := ui.NewWorkflowRunHandler(s.storage)
			uiAPIV2.Get("/workflow-runs", ginHandler(workflowRunsHandler.ListWorkflowRunsHandler))
			uiAPIV2.Get("/workflow-runs/:run_id", ginHandler(workflowRunsHandler.GetWorkflowRunDetailHandler))
		}
	}

	// Agent API routes
	agentAPI := s.App.Group("/v1")
	{
		// Health check endpoint for container orchestration
		agentAPI.Get("/health", ginHandler(s.healthCheckHandler))

		// Discovery endpoints
		discovery := agentAPI.Group("/discovery")
		{
			discovery.Get("/capabilities", ginHandler(handlers.DiscoveryCapabilitiesHandler(s.storage)))
		}

		// Node management endpoints
		agentAPI.Post("/nodes/register", ginHandler(handlers.RegisterNodeHandler(s.storage, s.uiService, s.didService, s.presenceManager)))
		agentAPI.Post("/nodes", ginHandler(handlers.RegisterNodeHandler(s.storage, s.uiService, s.didService, s.presenceManager)))
		agentAPI.Post("/nodes/register-serverless", ginHandler(handlers.RegisterServerlessAgentHandler(s.storage, s.uiService, s.didService, s.presenceManager)))
		agentAPI.Get("/nodes", ginHandler(handlers.ListNodesHandler(s.storage)))
		agentAPI.Get("/nodes/:node_id", ginHandler(handlers.GetNodeHandler(s.storage)))
		agentAPI.Post("/nodes/:node_id/heartbeat", ginHandler(handlers.HeartbeatHandler(s.storage, s.uiService, s.healthMonitor, s.statusManager, s.presenceManager)))
		agentAPI.Delete("/nodes/:node_id/monitoring", ginHandler(s.unregisterAgentFromMonitoring))

		// New unified status API endpoints
		agentAPI.Get("/nodes/:node_id/status", ginHandler(handlers.GetNodeStatusHandler(s.statusManager)))
		agentAPI.Post("/nodes/:node_id/status/refresh", ginHandler(handlers.RefreshNodeStatusHandler(s.statusManager)))
		agentAPI.Post("/nodes/status/bulk", ginHandler(handlers.BulkNodeStatusHandler(s.statusManager, s.storage)))
		agentAPI.Post("/nodes/status/refresh", ginHandler(handlers.RefreshAllNodeStatusHandler(s.statusManager, s.storage)))

		// Enhanced lifecycle management endpoints
		agentAPI.Post("/nodes/:node_id/start", ginHandler(handlers.StartNodeHandler(s.statusManager, s.storage)))
		agentAPI.Post("/nodes/:node_id/stop", ginHandler(handlers.StopNodeHandler(s.statusManager, s.storage)))
		agentAPI.Post("/nodes/:node_id/lifecycle/status", ginHandler(handlers.UpdateLifecycleStatusHandler(s.storage, s.uiService, s.statusManager)))
		agentAPI.Patch("/nodes/:node_id/status", ginHandler(handlers.NodeStatusLeaseHandler(s.storage, s.statusManager, s.presenceManager, handlers.DefaultLeaseTTL)))
		agentAPI.Post("/nodes/:node_id/actions/ack", ginHandler(handlers.NodeActionAckHandler(s.storage, s.presenceManager, handlers.DefaultLeaseTTL)))
		agentAPI.Post("/nodes/:node_id/shutdown", ginHandler(handlers.NodeShutdownHandler(s.storage, s.statusManager, s.presenceManager)))
		agentAPI.Post("/actions/claim", ginHandler(handlers.ClaimActionsHandler(s.storage, s.presenceManager, handlers.DefaultLeaseTTL)))

		// TODO: Add other node routes (DeleteNode)

		// Reasoner execution endpoints (legacy)
		agentAPI.Post("/reasoners/:reasoner_id", ginHandler(handlers.ExecuteReasonerHandler(s.storage)))

		// Skill execution endpoints (legacy)
		agentAPI.Post("/skills/:skill_id", ginHandler(handlers.ExecuteSkillHandler(s.storage)))

		// Unified execution endpoints (path-based)
		agentAPI.Post("/execute/:target", ginHandler(handlers.ExecuteHandler(s.storage, s.payloadStore, s.webhookDispatcher, s.config.HanzoAgents.ExecutionQueue.AgentCallTimeout)))
		agentAPI.Post("/execute/async/:target", ginHandler(handlers.ExecuteAsyncHandler(s.storage, s.payloadStore, s.webhookDispatcher, s.config.HanzoAgents.ExecutionQueue.AgentCallTimeout)))
		agentAPI.Get("/executions/:execution_id", ginHandler(handlers.GetExecutionStatusHandler(s.storage)))
		agentAPI.Post("/executions/batch-status", ginHandler(handlers.BatchExecutionStatusHandler(s.storage)))
		agentAPI.Post("/executions/:execution_id/status", ginHandler(handlers.UpdateExecutionStatusHandler(s.storage, s.payloadStore, s.webhookDispatcher, s.config.HanzoAgents.ExecutionQueue.AgentCallTimeout)))

		// Execution notes endpoints for app.note() feature
		agentAPI.Post("/executions/note", ginHandler(handlers.AddExecutionNoteHandler(s.storage)))
		agentAPI.Get("/executions/:execution_id/notes", ginHandler(handlers.GetExecutionNotesHandler(s.storage)))
		agentAPI.Post("/workflow/executions/events", ginHandler(handlers.WorkflowExecutionEventHandler(s.storage)))

		// Workflow endpoints will be reintroduced once the simplified execution pipeline lands.

		// Memory endpoints
		agentAPI.Post("/memory/set", ginHandler(handlers.SetMemoryHandler(s.storage)))
		agentAPI.Post("/memory/get", ginHandler(handlers.GetMemoryHandler(s.storage)))
		agentAPI.Post("/memory/delete", ginHandler(handlers.DeleteMemoryHandler(s.storage)))
		agentAPI.Get("/memory/list", ginHandler(handlers.ListMemoryHandler(s.storage)))

		// Vector Memory endpoints (RESTful)
		agentAPI.Post("/memory/vector", ginHandler(handlers.SetVectorHandler(s.storage)))
		agentAPI.Get("/memory/vector/:key", ginHandler(handlers.GetVectorHandler(s.storage)))
		agentAPI.Post("/memory/vector/search", ginHandler(handlers.SimilaritySearchHandler(s.storage)))
		agentAPI.Delete("/memory/vector/:key", ginHandler(handlers.DeleteVectorHandler(s.storage)))

		// Legacy Vector Memory endpoints (for backward compatibility)
		agentAPI.Post("/memory/vector/set", ginHandler(handlers.SetVectorHandler(s.storage)))
		agentAPI.Post("/memory/vector/delete", ginHandler(handlers.DeleteVectorHandler(s.storage)))
		agentAPI.Delete("/memory/vector/namespace", ginHandler(handlers.DeleteNamespaceVectorsHandler(s.storage)))

		// Memory events endpoints. The WebSocket upgrade rides the same seam:
		// the net/http adaptor supports connection hijack.
		memoryEventsHandler := handlers.NewMemoryEventsHandler(s.storage)
		agentAPI.Get("/memory/events/ws", ginHandler(memoryEventsHandler.WebSocketHandler))
		agentAPI.Get("/memory/events/sse", ginHandler(memoryEventsHandler.SSEHandler))
		agentAPI.Get("/memory/events/history", ginHandler(handlers.GetEventHistoryHandler(s.storage)))

		// DID/VC endpoints - use service-backed handlers if DID is enabled
		logger.Logger.Debug().
			Bool("did_enabled", s.config.Features.DID.Enabled).
			Bool("did_service_available", s.didService != nil).
			Bool("vc_service_available", s.vcService != nil).
			Msg("DID Route Registration Check")

		if s.config.Features.DID.Enabled && s.didService != nil && s.vcService != nil {
			logger.Logger.Debug().Msg("Registering DID routes - all conditions met")
			// Create DID handlers instance with services
			didHandlers := handlers.NewDIDHandlers(s.didService, s.vcService)

			// Register service-backed DID routes
			didHandlers.RegisterRoutes(agentAPI, ginHandler)

			// Add af server DID endpoint
			agentAPI.Get("/did/hanzo-agents-server", ginHandler(func(c *gin.Context) {
				// Get af server ID dynamically
				hanzoAgentsServerID, err := s.didService.GetHanzoAgentsServerID()
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{
						"error":   "Failed to get af server ID",
						"details": fmt.Sprintf("HanzoAgents server ID error: %v", err),
					})
					return
				}

				// Get the actual af server DID from the registry
				registry, err := s.didService.GetRegistry(hanzoAgentsServerID)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{
						"error":   "Failed to get af server DID",
						"details": fmt.Sprintf("Registry error: %v", err),
					})
					return
				}

				if registry == nil {
					c.JSON(http.StatusNotFound, gin.H{
						"error":   "HanzoAgents server DID not found",
						"details": "No DID registry exists for af server 'default'. The DID system may not be properly initialized.",
					})
					return
				}

				if registry.RootDID == "" {
					c.JSON(http.StatusInternalServerError, gin.H{
						"error":   "HanzoAgents server DID is empty",
						"details": "Registry exists but root DID is empty. The DID system may be corrupted.",
					})
					return
				}

				c.JSON(http.StatusOK, gin.H{
					"hanzo_agents_server_id":  "default",
					"hanzo_agents_server_did": registry.RootDID,
					"message":                 "HanzoAgents server DID retrieved successfully",
				})
			}))
		} else {
			logger.Logger.Warn().
				Bool("did_enabled", s.config.Features.DID.Enabled).
				Bool("did_service_available", s.didService != nil).
				Bool("vc_service_available", s.vcService != nil).
				Msg("DID routes NOT registered - conditions not met")
		}
		// Note: Removed unused/unimplemented DID endpoint placeholders for system simplification

		// Settings API routes (observability webhook configuration)
		settings := agentAPI.Group("/settings")
		{
			obsHandler := ui.NewObservabilityWebhookHandler(s.storage, s.observabilityForwarder)
			settings.Get("/observability-webhook", ginHandler(obsHandler.GetWebhookHandler))
			settings.Post("/observability-webhook", ginHandler(obsHandler.SetWebhookHandler))
			settings.Delete("/observability-webhook", ginHandler(obsHandler.DeleteWebhookHandler))
			settings.Get("/observability-webhook/status", ginHandler(obsHandler.GetStatusHandler))
			settings.Post("/observability-webhook/redrive", ginHandler(obsHandler.RedriveHandler))
			settings.Get("/observability-webhook/dlq", ginHandler(obsHandler.GetDeadLetterQueueHandler))
			settings.Delete("/observability-webhook/dlq", ginHandler(obsHandler.ClearDeadLetterQueueHandler))
			settings.Post("/observability-webhook/presets/console", ginHandler(obsHandler.SetConsolePresetHandler))
		}
	}

	// SPA fallback - serve index.html for all /ui/* routes that don't match static files
	// Only add this if we're NOT using embedded UI (since embedded UI handles its own NoRoute)
	if s.config.UI.Enabled && (s.config.UI.Mode != "embedded" || !client.IsUIEmbedded()) {
		// zip resolves routes by specificity, so this catch-all is reached
		// only when nothing else matched — gin's NoRoute semantics. "+"
		// rather than "*": it requires at least one character, which leaves
		// the root route (registered just above, in the same condition) to
		// win "/". A "/*" catch-all would shadow it.
		s.App.All("/+", ginHandler(func(c *gin.Context) {
			// Only handle /ui/* paths
			if strings.HasPrefix(c.Request.URL.Path, "/ui/") {
				// Check if it's a static asset by looking for common web asset file extensions
				// This prevents reasoner IDs with dots (like "deepresearchagent.meta_research_methodology_reasoner")
				// from being treated as static assets
				path := strings.ToLower(c.Request.URL.Path)
				isStaticAsset := strings.HasSuffix(path, ".js") ||
					strings.HasSuffix(path, ".css") ||
					strings.HasSuffix(path, ".html") ||
					strings.HasSuffix(path, ".ico") ||
					strings.HasSuffix(path, ".png") ||
					strings.HasSuffix(path, ".jpg") ||
					strings.HasSuffix(path, ".jpeg") ||
					strings.HasSuffix(path, ".gif") ||
					strings.HasSuffix(path, ".svg") ||
					strings.HasSuffix(path, ".woff") ||
					strings.HasSuffix(path, ".woff2") ||
					strings.HasSuffix(path, ".ttf") ||
					strings.HasSuffix(path, ".eot") ||
					strings.HasSuffix(path, ".map") ||
					strings.HasSuffix(path, ".json") ||
					strings.HasSuffix(path, ".xml") ||
					strings.HasSuffix(path, ".txt")

				if isStaticAsset {
					// Let it 404 for missing static assets
					c.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
					return
				}

				// For SPA routes (including reasoner detail pages), serve index.html from filesystem
				distPath := s.config.UI.DistPath
				if distPath == "" {
					// Get the executable path and find UI dist relative to it
					execPath, err := os.Executable()
					if err != nil {
						distPath = filepath.Join("apps", "platform", "hanzo-agents", "web", "client", "dist")
						if _, statErr := os.Stat(distPath); os.IsNotExist(statErr) {
							distPath = filepath.Join("web", "client", "dist")
						}
					} else {
						execDir := filepath.Dir(execPath)
						// Look for web/client/dist relative to the executable directory
						distPath = filepath.Join(execDir, "web", "client", "dist")

						// If that doesn't exist, try going up one level (if binary is in apps/platform/hanzo-agents/)
						if _, err := os.Stat(distPath); os.IsNotExist(err) {
							distPath = filepath.Join(filepath.Dir(execDir), "apps", "platform", "hanzo-agents", "web", "client", "dist")
						}

						// Final fallback to current working directory
						if _, err := os.Stat(distPath); os.IsNotExist(err) {
							altPath := filepath.Join("apps", "platform", "hanzo-agents", "web", "client", "dist")
							if _, altErr := os.Stat(altPath); altErr == nil {
								distPath = altPath
							} else {
								distPath = filepath.Join("web", "client", "dist")
							}
						}
					}
				}
				c.File(filepath.Join(distPath, "index.html"))
			} else {
				// For non-UI paths, return 404
				c.JSON(http.StatusNotFound, gin.H{"error": "endpoint not found"})
			}
		}))
	}
}

// generateHanzoAgentsServerID creates a deterministic af server ID based on the hanzo-agents home directory.
// This ensures each hanzo-agents instance has a unique ID while being deterministic for the same installation.
func generateHanzoAgentsServerID(hanzoAgentsHome string) string {
	// Use the absolute path of hanzo-agents home to generate a deterministic ID
	absPath, err := filepath.Abs(hanzoAgentsHome)
	if err != nil {
		// Fallback to the original path if absolute path fails
		absPath = hanzoAgentsHome
	}

	// Create a hash of the hanzo-agents home path to generate a unique but deterministic ID
	hash := sha256.Sum256([]byte(absPath))

	// Use first 16 characters of the hex hash as the af server ID
	// This provides uniqueness while keeping the ID manageable
	hanzoAgentsServerID := hex.EncodeToString(hash[:])[:16]

	return hanzoAgentsServerID
}
