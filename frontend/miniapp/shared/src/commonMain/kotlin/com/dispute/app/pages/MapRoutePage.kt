package com.dispute.app.pages

import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Brush
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.dispute.app.LocalAppState
import com.dispute.app.LocalApiClient
import com.dispute.app.LocalRouter
import com.dispute.app.Route
import com.dispute.app.api.PlanRouteRequest
import com.dispute.app.api.RoutePointRequest
import com.dispute.app.components.AppCard
import com.dispute.app.model.GridTask
import com.dispute.app.model.MapRoute
import com.dispute.app.model.RoutePointInfo
import com.dispute.app.model.TaskPoint
import com.dispute.app.platform.MapMarker as PlatformMapMarker
import com.dispute.app.platform.MapRoutePath
import com.dispute.app.platform.MapType as PlatformMapType
import com.dispute.app.platform.PlatformMapService
import com.dispute.app.utils.RouteOptimizer

private const val MAP_CONTAINER_ID = "patrol-route-map"

@Composable
fun MapRoutePage() = MapRouteContent()

@Composable
private fun MapRouteContent() {
    val appState = LocalAppState.current
    val router = LocalRouter.current
    val apiClient = LocalApiClient.current

    val task = appState.selectedGridTask
    var route by remember { mutableStateOf<MapRoute?>(null) }
    var isLoading by remember { mutableStateOf(true) }
    var selectedPointIndex by remember { mutableStateOf(0) }
    var mapType by remember { mutableStateOf(MapType.NORMAL) }
    var useRealMap by remember { mutableStateOf(false) }
    var fallbackMessage by remember { mutableStateOf<String?>(null) }

    LaunchedEffect(Unit) {
        if (task != null) {
            val points = task.pointList
            if (points.isNotEmpty()) {
                val startLng = appState.currentLocation?.first ?: points.first().longitude
                val startLat = appState.currentLocation?.second ?: points.first().latitude

                val routePointRequests = points.map { p ->
                    RoutePointRequest(
                        pointName = p.name,
                        address = p.address,
                        longitude = p.longitude,
                        latitude = p.latitude
                    )
                }

                try {
                    route = apiClient.gridWorker.planRoute(
                        PlanRouteRequest(
                            startLng = startLng,
                            startLat = startLat,
                            points = routePointRequests,
                            strategy = 10
                        )
                    )
                } catch (e: Exception) {
                    fallbackMessage = "服务不可用，使用本地排序"
                    route = buildLocalFallbackRoute(startLng, startLat, points)
                } finally {
                    isLoading = false
                }

                if (PlatformMapService.isAvailable && PlatformMapService.config.webKey.isNotBlank()) {
                    val sdkLoaded = try {
                        PlatformMapService.loadSDK()
                    } catch (_: Exception) {
                        false
                    }
                    if (sdkLoaded) {
                        useRealMap = true
                    }
                }
            } else {
                isLoading = false
            }
        } else {
            isLoading = false
        }
    }

    LaunchedEffect(route, useRealMap, mapType, selectedPointIndex) {
        if (useRealMap && route != null && task != null) {
            try {
                val routePoints = route!!.points
                if (routePoints.isNotEmpty()) {
                    val centerLng = routePoints.map { it.longitude }.average()
                    val centerLat = routePoints.map { it.latitude }.average()

                    val platformMapType = when (mapType) {
                        MapType.NORMAL -> PlatformMapType.NORMAL
                        MapType.SATELLITE -> PlatformMapType.SATELLITE
                        MapType.TRAFFIC -> PlatformMapType.TRAFFIC
                    }

                    PlatformMapService.renderMap(
                        containerId = MAP_CONTAINER_ID,
                        centerLng = centerLng,
                        centerLat = centerLat,
                        zoom = 14,
                        mapType = platformMapType
                    )

                    val markers = routePoints.mapIndexed { idx, rp ->
                        val taskPoint = task.pointList.getOrNull(rp.originalIndex)
                        PlatformMapMarker(
                            id = "marker-${rp.originalIndex}",
                            longitude = rp.longitude,
                            latitude = rp.latitude,
                            title = rp.pointName,
                            snippet = rp.address,
                            sortIndex = idx,
                            isCompleted = taskPoint?.status == TaskPoint.PointStatus.COMPLETED,
                            isSelected = idx == selectedPointIndex
                        )
                    }

                    PlatformMapService.addMarkers(
                        containerId = MAP_CONTAINER_ID,
                        markers = markers
                    ) { clicked ->
                        val foundIdx = routePoints.indexOfFirst { it.originalIndex.toString() == clicked.id.removePrefix("marker-") }
                        if (foundIdx >= 0) {
                            selectedPointIndex = foundIdx
                        }
                    }

                    val polylinePoints = mutableListOf<Pair<Double, Double>>()
                    val startLng = appState.currentLocation?.first ?: routePoints.first().longitude
                    val startLat = appState.currentLocation?.second ?: routePoints.first().latitude
                    polylinePoints.add(Pair(startLng, startLat))
                    routePoints.forEach { rp ->
                        polylinePoints.add(Pair(rp.longitude, rp.latitude))
                    }

                    val routePath = MapRoutePath(
                        polyline = polylinePoints,
                        distance = route!!.totalDistance,
                        duration = route!!.totalDuration.toLong(),
                        strategy = route!!.strategy,
                        strategyName = route!!.strategyName
                    )

                    PlatformMapService.drawRoute(
                        containerId = MAP_CONTAINER_ID,
                        route = routePath,
                        color = "#1D6CFF",
                        width = 6
                    )

                    PlatformMapService.fitMarkers(MAP_CONTAINER_ID, markers)
                }
            } catch (_: Exception) {
                useRealMap = false
            }
        }
    }

    val sortedPoints = remember(route, task) {
        if (route != null && task != null) {
            route!!.points.mapNotNull { rp ->
                task.pointList.getOrNull(rp.originalIndex)?.let { tp ->
                    Pair(rp, tp)
                }
            }
        } else {
            task?.pointList?.mapIndexed { idx, tp ->
                Pair(
                    RoutePointInfo(
                        originalIndex = idx,
                        sortedIndex = idx,
                        pointName = tp.name,
                        address = tp.address,
                        longitude = tp.longitude,
                        latitude = tp.latitude
                    ),
                    tp
                )
            } ?: emptyList()
        }
    }

    Column(
        modifier = Modifier
            .fillMaxSize()
            .background(MaterialTheme.colorScheme.background)
    ) {
        TopBar(
            title = "路线规划",
            onBack = { router.back() }
        )

        if (task == null) {
            Box(
                modifier = Modifier.fillMaxSize(),
                contentAlignment = Alignment.Center
            ) {
                Text("任务信息不存在", style = MaterialTheme.typography.bodyLarge)
            }
        } else {
            Column(modifier = Modifier.fillMaxSize()) {
                MapViewSection(
                    task = task,
                    route = route,
                    sortedPoints = sortedPoints,
                    selectedPointIndex = selectedPointIndex,
                    mapType = mapType,
                    useRealMap = useRealMap,
                    onMapTypeChange = { mapType = it }
                )

                RouteInfoCard(
                    task = task,
                    route = route,
                    isLoading = isLoading,
                    fallbackMessage = fallbackMessage
                )

                PointListSection(
                    sortedPairs = sortedPoints,
                    selectedPointIndex = selectedPointIndex,
                    onPointSelected = { selectedPointIndex = it },
                    onCheckInClick = { pointId ->
                        appState.setCheckInPoint(task.pointList.firstOrNull { it.id == pointId })
                        router.navigate(Route.CheckIn(task.id, pointId))
                    },
                    onStartNavigation = {
                        appState.showToast("开始导航到${sortedPoints.getOrNull(selectedPointIndex)?.second?.name ?: "目标点"}")
                    }
                )
            }
        }
    }
}

