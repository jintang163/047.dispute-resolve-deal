package service

import (
	"context"

	"github.com/dispute-resolve/common/model"
	"github.com/dispute-resolve/common/vector"
)

type CaseLibraryService interface {
	CreateCase(ctx context.Context, caseLib *model.CaseLibrary) error
	UpdateCase(ctx context.Context, caseLib *model.CaseLibrary) error
	DeleteCase(ctx context.Context, id int64) error
	GetCase(ctx context.Context, id int64) (*model.CaseLibrary, error)
	ListCases(ctx context.Context, page, pageSize int, keyword, disputeType string, difficultyLevel, status int) ([]*model.CaseLibrary, int64, error)
	SearchSimilarCases(ctx context.Context, query string, caseID int64, topK int) ([]*vector.CaseSearchResult, error)
	ScoreCase(ctx context.Context, score *model.CaseLibraryScore) error
	QuoteCase(ctx context.Context, quote *model.CaseLibraryQuote) error
	GetQuoteList(ctx context.Context, sourceCaseID int64) ([]*model.CaseLibraryQuote, error)
	ArchiveCase(ctx context.Context, id int64, archivedBy int64, reason int32) error
	ArchiveUnusedCases(ctx context.Context) (int, error)
	VectorizeCase(ctx context.Context, id int64) error
	VectorizeAllCases(ctx context.Context) (int, error)
	RestoreFromArchive(ctx context.Context, id int64) error
	GetArchiveList(ctx context.Context, page, pageSize int) ([]*model.CaseLibraryArchive, int64, error)
}
