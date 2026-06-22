package com.dispute.app.platform

expect class AMapService() {

    suspend fun planDrivingRoute(
        startLng: Double,
        startLat: Double,
        points: List<RoutePoint>,
        strategy: Int = 10
    ): AMapRouteResult

    suspend fun planWalkingRoute(
        startLng: Double,
        startLat: Double,
        endLng: Double,
        endLat: Double
    ): AMapRouteResult

    suspend fun reverseGeocode(
        longitude: Double,
        latitude: Double
    ): AddressResult

    suspend fun geocode(address: String, city: String? = null): GeoCodeResult

    fun calculateLineDistance(
        lng1: Double,
        lat1: Double,
        lng2: Double,
        lat2: Double
    ): Double

    fun setApiKey(key: String)

    fun release()
}

data class AMapRouteResult(
    val totalDistance: Int = 0,
    val totalDuration: Int = 0,
    val totalTaxiCost: Float = 0f,
    val strategy: Int = 10,
    val paths: List<AMapPath> = emptyList(),
    val orderedPoints: List<RoutePoint> = emptyList()
)

data class AMapPath(
    val distance: Int = 0,
    val duration: Int = 0,
    val strategy: String = "",
    val steps: List<AMapStep> = emptyList(),
    val polyline: String = ""
)

data class AMapStep(
    val instruction: String = "",
    val road: String = "",
    val distance: Int = 0,
    val duration: Int = 0,
    val polyline: String = "",
    val action: String = "",
    val orientation: String = ""
)

data class GeoCodeResult(
    val longitude: Double = 0.0,
    val latitude: Double = 0.0,
    val formattedAddress: String = "",
    val level: String = ""
)

expect fun isAMapServiceSupported(): Boolean

object RouteVisualizer() {

    fun generateGoogleMapsPolylineUrl(
        apiKey: String,
        points: List<Pair<Double, Double>>,
        width: Int = 600,
        height: Int = 400,
        color: String = "blue",
        weight: Int = 5,
        markersColor: String = "red"
    ): String {
        val pathStr = points.joinToString("|") { "${it.second},${it.first}" }
        val markersStr = points.mapIndexed { index, point ->
            val label = ('A'.code + index).toChar()
            "markers=color:$markersColor%7Clabel:$label%7C${point.second},${point.first}" }
            .joinToString("&")

        return "https://maps.googleapis.com/maps/api/staticmap?" +
                "size=${width}x${height}" +
                "&path=color:$color|weight:$weight|$pathStr" +
                "&$markersStr" +
                "&key=$apiKey"
    }

    fun generateAMapStaticUrl(
        apiKey: String,
        points: List<Pair<Double, Double>>,
        width: Int = 600,
        height: Int = 400,
        size: Int = 2,
        scale: Int = 2
    ): String {
        if (points.isEmpty()) return ""
        val locations = points.joinToString(";") { "${it.first},${it.second}" }
        return "https://restapi.amap.com/v3/staticmap" +
                "?location=${points.first().first},${points.first().second}" +
                "&zoom=14" +
                "&size=${width}*${height}" +
                "&scale=$scale" +
                "&markers=large,0xFF0000,A:${points.first().first},${points.first().second}" +
                "&paths=2,0x1D6CFF,0.7:$locations" +
                "&key=$apiKey"
    }

    fun getMapConfig(
        centerLng: Double,
        centerLat: Double,
        points: List<RoutePoint>,
        level: Int = 14,
        zoom: Int = 14,
        mapType: Int = 0
    ): MapViewConfig {
        return MapViewConfig(
            centerLng = centerLng,
            centerLat = centerLat,
            points = points,
            level = level,
            zoom = zoom,
            mapType = mapType,
            markers = points.mapIndexed { index, point ->
                MapMarker(
                    position = index,
                    longitude = point.longitude,
                    latitude = point.latitude,
                    title = point.name,
                    snippet = point.address,
                    isDraggable = false
                )
            }
        )
    }
}

data class MapViewConfig(
    val centerLng: Double = 0.0,
    val centerLat: Double = 0.0,
    val level: Int = 14,
    val zoom: Int = 14,
    val mapType: Int = 0,
    val points: List<RoutePoint> = emptyList(),
    val markers: List<MapMarker> = emptyList(),
    val polylines: List<MapPolyline> = emptyList(),
    val showTraffic: Boolean = false,
    val showBuilding: Boolean = true,
    val showScale: Boolean = true
)