private fun buildLocalFallbackRoute(
    startLng: Double,
    startLat: Double,
    points: List<TaskPoint>
): MapRoute {
    val routePoints = points.mapIndexed { idx, tp ->
        RouteOptimizer.RoutePoint(
            index = idx,
            lng = tp.longitude,
            lat = tp.latitude,
            name = tp.name
        )
    }

    val result = RouteOptimizer.nearestNeighbor(startLng, startLat, routePoints)

    val routePointInfos = result.orderedIndices.mapIndexed { sortedIdx, originalIdx ->
        val tp = points[originalIdx]
        RoutePointInfo(
            originalIndex = originalIdx,
            sortedIndex = sortedIdx,
            pointName = tp.name,
            address = tp.address,
            longitude = tp.longitude,
            latitude = tp.latitude,
            distanceFromPrev = result.pointDistances.getOrElse(sortedIdx) { 0.0 }
        )
    }

    val totalDurationSec = (result.totalDistance / 1.39).toInt()
    val taxiCost = if (result.totalDistance > 0) {
        val km = result.totalDistance / 1000.0
        val cost = 13.0 + (km - 3) * 2.3
        if (cost < 13) 13.0 else cost
    } else 0.0

    return MapRoute(
        totalDistance = result.totalDistance,
        totalDuration = totalDurationSec,
        totalTaxiCost = taxiCost,
        strategy = 10,
        strategyName = "速度优先(最近邻·本地)",
        points = routePointInfos,
        localOptimization = true,
        amapAvailable = false
    )
}

