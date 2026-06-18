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
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.shape.RoundedCornerShape
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
import com.dispute.app.components.EmptyCard
import com.dispute.app.components.InfoCard
import com.dispute.app.components.TopBarWithBackList
import com.dispute.app.model.JudicialConfirmation
import com.dispute.app.model.JudicialConfirmLog
import com.dispute.app.model.MockData
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.foundation.lazy.items

@Composable
fun JudicialDetailPage(id: Long) {
    val appState = androidx.compose.runtime.remember { com.dispute.app.AppState() }
    val router = androidx.compose.runtime.remember { com.dispute.app.Router(appState) }

    CompositionLocalProvider(
        LocalAppState provides appState,
        LocalRouter provides router
    ) {
        JudicialDetailContent(id)
    }
}

@Composable
private fun JudicialDetailContent(id: Long) {
    val appState = LocalAppState.current
    val router = LocalRouter.current
    val selectedJudicial by appState.selectedJudicial

    var confirmation by remember { mutableStateOf<JudicialConfirmation?>(null) }
    var logs by remember { mutableStateOf<List<JudicialConfirmLog>>(emptyList()) }

    LaunchedEffect(id) {
        val existing = appState.findJudicialConfirmation(id)
        if (existing != null) {
            confirmation = existing
        } else {
            confirmation = MockData.mockJudicialConfirmations.firstOrNull { it.id == id }
        }
    }

    if (confirmation == null) {
        Column(
            modifier = Modifier
                .fillMaxSize()
                .background(MaterialTheme.colorScheme.background)
        ) {
            TopBarWithBackList(
                title = "确认详情",
                onBack = { router.back() }
            )
            EmptyCard(
                icon = "❌",
                title = "确认记录不存在",
                description = "请返回列表重新选择"
            )
        }
        return
    }

    val conf = confirmation!!
    val statusColor = Color(conf.statusColor)
    val isWarning = conf.isWarning
    val isExpired = conf.isExpired

    Column(
        modifier = Modifier
            .fillMaxSize()
            .background(MaterialTheme.colorScheme.background)
    ) {
        TopBarWithBackList(
            title = "确认详情",
            onBack = { router.back() }
        )

        LazyColumn(
            modifier = Modifier
                .fillMaxSize()
                .padding(horizontal = 16.dp),
            verticalArrangement = Arrangement.spacedBy(16.dp)
        ) {
            item {
                Spacer(modifier = Modifier.height(8.dp))
                StatusHeader(conf, statusColor, isWarning, isExpired)
            }

            item {
                InfoCard(
                    icon = { Text("📁", fontSize = 20.sp) },
                    title = "关联案件",
                    value = conf.caseTitle
                )
            }

            item {
                Row(
                    modifier = Modifier.fillMaxWidth(),
                    horizontalArrangement = Arrangement.spacedBy(12.dp)
                ) {
                    InfoCard(
                        modifier = Modifier.weight(1f),
                        icon = { Text("👤", fontSize = 20.sp) },
                        title = "申请人",
                        value = conf.applicantName
                    )
                    InfoCard(
                        modifier = Modifier.weight(1f),
                        icon = { Text("👥", fontSize = 20.sp) },
                        title = "被申请人",
                        value = conf.respondentName
                    )
                }
            }

            item {
                InfoCard(
                    icon = { Text("🏛️", fontSize = 20.sp) },
                    title = "管辖法院",
                    value = conf.courtName
                )
            }

            if (conf.courtCaseNo != null) {
                item {
                    InfoCard(
                        icon = { Text("📋", fontSize = 20.sp) },
                        title = "法院案件编号",
                        value = conf.courtCaseNo!!
                    )
                }
            }

            item {
                AgreementContentCard(conf.agreementContent)
            }

            if (conf.performanceDeadline != null) {
                item {
                    PerformanceDeadlineCard(
                        deadline = conf.performanceDeadline!!,
                        daysLeft = conf.daysLeft,
                        isWarning = isWarning,
                        isExpired = isExpired,
                        amount = conf.confirmAmount
                    )
                }
            }

            if (conf.documentUrl != null) {
                item {
                    DocumentCard(
                        documentUrl = conf.documentUrl!!,
                        sealTime = conf.sealTime,
                        documentNo = conf.documentNo
                    )
                }
            }

            if (conf.remark != null) {
                item {
                    InfoCard(
                        icon = { Text("📝", fontSize = 20.sp) },
                        title = "备注",
                        value = conf.remark!!
                    )
                }
            }

            item {
                ProgressTimelineSection(logs)
            }

            item {
                Spacer(modifier = Modifier.height(24.dp))
            }
        }
    }
}

