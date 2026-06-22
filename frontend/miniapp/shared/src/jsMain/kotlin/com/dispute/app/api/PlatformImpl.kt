package com.dispute.app.api

import io.ktor.client.engine.HttpClientEngine
import io.ktor.client.engine.js.Js

actual fun createHttpClientEngine(): HttpClientEngine = Js.create()

actual object PlatformApiConfigImpl : PlatformApiConfig {
    private const val KEY_BASE_URL = "dispute_base_url"
    private const val KEY_AUTH_TOKEN = "dispute_auth_token"
    private const val KEY_DEVICE_ID = "dispute_device_id"
    private const val KEY_PLATFORM = "dispute_platform"

    override val apiPrefix: String = "/api/v1"

    override var baseUrl: String
        get() = getLocalStorage(KEY_BASE_URL) ?: "http://localhost:8080"
        set(value) = setLocalStorage(KEY_BASE_URL, value)

    override var authToken: String?
        get() = getLocalStorage(KEY_AUTH_TOKEN)
        set(value) {
            if (value != null) {
                setLocalStorage(KEY_AUTH_TOKEN, value)
            } else {
                removeLocalStorage(KEY_AUTH_TOKEN)
            }
        }

    override var deviceId: String?
        get() = getLocalStorage(KEY_DEVICE_ID)
        set(value) {
            if (value != null) {
                setLocalStorage(KEY_DEVICE_ID, value)
            } else {
                removeLocalStorage(KEY_DEVICE_ID)
            }
        }

    override var platform: String
        get() = getLocalStorage(KEY_PLATFORM) ?: detectPlatform()
        set(value) = setLocalStorage(KEY_PLATFORM, value)

    private fun detectPlatform(): String {
        val userAgent = js("navigator.userAgent") as? String ?: ""
        return when {
            js("typeof wx !== 'undefined'") as Boolean -> "wechat-miniapp"
            js("typeof my !== 'undefined'") as Boolean -> "alipay-miniapp"
            userAgent.contains("MicroMessenger") -> "wechat-h5"
            userAgent.contains("AlipayClient") -> "alipay-h5"
            userAgent.contains("Android") -> "android-h5"
            userAgent.contains("iPhone") || userAgent.contains("iPad") -> "ios-h5"
            else -> "web-h5"
        }
    }

    private fun getLocalStorage(key: String): String? {
        return try {
            val storage = js("localStorage")
            storage.getItem(key) as? String
        } catch (e: Exception) {
            null
        }
    }

    private fun setLocalStorage(key: String, value: String) {
        try {
            val storage = js("localStorage")
            storage.setItem(key, value)
        } catch (e: Exception) {
        }
    }

    private fun removeLocalStorage(key: String) {
        try {
            val storage = js("localStorage")
            storage.removeItem(key)
        } catch (e: Exception) {
        }
    }

    init {
        if (deviceId == null) {
            deviceId = "dev_" + generateRandomId()
        }
    }

    private fun generateRandomId(): String {
        val chars = "abcdefghijklmnopqrstuvwxyz0123456789"
        var result = ""
        repeat(16) {
            val index = (kotlin.math.random() * chars.length).toInt()
            result += chars[index]
        }
        return result
    }
}
