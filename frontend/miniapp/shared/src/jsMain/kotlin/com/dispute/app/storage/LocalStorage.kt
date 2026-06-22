package com.dispute.app.storage

import kotlinx.browser.window
import kotlinx.coroutines.await
import kotlinx.serialization.decodeFromString
import kotlinx.serialization.encodeToString
import kotlinx.serialization.json.Json
import kotlin.js.Promise

private val json = Json {
    prettyPrint = false
    isLenient = true
    ignoreUnknownKeys = true
    encodeDefaults = true
}

private const val DB_NAME = "DisputeAppDB"
private const val DB_VERSION = 1

private const val STORE_AUDIO_FILES = "audio_files"
private const val STORE_TRANSCRIBE_TASKS = "transcribe_tasks"
private const val STORE_KV = "kv_store"

private const val KEY_AUDIO_PREFIX = "audio_"
private const val KEY_AUDIO_META_LIST = "audio_meta_list"
private const val KEY_TRANSCRIBE_TASKS = "transcribe_tasks"

actual object LocalStorage {
    private var dbInitialized = false
    private var useIndexedDB = false
    private var db: dynamic = null

    private suspend fun ensureDb(): Boolean {
        if (dbInitialized) return useIndexedDB
        dbInitialized = true
        useIndexedDB = try {
            openDatabase().await()
            true
        } catch (e: Exception) {
            console.warn("IndexedDB 不可用，使用 localStorage 降级，音频容量受限", e)
            false
        }
        return useIndexedDB
    }

    private fun openDatabase(): Promise<dynamic> = Promise { resolve, reject ->
        try {
            val indexedDB = window.asDynamic().indexedDB
            if (indexedDB == null) {
                reject(Exception("IndexedDB not supported"))
                return@Promise
            }
            val request = indexedDB.open(DB_NAME, DB_VERSION)
            request.onerror = { _ ->
                reject(request.error ?: Exception("Failed to open IndexedDB"))
            }
            request.onsuccess = { _ ->
                db = request.result
                resolve(db)
            }
            request.onupgradeneeded = { event ->
                val database = event.target.result
                if (!database.objectStoreNames.contains(STORE_AUDIO_FILES)) {
                    database.createObjectStore(STORE_AUDIO_FILES, js("({keyPath: \"taskId\"})"))
                }
                if (!database.objectStoreNames.contains(STORE_TRANSCRIBE_TASKS)) {
                    database.createObjectStore(STORE_TRANSCRIBE_TASKS, js("({keyPath: \"taskId\"})"))
                }
                if (!database.objectStoreNames.contains(STORE_KV)) {
                    database.createObjectStore(STORE_KV, js("({keyPath: \"key\"})"))
                }
            }
        } catch (e: Exception) {
            reject(e)
        }
    }

    private fun <T> idbRequestToPromise(request: dynamic): Promise<T> = Promise { resolve, reject ->
        request.onsuccess = { _ -> resolve(request.result as T) }
        request.onerror = { _ -> reject(request.error ?: Exception("IDBRequest failed")) }
    }

    private fun transaction(storeName: String, mode: String): dynamic {
        val tx = db.transaction(arrayOf(storeName), mode)
        return tx.objectStore(storeName)
    }

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
        if (!ensureDb()) {
            return putAudioFallback(taskId, audioBase64, meta)
        }
        return try {
            val now = System.currentTimeMillis()
            val metaJson = json.encodeToString(meta)
            val record = js("({})").apply {
                this.taskId = taskId
                this.audioBase64 = audioBase64
                this.meta = metaJson
                this.createdAt = now
            }
            val store = transaction(STORE_AUDIO_FILES, "readwrite")
            idbRequestToPromise<Unit>(store.put(record)).await()
            true
        } catch (e: Exception) {
            console.warn("IndexedDB putAudio 失败，尝试 localStorage", e)
            putAudioFallback(taskId, audioBase64, meta)
        }
    }

    private fun putAudioFallback(taskId: String, audioBase64: String, meta: AudioStorageMeta): Boolean {
        return try {
            setString(KEY_AUDIO_PREFIX + taskId, audioBase64)
            val metaList = getAllAudioMetaFallback().toMutableList()
            metaList.removeAll { it.taskId == taskId }
            metaList.add(meta)
            saveAudioMetaListFallback(metaList)
            true
        } catch (e: Exception) {
            false
        }
    }

    actual suspend fun getAudio(taskId: String): String? {
        if (!ensureDb()) {
            return getAudioFallback(taskId)
        }
        return try {
            val store = transaction(STORE_AUDIO_FILES, "readonly")
            val result = idbRequestToPromise<dynamic?>(store.get(taskId)).await()
            if (result != null) result.audioBase64 as String? else null
        } catch (e: Exception) {
            getAudioFallback(taskId)
        }
    }

    private fun getAudioFallback(taskId: String): String? = getString(KEY_AUDIO_PREFIX + taskId)

    actual suspend fun removeAudio(taskId: String): Boolean {
        if (!ensureDb()) {
            return removeAudioFallback(taskId)
        }
        return try {
            val store = transaction(STORE_AUDIO_FILES, "readwrite")
            idbRequestToPromise<Unit>(store.delete(taskId)).await()
            true
        } catch (e: Exception) {
            removeAudioFallback(taskId)
        }
    }

    private fun removeAudioFallback(taskId: String): Boolean {
        return try {
            remove(KEY_AUDIO_PREFIX + taskId)
            val metaList = getAllAudioMetaFallback().toMutableList()
            metaList.removeAll { it.taskId == taskId }
            saveAudioMetaListFallback(metaList)
            true
        } catch (e: Exception) {
            false
        }
    }

    actual suspend fun getAllAudioTasks(): List<AudioStorageMeta> {
        if (!ensureDb()) {
            return getAllAudioMetaFallback()
        }
        return try {
            val store = transaction(STORE_AUDIO_FILES, "readonly")
            val request = store.openCursor()
            val result = mutableListOf<AudioStorageMeta>()
            Promise<List<AudioStorageMeta>> { resolve, _ ->
                request.onsuccess = { event ->
                    val cursor = event.target.result
                    if (cursor != null) {
                        try {
                            val metaJson = cursor.value.meta as String
                            val meta = json.decodeFromString<AudioStorageMeta>(metaJson)
                            result.add(meta)
                        } catch (_: Exception) {
                        }
                        cursor.continue()
                    } else {
                        resolve(result)
                    }
                }
                request.onerror = { _ -> resolve(result) }
            }.await()
        } catch (e: Exception) {
            getAllAudioMetaFallback()
        }
    }

    private fun getAllAudioMetaFallback(): List<AudioStorageMeta> {
        return try {
            val jsonStr = getString(KEY_AUDIO_META_LIST) ?: return emptyList()
            json.decodeFromString<List<AudioStorageMeta>>(jsonStr)
        } catch (e: Exception) {
            emptyList()
        }
    }

    private fun saveAudioMetaListFallback(list: List<AudioStorageMeta>) {
        try {
            setString(KEY_AUDIO_META_LIST, json.encodeToString(list))
        } catch (e: Exception) {
        }
    }

    actual suspend fun addTranscribeTask(task: OfflineTranscribeTask): Boolean {
        if (!ensureDb()) {
            return addTranscribeTaskFallback(task)
        }
        return try {
            val store = transaction(STORE_TRANSCRIBE_TASKS, "readwrite")
            idbRequestToPromise<Unit>(store.put(task)).await()
            true
        } catch (e: Exception) {
            addTranscribeTaskFallback(task)
        }
    }

    private fun addTranscribeTaskFallback(task: OfflineTranscribeTask): Boolean {
        return try {
            val tasks = getTranscribeTaskListFallback().toMutableList()
            tasks.removeAll { it.taskId == task.taskId }
            tasks.add(task)
            saveTranscribeTasksFallback(tasks)
            true
        } catch (e: Exception) {
            false
        }
    }

    actual suspend fun updateTranscribeTask(task: OfflineTranscribeTask): Boolean {
        if (!ensureDb()) {
            return updateTranscribeTaskFallback(task)
        }
        return try {
            val store = transaction(STORE_TRANSCRIBE_TASKS, "readwrite")
            val existing = idbRequestToPromise<dynamic?>(store.get(task.taskId)).await()
            if (existing != null) {
                idbRequestToPromise<Unit>(store.put(task)).await()
                true
            } else {
                false
            }
        } catch (e: Exception) {
            updateTranscribeTaskFallback(task)
        }
    }

    private fun updateTranscribeTaskFallback(task: OfflineTranscribeTask): Boolean {
        return try {
            val tasks = getTranscribeTaskListFallback().toMutableList()
            val index = tasks.indexOfFirst { it.taskId == task.taskId }
            if (index >= 0) {
                tasks[index] = task
                saveTranscribeTasksFallback(tasks)
                true
            } else {
                false
            }
        } catch (e: Exception) {
            false
        }
    }

    actual suspend fun getTranscribeTasks(status: String?): List<OfflineTranscribeTask> {
        if (!ensureDb()) {
            return getTranscribeTasksFallback(status)
        }
        return try {
            val store = transaction(STORE_TRANSCRIBE_TASKS, "readonly")
            val request = store.openCursor()
            val result = mutableListOf<OfflineTranscribeTask>()
            Promise<List<OfflineTranscribeTask>> { resolve, _ ->
                request.onsuccess = { event ->
                    val cursor = event.target.result
                    if (cursor != null) {
                        try {
                            val task = convertToTask(cursor.value)
                            if (task != null) {
                                if (status.isNullOrBlank() || task.status.value == status) {
                                    result.add(task)
                                }
                            }
                        } catch (_: Exception) {
                        }
                        cursor.continue()
                    } else {
                        resolve(result)
                    }
                }
                request.onerror = { _ -> resolve(result) }
            }.await()
        } catch (e: Exception) {
            getTranscribeTasksFallback(status)
        }
    }

    private fun convertToTask(obj: dynamic): OfflineTranscribeTask? {
        return try {
            val jsonStr = JSON.stringify(obj)
            json.decodeFromString<OfflineTranscribeTask>(jsonStr)
        } catch (e: Exception) {
            null
        }
    }

    private fun getTranscribeTasksFallback(status: String?): List<OfflineTranscribeTask> {
        val allTasks = getTranscribeTaskListFallback()
        return if (status.isNullOrBlank()) {
            allTasks
        } else {
            allTasks.filter { it.status.value == status }
        }
    }

    actual suspend fun removeTranscribeTask(taskId: String): Boolean {
        if (!ensureDb()) {
            return removeTranscribeTaskFallback(taskId)
        }
        return try {
            val store = transaction(STORE_TRANSCRIBE_TASKS, "readwrite")
            idbRequestToPromise<Unit>(store.delete(taskId)).await()
            true
        } catch (e: Exception) {
            removeTranscribeTaskFallback(taskId)
        }
    }

    private fun removeTranscribeTaskFallback(taskId: String): Boolean {
        return try {
            val tasks = getTranscribeTaskListFallback().toMutableList()
            tasks.removeAll { it.taskId == taskId }
            saveTranscribeTasksFallback(tasks)
            true
        } catch (e: Exception) {
            false
        }
    }

    private fun getTranscribeTaskListFallback(): List<OfflineTranscribeTask> {
        return try {
            val jsonStr = getString(KEY_TRANSCRIBE_TASKS) ?: return emptyList()
            json.decodeFromString<List<OfflineTranscribeTask>>(jsonStr)
        } catch (e: Exception) {
            emptyList()
        }
    }

    private fun saveTranscribeTasksFallback(tasks: List<OfflineTranscribeTask>) {
        try {
            setString(KEY_TRANSCRIBE_TASKS, json.encodeToString(tasks))
        } catch (e: Exception) {
        }
    }
}
