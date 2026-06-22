package com.dispute.app.utils

import kotlin.math.*

object RouteOptimizer {
    private const val EARTH_RADIUS = 6371000.0

    data class RoutePoint(
        val index: Int,
        val lng: Double,
        val lat: Double,
        val name: String = ""
    )

    data class OptimizedResult(
        val orderedIndices: List<Int>,
        val totalDistance: Double,
        val pointDistances: List<Double>
    )

    fun haversine(lng1: Double, lat1: Double, lng2: Double, lat2: Double): Double {
        val dLat = (lat2 - lat1) * PI / 180
        val dLng = (lng2 - lng1) * PI / 180
        val a = sin(dLat / 2).pow(2) +
                cos(lat1 * PI / 180) * cos(lat2 * PI / 180) *
                sin(dLng / 2).pow(2)
        val c = 2 * atan2(sqrt(a), sqrt(1 - a))
        return EARTH_RADIUS * c
    }

    fun nearestNeighbor(
        startLng: Double,
        startLat: Double,
        points: List<RoutePoint>
    ): OptimizedResult {
        val n = points.size
        if (n == 0) {
            return OptimizedResult(emptyList(), 0.0, emptyList())
        }
        if (n == 1) {
            val dist = haversine(startLng, startLat, points[0].lng, points[0].lat)
            return OptimizedResult(listOf(0), dist, listOf(dist))
        }

        val visited = BooleanArray(n)
        val orderedIndices = mutableListOf<Int>()
        val pointDistances = mutableListOf<Double>()
        var totalDistance = 0.0
        var currentLng = startLng
        var currentLat = startLat

        for (i in 0 until n) {
            var nearestIdx = -1
            var minDist = Double.MAX_VALUE
            for (j in 0 until n) {
                if (!visited[j]) {
                    val d = haversine(currentLng, currentLat, points[j].lng, points[j].lat)
                    if (d < minDist) {
                        minDist = d
                        nearestIdx = j
                    }
                }
            }
            if (nearestIdx >= 0) {
                visited[nearestIdx] = true
                orderedIndices.add(nearestIdx)
                pointDistances.add(minDist)
                totalDistance += minDist
                currentLng = points[nearestIdx].lng
                currentLat = points[nearestIdx].lat
            }
        }

        return twoOpt(startLng, startLat, points, orderedIndices, pointDistances, totalDistance)
    }

    private fun twoOpt(
        startLng: Double,
        startLat: Double,
        points: List<RoutePoint>,
        orderedIndices: List<Int>,
        pointDistances: List<Double>,
        initialTotalDistance: Double
    ): OptimizedResult {
        val n = orderedIndices.size
        if (n < 4) {
            return OptimizedResult(orderedIndices, initialTotalDistance, pointDistances)
        }

        val bestOrder = orderedIndices.toMutableList()
        var bestDist = calculateRouteDistance(startLng, startLat, points, bestOrder)
        var improved = true

        while (improved) {
            improved = false
            for (i in 0 until n - 1) {
                for (j in i + 1 until n) {
                    val newOrder = bestOrder.toMutableList()
                    var left = i
                    var right = j
                    while (left < right) {
                        val temp = newOrder[left]
                        newOrder[left] = newOrder[right]
                        newOrder[right] = temp
                        left++
                        right--
                    }
                    val newDist = calculateRouteDistance(startLng, startLat, points, newOrder)
                    if (newDist < bestDist - 1e-6) {
                        bestDist = newDist
                        bestOrder.clear()
                        bestOrder.addAll(newOrder)
                        improved = true
                    }
                }
            }
        }

        val finalDistances = calculatePointDistances(startLng, startLat, points, bestOrder)
        return OptimizedResult(bestOrder, bestDist, finalDistances)
    }

    private fun calculateRouteDistance(
        startLng: Double,
        startLat: Double,
        points: List<RoutePoint>,
        order: List<Int>
    ): Double {
        if (order.isEmpty()) return 0.0
        var total = haversine(startLng, startLat, points[order[0]].lng, points[order[0]].lat)
        for (i in 1 until order.size) {
            total += haversine(
                points[order[i - 1]].lng, points[order[i - 1]].lat,
                points[order[i]].lng, points[order[i]].lat
            )
        }
        return total
    }

    private fun calculatePointDistances(
        startLng: Double,
        startLat: Double,
        points: List<RoutePoint>,
        order: List<Int>
    ): List<Double> {
        val result = mutableListOf<Double>()
        if (order.isEmpty()) return result
        result.add(haversine(startLng, startLat, points[order[0]].lng, points[order[0]].lat))
        for (i in 1 until order.size) {
            result.add(
                haversine(
                    points[order[i - 1]].lng, points[order[i - 1]].lat,
                    points[order[i]].lng, points[order[i]].lat
                )
            )
        }
        return result
    }
}
