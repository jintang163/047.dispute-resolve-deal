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
		public.POST("/voice/recognize", handler.VoiceRecognize)
		public.POST("/receipt/qrcode", handler.GenerateReceiptQRCode)
		public.GET("/scan/:token", handler.ScanRedirect)
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

			mediationTemplate := dispute.Group("/mediation-template")
			{
				mediationTemplate.GET("", handler.GetMediationTemplateList)
				mediationTemplate.GET("/categories", handler.GetMediationTemplateCategories)
				mediationTemplate.GET("/recommend", handler.RecommendMediationTemplates)
				mediationTemplate.GET("/:id", handler.GetMediationTemplateDetail)
				mediationTemplate.POST("", handler.CreateMediationTemplate)
				mediationTemplate.PUT("/:id", handler.UpdateMediationTemplate)
				mediationTemplate.DELETE("/:id", handler.DeleteMediationTemplate)
				mediationTemplate.POST("/:id/apply", handler.ApplyMediationTemplate)
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

			dispute.POST("/export", handler.CreateCaseExport)
		}

		export := userAuth.Group("/export")
		{
			export.GET("/log", handler.GetExportList)
			export.GET("/log/:id", handler.GetExportDetail)
			export.GET("/log/:id/download", handler.DownloadExport)
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
			caseLibrary.POST("/search", handler.SearchSimilarCases)
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

		legalAid := userAuth.Group("/legal-aid")
		{
			legalAidOrg := legalAid.Group("/org")
			{
				legalAidOrg.GET("", handler.GetLegalAidOrgList)
				legalAidOrg.GET("/:id", handler.GetLegalAidOrgDetail)
				legalAidOrg.POST("", handler.CreateLegalAidOrg)
				legalAidOrg.PUT("/:id", handler.UpdateLegalAidOrg)
				legalAidOrg.DELETE("/:id", handler.DeleteLegalAidOrg)
			}

			legalAidLawyer := legalAid.Group("/lawyer")
			{
				legalAidLawyer.GET("", handler.GetLegalAidLawyerList)
				legalAidLawyer.GET("/:id", handler.GetLegalAidLawyerDetail)
			}

			legalAidApplication := legalAid.Group("/application")
			{
				legalAidApplication.GET("", handler.GetLegalAidApplyList)
				legalAidApplication.GET("/:id", handler.GetLegalAidApplyDetail)
				legalAidApplication.POST("", handler.ApplyLegalAid)
				legalAidApplication.POST("/:id/audit", handler.AuditLegalAidApply)
			}

			legalAidMaterial := legalAid.Group("/material")
			{
				legalAidMaterial.GET("", handler.GetLegalAidMaterialList)
				legalAidMaterial.POST("/upload", handler.UploadLegalAidMaterial)
				legalAidMaterial.DELETE("/:id", handler.DeleteLegalAidMaterial)
			}

			legalAidTransfer := legalAid.Group("/transfer")
			{
				legalAidTransfer.GET("", handler.GetLegalAidTransferList)
				legalAidTransfer.GET("/:id", handler.GetLegalAidTransferDetail)
				legalAidTransfer.POST("", handler.CreateLegalAidTransfer)
				legalAidTransfer.POST("/:id/accept", handler.AcceptLegalAidTransfer)
				legalAidTransfer.POST("/:id/close", handler.CloseLegalAidTransfer)
			}

			legalAid.POST("/recommend", handler.RecommendLegalAidOrgs)

			legalAidConsult := legalAid.Group("/consult")
			{
				legalAidConsult.GET("", handler.GetLegalAidConsultList)
				legalAidConsult.GET("/:id", handler.GetLegalAidConsultDetail)
				legalAidConsult.POST("", handler.CreateLegalAidConsult)
				legalAidConsult.POST("/:id/start", handler.StartLegalAidConsult)
				legalAidConsult.POST("/:id/end", handler.EndLegalAidConsult)
				legalAidConsult.POST("/:id/rate", handler.RateLegalAidConsult)
				legalAidConsult.GET("/:consultId/messages", handler.GetLegalAidConsultMessages)
				legalAidConsult.POST("/message", handler.SendLegalAidConsultMessage)
			}

			legalAid.GET("/case/:caseId/records", handler.GetCaseLegalAidRecords)
		}

		counseling := userAuth.Group("/counseling")
		{
			counselor := counseling.Group("/counselor")
			{
				counselor.GET("", handler.GetCounselorList)
				counselor.GET("/:id", handler.GetCounselorDetail)
				counselor.POST("", handler.CreateCounselor)
				counselor.PUT("/:id", handler.UpdateCounselor)
				counselor.DELETE("/:id", handler.DeleteCounselor)
				counselor.GET("/:id/available-slots", handler.GetCounselorAvailableSlots)
				counselor.GET("/:id/ratings", handler.GetCounselorRatingList)
				counselor.GET("/:id/stats", handler.GetCounselorStats)
			}

			counseling.POST("/counselor/recommend", handler.RecommendCounselors)

			appointment := counseling.Group("/appointment")
			{
				appointment.GET("", handler.GetAppointmentList)
				appointment.GET("/:id", handler.GetAppointmentDetail)
				appointment.POST("", handler.CreateAppointment)
				appointment.PUT("/:id", handler.UpdateAppointment)
				appointment.POST("/:id/cancel", handler.CancelAppointment)
			}

			counseling.POST("/rating", handler.CreateRating)

			schedule := counseling.Group("/schedule")
			{
				schedule.GET("", handler.GetScheduleList)
				schedule.POST("", handler.CreateSchedule)
				schedule.DELETE("/:id", handler.DeleteSchedule)
			}
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

		patrol := userAuth.Group("/patrol")
		{
			patrol.POST("/task", handler.CreatePatrolTask)
			patrol.GET("/task", handler.GetPatrolTaskList)
			patrol.GET("/task/:id", handler.GetPatrolTaskDetail)
			patrol.PUT("/task/:id", handler.UpdatePatrolTask)
			patrol.DELETE("/task/:id", handler.DeletePatrolTask)
			patrol.POST("/task/:id/cancel", handler.CancelPatrolTask)
			patrol.POST("/task/:id/start", handler.StartPatrolTask)
			patrol.POST("/task/:id/complete", handler.CompletePatrolTask)
			patrol.GET("/task/:taskId/points", handler.GetTaskPoints)

			patrol.GET("/my/tasks", handler.GetMemberTasks)

			patrol.POST("/route/plan", handler.PlanRoute)

			patrol.POST("/checkin", handler.Checkin)
			patrol.GET("/checkin/records", handler.GetCheckinRecords)
			patrol.GET("/checkin/statistics", handler.GetCheckinStatistics)

			patrol.POST("/visit", handler.CreateVisitRecord)
			patrol.GET("/visit", handler.GetVisitRecords)
			patrol.GET("/visit/:id", handler.GetVisitRecordDetail)
			patrol.PUT("/visit/:id", handler.UpdateVisitRecord)
			patrol.POST("/visit/:id/audit", handler.AuditVisitRecord)
			patrol.DELETE("/visit/:id", handler.DeleteVisitRecord)
			patrol.GET("/visit/statistics", handler.GetVisitStatistics)

			patrol.POST("/danger", handler.ReportDanger)
			patrol.GET("/danger", handler.GetDangerList)
			patrol.GET("/danger/:id", handler.GetDangerDetail)
			patrol.POST("/danger/:id/handle", handler.HandleDanger)
			patrol.GET("/danger/statistics", handler.GetDangerStatistics)

			patrol.GET("/member/me", handler.GetMemberMe)
			patrol.GET("/member", handler.GetMemberList)
			patrol.GET("/member/:id", handler.GetMemberDetail)
			patrol.POST("/member", handler.CreateMember)
			patrol.PUT("/member/:id", handler.UpdateMember)
			patrol.DELETE("/member/:id", handler.DeleteMember)
		}

		points := userAuth.Group("/points")
		{
			points.GET("/summary", handler.GetPointsSummary)
			points.POST("/add", handler.AddPoints)
			points.POST("/deduct", handler.DeductPoints)
			points.GET("/records", handler.GetPointsRecords)
			points.GET("/rules", handler.GetPointsRules)
			points.POST("/rules", handler.CreatePointsRule)
			points.PUT("/rules/:id", handler.UpdatePointsRule)
			points.DELETE("/rules/:id", handler.DeletePointsRule)
			points.POST("/exchange", handler.ExchangeGift)
			points.POST("/process-expired", handler.ProcessExpiredPoints)
		}

		gift := userAuth.Group("/gift")
		{
			gift.GET("", handler.GetGiftList)
			gift.GET("/:id", handler.GetGiftDetail)
			gift.POST("", handler.CreateGift)
			gift.PUT("/:id", handler.UpdateGift)
			gift.DELETE("/:id", handler.DeleteGift)

			gift.GET("/categories", handler.GetGiftCategories)
			gift.POST("/categories", handler.CreateGiftCategory)
			gift.PUT("/categories/:id", handler.UpdateGiftCategory)
			gift.DELETE("/categories/:id", handler.DeleteGiftCategory)

			gift.GET("/exchange", handler.GetExchangeList)
			gift.GET("/exchange/:id", handler.GetExchangeDetail)
			gift.POST("/exchange/:id/audit", handler.AuditExchange)
			gift.POST("/exchange/:id/ship", handler.ShipExchange)
			gift.POST("/exchange/:id/receive", handler.ReceiveExchange)
			gift.POST("/exchange/:id/cancel", handler.CancelExchange)
			gift.GET("/my/exchanges", handler.GetMemberExchanges)

			gift.GET("/statistics", handler.GetGiftStatistics)
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
