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
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.text.KeyboardOptions
import androidx.compose.foundation.verticalScroll
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.OutlinedTextField
import androidx.compose.material3.Text
import androidx.compose.material3.TextFieldDefaults
import androidx.compose.runtime.Composable
import androidx.compose.runtime.CompositionLocalProvider
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.input.KeyboardType
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.dispute.app.LocalAppState
import com.dispute.app.LocalApiClient
import com.dispute.app.LocalRouter
import com.dispute.app.Route
import com.dispute.app.components.AppCard
import com.dispute.app.model.HazardReport
import kotlinx.coroutines.launch
import kotlinx.datetime.Clock
import kotlinx.datetime.TimeZone
import kotlinx.datetime.toLocalDateTime

@Composable
fun HazardReportPage() = HazardReportContent()

@Composable
private fun HazardReportContent() {
    val appState = LocalAppState.current
    val router = LocalRouter.current
    val apiClient = LocalApiClient.current

    var selectedHazardType by remember { mutableStateOf<HazardReport.HazardType?>(null) }
    var selectedUrgency by remember { mutableStateOf<HazardReport.UrgencyLevel?>(HazardReport.UrgencyLevel.NORMAL) }
    var hazardTitle by remember { mutableStateOf("") }
    var hazardDescription by remember { mutableStateOf("") }
    var hazardLocation by remember { mutableStateOf("") }
    var reporterName by remember { mutableStateOf("") }
    var reporterPhone by remember { mutableStateOf("") }
    var hasLocation by remember { mutableStateOf(false) }
    var photoCount by remember { mutableStateOf(0) }
    var isSubmitting by remember { mutableStateOf(false) }

    val hazardTypes = HazardReport.HazardType.values().toList()
    val urgencyLevels = HazardReport.UrgencyLevel.values().toList()

    val canSubmit = selectedHazardType != null && hazardTitle.isNotBlank() &&
            hazardDescription.isNotBlank() && hazardLocation.isNotBlank() && !isSubmitting

    Column(
        modifier = Modifier
            .fillMaxSize()
            .background(MaterialTheme.colorScheme.background)
    ) {
        TopBar(
            title = "隐患上报",
            onBack = { router.back() }
        )

        Column(
            modifier = Modifier
                .fillMaxWidth()
                .weight(1f)
                .verticalScroll(rememberScrollState())
                .padding(horizontal = 16.dp, vertical = 12.dp),
            verticalArrangement = Arrangement.spacedBy(16.dp)
        ) {
            AppCard(title = "隐患类型", subtitle = "请选择隐患类型") {
                Column(verticalArrangement = Arrangement.spacedBy(12.dp)) {
                    val chunks = hazardTypes.chunked(2)
                    chunks.forEach { chunk ->
                        Row(
                            modifier = Modifier.fillMaxWidth(),
                            horizontalArrangement = Arrangement.spacedBy(8.dp)
                        ) {
                            chunk.forEach { type ->
                                val isSelected = selectedHazardType == type
                                Box(
                                    modifier = Modifier
                                        .weight(1f)
                                        .background(
                                            if (isSelected)
                                                MaterialTheme.colorScheme.primary.copy(alpha = 0.15f)
                                            else
                                                MaterialTheme.colorScheme.surfaceVariant.copy(alpha = 0.5f),
                                            RoundedCornerShape(12.dp)
                                        )
                                        .clickable { selectedHazardType = type }
                                        .padding(vertical = 14.dp, horizontal = 8.dp),
                                    contentAlignment = Alignment.Center
                                ) {
                                    Column(horizontalAlignment = Alignment.CenterHorizontally) {
                                        Text(
                                            text = type.icon,
                                            fontSize = 24.sp
                                        )
                                        Spacer(modifier = Modifier.height(4.dp))
                                        Text(
                                            text = type.displayName,
                                            style = MaterialTheme.typography.labelMedium,
                                            fontWeight = if (isSelected) FontWeight.SemiBold else FontWeight.Normal,
                                            color = if (isSelected)
                                                MaterialTheme.colorScheme.primary
                                            else
                                                MaterialTheme.colorScheme.onSurface
                                        )
                                    }
                                }
                            }
                        }
                    }
                }
            }

            AppCard(title = "紧急程度") {
                Row(
                    modifier = Modifier.fillMaxWidth(),
                    horizontalArrangement = Arrangement.spacedBy(8.dp)
                ) {
                    urgencyLevels.forEach { urgency ->
                        val isSelected = selectedUrgency == urgency
                        Box(
                            modifier = Modifier
                                .weight(1f)
                                .background(
                                    if (isSelected)
                                        urgency.color.copy(alpha = 0.15f)
                                    else
                                        MaterialTheme.colorScheme.surfaceVariant.copy(alpha = 0.5f),
                                    RoundedCornerShape(12.dp)
                                )
                                .clickable { selectedUrgency = urgency }
                                .padding(vertical = 12.dp),
                            contentAlignment = Alignment.Center
                        ) {
                            Text(
                                text = urgency.displayName,
                                style = MaterialTheme.typography.labelMedium,
                                fontWeight = if (isSelected) FontWeight.SemiBold else FontWeight.Normal,
                                color = if (isSelected) urgency.color else MaterialTheme.colorScheme.onSurface
                            )
                        }
                    }
                }
            }

            AppCard(title = "隐患信息") {
                Column(verticalArrangement = Arrangement.spacedBy(12.dp)) {
                    FormInputField(
                        label = "隐患标题",
                        value = hazardTitle,
                        onValueChange = { hazardTitle = it },
                        placeholder = "请简要描述隐患标题",
                        required = true
                    )

                    Column {
                        Row {
                            Text(
                                text = "隐患描述",
                                style = MaterialTheme.typography.titleSmall,
                                fontWeight = FontWeight.Medium
                            )
                            Text(
                                text = " *",
                                color = Color(0xFFEF4444),
                                style = MaterialTheme.typography.titleSmall
                            )
                        }
                        Spacer(modifier = Modifier.height(8.dp))
                        OutlinedTextField(
                            value = hazardDescription,
                            onValueChange = { hazardDescription = it },
                            modifier = Modifier
                                .fillMaxWidth()
                                .height(120.dp),
                            placeholder = {
                                Text(
                                    text = "请详细描述隐患情况，包括位置、状态、可能造成的影响等...",
                                    style = MaterialTheme.typography.bodyMedium,
                                    color = MaterialTheme.colorScheme.onSurfaceVariant
                                )
                            },
                            colors = TextFieldDefaults.colors(
                                focusedContainerColor = Color.Transparent,
                                unfocusedContainerColor = Color.Transparent
                            ),
                            maxLines = 5
                        )
                    }

                    FormInputField(
                        label = "隐患位置",
                        value = hazardLocation,
                        onValueChange = { hazardLocation = it },
                        placeholder = "请输入隐患具体位置",
                        required = true
                    )
                }
            }

            AppCard(title = "上报人信息") {
                Column(verticalArrangement = Arrangement.spacedBy(12.dp)) {
                    FormInputField(
                        label = "上报人姓名",
                        value = reporterName,
                        onValueChange = { reporterName = it },
                        placeholder = "请输入上报人姓名（选填）"
                    )

                    FormInputField(
                        label = "联系电话",
                        value = reporterPhone,
                        onValueChange = { reporterPhone = it },
                        placeholder = "请输入联系电话（选填）",
                        keyboardType = KeyboardType.Phone
                    )
                }
            }

            AppCard(title = "定位信息") {
                Row(
                    modifier = Modifier.fillMaxWidth(),
                    verticalAlignment = Alignment.CenterVertically,
                    horizontalArrangement = Arrangement.SpaceBetween
                ) {
                    Row(verticalAlignment = Alignment.CenterVertically) {
                        Text(if (hasLocation) "✅" else "📍", fontSize = 24.sp)
                        Spacer(modifier = Modifier.width(12.dp))
                        Column {
                            Text(
                                text = if (hasLocation) "已获取定位" else "未获取定位",
                                style = MaterialTheme.typography.titleSmall,
                                fontWeight = FontWeight.Medium,
                                color = if (hasLocation) Color(0xFF22C55E) else MaterialTheme.colorScheme.onSurface
                            )
                            Spacer(modifier = Modifier.height(4.dp))
                            Text(
                                text = if (hasLocation) "经度: 116.4074 | 纬度: 39.9042" else "点击获取当前位置",
                                style = MaterialTheme.typography.bodySmall,
                                color = MaterialTheme.colorScheme.onSurfaceVariant
                            )
                        }
                    }
                    Box(
                        modifier = Modifier
                            .background(
                                if (hasLocation)
                                    Color(0xFF22C55E).copy(alpha = 0.1f)
                                else
                                    MaterialTheme.colorScheme.primary.copy(alpha = 0.1f),
                                RoundedCornerShape(12.dp)
                            )
                            .clickable { hasLocation = !hasLocation }
                            .padding(horizontal = 16.dp, vertical = 10.dp)
                    ) {
                        Text(
                            text = if (hasLocation) "已定位" else "获取定位",
                            style = MaterialTheme.typography.labelMedium,
                            fontWeight = FontWeight.SemiBold,
                            color = if (hasLocation) Color(0xFF22C55E) else MaterialTheme.colorScheme.primary
                        )
                    }
                }
            }

            AppCard(title = "现场照片") {
                Row(
                    modifier = Modifier.fillMaxWidth(),
                    horizontalArrangement = Arrangement.spacedBy(12.dp)
                ) {
                    if (photoCount > 0) {
                        repeat(photoCount) { index ->
                            PhotoItem(
                                index = index + 1,
                                onRemove = { photoCount-- }
                            )
                        }
                    }
                    Box(
                        modifier = Modifier
                            .size(80.dp)
                            .background(
                                MaterialTheme.colorScheme.surfaceVariant,
                                RoundedCornerShape(12.dp)
                            )
                            .clickable { if (photoCount < 9) photoCount++ },
                        contentAlignment = Alignment.Center
                    ) {
                        Column(horizontalAlignment = Alignment.CenterHorizontally) {
                            Text("📷", fontSize = 28.sp)
                            Spacer(modifier = Modifier.height(4.dp))
                            Text(
                                text = "$photoCount/9",
                                style = MaterialTheme.typography.labelSmall,
                                color = MaterialTheme.colorScheme.onSurfaceVariant
                            )
                        }
                    }
                }
            }

            Spacer(modifier = Modifier.height(24.dp))
        }

        BottomSubmitButton(
            enabled = canSubmit,
            isSubmitting = isSubmitting,
            onClick = {
                isSubmitting = true
                appState.appScope.launch {
                    try {
                        val now = Clock.System.now().toLocalDateTime(TimeZone.currentSystemDefault())
                        val request = com.dispute.app.api.ReportHazardRequest(
                            workerId = "gw001",
                            hazardType = selectedHazardType!!,
                            urgencyLevel = selectedUrgency ?: HazardReport.UrgencyLevel.NORMAL,
                            title = hazardTitle,
                            description = hazardDescription,
                            location = hazardLocation,
                            reporterName = reporterName.ifBlank { null },
                            reporterPhone = reporterPhone.ifBlank { null },
                            longitude = if (hasLocation) 116.4074 else null,
                            latitude = if (hasLocation) 39.9042 else null
                        )

                        val response = apiClient.gridWorker.reportHazard(request)

                        appState.addHazardReport(response)
                        appState.showToast("隐患上报成功")
                        router.back()
                    } catch (e: Exception) {
                        appState.showToast("提交失败: ${e.message}")
                    } finally {
                        isSubmitting = false
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
private fun FormInputField(
    label: String,
    value: String,
    onValueChange: (String) -> Unit,
    placeholder: String,
    required: Boolean = false,
    keyboardType: KeyboardType = KeyboardType.Text
) {
    Column {
        Row {
            Text(
                text = label,
                style = MaterialTheme.typography.titleSmall,
                fontWeight = FontWeight.Medium
            )
            if (required) {
                Text(
                    text = " *",
                    color = Color(0xFFEF4444),
                    style = MaterialTheme.typography.titleSmall
                )
            }
        }
        Spacer(modifier = Modifier.height(8.dp))
        OutlinedTextField(
            value = value,
            onValueChange = onValueChange,
            modifier = Modifier.fillMaxWidth(),
            placeholder = {
                Text(
                    text = placeholder,
                    style = MaterialTheme.typography.bodyMedium,
                    color = MaterialTheme.colorScheme.onSurfaceVariant
                )
            },
            colors = TextFieldDefaults.colors(
                focusedContainerColor = Color.Transparent,
                unfocusedContainerColor = Color.Transparent
            ),
            keyboardOptions = KeyboardOptions(keyboardType = keyboardType),
            singleLine = true
        )
    }
}

@Composable
private fun PhotoItem(index: Int, onRemove: () -> Unit) {
    Box(
        modifier = Modifier.size(80.dp)
    ) {
        Box(
            modifier = Modifier
                .fillMaxSize()
                .background(
                    androidx.compose.ui.graphics.Brush.linearGradient(
                        colors = listOf(
                            Color(0xFFFFCCBC),
                            Color(0xFFFFAB91)
                        )
                    ),
                    RoundedCornerShape(12.dp)
                ),
            contentAlignment = Alignment.Center
        ) {
            Column(horizontalAlignment = Alignment.CenterHorizontally) {
                Text("✅", fontSize = 24.sp)
                Spacer(modifier = Modifier.height(2.dp))
                Text(
                    text = "已上传",
                    style = MaterialTheme.typography.labelSmall,
                    color = Color(0xFFE64A19),
                    fontSize = 10.sp
                )
            }
        }
        Box(
            modifier = Modifier
                .size(20.dp)
                .background(Color(0xFFEF4444), RoundedCornerShape(10.dp))
                .align(Alignment.TopEnd)
                .clickable(onClick = onRemove),
            contentAlignment = Alignment.Center
        ) {
            Text("×", color = Color.White, fontSize = 14.sp, fontWeight = FontWeight.Bold)
        }
    }
}

@Composable
private fun BottomSubmitButton(
    enabled: Boolean,
    isSubmitting: Boolean,
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
                    if (enabled)
                        androidx.compose.ui.graphics.Brush.linearGradient(
                            colors = listOf(
                                Color(0xFFF59E0B),
                                Color(0xFFFBBF24)
                            )
                        )
                    else
                        androidx.compose.ui.graphics.Brush.linearGradient(
                            colors = listOf(
                                Color(0xFF9CA3AF),
                                Color(0xFFD1D5DB)
                            )
                        ),
                    RoundedCornerShape(28.dp)
                )
                .clickable(enabled = enabled, onClick = onClick)
                .padding(vertical = 16.dp),
            contentAlignment = Alignment.Center
        ) {
            if (isSubmitting) {
                Row(verticalAlignment = Alignment.CenterVertically) {
                    androidx.compose.material3.CircularProgressIndicator(
                        color = Color.White,
                        strokeWidth = 2.dp,
                        modifier = Modifier.size(20.dp)
                    )
                    Spacer(modifier = Modifier.width(8.dp))
                    Text(
                        text = "提交中...",
                        color = Color.White,
                        style = MaterialTheme.typography.titleMedium,
                        fontWeight = FontWeight.SemiBold
                    )
                }
            } else {
                Text(
                    text = if (enabled) "提交隐患上报" else "请填写完整信息",
                    color = Color.White,
                    style = MaterialTheme.typography.titleMedium,
                    fontWeight = FontWeight.SemiBold
                )
            }
        }
    }
}
