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
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.Button
import androidx.compose.material3.ButtonDefaults
import androidx.compose.material3.MaterialTheme
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
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.dispute.app.LocalAppState
import com.dispute.app.LocalRouter
import com.dispute.app.Route
import com.dispute.app.components.AppCard
import com.dispute.app.components.InfoCard
import com.dispute.app.components.Timeline
import com.dispute.app.model.Case
import com.dispute.app.model.MockData
import kotlinx.coroutines.launch

@Composable
fun CaseDetailPage(caseNumber: String) {
    val appState = androidx.compose.runtime.remember { com.dispute.app.AppState() }
    val router = androidx.compose.runtime.remember { com.dispute.app.Router(appState) }

    CompositionLocalProvider(
        LocalAppState provides appState,
        LocalRouter provides router
    ) {
        CaseDetailContent(caseNumber)
    }
}

@Composable
private fun CaseDetailContent(caseNumber: String) {
    val appState = LocalAppState.current
    val router = LocalRouter.current

    var case by remember { mutableStateOf<Case?>(null) }

    androidx.compose.runtime.LaunchedEffect(Unit) {
        val found = appState.findCase(caseNumber)
            ?: MockData.mockCases.firstOrNull { it.caseNumber == caseNumber }
        case = found
        if (found != null) {
            appState.setSelectedCase(found)
        }
    }

    val statusColor = case?.statusColor?.let { Color(it) }
    val progress = MockData.mockProgress.filter { it.caseNumber == caseNumber }

    Column(
        modifier = Modifier
            .fillMaxSize()
            .background(MaterialTheme.colorScheme.background)
    ) {
        TopBarDetail(
            title = "案件详情",
            onBack = { router.back() }
        )

        if (case == null) {
            com.dispute.app.components.EmptyCard(
                icon = "🔍",
                title = "案件不存在",
                description = "未找到编号为 $caseNumber 的案件"
            )
        } else {
            LazyColumn(
                modifier = Modifier
                    .fillMaxWidth()
                    .weight(1f)
                    .padding(horizontal = 16.dp, vertical = 12.dp),
                verticalArrangement = Arrangement.spacedBy(14.dp)
            ) {
                item {
                    CaseHeaderCard(case!!, statusColor!!)
                }

                item {
                    MediatorCard(case!!, appState, router)
                }

                item {
                    AppCard(title = "基本信息") {
                        Column(verticalArrangement = Arrangement.spacedBy(10.dp)) {
                            InfoCard(title = "案件编号", value = case!!.caseNumber)
                            InfoCard(title = "纠纷类型", value = case!!.fullTypeName)
                            InfoCard(title = "提交时间", value = case!!.submitTime)
                            case!!.lastUpdateTime?.let {
                                InfoCard(title = "最近更新", value = it)
                            }
                            case!!.estimatedDays?.let {
                                InfoCard(title = "预计完成", value = "$it 个工作日内")
                            }
                        }
                    }
                }

                item {
                    AppCard(title = "当事人信息") {
                        Column(verticalArrangement = Arrangement.spacedBy(10.dp)) {
                            InfoCard(title = "申请人", value = case!!.applicantName)
                            InfoCard(title = "申请电话", value = maskPhone(case!!.applicantPhone))
                            InfoCard(title = "对方姓名", value = case!!.opponentName)
                            case!!.opponentPhone?.let {
                                InfoCard(title = "对方电话", value = maskPhone(it))
                            }
                        }
                    }
                }

                item {
                    AppCard(title = "纠纷描述") {
                        Text(
                            text = case!!.description,
                            style = MaterialTheme.typography.bodyMedium,
                            lineHeight = 20.sp
                        )
                    }
                }

                if (case!!.expectedResolution != null) {
                    item {
                        AppCard(title = "期望解决方式") {
                            Text(
                                text = case!!.expectedResolution!!,
                                style = MaterialTheme.typography.bodyMedium,
                                color = MaterialTheme.colorScheme.primary,
                                fontWeight = FontWeight.Medium
                            )
                        }
                    }
                }

                item {
                    AppCard(title = "调解进度（${progress.size}条记录）") {
                        Timeline(progressList = progress)
                    }
                }

                if (case!!.evidenceList.isNotEmpty()) {
                    item {
                        AppCard(title = "证据材料（${case!!.evidenceList.size}个）") {
                            Column(verticalArrangement = Arrangement.spacedBy(8.dp)) {
                                case!!.evidenceList.forEach { evidence ->
                                    Row(
                                        modifier = Modifier
                                            .fillMaxWidth()
                                            .background(
                                                MaterialTheme.colorScheme.surfaceVariant.copy(alpha = 0.4f),
                                                RoundedCornerShape(10.dp)
                                            )
                                            .padding(12.dp),
                                        verticalAlignment = Alignment.CenterVertically
                                    ) {
                                        Text(evidence.type.icon, fontSize = 24.sp)
                                        Spacer(modifier = Modifier.width(10.dp))
                                        Column(modifier = Modifier.weight(1f)) {
                                            Text(
                                                text = evidence.name,
                                                style = MaterialTheme.typography.bodyMedium,
                                                fontWeight = FontWeight.Medium
                                            )
                                            Text(
                                                text = "${evidence.displaySize} · ${evidence.uploadTime}",
                                                style = MaterialTheme.typography.labelSmall,
                                                color = Color(0xFF9CA3AF)
                                            )
                                        }
                                    }
                                }
                            }
                        }
                    }
                }

                if (case!!.satisfactionRating != null) {
                    item {
                        AppCard(title = "服务评价", backgroundColor = Color(0xFFFEFCE8)) {
                            Column {
                                Row(verticalAlignment = Alignment.CenterVertically) {
                                    Text(
                                        text = "★".repeat(case!!.satisfactionRating!!) +
                                            "☆".repeat(5 - case!!.satisfactionRating!!),
                                        color = Color(0xFFFFD700),
                                        fontSize = 22.sp
                                    )
                                    Spacer(modifier = Modifier.width(8.dp))
                                    Text(
                                        text = getRatingText(case!!.satisfactionRating!!),
                                        fontWeight = FontWeight.SemiBold
                                    )
                                }
                                if (!case!!.satisfactionComment.isNullOrBlank()) {
                                    Spacer(modifier = Modifier.height(8.dp))
                                    Text(
                                        text = case!!.satisfactionComment!!,
                                        style = MaterialTheme.typography.bodyMedium,
                                        color = MaterialTheme.colorScheme.onSurfaceVariant
                                    )
                                }
                            }
                        }
                    }
                }

                item { Spacer(modifier = Modifier.height(20.dp)) }
            }

            BottomActionsDetail(
                case = case!!,
                appState = appState,
                router = router
            )
        }
    }
}

