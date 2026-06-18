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
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.shape.CircleShape
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
import com.dispute.app.components.InfoCard
import com.dispute.app.model.Case
import com.dispute.app.model.MockData
import kotlinx.coroutines.launch

@Composable
fun HomePage() {
    val appState = androidx.compose.runtime.remember { com.dispute.app.AppState() }
    val router = androidx.compose.runtime.remember { com.dispute.app.Router(appState) }
    val apiClient = androidx.compose.runtime.remember { com.dispute.app.api.ApiClient() }

    CompositionLocalProvider(
        LocalAppState provides appState,
        LocalRouter provides router,
        LocalApiClient provides apiClient
    ) {
        HomeContent()
    }
}

@Composable
private fun HomeContent() {
    val appState = LocalAppState.current
    val router = LocalRouter.current
    val user by appState.currentUser
    val caseList by appState.caseList

    androidx.compose.runtime.LaunchedEffect(Unit) {
        if (caseList.isEmpty()) {
            appState.setCaseList(MockData.mockCases)
        }
    }

    val pendingCount = caseList.count { it.status == Case.Status.PENDING_REVIEW || it.status == Case.Status.REVIEWING }
    val mediatingCount = caseList.count { it.status == Case.Status.MEDIATING || it.status == Case.Status.ASSIGNED }
    val doneCount = caseList.count { it.status == Case.Status.SUCCESSFUL || it.status == Case.Status.CLOSED }

    LazyColumn(
        modifier = Modifier
            .fillMaxSize()
            .background(MaterialTheme.colorScheme.background),
        verticalArrangement = Arrangement.spacedBy(16.dp)
    ) {
        item {
            HeaderSection(
                userName = user?.displayName ?: "用户",
                onProfileClick = { router.navigate(Route.Profile) }
            )
        }

        item {
            QuickActionsGrid(
                onRegister = { router.navigate(Route.RegisterCase) },
                onMyCases = { router.navigate(Route.CaseList) },
                onProgress = { router.navigate(Route.Progress) },
                onAIConsult = { router.navigate(Route.AIConsult) }
            )
        }

        item {
            StatsRow(
                pendingCount = pendingCount,
                mediatingCount = mediatingCount,
                doneCount = doneCount,
                onPendingClick = { appState.setCaseStatusFilter(Case.Status.PENDING_REVIEW); router.navigate(Route.CaseList) },
                onMediatingClick = { appState.setCaseStatusFilter(Case.Status.MEDIATING); router.navigate(Route.CaseList) },
                onDoneClick = { appState.setCaseStatusFilter(Case.Status.SUCCESSFUL); router.navigate(Route.CaseList) }
            )
        }

        item {
            RecentCasesHeader(onSeeAll = { router.navigate(Route.CaseList) })
        }

        items(caseList.take(3), key = { it.caseNumber }) { case ->
            CaseMiniCard(
                case = case,
                onClick = { router.navigate(Route.CaseDetail(case.caseNumber)) }
            )
        }

        if (caseList.isEmpty()) {
            item {
                com.dispute.app.components.EmptyCard(
                    icon = "📋",
                    title = "暂无案件",
                    description = "点击上方快速登记按钮，开始您的第一次纠纷登记"
                )
            }
        }

        item {
            ServiceHotlineCard()
        }

        item {
            Spacer(modifier = Modifier.height(32.dp))
        }
    }
}

