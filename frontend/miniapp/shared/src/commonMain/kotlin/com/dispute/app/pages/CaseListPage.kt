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
import androidx.compose.runtime.*
import com.dispute.app.LocalRouter
import com.dispute.app.Route
import com.dispute.app.components.EmptyCard
import com.dispute.app.model.Case
import com.dispute.app.model.MockData

@Composable
fun CaseListPage() = CaseListContent()

@Composable
private fun CaseListContent() {
    val appState = LocalAppState.current
    val router = LocalRouter.current
    val allCases by appState.caseList
    val currentFilter by appState.caseStatusFilter

    var filter by remember { mutableStateOf(currentFilter) }

    val filters = listOf<Pair<String, Case.Status?>>(
        "全部" to null,
        "待处理" to Case.Status.PENDING_REVIEW,
        "调解中" to Case.Status.MEDIATING,
        "已完成" to Case.Status.SUCCESSFUL,
        "已结案" to Case.Status.CLOSED
    )

    val filteredCases = when (filter) {
        null -> allCases
        else -> allCases.filter {
            if (filter == Case.Status.PENDING_REVIEW) {
                it.status == Case.Status.PENDING_REVIEW || it.status == Case.Status.REVIEWING
            } else if (filter == Case.Status.MEDIATING) {
                it.status == Case.Status.MEDIATING || it.status == Case.Status.ASSIGNED
            } else {
                it.status == filter
            }
        }
    }

    androidx.compose.runtime.LaunchedEffect(Unit) {
        if (allCases.isEmpty()) {
            appState.setCaseList(MockData.mockCases)
        }
    }

    Column(
        modifier = Modifier
            .fillMaxSize()
            .background(MaterialTheme.colorScheme.background)
    ) {
        TopBarWithBackList(
            title = "我的案件",
            onBack = { router.back() }
        )

        FilterTabs(
            filters = filters,
            selected = filter,
            onSelect = { newFilter ->
                filter = newFilter
                appState.setCaseStatusFilter(newFilter)
            }
        )

        if (filteredCases.isEmpty()) {
            EmptyCard(
                icon = "📋",
                title = "暂无案件",
                description = "点击首页的快速登记按钮开始登记纠纷"
            )
        } else {
            LazyColumn(
                modifier = Modifier
                    .fillMaxWidth()
                    .weight(1f)
                    .padding(horizontal = 16.dp, vertical = 12.dp),
                verticalArrangement = Arrangement.spacedBy(12.dp)
            ) {
                items(filteredCases, key = { it.caseNumber }) { case ->
                    CaseListItem(
                        case = case,
                        onClick = { router.navigate(Route.CaseDetail(case.caseNumber)) }
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
private fun TopBarWithBackList(title: String, onBack: () -> Unit) {
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
private fun FilterTabs(
    filters: List<Pair<String, Case.Status?>>,
    selected: Case.Status?,
    onSelect: (Case.Status?) -> Unit
) {
    Row(
        modifier = Modifier
            .fillMaxWidth()
            .background(Color.White)
            .padding(horizontal = 16.dp, vertical = 8.dp),
        horizontalArrangement = Arrangement.spacedBy(8.dp)
    ) {
        filters.forEach { (label, status) ->
            val isSelected = (status == selected) || (status == null && selected == null)
            Box(
                modifier = Modifier
                    .background(
                        if (isSelected) MaterialTheme.colorScheme.primary.copy(alpha = 0.1f)
                        else MaterialTheme.colorScheme.surfaceVariant.copy(alpha = 0.4f),
                        RoundedCornerShape(20.dp)
                    )
                    .clickable { onSelect(status) }
                    .padding(horizontal = 14.dp, vertical = 8.dp)
            ) {
                Text(
                    text = label,
                    color = if (isSelected) MaterialTheme.colorScheme.primary else MaterialTheme.colorScheme.onSurfaceVariant,
                    fontWeight = if (isSelected) FontWeight.SemiBold else FontWeight.Normal,
                    fontSize = 13.sp
                )
            }
        }
    }
}

@Composable
private fun CaseListItem(case: Case, onClick: () -> Unit) {
    val statusColor = Color(case.statusColor)

    Column(
        modifier = Modifier
            .fillMaxWidth()
            .background(Color.White, RoundedCornerShape(16.dp))
            .clickable(onClick = onClick)
            .padding(16.dp)
    ) {
        Row(
            modifier = Modifier.fillMaxWidth(),
            horizontalArrangement = Arrangement.SpaceBetween,
            verticalAlignment = Alignment.CenterVertically
        ) {
            Text(
                text = case.caseNumber,
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
                    text = case.status.displayName,
                    color = statusColor,
                    fontSize = 12.sp,
                    fontWeight = FontWeight.SemiBold
                )
            }
        }

        Spacer(modifier = Modifier.height(12.dp))

        Text(
            text = case.fullTypeName,
            style = MaterialTheme.typography.titleMedium,
            fontWeight = FontWeight.Bold
        )

        Spacer(modifier = Modifier.height(8.dp))

        Text(
            text = "对方：${case.opponentName}",
            style = MaterialTheme.typography.bodyMedium,
            color = MaterialTheme.colorScheme.onSurfaceVariant,
            maxLines = 1
        )

        if (case.description.isNotBlank()) {
            Spacer(modifier = Modifier.height(6.dp))
            Text(
                text = case.description,
                style = MaterialTheme.typography.bodySmall,
                color = MaterialTheme.colorScheme.onSurfaceVariant.copy(alpha = 0.7f),
                maxLines = 2
            )
        }

        Spacer(modifier = Modifier.height(12.dp))

        Row(
            modifier = Modifier.fillMaxWidth(),
            horizontalArrangement = Arrangement.SpaceBetween,
            verticalAlignment = Alignment.CenterVertically
        ) {
            Text(
                text = case.submitTime.substringBefore(" "),
                style = MaterialTheme.typography.labelMedium,
                color = Color(0xFF9CA3AF)
            )
            if (case.mediatorName != null) {
                Row(verticalAlignment = Alignment.CenterVertically) {
                    Box(
                        modifier = Modifier
                            .size(20.dp)
                            .background(
                                MaterialTheme.colorScheme.primary.copy(alpha = 0.1f),
                                androidx.compose.foundation.shape.CircleShape
                            ),
                        contentAlignment = Alignment.Center
                    ) {
                        Text("👤", fontSize = 11.sp)
                    }
                    Spacer(modifier = Modifier.width(4.dp))
                    Text(
                        text = case.mediatorName,
                        style = MaterialTheme.typography.labelMedium,
                        color = MaterialTheme.colorScheme.primary,
                        fontWeight = FontWeight.Medium
                    )
                }
            } else if (case.status == Case.Status.PENDING_REVIEW || case.status == Case.Status.REVIEWING) {
                Box(
                    modifier = Modifier
                        .background(Color(0xFFFFF7ED), RoundedCornerShape(6.dp))
                        .padding(horizontal = 8.dp, vertical = 4.dp)
                ) {
                    Text(
                        text = "等待分配调解员",
                        fontSize = 11.sp,
                        color = Color(0xFFEA580C)
                    )
                }
            }
        }

        if (case.satisfactionRating == null && (case.status == Case.Status.SUCCESSFUL || case.status == Case.Status.CLOSED)) {
            Spacer(modifier = Modifier.height(12.dp))
            Box(
                modifier = Modifier
                    .fillMaxWidth()
                    .background(Color(0xFFFEF3C7), RoundedCornerShape(10.dp))
                    .clickable { }
                    .padding(horizontal = 12.dp, vertical = 10.dp)
            ) {
                Row(verticalAlignment = Alignment.CenterVertically) {
                    Text("⭐", fontSize = 16.sp)
                    Spacer(modifier = Modifier.width(8.dp))
                    Text(
                        text = "本案已完成，请对我们的服务进行评价",
                        fontSize = 12.sp,
                        color = Color(0xFF92400E),
                        modifier = Modifier.weight(1f)
                    )
                    Text("去评价 ›", fontSize = 12.sp, color = Color(0xFFB45309), fontWeight = FontWeight.SemiBold)
                }
            }
        }
    }
}
