package ai

type LawArticle struct {
	Id        int64  `json:"id" th:"1,optional"`
	Title     string `json:"title" th:"2,optional"`
	Content   string `json:"content" th:"3,optional"`
	Category  string `json:"category" th:"4,optional"`
	LawName   string `json:"law_name" th:"5,optional"`
	ArticleNo string `json:"article_no" th:"6,optional"`
	Keywords  string `json:"keywords" th:"7,optional"`
	VectorId  string `json:"vector_id" th:"8,optional"`
	Status    int32  `json:"status" th:"9,optional"`
	CreatedAt string `json:"created_at" th:"10,optional"`
}

type AIConsultRequest struct {
	Question string `json:"question" th:"1,optional"`
	UserId   int64  `json:"user_id" th:"2,optional"`
}

type AIConsultResponse struct {
	Code            int32         `json:"code" th:"1,optional"`
	Message         string        `json:"message" th:"2,optional"`
	Answer          string        `json:"answer" th:"3,optional"`
	RelatedArticles []*LawArticle `json:"related_articles" th:"4,optional"`
}

type GetLawArticlesRequest struct {
	Page     int32  `json:"page" th:"1,optional"`
	PageSize int32  `json:"page_size" th:"2,optional"`
	Keyword  string `json:"keyword" th:"3,optional"`
	Category string `json:"category" th:"4,optional"`
}

type GetLawArticlesResponse struct {
	Code     int32         `json:"code" th:"1,optional"`
	Message  string        `json:"message" th:"2,optional"`
	Articles []*LawArticle `json:"articles" th:"3,optional"`
	Total    int64         `json:"total" th:"4,optional"`
}

type CreateLawArticleRequest struct {
	Article *LawArticle `json:"article" th:"1,optional"`
}

type CreateLawArticleResponse struct {
	Code    int32  `json:"code" th:"1,optional"`
	Message string `json:"message" th:"2,optional"`
	Id      int64  `json:"id" th:"3,optional"`
}

type UpdateLawArticleRequest struct {
	Article *LawArticle `json:"article" th:"1,optional"`
}

type UpdateLawArticleResponse struct {
	Code    int32  `json:"code" th:"1,optional"`
	Message string `json:"message" th:"2,optional"`
}

type DeleteLawArticleRequest struct {
	Id int64 `json:"id" th:"1,optional"`
}

type DeleteLawArticleResponse struct {
	Code    int32  `json:"code" th:"1,optional"`
	Message string `json:"message" th:"2,optional"`
}

type VectorizeLawArticlesRequest struct {
	Ids []int64 `json:"ids" th:"1,optional"`
}

type VectorizeLawArticlesResponse struct {
	Code           int32  `json:"code" th:"1,optional"`
	Message        string `json:"message" th:"2,optional"`
	ProcessedCount int32  `json:"processed_count" th:"3,optional"`
}

type GenerateSummaryRequest struct {
	CaseId           int64  `json:"case_id" th:"1,optional"`
	MediationContent string `json:"mediation_content" th:"2,optional"`
}

type GenerateSummaryResponse struct {
	Code    int32  `json:"code" th:"1,optional"`
	Message string `json:"message" th:"2,optional"`
	Summary string `json:"summary" th:"3,optional"`
}

type SearchSimilarRequest struct {
	Question string `json:"question" th:"1,optional"`
	TopK     int32  `json:"top_k" th:"2,optional"`
}

type SearchSimilarResponse struct {
	Code     int32         `json:"code" th:"1,optional"`
	Message  string        `json:"message" th:"2,optional"`
	Articles []*LawArticle `json:"articles" th:"3,optional"`
	Scores   []float64     `json:"scores" th:"4,optional"`
}