@Composable
private fun TopBar(title: String, onBack: () -> Unit) {
    Row(
        modifier = Modifier
            .fillMaxWidth()
            .height(56.dp)
            .background(Color.White)
            .padding(horizontal = 16.dp),
        verticalAlignment = Alignment.CenterVertically
    ) {
        Text(
            text = "←",
            fontSize = 24.sp,
            modifier = Modifier.clickable(onClick = onBack)
        )
        Spacer(modifier = Modifier.width(8.dp))
        Text(
            text = title,
            style = MaterialTheme.typography.titleMedium,
            fontWeight = FontWeight.SemiBold
        )
    }
}

@Composable
private fun MapViewSection(
    task: GridTask,
    route: MapRoute?,
    sortedPoints: List<Pair<RoutePointInfo, TaskPoint>>,
    selectedPointIndex: Int,
    mapType: MapType,
    useRealMap: Boolean,
    onMapTypeChange: (MapType) -> Unit
) {
    Box(
        modifier = Modifier
            .fillMaxWidth()
            .height(280.dp)
            .background(
                Brush.linearGradient(
                    colors = listOf(
                        Color(0xFFE3F2FD),
                        Color(0xFFBBDEFB)
                    )
                )
            )
    ) {
        Box(
            modifier = Modifier
                .fillMaxSize()
                .padding(16.dp)
        ) {
            if (useRealMap) {
                RealMapView(containerId = MAP_CONTAINER_ID)
            } else {
                val routePointInfoList = sortedPoints.map { it.first }
                val taskPointList = sortedPoints.map { it.second }
                MapMockView(
                    points = taskPointList,
                    selectedPointIndex = selectedPointIndex,
                    route = route,
                    routeInfos = routePointInfoList
                )
            }

            Row(
                modifier = Modifier
                    .align(Alignment.TopEnd)
                    .background(Color.White, RoundedCornerShape(8.dp))
                    .padding(4.dp),
                horizontalArrangement = Arrangement.spacedBy(4.dp)
            ) {
                MapType.values().forEach { type ->
                    val isSelected = mapType == type
                    Box(
                        modifier = Modifier
                            .background(
                                if (isSelected) MaterialTheme.colorScheme.primary else Color.Transparent,
                                RoundedCornerShape(6.dp)
                            )
                            .clickable { onMapTypeChange(type) }
                            .padding(horizontal = 12.dp, vertical = 6.dp)
                    ) {
                        Text(
                            text = type.displayName,
                            style = MaterialTheme.typography.labelSmall,
                            color = if (isSelected) Color.White else MaterialTheme.colorScheme.onSurface
                        )
                    }
                }
            }

            Box(
                modifier = Modifier
                    .align(Alignment.BottomStart)
                    .background(Color.White, RoundedCornerShape(8.dp))
                    .padding(horizontal = 12.dp, vertical = 8.dp)
            ) {
                Row(verticalAlignment = Alignment.CenterVertically) {
                    Text("📍", fontSize = 16.sp)
                    Spacer(modifier = Modifier.width(4.dp))
                    Text(
                        text = if (useRealMap) "高德地图" else "模拟地图",
                        style = MaterialTheme.typography.labelMedium,
                        fontWeight = FontWeight.Medium
                    )
                }
            }
        }
    }
}

