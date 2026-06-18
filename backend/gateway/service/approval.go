package service

import (
	"context"

	"github.com/dispute-resolve/common/model"
)

type ApprovalService interface {
	SubmitApproval(ctx context.Context, caseID int64, userID int64, workflowID int64) (*model.ApprovalRecord, error)
	ApproveApproval(ctx context.Context, approvalID int64, userID int64, remark string) error
	RejectApproval(ctx context.Context, approvalID int64, userID int64, remark string) error
	ReturnApproval(ctx context.Context, approvalID int64, userID int64, remark string) error
	AddSignApproval(ctx context.Context, approvalID int64, userID int64, signUserID int64, remark string) error
	TransferApproval(ctx context.Context, approvalID int64, userID int64, transferUserID int64, remark string) error
	GetApprovalProgress(ctx context.Context, caseID int64) ([]*model.ApprovalRecord, error)
	GetApprovalTodoList(ctx context.Context, userID int64, page, pageSize int) ([]*model.ApprovalRecord, int64, error)
	GetApprovalDoneList(ctx context.Context, userID int64, page, pageSize int) ([]*model.ApprovalRecord, int64, error)
	ProcessTimeoutUpgrade(ctx context.Context) error
}
