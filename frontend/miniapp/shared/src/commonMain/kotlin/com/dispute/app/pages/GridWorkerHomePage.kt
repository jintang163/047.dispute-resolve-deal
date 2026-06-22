package com.dispute.app.pages

import androidx.compose.foundation.Image
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
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.CompositionLocalProvider
import androidx.compose.runtime.getValue
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
import com.dispute.app.model.GridWorkerMockData
import com.dispute.app.model.PointRecord
import com.dispute.app.model.TaskPoint
import kotlinx.coroutines.launch
import com.dispute.app.components.EmptyCard

@Composable
fun GridWorkerHomePage() = GridWorkerHomeContent()

@Composable
private fun GridWorkerHomeContent() {
    val appState = LocalAppState.current
    val router = LocalRouter.current
    val apiClient = LocalApiClient.current
    val gridWorker by appState.gridWorker
    val taskList by appState.gridTaskList

    androidx.compose.runtime.LaunchedEffect(Unit) {
        if (gridWorker == null) {
            appState.setGridWorker(GridWorkerMockData.mockGridWorker)
        }
        if (taskList.isEmpty()) {
            appState.setGridTaskList(GridWorkerMockData.mockTasks)
        }
    }

    val pendingTasks = taskList.count { it.status == GridTask.TaskStatus.PENDING }
    val inProgressTasks = taskList.count { it.status == GridTask.TaskStatus.IN_PROGRESS }
    val completedTasks = taskList.count { it.status == GridTask.TaskStatus.COMPLETED }

    LazyColumn(
        modifier = Modifier
            .fillMaxSize()
            .background(MaterialTheme.colorScheme.background),
        verticalArrangement = Arrangement.spacedBy(16.dp)
    ) {
        item {
            GridWorkerHeaderSection(
                workerName = gridWorker?.realName ?: "网格员",
                level = gridWorker?.level ?: "初级网格员",
                gridName = gridWorker?.gridName ?: "未知网格",
                points = gridWorker?.points ?: 0,
                onProfileClick = { router.navigate(Route.Profile) }
            )
        }

        item {
            TodayCheckInCard(
                hasCheckedIn = false,
                checkInCount = 0,
                onCheckInClick = {
                    val pendingTask = taskList.firstOrNull { it.status == GridTask.TaskStatus.IN_PROGRESS }
                        ?: taskList.firstOrNull { it.status == GridTask.TaskStatus.PENDING }
                    pendingTask?.let { task ->
                        val pendingPoint = task.pointList.firstOrNull { it.checkInStatus == TaskPoint.CheckInStatus.PENDING }
                        pendingPoint?.let { point ->
                            router.navigate(Route.CheckIn(task.id, point.id))
                        }
                    }
                }
            )
        }

        item {
            PointsOverviewCard(
                points = gridWorker?.points ?: 0,
                todayEarned = 0,
                onPointsClick = { router.navigate(Route.PointCenter) }
            )
        }

        item {
            QuickActionsGrid(
                onTaskList = { router.navigate(Route.GridTaskList) },
                onVisitRecord = { router.navigate(Route.VisitRecordList) },
                onHazardReport = { router.navigate(Route.HazardReport) },
                onGiftMall = { router.navigate(Route.GiftMall) }
            )
        }

        item {
            TaskStatsRow(
                pendingCount = pendingTasks,
                inProgressCount = inProgressTasks,
                completedCount = completedTasks,
                onPendingClick = {
                    appState.setGridTaskStatusFilter(GridTask.TaskStatus.PENDING)
                    router.navigate(Route.GridTaskList)
                },
                onInProgressClick = {
                    appState.setGridTaskStatusFilter(GridTask.TaskStatus.IN_PROGRESS)
                    router.navigate(Route.GridTaskList)
                },
                onCompletedClick = {
                    appState.setGridTaskStatusFilter(GridTask.TaskStatus.COMPLETED)
                    router.navigate(Route.GridTaskList)
                }
            )
        }

        item {
            RecentTasksHeader(onSeeAll = { router.navigate(Route.GridTaskList) })
        }

        items(taskList.take(3), key = { it.id }) { task ->
            TaskMiniCard(
                task = task,
                onClick = {
                    appState.setSelectedGridTask(task)
                    router.navigate(Route.GridTaskDetail(task.id))
                }
            )
        }

        if (taskList.isEmpty()) {
            item {
                EmptyCard(
                    icon = "📋",
                    title = "暂无任务",
                    description = "请等待任务分配或联系管理员"
                )
            }
        }

        item {
            Spacer(modifier = Modifier.height(32.dp))
        }
    }
}