@Composable
private fun RealMapView(containerId: String) {
    Box(
        modifier = Modifier
            .fillMaxSize()
            .background(Color(0xFFE8F5E9), RoundedCornerShape(12.dp))
    ) {
        androidx.compose.runtime.DisposableEffect(containerId) {
            onDispose {
                try {
                    PlatformMapService.clearMap(containerId)
                } catch (_: Exception) {}
            }
        }
        Box(
            modifier = Modifier.fillMaxSize(),
            contentAlignment = Alignment.Center
        ) {
            Text(
                text = "🗺️ 高德地图加载中...",
                style = MaterialTheme.typography.bodySmall,
                color = Color(0xFF4CAF50)
            )
        }
    }
}

@Composable
private fun MapMockView(
    points: List<TaskPoint>,
    selectedPointIndex: Int,
    route: MapRoute?,
    routeInfos: List<RoutePointInfo>
) {
    Box(modifier = Modifier.fillMaxSize()) {
        Box(
            modifier = Modifier
                .fillMaxSize()
                .background(
                    Brush.linearGradient(
                        colors = listOf(
                            Color(0xFFE8F5E9),
                            Color(0xFFC8E6C9)
                        )
                    ),
                    RoundedCornerShape(12.dp)
                )
        ) {
            Text(
                text = "🗺️",
                fontSize = 48.sp,
                modifier = Modifier.align(Alignment.Center)
            )

            points.forEachIndexed { index, point ->
                val position = getPointPosition(index, points.size)
                val isSelected = index == selectedPointIndex
                val isCompleted = point.status == TaskPoint.PointStatus.COMPLETED

                Box(
                    modifier = Modifier
                        .align(position.alignment)
                        .padding(position.padding)
                ) {
                    Column(horizontalAlignment = Alignment.CenterHorizontally) {
                        Box(
                            modifier = Modifier
                                .size(if (isSelected) 40.dp else 32.dp)
                                .background(
                                    when {
                                        isCompleted -> Color(0xFF22C55E)
                                        isSelected -> MaterialTheme.colorScheme.primary
                                        else -> Color(0xFF9CA3AF)
                                    },
                                    RoundedCornerShape(if (isSelected) 20.dp else 16.dp)
                                )
                                .padding(4.dp),
                            contentAlignment = Alignment.Center
                        ) {
                            val displayIdx = routeInfos.getOrNull(index)?.let { it.sortedIndex + 1 } ?: (index + 1)
                            Text(
                                text = if (isCompleted) "✓" else "$displayIdx",
                                color = Color.White,
                                style = MaterialTheme.typography.labelMedium,
                                fontWeight = FontWeight.Bold
                            )
                        }
                        Spacer(modifier = Modifier.height(2.dp))
                        Text(
                            text = point.name,
                            style = MaterialTheme.typography.labelSmall,
                            fontWeight = if (isSelected) FontWeight.SemiBold else FontWeight.Normal,
                            color = if (isSelected) MaterialTheme.colorScheme.primary else MaterialTheme.colorScheme.onSurface
                        )
                    }
                }
            }
        }
    }
}

