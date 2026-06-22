package com.dispute.app.components

import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.Button
import androidx.compose.material3.ButtonDefaults
import androidx.compose.material3.LinearProgressIndicator
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.DisposableEffect
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableLongStateOf
import androidx.compose.runtime.mutableStateListOf
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.dispute.app.LocalAppState
import com.dispute.app.audio.AudioRecorder
import com.dispute.app.audio.isAudioRecordingSupported
import com.dispute.app.audio.toBase64
import com.dispute.app.platform.NetworkStatusService
import com.dispute.app.storage.OfflineTranscribeTask
import com.dispute.app.storage.TranscribeTaskStatus
import com.dispute.app.utils.formatTimestamp
import kotlinx.coroutines.delay
import kotlinx.coroutines.launch
import kotlin.math.absoluteValue
import kotlin.math.sin

@Composable
fun VoiceRecordSection(
    caseId: String? = null,
    recordId: String? = null,
    onTranscribeComplete: ((String) -> Unit)? = null,
    modifier: Modifier = Modifier
) {
    val appState = LocalAppState.current
    val isOnline by NetworkStatusService.isOnlineState

    var isRecording by remember { mutableStateOf(false) }
    var recordingDuration by remember { mutableLongStateOf(0L) }
    var audioRecorder by remember { mutableStateOf<AudioRecorder?>(null) }
    var errorMessage by remember { mutableStateOf<String?>(null) }

    val transcriptList = remember { mutableStateListOf<OfflineTranscribeTask>() }

    val supported = remember { isAudioRecordingSupported() }

    LaunchedEffect(Unit) {
        audioRecorder = AudioRecorder().apply {
            setOnRecordingStart {
                isRecording = true
                recordingDuration = 0L
                errorMessage = null
            }
            setOnRecordingStop { audioData, fileName, format ->
                isRecording = false
                appState.appScope.launch {
                    try {
                        val base64 = audioData.toBase64()
                        val taskId = "task_${System.currentTimeMillis()}"
                        val task = OfflineTranscribeTask(
                            taskId = taskId,
                            status = TranscribeTaskStatus.PENDING,
                            audioBase64 = base64,
                            fileName = fileName,
                            format = format,
                            sizeBytes = audioData.size.toLong(),
                            durationMs = recordingDuration,
                            caseId = caseId,
                            recordId = recordId,
                            createdAt = System.currentTimeMillis(),
                            updatedAt = System.currentTimeMillis()
                        )

                        if (appState.isTranscribeManagerInitialized()) {
                            appState.transcribeManager.submitTask(task)
                        }
                        transcriptList.add(0, task)

                        if (!isOnline) {
                            appState.showToast("录音已保存，网络恢复后自动转写")
                        }
                    } catch (e: Exception) {
                        errorMessage = "处理录音失败: ${e.message}"
                    }
                }
                Unit
            }
            setOnError { message ->
                isRecording = false
                errorMessage = message
            }
        }
    }

    LaunchedEffect(isRecording) {
        if (isRecording) {
            while (isRecording) {
                delay(1000L)
                recordingDuration += 1000L
            }
        }
    }

    LaunchedEffect(Unit) {
        if (caseId != null && appState.isTranscribeManagerInitialized()) {
            val tasks = appState.transcribeManager.getTasksByCase(caseId)
            transcriptList.clear()
            transcriptList.addAll(tasks)
        }
    }

    DisposableEffect(Unit) {
        onDispose {
            audioRecorder?.release()
        }
    }

    AppCard(title = "语音记录", modifier = modifier) {
        Column(
            verticalArrangement = Arrangement.spacedBy(12.dp)
        ) {
            if (!supported) {
                Text(
                    text = "当前设备不支持录音功能",
                    color = MaterialTheme.colorScheme.error,
                    style = MaterialTheme.typography.bodyMedium
                )
            } else {
                Row(
                    modifier = Modifier.fillMaxWidth(),
                    verticalAlignment = Alignment.CenterVertically,
                    horizontalArrangement = Arrangement.spacedBy(12.dp)
                ) {
                    RecordButton(
                        isRecording = isRecording,
                        onClick = {
                            if (isRecording) {
                                audioRecorder?.stopRecording()
                            } else {
                                errorMessage = null
                                audioRecorder?.startRecording()
                            }
                        }
                    )

                    Column(modifier = Modifier.weight(1f)) {
                        Text(
                            text = if (isRecording) "录音中..." else "点击开始录音",
                            style = MaterialTheme.typography.titleMedium,
                            fontWeight = FontWeight.Medium
                        )
                        Spacer(modifier = Modifier.height(4.dp))
                        Text(
                            text = formatDuration(recordingDuration),
                            style = MaterialTheme.typography.bodyLarge,
                            color = if (isRecording) MaterialTheme.colorScheme.primary
                            else MaterialTheme.colorScheme.onSurfaceVariant
                        )
                    }

                    if (isRecording) {
                        RecordingWaveform()
                    }
                }

                if (!isOnline) {
                    Row(
                        modifier = Modifier
                            .fillMaxWidth()
                            .background(
                                Color(0xFFFFF7ED),
                                RoundedCornerShape(8.dp)
                            )
                            .padding(10.dp),
                        verticalAlignment = Alignment.CenterVertically
                    ) {
                        Text("📡", fontSize = 16.sp)
                        Spacer(modifier = Modifier.width(8.dp))
                        Text(
                            text = "当前离线，录音将保存到本地，网络恢复后自动转写",
                            style = MaterialTheme.typography.bodySmall,
                            color = Color(0xFFEA580C)
                        )
                    }
                }

                errorMessage?.let { msg ->
                    Text(
                        text = msg,
                        color = MaterialTheme.colorScheme.error,
                        style = MaterialTheme.typography.bodySmall
                    )
                }
            }

            if (transcriptList.isNotEmpty()) {
                Spacer(modifier = Modifier.height(8.dp))
                Text(
                    text = "转写记录（${transcriptList.size}条）",
                    style = MaterialTheme.typography.titleSmall,
                    fontWeight = FontWeight.SemiBold
                )

                transcriptList.forEach { task ->
                    TranscriptItem(
                        task = task,
                        onRetry = {
                            appState.appScope.launch {
                                if (appState.isTranscribeManagerInitialized()) {
                                    appState.transcribeManager.retryTask(task.taskId)
                                    val updated = appState.transcribeManager.getTask(task.taskId)
                                    updated?.let {
                                        val idx = transcriptList.indexOfFirst { it.taskId == task.taskId }
                                        if (idx >= 0) {
                                            transcriptList[idx] = it
                                        }
                                    }
                                }
                            }
                        },
                        onUseText = { text ->
                            onTranscribeComplete?.invoke(text)
                        }
                    )
                }
            }
        }
    }
}

