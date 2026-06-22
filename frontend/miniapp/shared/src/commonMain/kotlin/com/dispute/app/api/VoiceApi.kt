package com.dispute.app.api

import kotlinx.serialization.Serializable

class VoiceApi(private val client: ApiClient) {

    suspend fun recognizeSpeech(fileName: String, fileBase64: String, format: String): VoiceRecognizeResult {
        val request = VoiceRecognizeRequest(fileName, fileBase64, format)
        val response: ApiResponse<VoiceRecognizeResult> = client.post("/public/voice/recognize", request)
        return response.getOrThrow()
    }

    suspend fun submitTranscribeTask(
        fileBase64: String,
        fileName: String,
        format: String,
        caseId: String? = null,
        recordId: String? = null,
        enableDiarization: Boolean = false,
        speakerCount: Int = 2
    ): VoiceTranscribeTaskResult {
        val request = SubmitTranscribeRequest(
            fileBase64 = fileBase64,
            fileName = fileName,
            format = format,
            caseId = caseId,
            recordId = recordId,
            enableDiarization = enableDiarization,
            speakerCount = speakerCount
        )
        val response: ApiResponse<VoiceTranscribeTaskResult> = client.post("/public/voice/transcribe/submit", request)
        return response.getOrThrow()
    }

    suspend fun getTranscribeTask(taskId: String): VoiceTranscribeTaskResult {
        val response: ApiResponse<VoiceTranscribeTaskResult> = client.get("/public/voice/transcribe/task/$taskId")
        return response.getOrThrow()
    }

    suspend fun syncTranscribe(fileBase64: String, fileName: String, format: String): VoiceRecognizeResult {
        val request = VoiceRecognizeRequest(fileName, fileBase64, format)
        val response: ApiResponse<VoiceRecognizeResult> = client.post("/public/voice/transcribe/sync", request)
        return response.getOrThrow()
    }

    suspend fun saveTranscribeResult(
        caseId: String,
        recordId: String,
        transcriptText: String,
        transcribeTaskId: String? = null,
        audioUrl: String? = null,
        duration: Int? = null
    ): Boolean {
        val request = SaveTranscribeResultRequest(
            transcriptText = transcriptText,
            transcribeTaskId = transcribeTaskId,
            audioUrl = audioUrl,
            duration = duration
        )
        val response: ApiResponse<Unit> = client.post("/dispute/$caseId/mediation/$recordId/transcribe-result", request)
        return response.isSuccess
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

@Serializable
data class SubmitTranscribeRequest(
    val fileBase64: String,
    val fileName: String,
    val format: String,
    val caseId: String? = null,
    val recordId: String? = null,
    val enableDiarization: Boolean = false,
    val speakerCount: Int = 2
)

@Serializable
data class VoiceTranscribeTaskResult(
    val taskId: String,
    val status: String,
    val transcriptText: String? = null,
    val duration: Int = 0,
    val wordCount: Int = 0,
    val sentences: List<TranscribeSentence> = emptyList(),
    val errorMsg: String? = null
)

@Serializable
data class TranscribeSentence(
    val text: String,
    val beginTime: Int = 0,
    val endTime: Int = 0,
    val speakerId: Int = 0
)

@Serializable
data class SaveTranscribeResultRequest(
    val transcriptText: String,
    val transcribeTaskId: String? = null,
    val audioUrl: String? = null,
    val duration: Int? = null
)
