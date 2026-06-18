package com.dispute.app.api

import com.dispute.app.model.AIMessage
import kotlinx.serialization.Serializable

class AIApi(private val client: ApiClient) {

    suspend fun sendMessage(
        message: String,
        conversationId: String? = null,
        disputeType: String? = null
    ): AIConversationResponse {
        val request = SendMessageRequest(message, conversationId, disputeType)
        val response: ApiResponse<AIConversationResponse> = client.post("/api/ai/chat", request)
        return response.getOrThrow()
    }

    suspend fun getQuickQuestions(disputeType: String? = null): List<String> {
        val params = disputeType?.let { mapOf("disputeType" to it) } ?: emptyMap()
        val response: ApiResponse<List<String>> = client.get("/api/ai/quick-questions", params)
        return response.getOrThrow()
    }

    suspend fun getConversationHistory(conversationId: String): List<AIMessage> {
        val response: ApiResponse<List<AIMessage>> = client.get("/api/ai/history/$conversationId")
        return response.getOrThrow()
    }

    suspend fun clearConversation(conversationId: String) {
        val response: ApiResponse<Unit> = client.delete("/api/ai/conversation/$conversationId")
        if (!response.isSuccess) {
            throw ApiException.BusinessError(response.code, response.message)
        }
    }

    suspend fun getLegalKnowledge(keyword: String? = null): List<LegalKnowledge> {
        val params = keyword?.let { mapOf("keyword" to it) } ?: emptyMap()
        val response: ApiResponse<List<LegalKnowledge>> = client.get("/api/ai/knowledge", params)
        return response.getOrThrow()
    }

    suspend fun generateSuggestions(disputeType: String, description: String): List<String> {
        val request = SuggestionRequest(disputeType, description)
        val response: ApiResponse<List<String>> = client.post("/api/ai/suggestions", request)
        return response.getOrThrow()
    }
}

@Serializable
data class SendMessageRequest(
    val message: String,
    val conversationId: String?,
    val disputeType: String?
)

@Serializable
data class AIConversationResponse(
    val conversationId: String,
    val reply: String,
    val suggestedQuestions: List<String>? = null,
    val relatedLaws: List<RelatedLaw>? = null
)

@Serializable
data class RelatedLaw(
    val title: String,
    val article: String,
    val content: String
)

@Serializable
data class LegalKnowledge(
    val id: String,
    val title: String,
    val summary: String,
    val category: String,
    val content: String? = null
)

@Serializable
data class SuggestionRequest(
    val disputeType: String,
    val description: String
)
