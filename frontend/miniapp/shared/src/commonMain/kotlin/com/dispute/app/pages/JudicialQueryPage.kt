package com.dispute.app.pages

import androidx.compose.foundation.background
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
import com.dispute.app.components.InputField
import com.dispute.app.components.LargeButton
import com.dispute.app.components.TopBarWithBackList
import androidx.compose.runtime.LaunchedEffect

@Composable
fun JudicialQueryPage() = JudicialQueryContent()

@Composable
private fun JudicialQueryContent() {
    val appState = LocalAppState.current
    val router = LocalRouter.current

    var confirmNo by remember { mutableStateOf("") }
    var isQuerying by remember { mutableStateOf(false) }
    var errorMessage by remember { mutableStateOf<String?>(null) }

    fun handleQuery() {
        if (confirmNo.isBlank()) {
            errorMessage = "请输入司法确认编号"
            return
        }

        errorMessage = null
        isQuerying = true

        appState.appScope.launch {
            try {
                val result = com.dispute.app.api.JudicialApi.queryJudicialByNo(confirmNo)
                if (result.success && result.data != null) {
                    val confirmation = result.data!!
                    appState.addJudicialConfirmation(confirmation)
                    appState.setSelectedJudicial(confirmation)
                    router.navigate(Route.JudicialDetail(confirmation.id))
                } else {
                    errorMessage = result.message ?: "未找到该编号的司法确认记录"
                }
            } catch (e: Exception) {
                errorMessage = "网络异常，请检查网络后重试"
            } finally {
                isQuerying = false
            }
        }
    }

    Column(
        modifier = Modifier
            .fillMaxSize()
            .background(MaterialTheme.colorScheme.background)
    ) {
        TopBarWithBackList(
            title = "查询确认书",
            onBack = { router.back() }
        )

        Column(
            modifier = Modifier
                .fillMaxSize()
                .padding(24.dp),
            horizontalAlignment = Alignment.CenterHorizontally
        ) {
            Spacer(modifier = Modifier.height(32.dp))

            Box(
                modifier = Modifier
                    .size(100.dp)
                    .background(
                        MaterialTheme.colorScheme.primary.copy(alpha = 0.1f),
                        RoundedCornerShape(24.dp)
                    ),
                contentAlignment = Alignment.Center
            ) {
                Text("🔍", fontSize = 48.sp)
            }

            Spacer(modifier = Modifier.height(24.dp))

            Text(
                text = "输入司法确认编号",
                style = MaterialTheme.typography.titleLarge,
                fontWeight = FontWeight.Bold,
                color = MaterialTheme.colorScheme.onSurface
            )

            Spacer(modifier = Modifier.height(8.dp))

            Text(
                text = "请输入您的司法确认书编号查询进度",
                style = MaterialTheme.typography.bodyMedium,
                color = MaterialTheme.colorScheme.onSurfaceVariant
            )

            Spacer(modifier = Modifier.height(32.dp))

            InputField(
                label = "确认编号",
                placeholder = "请输入确认编号，如SF20240115000001",
                value = confirmNo,
                onValueChange = { confirmNo = it.uppercase() },
                error = errorMessage,
                onImeAction = { handleQuery() }
            )

            Spacer(modifier = Modifier.height(16.dp))

            Box(
                modifier = Modifier
                    .fillMaxWidth()
                    .background(Color(0xFFEFF6FF), RoundedCornerShape(8.dp))
                    .padding(12.dp)
            ) {
                Column {
                    Text(
                        text = "💡 编号格式说明",
                        style = MaterialTheme.typography.labelMedium,
                        fontWeight = FontWeight.SemiBold,
                        color = Color(0xFF1D4ED8)
                    )
                    Spacer(modifier = Modifier.height(4.dp))
                    Text(
                        text = "司法确认编号通常以SF开头，共16位，例如：SF20240115000001",
                        style = MaterialTheme.typography.bodySmall,
                        color = Color(0xFF3B82F6),
                        lineHeight = 18.sp
                    )
                }
            }

            Spacer(modifier = Modifier.height(32.dp))

            LargeButton(
                text = "查询",
                onClick = { handleQuery() },
                loading = isQuerying,
                enabled = !isQuerying && confirmNo.isNotBlank()
            )

            Spacer(modifier = Modifier.height(16.dp))

            Row(
                verticalAlignment = Alignment.CenterVertically
            ) {
                Text(
                    text = "或",
                    style = MaterialTheme.typography.bodyMedium,
                    color = MaterialTheme.colorScheme.onSurfaceVariant
                )
                Spacer(modifier = Modifier.width(8.dp))
                Text(
                    text = "查看我的申请",
                    style = MaterialTheme.typography.bodyMedium,
                    color = MaterialTheme.colorScheme.primary,
                    fontWeight = FontWeight.SemiBold,
                    modifier = androidx.compose.foundation.clickable {
                        router.navigate(Route.JudicialList)
                    }
                )
            }
        }
    }
}
