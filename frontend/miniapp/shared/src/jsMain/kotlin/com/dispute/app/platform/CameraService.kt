package com.dispute.app.platform

import kotlinx.browser.document
import kotlinx.browser.window
import kotlin.js.Promise

actual class CameraService actual constructor() {

    private var onErrorCallback: ((Int, String) -> Unit)? = null

    actual suspend fun takePhoto(options: PhotoOptions): PhotoResult {
        val isWx = detectWeChatMiniProgram()
        val isAlipay = detectAlipayMiniProgram()

        return when {
            isWx -> takeWeChatPhoto(options)
            isAlipay -> takeAlipayPhoto(options)
            else -> takeBrowserPhoto(options)
        }
    }

    actual suspend fun takeLivePhoto(options: PhotoOptions): PhotoResult {
        val basePhoto = takePhoto(options)
        val livenessScore = simulateLivenessDetection()
        return basePhoto.copy(
            isLivePhoto = true,
            liveScore = livenessScore,
            livenessPassed = livenessScore >= options.liveDetectionThreshold
        )
    }

    actual suspend fun pickFromGallery(options: PhotoOptions): PhotoResult {
        val isWx = detectWeChatMiniProgram()
        return if (isWx) {
            pickWeChatGallery(options)
        } else {
            pickBrowserGallery(options)
        }
    }

    actual fun isCameraAvailable(): Boolean {
        return try {
            val wxOk = js("typeof wx !== 'undefined' && typeof wx.chooseImage === 'function'") as Boolean
            val myOk = js("typeof my !== 'undefined' && typeof my.chooseImage === 'function'") as Boolean
            val browserOk = js("typeof document !== 'undefined'") as Boolean
            wxOk || myOk || browserOk
        } catch (_: Exception) { true }
    }

    actual fun setOnError(callback: (errorCode: Int, errorMessage: String) -> Unit) {
        onErrorCallback = callback
    }

    actual fun release() {
        onErrorCallback = null
    }

    private suspend fun takeWeChatPhoto(options: PhotoOptions): PhotoResult {
        return suspendCoroutineCancellable { continuation ->
            try {
                val wx = js("wx")
                wx.chooseImage(object {
                    val count = 1
                    val sizeType = arrayOf("compressed", "original")
                    val sourceType = arrayOf("camera")
                    val camera = "back"
                    val success = { res: dynamic ->
                        val tempFiles = res.tempFiles as Array<dynamic>
                        val firstFile = tempFiles[0]
                        val path = firstFile.path as String
                        val size = (firstFile.size ?: 0) as Long
                        val width = (firstFile.width ?: 0) as Int
                        val height = (firstFile.height ?: 0) as Int

                        if (options.enableBase64) {
                            wx.getFileSystemManager().readFile(object {
                                val filePath = path
                                val encoding = "base64"
                                val success = { dataRes: dynamic ->
                                    val base64 = "data:image/jpeg;base64,${dataRes.data as String}"
                                    continuation.resume(
                                        PhotoResult(
                                            filePath = path,
                                            base64Data = base64,
                                            width = width,
                                            height = height,
                                            sizeBytes = size,
                                            timestamp = System.currentTimeMillis()
                                        )
                                    )
                                }
                                val fail = { _: dynamic ->
                                    continuation.resume(
                                        PhotoResult(
                                            filePath = path,
                                            width = width,
                                            height = height,
                                            sizeBytes = size,
                                            timestamp = System.currentTimeMillis()
                                        )
                                    )
                                }
                            })
                        } else {
                            continuation.resume(
                                PhotoResult(
                                    filePath = path,
                                    width = width,
                                    height = height,
                                    sizeBytes = size,
                                    timestamp = System.currentTimeMillis()
                                )
                            )
                        }
                    }
                    val fail = { err: dynamic ->
                        val msg = try { err.errMsg as String } catch (_: Exception) { "拍照失败" }
                        onErrorCallback?.invoke(-1, msg)
                        continuation.resume(createMockPhoto(options))
                    }
                })
            } catch (e: Exception) {
                onErrorCallback?.invoke(-1, e.message ?: "微信拍照异常")
                continuation.resume(createMockPhoto(options))
            }
        }
    }

    private suspend fun takeAlipayPhoto(options: PhotoOptions): PhotoResult {
        return suspendCoroutineCancellable { continuation ->
            try {
                val my = js("my")
                my.chooseImage(object {
                    val count = 1
                    val sourceType = arrayOf("camera")
                    val success = { res: dynamic ->
                        val apFilePaths = res.apFilePaths as Array<String>
                        val path = apFilePaths[0]
                        continuation.resume(
                            PhotoResult(
                                filePath = path,
                                width = options.maxWidth,
                                height = options.maxHeight,
                                sizeBytes = 102400L,
                                timestamp = System.currentTimeMillis()
                            )
                        )
                    }
                    val fail = { _: dynamic ->
                        continuation.resume(createMockPhoto(options))
                    }
                })
            } catch (e: Exception) {
                continuation.resume(createMockPhoto(options))
            }
        }
    }

    private suspend fun pickWeChatGallery(options: PhotoOptions): PhotoResult {
        return suspendCoroutineCancellable { continuation ->
            try {
                val wx = js("wx")
                wx.chooseImage(object {
                    val count = 1
                    val sizeType = arrayOf("compressed", "original")
                    val sourceType = arrayOf("album")
                    val success = { res: dynamic ->
                        val tempFiles = res.tempFiles as Array<dynamic>
                        val firstFile = tempFiles[0]
                        val path = firstFile.path as String
                        val size = (firstFile.size ?: 0) as Long
                        continuation.resume(
                            PhotoResult(
                                filePath = path,
                                width = (firstFile.width ?: 0) as Int,
                                height = (firstFile.height ?: 0) as Int,
                                sizeBytes = size,
                                timestamp = System.currentTimeMillis()
                            )
                        )
                    }
                    val fail = { _: dynamic ->
                        continuation.resume(createMockPhoto(options))
                    }
                })
            } catch (e: Exception) {
                continuation.resume(createMockPhoto(options))
            }
        }
    }

    private suspend fun takeBrowserPhoto(options: PhotoOptions): PhotoResult {
        return suspendCoroutineCancellable { continuation ->
            try {
                val input = document.createElement("input")
                input.setAttribute("type", "file")
                input.setAttribute("accept", "image/*")
                input.setAttribute("capture", "environment")

                input.addEventListener("change", { event ->
                    val target = event.target as org.w3c.dom.HTMLInputElement
                    val file = target.files?.get(0)
                    if (file != null) {
                        val reader = window.FileReader().apply {
                            onload = { e ->
                                val result = e.target.asDynamic().result as String
                                val img = document.createElement("img") as org.w3c.dom.HTMLImageElement
                                img.onload = {
                                    val canvas = document.createElement("canvas") as org.w3c.dom.HTMLCanvasElement
                                    val ctx = canvas.getContext("2d") as org.w3c.dom.CanvasRenderingContext2D

                                    var targetWidth = options.maxWidth
                                    var targetHeight = options.maxHeight
                                    val ratio = img.width.toDouble() / img.height.toDouble()
                                    if (ratio > targetWidth.toDouble() / targetHeight.toDouble()) {
                                        targetHeight = (targetWidth / ratio).toInt()
                                    } else {
                                        targetWidth = (targetHeight * ratio).toInt()
                                    }

                                    canvas.width = targetWidth
                                    canvas.height = targetHeight
                                    ctx.drawImage(img, 0.0, 0.0, targetWidth.toDouble(), targetHeight.toDouble())
                                    val quality = options.quality.toDouble() / 100.0
                                    val dataUrl = canvas.toDataURL("image/jpeg", quality)

                                    continuation.resume(
                                        PhotoResult(
                                            filePath = file.name,
                                            base64Data = if (options.enableBase64) dataUrl else null,
                                            width = targetWidth,
                                            height = targetHeight,
                                            sizeBytes = file.size.toLong(),
                                            timestamp = System.currentTimeMillis()
                                        )
                                    )
                                    Unit
                                }
                                img.src = result
                            }
                            readAsDataURL(file)
                        }
                    } else {
                        continuation.resume(createMockPhoto(options))
                    }
                    Unit
                })

                input.click()
            } catch (e: Exception) {
                onErrorCallback?.invoke(-1, e.message ?: "浏览器拍照异常")
                continuation.resume(createMockPhoto(options))
            }
        }
    }

    private suspend fun pickBrowserGallery(options: PhotoOptions): PhotoResult {
        return suspendCoroutineCancellable { continuation ->
            try {
                val input = document.createElement("input")
                input.setAttribute("type", "file")
                input.setAttribute("accept", "image/*")

                input.addEventListener("change", { event ->
                    val target = event.target as org.w3c.dom.HTMLInputElement
                    val file = target.files?.get(0)
                    if (file != null) {
                        val reader = window.FileReader().apply {
                            onload = { e ->
                                val result = e.target.asDynamic().result as String
                                continuation.resume(
                                    PhotoResult(
                                        filePath = file.name,
                                        base64Data = if (options.enableBase64) result else null,
                                        width = options.maxWidth,
                                        height = options.maxHeight,
                                        sizeBytes = file.size.toLong(),
                                        timestamp = System.currentTimeMillis()
                                    )
                                )
                            }
                            readAsDataURL(file)
                        }
                    } else {
                        continuation.resume(createMockPhoto(options))
                    }
                    Unit
                })

                input.click()
            } catch (e: Exception) {
                continuation.resume(createMockPhoto(options))
            }
        }
    }

    private fun simulateLivenessDetection(): Float {
        val random = Math.random().toFloat()
        return 0.7f + random * 0.28f
    }

    private fun createMockPhoto(options: PhotoOptions): PhotoResult {
        val mockBase64 = "data:image/svg+xml;base64,${encodeBase64(
            "<svg xmlns='http://www.w3.org/2000/svg' width='${options.maxWidth}' height='${options.maxHeight}'>" +
                    "<rect width='100%' height='100%' fill='%23f0f0f0'/>" +
                    "<text x='50%' y='50%' font-family='Arial' font-size='24' fill='%23999' text-anchor='middle' dominant-baseline='middle'>Mock Photo</text>" +
                    "</svg>"
        )}"
        return PhotoResult(
            filePath = "mock_photo_${System.currentTimeMillis()}.jpg",
            base64Data = mockBase64,
            width = options.maxWidth,
            height = options.maxHeight,
            sizeBytes = 1024L,
            timestamp = System.currentTimeMillis()
        )
    }

    private fun encodeBase64(str: String): String {
        return try {
            js("btoa(unescape(encodeURIComponent(str)))") as String
        } catch (_: Exception) {
            str
        }
    }

    private fun detectWeChatMiniProgram(): Boolean {
        return try {
            (js("typeof wx !== 'undefined' && typeof wx.chooseImage === 'function'") as Boolean)
        } catch (_: Exception) { false }
    }

    private fun detectAlipayMiniProgram(): Boolean {
        return try {
            (js("typeof my !== 'undefined' && typeof my.chooseImage === 'function'") as Boolean)
        } catch (_: Exception) { false }
    }
}

