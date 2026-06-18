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
import androidx.compose.material3.Button
import androidx.compose.material3.ButtonDefaults
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Switch
import androidx.compose.material3.SwitchDefaults
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
import com.dispute.app.model.User
import kotlinx.coroutines.launch

@Composable
fun ProfilePage() {
    val appState = androidx.compose.runtime.remember { com.dispute.app.AppState() }
    val router = androidx.compose.runtime.remember { com.dispute.app.Router(appState) }

    CompositionLocalProvider(
        LocalAppState provides appState,
        LocalRouter provides router
    ) {
        ProfileContent()
    }
}

@Composable
private fun ProfileContent() {
    val appState = LocalAppState.current
    val router = LocalRouter.current
    val user by appState.currentUser
    val caseList by appState.caseList

    var pushNotification by remember { mutableStateOf(true) }
    var darkMode by remember { mutableStateOf(false) }

    val displayUser = user ?: User(
        id = "guest",
        nickname = "登录用户",
        phone = "138****8000",
        isVerified = true
    )

    Column(
        modifier = Modifier
            .fillMaxSize()
            .background(MaterialTheme.colorScheme.background)
    ) {
        Box(
            modifier = Modifier
                .fillMaxWidth()
                .height(200.dp)
                .background(
                    androidx.compose.ui.graphics.Brush.linearGradient(
                        listOf(Color(0xFF1D6CFF), Color(0xFF4D8CFF))
                    )
                )
        ) {
            Column(
                modifier = Modifier
                    .fillMaxSize()
                    .padding(20.dp),
                horizontalAlignment = Alignment.CenterHorizontally
            ) {
                Spacer(modifier = Modifier.height(20.dp))
                Row(
                    modifier = Modifier.fillMaxWidth(),
                    horizontalArrangement = Arrangement.SpaceBetween,
                    verticalAlignment = Alignment.CenterVertically
                ) {
                    Text(
                        text = "←",
                        fontSize = 24.sp,
                        color = Color.White,
                        modifier = Modifier.clickable { router.back() }
                    )
                    Text(
                        "个人中心",
                        style = MaterialTheme.typography.titleMedium,
                        color = Color.White,
                        fontWeight = FontWeight.SemiBold
                    )
                    Text(
                        "⚙️",
                        fontSize = 22.sp,
                        modifier = Modifier.clickable { appState.showToast("设置功能开发中") }
                    )
                }

                Spacer(modifier = Modifier.height(20.dp))

                Row(
                    modifier = Modifier.fillMaxWidth(),
                    verticalAlignment = Alignment.CenterVertically
                ) {
                    Box(
                        modifier = Modifier
                            .size(64.dp)
                            .background(Color.White.copy(alpha = 0.25f), CircleShape),
                        contentAlignment = Alignment.Center
                    ) {
                        Text(
                            displayUser.displayName.firstOrNull()?.toString() ?: "U",
                            color = Color.White,
                            fontSize = 26.sp,
                            fontWeight = FontWeight.Bold
                        )
                    }
                    Spacer(modifier = Modifier.width(14.dp))
                    Column(modifier = Modifier.weight(1f)) {
                        Row(verticalAlignment = Alignment.CenterVertically) {
                            Text(
                                displayUser.displayName,
                                color = Color.White,
                                style = MaterialTheme.typography.titleLarge,
                                fontWeight = FontWeight.Bold
                            )
                            Spacer(modifier = Modifier.width(8.dp))
                            if (displayUser.isVerified) {
                                Box(
                                    modifier = Modifier
                                        .background(Color(0xFF22C55E), RoundedCornerShape(6.dp))
                                        .padding(horizontal = 6.dp, vertical = 2.dp)
                                ) {
                                    Text("已认证", color = Color.White, fontSize = 10.sp)
                                }
                            }
                        }
                        Spacer(modifier = Modifier.height(4.dp))
                        Text(
                            displayUser.maskedPhone ?: "未绑定手机号",
                            color = Color.White.copy(alpha = 0.9f),
                            style = MaterialTheme.typography.bodyMedium
                        )
                    }
                    Box(
                        modifier = Modifier
                            .background(Color.White.copy(alpha = 0.18f), RoundedCornerShape(12.dp))
                            .clickable { appState.showToast("编辑个人资料") }
                            .padding(horizontal = 12.dp, vertical = 8.dp)
                    ) {
                        Text("编辑", color = Color.White, fontSize = 13.sp)
                    }
                }
            }
        }

        Row(
            modifier = Modifier
                .fillMaxWidth()
                .padding(horizontal = 16.dp)
                .padding(top = (-30).dp),
            horizontalArrangement = Arrangement.spacedBy(10.dp)
        ) {
            StatBoxBig(count = caseList.size, label = "案件总数", onClick = { router.navigate(Route.CaseList) })
            StatBoxBig(
                count = caseList.count { it.status == com.dispute.app.model.Case.Status.SUCCESSFUL },
                label = "已完成",
                onClick = { router.navigate(Route.CaseList) }
            )
            StatBoxBig(
                count = caseList.count { it.satisfactionRating != null },
                label = "已评价",
                onClick = { appState.showToast("评价记录") }
            )
        }

        Column(
            modifier = Modifier
                .fillMaxWidth()
                .weight(1f)
                .padding(16.dp),
            verticalArrangement = Arrangement.spacedBy(14.dp)
        ) {
            AppCard {
                Column(verticalArrangement = Arrangement.spacedBy(2.dp)) {
                    ProfileMenuItem(
                        icon = "📋",
                        title = "我的案件",
                        subtitle = "查看所有纠纷登记记录",
                        onClick = { router.navigate(Route.CaseList) }
                    )
                    ProfileMenuItem(
                        icon = "📝",
                        title = "快速登记",
                        subtitle = "提交新的纠纷调解申请",
                        onClick = { router.navigate(Route.RegisterCase) }
                    )
                    ProfileMenuItem(
                        icon = "🔍",
                        title = "进度查询",
                        subtitle = "查询案件调解进度",
                        onClick = { router.navigate(Route.Progress) }
                    )
                    ProfileMenuItem(
                        icon = "⭐",
                        title = "我的评价",
                        subtitle = "查看已提交的服务评价",
                        onClick = { appState.showToast("评价记录功能") }
                    )
                }
            }

            AppCard(title = "账号与安全") {
                Column(verticalArrangement = Arrangement.spacedBy(2.dp)) {
                    ProfileMenuItem(
                        icon = "🪪",
                        title = "实名认证",
                        subtitle = if (displayUser.isVerified) "已完成实名认证" else "未认证，去认证",
                        showBadge = !displayUser.isVerified,
                        onClick = { appState.showToast("实名认证页面") }
                    )
                    ProfileMenuItem(
                        icon = "📱",
                        title = "绑定手机",
                        subtitle = displayUser.maskedPhone ?: "未绑定",
                        onClick = { appState.showToast("修改绑定手机号") }
                    )
                    ProfileMenuItem(
                        icon = "💬",
                        title = "绑定微信",
                        subtitle = if (displayUser.wechatOpenId != null) "已绑定" else "未绑定",
                        onClick = { appState.showToast("微信绑定") }
                    )
                }
            }

            AppCard(title = "通用设置") {
                Column(verticalArrangement = Arrangement.spacedBy(2.dp)) {
                    ProfileMenuItemSwitch(
                        icon = "🔔",
                        title = "消息通知",
                        subtitle = "接收案件进度和消息推送",
                        checked = pushNotification,
                        onCheckedChange = {
                            pushNotification = it
                            appState.showToast("通知已" + if (it) "开启" else "关闭")
                        }
                    )
                    ProfileMenuItemSwitch(
                        icon = "🌙",
                        title = "深色模式",
                        subtitle = "切换应用显示主题",
                        checked = darkMode,
                        onCheckedChange = {
                            darkMode = it
                            appState.showToast("主题切换需重启应用")
                        }
                    )
                    ProfileMenuItem(
                        icon = "🌐",
                        title = "语言",
                        subtitle = "简体中文",
                        onClick = { appState.showToast("语言设置") }
                    )
                    ProfileMenuItem(
                        icon = "💾",
                        title = "清除缓存",
                        subtitle = "清除本地缓存数据",
                        onClick = { appState.showToast("缓存已清除（模拟）") }
                    )
                }
            }

            AppCard(title = "帮助与关于") {
                Column(verticalArrangement = Arrangement.spacedBy(2.dp)) {
                    ProfileMenuItem(
                        icon = "❓",
                        title = "帮助中心",
                        subtitle = "常见问题和使用指南",
                        onClick = { router.navigate(Route.AIConsult) }
                    )
                    ProfileMenuItem(
                        icon = "📞",
                        title = "联系客服",
                        subtitle = "服务热线：12348",
                        onClick = { appState.showToast("正在拨打12348...") }
                    )
                    ProfileMenuItem(
                        icon = "📜",
                        title = "用户协议",
                        onClick = { appState.showToast("用户协议页面") }
                    )
                    ProfileMenuItem(
                        icon = "🔒",
                        title = "隐私政策",
                        onClick = { appState.showToast("隐私政策页面") }
                    )
                    ProfileMenuItem(
                        icon = "ℹ️",
                        title = "关于我们",
                        subtitle = "版本 v1.0.0",
                        onClick = { appState.showToast("纠纷多元化解服务平台 v1.0.0") }
                    )
                }
            }

            Button(
                onClick = {
                    appState.logout()
                    router.navigateToLogin()
                },
                modifier = Modifier
                    .fillMaxWidth()
                    .height(50.dp),
                shape = RoundedCornerShape(14.dp),
                colors = ButtonDefaults.buttonColors(
                    containerColor = Color(0xFFFEF2F2),
                    contentColor = Color(0xFFDC2626)
                )
            ) {
                Text("退出登录", fontWeight = FontWeight.SemiBold)
            }

            Spacer(modifier = Modifier.height(16.dp))
            Text(
                "© 2024 纠纷多元化解服务中心",
                modifier = Modifier.fillMaxWidth(),
                style = MaterialTheme.typography.labelSmall,
                color = Color(0xFF9CA3AF),
                textAlign = androidx.compose.ui.text.style.TextAlign.Center
            )
        }
    }
}

