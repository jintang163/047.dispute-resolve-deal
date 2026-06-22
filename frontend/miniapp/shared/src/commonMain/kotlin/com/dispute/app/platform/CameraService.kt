package com.dispute.app.platform

expect class CameraService() {

    suspend fun takePhoto(options: PhotoOptions = PhotoOptions()): PhotoResult

    suspend fun takeLivePhoto(options: PhotoOptions = PhotoOptions()): PhotoResult

    suspend fun pickFromGallery(options: PhotoOptions = PhotoOptions()): PhotoResult

    fun isCameraAvailable(): Boolean

    fun setOnError(callback: (errorCode: Int, errorMessage: String) -> Unit)

    fun release()
}

data class PhotoOptions(
    val maxWidth: Int = 1920,
    val maxHeight: Int = 1080,
    val quality: Int = 85,
    val enableBase64: Boolean = true,
    val requireLiveDetection: Boolean = false,
    val liveDetectionThreshold: Float = 0.85f
)

data class PhotoResult(
    val filePath: String,
    val base64Data: String? = null,
    val width: Int = 0,
    val height: Int = 0,
    val sizeBytes: Long = 0L,
    val mimeType: String = "image/jpeg",
    val latitude: Double? = null,
    val longitude: Double? = null,
    val timestamp: Long = System.currentTimeMillis(),
    val isLivePhoto: Boolean = false,
    val liveScore: Float = 0f,
    val livenessPassed: Boolean = false
)

expect fun isCameraServiceSupported(): Boolean

data class LivenessResult(
    val passed: Boolean,
    val score: Float,
    val faceDetected: Boolean,
    val eyeOpen: Boolean,
    val smile: Boolean,
    val motionDetected: Boolean,
    val errorMessage: String? = null
)

expect class LivenessDetector() {

    suspend fun detectLiveness(
        requiredActions: List<LivenessAction> = listOf(
            LivenessAction.BLINK,
            LivenessAction.SMILE
        ),
        timeoutMs: Long = 15000L
    ): LivenessResult

    fun setOnProgress(callback: (action: LivenessAction, progress: Float) -> Unit)

    fun setOnError(callback: (errorCode: Int, errorMessage: String) -> Unit)

    fun cancel()

    fun release()
}

enum class LivenessAction {
    BLINK,
    SMILE,
    NOD,
    MOUTH_OPEN,
    TURN_LEFT,
    TURN_RIGHT,
    RANDOM
}

expect fun isLivenessDetectionSupported(): Boolean

class PhotoUploadHelper() {

    private var uploads = mutableListOf<PhotoResult>()

    fun addPhoto(photo: PhotoResult) {
        uploads.add(photo)
    }

    fun getPhotos(): List<PhotoResult> = uploads.toList()

    fun clear() {
        uploads.clear()
    }

    fun getBase64List(): List<String> {
        return uploads.mapNotNull { it.base64Data }
    }

    fun getFilePathList(): List<String> {
        return uploads.map { it.filePath }
    }

    fun joinBase64Urls(): String {
        return getBase64List().joinToString(",")
    }

    fun totalSize(): Long {
        return uploads.sumOf { it.sizeBytes }
    }

    fun count(): Int = uploads.size

    fun removeAt(index: Int) {
        if (index in uploads.indices) {
            uploads.removeAt(index)
        }
    }
}
