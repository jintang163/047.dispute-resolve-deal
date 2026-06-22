package com.dispute.app.platform

import kotlinx.serialization.Serializable

@Serializable
data class MapMarker(
    val id: String,
    val longitude: Double,
    val latitude: Double,
    val title: String = "",
    val snippet: String = "",
    val icon: String = "red",
    val sortIndex: Int = 0,
    val isCompleted: Boolean = false,
    val isSelected: Boolean = false
)

@Serializable
data class MapRoutePath(
    val polyline: List<Pair<Double, Double>> = emptyList(),
    val distance: Double = 0.0,
    val duration: Long = 0L,
    val strategy: Int = 10,
    val strategyName: String = "速度优先"
)

@Serializable
data class AmapConfig(
    val webKey: String = "",
    val securityCode: String = "",
    val defaultCity: String = "北京"
)

enum class MapType(val value: Int) {
    NORMAL(1),
    SATELLITE(2),
    TRAFFIC(3),
    NIGHT(4)
}

expect object PlatformMapService {
    val isAvailable: Boolean
    var config: AmapConfig

    suspend fun loadSDK(): Boolean
    suspend fun renderMap(
        containerId: String,
        centerLng: Double,
        centerLat: Double,
        zoom: Int = 14,
        mapType: MapType = MapType.NORMAL
    ): Boolean

    suspend fun addMarkers(
        containerId: String,
        markers: List<MapMarker>,
        onClick: ((MapMarker) -> Unit)? = null
    )

    suspend fun drawRoute(
        containerId: String,
        route: MapRoutePath,
        startMarker: MapMarker? = null,
        endMarker: MapMarker? = null,
        color: String = "#1D6CFF",
        width: Int = 6
    )

    suspend fun clearMap(containerId: String)
    suspend fun setMapType(containerId: String, mapType: MapType)
    suspend fun fitMarkers(containerId: String, markers: List<MapMarker>)
}
