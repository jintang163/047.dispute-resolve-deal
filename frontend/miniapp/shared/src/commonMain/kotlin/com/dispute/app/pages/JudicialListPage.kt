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
import com.dispute.app.components.EmptyCard
import com.dispute.app.components.FilterTabs
import com.dispute.app.components.TopBarWithBackList
import com.dispute.app.model.JudicialConfirmation
import com.dispute.app.model.MockData
import androidx.compose.runtime.LaunchedEffect

@Composable
fun JudicialListPage() {
    val appState = androidx.compose.runtime.remember { com.dispute.app.AppState() }
    val router = androidx.compose.runtime.remember { com.dispute.app.Router(appState) }

    CompositionLocalProvider(
        LocalAppState provides appState,
        LocalRouter provides router
    ) {
        JudicialListContent()
    }
}

@Composable
private fun JudicialListContent() {
    val appState = LocalAppState.current
    val router = LocalRouter.current
    val allConfirmations by appState.judicialList
    val currentFilter by appState.judicialStatusFilter

    var filter by remember { mutableStateOf(currentFilter) }

    val filters = listOf<Pair<String, JudicialConfirmation.Status?>>(
        "全部" to null,
        "已提交" to JudicialConfirmation.Status.SUBMITTED,
        "审核中" to JudicialConfirmation.Status.REVIEWING,
        "已确认" to JudicialConfirmation.Status.CONFIRMED,
        "已驳回" to JudicialConfirmation.Status.REJECTED,
        "已失效" to JudicialConfirmation.Status.EXPIRED
    )

    val filteredConfirmations = when (filter) {
        null -> allConfirmations
        else -> allConfirmations.filter { it.status == filter }
    }

    LaunchedEffect(Unit) {
        if (allConfirmations.isEmpty()) {
            appState.setJudicialList(MockData.mockJudicialConfirmations)
        }
    }

    Column(
        modifier = Modifier
            .fillMaxSize()
            .background(MaterialTheme.colorScheme.background)
    ) {
        TopBarWithBackList(
            title = "司法确认",
            onBack = { router.back() },
            actions = {
                Text(
                    text = "申请",
                    color = MaterialTheme.colorScheme.primary,
                    fontWeight = FontWeight.SemiBold,
                    modifier = Modifier
                        .padding(end = 16.dp)
                        .clickable { router.navigate(Route.JudicialApply) }
                )
            }
        )

        FilterTabs(
            filters = filters.map { it.first },
            selectedIndex = filters.indexOfFirst { it.second == filter },
            onSelected = { index ->
                filter = filters[index].second
                appState.setJudicialStatusFilter(filter)
            }
        )

        if (filteredConfirmations.isEmpty()) {
            EmptyCard(
                icon = "📋",
                title = "暂无司法确认记录",
                description = "点击右上角申请司法确认"
            )
        } else {
            LazyColumn(
                modifier = Modifier
                    .fillMaxSize()
                    .padding(horizontal = 16.dp),
                verticalArrangement = Arrangement.spacedBy(12.dp)
            ) {
                items(filteredConfirmations) { confirmation ->
                    JudicialConfirmationCard(
                        confirmation = confirmation,
                        onClick = {
                            appState.setSelectedJudicial(confirmation)
                            router.navigate(Route.JudicialDetail(confirmation.id))
                        }
                    )
                }
                item {
                    Spacer(modifier = Modifier.height(24.dp))
                }
            }
        }
    }
}

