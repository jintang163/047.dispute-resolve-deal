package com.dispute.app.platform

import com.dispute.app.api.ApiClient
import com.dispute.app.api.ApiResponse
import io.ktor.client.request.get
import io.ktor.client.statement.bodyAsText
import kotlinx.serialization.Serializable
import kotlinx.serialization.json.Json

actual class AMapService actual constructor() {

    private var apiKey: String = ""
    private val json = Json {
        prettyPrint = true
        isLenient = true
        ignoreUnknownKeys = true
        encodeDefaults = true
    }

    actual suspend fun planDrivingRoute(
        startLng: Double,
        startLat: Double,
        points: List<RoutePoint>,
        strategy: Int
    ): AMapRouteResult {
        if (points.isEmpty()) {
            return AMapRouteResult(
                orderedPoints = listOf(RoutePoint(longitude = startLng, latitude = startLat, index = 0))
            )
        }

        val optimizer = RouteOptimizer()
        val startPoint = RoutePoint(
            longitude = startLng,
            latitude = startLat,
            name = "起点",
            address = "当前位置",
            index = 0
        )
        val optimized = optimizer.optimizeRouteTwoOpt(startPoint, points)

        val totalDistance = optimized.totalDistance
        val totalDuration = optimized.totalDuration

        val ordered = optimized.points.mapIndexed { idx, point ->
            point.copy(index = idx)
        }

        val routeSteps = mutableListOf<AMapStep>()
        for (i in 0 until ordered.size - 1) {
            val from = ordered[i]
            val to = ordered[i + 1]
            val dist = RouteOptimizer.calculateDistance(
                from.longitude, from.latitude,
                to.longitude, to.latitude
            )
            routeSteps.add(
                AMapStep(
                    instruction = "从${from.name}前往${to.name}",
                    road = to.address,
                    distance = dist.toInt(),
                    duration = (dist / 1.4 / 60.0).toInt(),
                    polyline = "${from.longitude},${from.latitude};${to.longitude},${to.latitude}",
                    action = "前往",
                    orientation = calculateBearing(from, to)
                )
            )
        }

        return AMapRouteResult(
            totalDistance = totalDistance.toInt(),
            totalDuration = (totalDuration / 1000.0).toInt(),
            totalTaxiCost = (totalDistance / 1000.0 * 2.5f).toFloat(),
            strategy = strategy,
            paths = listOf(
                AMapPath(
                    distance = totalDistance.toInt(),
                    duration = (totalDuration / 1000.0).toInt(),
                    strategy = "最优路线",
                    steps = routeSteps,
                    polyline = ordered.joinToString(";") { "${it.longitude},${it.latitude}" }
                )
            ),
            orderedPoints = ordered
        )
    }

    actual suspend fun planWalkingRoute(
        startLng: Double,
        startLat: Double,
        endLng: Double,
        endLat: Double
    ): AMapRouteResult {
        val distance = RouteOptimizer.calculateDistance(startLng, startLat, endLng, endLat)
        val duration = (distance / 1.4).toInt()

        return AMapRouteResult(
            totalDistance = distance.toInt(),
            totalDuration = duration,
            totalTaxiCost = 0f,
            strategy = 10,
            paths = listOf(
                AMapPath(
                    distance = distance.toInt(),
                    duration = duration,
                    strategy = "步行路线",
                    steps = listOf(
                        AMapStep(
                            instruction = "步行前往目的地",
                            road = "",
                            distance = distance.toInt(),
                            duration = duration,
                            polyline = "$startLng,$startLat;$endLng,$endLat",
                            action = "步行",
                            orientation = ""
                        )
                    ),
                    polyline = "$startLng,$startLat;$endLng,$endLat"
                )
            )
        )
    }

    actual suspend fun reverseGeocode(
        longitude: Double,
        latitude: Double
    ): AddressResult {
        if (apiKey.isBlank()) {
            return mockReverseGeocode(longitude, latitude)
        }

        return try {
            val url = "https://restapi.amap.com/v3/geocode/regeo" +
                    "?location=$longitude,$latitude" +
                    "&key=$apiKey&radius=1000&extensions=all"
            val client = createDefaultClient()
            val responseText = client.get(url).bodyAsText()
            parseRegeoResponse(longitude, latitude, responseText)
        } catch (e: Exception) {
            mockReverseGeocode(longitude, latitude)
        }
    }

    actual suspend fun geocode(address: String, city: String?): GeoCodeResult {
        if (apiKey.isBlank()) {
            return mockGeocode(address)
        }

        return try {
            val url = "https://restapi.amap.com/v3/geocode/geo" +
                    "?address=$address" +
                    (city?.let { "&city=$it" } ?: "") +
                    "&key=$apiKey"
            val client = createDefaultClient()
            val responseText = client.get(url).bodyAsText()
            parseGeoResponse(address, responseText)
        } catch (e: Exception) {
            mockGeocode(address)
        }
    }

    actual fun calculateLineDistance(
        lng1: Double,
        lat1: Double,
        lng2: Double,
        lat2: Double
    ): Double {
        return RouteOptimizer.calculateDistance(lng1, lat1, lng2, lat2)
    }

    actual fun setApiKey(key: String) {
        apiKey = key
    }

    actual fun release() {
        apiKey = ""
    }

    private fun calculateBearing(from: RoutePoint, to: RoutePoint): String {
        val lngDiff = to.longitude - from.longitude
        val latDiff = to.latitude - from.latitude
        return when {
            Math.abs(lngDiff) < 0.0001 && latDiff > 0 -> "北"
            Math.abs(lngDiff) < 0.0001 && latDiff < 0 -> "南"
            Math.abs(latDiff) < 0.0001 && lngDiff > 0 -> "东"
            Math.abs(latDiff) < 0.0001 && lngDiff < 0 -> "西"
            lngDiff > 0 && latDiff > 0 -> "东北"
            lngDiff > 0 && latDiff < 0 -> "东南"
            lngDiff < 0 && latDiff > 0 -> "西北"
            else -> "西南"
        }
    }

    private fun mockReverseGeocode(longitude: Double, latitude: Double): AddressResult {
        val (district, street, streetNum) = listOf(
            listOf("朝阳区", "建国路", "88号"),
            listOf("海淀区", "中关村大街", "1号"),
            listOf("东城区", "王府井大街", "201号"),
            listOf("西城区", "金融街", "35号"),
            listOf("丰台区", "西三环南路", "16号")
        ).random()

        return AddressResult(
            longitude = longitude,
            latitude = latitude,
            formattedAddress = "北京市$district$street$streetNum",
            province = "北京市",
            city = "北京市",
            district = district,
            street = street,
            streetNumber = streetNum,
            poiName = "${street}街道办事处",
            adCode = "110105",
            cityCode = "010"
        )
    }

    private fun mockGeocode(address: String): GeoCodeResult {
        val baseLng = 116.0 + Math.random() * 2.0
        val baseLat = 39.0 + Math.random() * 2.0
        return GeoCodeResult(
            longitude = baseLng,
            latitude = baseLat,
            formattedAddress = address,
            level = "门牌号"
        )
    }

    private fun parseRegeoResponse(lng: Double, lat: Double, text: String): AddressResult {
        return try {
            val wrapper = json.decodeFromString(RegeoResponseWrapper.serializer(), text)
            val regeo = wrapper.regeocode
            val addr = regeo?.addressComponent
            AddressResult(
                longitude = lng,
                latitude = lat,
                formattedAddress = regeo?.formatted_address ?: "",
                province = addr?.province ?: "",
                city = addr?.city?.takeIf { it.isNotBlank() } ?: addr?.province ?: "",
                district = addr?.district ?: "",
                street = addr?.township ?: addr?.street?.firstOrNull() ?: "",
                streetNumber = addr?.streetNumber?.number ?: "",
                poiName = regeo?.pois?.firstOrNull()?.name ?: "",
                adCode = addr?.adcode ?: "",
                cityCode = addr?.citycode ?: ""
            )
        } catch (_: Exception) {
            mockReverseGeocode(lng, lat)
        }
    }

    private fun parseGeoResponse(address: String, text: String): GeoCodeResult {
        return try {
            val wrapper = json.decodeFromString(GeoResponseWrapper.serializer(), text)
            val loc = wrapper.geocodes?.firstOrNull()?.location
            if (loc != null && loc.contains(",")) {
                val parts = loc.split(",")
                GeoCodeResult(
                    longitude = parts[0].toDouble(),
                    latitude = parts[1].toDouble(),
                    formattedAddress = wrapper.geocodes?.firstOrNull()?.formatted_address ?: address,
                    level = wrapper.geocodes?.firstOrNull()?.level ?: ""
                )
            } else {
                mockGeocode(address)
            }
        } catch (_: Exception) {
            mockGeocode(address)
        }
    }

    private fun createDefaultClient(): io.ktor.client.HttpClient {
        return io.ktor.client.HttpClient {
            install(io.ktor.client.plugins.contentnegotiation.ContentNegotiation) {
                json(json)
            }
        }
    }
}

