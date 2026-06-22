package com.dispute.app.storage

import kotlinx.browser.window
import kotlinx.serialization.Serializable
import kotlinx.serialization.decodeFromString
import kotlinx.serialization.encodeToString
import kotlinx.serialization.json.Json

private val json = Json {
    prettyPrint = false
    isLenient = true
    ignoreUnknownKeys = true
    encodeDefaults = true
}

actual object LocalStorage {
    private const val KEY_AUDIO_PREFIX = "audio_"
    private const val KEY_AUDIO_META_LIST = "audio_meta_list"
    private const val KEY_TRANSCRIBE_TASKS = "transcribe_tasks"

    actual fun getString(key: String): String? {
        return try {
            window.localStorage.getItem(key)
        } catch (e: Exception) {
            null
        }
    }

    actual fun setString(key: String, value: String) {
        try {
            window.localStorage.setItem(key, value)
        } catch (e: Exception) {
        }
    }

    actual fun remove(key: String) {
        try {
            window.localStorage.removeItem(key)
        } catch (e: Exception) {
        }
    }

    actual suspend fun putAudio(taskId: String, audioBase64: String, meta: AudioStorageMeta): Boolean {
        return try {
            setString(KEY_AUDIO_PREFIX + taskId, audioBase64)

            val metaList = getAllAudioMeta().toMutableList()
            metaList.removeAll { it.taskId == taskId }
            metaList.add(meta)
            saveAudioMetaList(metaList)

            true
        } catch (e: Exception) {
            false
        }
    }

    actual suspend fun getAudio(taskId: String): String? {
        return getString(KEY_AUDIO_PREFIX + taskId)
    }

    actual suspend fun removeAudio(taskId: String): Boolean {
        return try {
            remove(KEY_AUDIO_PREFIX + taskId)

            val metaList = getAllAudioMeta().toMutableList()
            metaList.removeAll { it.taskId == taskId }
            saveAudioMetaList(metaList)

            true
        } catch (e: Exception) {
            false
        }
    }

    actual suspend fun getAllAudioTasks(): List<AudioStorageMeta> {
        return getAllAudioMeta()
    }

    private fun getAllAudioMeta(): List<AudioStorageMeta> {
        return try {
            val jsonStr = getString(KEY_AUDIO_META_LIST) ?: return emptyList()
            json.decodeFromString<List<AudioStorageMeta>>(jsonStr)
        } catch (e: Exception) {
            emptyList()
        }
    }

    private fun saveAudioMetaList(list: List<AudioStorageMeta>) {
        try {
            setString(KEY_AUDIO_META_LIST, json.encodeToString(list))
        } catch (e: Exception) {
        }
    }

    actual suspend fun addTranscribeTask(task: OfflineTranscribeTask): Boolean {
        return try {
            val tasks = getTranscribeTaskList().toMutableList()
            tasks.removeAll { it.taskId == task.taskId }
            tasks.add(task)
            saveTranscribeTasks(tasks)
            true
        } catch (e: Exception) {
            false
        }
    }

    actual suspend fun updateTranscribeTask(task: OfflineTranscribeTask): Boolean {
        return try {
            val tasks = getTranscribeTaskList().toMutableList()
            val index = tasks.indexOfFirst { it.taskId == task.taskId }
            if (index >= 0) {
                tasks[index] = task
                saveTranscribeTasks(tasks)
                true
            } else {
                false
            }
        } catch (e: Exception) {
            false
        }
    }

    actual suspend fun getTranscribeTasks(status: String?): List<OfflineTranscribeTask> {
        val allTasks = getTranscribeTaskList()
        return if (status.isNullOrBlank()) {
            allTasks
        } else {
            allTasks.filter { it.status.value == status }
        }
    }

    actual suspend fun removeTranscribeTask(taskId: String): Boolean {
        return try {
            val tasks = getTranscribeTaskList().toMutableList()
            tasks.removeAll { it.taskId == taskId }
            saveTranscribeTasks(tasks)
            true
        } catch (e: Exception) {
            false
        }
    }

    private fun getTranscribeTaskList(): List<OfflineTranscribeTask> {
        return try {
            val jsonStr = getString(KEY_TRANSCRIBE_TASKS) ?: return emptyList()
            json.decodeFromString<List<OfflineTranscribeTask>>(jsonStr)
        } catch (e: Exception) {
            emptyList()
        }
    }

    private fun saveTranscribeTasks(tasks: List<OfflineTranscribeTask>) {
        try {
            setString(KEY_TRANSCRIBE_TASKS, json.encodeToString(tasks))
        } catch (e: Exception) {
        }
    }
}