@Composable
private fun StatusHeader(
    conf: JudicialConfirmation,
    statusColor: Color,
    isWarning: Boolean,
    isExpired: Boolean
) {
    androidx.compose.material3.Card(
        modifier = Modifier.fillMaxWidth(),
        shape = RoundedCornerShape(16.dp),
        colors = androidx.compose.material3.CardDefaults.cardColors(
            containerColor = statusColor.copy(alpha = 0.1f)
        ),
        elevation = androidx.compose.material3.CardDefaults.cardElevation(defaultElevation = 0.dp)
    ) {
        Column(
            modifier = Modifier
                .fillMaxWidth()
                .padding(20.dp)
        ) {
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.SpaceBetween,
                verticalAlignment = Alignment.CenterVertically
            ) {
                Text(
                    text = conf.confirmNo,
                    style = MaterialTheme.typography.titleLarge,
                    fontWeight = FontWeight.Bold,
                    color = MaterialTheme.colorScheme.onSurface
                )
                Box(
                    modifier = Modifier
                        .background(
                            statusColor.copy(alpha = 0.2f),
                            RoundedCornerShape(8.dp)
                        )
                        .padding(horizontal = 12.dp, vertical = 6.dp)
                ) {
                    Text(
                        text = conf.statusText,
                        color = statusColor,
                        fontSize = 13.sp,
                        fontWeight = FontWeight.SemiBold
                    )
                }
            }

            Spacer(modifier = Modifier.height(12.dp))

            Text(
                text = conf.caseTitle,
                style = MaterialTheme.typography.bodyLarge,
                color = MaterialTheme.colorScheme.onSurface,
                maxLines = 2
            )

            if (conf.daysLeft != null && conf.status == JudicialConfirmation.Status.CONFIRMED) {
                Spacer(modifier = Modifier.height(12.dp))
                Row(
                    verticalAlignment = Alignment.CenterVertically
                ) {
                    Text(
                        text = if (conf.daysLeft > 0) "⏰ 履行期限剩余${conf.daysLeft}天" else "⚠️ 已超过履行期限",
                        style = MaterialTheme.typography.bodyMedium,
                        color = if (isExpired) Color(0xFFEF4444)
                        else if (isWarning) Color(0xFFF59E0B)
                        else Color(0xFF22C55E),
                        fontWeight = FontWeight.SemiBold
                    )
                }
            }
        }
    }
}

@Composable
private fun AgreementContentCard(content: String) {
    androidx.compose.material3.Card(
        modifier = Modifier.fillMaxWidth(),
        shape = RoundedCornerShape(16.dp),
        colors = androidx.compose.material3.CardDefaults.cardColors(
            containerColor = MaterialTheme.colorScheme.surface
        ),
        elevation = androidx.compose.material3.CardDefaults.cardElevation(defaultElevation = 2.dp)
    ) {
        Column(
            modifier = Modifier
                .fillMaxWidth()
                .padding(16.dp)
        ) {
            Text(
                text = "📄 协议内容",
                style = MaterialTheme.typography.titleMedium,
                fontWeight = FontWeight.SemiBold,
                color = MaterialTheme.colorScheme.onSurface
            )
            Spacer(modifier = Modifier.height(12.dp))
            Text(
                text = content,
                style = MaterialTheme.typography.bodyMedium,
                color = MaterialTheme.colorScheme.onSurfaceVariant,
                lineHeight = 24.sp
            )
        }
    }
}

