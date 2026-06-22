package com.dispute.app.platform

import androidx.compose.runtime.State

expect object NetworkStatusService {
    val isOnline: Boolean
    val isOnlineState: State<Boolean>

    fun startMonitoring()
    fun stopMonitoring()
    fun setOnNetworkChangeListener(callback: (isOnline: Boolean, networkType: String) -> Unit)
}
