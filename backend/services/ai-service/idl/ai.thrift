namespace go ai

struct LawArticle {
    1: i64 id
    2: string title
    3: string content
    4: string category
    5: string law_name
    6: string article_no
    7: string keywords
    8: string vector_id
    9: i32 status
    10: string created_at
}

struct AIConsultRequest {
    1: string question
    2: i64 user_id
}

struct AIConsultResponse {
    1: i32 code
    2: string message
    3: string answer
    4: list<LawArticle> related_articles
}

struct GetLawArticlesRequest {
    1: i32 page
    2: i32 page_size
    3: string keyword
    4: string category
}

struct GetLawArticlesResponse {
    1: i32 code
    2: string message
    3: list<LawArticle> articles
    4: i64 total
}

struct CreateLawArticleRequest {
    1: LawArticle article
}

struct CreateLawArticleResponse {
    1: i32 code
    2: string message
    3: i64 id
}

struct UpdateLawArticleRequest {
    1: LawArticle article
}

struct UpdateLawArticleResponse {
    1: i32 code
    2: string message
}

struct DeleteLawArticleRequest {
    1: i64 id
}

struct DeleteLawArticleResponse {
    1: i32 code
    2: string message
}

struct VectorizeLawArticlesRequest {
    1: list<i64> ids
}

struct VectorizeLawArticlesResponse {
    1: i32 code
    2: string message
    3: i32 processed_count
}

struct GenerateSummaryRequest {
    1: i64 case_id
    2: string mediation_content
}

struct GenerateSummaryResponse {
    1: i32 code
    2: string message
    3: string summary
}

struct SearchSimilarRequest {
    1: string question
    2: i32 top_k
}

struct SearchSimilarResponse {
    1: i32 code
    2: string message
    3: list<LawArticle> articles
    4: list<double> scores
}

service AIService {
    AIConsultResponse AIConsult(1: AIConsultRequest request)
    GetLawArticlesResponse GetLawArticles(1: GetLawArticlesRequest request)
    CreateLawArticleResponse CreateLawArticle(1: CreateLawArticleRequest request)
    UpdateLawArticleResponse UpdateLawArticle(1: UpdateLawArticleRequest request)
    DeleteLawArticleResponse DeleteLawArticle(1: DeleteLawArticleRequest request)
    VectorizeLawArticlesResponse VectorizeLawArticles(1: VectorizeLawArticlesRequest request)
    GenerateSummaryResponse GenerateMediationSummary(1: GenerateSummaryRequest request)
    SearchSimilarResponse SearchSimilarLawArticles(1: SearchSimilarRequest request)
}
