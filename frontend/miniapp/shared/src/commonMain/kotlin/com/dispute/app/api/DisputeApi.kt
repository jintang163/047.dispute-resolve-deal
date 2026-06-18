package com.dispute.app.api

import com.dispute.app.model.Case
import com.dispute.app.model.DisputeType
import com.dispute.app.model.Evidence
import com.dispute.app.model.MediationProgress
import kotlinx.serialization.Serializable

class DisputeApi(private val client: ApiClient) {

    suspend fun getDisputeTypes(): List<DisputeType> {
        val response: ApiResponse<List<DisputeType>> = client.get("/api/dispute/types")
        return response.getOrThrow()
    }

    suspend fun submitCase(request: SubmitCaseRequest): SubmitCaseResponse {
        val response: ApiResponse<SubmitCaseResponse> = client.post("/api/case/submit", request)
        return response.getOrThrow()
    }

    suspend fun getCaseList(status: Case.Status? = null): List<Case> {
        val params = status?.let { mapOf("status" to it.name) } ?: emptyMap()
        val response: ApiResponse<List<Case>> = client.get("/api/case/list", params)
        return response.getOrThrow()
    }

    suspend fun getCaseDetail(caseNumber: String): Case {
        val response: ApiResponse<Case> = client.get("/api/case/detail/$caseNumber")
        return response.getOrThrow()
    }

    suspend fun getCaseProgress(caseNumber: String): List<MediationProgress> {
        val response: ApiResponse<List<MediationProgress>> =
            client.get("/api/case/progress/$caseNumber")
        return response.getOrThrow()
    }

    suspend fun urgeCase(caseNumber: String, reason: String = "") {
        val request = mapOf(
            "caseNumber" to caseNumber,
            "reason" to reason
        )
        val response: ApiResponse<Unit> = client.post("/api/case/urge", request)
        if (!response.isSuccess) {
            throw ApiException.BusinessError(response.code, response.message)
        }
    }

    suspend fun supplementEvidence(caseNumber: String, evidenceList: List<Evidence>) {
        val request = SupplementEvidenceRequest(caseNumber, evidenceList)
        val response: ApiResponse<Unit> = client.post("/api/case/supplement", request)
        if (!response.isSuccess) {
            throw ApiException.BusinessError(response.code, response.message)
        }
    }

    suspend fun submitSatisfaction(caseNumber: String, rating: Int, comment: String) {
        val request = SatisfactionRequest(caseNumber, rating, comment)
        val response: ApiResponse<Unit> = client.post("/api/case/satisfaction", request)
        if (!response.isSuccess) {
            throw ApiException.BusinessError(response.code, response.message)
        }
    }

    suspend fun uploadEvidence(fileName: String, fileBytes: ByteArray, fileType: String): Evidence {
        val request = UploadEvidenceRequest(fileName, fileBytes.encodeToString(), fileType)
        val response: ApiResponse<Evidence> = client.post("/api/evidence/upload", request)
        return response.getOrThrow()
    }

    suspend fun getReceipt(caseNumber: String): ReceiptData {
        val response: ApiResponse<ReceiptData> = client.get("/api/case/receipt/$caseNumber")
        return response.getOrThrow()
    }
}

@Serializable
data class SubmitCaseRequest(
    val userId: String,
    val userName: String,
    val idNumber: String,
    val phone: String,
    val disputeTypePath: List<String>,
    val disputeTypeName: String,
    val opponentName: String,
    val opponentPhone: String,
    val opponentAddress: String,
    val description: String,
    val expectedResolution: String,
    val evidenceList: List<Evidence>
)

@Serializable
data class SubmitCaseResponse(
    val caseNumber: String,
    val createdAt: String,
    val estimatedDays: Int,
    val mediatorName: String?,
    val mediatorPhone: String?
)

@Serializable
data class SupplementEvidenceRequest(
    val caseNumber: String,
    val evidenceList: List<Evidence>
)

@Serializable
data class SatisfactionRequest(
    val caseNumber: String,
    val rating: Int,
    val comment: String
)

@Serializable
data class UploadEvidenceRequest(
    val fileName: String,
    val fileBase64: String,
    val fileType: String
)

@Serializable
data class ReceiptData(
    val caseNumber: String,
    val applicantName: String,
    val disputeType: String,
    val submitTime: String,
    val mediatorName: String?,
    val mediatorPhone: String?,
    val serviceHotline: String,
    val qrCodeContent: String
)
