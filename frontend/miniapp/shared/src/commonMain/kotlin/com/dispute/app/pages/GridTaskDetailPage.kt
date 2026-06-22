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
import androidx.compose.runtime.getValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
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
import com.dispute.app.model.TaskPoint

@Composable
fun GridTaskDetailPage() {
    val appState = androidx.compose.runtime.remember { com.dispute.app.AppState() }
    val router = androidx.compose.runtime.remember { com.dispute.app.Router(appState) }
    val apiClient = androidx.compose.runtime.remember { com.dispute.app.api.ApiClient() }

    CompositionLocalProvider(
        LocalAppState provides appState,
        LocalRouter provides router,
        LocalApiClient provides apiClient
    ) {
        GridTaskDetailContent()
    }
}

@Composable
private fun GridTaskDetailContent() {
    val appState = LocalAppState.current
    val router = LocalRouter.current
    val apiClient = LocalApiClient.current
    val currentRoute = router.currentRoute.value
    val selectedTask by appState.selectedGridTask

    val taskId = (currentRoute as? Route.GridTaskDetail)?.taskId ?: selectedTask?.id ?: ""

    androidx.compose.runtime.LaunchedEffect(taskId) {
        if (selectedTask == null) {
            val task = GridWorkerMockData.mockTasks.find { it.id == taskId }
            task?.let { appState.setSelectedGridTask(it) }
        }
    }

    val task = selectedTask ?: return

    val statusColor = Color(task.statusColor)
    val priorityColor = Color(task.priorityColor)
    val checkedInCount = task.pointList.count { it.checkInStatus == TaskPoint.CheckInStatus.CHECKED_IN }

    LazyColumn(
        modifier = Modifier
            .fillMaxSize()
            .background(MaterialTheme.colorScheme.background),
        verticalArrangement = Arrangement.spacedBy(16.dp)
    ) {
        item {
            DetailTopBar(
                title = "任务详情",
                onBack = { router.back() }
            )
        }

        item {
            TaskBasicInfoCard(
                task = task,
                statusColor = statusColor,
                priorityColor = priorityColor
            )
        }

        item {
            TaskProgressCard(
                totalPoints = task.pointList.size,
                checkedInCount = checkedInCount,
                onRoutePlanClick = { router.navigate(Route.MapRoute(task.id)) }
            )
        }

        item {
            PointsInfoCard(
                expectedPoints = task.expectedPoints
            )
        }

        item {
            SectionHeader(title = "点位列表")
        }

        items(task.pointList, key = { it.id }) { point ->
            PointListItem(
                point = point,
                onCheckInClick = {
                    appState.setCheckInPoint(point)
                    router.navigate(Route.CheckIn(task.id, point.id))
                }
            )
        }

        item {
            Spacer(modifier = Modifier.height(100.dp))
        }
    }

    BottomActionButton(
        task = task,
        checkedInCount = checkedInCount,
        onStartTask = {
            appState.launchWithLoading {
                apiClient.gridWorker.startTask(task.id, "gw001")
                appState.updateGridTask(task.id) { it.copy(status = GridTask.TaskStatus.IN_PROGRESS) }
                appState.showToast("任务已开始")
            }
        },
        onCompleteTask = {
            appState.launchWithLoading {
                apiClient.gridWorker.completeTask(task.id, "gw001")
                appState.updateGridTask(task.id) { it.copy(status = GridTask.TaskStatus.COMPLETED) }
                appState.showToast("任务已完成")
            }
        },
        onRoutePlan = { router.navigate(Route.MapRoute(task.id)) }
    )
}

@Composable
private fun DetailTopBar(title: String, onBack: () -> Unit) {
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
private fun TaskBasicInfoCard(
    task: GridTask,
    statusColor: Color,
    priorityColor: Color
) {
    AppCard(
        modifier = Modifier.padding(horizontal = 16.dp)
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
                style = MaterialTheme.typography.titleLarge,
                fontWeight = FontWeight.Bold
            )

            Spacer(modifier = Modifier.height(8.dp))

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
                    text = task.type.displayName,
                    style = MaterialTheme.typography.bodyMedium,
                    color = MaterialTheme.colorScheme.onSurfaceVariant
                )
            }

            Spacer(modifier = Modifier.height(12.dp))

            Text(
                text = "任务描述",
                style = MaterialTheme.typography.titleSmall,
                fontWeight = FontWeight.SemiBold
            )

            Spacer(modifier = Modifier.height(8.dp))

            Text(
                text = task.description,
                style = MaterialTheme.typography.bodyMedium,
                color = MaterialTheme.colorScheme.onSurfaceVariant
            )

            task.deadline?.let { deadline ->
                Spacer(modifier = Modifier.height(12.dp))
                Row(verticalAlignment = Alignment.CenterVertically) {
                    Text("⏰", fontSize = 14.sp)
                    Spacer(modifier = Modifier.width(6.dp))
                    Text(
                        text = "截止时间: $deadline",
                        style = MaterialTheme.typography.bodyMedium,
                        color = Color(0xFFEF4444)
                    )
                }
            }

            task.assignedName?.let { name ->
                Spacer(modifier = Modifier.height(8.dp))
                Row(verticalAlignment = Alignment.CenterVertically) {
                    Text("👤", fontSize = 14.sp)
                    Spacer(modifier = Modifier.width(6.dp))
                    Text(
                        text = "负责人: $name",
                        style = MaterialTheme.typography.bodyMedium,
                        color = MaterialTheme.colorScheme.onSurfaceVariant
                    )
                }
            }

            Spacer(modifier = Modifier.height(8.dp))
            Row(verticalAlignment = Alignment.CenterVertically) {
                Text("📍", fontSize = 14.sp)
                Spacer(modifier = Modifier.width(6.dp))
                Text(
                    text = "所属网格: ${task.gridName}",
                    style = MaterialTheme.typography.bodyMedium,
                    color = MaterialTheme.colorScheme.onSurfaceVariant
                )
            }

            Spacer(modifier = Modifier.height(8.dp))
            Row(verticalAlignment = Alignment.CenterVertically) {
                Text("📅", fontSize = 14.sp)
                Spacer(modifier = Modifier.width(6.dp))
                Text(
                    text = "创建时间: ${task.createTime}",
                    style = MaterialTheme.typography.bodyMedium,
                    color = MaterialTheme.colorScheme.onSurfaceVariant
                )
            }
        }
    }
}