@Composable
private fun HeaderSection(
    userName: String,
    onProfileClick: () -> Unit
) {
    Box(
        modifier = Modifier
            .fillMaxWidth()
            .height(160.dp)
            .background(
                androidx.compose.ui.graphics.Brush.linearGradient(
                    colors = listOf(
                        Color(0xFF1D6CFF),
                        Color(0xFF4D8CFF)
                    )
                )
            )
    ) {
        Column(
            modifier = Modifier
                .fillMaxSize()
                .padding(horizontal = 20.dp, vertical = 24.dp),
            verticalArrangement = Arrangement.Center
        ) {
            Row(
                modifier = Modifier.fillMaxWidth(),
                verticalAlignment = Alignment.CenterVertically
            ) {
                Box(
                    modifier = Modifier
                        .size(56.dp)
                        .background(Color.White.copy(alpha = 0.25f), CircleShape),
                    contentAlignment = Alignment.Center
                ) {
                    Text(
                        text = userName.firstOrNull()?.toString() ?: "U",
                        color = Color.White,
                        fontSize = 24.sp,
                        fontWeight = FontWeight.Bold
                    )
                }
                Spacer(modifier = Modifier.width(14.dp))
                Column(modifier = Modifier.weight(1f)) {
                    Text(
                        text = "您好，$userName",
                        color = Color.White,
                        fontSize = 20.sp,
                        fontWeight = FontWeight.SemiBold
                    )
                    Text(
                        text = "欢迎使用纠纷调解服务平台",
                        color = Color.White.copy(alpha = 0.85f),
                        fontSize = 13.sp,
                        modifier = Modifier.padding(top = 4.dp)
                    )
                }
                Text(
                    text = "⚙️",
                    fontSize = 26.sp,
                    modifier = Modifier.clickable(onClick = onProfileClick)
                )
            }
        }
    }
}

@Composable
private fun QuickActionsGrid(
    onRegister: () -> Unit,
    onMyCases: () -> Unit,
    onProgress: () -> Unit,
    onAIConsult: () -> Unit
) {
    val actions = listOf(
        Triple("📝", "快速登记", onRegister),
        Triple("📋", "我的案件", onMyCases),
        Triple("🔍", "进度查询", onProgress),
        Triple("🤖", "AI咨询", onAIConsult)
    )

    androidx.compose.foundation.layout.Column(
        modifier = Modifier
            .fillMaxWidth()
            .padding(horizontal = 16.dp)
    ) {
        Row(
            modifier = Modifier
                .fillMaxWidth()
                .padding(top = (-40).dp)
                .background(Color.White, RoundedCornerShape(20.dp))
                .padding(vertical = 24.dp, horizontal = 8.dp),
            horizontalArrangement = Arrangement.SpaceEvenly
        ) {
            actions.forEach { (icon, label, onClick) ->
                ActionItem(
                    icon = icon,
                    label = label,
                    onClick = onClick
                )
            }
        }
    }
}

@Composable
private fun ActionItem(icon: String, label: String, onClick: () -> Unit) {
    Column(
        horizontalAlignment = Alignment.CenterHorizontally,
        modifier = Modifier
            .clickable(onClick = onClick)
            .padding(horizontal = 12.dp, vertical = 8.dp)
    ) {
        Box(
            modifier = Modifier
                .size(56.dp)
                .background(
                    MaterialTheme.colorScheme.primary.copy(alpha = 0.1f),
                    RoundedCornerShape(16.dp)
                ),
            contentAlignment = Alignment.Center
        ) {
            Text(
                text = icon,
                fontSize = 28.sp
            )
        }
        Text(
            text = label,
            modifier = Modifier.padding(top = 10.dp),
            style = MaterialTheme.typography.labelLarge,
            fontWeight = FontWeight.Medium
        )
    }
}

@Composable
private fun StatsRow(
    pendingCount: Int,
    mediatingCount: Int,
    doneCount: Int,
    onPendingClick: () -> Unit,
    onMediatingClick: () -> Unit,
    onDoneClick: () -> Unit
) {
    Row(
        modifier = Modifier
            .fillMaxWidth()
            .padding(horizontal = 16.dp),
        horizontalArrangement = Arrangement.spacedBy(12.dp)
    ) {
        StatCard(
            count = pendingCount,
            label = "待审核",
            color = Color(0xFFF59E0B),
            onClick = onPendingClick,
            modifier = Modifier.weight(1f)
        )
        StatCard(
            count = mediatingCount,
            label = "调解中",
            color = Color(0xFF1D6CFF),
            onClick = onMediatingClick,
            modifier = Modifier.weight(1f)
        )
        StatCard(
            count = doneCount,
            label = "已完成",
            color = Color(0xFF22C55E),
            onClick = onDoneClick,
            modifier = Modifier.weight(1f)
        )
    }
}