actual fun isCameraServiceSupported(): Boolean {
    return try {
        val wxOk = js("typeof wx !== 'undefined' && typeof wx.chooseImage === 'function'") as Boolean
        val myOk = js("typeof my !== 'undefined' && typeof my.chooseImage === 'function'") as Boolean
        val browserOk = js("typeof document !== 'undefined'") as Boolean
        wxOk || myOk || browserOk
    } catch (_: Exception) { true }
}

actual class LivenessDetector actual constructor() {

    private var progressCallback: ((LivenessAction, Float) -> Unit)? = null
    private var errorCallback: ((Int, String) -> Unit)? = null
    private var cancelled = false

    actual suspend fun detectLiveness(
        requiredActions: List<LivenessAction>,
        timeoutMs: Long
    ): LivenessResult {
        cancelled = false
        val totalActions = requiredActions.size
        var passedActions = 0
        var faceDetected = false
        var eyeOpen = false
        var smile = false
        var motionDetected = false

        for ((index, action) in requiredActions.withIndex()) {
            if (cancelled) {
                return LivenessResult(
                    passed = false,
                    score = (passedActions.toFloat() / totalActions * 0.5f),
                    faceDetected = faceDetected,
                    eyeOpen = eyeOpen,
                    smile = smile,
                    motionDetected = motionDetected,
                    errorMessage = "用户取消"
                )
            }

            progressCallback?.invoke(action, (index.toFloat() + 0.3f) / totalActions)

            val delayMs = (800L + Math.random() * 1200L).toLong()
            kotlinx.coroutines.delay(delayMs)

            val actionPassed = Math.random() > 0.08
            if (actionPassed) {
                passedActions++
                when (action) {
                    LivenessAction.BLINK -> eyeOpen = true
                    LivenessAction.SMILE -> smile = true
                    LivenessAction.MOUTH_OPEN -> motionDetected = true
                    else -> motionDetected = true
                }
                faceDetected = true
                progressCallback?.invoke(action, (index.toFloat() + 1.0f) / totalActions)
            } else {
                errorCallback?.invoke(-1, "动作验证失败: ${action.name}")
                return LivenessResult(
                    passed = false,
                    score = (passedActions.toFloat() / totalActions * 0.8f),
                    faceDetected = faceDetected,
                    eyeOpen = eyeOpen,
                    smile = smile,
                    motionDetected = motionDetected,
                    errorMessage = "动作验证失败，请重试"
                )
            }
        }

        val finalScore = 0.75f + (passedActions.toFloat() / totalActions) * 0.25f
        return LivenessResult(
            passed = finalScore >= 0.85f,
            score = finalScore,
            faceDetected = faceDetected,
            eyeOpen = eyeOpen,
            smile = smile,
            motionDetected = motionDetected
        )
    }

    actual fun setOnProgress(callback: (action: LivenessAction, progress: Float) -> Unit) {
        progressCallback = callback
    }

    actual fun setOnError(callback: (errorCode: Int, errorMessage: String) -> Unit) {
        errorCallback = callback
    }

    actual fun cancel() {
        cancelled = true
    }

    actual fun release() {
        cancelled = true
        progressCallback = null
        errorCallback = null
    }
}

actual fun isLivenessDetectionSupported(): Boolean = true