private fun getPointPosition(index: Int, total: Int): PointPosition {
    return when (index) {
        0 -> PointPosition(Alignment.TopStart, androidx.compose.foundation.layout.PaddingValues(top = 20.dp, start = 20.dp))
        1 -> PointPosition(Alignment.TopEnd, androidx.compose.foundation.layout.PaddingValues(top = 40.dp, end = 30.dp))
        2 -> PointPosition(Alignment.CenterStart, androidx.compose.foundation.layout.PaddingValues(start = 40.dp))
        3 -> PointPosition(Alignment.CenterEnd, androidx.compose.foundation.layout.PaddingValues(end = 20.dp))
        4 -> PointPosition(Alignment.BottomStart, androidx.compose.foundation.layout.PaddingValues(bottom = 30.dp, start = 30.dp))
        else -> PointPosition(Alignment.BottomEnd, androidx.compose.foundation.layout.PaddingValues(bottom = 20.dp, end = 20.dp))
    }
}

private data class PointPosition(
    val alignment: Alignment,
    val padding: androidx.compose.foundation.layout.PaddingValues
)

@Composable
private fun RouteInfoCard(
    task: GridTask,
    route: MapRoute?,
    isLoading: Boolean,
    fallbackMessage: String?
) {
    AppCard(
        modifier = Modifier.padding(horizontal = 16.dp, vertical = 12.dp)
    ) {
        Column(verticalArrangement = Arrangement.spacedBy(12.dp)) {
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.SpaceBetween,
                verticalAlignment = Alignment.CenterVertically
            ) {
                Text(
                    text = "路线概览",
                    style = MaterialTheme.typography.titleSmall,
                    fontWeight = FontWeight.SemiBold
                )
                if (route != null) {
                    Text(
                        text = route.strategyName,
                        style = MaterialTheme.typography.labelSmall,
                        color = MaterialTheme.colorScheme.primary
                    )
                }
            }

            if (fallbackMessage != null) {
                Row(
                    modifier = Modifier
                        .fillMaxWidth()
                        .background(Color(0xFFFFF8E1), RoundedCornerShape(8.dp))
                        .padding(horizontal = 12.dp, vertical = 8.dp),
                    verticalAlignment = Alignment.CenterVertically
                ) {
                    Text("⚠️", fontSize = 14.sp)
                    Spacer(modifier = Modifier.width(6.dp))
                    Text(
                        text = fallbackMessage,
                        style = MaterialTheme.typography.labelSmall,
                        color = Color(0xFFF57C00)
                    )
                }
            }

            if (isLoading) {
                Row(
                    modifier = Modifier.fillMaxWidth(),
                    horizontalArrangement = Arrangement.Center
                ) {
                    androidx.compose.material3.CircularProgressIndicator(
                        modifier = Modifier.size(24.dp),
                        strokeWidth = 2.dp
                    )
                }
            } else if (route != null) {
                Row(
                    modifier = Modifier.fillMaxWidth(),
                    horizontalArrangement = Arrangement.SpaceEvenly
                ) {
                    RouteInfoItem(
                        icon = "📍",
                        label = "点位数量",
                        value = "${task.pointList.size}个"
                    )
                    RouteInfoItem(
                        icon = "📏",
                        label = "总距离",
                        value = route.formattedDistance
                    )
                    RouteInfoItem(
                        icon = "⏱️",
                        label = "预计时间",
                        value = route.formattedDuration
                    )
                }

                if (route.totalTaxiCost > 0) {
                    Spacer(modifier = Modifier.height(4.dp))
                    Row(
                        modifier = Modifier.fillMaxWidth(),
                        horizontalArrangement = Arrangement.Center
                    ) {
                        Text(
                            text = "打车费用约 ¥${String.format("%.2f", route.totalTaxiCost)}",
                            style = MaterialTheme.typography.labelSmall,
                            color = Color(0xFF78909C)
                        )
                    }
                }

                Spacer(modifier = Modifier.height(4.dp))

                Row(
                    modifier = Modifier.fillMaxWidth(),
                    verticalAlignment = Alignment.CenterVertically
                ) {
                    Text("✅", fontSize = 16.sp)
                    Spacer(modifier = Modifier.width(8.dp))
                    Text(
                        text = if (route.localOptimization)
                            "已使用本地最优算法规划路线，建议按顺序走访"
                        else
                            "已规划最优路线，建议按顺序走访",
                        style = MaterialTheme.typography.bodySmall,
                        color = Color(0xFF22C55E)
                    )
                }
            }
        }
    }
}

