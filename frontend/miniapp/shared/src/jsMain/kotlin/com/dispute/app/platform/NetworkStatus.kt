package com.dispute.app.platform

import androidx.compose.runtime.State
import androidx.compose.runtime.mutableStateOf
import kotlinx.browser.window
import org.w3c.dom.events.Event

actual object NetworkStatusService {
    private val _isOnlineState = mutableStateOf(true)
    actual val isOnlineState: State<Boolean> = _isOnlineState

    actual val isOnline: Boolean
        get() = _isOnlineState.value

    private var onNetworkChangeCallback: ((Boolean, String) -> Unit)? = null
    private var monitoring = false

    private val onlineHandler: (Event) -> Unit = {
        _isOnlineState.value = true
        onNetworkChangeCallback?.invoke(true, getNetworkType())
        Unit
    }

    private val offlineHandler: (Event) -> Unit = {
        _isOnlineState.value = false
        onNetworkChangeCallback?.invoke(false, "none")
        Unit
    }

    actual fun startMonitoring() {
        if (monitoring) return

        _isOnlineState.value = try {
            window.navigator.asDynamic().onLine as? Boolean ?: true
        } catch (e: Exception) {
            true
        }

        window.addEventListener("online", onlineHandler)
        window.addEventListener("offline", offlineHandler)

        monitoring = true
    }

    actual fun stopMonitoring() {
        if (!monitoring) return

        window.removeEventListener("online", onlineHandler)
        window.removeEventListener("offline", offlineHandler)

        monitoring = false
    }

    actual fun setOnNetworkChangeListener(callback: (isOnline: Boolean, networkType: String) -> Unit) {
        onNetworkChangeCallback = callback
    }

    private fun getNetworkType(): String {
        return try {
            val connection = window.navigator.asDynamic().connection
                ?: window.navigator.asDynamic().webkitConnection
            connection?.effectiveType as? String ?: "unknown"
        } catch (e: Exception) {
            "unknown"
        }
    }
}
