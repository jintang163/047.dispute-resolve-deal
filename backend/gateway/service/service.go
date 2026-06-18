package service

import (
	"sync"
)

var (
	userServiceInstance              UserService
	disputeServiceInstance           DisputeService
	mediationServiceInstance         MediationService
	approvalServiceInstance          ApprovalService
	aiServiceInstance                AIService
	statsServiceInstance             StatsService
	notificationServiceInstance      NotificationService
	performanceServiceInstance       PerformanceService
	dispatchServiceInstance          DispatchService
	videoServiceInstance             VideoService
	esignServiceInstance             ESignService
	judicialConfirmationServiceInstance JudicialConfirmationService
	once                             sync.Once
)

func InitServices(
	user UserService,
	dispute DisputeService,
	mediation MediationService,
	approval ApprovalService,
	ai AIService,
	stats StatsService,
	notification NotificationService,
	performance PerformanceService,
	dispatch DispatchService,
	video VideoService,
	esign ESignService,
	judicialConfirmation JudicialConfirmationService,
) {
	once.Do(func() {
		userServiceInstance = user
		disputeServiceInstance = dispute
		mediationServiceInstance = mediation
		approvalServiceInstance = approval
		aiServiceInstance = ai
		statsServiceInstance = stats
		notificationServiceInstance = notification
		performanceServiceInstance = performance
		dispatchServiceInstance = dispatch
		videoServiceInstance = video
		esignServiceInstance = esign
		judicialConfirmationServiceInstance = judicialConfirmation
	})
}

func UserServiceInst() UserService         { return userServiceInstance }
func DisputeServiceInst() DisputeService      { return disputeServiceInstance }
func MediationServiceInst() MediationService    { return mediationServiceInstance }
func ApprovalServiceInst() ApprovalService     { return approvalServiceInstance }
func AIServiceInst() AIService           { return aiServiceInstance }
func StatsServiceInst() StatsService        { return statsServiceInstance }
func NotificationServiceInst() NotificationService { return notificationServiceInstance }
func PerformanceServiceInst() PerformanceService  { return performanceServiceInstance }
func DispatchServiceInst() DispatchService     { return dispatchServiceInstance }
func VideoServiceInst() VideoService              { return videoServiceInstance }
func ESignServiceInst() ESignService              { return esignServiceInstance }
func JudicialConfirmationServiceInst() JudicialConfirmationService { return judicialConfirmationServiceInstance }