@Composable
private fun RouteInfoItem(
    icon: String,
    label: String,
    value: String
) {
    Column(horizontalAlignment = Alignment.CenterHorizontally) {
        Text(icon, fontSize = 24.sp)
        Spacer(modifier = Modifier.height(4.dp))
        Text(
            text = value,
            style = MaterialTheme.typography.titleSmall,
            fontWeight = FontWeight.SemiBold
        )
        Text(
            text = label,
            style = MaterialTheme.typography.labelSmall,
            color = MaterialTheme.colorScheme.onSurfaceVariant
        )
    }
}

@Composable
private fun PointListSection(
    sortedPairs: List<Pair<RoutePointInfo, TaskPoint>>,
    selectedPointIndex: Int,
    onPointSelected: (Int) -> Unit,
    onCheckInClick: (String) -> Unit,
    onStartNavigation: () -> Unit
) {
    Column(
        modifier = Modifier
            .fillMaxWidth()
            .weight(1f)
    ) {
        Row(
            modifier = Modifier
                .fillMaxWidth()
                .padding(horizontal = 16.dp, vertical = 8.dp),
            verticalAlignment = Alignment.CenterVertically,
            horizontalArrangement = Arrangement.SpaceBetween
        ) {
            Text(
                text = "点位列表（已排序）",
                style = MaterialTheme.typography.titleSmall,
                fontWeight = FontWeight.SemiBold
            )
            Text(
                text = "${sortedPairs.count { it.second.status == TaskPoint.PointStatus.COMPLETED }}/${sortedPairs.size} 已完成",
                style = MaterialTheme.typography.bodySmall,
                color = MaterialTheme.colorScheme.onSurfaceVariant
            )
        }

        LazyColumn(
            modifier = Modifier.weight(1f),
            contentPadding = androidx.compose.foundation.layout.PaddingValues(horizontal = 16.dp, vertical = 8.dp),
            verticalArrangement = Arrangement.spacedBy(8.dp)
        ) {
            items(sortedPairs) { pair ->
                val (rp, tp) = pair
                val index = rp.sortedIndex
                val isSelected = index == selectedPointIndex
                PointListItem(
                    routeInfo = rp,
                    point = tp,
                    index = index,
                    isSelected = isSelected,
                    onClick = { onPointSelected(index) },
                    onCheckInClick = { onCheckInClick(tp.id) }
                )
            }
        }

        BottomActionButton(
            text = "导航到当前点位",
            onClick = onStartNavigation
        )
    }
}

