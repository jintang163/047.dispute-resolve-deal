package router

import (
	"context"
	"time"

	"github.com/dispute-resolve/common/auth"
	"github.com/dispute-resolve/gateway/handler"
	"github.com/dispute-resolve/gateway/middleware"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/middlewares/server/cors"
	"github.com/cloudwego/hertz/pkg/app/middlewares/server/recovery"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
)

var timeout = 30 * time.Second

func RegisterRoutes(h *app.Hertz) {
	h.Use(
		cors.New(cors.Config{
			AllowAllOrigins:  true,
			AllowMethods:     []string{consts.MethodGet, consts.MethodPost, consts.MethodPut, consts.MethodDelete, consts.MethodOptions},
			AllowHeaders:     []string{"Origin", "Content-Type", "Content-Length", "Accept-Encoding", "X-CSRF-Token", "Authorization"},
			ExposeHeaders:    []string{"Content-Length"},
			AllowCredentials: true,
			MaxAge:           12 * time.Hour,
		}),
		recovery.New(recovery.WithRecoveryHandler(func(c context.Context, ctx *app.RequestContext, err interface{}, stack []byte) {
			hlog.SystemLogger().CtxErrorf(c, "[Recovery] err=%v\nstack=%s", err, stack)
			ctx.JSON(consts.StatusInternalServerError, map[string]interface{}{
				"code":    500,
				"message": "Internal Server Error",
			})
		})),
		middleware.RequestIDMiddleware(),
		middleware.LoggerMiddleware(),
		middleware.RateLimitMiddleware(),
	)

	api := h.Group("/api/v1")

	public := api.Group("/public")
	{
		public.GET("/health", handler.HealthCheck)
		public.POST("/login", handler.Login)
		public.POST("/kiosk/login", handler.KioskLogin)
		public.POST("/miniapp/login", handler.MiniAppLogin)
		public.GET("/dispute/types", handler.GetDisputeTypes)
		public.POST("/dispute/kiosk/create", handler.KioskCreateDispute)
		public.POST("/dispute/miniapp/create", handler.MiniAppCreateDispute)
		public.GET("/dispute/progress", handler.GetDisputeProgress)
		public.POST("/ai/consult", handler.AIConsult)
	}

	userAuth := api.Group("", middleware.JWTAuthMiddleware())
	{
		user := userAuth.Group("/user")
		{
			user.GET("/info", handler.GetUserInfo)
			user.PUT("/password", handler.ChangePassword)
		}

		dispute := userAuth.Group("/dispute")
		{
			dispute.GET("", handler.GetDisputeList)
			dispute.GET("/:id", handler.GetDisputeDetail)
			dispute.POST("", handler.CreateDispute)
			dispute.POST("/:id/assign", handler.AssignDispute)
			dispute.POST("/:id/urge", handler.UrgeDispute)
			dispute.POST("/:id/status", handler.UpdateDisputeStatus)
			dispute.GET("/:id/history", handler.GetDisputeHistory)

			evidence := dispute.Group("/:id/evidence")
			{
				evidence.POST("", handler.UploadEvidence)
				evidence.GET("", handler.GetEvidenceList)
				evidence.DELETE("/:evidenceId", handler.DeleteEvidence)
				evidence.POST("/batch-delete", handler.BatchDeleteEvidence)
				evidence.PUT("/:evidenceId/remark", handler.UpdateEvidenceRemark)
			}

			mediation := dispute.Group("/:id/mediation")
			{
				mediation.GET("", handler.GetMediationRecords)
				mediation.POST("", handler.CreateMediationRecord)
				mediation.PUT("/:recordId", handler.UpdateMediationRecord)
				mediation.GET("/:recordId/ai-summary", handler.GetAISummary)
			}

			approval := dispute.Group("/:id/approval")
			{
				approval.GET("", handler.GetApprovalProgress)
				approval.POST("/submit", handler.SubmitApproval)
				approval.POST("/approve", handler.ApproveApproval)
				approval.POST("/reject", handler.RejectApproval)
				approval.POST("/return", handler.ReturnApproval)
				approval.POST("/add-sign", handler.AddSignApproval)
				approval.POST("/transfer", handler.TransferApproval)
			}

			video := dispute.Group("/:id/video")
			{
				video.POST("/create", handler.CreateVideoRoom)
				video.GET("", handler.GetVideoRoomList)
				video.GET("/:roomId", handler.GetVideoRoomDetail)
				video.POST("/:roomId/join", handler.JoinVideoRoom)
				video.GET("/:roomId/token", handler.GetVideoRoomToken)
				video.POST("/:roomId/end", handler.EndVideoRoom)
				video.POST("/:roomId/cancel", handler.CancelVideoRoom)
			}

			esign := dispute.Group("/:id/esign")
			{
				esign.POST("", handler.CreateEsignFlow)
				esign.GET("", handler.GetEsignList)
				esign.GET("/:esignId", handler.GetEsignDetail)
				esign.POST("/:esignId/sign", handler.SignDocument)
				esign.POST("/:esignId/revoke", handler.RevokeEsignFlow)
				esign.POST("/:esignId/send-code", handler.SendEsignVerifyCode)
			}

			dispatch := dispute.Group("/dispatch")
			{
				dispatch.POST("/intelligent", handler.IntelligentDispatch)
				dispatch.POST("/batch", handler.BatchIntelligentDispatch)
				dispatch.GET("/config", handler.GetDispatchConfig)
				dispatch.PUT("/config", handler.UpdateDispatchConfig)
				dispatch.GET("/mediator-load", handler.GetMediatorLoadStats)
			}
		}

		judicial := userAuth.Group("/judicial")
		{
			judicial.GET("/list", handler.GetJudicialConfirmationList)
			judicial.GET("/:id", handler.GetJudicialConfirmationDetail)
			judicial.GET("/query", handler.QueryJudicialConfirmationByNo)
			judicial.POST("", handler.CreateJudicialConfirmation)
			judicial.POST("/:id/submit", handler.SubmitJudicialToCourt)
			judicial.POST("/:id/query-status", handler.QueryCourtStatus)
			judicial.POST("/:id/generate-doc", handler.GenerateConfirmationDocument)
			judicial.POST("/:id/seal", handler.SealConfirmationDocument)
			judicial.GET("/:id/logs", handler.GetConfirmationLogs)

			remind := judicial.Group("/:id/remind")
			{
				remind.POST("/performance", handler.SendPerformanceReminder)
				remind.POST("/expiration", handler.SendExpirationReminder)
			}
		}

		approval := userAuth.Group("/approval")
		{
			approval.GET("/todo", handler.GetApprovalTodoList)
			approval.GET("/done", handler.GetApprovalDoneList)
		}

		stats := userAuth.Group("/stats")
		{
			stats.GET("/dashboard", handler.GetDashboardStats)
			stats.GET("/heatmap", handler.GetHeatmapData)
			stats.GET("/organization", handler.GetOrganizationStats)
			stats.GET("/yearly-comparison", handler.GetYearlyComparison)
			stats.GET("/mediator-ranking", handler.GetMediatorRanking)
			stats.POST("/refresh-cache", handler.RefreshStatsCache)
		}

		performance := userAuth.Group("/performance")
		{
			performance.GET("", handler.GetPerformanceScoreList)
			performance.GET("/my", handler.GetMyPerformance)
			performance.GET("/:userId", handler.GetPerformanceDetail)
			performance.POST("/calculate", handler.CalculatePerformanceScore)
			performance.GET("/trend", handler.GetPerformanceTrend)
			performance.GET("/ranking", handler.GetPerformanceRanking)
			performance.GET("/indicator-config", handler.GetPerformanceIndicatorConfig)
		}

		notification := userAuth.Group("/notification")
		{
			notification.GET("", handler.GetMyNotifications)
			notification.GET("/unread-count", handler.GetUnreadCount)
			notification.GET("/templates", handler.GetNotificationTemplates)
			notification.GET("/:id", handler.GetNotificationDetail)
			notification.PUT("/:id/read", handler.MarkAsRead)
			notification.PUT("/read-all", handler.MarkAllAsRead)
			notification.POST("/send", handler.SendNotification)
			notification.DELETE("/:id", handler.DeleteNotification)
			notification.POST("/batch-delete", handler.BatchDeleteNotifications)
		}

		ws := userAuth.Group("/ws")
		{
			ws.GET("", handler.HandleWebSocket)
			ws.GET("/case/:caseId", handler.HandleCaseWebSocket)
			ws.GET("/online-users", handler.GetOnlineUsers)
			ws.POST("/broadcast", handler.BroadcastNotification)
		}

		system := userAuth.Group("/system", middleware.AdminRequiredMiddleware())
		{
			user := system.Group("/user")
			{
				user.GET("", handler.GetUserList)
				user.GET("/:id", handler.GetUserDetail)
				user.POST("", handler.CreateUser)
				user.PUT("/:id", handler.UpdateUser)
				user.DELETE("/:id", handler.DeleteUser)
				user.PUT("/:id/reset-password", handler.ResetPassword)
			}

			org := system.Group("/organization")
			{
				org.GET("/tree", handler.GetOrganizationTree)
				org.POST("", handler.CreateOrganization)
				org.PUT("/:id", handler.UpdateOrganization)
				org.DELETE("/:id", handler.DeleteOrganization)
			}

			mediator := system.Group("/mediator")
			{
				mediator.GET("", handler.GetMediatorList)
			}

			log := system.Group("/log")
			{
				log.GET("/operation", handler.GetOperationLogList)
			}

			ai := system.Group("/ai")
			{
				ai.GET("/law-articles", handler.GetLawArticles)
				ai.POST("/law-articles", handler.CreateLawArticle)
				ai.PUT("/law-articles/:id", handler.UpdateLawArticle)
				ai.DELETE("/law-articles/:id", handler.DeleteLawArticle)
				ai.POST("/law-articles/vectorize", handler.VectorizeLawArticles)
				ai.GET("/config", handler.GetAIConfig)
				ai.PUT("/config", handler.UpdateAIConfig)
			}

			court := system.Group("/court")
			{
				court.GET("/config/list", handler.GetCourtConfigList)
				court.GET("/config/:id", handler.GetCourtConfigDetail)
				court.POST("/config", handler.CreateCourtConfig)
				court.PUT("/config/:id", handler.UpdateCourtConfig)
				court.DELETE("/config/:id", handler.DeleteCourtConfig)
				court.GET("/options", handler.GetCourtOptions)
			}
		}
	}

	wsPublic := h.Group("/ws")
	{
		wsPublic.GET("/video/:roomId", handler.HandleVideoWebSocket)
	}
}
