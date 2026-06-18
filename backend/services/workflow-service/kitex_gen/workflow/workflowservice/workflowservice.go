package workflowservice

import (
	"context"

	workflow "github.com/dispute-resolve/workflow-service/kitex_gen/workflow"

	"github.com/cloudwego/kitex/pkg/serviceinfo"
)

type WorkflowService interface {
	SubmitApproval(ctx context.Context, request *workflow.SubmitApprovalRequest) (r *workflow.SubmitApprovalResponse, err error)
	ProcessApproval(ctx context.Context, request *workflow.ProcessApprovalRequest) (r *workflow.ProcessApprovalResponse, err error)
	GetApprovalProgress(ctx context.Context, request *workflow.GetApprovalProgressRequest) (r *workflow.GetApprovalProgressResponse, err error)
	GetApprovalTodoList(ctx context.Context, request *workflow.GetApprovalListRequest) (r *workflow.GetApprovalListResponse, err error)
	GetApprovalDoneList(ctx context.Context, request *workflow.GetApprovalListRequest) (r *workflow.GetApprovalListResponse, err error)
	ProcessTimeoutUpgrade(ctx context.Context) (r *workflow.ProcessTimeoutUpgradeResponse, err error)
}

type Server interface {
	RegisterService(svc *serviceinfo.ServiceInfo)
}

type Client interface {
	WorkflowService
}