@Composable
private fun PointListItem(
    routeInfo: RoutePointInfo,
    point: TaskPoint,
    index: Int,
    isSelected: Boolean,
    onClick: () -> Unit,
    onCheckInClick: () -> Unit
) {
    val isCompleted = point.status == TaskPoint.PointStatus.COMPLETED

    Box(
        modifier = Modifier
            .fillMaxWidth()
            .background(
                if (isSelected)
                    MaterialTheme.colorScheme.primary.copy(alpha = 0.08f)
                else
                    MaterialTheme.colorScheme.surface,
                RoundedCornerShape(12.dp)
            )
            .clickable(onClick = onClick)
            .padding(12.dp)
    ) {
        Row(
            modifier = Modifier.fillMaxWidth(),
            verticalAlignment = Alignment.CenterVertically
        ) {
            Box(
                modifier = Modifier
                    .size(36.dp)
                    .background(
                        when {
                            isCompleted -> Color(0xFF22C55E)
                            isSelected -> MaterialTheme.colorScheme.primary
                            else -> MaterialTheme.colorScheme.surfaceVariant
                        },
                        RoundedCornerShape(18.dp)
                    ),
                contentAlignment = Alignment.Center
            ) {
                Text(
                    text = if (isCompleted) "✓" else "${index + 1}",
                    color = if (isCompleted || isSelected) Color.White else MaterialTheme.colorScheme.onSurface,
                    style = MaterialTheme.typography.labelMedium,
                    fontWeight = FontWeight.Bold
                )
            }

            Spacer(modifier = Modifier.width(12.dp))

            Column(modifier = Modifier.weight(1f)) {
                Text(
                    text = point.name,
                    style = MaterialTheme.typography.titleSmall,
                    fontWeight = FontWeight.SemiBold
                )
                Spacer(modifier = Modifier.height(2.dp))
                Text(
                    text = point.address,
                    style = MaterialTheme.typography.bodySmall,
                    color = MaterialTheme.colorScheme.onSurfaceVariant
                )
                if (routeInfo.distanceFromPrev > 0 && index > 0) {
                    Spacer(modifier = Modifier.height(2.dp))
                    val distStr = if (routeInfo.distanceFromPrev >= 1000)
                        "${String.format("%.2f", routeInfo.distanceFromPrev / 1000)} 公里"
                    else
                        "${String.format("%.0f", routeInfo.distanceFromPrev)} 米"
                    Text(
                        text = "距上一点：$distStr",
                        style = MaterialTheme.typography.labelSmall,
                        color = Color(0xFF78909C)
                    )
                }
                if (point.status == TaskPoint.PointStatus.COMPLETED && point.checkInTime != null) {
                    Spacer(modifier = Modifier.height(2.dp))
                    Text(
                        text = "已签到: ${point.checkInTime}",
                        style = MaterialTheme.typography.labelSmall,
                        color = Color(0xFF22C55E)
                    )
                }
            }

            Spacer(modifier = Modifier.width(8.dp))

            if (point.status != TaskPoint.PointStatus.COMPLETED) {
                Box(
                    modifier = Modifier
                        .background(
                            MaterialTheme.colorScheme.primary,
                            RoundedCornerShape(16.dp)
                        )
                        .clickable(onClick = onCheckInClick)
                        .padding(horizontal = 12.dp, vertical = 6.dp)
                ) {
                    Text(
                        text = "签到",
                        color = Color.White,
                        style = MaterialTheme.typography.labelSmall,
                        fontWeight = FontWeight.SemiBold
                    )
                }
            } else {
                Box(
                    modifier = Modifier
                        .background(
                            Color(0xFF22C55E).copy(alpha = 0.1f),
                            RoundedCornerShape(16.dp)
                        )
                        .padding(horizontal = 12.dp, vertical = 6.dp)
                ) {
                    Text(
                        text = "已完成",
                        color = Color(0xFF22C55E),
                        style = MaterialTheme.typography.labelSmall,
                        fontWeight = FontWeight.SemiBold
                    )
                }
            }
        }
    }
}

@Composable
private fun BottomActionButton(
    text: String,
    onClick: () -> Unit
) {
    Box(
        modifier = Modifier
            .fillMaxWidth()
            .padding(horizontal = 16.dp, vertical = 16.dp)
    ) {
        Box(
            modifier = Modifier
                .fillMaxWidth()
                .background(
                    Brush.linearGradient(
                        colors = listOf(
                            Color(0xFF10B981),
                            Color(0xFF34D399)
                        )
                    ),
                    RoundedCornerShape(28.dp)
                )
                .clickable(onClick = onClick)
                .padding(vertical = 16.dp),
            contentAlignment = Alignment.Center
        ) {
            Row(verticalAlignment = Alignment.CenterVertically) {
                Text("🧭", fontSize = 20.sp)
                Spacer(modifier = Modifier.width(8.dp))
                Text(
                    text = text,
                    color = Color.White,
                    style = MaterialTheme.typography.titleMedium,
                    fontWeight = FontWeight.SemiBold
                )
            }
        }
    }
}

enum class MapType(val displayName: String) {
    NORMAL("标准"),
    SATELLITE("卫星"),
    TRAFFIC("路况")
}
