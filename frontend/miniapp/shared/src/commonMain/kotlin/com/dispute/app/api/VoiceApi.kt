package com.dispute.app.api

import kotlinx.serialization.Serializable

class VoiceApi(private val client: ApiClient) {

    suspend fun recognizeSpeech(fileName: String, fileBase64: String, format: String): VoiceRecognizeResult {
        val request = VoiceRecognizeRequest(fileName, fileBase64, format)
        val response: ApiResponse<VoiceRecognizeResult> = client.post("/api/v1/public/voice/recognize", request)
        return response.getOrThrow()
    }
}

@Serializable
data class VoiceRecognizeRequest(
    val fileName: String,
    val fileBase64: String,
    val format: String
)

@Serializable
data class VoiceRecognizeResult(
    val text: String,
    val duration: Int,
    val taskId: String,
    val format: String,
    val fileSize: Long
)