@Composable
private fun JudicialConfirmationCard(
    confirmation: JudicialConfirmation,
    onClick: () -> Unit
) {
    val statusColor = Color(confirmation.statusColor)
    val isWarning = confirmation.isWarning
    val isExpired = confirmation.isExpired

    androidx.compose.material3.Card(
        modifier = Modifier
            .fillMaxWidth()
            .clickable(onClick = onClick),
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
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.SpaceBetween,
                verticalAlignment = Alignment.CenterVertically
            ) {
                Text(
                    text = confirmation.confirmNo,
                    style = MaterialTheme.typography.titleMedium,
                    fontWeight = FontWeight.SemiBold,
                    color = MaterialTheme.colorScheme.onSurface
                )
                Box(
                    modifier = Modifier
                        .background(
                            statusColor.copy(alpha = 0.15f),
                            RoundedCornerShape(8.dp)
                        )
                        .padding(horizontal = 10.dp, vertical = 4.dp)
                ) {
                    Text(
                        text = confirmation.statusText,
                        color = statusColor,
                        fontSize = 12.sp,
                        fontWeight = FontWeight.SemiBold
                    )
                }
            }

            Spacer(modifier = Modifier.height(12.dp))

            Text(
                text = confirmation.caseTitle,
                style = MaterialTheme.typography.bodyLarge,
                color = MaterialTheme.colorScheme.onSurface,
                maxLines = 2
            )

            Spacer(modifier = Modifier.height(8.dp))

            Row(
                modifier = Modifier.fillMaxWidth(),
                verticalAlignment = Alignment.CenterVertically
            ) {
                Text(
                    text = "申请人：${confirmation.applicantName}",
                    style = MaterialTheme.typography.bodyMedium,
                    color = MaterialTheme.colorScheme.onSurfaceVariant
                )
                Spacer(modifier = Modifier.width(16.dp))
                Text(
                    text = "被申请人：${confirmation.respondentName}",
                    style = MaterialTheme.typography.bodyMedium,
                    color = MaterialTheme.colorScheme.onSurfaceVariant
                )
            }

            Spacer(modifier = Modifier.height(8.dp))

            Row(
                modifier = Modifier.fillMaxWidth(),
                verticalAlignment = Alignment.CenterVertically
            ) {
                Text(
                    text = "法院：${confirmation.courtName}",
                    style = MaterialTheme.typography.bodyMedium,
                    color = MaterialTheme.colorScheme.onSurfaceVariant
                )
            }

            if (confirmation.performanceDeadline != null && confirmation.status == JudicialConfirmation.Status.CONFIRMED) {
                Spacer(modifier = Modifier.height(12.dp))
                Row(
                    modifier = Modifier.fillMaxWidth(),
                    horizontalArrangement = Arrangement.SpaceBetween,
                    verticalAlignment = Alignment.CenterVertically
                ) {
                    Text(
                        text = "履行期限：${confirmation.performanceDeadline}",
                        style = MaterialTheme.typography.bodyMedium,
                        color = MaterialTheme.colorScheme.onSurfaceVariant
                    )
                    if (confirmation.daysLeft != null) {
                        Box(
                            modifier = Modifier
                                .background(
                                    if (isExpired) Color(0xFFEF4444).copy(alpha = 0.15f)
                                    else if (isWarning) Color(0xFFF59E0B).copy(alpha = 0.15f)
                                    else Color(0xFF22C55E).copy(alpha = 0.15f),
                                    RoundedCornerShape(6.dp)
                                )
                                .padding(horizontal = 8.dp, vertical = 2.dp)
                        ) {
                            Text(
                                text = if (confirmation.daysLeft > 0) "剩余${confirmation.daysLeft}天" else "已逾期",
                                color = if (isExpired) Color(0xFFEF4444)
                                else if (isWarning) Color(0xFFF59E0B)
                                else Color(0xFF22C55E),
                                fontSize = 11.sp,
                                fontWeight = FontWeight.SemiBold
                            )
                        }
                    }
                }
            }

            if (confirmation.confirmAmount != null && confirmation.confirmAmount > 0) {
                Spacer(modifier = Modifier.height(8.dp))
                Text(
                    text = "确认金额：¥${String.format("%.2f", confirmation.confirmAmount)}",
                    style = MaterialTheme.typography.bodyMedium,
                    color = MaterialTheme.colorScheme.primary,
                    fontWeight = FontWeight.SemiBold
                )
            }

            Spacer(modifier = Modifier.height(12.dp))
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.SpaceBetween,
                verticalAlignment = Alignment.CenterVertically
            ) {
                Text(
                    text = "申请时间：${confirmation.createTime}",
                    style = MaterialTheme.typography.labelMedium,
                    color = MaterialTheme.colorScheme.onSurfaceVariant
                )
                Row(
                    verticalAlignment = Alignment.CenterVertically
                ) {
                    Text(
                        text = "查看详情",
                        style = MaterialTheme.typography.labelMedium,
                        color = MaterialTheme.colorScheme.primary,
                        fontWeight = FontWeight.Medium
                    )
                    Spacer(modifier = Modifier.width(4.dp))
                    Text(
                        text = "›",
                        style = MaterialTheme.typography.labelLarge,
                        color = MaterialTheme.colorScheme.primary
                    )
                }
            }
        }
    }
}
