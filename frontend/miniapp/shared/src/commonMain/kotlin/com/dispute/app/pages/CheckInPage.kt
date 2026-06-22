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
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
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
import com.dispute.app.model.GridWorkerMockData
import com.dispute.app.model.TaskPoint
import com.dispute.app.platform.CameraService
import com.dispute.app.platform.LivenessAction
import com.dispute.app.platform.LivenessDetector
import com.dispute.app.platform.LocationResult
import com.dispute.app.platform.LocationService
import com.dispute.app.platform.PhotoOptions
import com.dispute.app.platform.PhotoResult
import kotlinx.coroutines.launch
import kotlinx.datetime.Clock
import kotlinx.datetime.TimeZone
import kotlinx.datetime.toLocalDateTime

@Composable
fun CheckInPage() = CheckInContent()

@Composable
private fun CheckInContent() {
    val appState = LocalAppState.current
    val router = LocalRouter.current
    val apiClient = LocalApiClient.current
    val currentRoute = router.currentRoute.value
    val checkInPoint by appState.checkInPoint
    val selectedTask by appState.selectedGridTask

    val locationService = remember { LocationService() }
    val cameraService = remember { CameraService() }
    val livenessDetector = remember { LivenessDetector() }

    val routeParams = currentRoute as? Route.CheckIn
    val taskId = routeParams?.taskId ?: selectedTask?.id ?: ""
    val pointId = routeParams?.pointId ?: checkInPoint?.id ?: ""

    var location by remember { mutableStateOf<Pair<Double, Double>?>(null) }
    var locationAccuracy by remember { mutableStateOf<Float?>(null) }
    var locationResult by remember { mutableStateOf<LocationResult?>(null) }
    var address by remember { mutableStateOf("定位中...") }
    var hasPhoto by remember { mutableStateOf(false) }
    var photoDataUrl by remember { mutableStateOf<String?>(null) }
    var livenessVerified by remember { mutableStateOf(false) }
    var livenessScore by remember { mutableStateOf<Float?>(null) }
    var livenessPhotoDataUrl by remember { mutableStateOf<String?>(null) }
    var isCheckingIn by remember { mutableStateOf(false) }
    var checkInSuccess by remember { mutableStateOf(false) }

    val point = checkInPoint ?: selectedTask?.pointList?.find { it.id == pointId }
        ?: GridWorkerMockData.mockTasks.flatMap { it.pointList }.find { it.id == pointId }

    androidx.compose.runtime.LaunchedEffect(Unit) {
        point?.let {
            appState.setCheckInPoint(it)
            location = Pair(it.longitude, it.latitude)
            address = it.address
        }
    }

    if (point == null) {
        Box(modifier = Modifier.fillMaxSize(), contentAlignment = Alignment.Center) {
            Text("点位信息不存在", style = MaterialTheme.typography.bodyLarge)
        }
        return
    }

    if (checkInSuccess) {
        CheckInSuccessScreen(
            point = point,
            onBack = { router.back() }
        )
        return
    }

    Column(
        modifier = Modifier
            .fillMaxSize()
            .background(MaterialTheme.colorScheme.background)
    ) {
        TopBar(
            title = "签到打卡",
            onBack = { router.back() }
        )

        Column(
            modifier = Modifier
                .fillMaxWidth()
                .weight(1f)
                .padding(horizontal = 16.dp),
            verticalArrangement = Arrangement.spacedBy(16.dp)
        ) {
            Spacer(modifier = Modifier.height(8.dp))

            PointInfoCard(point = point)

            LocationCard(
                address = address,
                longitude = location?.first,
                latitude = location?.second,
                accuracy = locationAccuracy,
                onRefreshLocation = {
                    address = "正在获取定位..."
                    appState.appScope.launch {
                        try {
                            val result = locationService.getCurrentLocation()
                            locationResult = result
                            location = Pair(result.longitude, result.latitude)
                            locationAccuracy = result.accuracy
                            address = "经度: ${"%.6f".format(result.longitude)}, 纬度: ${"%.6f".format(result.latitude)}"
                            if (result.isMock) {
                                appState.showToast("使用模拟定位，请检查定位权限")
                            } else {
                                appState.showToast("定位成功")
                            }
                        } catch (e: Exception) {
                            location = Pair(point.longitude, point.latitude)
                            address = point.address
                            appState.showToast("定位失败: ${e.message}")
                        }
                    }
                }
            )

            PhotoCard(
                hasPhoto = hasPhoto,
                photoDataUrl = photoDataUrl,
                onTakePhoto = {
                    appState.appScope.launch {
                        try {
                            val result = cameraService.takePhoto(
                                PhotoOptions(
                                    enableBase64 = true,
                                    requireLiveDetection = false
                                )
                            )
                            hasPhoto = true
                            photoDataUrl = result.base64Data ?: result.filePath
                            appState.showToast("拍照成功")
                        } catch (e: Exception) {
                            appState.showToast("拍照失败: ${e.message}")
                        }
                    }
                }
            )

            LivenessCard(
                verified = livenessVerified,
                score = livenessScore,
                onVerify = {
                    appState.appScope.launch {
                        try {
                            val actions = listOf(LivenessAction.BLINK, LivenessAction.SMILE)
                            val livenessResult = livenessDetector.detectLiveness(
                                requiredActions = actions,
                                timeoutMs = 15000L
                            )
                            livenessVerified = livenessResult.passed
                            livenessScore = livenessResult.score
                            if (livenessResult.passed) {
                                appState.showToast("活体验证通过，分数: ${"%.2f".format(livenessResult.score)}")
                            } else {
                                appState.showToast(livenessResult.errorMessage ?: "活体验证失败")
                            }
                        } catch (e: Exception) {
                            appState.showToast("活体验证失败: ${e.message}")
                        }
                    }
                }
            )

            CheckInTips()
        }

        BottomCheckInButton(
            enabled = location != null && hasPhoto && livenessVerified && !isCheckingIn,
            isCheckingIn = isCheckingIn,
            onClick = {
                isCheckingIn = true
                appState.appScope.launch {
                    try {
                        val now = Clock.System.now().toLocalDateTime(TimeZone.currentSystemDefault())
                        val checkInRequest = com.dispute.app.api.CheckInRequest(
                            taskId = taskId,
                            pointId = point.id,
                            workerId = "gw001",
                            longitude = location!!.first,
                            latitude = location!!.second,
                            locationAccuracy = locationAccuracy?.toDouble(),
                            address = address,
                            photoUrl = photoDataUrl,
                            livePhotoUrl = livenessPhotoDataUrl,
                            livenessVerified = livenessVerified
                        )
                        apiClient.gridWorker.checkIn(checkInRequest)

                        appState.updateCheckInPoint(point.id) {
                            it.copy(
                                checkInStatus = TaskPoint.CheckInStatus.CHECKED_IN,
                                checkInTime = now.toString()
                            )
                        }

                        appState.updateGridTask(taskId) { task ->
                            val updatedPoints = task.pointList.map { p ->
                                if (p.id == point.id) {
                                    p.copy(
                                        checkInStatus = TaskPoint.CheckInStatus.CHECKED_IN,
                                        checkInTime = now.toString()
                                    )
                                } else p
                            }
                            task.copy(pointList = updatedPoints)
                        }

                        checkInSuccess = true
                    } catch (e: Exception) {
                        appState.showToast("签到失败: ${e.message}")
                    } finally {
                        isCheckingIn = false
                    }
                }
            }
        )
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
private fun PointInfoCard(point: TaskPoint) {
    AppCard {
        Column {
            Row(
                modifier = Modifier.fillMaxWidth(),
                verticalAlignment = Alignment.CenterVertically
            ) {
                Box(
                    modifier = Modifier
                        .size(48.dp)
                        .background(MaterialTheme.colorScheme.primary.copy(alpha = 0.1f), RoundedCornerShape(12.dp)),
                    contentAlignment = Alignment.Center
                ) {
                    Text("📍", fontSize = 24.sp)
                }
                Spacer(modifier = Modifier.width(12.dp))
                Column(modifier = Modifier.weight(1f)) {
                    Text(
                        text = point.name,
                        style = MaterialTheme.typography.titleMedium,
                        fontWeight = FontWeight.SemiBold
                    )
                    Spacer(modifier = Modifier.height(4.dp))
                    Text(
                        text = point.address,
                        style = MaterialTheme.typography.bodyMedium,
                        color = MaterialTheme.colorScheme.onSurfaceVariant
                    )
                }
            }
        }
    }
}

@Composable
private fun LocationCard(
    address: String,
    longitude: Double?,
    latitude: Double?,
    accuracy: Float?,
    onRefreshLocation: () -> Unit
) {
    AppCard(
        title = "GPS定位",
        subtitle = "请确保您已到达指定位置"
    ) {
        Column {
            Row(
                modifier = Modifier.fillMaxWidth(),
                verticalAlignment = Alignment.CenterVertically,
                horizontalArrangement = Arrangement.SpaceBetween
            ) {
                Column(modifier = Modifier.weight(1f)) {
                    Text(
                        text = address,
                        style = MaterialTheme.typography.bodyMedium,
                        fontWeight = FontWeight.Medium
                    )
                    if (longitude != null && latitude != null) {
                        Spacer(modifier = Modifier.height(4.dp))
                        val accuracyText = accuracy?.let { " | 精度: ${"%.0f".format(it)}m" } ?: ""
                        Text(
                            text = "经度: ${"%.6f".format(longitude)} | 纬度: ${"%.6f".format(latitude)}$accuracyText",
                            style = MaterialTheme.typography.labelSmall,
                            color = MaterialTheme.colorScheme.onSurfaceVariant
                        )
                    }
                }
                Box(
                    modifier = Modifier
                        .background(
                            MaterialTheme.colorScheme.primary.copy(alpha = 0.1f),
                            RoundedCornerShape(12.dp)
                        )
                        .clickable(onClick = onRefreshLocation)
                        .padding(horizontal = 12.dp, vertical = 8.dp)
                ) {
                    Text(
                        text = "刷新定位",
                        style = MaterialTheme.typography.labelMedium,
                        color = MaterialTheme.colorScheme.primary,
                        fontWeight = FontWeight.Medium
                    )
                }
            }

            Spacer(modifier = Modifier.height(12.dp))

            Box(
                modifier = Modifier
                    .fillMaxWidth()
                    .height(120.dp)
                    .background(
                        Brush.linearGradient(
                            colors = listOf(
                                Color(0xFFE3F2FD),
                                Color(0xFFBBDEFB)
                            )
                        ),
                        RoundedCornerShape(12.dp)
                    ),
                contentAlignment = Alignment.Center
            ) {
                Column(horizontalAlignment = Alignment.CenterHorizontally) {
                    Text("🗺️", fontSize = 48.sp)
                    Spacer(modifier = Modifier.height(8.dp))
                    Text(
                        text = "高德地图",
                        style = MaterialTheme.typography.labelMedium,
                        color = MaterialTheme.colorScheme.onSurfaceVariant
                    )
                }
            }
        }
    }
}

@Composable
private fun PhotoCard(
    hasPhoto: Boolean,
    photoDataUrl: String?,
    onTakePhoto: () -> Unit
) {
    AppCard(
        title = "现场拍照",
        subtitle = "请拍摄现场照片作为签到凭证"
    ) {
        Row(
            modifier = Modifier.fillMaxWidth(),
            horizontalArrangement = Arrangement.spacedBy(12.dp)
        ) {
            if (hasPhoto) {
                Box(
                    modifier = Modifier
                        .size(100.dp)
                        .background(
                            Brush.linearGradient(
                                colors = listOf(
                                    Color(0xFFC8E6C9),
                                    Color(0xFFA5D6A7)
                                )
                            ),
                            RoundedCornerShape(12.dp)
                        ),
                    contentAlignment = Alignment.Center
                ) {
                    Column(horizontalAlignment = Alignment.CenterHorizontally) {
                        Text("✅", fontSize = 32.sp)
                        Spacer(modifier = Modifier.height(4.dp))
                        Text(
                            text = "已拍摄",
                            style = MaterialTheme.typography.labelSmall,
                            color = Color(0xFF2E7D32)
                        )
                    }
                }
            }
            Box(
                modifier = Modifier
                    .size(100.dp)
                    .background(
                        MaterialTheme.colorScheme.surfaceVariant,
                        RoundedCornerShape(12.dp)
                    )
                    .clickable(onClick = onTakePhoto),
                contentAlignment = Alignment.Center
            ) {
                Column(horizontalAlignment = Alignment.CenterHorizontally) {
                    Text("📷", fontSize = 32.sp)
                    Spacer(modifier = Modifier.height(4.dp))
                    Text(
                        text = if (hasPhoto) "重新拍摄" else "拍摄照片",
                        style = MaterialTheme.typography.labelSmall,
                        color = MaterialTheme.colorScheme.onSurfaceVariant
                    )
                }
            }
        }
    }
}

@Composable
private fun LivenessCard(
    verified: Boolean,
    score: Float?,
    onVerify: () -> Unit
) {
    AppCard(
        title = "活体验证",
        subtitle = "请进行人脸识别验证本人身份"
    ) {
        Row(
            modifier = Modifier.fillMaxWidth(),
            verticalAlignment = Alignment.CenterVertically,
            horizontalArrangement = Arrangement.SpaceBetween
        ) {
            Row(verticalAlignment = Alignment.CenterVertically) {
                Box(
                    modifier = Modifier
                        .size(56.dp)
                        .background(
                            if (verified) Color(0xFF22C55E).copy(alpha = 0.15f)
                            else MaterialTheme.colorScheme.surfaceVariant,
                            CircleShape
                        ),
                    contentAlignment = Alignment.Center
                ) {
                    Text(
                        text = if (verified) "✅" else "👤",
                        fontSize = 28.sp
                    )
                }
                Spacer(modifier = Modifier.width(12.dp))
                Column {
                    Text(
                        text = if (verified) "验证通过" else "未验证",
                        style = MaterialTheme.typography.titleSmall,
                        fontWeight = FontWeight.SemiBold,
                        color = if (verified) Color(0xFF22C55E) else MaterialTheme.colorScheme.onSurface
                    )
                    Spacer(modifier = Modifier.height(4.dp))
                    val scoreText = score?.let { " | 分数: ${"%.2f".format(it)}" } ?: ""
                    Text(
                        text = if (verified) "本人身份已确认$scoreText" else "点击开始人脸识别",
                        style = MaterialTheme.typography.bodySmall,
                        color = MaterialTheme.colorScheme.onSurfaceVariant
                    )
                }
            }
            if (!verified) {
                Box(
                    modifier = Modifier
                        .background(MaterialTheme.colorScheme.primary, RoundedCornerShape(20.dp))
                        .clickable(onClick = onVerify)
                        .padding(horizontal = 16.dp, vertical = 10.dp)
                ) {
                    Text(
                        text = "开始验证",
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
private fun CheckInTips() {
    AppCard(
        backgroundColor = Color(0xFFFFF8E7)
    ) {
        Column {
            Row(verticalAlignment = Alignment.CenterVertically) {
                Text("💡", fontSize = 20.sp)
                Spacer(modifier = Modifier.width(8.dp))
                Text(
                    text = "签到须知",
                    style = MaterialTheme.typography.titleSmall,
                    fontWeight = FontWeight.SemiBold,
                    color = Color(0xFF92400E)
                )
            }
            Spacer(modifier = Modifier.height(8.dp))
            Text(
                text = "1. 请确保您已到达指定位置后再进行签到\n" +
                        "2. 现场照片需清晰反映实际环境\n" +
                        "3. 活体验证需本人完成，不可替代\n" +
                        "4. 虚假签到将导致积分扣除并影响考核",
                style = MaterialTheme.typography.bodySmall,
                color = Color(0xFF92400E),
                lineHeight = 20.sp
            )
        }
    }
}

@Composable
private fun BottomCheckInButton(
    enabled: Boolean,
    isCheckingIn: Boolean,
    onClick: () -> Unit
) {
    Box(
        modifier = Modifier
            .fillMaxWidth()
            .padding(horizontal = 16.dp, vertical = 24.dp)
    ) {
        Box(
            modifier = Modifier
                .fillMaxWidth()
                .background(
                    if (enabled) {
                        Brush.linearGradient(
                            colors = listOf(
                                Color(0xFF1D6CFF),
                                Color(0xFF4D8CFF)
                            )
                        )
                    } else {
                        Brush.linearGradient(
                            colors = listOf(
                                Color(0xFF9CA3AF),
                                Color(0xFFD1D5DB)
                            )
                        )
                    },
                    RoundedCornerShape(28.dp)
                )
                .clickable(enabled = enabled, onClick = onClick)
                .padding(vertical = 16.dp),
            contentAlignment = Alignment.Center
        ) {
            if (isCheckingIn) {
                Row(verticalAlignment = Alignment.CenterVertically) {
                    androidx.compose.material3.CircularProgressIndicator(
                        color = Color.White,
                        strokeWidth = 2.dp,
                        modifier = Modifier.size(20.dp)
                    )
                    Spacer(modifier = Modifier.width(8.dp))
                    Text(
                        text = "签到中...",
                        color = Color.White,
                        style = MaterialTheme.typography.titleMedium,
                        fontWeight = FontWeight.SemiBold
                    )
                }
            } else {
                Text(
                    text = if (enabled) "立即签到" else "请完成所有验证步骤",
                    color = Color.White,
                    style = MaterialTheme.typography.titleMedium,
                    fontWeight = FontWeight.SemiBold
                )
            }
        }
    }
}

@Composable
private fun CheckInSuccessScreen(
    point: TaskPoint,
    onBack: () -> Unit
) {
    Box(
        modifier = Modifier
            .fillMaxSize()
            .background(
                Brush.linearGradient(
                    colors = listOf(
                        Color(0xFF22C55E),
                        Color(0xFF4ADE80)
                    )
                )
            )
    ) {
        Column(
            modifier = Modifier
                .fillMaxSize()
                .padding(horizontal = 24.dp),
            horizontalAlignment = Alignment.CenterHorizontally,
            verticalArrangement = Arrangement.Center
        ) {
            Box(
                modifier = Modifier
                    .size(100.dp)
                    .background(Color.White.copy(alpha = 0.2f), CircleShape),
                contentAlignment = Alignment.Center
            ) {
                Text("✅", fontSize = 56.sp)
            }

            Spacer(modifier = Modifier.height(24.dp))

            Text(
                text = "签到成功！",
                color = Color.White,
                style = MaterialTheme.typography.displayMedium,
                fontWeight = FontWeight.Bold
            )

            Spacer(modifier = Modifier.height(12.dp))

            Text(
                text = "+5 积分",
                color = Color.White.copy(alpha = 0.9f),
                style = MaterialTheme.typography.titleLarge,
                fontWeight = FontWeight.SemiBold
            )

            Spacer(modifier = Modifier.height(32.dp))

            AppCard(
                modifier = Modifier.fillMaxWidth()
            ) {
                Column(verticalArrangement = Arrangement.spacedBy(12.dp)) {
                    InfoRow("点位名称", point.name)
                    InfoRow("位置", point.address)
                    InfoRow(
                        "签到时间",
                        Clock.System.now().toLocalDateTime(TimeZone.currentSystemDefault()).toString()
                            .replace("T", " ")
                    )
                    InfoRow("获得积分", "+5 积分", valueColor = Color(0xFFFF9500))
                }
            }

            Spacer(modifier = Modifier.height(32.dp))

            Box(
                modifier = Modifier
                    .fillMaxWidth()
                    .background(Color.White, RoundedCornerShape(24.dp))
                    .clickable(onClick = onBack)
                    .padding(vertical = 16.dp),
                contentAlignment = Alignment.Center
            ) {
                Text(
                    text = "返回",
                    color = Color(0xFF22C55E),
                    style = MaterialTheme.typography.titleMedium,
                    fontWeight = FontWeight.SemiBold
                )
            }
        }
    }
}

@Composable
private fun InfoRow(label: String, value: String, valueColor: Color = MaterialTheme.colorScheme.onSurface) {
    Row(
        modifier = Modifier.fillMaxWidth(),
        horizontalArrangement = Arrangement.SpaceBetween
    ) {
        Text(
            text = label,
            style = MaterialTheme.typography.bodyMedium,
            color = MaterialTheme.colorScheme.onSurfaceVariant
        )
        Text(
            text = value,
            style = MaterialTheme.typography.bodyMedium,
            fontWeight = FontWeight.Medium,
            color = valueColor
        )
    }
}