@Composable
private fun GridWorkerHeaderSection(
    workerName: String,
    level: String,
    gridName: String,
    points: Int,
    onProfileClick: () -> Unit
) {
    Box(
        modifier = Modifier
            .fillMaxWidth()
            .height(180.dp)
            .background(
                Brush.linearGradient(
                    colors = listOf(
                        Color(0xFF1D6CFF),
                        Color(0xFF4D8CFF)
                    )
                )
            )
    ) {
        Column(
            modifier = Modifier
                .fillMaxSize()
                .padding(horizontal = 20.dp, vertical = 24.dp),
            verticalArrangement = Arrangement.Center
        ) {
            Row(
                modifier = Modifier.fillMaxWidth(),
                verticalAlignment = Alignment.CenterVertically
            ) {
                Box(
                    modifier = Modifier
                        .size(64.dp)
                        .background(Color.White.copy(alpha = 0.25f), CircleShape),
                    contentAlignment = Alignment.Center
                ) {
                    Text(
                        text = workerName.firstOrNull()?.toString() ?: "U",
                        color = Color.White,
                        fontSize = 28.sp,
                        fontWeight = FontWeight.Bold
                    )
                }
                Spacer(modifier = Modifier.width(14.dp))
                Column(modifier = Modifier.weight(1f)) {
                    Text(
                        text = "您好，$workerName",
                        color = Color.White,
                        fontSize = 20.sp,
                        fontWeight = FontWeight.SemiBold
                    )
                    Spacer(modifier = Modifier.height(4.dp))
                    Row(verticalAlignment = Alignment.CenterVertically) {
                        Text(
                            text = level,
                            color = Color.White.copy(alpha = 0.9f),
                            fontSize = 13.sp,
                            modifier = Modifier
                                .background(Color.White.copy(alpha = 0.2f), RoundedCornerShape(8.dp))
                                .padding(horizontal = 8.dp, vertical = 2.dp)
                        )
                        Spacer(modifier = Modifier.width(8.dp))
                        Text(
                            text = "📍 $gridName",
                            color = Color.White.copy(alpha = 0.85f),
                            fontSize = 12.sp
                        )
                    }
                }
                Text(
                    text = "⚙️",
                    fontSize = 26.sp,
                    modifier = Modifier.clickable(onClick = onProfileClick)
                )
            }
            Spacer(modifier = Modifier.height(16.dp))
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.spacedBy(12.dp)
            ) {
                Text(
                    text = "💰 积分余额: $points",
                    color = Color.White,
                    fontSize = 14.sp,
                    fontWeight = FontWeight.Medium,
                    modifier = Modifier
                        .background(Color.White.copy(alpha = 0.15f), RoundedCornerShape(12.dp))
                        .padding(horizontal = 12.dp, vertical = 6.dp)
                )
            }
        }
    }
}

@Composable
private fun TodayCheckInCard(
    hasCheckedIn: Boolean,
    checkInCount: Int,
    onCheckInClick: () -> Unit
) {
    androidx.compose.foundation.layout.Column(
        modifier = Modifier
            .fillMaxWidth()
            .padding(horizontal = 16.dp)
    ) {
        Row(
            modifier = Modifier
                .fillMaxWidth()
                .padding(top = (-40).dp)
                .background(Color.White, RoundedCornerShape(20.dp))
                .padding(vertical = 20.dp, horizontal = 16.dp),
            verticalAlignment = Alignment.CenterVertically,
            horizontalArrangement = Arrangement.SpaceBetween
        ) {
            Row(verticalAlignment = Alignment.CenterVertically) {
                Box(
                    modifier = Modifier
                        .size(56.dp)
                        .background(
                            if (hasCheckedIn) Color(0xFF22C55E).copy(alpha = 0.15f)
                            else Color(0xFF1D6CFF).copy(alpha = 0.15f),
                            RoundedCornerShape(16.dp)
                        ),
                    contentAlignment = Alignment.Center
                ) {
                    Text(
                        text = if (hasCheckedIn) "✅" else "📍",
                        fontSize = 28.sp
                    )
                }
                Spacer(modifier = Modifier.width(14.dp))
                Column {
                    Text(
                        text = "今日签到",
                        style = MaterialTheme.typography.titleMedium,
                        fontWeight = FontWeight.SemiBold
                    )
                    Spacer(modifier = Modifier.height(4.dp))
                    Text(
                        text = if (hasCheckedIn) "已签到 $checkInCount 个点位" else "今日还未签到",
                        style = MaterialTheme.typography.bodyMedium,
                        color = MaterialTheme.colorScheme.onSurfaceVariant
                    )
                }
            }
            Box(
                modifier = Modifier
                    .background(
                        if (hasCheckedIn) Color(0xFF22C55E) else Color(0xFF1D6CFF),
                        RoundedCornerShape(24.dp)
                    )
                    .clickable(onClick = onCheckInClick)
                    .padding(horizontal = 20.dp, vertical = 10.dp)
            ) {
                Text(
                    text = if (hasCheckedIn) "已完成" else "立即签到",
                    color = Color.White,
                    fontWeight = FontWeight.SemiBold
                )
            }
        }
    }
}

