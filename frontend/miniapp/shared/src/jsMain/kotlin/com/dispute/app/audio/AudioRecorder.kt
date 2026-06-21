package com.dispute.app.audio

import kotlinx.browser.window
import org.w3c.dom.Blob
import org.w3c.dom.MediaRecorder
import org.w3c.dom.MediaRecorderOptions
import org.w3c.dom.MediaStream
import org.w3c.files.FileReader
import kotlin.js.Date
import kotlin.js.Promise

actual class AudioRecorder actual constructor() {
    private var mediaRecorder: dynamic = null
    private var audioChunks: MutableList<dynamic> = mutableListOf()
    private var recording = false
    private var onStartCallback: (() -> Unit)? = null
    private var onStopCallback: ((ByteArray, String, String) -> Unit)? = null
    private var onErrorCallback: ((String) -> Unit)? = null
    private var mediaStream: MediaStream? = null
    private var wxRecorderManager: dynamic = null

    actual fun startRecording() {
        if (recording) return

        if (detectWeChat()) {
            startWeChatRecording()
        } else {
            startBrowserRecording()
        }
    }

    private fun detectWeChat(): Boolean {
        return try {
            js("typeof wx !== 'undefined' && typeof wx.getRecorderManager === 'function'") as Boolean
        } catch (e: Exception) {
            false
        }
    }

    private fun startWeChatRecording() {
        try {
            val wx = js("wx")
            wxRecorderManager = wx.getRecorderManager()
            val self = this@AudioRecorder

            wxRecorderManager.onStart = {
                recording = true
                self.onStartCallback?.invoke()
                Unit
            }

            wxRecorderManager.onStop = { res: dynamic ->
                recording = false
                val tempFilePath = res.tempFilePath as String

                js("""
                    (function(wx, tempFilePath, onSuccess, onFail) {
                        try {
                            var fsManager = wx.getFileSystemManager();
                            fsManager.readFile({
                                filePath: tempFilePath,
                                encoding: 'base64',
                                success: function(readRes) {
                                    onSuccess(readRes.data);
                                },
                                fail: function(err) {
                                    onFail(err.errMsg || '读取录音文件失败');
                                }
                            });
                        } catch(e) {
                            onFail(e.message || '读取录音文件异常');
                        }
                    })
                """)(wx, tempFilePath,
                    { base64Data: String ->
                        val bytes = base64ToByteArray(base64Data)
                        val fileName = "record_${Date.now().toLong()}.mp3"
                        self.onStopCallback?.invoke(bytes, fileName, "mp3")
                        Unit
                    },
                    { errMsg: String ->
                        self.onErrorCallback?.invoke(errMsg)
                        Unit
                    }
                )
                Unit
            }

            wxRecorderManager.onError = { err: dynamic ->
                recording = false
                val errMsg = (err.errMsg as? String) ?: "录音出错"
                this@AudioRecorder.onErrorCallback?.invoke(errMsg)
                Unit
            }

            wxRecorderManager.start(js("{duration: 60000, sampleRate: 16000, numberOfChannels: 1, encodeBitRate: 48000, format: 'mp3'}"))
        } catch (e: Exception) {
            startBrowserRecording()
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

            val constraints = js("{audio: true, video: false}")

            val promise = mediaDevices.getUserMedia(constraints) as Promise<MediaStream>
            promise.then { stream ->
                mediaStream = stream
                audioChunks.clear()

                val recorder = try {
                    val options = js("{mimeType: 'audio/webm;codecs=opus'}")
                    MediaRecorder(stream, options.unsafeCast<MediaRecorderOptions>())
                } catch (e: Exception) {
                    try {
                        MediaRecorder(stream)
                    } catch (e2: Exception) {
                        onErrorCallback?.invoke("浏览器不支持MediaRecorder")
                        return@then
                    }
                }

                mediaRecorder = recorder

                recorder.ondataavailable = { event ->
                    if (event.data.size > 0) {
                        audioChunks.add(event.data)
                    }
                    Unit
                }

                recorder.onstop = {
                    if (audioChunks.isEmpty()) {
                        onErrorCallback?.invoke("未录制到音频数据")
                        return<Unit>
                    }
                    val blob = Blob(audioChunks.toTypedArray(), js("{type: 'audio/webm'}"))
                    val reader = FileReader()
                    val self = this@AudioRecorder

                    reader.onloadend = {
                        val result = reader.result
                        val base64 = (result as String).substringAfter(",")
                        val bytes = base64ToByteArray(base64)
                        val fileName = "record_${Date.now().toLong()}.webm"
                        self.onStopCallback?.invoke(bytes, fileName, "webm")
                        mediaStream?.getTracks()?.forEach { it.stop() }
                        mediaStream = null
                        Unit
                    }

                    reader.onerror = {
                        self.onErrorCallback?.invoke("读取音频文件失败")
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
                val msg = error?.asDynamic()?.message as? String ?: "请检查麦克风权限设置"
                onErrorCallback?.invoke("获取录音权限失败: $msg")
                Unit
            }
        } catch (e: Exception) {
            onErrorCallback?.invoke("启动录音失败: ${e.message}")
        }
    }

    actual fun stopRecording() {
        if (!recording) return

        if (detectWeChat() && wxRecorderManager != null) {
            try {
                wxRecorderManager.stop()
            } catch (e: Exception) {
                onErrorCallback?.invoke("停止录音失败: ${e.message}")
            }
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
        wxRecorderManager = null
    }
}

actual fun isAudioRecordingSupported(): Boolean {
    val hasWx = try {
        js("typeof wx !== 'undefined' && typeof wx.getRecorderManager === 'function'") as Boolean
    } catch (e: Exception) {
        false
    }
    if (hasWx) return true

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
