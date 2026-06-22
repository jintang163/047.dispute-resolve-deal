package com.dispute.app.platform

import kotlinx.browser.document
import kotlinx.browser.window
import org.w3c.dom.HTMLScriptElement

actual object PlatformMapService {

    actual var config: AmapConfig = AmapConfig()

    private var sdkLoading = false
    private var sdkLoaded = false

    actual val isAvailable: Boolean
        get() = try {
            js("typeof window !== 'undefined' && typeof document !== 'undefined'") as Boolean
        } catch (_: Exception) {
            false
        }

    actual suspend fun loadSDK(): Boolean {
        if (sdkLoaded) return true
        if (!isAvailable) return false

        if (sdkLoading) {
            return waitForSDK(15000L)
        }

        sdkLoading = true

        return try {
            if (config.securityCode.isNotBlank()) {
                js("window._AMapSecurityConfig = { securityJsCode: config.securityCode }")
            }

            val result = suspendCoroutineCancellable<Boolean> { continuation ->
                try {
                    val existing = document.querySelector("script[src*='webapi.amap.com']")
                    if (existing != null) {
                        sdkLoaded = true
                        sdkLoading = false
                        continuation.resume(true)
                        return@suspendCoroutineCancellable
                    }

                    val script = document.createElement("script") as HTMLScriptElement
                    script.type = "text/javascript"
                    script.src = "https://webapi.amap.com/maps?v=2.0&key=${config.webKey}"
                    script.async = true

                    script.onload = {
                        sdkLoaded = true
                        sdkLoading = false
                        continuation.resume(true)
                        Unit
                    }

                    script.onerror = { _, _, _, _, _ ->
                        sdkLoading = false
                        continuation.resume(false)
                        Unit
                    }

                    document.head?.appendChild(script)
                } catch (e: Exception) {
                    sdkLoading = false
                    continuation.resume(false)
                }
            }

            result
        } catch (e: Exception) {
            sdkLoading = false
            false
        }
    }

    private suspend fun waitForSDK(timeoutMs: Long): Boolean {
        val startTime = System.currentTimeMillis()
        while (System.currentTimeMillis() - startTime < timeoutMs) {
            if (isAMapReady()) {
                sdkLoaded = true
                sdkLoading = false
                return true
            }
            kotlinx.coroutines.delay(100L)
        }
        return sdkLoaded
    }

    private fun isAMapReady(): Boolean {
        return try {
            val amap = js("typeof AMap !== 'undefined'")
            (amap as? Boolean) ?: false
        } catch (_: Exception) {
            false
        }
    }

    private fun getMapInstances(): dynamic {
        return try {
            val instances = js("window._amapInstances")
            if (instances == null || instances == undefined) {
                js("window._amapInstances = {}")
                js("window._amapInstances")
            } else {
                instances
            }
        } catch (_: Exception) {
            js("({})")
        }
    }

    private fun getMapInstance(containerId: String): dynamic? {
        return try {
            val instances = getMapInstances()
            val map = instances[containerId]
            if (map == null || map == undefined) null else map
        } catch (_: Exception) {
            null
        }
    }

    private fun setMapInstance(containerId: String, map: dynamic) {
        try {
            val instances = getMapInstances()
            instances[containerId] = map
        } catch (_: Exception) {}
    }

    actual suspend fun renderMap(
        containerId: String,
        centerLng: Double,
        centerLat: Double,
        zoom: Int,
        mapType: MapType
    ): Boolean {
        if (!loadSDK()) return false

        return try {
            val container = js("document.getElementById(containerId)")
            if (container == null || container == undefined) {
                return false
            }

            val existingMap = getMapInstance(containerId)
            if (existingMap != null) {
                try {
                    existingMap.destroy()
                } catch (_: Exception) {}
            }

            val mapStyle = when (mapType) {
                MapType.NORMAL -> "amap://styles/standard"
                MapType.SATELLITE -> "amap://styles/satellite"
                MapType.TRAFFIC -> "amap://styles/standard"
                MapType.NIGHT -> "amap://styles/dark"
            }

            val AMap = js("AMap")
            val map = AMap.Map(container, js("({
                center: [centerLng, centerLat],
                zoom: zoom,
                mapStyle: mapStyle,
                viewMode: '2D'
            })"))

            if (mapType == MapType.SATELLITE) {
                try {
                    val satellite = AMap.TileLayer.Satellite()
                    val roadNet = AMap.TileLayer.RoadNet()
                    map.add([satellite, roadNet])
                } catch (_: Exception) {}
            }

            if (mapType == MapType.TRAFFIC) {
                try {
                    val traffic = AMap.TileLayer.Traffic(js("({
                        autoRefresh: true,
                        interval: 180
                    })"))
                    map.add(traffic)
                } catch (_: Exception) {}
            }

            setMapInstance(containerId, map)
            true
        } catch (e: Exception) {
            false
        }
    }

    actual suspend fun addMarkers(
        containerId: String,
        markers: List<MapMarker>,
        onClick: ((MapMarker) -> Unit)?
    ) {
        val map = getMapInstance(containerId) ?: return
        if (!isAMapReady()) return

        try {
            val AMap = js("AMap")

            for (marker in markers) {
                val markerColor = when {
                    marker.isSelected -> "#1D6CFF"
                    marker.isCompleted -> "#22C55E"
                    else -> "#EF4444"
                }

                val iconContent = generateMarkerIcon(markerColor, marker.title)

                val amapMarker = AMap.Marker(js("({
                    position: [marker.longitude, marker.latitude],
                    title: marker.title,
                    content: iconContent,
                    offset: AMap.Pixel(-13, -30),
                    zIndex: if (marker.isSelected) 200 else 100 + marker.sortIndex
                })"))

                if (onClick != null) {
                    amapMarker.on("click", { _: dynamic ->
                        onClick(marker)
                        Unit
                    })
                }

                map.add(amapMarker)
            }
        } catch (_: Exception) {}
    }

    private fun generateMarkerIcon(color: String, label: String): String {
        val shortLabel = if (label.length > 2) label.substring(0, 2) else label
        return """
            <div style="position:relative;display:flex;flex-direction:column;align-items:center;">
                <div style="
                    background:$color;
                    color:#fff;
                    font-size:11px;
                    font-weight:bold;
                    padding:3px 8px;
                    border-radius:4px;
                    box-shadow:0 2px 6px rgba(0,0,0,0.3);
                    white-space:nowrap;
                    max-width:100px;
                    overflow:hidden;
                    text-overflow:ellipsis;
                ">${shortLabel.ifBlank { "📍" }}</div>
                <div style="
                    width:0;
                    height:0;
                    border-left:6px solid transparent;
                    border-right:6px solid transparent;
                    border-top:8px solid $color;
                    margin-top:-1px;
                "></div>
            </div>
        """.trimIndent()
    }

    actual suspend fun drawRoute(
        containerId: String,
        route: MapRoutePath,
        startMarker: MapMarker?,
        endMarker: MapMarker?,
        color: String,
        width: Int
    ) {
        val map = getMapInstance(containerId) ?: return
        if (!isAMapReady()) return

        try {
            val AMap = js("AMap")

            if (route.polyline.isNotEmpty()) {
                val path = route.polyline.map { pair ->
                    arrayOf(pair.first, pair.second)
                }.toTypedArray()

                val polyline = AMap.Polyline(js("({
                    path: path,
                    strokeColor: color,
                    strokeWeight: width,
                    strokeOpacity: 0.9,
                    strokeStyle: 'solid',
                    lineJoin: 'round'
                })"))

                map.add(polyline)
            }

            if (startMarker != null) {
                addSingleMarker(map, AMap, startMarker, "#22C55E", "起点")
            }
            if (endMarker != null) {
                addSingleMarker(map, AMap, endMarker, "#EF4444", "终点")
            }
        } catch (_: Exception) {}
    }

    private fun addSingleMarker(map: dynamic, AMap: dynamic, marker: MapMarker, color: String, defaultLabel: String) {
        try {
            val iconContent = generateMarkerIcon(color, marker.title.ifBlank { defaultLabel })
            val amapMarker = AMap.Marker(js("({
                position: [marker.longitude, marker.latitude],
                title: marker.title,
                content: iconContent,
                offset: AMap.Pixel(-13, -30),
                zIndex: 150
            })"))
            map.add(amapMarker)
        } catch (_: Exception) {}
    }

    actual suspend fun clearMap(containerId: String) {
        val map = getMapInstance(containerId) ?: return
        try {
            map.clearMap()
        } catch (_: Exception) {}
    }

    actual suspend fun setMapType(containerId: String, mapType: MapType) {
        val map = getMapInstance(containerId) ?: return
        if (!isAMapReady()) return

        try {
            val AMap = js("AMap")

            map.clearMap()

            val allLayers = map.getLayers() ?: emptyArray<dynamic>()
            for (layer in allLayers) {
                try {
                    map.remove(layer)
                } catch (_: Exception) {}
            }

            when (mapType) {
                MapType.NORMAL -> {
                    map.setMapStyle("amap://styles/standard")
                }
                MapType.SATELLITE -> {
                    map.setMapStyle("amap://styles/satellite")
                    try {
                        val satellite = AMap.TileLayer.Satellite()
                        val roadNet = AMap.TileLayer.RoadNet()
                        map.add([satellite, roadNet])
                    } catch (_: Exception) {}
                }
                MapType.TRAFFIC -> {
                    map.setMapStyle("amap://styles/standard")
                    try {
                        val traffic = AMap.TileLayer.Traffic(js("({
                            autoRefresh: true,
                            interval: 180
                        })"))
                        map.add(traffic)
                    } catch (_: Exception) {}
                }
                MapType.NIGHT -> {
                    map.setMapStyle("amap://styles/dark")
                }
            }
        } catch (_: Exception) {}
    }

    actual suspend fun fitMarkers(containerId: String, markers: List<MapMarker>) {
        val map = getMapInstance(containerId) ?: return
        if (markers.isEmpty()) return
        if (!isAMapReady()) return

        try {
            val AMap = js("AMap")
            val positions = markers.map { marker ->
                AMap.LngLat(marker.longitude, marker.latitude)
            }.toTypedArray()

            map.setFitView(null, false, [60, 60, 60, 60])
        } catch (_: Exception) {
            try {
                val lngs = markers.map { it.longitude }
                val lats = markers.map { it.latitude }
                val minLng = lngs.minOrNull() ?: return
                val maxLng = lngs.maxOrNull() ?: return
                val minLat = lats.minOrNull() ?: return
                val maxLat = lats.maxOrNull() ?: return
                val centerLng = (minLng + maxLng) / 2
                val centerLat = (minLat + maxLat) / 2
                map.setCenter([centerLng, centerLat])
            } catch (_: Exception) {}
        }
    }
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