@Composable
private fun PointsOverviewCard(
    points: Int,
    todayEarned: Int,
    onPointsClick: () -> Unit
) {
    AppCard(
        modifier = Modifier.padding(horizontal = 16.dp),
        onClick = onPointsClick
    ) {
        Row(
            modifier = Modifier.fillMaxWidth(),
            verticalAlignment = Alignment.CenterVertically,
            horizontalArrangement = Arrangement.SpaceBetween
        ) {
            Row(verticalAlignment = Alignment.CenterVertically) {
                Box(
                    modifier = Modifier
                        .size(48.dp)
                        .background(Color(0xFFFFF3CD), RoundedCornerShape(12.dp)),
                    contentAlignment = Alignment.Center
                ) {
                    Text("💰", fontSize = 24.sp)
                }
                Spacer(modifier = Modifier.width(14.dp))
                Column {
                    Text(
                        text = "积分余额",
                        style = MaterialTheme.typography.bodyMedium,
                        color = MaterialTheme.colorScheme.onSurfaceVariant
                    )
                    Text(
                        text = points.toString(),
                        style = MaterialTheme.typography.displayMedium,
                        color = Color(0xFFFF9500),
                        fontWeight = FontWeight.Bold
                    )
                }
            }
            Column(horizontalAlignment = Alignment.End) {
                Text(
                    text = "今日获得",
                    style = MaterialTheme.typography.bodySmall,
                    color = MaterialTheme.colorScheme.onSurfaceVariant
                )
                Text(
                    text = "+$todayEarned",
                    style = MaterialTheme.typography.titleMedium,
                    color = Color(0xFF22C55E),
                    fontWeight = FontWeight.SemiBold
                )
                Spacer(modifier = Modifier.height(4.dp))
                Text(
                    text = "查看详情 ›",
                    style = MaterialTheme.typography.labelSmall,
                    color = MaterialTheme.colorScheme.primary,
                    fontWeight = FontWeight.Medium
                )
            }
        }
    }
}

@Composable
private fun QuickActionsGrid(
    onTaskList: () -> Unit,
    onVisitRecord: () -> Unit,
    onHazardReport: () -> Unit,
    onGiftMall: () -> Unit
) {
    val actions = listOf(
        Triple("📋", "任务列表", onTaskList),
        Triple("👣", "走访记录", onVisitRecord),
        Triple("⚠️", "隐患上报", onHazardReport),
        Triple("🎁", "礼品商城", onGiftMall)
    )

    AppCard(
        modifier = Modifier.padding(horizontal = 16.dp)
    ) {
        Row(
            modifier = Modifier.fillMaxWidth(),
            horizontalArrangement = Arrangement.SpaceEvenly
        ) {
            actions.forEach { (icon, label, onClick) ->
                QuickActionItem(
                    icon = icon,
                    label = label,
                    onClick = onClick
                )
            }
        }
    }
}

@Composable
private fun QuickActionItem(icon: String, label: String, onClick: () -> Unit) {
    Column(
        horizontalAlignment = Alignment.CenterHorizontally,
        modifier = Modifier
            .clickable(onClick = onClick)
            .padding(horizontal = 8.dp, vertical = 8.dp)
    ) {
        Box(
            modifier = Modifier
                .size(56.dp)
                .background(
                    MaterialTheme.colorScheme.primary.copy(alpha = 0.1f),
                    RoundedCornerShape(16.dp)
                ),
            contentAlignment = Alignment.Center
        ) {
            Text(
                text = icon,
                fontSize = 28.sp
            )
        }
        Text(
            text = label,
            modifier = Modifier.padding(top = 10.dp),
            style = MaterialTheme.typography.labelLarge,
            fontWeight = FontWeight.Medium
        )
    }
}

@Composable
private fun TaskStatsRow(
    pendingCount: Int,
    inProgressCount: Int,
    completedCount: Int,
    onPendingClick: () -> Unit,
    onInProgressClick: () -> Unit,
    onCompletedClick: () -> Unit
) {
    Row(
        modifier = Modifier
            .fillMaxWidth()
            .padding(horizontal = 16.dp),
        horizontalArrangement = Arrangement.spacedBy(12.dp)
    ) {
        TaskStatCard(
            count = pendingCount,
            label = "待执行",
            color = Color(0xFFF59E0B),
            onClick = onPendingClick,
            modifier = Modifier.weight(1f)
        )
        TaskStatCard(
            count = inProgressCount,
            label = "进行中",
            color = Color(0xFF1D6CFF),
            onClick = onInProgressClick,
            modifier = Modifier.weight(1f)
        )
        TaskStatCard(
            count = completedCount,
            label = "已完成",
            color = Color(0xFF22C55E),
            onClick = onCompletedClick,
            modifier = Modifier.weight(1f)
        )
    }
}