actual fun isAMapServiceSupported(): Boolean = true

@Serializable
private data class RegeoResponseWrapper(
    val status: String? = null,
    val regeocode: RegeoCode? = null
)

@Serializable
private data class RegeoCode(
    val formatted_address: String? = null,
    val addressComponent: AddressComponent? = null,
    val pois: List<RegeoPoi>? = null
)

@Serializable
private data class AddressComponent(
    val province: String? = null,
    val city: String? = null,
    val district: String? = null,
    val citycode: String? = null,
    val adcode: String? = null,
    val township: String? = null,
    val towncode: String? = null,
    val neighborhood: Neighborhood? = null,
    val building: Building? = null,
    val street: List<StreetInfo>? = null,
    val streetNumber: StreetNumber? = null
)

@Serializable
private data class Neighborhood(val name: String? = null)

@Serializable
private data class Building(val name: String? = null)

@Serializable
private data class StreetInfo(val id: String? = null, val name: String? = null)

@Serializable
private data class StreetNumber(val number: String? = null, val location: String? = null)

@Serializable
private data class RegeoPoi(val id: String? = null, val name: String? = null)

@Serializable
private data class GeoResponseWrapper(
    val status: String? = null,
    val geocodes: List<GeoCodeItem>? = null
)

@Serializable
private data class GeoCodeItem(
    val formatted_address: String? = null,
    val location: String? = null,
    val level: String? = null
)
