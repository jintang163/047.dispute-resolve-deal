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
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.text.KeyboardOptions
import androidx.compose.material3.Button
import androidx.compose.material3.ButtonDefaults
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.OutlinedTextField
import androidx.compose.material3.OutlinedTextFieldDefaults
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
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.input.KeyboardType
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.dispute.app.LocalAppState
import com.dispute.app.LocalRouter
import com.dispute.app.Route
import com.dispute.app.components.AppCard
import com.dispute.app.components.Timeline
import com.dispute.app.model.Case
import com.dispute.app.model.MockData

@Composable
fun ProgressPage(initialCaseNo: String = "", initialPhone: String? = null) = ProgressContent(initialCaseNo = initialCaseNo, initialPhone = initialPhone)

@Composable
private fun ProgressContent(initialCaseNo: String = "", initialPhone: String? = null) {
    val appState = LocalAppState.current
    val router = LocalRouter.current

    var caseNumberInput by remember { mutableStateOf(initialCaseNo) }
    var phoneInput by remember { mutableStateOf(initialPhone ?: "") }
    var resultCase by remember { mutableStateOf<Case?>(null) }
    var showResult by remember { mutableStateOf(false) }
    var searchError by remember { mutableStateOf<String?>(null) }
    var autoSearched by remember { mutableStateOf(false) }

    val recentCases = MockData.mockCases.take(3)
    val progress = resultCase?.let { MockData.mockProgress.filter { p -> p.caseNumber == it.caseNumber } }

    LaunchedEffect(initialCaseNo) {
        if (initialCaseNo.isNotBlank() && !autoSearched) {
            autoSearched = true
            appState.launchWithLoading {
                kotlinx.coroutines.delay(500)
                val found = MockData.mockCases.firstOrNull {
                    it.caseNumber == initialCaseNo
                }
                if (found != null) {
                    resultCase = found
                    showResult = true
                } else {
                    searchError = "扫码跳转：案件编号 $initialCaseNo 暂未查询到结果，请确认编号是否正确"
                    showResult = false
                }
            }
        }
    }

    Column(
        modifier = Modifier
            .fillMaxSize()
            .background(MaterialTheme.colorScheme.background)
    ) {
        TopBarProgress(
            title = "进度查询",
            onBack = { router.back() }
        )

        Column(
            modifier = Modifier
                .fillMaxWidth()
                .weight(1f)
                .padding(16.dp),
            verticalArrangement = Arrangement.spacedBy(16.dp)
        ) {
            AppCard(
                title = "输入查询信息",
                subtitle = "请输入案件编号和手机号查询调解进度"
            ) {
                Column(verticalArrangement = Arrangement.spacedBy(14.dp)) {
                    LabeledField(
                        label = "案件编号",
                        value = caseNumberInput,
                        onValueChange = { caseNumberInput = it.uppercase() },
                        placeholder = "如：JF202412010001"
                    )
                    LabeledField(
                        label = "手机号",
                        value = phoneInput,
                        onValueChange = { if (it.length <= 11) phoneInput = it },
                        placeholder = "请输入登记时的手机号",
                        isPhone = true
                    )
                    Button(
                        onClick = {
                            searchError = null
                            when {
                                caseNumberInput.isBlank() -> searchError = "请输入案件编号"
                                !caseNumberInput.matches(Regex("^JF\\d{12}$")) ->
                                    searchError = "案件编号格式不正确"
                                phoneInput.isBlank() -> searchError = "请输入手机号"
                                !phoneInput.matches(Regex("^1[3-9]\\d{9}$")) ->
                                    searchError = "请输入正确的手机号"
                                else -> {
                                    appState.launchWithLoading {
                                        kotlinx.coroutines.delay(800)
                                        val found = MockData.mockCases.firstOrNull {
                                            it.caseNumber == caseNumberInput
                                        }
                                        if (found != null) {
                                            resultCase = found
                                            showResult = true
                                        } else {
                                            searchError = "未找到对应案件，请确认信息是否正确"
                                            showResult = false
                                        }
                                    }
                                }
                            }
                        },
                        modifier = Modifier
                            .fillMaxWidth()
                            .height(50.dp),
                        shape = RoundedCornerShape(14.dp),
                        colors = ButtonDefaults.buttonColors(
                            containerColor = MaterialTheme.colorScheme.primary
                        )
                    ) {
                        Text("查询进度", fontWeight = FontWeight.SemiBold, fontSize = 15.sp)
                    }
                    searchError?.let {
                        Box(
                            modifier = Modifier
                                .fillMaxWidth()
                                .background(Color(0xFFFEF2F2), RoundedCornerShape(10.dp))
                                .padding(12.dp)
                        ) {
                            Row(verticalAlignment = Alignment.CenterVertically) {
                                Text("⚠️", fontSize = 16.sp)
                                Spacer(modifier = Modifier.width(8.dp))
                                Text(it, color = Color(0xFFDC2626), fontSize = 13.sp)
                            }
                        }
                    }
                }
            }

            if (showResult && resultCase != null) {
                val statusColor = Color(resultCase!!.statusColor)
                AppCard(title = "查询结果") {
                    Column(verticalArrangement = Arrangement.spacedBy(12.dp)) {
                        Row(
                            modifier = Modifier.fillMaxWidth(),
                            horizontalArrangement = Arrangement.SpaceBetween,
                            verticalAlignment = Alignment.CenterVertically
                        ) {
                            Column {
                                Text(
                                    text = resultCase!!.caseNumber,
                                    style = MaterialTheme.typography.titleSmall,
                                    fontWeight = FontWeight.SemiBold
                                )
                                Spacer(modifier = Modifier.height(4.dp))
                                Text(
                                    text = resultCase!!.fullTypeName,
                                    style = MaterialTheme.typography.labelMedium,
                                    color = MaterialTheme.colorScheme.onSurfaceVariant
                                )
                            }
                            Box(
                                modifier = Modifier
                                    .background(statusColor.copy(alpha = 0.15f), RoundedCornerShape(8.dp))
                                    .padding(horizontal = 10.dp, vertical = 5.dp)
                            ) {
                                Text(
                                    text = resultCase!!.status.displayName,
                                    color = statusColor,
                                    fontSize = 12.sp,
                                    fontWeight = FontWeight.Bold
                                )
                            }
                        }

                        androidx.compose.material3.Divider(color = Color(0xFFF3F4F6))

                        Row(horizontalArrangement = Arrangement.spacedBy(8.dp)) {
                            TagChip("调解员: ${resultCase!!.mediatorName ?: "分配中"}")
                            TagChip("提交: ${resultCase!!.submitTime.substringBefore(" ")}")
                        }

                        if (!progress.isNullOrEmpty()) {
                            Spacer(modifier = Modifier.height(8.dp))
                            Timeline(progressList = progress)
                        }

                        Spacer(modifier = Modifier.height(4.dp))
                        Button(
                            onClick = {
                                router.navigate(Route.CaseDetail(resultCase!!.caseNumber))
                            },
                            modifier = Modifier
                                .fillMaxWidth()
                                .height(44.dp),
                            shape = RoundedCornerShape(12.dp),
                            colors = ButtonDefaults.buttonColors(
                                containerColor = MaterialTheme.colorScheme.primary.copy(alpha = 0.1f),
                                contentColor = MaterialTheme.colorScheme.primary
                            )
                        ) {
                            Text("查看完整详情 ›", fontWeight = FontWeight.SemiBold)
                        }
                    }
                }
            }

            AppCard(title = "最近查询", subtitle = "快速查看最近登记的案件") {
                Column(verticalArrangement = Arrangement.spacedBy(10.dp)) {
                    recentCases.forEach { case ->
                        RecentSearchItem(
                            caseNumber = case.caseNumber,
                            typeName = case.fullTypeName,
                            status = case.status.displayName,
                            statusColor = Color(case.statusColor),
                            onClick = {
                                caseNumberInput = case.caseNumber
                                phoneInput = case.applicantPhone
                                resultCase = case
                                showResult = true
                                searchError = null
                            }
                        )
                    }
                }
            }

            AppCard(
                backgroundColor = Color(0xFFEFF6FF),
                borderColor = Color(0xFFBFDBFE)
            ) {
                Column {
                    Row(verticalAlignment = Alignment.Top) {
                        Text("💡", fontSize = 18.sp, modifier = Modifier.padding(end = 8.dp))
                        Column {
                            Text(
                                text = "温馨提示",
                                fontWeight = FontWeight.Bold,
                                color = Color(0xFF1E40AF)
                            )
                            Spacer(modifier = Modifier.height(4.dp))
                            Text(
                                text = "• 案件编号可在登记回执或短信中找到\n" +
                                    "• 如忘记编号，可拨打服务热线12348查询\n" +
                                    "• 调解进度一般1-3个工作日更新一次",
                                style = MaterialTheme.typography.bodySmall,
                                color = Color(0xFF1E40AF),
                                lineHeight = 18.sp
                            )
                        }
                    }
                }
            }
        }
    }
}

