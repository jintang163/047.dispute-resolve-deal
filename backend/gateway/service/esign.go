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
	GetEsignProgress(ctx context.Context, flowID string) (map[string]interface{}, error)
	StoreSignedDocToBlockchain(ctx context.Context, esignID int64) (map[string]interface{}, error)
	NotifySignerProgress(ctx context.Context, flowID string, signerID int64, notifyType string) error
}

type BlockchainService interface {
	StoreEvidence(ctx context.Context, caseID int64, evidenceID string, evidenceType string, evidenceName string, evidenceHash string, flowID string, metadata string, creatorID int64) (map[string]interface{}, error)
	GetCertList(ctx context.Context, caseID int64, evidenceType string, page, pageSize int, userID int64, role int32) ([]map[string]interface{}, int64, error)
	GetCertDetail(ctx context.Context, certNo string) (map[string]interface{}, error)
	VerifyEvidence(ctx context.Context, certNo string) (map[string]interface{}, error)
	DownloadCert(ctx context.Context, certNo string) (map[string]interface{}, error)
	PublicVerify(ctx context.Context, certNo string) (map[string]interface{}, error)
}
