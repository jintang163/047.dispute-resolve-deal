package disputeservice

import (
	"context"

	dispute "github.com/dispute-resolve/dispute-service/kitex_gen/dispute"

	"github.com/cloudwego/kitex/pkg/serviceinfo"
)

type DisputeService interface {
	GetDisputeList(ctx context.Context, request *dispute.GetDisputeListRequest) (r *dispute.GetDisputeListResponse, err error)
	GetDisputeDetail(ctx context.Context, request *dispute.GetDisputeDetailRequest) (r *dispute.GetDisputeDetailResponse, err error)
	CreateDispute(ctx context.Context, request *dispute.CreateDisputeRequest) (r *dispute.CreateDisputeResponse, err error)
	KioskCreateDispute(ctx context.Context, request *dispute.KioskCreateDisputeRequest) (r *dispute.CreateDisputeResponse, err error)
	MiniAppCreateDispute(ctx context.Context, request *dispute.CreateDisputeRequest) (r *dispute.CreateDisputeResponse, err error)
	AssignDispute(ctx context.Context, request *dispute.AssignDisputeRequest) (r *dispute.AssignDisputeResponse, err error)
	UrgeDispute(ctx context.Context, request *dispute.UrgeDisputeRequest) (r *dispute.UrgeDisputeResponse, err error)
	UpdateDisputeStatus(ctx context.Context, request *dispute.UpdateDisputeStatusRequest) (r *dispute.UpdateDisputeStatusResponse, err error)
	GetDisputeHistory(ctx context.Context, request *dispute.GetDisputeHistoryRequest) (r *dispute.GetDisputeHistoryResponse, err error)
	GetDisputeProgress(ctx context.Context, request *dispute.GetDisputeProgressRequest) (r *dispute.GetDisputeProgressResponse, err error)
	GetDisputeTypes(ctx context.Context) (r *dispute.GetDisputeTypesResponse, err error)
	UploadEvidence(ctx context.Context, request *dispute.UploadEvidenceRequest) (r *dispute.UploadEvidenceResponse, err error)
	GetEvidenceList(ctx context.Context, request *dispute.GetEvidenceListRequest) (r *dispute.GetEvidenceListResponse, err error)
	DeleteEvidence(ctx context.Context, request *dispute.DeleteEvidenceRequest) (r *dispute.DeleteEvidenceResponse, err error)
	BatchDeleteEvidence(ctx context.Context, request *dispute.BatchDeleteEvidenceRequest) (r *dispute.BatchDeleteEvidenceResponse, err error)
	UpdateEvidenceRemark(ctx context.Context, request *dispute.UpdateEvidenceRemarkRequest) (r *dispute.UpdateEvidenceRemarkResponse, err error)
}

type Server interface {
	RegisterService(svc *serviceinfo.ServiceInfo)
}

type Client interface {
	DisputeService
}
