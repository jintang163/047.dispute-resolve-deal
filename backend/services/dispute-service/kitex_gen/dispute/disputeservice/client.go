package disputeservice

import (
	"context"

	clientpkg "github.com/cloudwego/kitex/client"

	dispute "github.com/dispute-resolve/dispute-service/kitex_gen/dispute"
)

type kClient struct {
	c clientpkg.Client
}

func NewClient(destService string, opts ...clientpkg.Option) (Client, error) {
	cli, err := clientpkg.NewClient(destService, opts...)
	if err != nil {
		return nil, err
	}
	return &kClient{c: cli}, nil
}

func (c *kClient) GetDisputeList(ctx context.Context, request *dispute.GetDisputeListRequest) (r *dispute.GetDisputeListResponse, err error) {
	r = &dispute.GetDisputeListResponse{Code: 500, Message: "RPC not implemented - using local service fallback"}
	return
}

func (c *kClient) GetDisputeDetail(ctx context.Context, request *dispute.GetDisputeDetailRequest) (r *dispute.GetDisputeDetailResponse, err error) {
	r = &dispute.GetDisputeDetailResponse{Code: 500, Message: "RPC not implemented"}
	return
}

func (c *kClient) CreateDispute(ctx context.Context, request *dispute.CreateDisputeRequest) (r *dispute.CreateDisputeResponse, err error) {
	r = &dispute.CreateDisputeResponse{Code: 500, Message: "RPC not implemented"}
	return
}

func (c *kClient) KioskCreateDispute(ctx context.Context, request *dispute.KioskCreateDisputeRequest) (r *dispute.CreateDisputeResponse, err error) {
	r = &dispute.CreateDisputeResponse{Code: 500, Message: "RPC not implemented"}
	return
}

func (c *kClient) MiniAppCreateDispute(ctx context.Context, request *dispute.CreateDisputeRequest) (r *dispute.CreateDisputeResponse, err error) {
	r = &dispute.CreateDisputeResponse{Code: 500, Message: "RPC not implemented"}
	return
}

func (c *kClient) AssignDispute(ctx context.Context, request *dispute.AssignDisputeRequest) (r *dispute.AssignDisputeResponse, err error) {
	r = &dispute.AssignDisputeResponse{Code: 500, Message: "RPC not implemented"}
	return
}

func (c *kClient) UrgeDispute(ctx context.Context, request *dispute.UrgeDisputeRequest) (r *dispute.UrgeDisputeResponse, err error) {
	r = &dispute.UrgeDisputeResponse{Code: 500, Message: "RPC not implemented"}
	return
}

func (c *kClient) UpdateDisputeStatus(ctx context.Context, request *dispute.UpdateDisputeStatusRequest) (r *dispute.UpdateDisputeStatusResponse, err error) {
	r = &dispute.UpdateDisputeStatusResponse{Code: 500, Message: "RPC not implemented"}
	return
}

func (c *kClient) GetDisputeHistory(ctx context.Context, request *dispute.GetDisputeHistoryRequest) (r *dispute.GetDisputeHistoryResponse, err error) {
	r = &dispute.GetDisputeHistoryResponse{Code: 500, Message: "RPC not implemented"}
	return
}

func (c *kClient) GetDisputeProgress(ctx context.Context, request *dispute.GetDisputeProgressRequest) (r *dispute.GetDisputeProgressResponse, err error) {
	r = &dispute.GetDisputeProgressResponse{Code: 500, Message: "RPC not implemented"}
	return
}

func (c *kClient) GetDisputeTypes(ctx context.Context) (r *dispute.GetDisputeTypesResponse, err error) {
	r = &dispute.GetDisputeTypesResponse{Code: 500, Message: "RPC not implemented"}
	return
}

func (c *kClient) UploadEvidence(ctx context.Context, request *dispute.UploadEvidenceRequest) (r *dispute.UploadEvidenceResponse, err error) {
	r = &dispute.UploadEvidenceResponse{Code: 500, Message: "RPC not implemented"}
	return
}

func (c *kClient) GetEvidenceList(ctx context.Context, request *dispute.GetEvidenceListRequest) (r *dispute.GetEvidenceListResponse, err error) {
	r = &dispute.GetEvidenceListResponse{Code: 500, Message: "RPC not implemented"}
	return
}

func (c *kClient) DeleteEvidence(ctx context.Context, request *dispute.DeleteEvidenceRequest) (r *dispute.DeleteEvidenceResponse, err error) {
	r = &dispute.DeleteEvidenceResponse{Code: 500, Message: "RPC not implemented"}
	return
}

func (c *kClient) BatchDeleteEvidence(ctx context.Context, request *dispute.BatchDeleteEvidenceRequest) (r *dispute.BatchDeleteEvidenceResponse, err error) {
	r = &dispute.BatchDeleteEvidenceResponse{Code: 500, Message: "RPC not implemented"}
	return
}

func (c *kClient) UpdateEvidenceRemark(ctx context.Context, request *dispute.UpdateEvidenceRemarkRequest) (r *dispute.UpdateEvidenceRemarkResponse, err error) {
	r = &dispute.UpdateEvidenceRemarkResponse{Code: 500, Message: "RPC not implemented"}
	return
}
