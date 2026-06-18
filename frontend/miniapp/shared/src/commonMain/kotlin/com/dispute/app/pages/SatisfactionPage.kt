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
import com.dispute.app.components.RatingBar
import com.dispute.app.components.StarRating
import com.dispute.app.model.Case
import com.dispute.app.model.MockData
import kotlinx.coroutines.launch

@Composable
fun SatisfactionPage(caseNumber: String) {
    val appState = androidx.compose.runtime.remember { com.dispute.app.AppState() }
    val router = androidx.compose.runtime.remember { com.dispute.app.Router(appState) }

    CompositionLocalProvider(
        LocalAppState provides appState,
        LocalRouter provides router
    ) {
        SatisfactionContent(caseNumber)
    }
}

@Composable
private fun SatisfactionContent(caseNumber: String) {
    val appState = LocalAppState.current
    val router = LocalRouter.current

    val case = MockData.mockCases.firstOrNull { it.caseNumber == caseNumber }

    var overallRating by remember { mutableStateOf(0) }
    var mediatorRating by remember { mutableStateOf(0) }
    var efficiencyRating by remember { mutableStateOf(0) }
    var comment by remember { mutableStateOf("") }
    var isAnonymous by remember { mutableStateOf(false) }
    var submitted by remember { mutableStateOf(false) }

    val canSubmit = overallRating > 0

    Column(
        modifier = Modifier
            .fillMaxSize()
            .background(MaterialTheme.colorScheme.background)
    ) {
        TopBarSatisfaction(
            title = "服务评价",
            onBack = { router.back() }
        )

        if (submitted) {
            Column(
                modifier = Modifier
                    .fillMaxWidth()
                    .weight(1f),
                horizontalAlignment = Alignment.CenterHorizontally,
                verticalArrangement = Arrangement.Center
            ) {
                Box(
                    modifier = Modifier
                        .size(100.dp)
                        .background(Color(0xFF22C55E), RoundedCornerShape(50)),
                    contentAlignment = Alignment.Center
                ) {
                    Text("✓", color = Color.White, fontSize = 56.sp, fontWeight = FontWeight.Bold)
                }
                Spacer(modifier = Modifier.height(24.dp))
                Text(
                    "感谢您的评价！",
                    style = MaterialTheme.typography.headlineMedium,
                    fontWeight = FontWeight.Bold
                )
                Spacer(modifier = Modifier.height(12.dp))
                Text(
                    "您的反馈是我们改进服务的动力",
                    style = MaterialTheme.typography.bodyMedium,
                    color = MaterialTheme.colorScheme.onSurfaceVariant
                )
                Spacer(modifier = Modifier.height(32.dp))
                Button(
                    onClick = { router.navigateWithReplace(Route.Home) },
                    modifier = Modifier
                        .width(220.dp)
                        .height(48.dp),
                    shape = RoundedCornerShape(24.dp),
                    colors = ButtonDefaults.buttonColors(containerColor = MaterialTheme.colorScheme.primary)
                ) {
                    Text("返回首页", fontWeight = FontWeight.SemiBold)
                }
            }
        } else {
            Column(
                modifier = Modifier
                    .fillMaxWidth()
                    .weight(1f)
                    .padding(16.dp),
                verticalArrangement = Arrangement.spacedBy(16.dp)
            ) {
                if (case != null) {
                    CaseInfoSummary(case = case)
                }

                RatingSection(
                    title = "整体满意度",
                    description = "您对本次调解服务的总体印象如何？",
                    rating = overallRating,
                    onRatingChanged = { overallRating = it }
                )

                RatingSection(
                    title = "调解员服务",
                    description = "调解员的专业水平、态度和沟通能力",
                    rating = mediatorRating,
                    onRatingChanged = { mediatorRating = it }
                )

                RatingSection(
                    title = "办理效率",
                    description = "案件处理速度和响应及时性",
                    rating = efficiencyRating,
                    onRatingChanged = { efficiencyRating = it }
                )

                com.dispute.app.components.AppCard(title = "详细评价（选填）") {
                    Column {
                        val quickTags = listOf(
                            "调解员专业", "响应速度快", "沟通耐心",
                            "环境舒适", "流程清晰", "结果满意"
                        )
                        var selectedTags by remember { mutableStateOf(setOf<String>()) }

                        Text(
                            "快捷标签：",
                            style = MaterialTheme.typography.labelLarge,
                            color = MaterialTheme.colorScheme.onSurfaceVariant,
                            modifier = Modifier.padding(bottom = 8.dp)
                        )
                        Row(
                            horizontalArrangement = Arrangement.spacedBy(8.dp),
                            modifier = androidx.compose.foundation.layout.padding(bottom = 12.dp)
                        ) {
                            quickTags.take(3).forEach { tag ->
                                val selected = selectedTags.contains(tag)
                                Box(
                                    modifier = Modifier
                                        .background(
                                            if (selected) MaterialTheme.colorScheme.primary.copy(alpha = 0.12f)
                                            else MaterialTheme.colorScheme.surfaceVariant.copy(alpha = 0.5f),
                                            RoundedCornerShape(16.dp)
                                        )
                                        .clickable {
                                            selectedTags = if (selected) {
                                                selectedTags - tag
                                            } else selectedTags + tag
                                            comment = selectedTags.joinToString("、")
                                        }
                                        .padding(horizontal = 12.dp, vertical = 8.dp)
                                ) {
                                    Text(
                                        tag,
                                        fontSize = 12.sp,
                                        color = if (selected) MaterialTheme.colorScheme.primary
                                        else MaterialTheme.colorScheme.onSurfaceVariant
                                    )
                                }
                            }
                        }
                        Row(horizontalArrangement = Arrangement.spacedBy(8.dp)) {
                            quickTags.drop(3).forEach { tag ->
                                val selected = selectedTags.contains(tag)
                                Box(
                                    modifier = Modifier
                                        .background(
                                            if (selected) MaterialTheme.colorScheme.primary.copy(alpha = 0.12f)
                                            else MaterialTheme.colorScheme.surfaceVariant.copy(alpha = 0.5f),
                                            RoundedCornerShape(16.dp)
                                        )
                                        .clickable {
                                            selectedTags = if (selected) selectedTags - tag
                                            else selectedTags + tag
                                            comment = selectedTags.joinToString("、")
                                        }
                                        .padding(horizontal = 12.dp, vertical = 8.dp)
                                ) {
                                    Text(
                                        tag,
                                        fontSize = 12.sp,
                                        color = if (selected) MaterialTheme.colorScheme.primary
                                        else MaterialTheme.colorScheme.onSurfaceVariant
                                    )
                                }
                            }
                        }

                        Spacer(modifier = Modifier.height(14.dp))
                        Text(
                            "您的建议：",
                            style = MaterialTheme.typography.labelLarge,
                            color = MaterialTheme.colorScheme.onSurfaceVariant,
                            modifier = Modifier.padding(bottom = 6.dp)
                        )
                        OutlinedTextField(
                            value = comment,
                            onValueChange = { comment = it },
                            placeholder = {
                                Text(
                                    "请分享您的宝贵意见和建议，帮助我们做得更好...",
                                    color = MaterialTheme.colorScheme.onSurfaceVariant,
                                    fontSize = 12.sp
                                )
                            },
                            modifier = Modifier
                                .fillMaxWidth()
                                .height(110.dp),
                            shape = RoundedCornerShape(12.dp),
                            colors = OutlinedTextFieldDefaults.colors(
                                focusedBorderColor = MaterialTheme.colorScheme.primary,
                                unfocusedBorderColor = Color(0xFFE5E7EB)
                            )
                        )
                    }
                }

                Row(
                    verticalAlignment = Alignment.CenterVertically,
                    modifier = Modifier
                        .fillMaxWidth()
                        .background(Color(0xFFFFFBEB), RoundedCornerShape(12.dp))
                        .clickable { isAnonymous = !isAnonymous }
                        .padding(14.dp)
                ) {
                    Box(
                        modifier = Modifier
                            .size(20.dp)
                            .background(
                                if (isAnonymous) MaterialTheme.colorScheme.primary else Color.White,
                                RoundedCornerShape(6.dp)
                            )
                            .then(
                                if (!isAnonymous) Modifier.background(
                                    Color.White,
                                    RoundedCornerShape(6.dp)
                                ) else Modifier
                            ),
                        contentAlignment = Alignment.Center
                    ) {
                        if (isAnonymous) {
                            Text("✓", color = Color.White, fontSize = 13.sp, fontWeight = FontWeight.Bold)
                        }
                    }
                    Spacer(modifier = Modifier.width(10.dp))
                    Text(
                        "匿名提交（您的身份信息不会被记录）",
                        style = MaterialTheme.typography.bodyMedium,
                        color = Color(0xFF92400E),
                        modifier = Modifier.weight(1f)
                    )
                }
            }

            Column(
                modifier = Modifier
                    .fillMaxWidth()
                    .background(Color.White)
                    .padding(horizontal = 16.dp, vertical = 12.dp)
            ) {
                Button(
                    onClick = {
                        appState.launchWithLoading {
                            kotlinx.coroutines.delay(800)
                            appState.showToast("评价提交成功，感谢您的反馈！")
                            kotlinx.coroutines.delay(500)
                            submitted = true
                            if (case != null) {
                                appState.updateCase(case.caseNumber) {
                                    it.copy(
                                        satisfactionRating = overallRating,
                                        satisfactionComment = comment
                                    )
                                }
                            }
                        }
                    },
                    modifier = Modifier
                        .fillMaxWidth()
                        .height(50.dp),
                    shape = RoundedCornerShape(14.dp),
                    enabled = canSubmit,
                    colors = ButtonDefaults.buttonColors(
                        containerColor = MaterialTheme.colorScheme.primary,
                        disabledContainerColor = MaterialTheme.colorScheme.primary.copy(alpha = 0.35f)
                    )
                ) {
                    Text("提交评价", fontWeight = FontWeight.SemiBold, fontSize = 15.sp)
                }
            }
        }
    }
}

