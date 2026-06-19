package bootstrap

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/dispute-resolve/common/ai"
	"github.com/dispute-resolve/common/cache"
	"github.com/dispute-resolve/common/config"
	"github.com/dispute-resolve/common/court"
	"github.com/dispute-resolve/common/database"
	"github.com/dispute-resolve/common/logger"
	"github.com/dispute-resolve/common/trtc"
	"github.com/dispute-resolve/common/utils"
	"github.com/dispute-resolve/common/vector"
	"github.com/dispute-resolve/common/workflow"
)

type InitOptions struct {
	ConfigPath     string
	ServiceName    string
	EnableRedis    bool
	EnableFlowable bool
	EnableAI       bool
	EnableMilvus   bool
	EnableMinIO    bool
	EnableCourt    bool
	EnableTRTC     bool
	LogLevel       string
}

type InitResult struct {
	Config *config.Config
	Stop   func()
}

func InitService(serviceName string) *InitResult {
	return InitServiceWithOptions(InitOptions{
		ConfigPath:  "../../config/config.yaml",
		ServiceName: serviceName,
		EnableRedis: false,
		LogLevel:    "info",
	})
}

func InitServiceWithRedis(serviceName string) *InitResult {
	return InitServiceWithOptions(InitOptions{
		ConfigPath:  "../../config/config.yaml",
		ServiceName: serviceName,
		EnableRedis: true,
		LogLevel:    "info",
	})
}

func InitServiceWithOptions(opts InitOptions) *InitResult {
	if opts.ConfigPath == "" {
		opts.ConfigPath = "../../config/config.yaml"
	}
	if opts.LogLevel == "" {
		opts.LogLevel = "info"
	}

	if err := config.LoadConfig(opts.ConfigPath); err != nil {
		log.Fatalf("[%s] Load config failed: %v", opts.ServiceName, err)
	}
	log.Printf("[%s] Config loaded successfully", opts.ServiceName)

	logger.InitLogger(opts.LogLevel, opts.ServiceName)
	logger.Info("Logger initialized", logger.String("service", opts.ServiceName))

	database.InitDB(&config.GlobalConfig.Database)
	logger.Info("Database initialized")

	utils.InitIDGenerator(1)
	logger.Info("ID generator initialized")

	if opts.EnableRedis {
		cache.InitRedis(&config.GlobalConfig.Redis)
		logger.Info("Redis initialized")
	}

	if opts.EnableFlowable {
		workflow.InitFlowable()
		logger.Info("Flowable workflow initialized")
	}

	if opts.EnableAI {
		ai.InitDeepSeek()
		logger.Info("DeepSeek AI initialized")
	}

	if opts.EnableMilvus {
		if err := vector.InitMilvus(); err != nil {
			logger.Warn("Milvus initialization failed, will use fallback search", logger.Error(err))
		} else {
			logger.Info("Milvus vector store initialized")
		}
	}

	if opts.EnableMinIO {
		database.InitMinIO(&config.GlobalConfig.MinIO)
		logger.Info("MinIO client initialized")
	}

	if opts.EnableCourt {
		court.InitMicroCourt()
		logger.Info("MicroCourt client initialized")
	}

	if opts.EnableTRTC {
		trtc.InitTRTC()
		trtc.InitCloudRecordService()
		trtc.InitVideoQueueService()
		logger.Info("TRTC video mediation services initialized")
	}

	stop := setupGracefulShutdown(opts.ServiceName)

	return &InitResult{
		Config: config.GlobalConfig,
		Stop:   stop,
	}
}

func setupGracefulShutdown(serviceName string) func() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	stop := func() {
		logger.Info("Service stopping", logger.String("service", serviceName))
	}

	go func() {
		sig := <-sigChan
		logger.Info("Received signal",
			logger.String("service", serviceName),
			logger.String("signal", sig.String()))
		stop()
		os.Exit(0)
	}()

	return stop
}

func GetServicePort(serviceName string) int {
	cfg := config.GlobalConfig.ServicePorts
	switch serviceName {
	case "user-service":
		return cfg.User
	case "dispute-service":
		return cfg.Dispute
	case "workflow-service":
		return cfg.Workflow
	case "ai-service":
		return cfg.AI
	case "notification-service":
		return cfg.Notification
	default:
		return 8080
	}
}

func Shutdown(ctx context.Context) {
	logger.Info("Service shutdown complete")
}