@Composable
private fun StatBoxBig(count: Int, label: String, onClick: () -> Unit) {
    Box(
        modifier = Modifier
            .weight(1f)
            .background(Color.White, RoundedCornerShape(16.dp))
            .clickable(onClick = onClick)
            .padding(vertical = 18.dp),
        contentAlignment = Alignment.Center
    ) {
        Column(horizontalAlignment = Alignment.CenterHorizontally) {
            Text(
                count.toString(),
                color = MaterialTheme.colorScheme.primary,
                style = MaterialTheme.typography.headlineMedium,
                fontWeight = FontWeight.Bold
            )
            Spacer(modifier = Modifier.height(4.dp))
            Text(
                label,
                style = MaterialTheme.typography.labelMedium,
                color = MaterialTheme.colorScheme.onSurfaceVariant
            )
        }
    }
}

@Composable
private fun ProfileMenuItem(
    icon: String,
    title: String,
    subtitle: String? = null,
    showBadge: Boolean = false,
    onClick: () -> Unit
) {
    Row(
        modifier = Modifier
            .fillMaxWidth()
            .clickable(onClick = onClick)
            .padding(vertical = 14.dp),
        verticalAlignment = Alignment.CenterVertically
    ) {
        Box(
            modifier = Modifier
                .size(36.dp)
                .background(
                    MaterialTheme.colorScheme.primary.copy(alpha = 0.1f),
                    RoundedCornerShape(10.dp)
                ),
            contentAlignment = Alignment.Center
        ) {
            Text(icon, fontSize = 18.sp)
        }
        Spacer(modifier = Modifier.width(12.dp))
        Column(modifier = Modifier.weight(1f)) {
            Row(verticalAlignment = Alignment.CenterVertically) {
                Text(title, style = MaterialTheme.typography.titleSmall, fontWeight = FontWeight.SemiBold)
                if (showBadge) {
                    Spacer(modifier = Modifier.width(6.dp))
                    Box(
                        modifier = Modifier
                            .background(Color(0xFFEF4444), RoundedCornerShape(4.dp))
                            .padding(horizontal = 6.dp, vertical = 1.dp)
                    ) {
                        Text("未完成", color = Color.White, fontSize = 10.sp)
                    }
                }
            }
            if (subtitle != null) {
                Spacer(modifier = Modifier.height(2.dp))
                Text(
                    subtitle,
                    style = MaterialTheme.typography.labelMedium,
                    color = MaterialTheme.colorScheme.onSurfaceVariant
                )
            }
        }
        Text("›", color = Color(0xFFC7C7CC), fontSize = 22.sp)
    }
}