@Composable
private fun RecordButton(isRecording: Boolean, onClick: () -> Unit) {
    Box(
        modifier = Modifier
            .size(64.dp)
            .background(
                color = if (isRecording) Color(0xFFEF4444) else MaterialTheme.colorScheme.primary,
                shape = CircleShape
            )
            .clickable(onClick = onClick),
        contentAlignment = Alignment.Center
    ) {
        if (isRecording) {
            Box(
                modifier = Modifier
                    .size(20.dp)
                    .background(Color.White, RoundedCornerShape(4.dp))
            )
        } else {
            Box(
                modifier = Modifier
                    .size(24.dp)
                    .background(Color.White, CircleShape)
            )
        }
    }
}

@Composable
private fun RecordingWaveform() {
    val bars = 5
    Row(
        horizontalArrangement = Arrangement.spacedBy(3.dp),
        verticalAlignment = Alignment.CenterVertically
    ) {
        for (i in 0 until bars) {
            var height by remember { mutableStateOf(8f) }

            LaunchedEffect(Unit) {
                while (true) {
                    delay(100L + (i * 30L))
                    height = 12f + sin((System.currentTimeMillis() / 200.0 + i).toFloat())
                        .absoluteValue * 16f
                }
            }

            Box(
                modifier = Modifier
                    .width(3.dp)
                    .height(height.dp)
                    .background(
                        MaterialTheme.colorScheme.primary,
                        RoundedCornerShape(2.dp)
                    )
            )
        }
    }
}

