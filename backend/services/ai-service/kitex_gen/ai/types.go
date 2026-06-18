package ai

type LawArticle struct {
	Id        int64  `thrift:"id,1" frugal:"1,default,i64" json:"id"`
	Title     string `thrift:"title,2" frugal:"2,default,string" json:"title"`
	Content   string `thrift:"content,3" frugal:"3,default,string" json:"content"`
	Category  string `thrift:"category,4" frugal:"4,default,string" json:"category"`
	LawName   string `thrift:"law_name,5" frugal:"5,default,string" json:"lawName"`
	ArticleNo string `thrift:"article_no,6" frugal:"6,default,string" json:"articleNo"`
	Keywords  string `thrift:"keywords,7" frugal:"7,default,string" json:"keywords"`
	VectorId  string `thrift:"vector_id,8" frugal:"8,default,string" json:"vectorId"`
	Status    int32  `thrift:"status,9" frugal:"9,default,i32" json:"status"`
	CreatedAt string `thrift:"created_at,10" frugal:"10,default,string" json:"createdAt"`
}

type AIConsultRequest struct {
	Question string `thrift:"question,1" frugal:"1,default,string" json:"question"`
	UserId   int64  `thrift:"user_id,2" frugal:"2,default,i64" json:"userId"`
}

type AIConsultResponse struct {
	Code            int32         `thrift:"code,1" frugal:"1,default,i32" json:"code"`
	Message         string        `thrift:"message,2" frugal:"2,default,string" json:"message"`
	Answer          string        `thrift:"answer,3" frugal:"3,default,string" json:"answer"`
	RelatedArticles []*LawArticle `thrift:"related_articles,4" frugal:"4,default,list<*LawArticle>" json:"relatedArticles,omitempty"`
}

type GetLawArticlesRequest struct {
	Page     int32  `thrift:"page,1" frugal:"1,default,i32" json:"page"`
	PageSize int32  `thrift:"page_size,2" frugal:"2,default,i32" json:"pageSize"`
	Keyword  string `thrift:"keyword,3" frugal:"3,default,string" json:"keyword"`
	Category string `thrift:"category,4" frugal:"4,default,string" json:"category"`
}

type GetLawArticlesResponse struct {
	Code     int32         `thrift:"code,1" frugal:"1,default,i32" json:"code"`
	Message  string        `thrift:"message,2" frugal:"2,default,string" json:"message"`
	Articles []*LawArticle `thrift:"articles,3" frugal:"3,default,list<*LawArticle>" json:"articles,omitempty"`
	Total    int64         `thrift:"total,4" frugal:"4,default,i64" json:"total"`
}

type CreateLawArticleRequest struct {
	Article *LawArticle `thrift:"article,1" frugal:"1,default,*LawArticle" json:"article,omitempty"`
}

type CreateLawArticleResponse struct {
	Code    int32  `thrift:"code,1" frugal:"1,default,i32" json:"code"`
	Message string `thrift:"message,2" frugal:"2,default,string" json:"message"`
	Id      int64  `thrift:"id,3" frugal:"3,default,i64" json:"id"`
}

type UpdateLawArticleRequest struct {
	Article *LawArticle `thrift:"article,1" frugal:"1,default,*LawArticle" json:"article,omitempty"`
}

type UpdateLawArticleResponse struct {
	Code    int32  `thrift:"code,1" frugal:"1,default,i32" json:"code"`
	Message string `thrift:"message,2" frugal:"2,default,string" json:"message"`
}

type DeleteLawArticleRequest struct {
	Id int64 `thrift:"id,1" frugal:"1,default,i64" json:"id"`
}

type DeleteLawArticleResponse struct {
	Code    int32  `thrift:"code,1" frugal:"1,default,i32" json:"code"`
	Message string `thrift:"message,2" frugal:"2,default,string" json:"message"`
}

type VectorizeLawArticlesRequest struct {
	Ids []int64 `thrift:"ids,1" frugal:"1,default,list<i64>" json:"ids"`
}

type VectorizeLawArticlesResponse struct {
	Code           int32  `thrift:"code,1" frugal:"1,default,i32" json:"code"`
	Message        string `thrift:"message,2" frugal:"2,default,string" json:"message"`
	ProcessedCount int32  `thrift:"processed_count,3" frugal:"3,default,i32" json:"processedCount"`
}

type GenerateSummaryRequest struct {
	CaseId           int64  `thrift:"case_id,1" frugal:"1,default,i64" json:"caseId"`
	MediationContent string `thrift:"mediation_content,2" frugal:"2,default,string" json:"mediationContent"`
}

type GenerateSummaryResponse struct {
	Code    int32  `thrift:"code,1" frugal:"1,default,i32" json:"code"`
	Message string `thrift:"message,2" frugal:"2,default,string" json:"message"`
	Summary string `thrift:"summary,3" frugal:"3,default,string" json:"summary"`
}

type SearchSimilarRequest struct {
	Question string `thrift:"question,1" frugal:"1,default,string" json:"question"`
	TopK     int32  `thrift:"top_k,2" frugal:"2,default,i32" json:"topK"`
}

type SearchSimilarResponse struct {
	Code     int32         `thrift:"code,1" frugal:"1,default,i32" json:"code"`
	Message  string        `thrift:"message,2" frugal:"2,default,string" json:"message"`
	Articles []*LawArticle `thrift:"articles,3" frugal:"3,default,list<*LawArticle>" json:"articles,omitempty"`
	Scores   []float64     `thrift:"scores,4" frugal:"4,default,list<double>" json:"scores"`
}
