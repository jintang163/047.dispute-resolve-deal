package workflowservice

import (
	"context"

	clientpkg "github.com/cloudwego/kitex/client"

	workflow "github.com/dispute-resolve/workflow-service/kitex_gen/workflow"
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

func (c *kClient) SubmitApproval(ctx context.Context, request *workflow.SubmitApprovalRequest) (r *workflow.SubmitApprovalResponse, err error) {
	r = &workflow.SubmitApprovalResponse{Code: 500, Message: "RPC not implemented - using local service fallback"}
	return
}

func (c *kClient) ProcessApproval(ctx context.Context, request *workflow.ProcessApprovalRequest) (r *workflow.ProcessApprovalResponse, err error) {
	r = &workflow.ProcessApprovalResponse{Code: 500, Message: "RPC not implemented"}
	return
}

func (c *kClient) GetApprovalProgress(ctx context.Context, request *workflow.GetApprovalProgressRequest) (r *workflow.GetApprovalProgressResponse, err error) {
	r = &workflow.GetApprovalProgressResponse{Code: 500, Message: "RPC not implemented"}
	return
}

func (c *kClient) GetApprovalTodoList(ctx context.Context, request *workflow.GetApprovalListRequest) (r *workflow.GetApprovalListResponse, err error) {
	r = &workflow.GetApprovalListResponse{Code: 500, Message: "RPC not implemented"}
	return
}

func (c *kClient) GetApprovalDoneList(ctx context.Context, request *workflow.GetApprovalListRequest) (r *workflow.GetApprovalListResponse, err error) {
	r = &workflow.GetApprovalListResponse{Code: 500, Message: "RPC not implemented"}
	return
}

func (c *kClient) ProcessTimeoutUpgrade(ctx context.Context) (r *workflow.ProcessTimeoutUpgradeResponse, err error) {
	r = &workflow.ProcessTimeoutUpgradeResponse{Code: 500, Message: "RPC not implemented"}
	return
}
