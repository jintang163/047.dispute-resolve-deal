package com.dispute.app.platform

import kotlinx.coroutines.flow.Flow

expect class LocationService() {

    suspend fun getCurrentLocation(): LocationResult

    suspend fun requestLocationUpdates(intervalMs: Long = 5000): Flow<LocationResult>

    fun stopLocationUpdates()

    fun isLocationEnabled(): Boolean

    fun setOnLocationError(callback: (errorCode: Int, errorMessage: String) -> Unit)

    fun release()
}

data class LocationResult(
    val longitude: Double,
    val latitude: Double,
    val altitude: Double? = null,
    val accuracy: Float = 0f,
    val speed: Float? = null,
    val bearing: Float? = null,
    val timestamp: Long = System.currentTimeMillis(),
    val provider: String = "",
    val isMock: Boolean = false
)

expect fun isLocationServiceSupported(): Boolean

suspend fun LocationService.getAddressByLocation(
    longitude: Double,
    latitude: Double
): AddressResult {
    return AddressResult(
        longitude = longitude,
        latitude = latitude,
        formattedAddress = "",
        province = "",
        city = "",
        district = "",
        street = "",
        streetNumber = "",
        poiName = ""
    )
}

data class AddressResult(
    val longitude: Double,
    val latitude: Double,
    val formattedAddress: String,
    val province: String,
    val city: String,
    val district: String,
    val street: String,
    val streetNumber: String,
    val poiName: String,
    val adCode: String = "",
    val cityCode: String = ""
)

data class RoutePoint(
    val index: Int = 0,
    val longitude: Double,
    val latitude: Double,
    val name: String = "",
    val address: String = ""
)

data class OptimizedRoute(
    val points: List<RoutePoint>,
    val totalDistance: Double = 0.0,
    val totalDuration: Long = 0L
)

class RouteOptimizer {

    fun optimizeRouteNearestNeighbor(start: RoutePoint, points: List<RoutePoint>): OptimizedRoute {
        if (points.isEmpty()) {
            return OptimizedRoute(listOf(start), 0.0, 0L)
        }

        val unvisited = points.toMutableList()
        val result = mutableListOf(start)
        var current = start
        var totalDistance = 0.0
        var totalDuration = 0L

        while (unvisited.isNotEmpty()) {
            var nearestIndex = 0
            var nearestDistance = Double.MAX_VALUE

            unvisited.forEachIndexed { index, point ->
                val distance = calculateDistance(
                    current.longitude, current.latitude,
                    point.longitude, point.latitude
                )
                if (distance < nearestDistance) {
                    nearestDistance = distance
                    nearestIndex = index
                }
            }

            val nextPoint = unvisited.removeAt(nearestIndex)
            totalDistance += nearestDistance
            totalDuration += (nearestDistance / 5.0 / 1000.0 * 60.0 * 60.0).toLong()

            result.add(nextPoint.copy(index = result.size - 1))
            current = nextPoint
        }

        return OptimizedRoute(result, totalDistance, totalDuration)
    }

    fun optimizeRouteTwoOpt(start: RoutePoint, points: List<RoutePoint>): OptimizedRoute {
        val initial = optimizeRouteNearestNeighbor(start, points)
        var route = initial.points.toMutableList()
        var improved = true
        var bestDistance = initial.totalDistance

        while (improved) {
            improved = false

            for (i in 1 until route.size - 1) {
                for (j in i + 1 until route.size - 1) {
                    val newRoute = twoOptSwap(route, i, j)
                    val newDistance = calculateTotalDistance(newRoute)

                    if (newDistance < bestDistance - 0.001) {
                        route = newRoute
                        bestDistance = newDistance
                        improved = true
                    }
                }
            }
        }

        val totalDuration = (bestDistance / 5.0 / 1000.0 * 60.0 * 60.0).toLong()
        route = route.mapIndexed { index, point -> point.copy(index = index) }.toMutableList()

        return OptimizedRoute(route, bestDistance, totalDuration)
    }

    private fun twoOptSwap(route: MutableList<RoutePoint>, i: Int, k: Int): MutableList<RoutePoint> {
        val newRoute = route.toMutableList()
        var a = i
        var b = k
        while (a < b) {
            val temp = newRoute[a]
            newRoute[a] = newRoute[b]
            newRoute[b] = temp
            a++
            b--
        }
        return newRoute
    }

    private fun calculateTotalDistance(route: List<RoutePoint>): Double {
        var total = 0.0
        for (i in 0 until route.size - 1) {
            total += calculateDistance(
                route[i].longitude, route[i].latitude,
                route[i + 1].longitude, route[i + 1].latitude
            )
        }
        return total
    }

    companion object {
        fun calculateDistance(lng1: Double, lat1: Double, lng2: Double, lat2: Double): Double {
            val earthRadius = 6371000.0

            val lat1Rad = lat1 * Math.PI / 180
            val lat2Rad = lat2 * Math.PI / 180
            val deltaLat = (lat2 - lat1) * Math.PI / 180
            val deltaLng = (lng2 - lng1) * Math.PI / 180

            val a = Math.sin(deltaLat / 2) * Math.sin(deltaLat / 2) +
                    Math.cos(lat1Rad) * Math.cos(lat2Rad) *
                    Math.sin(deltaLng / 2) * Math.sin(deltaLng / 2)
            val c = 2 * Math.atan2(Math.sqrt(a), Math.sqrt(1 - a))

            return earthRadius * c
        }
    }
}
