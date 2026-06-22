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
import androidx.compose.foundation.lazy.LazyRow
import androidx.compose.foundation.lazy.items
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
import com.dispute.app.model.GridWorkerMockData
import com.dispute.app.model.VisitRecord
import kotlinx.coroutines.launch
import kotlinx.datetime.Clock
import kotlinx.datetime.TimeZone
import kotlinx.datetime.toLocalDateTime

@Composable
fun VisitRecordAddPage() = VisitRecordAddContent()

@Composable
private fun VisitRecordAddContent() {
    val appState = LocalAppState.current
    val router = LocalRouter.current
    val apiClient = LocalApiClient.current

    var residentName by remember { mutableStateOf("") }
    var residentPhone by remember { mutableStateOf("") }
    var residentAddress by remember { mutableStateOf("") }
    var selectedVisitType by remember { mutableStateOf<VisitRecord.VisitType?>(null) }
    var visitContent by remember { mutableStateOf("") }
    var visitResult by remember { mutableStateOf("") }
    var hasLocation by remember { mutableStateOf(false) }
    var photoCount by remember { mutableStateOf(0) }
    var isSubmitting by remember { mutableStateOf(false) }

    val visitTypes = VisitRecord.VisitType.values().toList()

    val canSubmit = residentName.isNotBlank() && residentAddress.isNotBlank() &&
            selectedVisitType != null && visitContent.isNotBlank() && !isSubmitting

    Column(
        modifier = Modifier
            .fillMaxSize()
            .background(MaterialTheme.colorScheme.background)
    ) {
        TopBar(
            title = "新增走访记录",
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
            AppCard(title = "居民信息") {
                Column(verticalArrangement = Arrangement.spacedBy(12.dp)) {
                    FormInputField(
                        label = "居民姓名",
                        value = residentName,
                        onValueChange = { residentName = it },
                        placeholder = "请输入居民姓名",
                        required = true
                    )

                    FormInputField(
                        label = "联系电话",
                        value = residentPhone,
                        onValueChange = { residentPhone = it },
                        placeholder = "请输入联系电话（选填）",
                        keyboardType = KeyboardType.Phone
                    )

                    FormInputField(
                        label = "居住地址",
                        value = residentAddress,
                        onValueChange = { residentAddress = it },
                        placeholder = "请输入居住地址",
                        required = true
                    )
                }
            }

            AppCard(title = "走访类型") {
                Column(verticalArrangement = Arrangement.spacedBy(12.dp)) {
                    val chunks = visitTypes.chunked(3)
                    chunks.forEach { chunk ->
                        Row(
                            modifier = Modifier.fillMaxWidth(),
                            horizontalArrangement = Arrangement.spacedBy(8.dp)
                        ) {
                            chunk.forEach { type ->
                                val isSelected = selectedVisitType == type
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
                                        .clickable { selectedVisitType = type }
                                        .padding(vertical = 12.dp),
                                    contentAlignment = Alignment.Center
                                ) {
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

            AppCard(title = "走访内容", subtitle = "请详细记录走访情况") {
                Column {
                    OutlinedTextField(
                        value = visitContent,
                        onValueChange = { visitContent = it },
                        modifier = Modifier
                            .fillMaxWidth()
                            .height(120.dp),
                        placeholder = {
                            Text(
                                text = "请输入走访内容详情...",
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
            }

            AppCard(title = "走访结果", subtitle = "请记录走访结果（选填）") {
                Column {
                    OutlinedTextField(
                        value = visitResult,
                        onValueChange = { visitResult = it },
                        modifier = Modifier
                            .fillMaxWidth()
                            .height(80.dp),
                        placeholder = {
                            Text(
                                text = "请输入走访结果...",
                                style = MaterialTheme.typography.bodyMedium,
                                color = MaterialTheme.colorScheme.onSurfaceVariant
                            )
                        },
                        colors = TextFieldDefaults.colors(
                            focusedContainerColor = Color.Transparent,
                            unfocusedContainerColor = Color.Transparent
                        ),
                        maxLines = 3
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
                        val request = com.dispute.app.api.CreateVisitRequest(
                            workerId = "gw001",
                            residentName = residentName,
                            residentPhone = residentPhone.ifBlank { null },
                            residentAddress = residentAddress,
                            visitType = selectedVisitType!!,
                            visitContent = visitContent,
                            visitResult = visitResult.ifBlank { null },
                            longitude = if (hasLocation) 116.4074 else null,
                            latitude = if (hasLocation) 39.9042 else null
                        )

                        val response = apiClient.gridWorker.createVisitRecord(request)

                        appState.addVisitRecord(response)
                        appState.showToast("走访记录提交成功")
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
                            Color(0xFFC8E6C9),
                            Color(0xFFA5D6A7)
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
                    color = Color(0xFF2E7D32),
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
                                Color(0xFF1D6CFF),
                                Color(0xFF4D8CFF)
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
                    text = if (enabled) "提交走访记录" else "请填写完整信息",
                    color = Color.White,
                    style = MaterialTheme.typography.titleMedium,
                    fontWeight = FontWeight.SemiBold
                )
            }
        }
    }
}