@Composable
private fun TopBarDetail(title: String, onBack: () -> Unit) {
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
private fun CaseHeaderCard(case: Case, statusColor: Color) {
    Column(
        modifier = Modifier
            .fillMaxWidth()
            .background(
                androidx.compose.ui.graphics.Brush.linearGradient(
                    listOf(
                        statusColor.copy(alpha = 0.9f),
                        statusColor.copy(alpha = 0.7f)
                    )
                ),
                RoundedCornerShape(20.dp)
            )
            .padding(20.dp)
    ) {
        Row(
            modifier = Modifier.fillMaxWidth(),
            horizontalArrangement = Arrangement.SpaceBetween,
            verticalAlignment = Alignment.CenterVertically
        ) {
            Column {
                Text(
                    text = case.status.displayName,
                    color = Color.White,
                    fontSize = 18.sp,
                    fontWeight = FontWeight.Bold
                )
                Spacer(modifier = Modifier.height(6.dp))
                Text(
                    text = case.caseNumber,
                    color = Color.White.copy(alpha = 0.9f),
                    fontSize = 13.sp
                )
            }
            Box(
                modifier = Modifier
                    .size(56.dp)
                    .background(Color.White.copy(alpha = 0.2f), RoundedCornerShape(16.dp)),
                contentAlignment = Alignment.Center
            ) {
                Text(getStatusIcon(case.status), fontSize = 28.sp)
            }
        }

        Spacer(modifier = Modifier.height(16.dp))

        Box(
            modifier = Modifier
                .fillMaxWidth()
                .background(Color.White.copy(alpha = 0.18f), RoundedCornerShape(12.dp))
                .padding(12.dp)
        ) {
            Row(verticalAlignment = Alignment.CenterVertically) {
                Text("📋", fontSize = 18.sp)
                Spacer(modifier = Modifier.width(8.dp))
                Text(
                    text = case.fullTypeName,
                    color = Color.White,
                    style = MaterialTheme.typography.titleSmall,
                    fontWeight = FontWeight.SemiBold
                )
            }
        }
    }
}

@Composable
private fun MediatorCard(case: Case, appState: com.dispute.app.AppState, router: com.dispute.app.Router) {
    if (case.mediatorName != null) {
        AppCard {
            Row(verticalAlignment = Alignment.CenterVertically) {
                Box(
                    modifier = Modifier
                        .size(52.dp)
                        .background(
                            MaterialTheme.colorScheme.primary.copy(alpha = 0.1f),
                            androidx.compose.foundation.shape.CircleShape
                        ),
                    contentAlignment = Alignment.Center
                ) {
                    Text(
                        text = case.mediatorName!!.firstOrNull()?.toString() ?: "调",
                        color = MaterialTheme.colorScheme.primary,
                        fontSize = 20.sp,
                        fontWeight = FontWeight.Bold
                    )
                }
                Spacer(modifier = Modifier.width(14.dp))
                Column(modifier = Modifier.weight(1f)) {
                    Row(verticalAlignment = Alignment.CenterVertically) {
                        Text(
                            text = case.mediatorName!!,
                            style = MaterialTheme.typography.titleMedium,
                            fontWeight = FontWeight.Bold
                        )
                        Spacer(modifier = Modifier.width(8.dp))
                        Box(
                            modifier = Modifier
                                .background(Color(0xFFDCFCE7), RoundedCornerShape(6.dp))
                                .padding(horizontal = 8.dp, vertical = 2.dp)
                        ) {
                            Text("调解员", fontSize = 11.sp, color = Color(0xFF15803D))
                        }
                    }
                    Spacer(modifier = Modifier.height(4.dp))
                    Text(
                        text = "专业调解人员 · 竭诚为您服务",
                        style = MaterialTheme.typography.labelMedium,
                        color = MaterialTheme.colorScheme.onSurfaceVariant
                    )
                }
                Row(horizontalArrangement = Arrangement.spacedBy(10.dp)) {
                    ActionSmallButton(icon = "📞", label = "联系") {
                        appState.showToast("正在拨打：${case.mediatorPhone ?: "12348"}")
                    }
                    ActionSmallButton(icon = "💬", label = "消息") {
                        appState.showToast("消息功能开发中")
                    }
                }
            }
        }
    }
}

@Composable
private fun ActionSmallButton(icon: String, label: String, onClick: () -> Unit) {
    Column(
        horizontalAlignment = Alignment.CenterHorizontally,
        modifier = Modifier
            .clickable(onClick = onClick)
            .padding(horizontal = 8.dp, vertical = 6.dp)
    ) {
        Box(
            modifier = Modifier
                .size(38.dp)
                .background(
                    MaterialTheme.colorScheme.primary.copy(alpha = 0.08f),
                    RoundedCornerShape(10.dp)
                ),
            contentAlignment = Alignment.Center
        ) {
            Text(icon, fontSize = 18.sp)
        }
        Spacer(modifier = Modifier.height(4.dp))
        Text(label, fontSize = 11.sp, color = MaterialTheme.colorScheme.onSurfaceVariant)
    }
}

@Composable
private fun BottomActionsDetail(
    case: Case,
    appState: com.dispute.app.AppState,
    router: com.dispute.app.Router
) {
    Row(
        modifier = Modifier
            .fillMaxWidth()
            .background(Color.White)
            .padding(horizontal = 16.dp, vertical = 12.dp),
        horizontalArrangement = Arrangement.spacedBy(10.dp)
    ) {
        if (case.status == Case.Status.MEDIATING || case.status == Case.Status.ASSIGNED) {
            Button(
                onClick = {
                    appState.showToast("已提交催办申请，调解员将尽快处理")
                },
                modifier = Modifier
                    .weight(1f)
                    .height(46.dp),
                shape = RoundedCornerShape(12.dp),
                colors = ButtonDefaults.buttonColors(
                    containerColor = Color(0xFFFFF7ED),
                    contentColor = Color(0xFFEA580C)
                )
            ) {
                Text("催办", fontWeight = FontWeight.SemiBold)
            }
        }

        if (case.status == Case.Status.MEDIATING || case.status == Case.Status.ASSIGNED ||
            case.status == Case.Status.PENDING_REVIEW) {
            Button(
                onClick = {
                    appState.showToast("补充材料功能开发中")
                },
                modifier = Modifier
                    .weight(1f)
                    .height(46.dp),
                shape = RoundedCornerShape(12.dp),
                colors = ButtonDefaults.buttonColors(
                    containerColor = MaterialTheme.colorScheme.surfaceVariant,
                    contentColor = MaterialTheme.colorScheme.onSurface
                )
            ) {
                Text("补充材料", fontWeight = FontWeight.SemiBold)
            }
        }

        if ((case.status == Case.Status.SUCCESSFUL || case.status == Case.Status.CLOSED)
            && case.satisfactionRating == null) {
            Button(
                onClick = { router.navigate(Route.Satisfaction(case.caseNumber)) },
                modifier = Modifier
                    .weight(1.5f)
                    .height(46.dp),
                shape = RoundedCornerShape(12.dp),
                colors = ButtonDefaults.buttonColors(
                    containerColor = MaterialTheme.colorScheme.primary
                )
            ) {
                Text("评价服务", fontWeight = FontWeight.SemiBold)
            }
        }

        if (case.status == Case.Status.MEDIATING || case.status == Case.Status.ASSIGNED) {
            Button(
                onClick = { appState.showToast("AI咨询入口") },
                modifier = Modifier
                    .weight(1.2f)
                    .height(46.dp),
                shape = RoundedCornerShape(12.dp),
                colors = ButtonDefaults.buttonColors(
                    containerColor = MaterialTheme.colorScheme.primary
                )
            ) {
                Text("咨询律师", fontWeight = FontWeight.SemiBold)
            }
        }
    }
}

private fun maskPhone(phone: String): String {
    if (phone.length >= 11) {
        return phone.substring(0, 3) + "****" + phone.substring(7)
    }
    return phone
}

private fun getRatingText(rating: Int): String {
    return when (rating) {
        1 -> "非常不满意"
        2 -> "不满意"
        3 -> "一般"
        4 -> "满意"
        5 -> "非常满意"
        else -> ""
    }
}

private fun getStatusIcon(status: Case.Status): String {
    return when (status) {
        Case.Status.PENDING_REVIEW, Case.Status.REVIEWING -> "⏳"
        Case.Status.ASSIGNED, Case.Status.MEDIATING -> "🤝"
        Case.Status.SUCCESSFUL -> "✅"
        Case.Status.UNSUCCESSFUL -> "⚠️"
        Case.Status.CLOSED -> "📋"
        else -> "📄"
    }
}