@Composable
private fun TaskStatCard(
    count: Int,
    label: String,
    color: Color,
    onClick: () -> Unit,
    modifier: Modifier = Modifier
) {
    Column(
        modifier = modifier
            .background(Color.White, RoundedCornerShape(16.dp))
            .clickable(onClick = onClick)
            .padding(vertical = 20.dp),
        horizontalAlignment = Alignment.CenterHorizontally
    ) {
        Text(
            text = count.toString(),
            color = color,
            fontSize = 32.sp,
            fontWeight = FontWeight.Bold
        )
        Text(
            text = label,
            modifier = Modifier.padding(top = 6.dp),
            style = MaterialTheme.typography.bodyMedium,
            color = MaterialTheme.colorScheme.onSurfaceVariant
        )
    }
}

@Composable
private fun RecentTasksHeader(onSeeAll: () -> Unit) {
    Row(
        modifier = Modifier
            .fillMaxWidth()
            .padding(horizontal = 16.dp, vertical = 4.dp),
        horizontalArrangement = Arrangement.SpaceBetween,
        verticalAlignment = Alignment.CenterVertically
    ) {
        Text(
            text = "最近任务",
            style = MaterialTheme.typography.titleMedium,
            fontWeight = FontWeight.Bold
        )
        Text(
            text = "查看全部 >",
            style = MaterialTheme.typography.labelLarge,
            color = MaterialTheme.colorScheme.primary,
            modifier = Modifier.clickable(onClick = onSeeAll)
        )
    }
}

@Composable
private fun TaskMiniCard(task: GridTask, onClick: () -> Unit) {
    val statusColor = androidx.compose.ui.graphics.Color(task.statusColor)
    val priorityColor = androidx.compose.ui.graphics.Color(task.priorityColor)

    AppCard(
        modifier = Modifier.padding(horizontal = 16.dp),
        onClick = onClick
    ) {
        Column {
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.SpaceBetween,
                verticalAlignment = Alignment.CenterVertically
            ) {
                Text(
                    text = task.taskNo,
                    style = MaterialTheme.typography.labelLarge,
                    fontWeight = FontWeight.SemiBold,
                    color = MaterialTheme.colorScheme.onSurfaceVariant
                )
                Box(
                    modifier = Modifier
                        .background(statusColor.copy(alpha = 0.15f), RoundedCornerShape(8.dp))
                        .padding(horizontal = 10.dp, vertical = 4.dp)
                ) {
                    Text(
                        text = task.status.displayName,
                        color = statusColor,
                        fontSize = 12.sp,
                        fontWeight = FontWeight.SemiBold
                    )
                }
            }

            Spacer(modifier = Modifier.height(12.dp))

            Text(
                text = task.title,
                style = MaterialTheme.typography.titleMedium,
                fontWeight = FontWeight.SemiBold,
                maxLines = 1
            )

            Spacer(modifier = Modifier.height(8.dp))

            Text(
                text = task.type.displayName,
                style = MaterialTheme.typography.bodyMedium,
                color = MaterialTheme.colorScheme.onSurfaceVariant
            )

            Spacer(modifier = Modifier.height(12.dp))

            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.SpaceBetween,
                verticalAlignment = Alignment.CenterVertically
            ) {
                Row(verticalAlignment = Alignment.CenterVertically) {
                    Box(
                        modifier = Modifier
                            .background(priorityColor.copy(alpha = 0.15f), RoundedCornerShape(6.dp))
                            .padding(horizontal = 8.dp, vertical = 3.dp)
                    ) {
                        Text(
                            text = task.priority.displayName,
                            color = priorityColor,
                            fontSize = 11.sp,
                            fontWeight = FontWeight.SemiBold
                        )
                    }
                    Spacer(modifier = Modifier.width(8.dp))
                    Text(
                        text = "📍 ${task.pointList.size} 个点位",
                        style = MaterialTheme.typography.labelMedium,
                        color = MaterialTheme.colorScheme.onSurfaceVariant
                    )
                }
                Text(
                    text = "+${task.expectedPoints}积分",
                    style = MaterialTheme.typography.labelMedium,
                    color = Color(0xFFFF9500),
                    fontWeight = FontWeight.SemiBold
                )
            }
        }
    }
}