@Composable
private fun TaskProgressCard(
    totalPoints: Int,
    checkedInCount: Int,
    onRoutePlanClick: () -> Unit
) {
    val progress = if (totalPoints > 0) (checkedInCount * 100) / totalPoints else 0

    AppCard(
        modifier = Modifier.padding(horizontal = 16.dp)
    ) {
        Column {
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.SpaceBetween,
                verticalAlignment = Alignment.CenterVertically
            ) {
                Text(
                    text = "执行进度",
                    style = MaterialTheme.typography.titleSmall,
                    fontWeight = FontWeight.SemiBold
                )
                Text(
                    text = "$progress%",
                    style = MaterialTheme.typography.titleMedium,
                    color = MaterialTheme.colorScheme.primary,
                    fontWeight = FontWeight.Bold
                )
            }

            Spacer(modifier = Modifier.height(12.dp))

            Box(
                modifier = Modifier
                    .fillMaxWidth()
                    .height(8.dp)
                    .background(
                        MaterialTheme.colorScheme.surfaceVariant,
                        RoundedCornerShape(4.dp)
                    )
            ) {
                Box(
                    modifier = Modifier
                        .fillMaxWidth(progress.toFloat() / 100f)
                        .height(8.dp)
                        .background(
                            MaterialTheme.colorScheme.primary,
                            RoundedCornerShape(4.dp)
                        )
                )
            }

            Spacer(modifier = Modifier.height(12.dp))

            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.SpaceBetween,
                verticalAlignment = Alignment.CenterVertically
            ) {
                Text(
                    text = "已完成 $checkedInCount / $totalPoints 个点位",
                    style = MaterialTheme.typography.bodyMedium,
                    color = MaterialTheme.colorScheme.onSurfaceVariant
                )
                Text(
                    text = "🗺️ 路线规划",
                    style = MaterialTheme.typography.labelLarge,
                    color = MaterialTheme.colorScheme.primary,
                    fontWeight = FontWeight.Medium,
                    modifier = Modifier.clickable(onClick = onRoutePlanClick)
                )
            }
        }
    }
}

@Composable
private fun PointsInfoCard(
    expectedPoints: Int
) {
    AppCard(
        modifier = Modifier.padding(horizontal = 16.dp)
    ) {
        Row(
            modifier = Modifier.fillMaxWidth(),
            verticalAlignment = Alignment.CenterVertically,
            horizontalArrangement = Arrangement.SpaceBetween
        ) {
            Row(verticalAlignment = Alignment.CenterVertically) {
                Text("💰", fontSize = 24.sp)
                Spacer(modifier = Modifier.width(12.dp))
                Column {
                    Text(
                        text = "预计获得积分",
                        style = MaterialTheme.typography.bodyMedium,
                        color = MaterialTheme.colorScheme.onSurfaceVariant
                    )
                    Text(
                        text = "+$expectedPoints 积分",
                        style = MaterialTheme.typography.titleLarge,
                        color = Color(0xFFFF9500),
                        fontWeight = FontWeight.Bold
                    )
                }
            }
        }
    }
}

@Composable
private fun SectionHeader(title: String) {
    Text(
        text = title,
        style = MaterialTheme.typography.titleMedium,
        fontWeight = FontWeight.Bold,
        modifier = Modifier.padding(horizontal = 16.dp)
    )
}

