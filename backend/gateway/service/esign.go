package service

import (
	"context"
)

type ESignService interface {
	CreateEsignFlow(ctx context.Context, caseID int64, title string, documentIDs []int64, signerIDs []int64, creatorID int64) (map[string]interface{}, error)
	GetEsignList(ctx context.Context, caseID int64, page, pageSize int, userID int64, role int32) ([]map[string]interface{}, int64, error)
	GetEsignDetail(ctx context.Context, esignID int64, userID int64) (map[string]interface{}, error)
	SignDocument(ctx context.Context, esignID int64, userID int64, verifyCode string, signatureData string) error
	RevokeEsignFlow(ctx context.Context, esignID int64, userID int64, reason string) error
	SendEsignVerifyCode(ctx context.Context, esignID int64, mobile string) error
}