@Composable
private fun TopBarSatisfaction(title: String, onBack: () -> Unit) {
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
private fun CaseInfoSummary(case: Case) {
    val statusColor = Color(case.statusColor)
    com.dispute.app.components.AppCard {
        Column {
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.SpaceBetween,
                verticalAlignment = Alignment.CenterVertically
            ) {
                Column {
                    Text(
                        case.caseNumber,
                        style = MaterialTheme.typography.titleSmall,
                        fontWeight = FontWeight.SemiBold
                    )
                    Spacer(modifier = Modifier.height(4.dp))
                    Text(
                        case.fullTypeName,
                        style = MaterialTheme.typography.labelMedium,
                        color = MaterialTheme.colorScheme.onSurfaceVariant
                    )
                }
                Box(
                    modifier = Modifier
                        .background(statusColor.copy(alpha = 0.15f), RoundedCornerShape(8.dp))
                        .padding(horizontal = 10.dp, vertical = 4.dp)
                ) {
                    Text(
                        case.status.displayName,
                        color = statusColor,
                        fontSize = 12.sp,
                        fontWeight = FontWeight.Bold
                    )
                }
            }
            androidx.compose.material3.Divider(
                color = Color(0xFFF3F4F6),
                modifier = Modifier.padding(vertical = 12.dp)
            )
            Row(horizontalArrangement = Arrangement.spacedBy(24.dp)) {
                if (case.mediatorName != null) {
                    InfoMini("调解员", case.mediatorName!!)
                }
                InfoMini("结案时间", case.lastUpdateTime?.substringBefore(" ") ?: "—")
            }
        }
    }
}

@Composable
private fun InfoMini(label: String, value: String) {
    Column {
        Text(
            label,
            style = MaterialTheme.typography.labelSmall,
            color = Color(0xFF9CA3AF)
        )
        Text(
            value,
            style = MaterialTheme.typography.bodyMedium,
            fontWeight = FontWeight.Medium
        )
    }
}

@Composable
private fun RatingSection(
    title: String,
    description: String,
    rating: Int,
    onRatingChanged: (Int) -> Unit
) {
    com.dispute.app.components.AppCard {
        Column {
            Text(
                title,
                style = MaterialTheme.typography.titleMedium,
                fontWeight = FontWeight.Bold
            )
            Spacer(modifier = Modifier.height(4.dp))
            Text(
                description,
                style = MaterialTheme.typography.bodySmall,
                color = MaterialTheme.colorScheme.onSurfaceVariant
            )
            Spacer(modifier = Modifier.height(16.dp))
            StarRating(
                rating = rating,
                onRatingChanged = onRatingChanged,
                starSize = 38,
                showLabels = true
            )
        }
    }
}