@Composable
private fun ProfileMenuItemSwitch(
    icon: String,
    title: String,
    subtitle: String? = null,
    checked: Boolean,
    onCheckedChange: (Boolean) -> Unit
) {
    Row(
        modifier = Modifier
            .fillMaxWidth()
            .padding(vertical = 14.dp),
        verticalAlignment = Alignment.CenterVertically
    ) {
        Box(
            modifier = Modifier
                .size(36.dp)
                .background(
                    MaterialTheme.colorScheme.primary.copy(alpha = 0.1f),
                    RoundedCornerShape(10.dp)
                ),
            contentAlignment = Alignment.Center
        ) {
            Text(icon, fontSize = 18.sp)
        }
        Spacer(modifier = Modifier.width(12.dp))
        Column(modifier = Modifier.weight(1f)) {
            Text(title, style = MaterialTheme.typography.titleSmall, fontWeight = FontWeight.SemiBold)
            if (subtitle != null) {
                Spacer(modifier = Modifier.height(2.dp))
                Text(
                    subtitle,
                    style = MaterialTheme.typography.labelMedium,
                    color = MaterialTheme.colorScheme.onSurfaceVariant
                )
            }
        }
        Switch(
            checked = checked,
            onCheckedChange = onCheckedChange,
            colors = SwitchDefaults.colors(
                checkedThumbColor = Color.White,
                checkedTrackColor = MaterialTheme.colorScheme.primary
            )
        )
    }
}
