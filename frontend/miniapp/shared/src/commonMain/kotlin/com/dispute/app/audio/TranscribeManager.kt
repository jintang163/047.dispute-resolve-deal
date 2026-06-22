package com.dispute.app.audio

import androidx.compose.runtime.State
import androidx.compose.runtime.mutableStateListOf
import androidx.compose.runtime.mutableStateOf
import com.dispute.app.api.ApiClient
import com.dispute.app.api.VoiceTranscribeTaskResult
import com.dispute.app.platform.NetworkStatusService
import com.dispute.app.storage.AudioStorageMeta
import com.dispute.app.storage.LocalStorage
import com.dispute.app.storage.OfflineTranscribeTask
import com.dispute.app.storage.TranscribeTaskStatus
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.Job
import kotlinx.coroutines.SupervisorJob
import kotlinx.coroutines.delay
import kotlinx.coroutines.launch

class TranscribeManager(
    private val apiClient: ApiClient
) {
    private val tasks = mutableStateListOf<OfflineTranscribeTask>()

    private val _pendingCount = mutableStateOf(0)
    val pendingCount: State<Int> = _pendingCount

    private val _processingCount = mutableStateOf(0)
    val processingCount: State<Int> = _processingCount

    private val scope = CoroutineScope(SupervisorJob() + Dispatchers.Main)
    private var pollJob: Job? = null
    private var started = false

    private var wasOnline = true

    fun start() {
        if (started) return
        started = true

        scope.launch {
            loadTasksFromStorage()
            updateCounts()
        }

        NetworkStatusService.startMonitoring()
        NetworkStatusService.setOnNetworkChangeListener { isOnline, _ ->
            scope.launch {
                if (isOnline && !wasOnline) {
                    syncPendingTasks()
                }
                wasOnline = isOnline
            }
        }

        wasOnline = NetworkStatusService.isOnline

        startPolling()
    }

    fun stop() {
        if (!started) return
        started = false

        stopPolling()
        NetworkStatusService.stopMonitoring()
    }

    suspend fun submitTask(task: OfflineTranscribeTask): Boolean {
        val existingIndex = tasks.indexOfFirst { it.taskId == task.taskId }
        if (existingIndex >= 0) {
            tasks[existingIndex] = task
            LocalStorage.updateTranscribeTask(task)
        } else {
            tasks.add(task)
            LocalStorage.addTranscribeTask(task)
        }

        val meta = AudioStorageMeta(
            taskId = task.taskId,
            fileName = task.fileName,
            format = task.format,
            sizeBytes = task.sizeBytes,
            durationMs = task.durationMs,
            createdAt = task.createdAt,
            caseId = task.caseId,
            recordId = task.recordId
        )
        LocalStorage.putAudio(task.taskId, task.audioBase64, meta)

        updateCounts()

        if (NetworkStatusService.isOnline) {
            scope.launch {
                processTask(task.taskId)
            }
        }

        return true
    }

    suspend fun retryTask(taskId: String): Boolean {
        val task = tasks.find { it.taskId == taskId } ?: return false
        val updatedTask = task.copy(
            status = TranscribeTaskStatus.PENDING,
            errorMessage = null,
            retryCount = task.retryCount + 1,
            updatedAt = System.currentTimeMillis()
        )
        return submitTask(updatedTask)
    }

    suspend fun syncPendingTasks() {
        val pendingTasks = tasks.filter {
            it.status == TranscribeTaskStatus.PENDING || it.status == TranscribeTaskStatus.FAILED
        }

        pendingTasks.forEach { task ->
            scope.launch {
                processTask(task.taskId)
            }
        }
    }

    suspend fun pollProcessingTasks() {
        val processingTasks = tasks.filter {
            it.status == TranscribeTaskStatus.UPLOADING || it.status == TranscribeTaskStatus.PROCESSING
        }

        processingTasks.forEach { task ->
            scope.launch {
                pollTaskStatus(task.taskId)
            }
        }
    }

    fun getTask(taskId: String): OfflineTranscribeTask? {
        return tasks.find { it.taskId == taskId }
    }

    fun getTasksByCase(caseId: String): List<OfflineTranscribeTask> {
        return tasks.filter { it.caseId == caseId }
    }

    fun getAllTasks(): List<OfflineTranscribeTask> {
        return tasks.toList()
    }

    private suspend fun loadTasksFromStorage() {
        val storedTasks = LocalStorage.getTranscribeTasks()
        tasks.clear()
        tasks.addAll(storedTasks)
    }

    private suspend fun processTask(taskId: String) {
        val taskIndex = tasks.indexOfFirst { it.taskId == taskId }
        if (taskIndex < 0) return

        val task = tasks[taskIndex]

        if (task.status == TranscribeTaskStatus.COMPLETED || task.status == TranscribeTaskStatus.PROCESSING) {
            return
        }

        if (task.retryCount >= MAX_RETRY_COUNT && task.status == TranscribeTaskStatus.FAILED) {
            return
        }

        try {
            val uploadingTask = task.copy(
                status = TranscribeTaskStatus.UPLOADING,
                updatedAt = System.currentTimeMillis()
            )
            tasks[taskIndex] = uploadingTask
            LocalStorage.updateTranscribeTask(uploadingTask)
            updateCounts()

            val result = apiClient.voice.submitTranscribeTask(
                fileBase64 = task.audioBase64,
                fileName = task.fileName,
                format = task.format,
                caseId = task.caseId,
                recordId = task.recordId
            )

            val processingTask = uploadingTask.copy(
                status = TranscribeTaskStatus.PROCESSING,
                remoteTaskId = result.taskId,
                updatedAt = System.currentTimeMillis()
            )
            tasks[tasks.indexOfFirst { it.taskId == taskId }] = processingTask
            LocalStorage.updateTranscribeTask(processingTask)
            updateCounts()

            pollTaskStatus(taskId)
        } catch (e: Exception) {
            val failedTask = task.copy(
                status = TranscribeTaskStatus.FAILED,
                errorMessage = e.message ?: "上传失败",
                updatedAt = System.currentTimeMillis()
            )
            val idx = tasks.indexOfFirst { it.taskId == taskId }
            if (idx >= 0) {
                tasks[idx] = failedTask
                LocalStorage.updateTranscribeTask(failedTask)
            }
            updateCounts()
        }
    }

    private suspend fun pollTaskStatus(taskId: String) {
        val taskIndex = tasks.indexOfFirst { it.taskId == taskId }
        if (taskIndex < 0) return

        val task = tasks[taskIndex]
        val remoteTaskId = task.remoteTaskId ?: return

        if (task.status == TranscribeTaskStatus.COMPLETED) return

        try {
            val result = apiClient.voice.getTranscribeTask(remoteTaskId)

            val finalStatus = result.statusCode.ifBlank {
                when (result.status) {
                    "3", "completed", "Completed", "已完成" -> "completed"
                    "4", "failed", "Failed", "失败" -> "failed"
                    "1", "2", "processing", "Processing", "Queuing", "转写中", "处理中" -> "processing"
                    "0", "pending", "Pending", "等待中", "排队中" -> "pending"
                    "5", "canceled", "Canceled", "已取消" -> "canceled"
                    else -> result.status.lowercase()
                }
            }

            when (finalStatus) {
                "completed" -> {
                    val completedTask = task.copy(
                        status = TranscribeTaskStatus.COMPLETED,
                        transcriptText = result.transcriptText,
                        updatedAt = System.currentTimeMillis()
                    )
                    tasks[taskIndex] = completedTask
                    LocalStorage.updateTranscribeTask(completedTask)
                    updateCounts()

                    onTaskCompleted(completedTask, result)
                }
                "failed", "canceled" -> {
                    val failedTask = task.copy(
                        status = TranscribeTaskStatus.FAILED,
                        errorMessage = result.errorMsg ?: result.errorMessage ?: "转写失败",
                        updatedAt = System.currentTimeMillis()
                    )
                    tasks[taskIndex] = failedTask
                    LocalStorage.updateTranscribeTask(failedTask)
                    updateCounts()
                }
                else -> {
                }
            }
        } catch (e: Exception) {
        }
    }

    private suspend fun onTaskCompleted(task: OfflineTranscribeTask, result: VoiceTranscribeTaskResult) {
        if (task.caseId != null && task.recordId != null && result.transcriptText != null) {
            try {
                apiClient.voice.saveTranscribeResult(
                    caseId = task.caseId,
                    recordId = task.recordId,
                    transcriptText = result.transcriptText,
                    transcribeTaskId = result.taskId,
                    duration = if (result.duration > 0) result.duration else null
                )
            } catch (e: Exception) {
                println("Failed to save transcribe result to mediation record: ${e.message}")
            }
        }
    }

    private fun startPolling() {
        if (pollJob != null) return

        pollJob = scope.launch {
            while (started) {
                delay(POLL_INTERVAL_MS)
                if (NetworkStatusService.isOnline) {
                    pollProcessingTasks()
                }
            }
        }
    }

    private fun stopPolling() {
        pollJob?.cancel()
        pollJob = null
    }

    private fun updateCounts() {
        _pendingCount.value = tasks.count {
            it.status == TranscribeTaskStatus.PENDING || it.status == TranscribeTaskStatus.FAILED
        }
        _processingCount.value = tasks.count {
            it.status == TranscribeTaskStatus.UPLOADING || it.status == TranscribeTaskStatus.PROCESSING
        }
    }

    companion object {
        private const val MAX_RETRY_COUNT = 3
        private const val POLL_INTERVAL_MS = 5000L
    }
}