@Composable
private fun PerformanceDeadlineCard(
    deadline: String,
    daysLeft: Int?,
    isWarning: Boolean,
    isExpired: Boolean,
    amount: Double?
) {
    val amountColor = MaterialTheme.colorScheme.primary
    val deadlineColor = if (isExpired) Color(0xFFEF4444)
    else if (isWarning) Color(0xFFF59E0B)
    else Color(0xFF22C55E)

    androidx.compose.material3.Card(
        modifier = Modifier.fillMaxWidth(),
        shape = RoundedCornerShape(16.dp),
        colors = androidx.compose.material3.CardDefaults.cardColors(
            containerColor = MaterialTheme.colorScheme.surface
        ),
        elevation = androidx.compose.material3.CardDefaults.cardElevation(defaultElevation = 2.dp)
    ) {
        Column(
            modifier = Modifier
                .fillMaxWidth()
                .padding(16.dp)
        ) {
            Text(
                text = "⏳ 履行信息",
                style = MaterialTheme.typography.titleMedium,
                fontWeight = FontWeight.SemiBold,
                color = MaterialTheme.colorScheme.onSurface
            )
            Spacer(modifier = Modifier.height(16.dp))
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.SpaceBetween
            ) {
                Column {
                    Text(
                        text = "履行期限",
                        style = MaterialTheme.typography.bodyMedium,
                        color = MaterialTheme.colorScheme.onSurfaceVariant
                    )
                    Spacer(modifier = Modifier.height(4.dp))
                    Text(
                        text = deadline,
                        style = MaterialTheme.typography.titleMedium,
                        fontWeight = FontWeight.SemiBold,
                        color = deadlineColor
                    )
                    if (daysLeft != null) {
                        Spacer(modifier = Modifier.height(4.dp))
                        Text(
                            text = if (daysLeft > 0) "剩余${daysLeft}天" else "已逾期${-daysLeft}天",
                            style = MaterialTheme.typography.labelMedium,
                            color = deadlineColor
                        )
                    }
                }
                if (amount != null && amount > 0) {
                    Column(horizontalAlignment = Alignment.End) {
                        Text(
                            text = "确认金额",
                            style = MaterialTheme.typography.bodyMedium,
                            color = MaterialTheme.colorScheme.onSurfaceVariant
                        )
                        Spacer(modifier = Modifier.height(4.dp))
                        Text(
                            text = "¥${String.format("%.2f", amount)}",
                            style = MaterialTheme.typography.titleMedium,
                            fontWeight = FontWeight.Bold,
                            color = amountColor
                        )
                    }
                }
            }
        }
    }
}

@Composable
private fun DocumentCard(
    documentUrl: String,
    sealTime: String?,
    documentNo: String?
) {
    androidx.compose.material3.Card(
        modifier = Modifier
            .fillMaxWidth()
            .clickable { },
        shape = RoundedCornerShape(16.dp),
        colors = androidx.compose.material3.CardDefaults.cardColors(
            containerColor = Color(0xFFECFDF5)
        ),
        elevation = androidx.compose.material3.CardDefaults.cardElevation(defaultElevation = 2.dp)
    ) {
        Column(
            modifier = Modifier
                .fillMaxWidth()
                .padding(16.dp)
        ) {
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.SpaceBetween,
                verticalAlignment = Alignment.CenterVertically
            ) {
                Row(verticalAlignment = Alignment.CenterVertically) {
                    Text("📄", fontSize = 24.sp)
                    Spacer(modifier = Modifier.width(12.dp))
                    Column {
                        Text(
                            text = "司法确认书",
                            style = MaterialTheme.typography.titleMedium,
                            fontWeight = FontWeight.SemiBold,
                            color = Color(0xFF065F46)
                        )
                        if (documentNo != null) {
                            Spacer(modifier = Modifier.height(2.dp))
                            Text(
                                text = "文书编号：$documentNo",
                                style = MaterialTheme.typography.bodySmall,
                                color = Color(0xFF059669)
                            )
                        }
                        if (sealTime != null) {
                            Spacer(modifier = Modifier.height(2.dp))
                            Text(
                                text = "签章时间：$sealTime",
                                style = MaterialTheme.typography.bodySmall,
                                color = Color(0xFF059669)
                            )
                        }
                    }
                }
                Box(
                    modifier = Modifier
                        .background(Color(0xFF10B981), RoundedCornerShape(8.dp))
                        .padding(horizontal = 12.dp, vertical = 6.dp)
                ) {
                    Text(
                        text = "查看",
                        color = Color.White,
                        fontSize = 12.sp,
                        fontWeight = FontWeight.SemiBold
                    )
                }
            }
        }
    }
}

