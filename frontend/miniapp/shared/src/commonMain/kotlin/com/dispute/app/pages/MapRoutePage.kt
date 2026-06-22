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
import androidx.compose.runtime.CompositionLocalProvider
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
import com.dispute.app.components.AppCard
import com.dispute.app.model.GridTask
import com.dispute.app.model.MapRoute
import com.dispute.app.model.TaskPoint
import kotlinx.coroutines.launch

@Composable
fun MapRoutePage() {
    val appState = androidx.compose.runtime.remember { com.dispute.app.AppState() }
    val router = androidx.compose.runtime.remember { com.dispute.app.Router(appState) }
    val apiClient = androidx.compose.runtime.remember { com.dispute.app.api.ApiClient() }

    CompositionLocalProvider(
        LocalAppState provides appState,
        LocalRouter provides router,
        LocalApiClient provides apiClient
    ) {
        MapRouteContent()
    }
}

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

    LaunchedEffect(Unit) {
        if (task != null) {
            try {
                route = apiClient.gridWorker.planRoute(task.id)
            } catch (e: Exception) {
                appState.showToast("加载路线失败: ${e.message}")
            } finally {
                isLoading = false
            }
        } else {
            isLoading = false
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
                    selectedPointIndex = selectedPointIndex,
                    mapType = mapType,
                    onMapTypeChange = { mapType = it }
                )

                RouteInfoCard(
                    task = task,
                    route = route,
                    isLoading = isLoading
                )

                PointListSection(
                    points = task.pointList,
                    selectedPointIndex = selectedPointIndex,
                    onPointSelected = { selectedPointIndex = it },
                    onCheckInClick = { pointId ->
                        appState.setCheckInPoint(task.pointList.firstOrNull { it.id == pointId })
                        router.navigate(Route.CheckIn(task.id, pointId))
                    },
                    onStartNavigation = {
                        appState.showToast("开始导航到${task.pointList[selectedPointIndex].name}")
                    }
                )
            }
        }
    }
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
    selectedPointIndex: Int,
    mapType: MapType,
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
            MapMockView(
                points = task.pointList,
                selectedPointIndex = selectedPointIndex,
                route = route
            )

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
                        text = "高德地图",
                        style = MaterialTheme.typography.labelMedium,
                        fontWeight = FontWeight.Medium
                    )
                }
            }
        }
    }
}

@Composable
private fun MapMockView(
    points: List<TaskPoint>,
    selectedPointIndex: Int,
    route: MapRoute?
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
                            Text(
                                text = if (isCompleted) "✓" else "${index + 1}",
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
    isLoading: Boolean
) {
    AppCard(
        modifier = Modifier.padding(horizontal = 16.dp, vertical = 12.dp)
    ) {
        Column(verticalArrangement = Arrangement.spacedBy(12.dp)) {
            Text(
                text = "路线概览",
                style = MaterialTheme.typography.titleSmall,
                fontWeight = FontWeight.SemiBold
            )

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
                        value = route.totalDistance
                    )
                    RouteInfoItem(
                        icon = "⏱️",
                        label = "预计时间",
                        value = route.estimatedTime
                    )
                }

                Spacer(modifier = Modifier.height(4.dp))

                Row(
                    modifier = Modifier.fillMaxWidth(),
                    verticalAlignment = Alignment.CenterVertically
                ) {
                    Text("✅", fontSize = 16.sp)
                    Spacer(modifier = Modifier.width(8.dp))
                    Text(
                        text = "已规划最优路线，建议按顺序走访",
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
    points: List<TaskPoint>,
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
                text = "点位列表",
                style = MaterialTheme.typography.titleSmall,
                fontWeight = FontWeight.SemiBold
            )
            Text(
                text = "${points.count { it.status == TaskPoint.PointStatus.COMPLETED }}/${points.size} 已完成",
                style = MaterialTheme.typography.bodySmall,
                color = MaterialTheme.colorScheme.onSurfaceVariant
            )
        }

        LazyColumn(
            modifier = Modifier.weight(1f),
            contentPadding = androidx.compose.foundation.layout.PaddingValues(horizontal = 16.dp, vertical = 8.dp),
            verticalArrangement = Arrangement.spacedBy(8.dp)
        ) {
            items(points) { point ->
                val index = points.indexOf(point)
                val isSelected = index == selectedPointIndex
                PointListItem(
                    point = point,
                    index = index,
                    isSelected = isSelected,
                    onClick = { onPointSelected(index) },
                    onCheckInClick = { onCheckInClick(point.id) }
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
