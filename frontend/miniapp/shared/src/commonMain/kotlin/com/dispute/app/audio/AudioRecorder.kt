package com.dispute.app.audio

expect class AudioRecorder() {
    fun startRecording()
    fun stopRecording()
    fun isRecording(): Boolean
    fun setOnRecordingStart(callback: () -> Unit)
    fun setOnRecordingStop(callback: (audioData: ByteArray, fileName: String, format: String) -> Unit)
    fun setOnError(callback: (message: String) -> Unit)
    fun release()
}

expect fun isAudioRecordingSupported(): Boolean