@Composable
private fun ProgressTimelineSection(logs: List<JudicialConfirmLog>) {
    Column(
        modifier = Modifier.fillMaxWidth()
    ) {
        Text(
            text = "📊 进度轨迹",
            style = MaterialTheme.typography.titleMedium,
            fontWeight = FontWeight.SemiBold,
            color = MaterialTheme.colorScheme.onSurface,
            modifier = Modifier.padding(bottom = 12.dp)
        )

        if (logs.isEmpty()) {
            EmptyCard(
                icon = "📋",
                title = "暂无进度记录",
                description = "进度更新后将在这里显示"
            )
        } else {
            Column(
                modifier = Modifier.fillMaxWidth(),
                verticalArrangement = Arrangement.spacedBy(0.dp)
            ) {
                logs.forEachIndexed { index, log ->
                    TimelineItem(
                        log = log,
                        isLast = index == logs.lastIndex,
                        isActive = index == 0
                    )
                }
            }
        }
    }
}

@Composable
private fun TimelineItem(
    log: JudicialConfirmLog,
    isLast: Boolean,
    isActive: Boolean
) {
    val indicatorColor = if (isActive) {
        MaterialTheme.colorScheme.primary
    } else {
        Color(0xFF22C55E)
    }

    val lineColor = Color(0xFFE5E7EB)
    val textColor = MaterialTheme.colorScheme.onSurface
    val subTextColor = MaterialTheme.colorScheme.onSurfaceVariant

    Row(
        modifier = Modifier.fillMaxWidth(),
        verticalAlignment = Alignment.Top
    ) {
        Column(
            horizontalAlignment = Alignment.CenterHorizontally,
            modifier = Modifier
                .width(48.dp)
                .padding(top = 4.dp)
        ) {
            Box(
                modifier = Modifier
                    .size(if (isActive) 18.dp else 14.dp)
                    .background(indicatorColor, RoundedCornerShape(50)),
                contentAlignment = Alignment.Center
            ) {
                if (!isActive) {
                    Box(
                        modifier = Modifier
                            .size(5.dp)
                            .background(Color.White, RoundedCornerShape(50))
                    )
                }
            }

            if (!isLast) {
                Box(
                    modifier = Modifier
                        .width(2.dp)
                        .height(48.dp)
                        .background(lineColor)
                        .padding(top = 8.dp)
                )
            }
        }

        Column(
            modifier = Modifier
                .fillMaxWidth()
                .padding(start = 8.dp, bottom = if (isLast) 0.dp else 16.dp)
        ) {
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.SpaceBetween,
                verticalAlignment = Alignment.CenterVertically
            ) {
                Text(
                    text = log.actionTypeName,
                    style = MaterialTheme.typography.titleMedium,
                    fontWeight = if (isActive) FontWeight.Bold else FontWeight.SemiBold,
                    color = textColor
                )
            }

            if (log.remark != null) {
                Spacer(modifier = Modifier.height(6.dp))
                Text(
                    text = log.remark!!,
                    style = MaterialTheme.typography.bodyMedium,
                    color = subTextColor,
                    lineHeight = 20.sp
                )
            }

            Spacer(modifier = Modifier.height(8.dp))
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.SpaceBetween,
                verticalAlignment = Alignment.CenterVertically
            ) {
                Text(
                    text = buildString {
                        append(log.operatorTypeName)
                        if (log.operatorName != null) {
                            append(" · ")
                            append(log.operatorName)
                        }
                    },
                    style = MaterialTheme.typography.labelMedium,
                    color = subTextColor
                )
                Text(
                    text = log.createTime,
                    style = MaterialTheme.typography.labelMedium,
                    color = subTextColor
                )
            }
        }
    }
}
