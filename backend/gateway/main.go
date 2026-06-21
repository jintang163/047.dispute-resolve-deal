package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dispute-resolve/common/bootstrap"
	"github.com/dispute-resolve/common/config"
	"github.com/dispute-resolve/common/database"
	"github.com/dispute-resolve/common/logger"
	"github.com/dispute-resolve/common/mq"
	"github.com/dispute-resolve/common/model"
	"github.com/dispute-resolve/gateway/cron"
	"github.com/dispute-resolve/gateway/router"
	"github.com/dispute-resolve/gateway/rpc"
	"github.com/dispute-resolve/gateway/service"
	"github.com/dispute-resolve/gateway/service/impl"

	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/cloudwego/hertz/pkg/network/netpoll"
	hertzzap "github.com/hertz-contrib/logger/zap"
)

var timeout = 30 * time.Second

func main() {
	initResult := bootstrap.InitServiceWithOptions(bootstrap.InitOptions{
		ConfigPath:     "./conf/config.yaml",
		ServiceName:    "gateway",
		EnableRedis:    true,
		EnableFlowable: true,
		EnableAI:       true,
		EnableMilvus:   true,
		EnableMinIO:    true,
		EnableCourt:    true,
		EnableTRTC:     true,
		LogLevel:       "info",
	})
	defer initResult.Stop()

	cfg := config.GlobalConfig

	if err := database.AutoMigrate(
		&model.TransferTemplate{},
		&model.DisputeTransfer{},
		&model.DisputeTransferUrge{},
		&model.DisputeEvidence{},
	); err != nil {
		logger.Error("Auto migrate tables failed", logger.Error(err))
	}
	logger.Info("Tables auto migrated")

	log := logger.GetLogger()
	hlog.SetLogger(hertzzap.NewLogger(hertzzap.WithLogger(log)))
	hlog.SetLevel(hlog.LevelDebug)

	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	h := server.New(
		server.WithHostPorts(addr),
		server.WithTransport(netpoll.NewTransporter),
		server.WithReadTimeout(60000),
		server.WithWriteTimeout(60000),
		server.WithIdleTimeout(120000),
		server.WithMaxRequestBodySize(50*1024*1024),
	)

	userService := impl.NewUserService()
	disputeService := impl.NewDisputeService()
	mediationService := impl.NewMediationService()
	approvalService := impl.NewApprovalService()
	aiService := impl.NewAIService()
	statsService := impl.NewStatsService()
	notificationService := impl.NewNotificationService()
	performanceService := impl.NewPerformanceService()
	dispatchService := impl.NewDispatchService()
	videoService := impl.NewVideoService()
	esignService := impl.NewESignService()
	judicialConfirmationService := impl.NewJudicialConfirmationService()

	service.InitServices(
		userService,
		disputeService,
		mediationService,
		approvalService,
		aiService,
		statsService,
		notificationService,
		performanceService,
		dispatchService,
		videoService,
		esignService,
		judicialConfirmationService,
	)

	rpc.InitRPCClients()

	mq.StartConsumers()

	service.InitCallbackService()
	service.InitSatisfactionService()
	service.InitTimeoutUrgeService()

	cron.StartCronTasks()

	router.RegisterRoutes(h)

	go func() {
		log.Info("Gateway server starting", hlog.String("addr", addr))
		if err := h.Run(); err != nil {
			log.Fatal("Gateway server run failed", hlog.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Gateway server shutting down")

	cron.StopCronTasks()

	mq.ShutdownAllConsumers()

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	h.Shutdown(ctx)
	log.Info("Gateway server exited")
}
