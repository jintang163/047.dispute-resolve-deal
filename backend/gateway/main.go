package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dispute-resolve/common/config"
	"github.com/dispute-resolve/common/database"
	"github.com/dispute-resolve/common/logger"
	"github.com/dispute-resolve/common/utils"
	"github.com/dispute-resolve/gateway/router"
	"github.com/dispute-resolve/gateway/rpc"
	"github.com/dispute-resolve/gateway/service"
	"github.com/dispute-resolve/gateway/service/impl"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/cloudwego/hertz/pkg/network/netpoll"
	hertzlog "github.com/cloudwego/hertz/pkg/common/hlog"
	hertzzap "github.com/hertz-contrib/logger/zap"
)

var timeout = 30 * time.Second

func main() {
	if err := config.LoadConfig("./conf/config.yaml"); err != nil {
		panic(fmt.Sprintf("Load config failed: %v", err))
	}
	cfg := config.GlobalConfig

	logger.InitLogger(cfg.Server.Mode, cfg.Server.Name)
	utils.InitIDGenerator(1)

	database.InitDB(&cfg.Database)

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
	)

	rpc.InitRPCClients()

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
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	h.Shutdown(ctx)
	log.Info("Gateway server exited")
}