@Composable
private fun TopBarProgress(title: String, onBack: () -> Unit) {
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
private fun LabeledField(
    label: String,
    value: String,
    onValueChange: (String) -> Unit,
    placeholder: String,
    isPhone: Boolean = false
) {
    Column {
        Text(
            text = label,
            style = MaterialTheme.typography.labelLarge,
            fontWeight = FontWeight.SemiBold,
            modifier = Modifier.padding(bottom = 6.dp)
        )
        OutlinedTextField(
            value = value,
            onValueChange = onValueChange,
            placeholder = {
                Text(placeholder, color = MaterialTheme.colorScheme.onSurfaceVariant, fontSize = 13.sp)
            },
            modifier = Modifier.fillMaxWidth(),
            shape = RoundedCornerShape(12.dp),
            keyboardOptions = KeyboardOptions(
                keyboardType = if (isPhone) KeyboardType.Phone else KeyboardType.Text
            ),
            colors = OutlinedTextFieldDefaults.colors(
                focusedBorderColor = MaterialTheme.colorScheme.primary,
                unfocusedBorderColor = Color(0xFFE5E7EB)
            )
        )
    }
}

@Composable
private fun TagChip(text: String) {
    Box(
        modifier = Modifier
            .background(
                MaterialTheme.colorScheme.surfaceVariant.copy(alpha = 0.6f),
                RoundedCornerShape(8.dp)
            )
            .padding(horizontal = 10.dp, vertical = 6.dp)
    ) {
        Text(text, fontSize = 12.sp, color = MaterialTheme.colorScheme.onSurfaceVariant)
    }
}

@Composable
private fun RecentSearchItem(
    caseNumber: String,
    typeName: String,
    status: String,
    statusColor: Color,
    onClick: () -> Unit
) {
    Row(
        modifier = Modifier
            .fillMaxWidth()
            .background(
                MaterialTheme.colorScheme.surfaceVariant.copy(alpha = 0.3f),
                RoundedCornerShape(12.dp)
            )
            .clickable(onClick = onClick)
            .padding(14.dp),
        verticalAlignment = Alignment.CenterVertically
    ) {
        Box(
            modifier = Modifier
                .size(40.dp)
                .background(Color.White, RoundedCornerShape(10.dp)),
            contentAlignment = Alignment.Center
        ) {
            Text("📋", fontSize = 20.sp)
        }
        Spacer(modifier = Modifier.width(12.dp))
        Column(modifier = Modifier.weight(1f)) {
            Text(
                text = caseNumber,
                style = MaterialTheme.typography.titleSmall,
                fontWeight = FontWeight.SemiBold
            )
            Spacer(modifier = Modifier.height(3.dp))
            Text(
                text = typeName,
                style = MaterialTheme.typography.labelMedium,
                color = MaterialTheme.colorScheme.onSurfaceVariant
            )
        }
        Box(
            modifier = Modifier
                .background(statusColor.copy(alpha = 0.15f), RoundedCornerShape(6.dp))
                .padding(horizontal = 8.dp, vertical = 4.dp)
        ) {
            Text(status, fontSize = 11.sp, color = statusColor, fontWeight = FontWeight.SemiBold)
        }
    }
}