@Composable
private fun StatCard(
    count: Int,
    label: String,
    color: Color,
    onClick: () -> Unit,
    modifier: Modifier = Modifier
) {
    Column(
        modifier = modifier
            .background(Color.White, RoundedCornerShape(16.dp))
            .clickable(onClick = onClick)
            .padding(vertical = 20.dp),
        horizontalAlignment = Alignment.CenterHorizontally
    ) {
        Text(
            text = count.toString(),
            color = color,
            fontSize = 32.sp,
            fontWeight = FontWeight.Bold
        )
        Text(
            text = label,
            modifier = Modifier.padding(top = 6.dp),
            style = MaterialTheme.typography.bodyMedium,
            color = MaterialTheme.colorScheme.onSurfaceVariant
        )
    }
}

@Composable
private fun RecentCasesHeader(onSeeAll: () -> Unit) {
    Row(
        modifier = Modifier
            .fillMaxWidth()
            .padding(horizontal = 16.dp, vertical = 4.dp),
        horizontalArrangement = Arrangement.SpaceBetween,
        verticalAlignment = Alignment.CenterVertically
    ) {
        Text(
            text = "最近案件",
            style = MaterialTheme.typography.titleMedium,
            fontWeight = FontWeight.Bold
        )
        Text(
            text = "查看全部 >",
            style = MaterialTheme.typography.labelLarge,
            color = MaterialTheme.colorScheme.primary,
            modifier = Modifier.clickable(onClick = onSeeAll)
        )
    }
}

@Composable
private fun CaseMiniCard(case: Case, onClick: () -> Unit) {
    val statusColor = androidx.compose.ui.graphics.Color(case.statusColor)

    AppCard(
        modifier = Modifier.padding(horizontal = 16.dp),
        onClick = onClick
    ) {
        Column {
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.SpaceBetween,
                verticalAlignment = Alignment.CenterVertically
            ) {
                Text(
                    text = case.caseNumber,
                    style = MaterialTheme.typography.titleSmall,
                    fontWeight = FontWeight.SemiBold
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
                fontWeight = FontWeight.SemiBold,
                maxLines = 1
            )

            Spacer(modifier = Modifier.height(8.dp))

            Text(
                text = "对方：${case.opponentName}",
                style = MaterialTheme.typography.bodyMedium,
                color = MaterialTheme.colorScheme.onSurfaceVariant
            )

            Spacer(modifier = Modifier.height(12.dp))

            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.SpaceBetween,
                verticalAlignment = Alignment.CenterVertically
            ) {
                Text(
                    text = case.submitTime.substringBefore(" "),
                    style = MaterialTheme.typography.labelMedium,
                    color = MaterialTheme.colorScheme.onSurfaceVariant
                )
                if (case.mediatorName != null) {
                    Text(
                        text = "调解员：${case.mediatorName}",
                        style = MaterialTheme.typography.labelMedium,
                        color = MaterialTheme.colorScheme.primary
                    )
                }
            }
        }
    }
}

@Composable
private fun ServiceHotlineCard() {
    AppCard(
        modifier = Modifier
            .fillMaxWidth()
            .padding(horizontal = 16.dp),
        backgroundColor = Color(0xFFFFF8E7)
    ) {
        Row(
            modifier = Modifier.fillMaxWidth(),
            verticalAlignment = Alignment.CenterVertically,
            horizontalArrangement = Arrangement.SpaceBetween
        ) {
            Row(verticalAlignment = Alignment.CenterVertically) {
                Box(
                    modifier = Modifier
                        .size(48.dp)
                        .background(Color(0xFFFF9500).copy(alpha = 0.15f), RoundedCornerShape(12.dp)),
                    contentAlignment = Alignment.Center
                ) {
                    Text("📞", fontSize = 24.sp)
                }
                Spacer(modifier = Modifier.width(14.dp))
                Column {
                    Text(
                        text = "调解服务热线",
                        style = MaterialTheme.typography.bodyMedium,
                        color = MaterialTheme.colorScheme.onSurfaceVariant
                    )
                    Text(
                        text = "12348",
                        style = MaterialTheme.typography.titleLarge,
                        color = Color(0xFFFF9500),
                        fontWeight = FontWeight.Bold
                    )
                }
            }
            Box(
                modifier = Modifier
                    .background(Color(0xFFFF9500), RoundedCornerShape(24.dp))
                    .clickable { }
                    .padding(horizontal = 18.dp, vertical = 10.dp)
            ) {
                Text(
                    text = "立即拨打",
                    color = Color.White,
                    fontWeight = FontWeight.SemiBold
                )
            }
        }
    }
}