@Composable
private fun TranscriptItem(
    task: OfflineTranscribeTask,
    onRetry: () -> Unit,
    onUseText: (String) -> Unit
) {
    Column(
        modifier = Modifier
            .fillMaxWidth()
            .background(
                MaterialTheme.colorScheme.surfaceVariant.copy(alpha = 0.4f),
                RoundedCornerShape(10.dp)
            )
            .padding(12.dp),
        verticalArrangement = Arrangement.spacedBy(8.dp)
    ) {
        Row(
            modifier = Modifier.fillMaxWidth(),
            verticalAlignment = Alignment.CenterVertically,
            horizontalArrangement = Arrangement.SpaceBetween
        ) {
            Row(verticalAlignment = Alignment.CenterVertically) {
                Text("🎙️", fontSize = 16.sp)
                Spacer(modifier = Modifier.width(6.dp))
                Text(
                    text = formatDuration(task.durationMs),
                    style = MaterialTheme.typography.bodyMedium,
                    fontWeight = FontWeight.Medium
                )
            }
            StatusBadge(status = task.status)
        }

        when (task.status) {
            TranscribeTaskStatus.PENDING -> {
                Row(
                    verticalAlignment = Alignment.CenterVertically,
                    horizontalArrangement = Arrangement.spacedBy(8.dp)
                ) {
                    androidx.compose.material3.CircularProgressIndicator(
                        modifier = Modifier.size(14.dp),
                        strokeWidth = 2.dp
                    )
                    Text(
                        text = "等待转写...",
                        style = MaterialTheme.typography.bodySmall,
                        color = MaterialTheme.colorScheme.onSurfaceVariant
                    )
                }
            }
            TranscribeTaskStatus.UPLOADING -> {
                Column(verticalArrangement = Arrangement.spacedBy(4.dp)) {
                    LinearProgressIndicator(
                        modifier = Modifier.fillMaxWidth(),
                    )
                    Text(
                        text = "上传中...",
                        style = MaterialTheme.typography.bodySmall,
                        color = MaterialTheme.colorScheme.onSurfaceVariant
                    )
                }
            }
            TranscribeTaskStatus.PROCESSING -> {
                Row(
                    verticalAlignment = Alignment.CenterVertically,
                    horizontalArrangement = Arrangement.spacedBy(8.dp)
                ) {
                    androidx.compose.material3.CircularProgressIndicator(
                        modifier = Modifier.size(14.dp),
                        strokeWidth = 2.dp
                    )
                    Text(
                        text = "转写中，请稍候...",
                        style = MaterialTheme.typography.bodySmall,
                        color = MaterialTheme.colorScheme.onSurfaceVariant
                    )
                }
            }
            TranscribeTaskStatus.COMPLETED -> {
                Text(
                    text = task.transcriptText ?: "（无文本）",
                    style = MaterialTheme.typography.bodyMedium,
                    lineHeight = 18.sp
                )
                Button(
                    onClick = { task.transcriptText?.let { onUseText(it) } },
                    modifier = Modifier.fillMaxWidth().height(36.dp),
                    shape = RoundedCornerShape(8.dp),
                    colors = ButtonDefaults.buttonColors(
                        containerColor = MaterialTheme.colorScheme.primaryContainer,
                        contentColor = MaterialTheme.colorScheme.onPrimaryContainer
                    )
                ) {
                    Text("使用此文本", fontSize = 12.sp, fontWeight = FontWeight.Medium)
                }
            }
            TranscribeTaskStatus.FAILED -> {
                Column(verticalArrangement = Arrangement.spacedBy(6.dp)) {
                    Text(
                        text = task.errorMessage ?: "转写失败",
                        style = MaterialTheme.typography.bodySmall,
                        color = MaterialTheme.colorScheme.error
                    )
                    if (task.retryCount < 3) {
                        Button(
                            onClick = onRetry,
                            modifier = Modifier.height(32.dp),
                            shape = RoundedCornerShape(8.dp),
                            colors = ButtonDefaults.buttonColors(
                                containerColor = MaterialTheme.colorScheme.errorContainer,
                                contentColor = MaterialTheme.colorScheme.onErrorContainer
                            )
                        ) {
                            Text("重试", fontSize = 12.sp)
                        }
                    }
                }
            }
        }

        Text(
            text = formatTimestamp(task.createdAt),
            style = MaterialTheme.typography.labelSmall,
            color = Color(0xFF9CA3AF)
        )
    }
}

@Composable
private fun StatusBadge(status: TranscribeTaskStatus) {
    val (bgColor, textColor, text) = when (status) {
        TranscribeTaskStatus.PENDING -> Triple(
            Color(0xFFFEF3C7),
            Color(0xFFD97706),
            "等待中"
        )
        TranscribeTaskStatus.UPLOADING -> Triple(
            Color(0xFFDBEAFE),
            Color(0xFF2563EB),
            "上传中"
        )
        TranscribeTaskStatus.PROCESSING -> Triple(
            Color(0xFFE0E7FF),
            Color(0xFF4F46E5),
            "转写中"
        )
        TranscribeTaskStatus.COMPLETED -> Triple(
            Color(0xFFDCFCE7),
            Color(0xFF15803D),
            "已完成"
        )
        TranscribeTaskStatus.FAILED -> Triple(
            Color(0xFFFEE2E2),
            Color(0xFFDC2626),
            "失败"
        )
    }

    Box(
        modifier = Modifier
            .background(bgColor, RoundedCornerShape(6.dp))
            .padding(horizontal = 8.dp, vertical = 3.dp),
        contentAlignment = Alignment.Center
    ) {
        Text(
            text = text,
            fontSize = 10.sp,
            fontWeight = FontWeight.Medium,
            color = textColor
        )
    }
}

private fun formatDuration(durationMs: Long): String {
    val totalSeconds = durationMs / 1000
    val minutes = totalSeconds / 60
    val seconds = totalSeconds % 60
    val minStr = if (minutes < 10) "0$minutes" else "$minutes"
    val secStr = if (seconds < 10) "0$seconds" else "$seconds"
    return "$minStr:$secStr"
}
