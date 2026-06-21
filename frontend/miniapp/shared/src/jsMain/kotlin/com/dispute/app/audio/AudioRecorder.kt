package com.dispute.app.audio

import kotlinx.browser.window
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import org.w3c.dom.Blob
import org.w3c.dom.MediaRecorder
import org.w3c.dom.MediaRecorderOptions
import org.w3c.dom.MediaStream
import org.w3c.files.FileReader
import kotlin.js.Promise

actual class AudioRecorder actual constructor() {
    private var mediaRecorder: dynamic = null
    private var audioChunks: MutableList<dynamic> = mutableListOf()
    private var recording = false
    private var onStartCallback: (() -> Unit)? = null
    private var onStopCallback: ((ByteArray, String, String) -> Unit)? = null
    private var onErrorCallback: ((String) -> Unit)? = null
    private var mediaStream: MediaStream? = null

    actual fun startRecording() {
        if (recording) return

        val wx = js("wx")
        if (wx != null && jsTypeOf(wx) != "undefined" &&
            wx.getRecorderManager != null) {
            startWeChatRecording(wx)
            return
        }

        startBrowserRecording()
    }

    private fun startWeChatRecording(wx: dynamic) {
        try {
            val recorderManager = wx.getRecorderManager()

            recorderManager.onStart = {
                recording = true
                onStartCallback?.invoke()
                Unit
            }

            recorderManager.onStop = { res: dynamic ->
                recording = false
                val tempFilePath = res.tempFilePath
                val fileSize = res.fileSize ?: 0
                val duration = res.duration ?: 0

                val fileSystemManager = wx.getFileSystemManager()
                fileSystemManager.readFile(
                    jsObject(
                        "filePath" to tempFilePath,
                        "encoding" to "base64"
                    ),
                    jsObject(
                        "success" to { readRes: dynamic ->
                            val base64Data = readRes.data as String
                            val bytes = base64ToByteArray(base64Data)
                            val fileName = "record_${System.currentTimeMillis()}.mp3"
                            onStopCallback?.invoke(bytes, fileName, "mp3")
                            Unit
                        },
                        "fail" to { err: dynamic ->
                            onErrorCallback?.invoke("读取录音文件失败: ${err.errMsg ?: "未知错误"}")
                            Unit
                        }
                    )
                )
                Unit
            }

            recorderManager.onError = { err: dynamic ->
                recording = false
                onErrorCallback?.invoke("录音出错: ${err.errMsg ?: "未知错误"}")
                Unit
            }

            recorderManager.start(
                jsObject(
                    "duration" to 60000,
                    "sampleRate" to 16000,
                    "numberOfChannels" to 1,
                    "encodeBitRate" to 48000,
                    "format" to "mp3"
                )
            )
        } catch (e: Exception) {
            onErrorCallback?.invoke("启动录音失败: ${e.message}")
        }
    }

    private fun startBrowserRecording() {
        try {
            val navigator = window.navigator
            val mediaDevices = navigator.asDynamic().mediaDevices

            if (mediaDevices == null || mediaDevices.getUserMedia == null) {
                onErrorCallback?.invoke("当前浏览器不支持录音功能")
                return
            }

            val constraints = jsObject(
                "audio" to true,
                "video" to false
            )

            val promise = mediaDevices.getUserMedia(constraints) as Promise<MediaStream>
            promise.then { stream ->
                mediaStream = stream
                audioChunks.clear()

                val options = js("({})")
                val recorder = try {
                    MediaRecorder(stream, options.unsafeCast<MediaRecorderOptions>())
                } catch (e: Exception) {
                    MediaRecorder(stream)
                }

                mediaRecorder = recorder

                recorder.ondataavailable = { event ->
                    if (event.data.size > 0) {
                        audioChunks.add(event.data)
                    }
                    Unit
                }

                recorder.onstop = {
                    val blob = Blob(audioChunks.toTypedArray(), jsObject("type" to "audio/webm"))
                    val reader = FileReader()

                    reader.onloadend = {
                        val result = reader.result
                        val base64 = (result as String).substringAfter(",")
                        val bytes = base64ToByteArray(base64)
                        val fileName = "record_${System.currentTimeMillis()}.webm"
                        onStopCallback?.invoke(bytes, fileName, "webm")
                        mediaStream?.getTracks()?.forEach { it.stop() }
                        mediaStream = null
                        Unit
                    }

                    reader.onerror = {
                        onErrorCallback?.invoke("读取音频文件失败")
                        Unit
                    }

                    reader.readAsDataURL(blob)
                    Unit
                }

                recorder.onerror = { event ->
                    recording = false
                    onErrorCallback?.invoke("录音出错")
                    Unit
                }

                recorder.start()
                recording = true
                onStartCallback?.invoke()
                Unit
            }.catch { error ->
                onErrorCallback?.invoke("获取录音权限失败: ${error.message ?: "请检查麦克风权限设置"}")
                Unit
            }
        } catch (e: Exception) {
            onErrorCallback?.invoke("启动录音失败: ${e.message}")
        }
    }

    actual fun stopRecording() {
        if (!recording) return

        val wx = js("wx")
        if (wx != null && jsTypeOf(wx) != "undefined" &&
            wx.getRecorderManager != null) {
            val recorderManager = wx.getRecorderManager()
            recorderManager.stop()
            return
        }

        try {
            mediaRecorder?.stop()
        } catch (e: Exception) {
            onErrorCallback?.invoke("停止录音失败: ${e.message}")
        }
    }

    actual fun isRecording(): Boolean = recording

    actual fun setOnRecordingStart(callback: () -> Unit) {
        onStartCallback = callback
    }

    actual fun setOnRecordingStop(callback: (audioData: ByteArray, fileName: String, format: String) -> Unit) {
        onStopCallback = callback
    }

    actual fun setOnError(callback: (message: String) -> Unit) {
        onErrorCallback = callback
    }

    actual fun release() {
        if (recording) {
            stopRecording()
        }
        mediaStream?.getTracks()?.forEach { it.stop() }
        mediaStream = null
        mediaRecorder = null
    }
}

actual fun isAudioRecordingSupported(): Boolean {
    val wx = js("wx")
    if (wx != null && jsTypeOf(wx) != "undefined" &&
        wx.getRecorderManager != null) {
        return true
    }

    return try {
        val navigator = window.navigator.asDynamic()
        navigator.mediaDevices != null &&
            navigator.mediaDevices.getUserMedia != null &&
            js("typeof MediaRecorder !== 'undefined'") as Boolean
    } catch (e: Exception) {
        false
    }
}

private fun base64ToByteArray(base64: String): ByteArray {
    val binary = js("atob(base64)") as String
    val bytes = ByteArray(binary.length)
    for (i in binary.indices) {
        bytes[i] = binary[i].code.toByte()
    }
    return bytes
}

private fun jsObject(vararg pairs: Pair<String, Any?>): dynamic {
    val obj = js("({})")
    pairs.forEach { (key, value) ->
        obj[key] = value
    }
    return obj
}

private fun jsTypeOf(value: dynamic): String {
    return js("typeof value") as String
}