data class MapMarker(
    val position: Int = 0,
    val longitude: Double = 0.0,
    val latitude: Double = 0.0,
    val title: String = "",
    val snippet: String = "",
    val icon: String = "",
    val isDraggable: Boolean = false
)

data class MapPolyline(
    val points: List<Pair<Double, Double>> = emptyList(),
    val color: Long = 0xFF1D6CFF,
    val width: Float = 5f,
    val dottedLine: Boolean = false
)

object CoordinateConverter() {

    fun bd09ToGcj02(bdLng: Double, bdLat: Double): Pair<Double, Double> {
        val x = bdLng - 0.0065
        val y = bdLat - 0.006
        val z = Math.sqrt(x * x + y * y) - 0.00002 * Math.sin(y * Math.PI * 3000.0 / 180.0 * Math.PI)
        val theta = Math.atan2(y, x) - 0.000003 * Math.cos(x * Math.PI * 3000.0 / 180.0 * Math.PI)
        val ggLng = z * Math.cos(theta)
        val ggLat = z * Math.sin(theta)
        return Pair(ggLng, ggLat)
    }

    fun gcj02ToWgs84(lng: Double, lat: Double): Pair<Double, Double> {
        if (outOfChina(lng, lat)) {
            return Pair(lng, lat)
        }
        val dLat = transformLat(lng - 105.0, lat - 35.0)
        val dLng = transformLng(lng - 105.0, lat - 35.0)
        val radLat = lat / 180.0 * Math.PI
        var magic = Math.sin(radLat)
        magic = 1 - 0.0066934216223212 * magic * magic
        val sqrtMagic = Math.sqrt(magic)
        val d = (20.0 * Math.sin((lng)) / (180.0 * Math.PI))
        magic = (180.0 - d) / Math.PI
        val realLat = lat + d * 2.0
        magic = d * 2.0
        val realLng = lng + magic
        return Pair(lng + (realLng - lng), lat + (realLat - lat))
    }

    fun wgs84ToGcj02(lng: Double, lat: Double): Pair<Double, Double> {
        if (outOfChina(lng, lat)) {
            return Pair(lng, lat)
        }
        val dLat = transformLat(lng - 105.0, lat - 35.0)
        val dLng = transformLng(lng - 105.0, lat - 35.0)
        val radLat = lat / 180.0 * Math.PI
        var magic = Math.sin(radLat)
        magic = 1 - 0.0066934216223212 * magic * magic
        val sqrtMagic = Math.sqrt(magic)
        val d = (20.0 * Math.sin(radLat * 3.0) / (180.0 * Math.PI))
        magic = (180.0 - d) / Math.PI
        val realLat = lat + d * 2.0
        magic = d * 2.0
        val realLng = lng + magic
        return Pair(lng + (realLng - lng), lat + (realLat - lat))
    }

    private fun outOfChina(lng: Double, lat: Double): Boolean {
        return !(lng > 73.66 && lng < 135.05 && lat > 3.86 && lat < 53.55)
    }

    private fun transformLat(x: Double, y: Double): Double {
        var ret = -100.0 + 2.0 * x + 3.0 * y + 0.2 * y * y + 0.1 * x * y + 0.2 * Math.sqrt(Math.abs(x))
        ret += (20.0 * Math.sin(6.0 * x * Math.PI) + 20.0 * Math.sin(2.0 * x * Math.PI)) * 2.0 / 3.0
        ret += (20.0 * Math.sin(y * Math.PI) + 40.0 * Math.sin(y / 3.0 * Math.PI) * 2.0 / 3.0
        ret += (160.0 * Math.sin(y / 12.0 * Math.PI) + 320 * Math.sin(y * Math.PI / 30.0)) * 2.0 / 3.0
        return ret
    }

    private fun transformLng(x: Double, y: Double): Double {
        var ret = 300.0 + x + 2.0 * y + 0.1 * x * x + 0.1 * x * y + 0.1 * Math.sqrt(Math.abs(x))
        ret += (20.0 * Math.sin(6.0 * x * Math.PI) + 20.0 * Math.sin(2.0 * x * Math.PI)) * 2.0 / 3.0
        ret += (20.0 * Math.sin(x * Math.PI) + 40.0 * Math.sin(x / 3.0 * Math.PI)) * 2.0 / 3.0
        ret += (150.0 * Math.sin(x / 12.0 * Math.PI) + 300.0 * Math.sin(x / 30.0 * Math.PI)) * 2.0 / 3.0
        return ret
    }
}
