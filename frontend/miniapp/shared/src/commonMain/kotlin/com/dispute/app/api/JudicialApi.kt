package com.dispute.app.api

import com.dispute.app.model.CourtOption
import com.dispute.app.model.CreateJudicialRequest
import com.dispute.app.model.CreateJudicialResponse
import com.dispute.app.model.JudicialConfirmLog
import com.dispute.app.model.JudicialConfirmation
import kotlinx.serialization.Serializable

class JudicialApi(private val client: ApiClient) {

    suspend fun getJudicialList(status: JudicialConfirmation.Status? = null): List<JudicialConfirmation> {
        val params = status?.let { mapOf("status" to it.ordinal) } ?: emptyMap()
        val response: ApiResponse<List<JudicialConfirmation>> = client.get("/api/v1/judicial/list", params)
        return response.getOrThrow()
    }

    suspend fun getJudicialDetail(id: Long): JudicialConfirmation {
        val response: ApiResponse<JudicialConfirmation> = client.get("/api/v1/judicial/$id")
        return response.getOrThrow()
    }

    suspend fun queryJudicialByNo(confirmNo: String, idCard: String): JudicialConfirmation {
        val params = mapOf(
            "confirmNo" to confirmNo,
            "idCard" to idCard
        )
        val response: ApiResponse<JudicialConfirmation> = client.get("/api/v1/judicial/query", params)
        return response.getOrThrow()
    }

    suspend fun createJudicialConfirmation(request: CreateJudicialRequest): CreateJudicialResponse {
        val response: ApiResponse<CreateJudicialResponse> = client.post("/api/v1/judicial", request)
        return response.getOrThrow()
    }

    suspend fun submitToCourt(id: Long) {
        val response: ApiResponse<Unit> = client.post("/api/v1/judicial/$id/submit", emptyMap<String, Any>())
        if (!response.isSuccess) {
            throw ApiException.BusinessError(response.code, response.message)
        }
    }

    suspend fun queryCourtStatus(id: Long) {
        val response: ApiResponse<Unit> = client.post("/api/v1/judicial/$id/query-status", emptyMap<String, Any>())
        if (!response.isSuccess) {
            throw ApiException.BusinessError(response.code, response.message)
        }
    }

    suspend fun generateDocument(id: Long): String {
        val response: ApiResponse<Map<String, String>> = client.post("/api/v1/judicial/$id/generate-doc", emptyMap<String, Any>())
        return response.getOrThrow()["documentUrl"] ?: ""
    }

    suspend fun sealDocument(id: Long) {
        val response: ApiResponse<Unit> = client.post("/api/v1/judicial/$id/seal", emptyMap<String, Any>())
        if (!response.isSuccess) {
            throw ApiException.BusinessError(response.code, response.message)
        }
    }

    suspend fun getConfirmLogs(id: Long): List<JudicialConfirmLog> {
        val response: ApiResponse<List<JudicialConfirmLog>> = client.get("/api/v1/judicial/$id/logs")
        return response.getOrThrow()
    }

    suspend fun sendPerformanceReminder(id: Long) {
        val response: ApiResponse<Unit> = client.post("/api/v1/judicial/$id/remind/performance", emptyMap<String, Any>())
        if (!response.isSuccess) {
            throw ApiException.BusinessError(response.code, response.message)
        }
    }

    suspend fun sendExpirationReminder(id: Long) {
        val response: ApiResponse<Unit> = client.post("/api/v1/judicial/$id/remind/expiration", emptyMap<String, Any>())
        if (!response.isSuccess) {
            throw ApiException.BusinessError(response.code, response.message)
        }
    }

    suspend fun getCourtOptions(): List<CourtOption> {
        val response: ApiResponse<List<CourtOption>> = client.get("/api/v1/court/options")
        return response.getOrThrow()
    }
}