@Composable
private fun PointListItem(
    point: TaskPoint,
    onCheckInClick: () -> Unit
) {
    val statusColor = when (point.checkInStatus) {
        TaskPoint.CheckInStatus.CHECKED_IN -> Color(0xFF22C55E)
        TaskPoint.CheckInStatus.SKIPPED -> Color(0xFF9CA3AF)
        else -> Color(0xFFF59E0B)
    }

    AppCard(
        modifier = Modifier.padding(horizontal = 16.dp),
        onClick = if (point.checkInStatus == TaskPoint.CheckInStatus.PENDING) onCheckInClick else null
    ) {
        Row(
            modifier = Modifier.fillMaxWidth(),
            verticalAlignment = Alignment.CenterVertically
        ) {
            Box(
                modifier = Modifier
                    .size(40.dp)
                    .background(
                        statusColor.copy(alpha = 0.15f),
                        RoundedCornerShape(12.dp)
                    ),
                contentAlignment = Alignment.Center
            ) {
                Text(
                    text = point.sortOrder.toString(),
                    style = MaterialTheme.typography.titleMedium,
                    color = statusColor,
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
                Spacer(modifier = Modifier.height(4.dp))
                Text(
                    text = point.address,
                    style = MaterialTheme.typography.bodySmall,
                    color = MaterialTheme.colorScheme.onSurfaceVariant,
                    maxLines = 1
                )
                point.checkInTime?.let { time ->
                    Spacer(modifier = Modifier.height(4.dp))
                    Text(
                        text = "签到时间: ${time.substringBefore(" ")} ${time.substringAfter(" ").substringBefore(".")}",
                        style = MaterialTheme.typography.labelSmall,
                        color = Color(0xFF22C55E)
                    )
                }
            }

            Spacer(modifier = Modifier.width(8.dp))

            if (point.checkInStatus == TaskPoint.CheckInStatus.CHECKED_IN) {
                Text("✅", fontSize = 24.sp)
            } else if (point.checkInStatus == TaskPoint.CheckInStatus.SKIPPED) {
                Text("⏭️", fontSize = 24.sp)
            } else {
                Box(
                    modifier = Modifier
                        .background(MaterialTheme.colorScheme.primary, RoundedCornerShape(16.dp))
                        .padding(horizontal = 12.dp, vertical = 6.dp)
                ) {
                    Text(
                        text = "签到",
                        color = Color.White,
                        style = MaterialTheme.typography.labelMedium,
                        fontWeight = FontWeight.SemiBold
                    )
                }
            }
        }
    }
}

@Composable
private fun BottomActionButton(
    task: GridTask,
    checkedInCount: Int,
    onStartTask: () -> Unit,
    onCompleteTask: () -> Unit,
    onRoutePlan: () -> Unit
) {
    Box(
        modifier = Modifier
            .fillMaxSize()
            .padding(bottom = 16.dp),
        contentAlignment = Alignment.BottomCenter
    ) {
        Row(
            modifier = Modifier
                .fillMaxWidth()
                .padding(horizontal = 16.dp),
            horizontalArrangement = Arrangement.spacedBy(12.dp)
        ) {
            Box(
                modifier = Modifier
                    .weight(1f)
                    .background(
                        MaterialTheme.colorScheme.primary.copy(alpha = 0.1f),
                        RoundedCornerShape(24.dp)
                    )
                    .clickable(onClick = onRoutePlan)
                    .padding(vertical = 14.dp),
                contentAlignment = Alignment.Center
            ) {
                Text(
                    text = "🗺️ 路线规划",
                    style = MaterialTheme.typography.titleSmall,
                    color = MaterialTheme.colorScheme.primary,
                    fontWeight = FontWeight.SemiBold
                )
            }

            when (task.status) {
                GridTask.TaskStatus.PENDING -> {
                    Box(
                        modifier = Modifier
                            .weight(2f)
                            .background(Color(0xFF1D6CFF), RoundedCornerShape(24.dp))
                            .clickable(onClick = onStartTask)
                            .padding(vertical = 14.dp),
                        contentAlignment = Alignment.Center
                    ) {
                        Text(
                            text = "开始任务",
                            style = MaterialTheme.typography.titleSmall,
                            color = Color.White,
                            fontWeight = FontWeight.SemiBold
                        )
                    }
                }
                GridTask.TaskStatus.IN_PROGRESS -> {
                    val canComplete = checkedInCount == task.pointList.size
                    Box(
                        modifier = Modifier
                            .weight(2f)
                            .background(
                                if (canComplete) Color(0xFF22C55E) else Color(0xFF9CA3AF),
                                RoundedCornerShape(24.dp)
                            )
                            .clickable(enabled = canComplete, onClick = onCompleteTask)
                            .padding(vertical = 14.dp),
                        contentAlignment = Alignment.Center
                    ) {
                        Text(
                            text = if (canComplete) "完成任务" else "请先完成所有点位签到",
                            style = MaterialTheme.typography.titleSmall,
                            color = Color.White,
                            fontWeight = FontWeight.SemiBold
                        )
                    }
                }
                else -> {
                    Box(
                        modifier = Modifier
                            .weight(2f)
                            .background(Color(0xFF9CA3AF), RoundedCornerShape(24.dp))
                            .padding(vertical = 14.dp),
                        contentAlignment = Alignment.Center
                    ) {
                        Text(
                            text = task.status.displayName,
                            style = MaterialTheme.typography.titleSmall,
                            color = Color.White,
                            fontWeight = FontWeight.SemiBold
                        )
                    }
                }
            }
        }
    }
}
