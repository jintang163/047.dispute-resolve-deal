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
		public.POST("/video/record-callback", handler.RecordCallback)
		public.POST("/esign/fadada-callback", handler.FaDaDaCallback)
		public.GET("/blockchain/verify/:certNo", handler.PublicVerifyEvidence)
		public.POST("/callback/aliyun-voice", handler.AliyunVoiceCallback)
		public.POST("/idcard/query", handler.QueryPopulationByIDCard)
	}

	userAuth := api.Group("", middleware.JWTAuthMiddleware())
	{
		userAuth.POST("/idcard/query", handler.QueryPopulationByIDCard)

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
			dispute.GET("/:id/escalations", handler.GetCaseEscalationListHandler)
			dispute.GET("/:id/transfers", handler.GetCaseTransferList)

			dispute.GET("/mediators", handler.GetMediatorList)
			dispute.GET("/mediators/:id/load", handler.GetMediatorLoad)

			evidence := dispute.Group("/:id/evidence")
			{
				evidence.POST("", handler.UploadEvidence)
				evidence.GET("", handler.GetEvidenceList)
				evidence.DELETE("/:evidenceId", handler.DeleteEvidence)
				evidence.POST("/batch-delete", handler.BatchDeleteEvidence)
				evidence.PUT("/:evidenceId/remark", handler.UpdateEvidenceRemark)
				evidence.PUT("/:evidenceId/category", handler.UpdateEvidenceCategory)
			}

			mediation := dispute.Group("/:id/mediation")
			{
				mediation.GET("", handler.GetMediationRecords)
				mediation.POST("", handler.CreateMediationRecord)
				mediation.PUT("/:recordId", handler.UpdateMediationRecord)
				mediation.GET("/:recordId/ai-summary", handler.GetAISummary)
				mediation.POST("/protocol/generate", handler.GenerateMediationProtocol)
				mediation.GET("/protocols", handler.GetMediationProtocolList)
				mediation.POST("/protocol/:protocolId/adopt", handler.AdoptMediationProtocol)
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
				video.POST("/trtc/usersig", handler.GetTRTCUserSig)
				video.POST("/record/start", handler.StartVideoRecord)
				video.POST("/record/stop", handler.StopVideoRecord)
				video.GET("/:roomId/segments", handler.GetVideoRecordSegments)
				video.POST("/screen-share", handler.UpdateScreenShare)
				video.POST("/virtual-bg", handler.UpdateVirtualBackground)
				video.POST("/beauty", handler.UpdateBeautyFilter)
				video.POST("/minutes/generate", handler.GenerateVideoMinutes)
				video.GET("/:roomId/minutes", handler.GetVideoMeetingMinutes)
				video.POST("/minutes/:minutesId/approve", handler.ApproveVideoMinutes)
			}

			videoQueue := dispute.Group("/video-queue")
			{
				videoQueue.POST("/enqueue", handler.EnqueueVideoMediation)
				videoQueue.GET("/position", handler.GetVideoQueuePosition)
				videoQueue.GET("/list", handler.GetVideoQueueList)
				videoQueue.POST("/leave", handler.LeaveVideoQueue)
				videoQueue.GET("/status", handler.GetVideoQueueStatus)
			}

			esign := dispute.Group("/:id/esign")
			{
				esign.POST("", handler.CreateEsignFlow)
				esign.GET("", handler.GetEsignList)
				esign.GET("/:flowId", handler.GetEsignDetail)
				esign.POST("/:flowId/sign", handler.SignDocument)
				esign.POST("/:flowId/revoke", handler.RevokeEsignFlow)
				esign.POST("/:flowId/send-code", handler.SendEsignVerifyCode)
				esign.GET("/:flowId/progress", handler.GetEsignProgress)
			}

			blockchainGroup := dispute.Group("/:id/blockchain")
			{
				blockchainGroup.POST("/store", handler.StoreEvidenceToBlockchain)
				blockchainGroup.GET("/certs", handler.GetBlockchainCertList)
				blockchainGroup.GET("/cert/:certNo", handler.GetBlockchainCertDetail)
				blockchainGroup.GET("/verify", handler.VerifyBlockchainEvidence)
				blockchainGroup.GET("/cert/:certNo/download", handler.DownloadBlockchainCert)
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
			stats.GET("/heatmap/timeline", handler.GetHeatmapTimeline)
			stats.GET("/heatmap/top-communities", handler.GetTopCommunities)
			stats.GET("/heatmap/drilldown", handler.GetHeatmapDrilldown)
			stats.GET("/heatmap/amap-config", handler.GetAmapConfig)
			stats.GET("/organization", handler.GetOrganizationStats)
			stats.GET("/yearly-comparison", handler.GetYearlyComparison)
			stats.GET("/mediator-ranking", handler.GetMediatorRanking)
			stats.GET("/keywords/aggregation", handler.GetKeywordStats)
			stats.POST("/refresh-cache", handler.RefreshStatsCache)
		}

		performance := userAuth.Group("/performance")
		{
			performance.GET("", handler.GetPerformanceScoreList)
			performance.GET("/my", handler.GetMyPerformance)
			performance.GET("/:userId", handler.GetPerformanceDetail)
			performance.POST("/calculate", handler.CalculatePerformanceScore)
			performance.POST("/batch-calculate", handler.BatchCalculatePerformanceScore)
			performance.GET("/trend", handler.GetPerformanceTrend)
			performance.GET("/ranking", handler.GetPerformanceRanking)
			performance.GET("/indicator-config", handler.GetPerformanceIndicatorConfig)
			performance.PUT("/indicator-config", handler.UpdatePerformanceIndicatorConfig)
			performance.GET("/dashboard", handler.GetPerformanceDashboard)
			performance.GET("/month-comparison", handler.GetPerformanceMonthComparison)
			performance.GET("/export", handler.ExportPerformanceExcel)

			interview := performance.Group("/interview")
			{
				interview.GET("", handler.GetPerformanceInterviewList)
				interview.POST("", handler.CreatePerformanceInterview)
				interview.GET("/:id", handler.GetPerformanceInterviewDetail)
				interview.POST("/:id/confirm", handler.ConfirmPerformanceInterview)
			}
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

		callback := userAuth.Group("/callback")
		{
			callback.GET("", handler.GetCallbackList)
			callback.GET("/:id", handler.GetCallbackDetail)
			callback.POST("", handler.CreateCallback)
			callback.POST("/:id/initiate", handler.InitiateCallback)
			callback.POST("/:id/retry", handler.RetryCallback)
			callback.POST("/:id/cancel", handler.CancelCallback)
			callback.POST("/:id/refresh", handler.RefreshCallbackResult)
			callback.POST("/:id/archive-recording", handler.DownloadAndArchiveRecording)
			callback.GET("/case/:caseId", handler.GetCallbacksByCase)
		}

		satisfaction := userAuth.Group("/satisfaction")
		{
			satisfaction.POST("/analyze/:caseId", handler.AnalyzeSatisfactionHandler)
			satisfaction.GET("/stats", handler.GetSatisfactionSentimentStatsHandler)
		}

		caseLibrary := userAuth.Group("/case-library")
		{
			caseLibrary.GET("", handler.GetCaseLibraryList)
			caseLibrary.GET("/search", handler.SearchSimilarCases)
			caseLibrary.GET("/archives", handler.GetCaseLibraryArchiveList)
			caseLibrary.GET("/:id", handler.GetCaseLibraryDetail)
			caseLibrary.POST("", handler.CreateCaseLibrary)
			caseLibrary.PUT("/:id", handler.UpdateCaseLibrary)
			caseLibrary.DELETE("/:id", handler.DeleteCaseLibrary)
			caseLibrary.POST("/:id/score", handler.ScoreCaseLibrary)
			caseLibrary.POST("/:id/quote", handler.QuoteCaseLibrary)
			caseLibrary.GET("/quotes", handler.GetCaseQuoteList)
			caseLibrary.POST("/:id/archive", handler.ArchiveCaseLibrary)
			caseLibrary.POST("/:id/restore", handler.RestoreCaseLibrary)
			caseLibrary.POST("/:id/vectorize", handler.VectorizeCaseLibrary)
			caseLibrary.POST("/vectorize-all", handler.VectorizeAllCaseLibrary)
		}

		improvement := userAuth.Group("/improvement")
		{
			improvement.GET("", handler.GetImprovementOrderListHandler)
			improvement.GET("/:id", handler.GetImprovementOrderDetailHandler)
			improvement.POST("/:id/rectify", handler.SubmitRectificationHandler)
			improvement.POST("/:id/review", handler.ReviewRectificationHandler)
			improvement.POST("/:id/close", handler.CloseImprovementOrderHandler)
		}

		escalation := userAuth.Group("/escalation")
		{
			escalation.GET("", handler.GetEscalationListHandler)
			escalation.GET("/:id", handler.GetEscalationDetailHandler)
			escalation.POST("/:id/handle", handler.HandleEscalationHandler)
			escalation.POST("/:id/close", handler.CloseEscalationHandler)
		}

		urge := userAuth.Group("/urge")
		{
			urge.GET("/case/:caseId", handler.GetCaseUrgeListHandler)
		}

		transfer := userAuth.Group("/transfer")
		{
			transfer.GET("", handler.GetTransferList)
			transfer.GET("/:id", handler.GetTransferDetail)
			transfer.POST("", handler.CreateTransfer)
			transfer.POST("/:id/receive", handler.ReceiveTransfer)
			transfer.POST("/:id/reject", handler.RejectTransfer)
			transfer.POST("/:id/process", handler.StartProcessTransfer)
			transfer.POST("/:id/complete", handler.CompleteTransfer)
			transfer.POST("/:id/urge", handler.UrgeTransfer)
			transfer.POST("/:id/cancel", handler.CancelTransfer)
			transfer.GET("/:transferId/urges", handler.GetTransferUrgeList)
			transfer.GET("/templates/available", handler.GetAvailableTransferTemplates)

			transfer.GET("/stats/dept", handler.GetTransferDeptStats)
			transfer.GET("/stats/duration-ranking", handler.GetTransferDurationRanking)
			transfer.GET("/stats/trend", handler.GetTransferTrendStats)
		}

		transferTemplate := userAuth.Group("/transfer-template", middleware.AdminRequiredMiddleware())
		{
			transferTemplate.GET("", handler.GetTransferTemplateList)
			transferTemplate.GET("/:id", handler.GetTransferTemplateDetail)
			transferTemplate.POST("", handler.CreateTransferTemplate)
			transferTemplate.PUT("/:id", handler.UpdateTransferTemplate)
			transferTemplate.DELETE("/:id", handler.DeleteTransferTemplate)
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
				mediator.GET("/:id/load", handler.GetMediatorLoad)
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
				ai.POST("/keywords/extract", handler.ExtractKeywords)
				ai.GET("/keywords/dict", handler.GetKeywordDict)
				ai.GET("/keywords/hot", handler.GetHotKeywords)
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
