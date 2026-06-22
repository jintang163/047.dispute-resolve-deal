package com.dispute.app.platform

import kotlinx.browser.window
import kotlinx.coroutines.channels.awaitClose
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.callbackFlow
import kotlin.js.Promise

actual class LocationService actual constructor() {

    private var onErrorCallback: ((Int, String) -> Unit)? = null
    private var watchId: Int? = null

    actual suspend fun getCurrentLocation(): LocationResult {
        val isWx = detectWeChatMiniProgram()
        val isAlipay = detectAlipayMiniProgram()

        return when {
            isWx -> getWeChatLocation()
            isAlipay -> getAlipayLocation()
            else -> getBrowserLocation()
        }
    }

    actual suspend fun requestLocationUpdates(intervalMs: Long): Flow<LocationResult> = callbackFlow {
        val isWx = detectWeChatMiniProgram()

        if (isWx) {
            try {
                val wx = js("wx")
                val listener: (dynamic) -> Unit = { res ->
                    val longitude = res.longitude as Double
                    val latitude = res.latitude as Double
                    val accuracy = (res.accuracy ?: 0) as? Float ?: 0f
                    val speed = (res.speed ?: 0) as? Float ?: 0f
                    trySend(
                        LocationResult(
                            longitude = longitude,
                            latitude = latitude,
                            accuracy = accuracy,
                            speed = speed,
                            provider = "wechat",
                            timestamp = System.currentTimeMillis()
                        )
                    )
                }
                wx.onLocationChange(listener)
                wx.startLocationUpdate(object {})
            } catch (e: Exception) {
                onErrorCallback?.invoke(-1, "微信定位失败: ${e.message}")
            }
        } else {
            try {
                val options = js("({enableHighAccuracy: true, timeout: 10000, maximumAge: 0})")
                val id = window.navigator.geolocation.watchPosition(
                    success = { pos ->
                        val coords = pos.coords
                        trySend(
                            LocationResult(
                                longitude = coords.longitude as Double,
                                latitude = coords.latitude as Double,
                                altitude = coords.altitude as? Double,
                                accuracy = coords.accuracy as? Float ?: 0f,
                                speed = coords.speed as? Float,
                                bearing = coords.heading as? Float,
                                provider = "gps",
                                timestamp = (pos.timestamp as Number).toLong()
                            )
                        )
                    },
                    error = { err ->
                        onErrorCallback?.invoke(err.code as Int, err.message as String)
                    },
                    options = options
                ) as Int
                watchId = id
            } catch (e: Exception) {
                onErrorCallback?.invoke(-1, "浏览器定位失败: ${e.message}")
            }
        }

        awaitClose {
            stopLocationUpdates()
        }
    }

    actual fun stopLocationUpdates() {
        val isWx = detectWeChatMiniProgram()
        if (isWx) {
            try {
                val wx = js("wx")
                wx.stopLocationUpdate()
                wx.offLocationChange()
            } catch (_: Exception) {}
        } else {
            watchId?.let {
                try {
                    window.navigator.geolocation.clearWatch(it)
                } catch (_: Exception) {}
                watchId = null
            }
        }
    }

    actual fun isLocationEnabled(): Boolean {
        val isWx = detectWeChatMiniProgram()
        return if (isWx) true else try {
            js("typeof navigator !== 'undefined' && 'geolocation' in navigator") as Boolean
        } catch (e: Exception) {
            false
        }
    }

    actual fun setOnLocationError(callback: (errorCode: Int, errorMessage: String) -> Unit) {
        onErrorCallback = callback
    }

    actual fun release() {
        stopLocationUpdates()
        onErrorCallback = null
    }

    private suspend fun getBrowserLocation(): LocationResult {
        return suspendCoroutineCancellable { continuation ->
            try {
                val options = js("({enableHighAccuracy: true, timeout: 10000, maximumAge: 3000})")
                window.navigator.geolocation.getCurrentPosition(
                    success = { pos ->
                        val coords = pos.coords
                        continuation.resume(
                            LocationResult(
                                longitude = coords.longitude as Double,
                                latitude = coords.latitude as Double,
                                altitude = coords.altitude as? Double,
                                accuracy = coords.accuracy as? Float ?: 0f,
                                speed = coords.speed as? Float,
                                bearing = coords.heading as? Float,
                                provider = "gps",
                                timestamp = (pos.timestamp as Number).toLong()
                            )
                        )
                    },
                    error = { err ->
                        val code = try { err.code as Int } catch (_: Exception) { -1 }
                        val msg = try { err.message as String } catch (_: Exception) { "定位失败" }
                        onErrorCallback?.invoke(code, msg)
                        continuation.resume(
                            LocationResult(
                                longitude = 116.397428,
                                latitude = 39.90923,
                                accuracy = 0f,
                                provider = "fallback",
                                timestamp = System.currentTimeMillis(),
                                isMock = true
                            )
                        )
                    },
                    options = options
                )
            } catch (e: Exception) {
                onErrorCallback?.invoke(-1, e.message ?: "定位异常")
                continuation.resume(
                    LocationResult(
                        longitude = 116.397428,
                        latitude = 39.90923,
                        accuracy = 0f,
                        provider = "fallback",
                        timestamp = System.currentTimeMillis(),
                        isMock = true
                    )
                )
            }
        }
    }

    private suspend fun getWeChatLocation(): LocationResult {
        return suspendCoroutineCancellable { continuation ->
            try {
                val wx = js("wx")
                wx.getLocation(object {
                    val type = "gcj02"
                    val isHighAccuracy = true
                    val highAccuracyExpireTime = 4000
                    val success = { res: dynamic ->
                        val longitude = res.longitude as Double
                        val latitude = res.latitude as Double
                        val accuracy = (res.accuracy ?: 0) as? Float ?: 0f
                        val speed = (res.speed ?: 0) as? Float ?: 0f
                        continuation.resume(
                            LocationResult(
                                longitude = longitude,
                                latitude = latitude,
                                accuracy = accuracy,
                                speed = speed,
                                provider = "wechat",
                                timestamp = System.currentTimeMillis()
                            )
                        )
                    }
                    val fail = { err: dynamic ->
                        val msg = try { err.errMsg as String } catch (_: Exception) { "微信定位失败" }
                        onErrorCallback?.invoke(-2, msg)
                        continuation.resume(getBrowserLocationSync())
                    }
                })
            } catch (e: Exception) {
                continuation.resume(getBrowserLocationSync())
            }
        }
    }

    private suspend fun getAlipayLocation(): LocationResult {
        return suspendCoroutineCancellable { continuation ->
            try {
                val my = js("my")
                my.getLocation(object {
                    val cacheTimeout = 30
                    val success = { res: dynamic ->
                        val longitude = res.longitude as Double
                        val latitude = res.latitude as Double
                        val accuracy = (res.accuracy ?: 0) as? Float ?: 0f
                        continuation.resume(
                            LocationResult(
                                longitude = longitude,
                                latitude = latitude,
                                accuracy = accuracy,
                                provider = "alipay",
                                timestamp = System.currentTimeMillis()
                            )
                        )
                    }
                    val fail = { _: dynamic ->
                        continuation.resume(getBrowserLocationSync())
                    }
                })
            } catch (e: Exception) {
                continuation.resume(getBrowserLocationSync())
            }
        }
    }

    private fun getBrowserLocationSync(): LocationResult {
        return LocationResult(
            longitude = 116.397428,
            latitude = 39.90923,
            accuracy = 100f,
            provider = "mock",
            timestamp = System.currentTimeMillis(),
            isMock = true
        )
    }

    private fun detectWeChatMiniProgram(): Boolean {
        return try {
            val wxExists = js("typeof wx !== 'undefined' && typeof wx.getLocation === 'function'")
            val isMini = try {
                val sysInfo = js("wx.getSystemInfoSync()")
                val env = sysInfo.environment as? String ?: ""
                env == "wx" || sysInfo.platform != null
            } catch (_: Exception) { false }
            (wxExists as Boolean) || isMini
        } catch (_: Exception) { false }
    }

    private fun detectAlipayMiniProgram(): Boolean {
        return try {
            (js("typeof my !== 'undefined' && typeof my.getLocation === 'function'") as Boolean)
        } catch (_: Exception) { false }
    }
}

actual fun isLocationServiceSupported(): Boolean {
    return try {
        val wxOk = js("typeof wx !== 'undefined' && typeof wx.getLocation === 'function'") as Boolean
        val myOk = js("typeof my !== 'undefined' && typeof my.getLocation === 'function'") as Boolean
        val browserOk = js("typeof navigator !== 'undefined' && 'geolocation' in navigator") as Boolean
        wxOk || myOk || browserOk
    } catch (_: Exception) { true }
}

internal suspend inline fun <T> suspendCoroutineCancellable(
    crossinline block: (kotlin.coroutines.Continuation<T>) -> Unit
): T = kotlin.coroutines.suspendCoroutine { cont ->
    try {
        block(cont)
    } catch (e: Exception) {
        cont.resumeWith(Result.failure(e))
    }
}

internal fun <T> kotlin.coroutines.Continuation<T>.resume(value: T) =
    resumeWith(Result.success(value))
