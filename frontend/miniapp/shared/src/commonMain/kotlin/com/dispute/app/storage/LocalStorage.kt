package com.dispute.app.storage

import kotlinx.serialization.Serializable

expect object LocalStorage {
    fun getString(key: String): String?
    fun setString(key: String, value: String)
    fun remove(key: String)

    suspend fun putAudio(taskId: String, audioBase64: String, meta: AudioStorageMeta): Boolean
    suspend fun getAudio(taskId: String): String?
    suspend fun removeAudio(taskId: String): Boolean
    suspend fun getAllAudioTasks(): List<AudioStorageMeta>

    suspend fun addTranscribeTask(task: OfflineTranscribeTask): Boolean
    suspend fun updateTranscribeTask(task: OfflineTranscribeTask): Boolean
    suspend fun getTranscribeTasks(status: String? = null): List<OfflineTranscribeTask>
    suspend fun removeTranscribeTask(taskId: String): Boolean
}

@Serializable
data class AudioStorageMeta(
    val taskId: String,
    val fileName: String,
    val format: String,
    val sizeBytes: Long,
    val durationMs: Long = 0L,
    val createdAt: Long = 0L,
    val caseId: String? = null,
    val recordId: String? = null
)

@Serializable
enum class TranscribeTaskStatus(val value: String) {
    PENDING("pending"),
    UPLOADING("uploading"),
    PROCESSING("processing"),
    COMPLETED("completed"),
    FAILED("failed");

    companion object {
        fun fromValue(value: String): TranscribeTaskStatus =
            values().find { it.value == value } ?: PENDING
    }
}

@Serializable
data class OfflineTranscribeTask(
    val taskId: String,
    val status: TranscribeTaskStatus,
    val audioBase64: String,
    val fileName: String,
    val format: String,
    val sizeBytes: Long,
    val durationMs: Long = 0L,
    val caseId: String? = null,
    val recordId: String? = null,
    val remoteTaskId: String? = null,
    val transcriptText: String? = null,
    val errorMessage: String? = null,
    val retryCount: Int = 0,
    val createdAt: Long = 0L,
    val updatedAt: Long = 0L
)
