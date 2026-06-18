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
import androidx.compose.foundation.lazy.rememberLazyListState
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.text.KeyboardActions
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
import androidx.compose.runtime.mutableStateListOf
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.rememberCoroutineScope
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.input.ImeAction
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.dispute.app.LocalAppState
import com.dispute.app.LocalRouter
import com.dispute.app.Route
import com.dispute.app.model.AIMessage
import kotlinx.coroutines.launch
import kotlinx.datetime.Clock
import kotlinx.datetime.TimeZone
import kotlinx.datetime.toLocalDateTime

@Composable
fun AIConsultPage() {
    val appState = androidx.compose.runtime.remember { com.dispute.app.AppState() }
    val router = androidx.compose.runtime.remember { com.dispute.app.Router(appState) }

    CompositionLocalProvider(
        LocalAppState provides appState,
        LocalRouter provides router
    ) {
        AIConsultContent()
    }
}

@Composable
private fun AIConsultContent() {
    val appState = LocalAppState.current
    val router = LocalRouter.current
    val scope = rememberCoroutineScope()
    val listState = rememberLazyListState()

    val messages = remember { mutableStateListOf<AIMessage>() }
    var inputText by remember { mutableStateOf("") }
    var isTyping by remember { mutableStateOf(false) }

    val quickQuestions = listOf(
        "公司拖欠工资怎么办？",
        "邻里噪音纠纷如何处理？",
        "合同到期不续签有赔偿吗？",
        "交通事故责任怎么认定？",
        "借贷没有借条能起诉吗？"
    )

    androidx.compose.runtime.LaunchedEffect(messages.size) {
        if (messages.isNotEmpty()) {
            listState.animateScrollToItem(messages.size)
        }
    }

    Column(
        modifier = Modifier
            .fillMaxSize()
            .background(MaterialTheme.colorScheme.background)
    ) {
        TopBarAI(
            title = "AI法律咨询",
            onBack = { router.back() },
            onClear = {
                messages.clear()
                appState.setAIConversationId(null)
                appState.showToast("对话已清空")
            }
        )

        if (messages.isEmpty()) {
            Column(
                modifier = Modifier
                    .fillMaxWidth()
                    .weight(1f)
                    .padding(16.dp),
                horizontalAlignment = Alignment.CenterHorizontally
            ) {
                Spacer(modifier = Modifier.height(30.dp))
                Box(
                    modifier = Modifier
                        .size(80.dp)
                        .background(
                            androidx.compose.ui.graphics.Brush.linearGradient(
                                listOf(Color(0xFF1D6CFF), Color(0xFF6366F1))
                            ),
                            RoundedCornerShape(24.dp)
                        ),
                    contentAlignment = Alignment.Center
                ) {
                    Text("🤖", fontSize = 44.sp)
                }
                Spacer(modifier = Modifier.height(16.dp))
                Text(
                    "AI法律助手",
                    style = MaterialTheme.typography.headlineMedium,
                    fontWeight = FontWeight.Bold
                )
                Spacer(modifier = Modifier.height(6.dp))
                Text(
                    "7x24小时在线解答法律问题",
                    style = MaterialTheme.typography.bodyMedium,
                    color = MaterialTheme.colorScheme.onSurfaceVariant
                )
                Spacer(modifier = Modifier.height(30.dp))

                Column(modifier = Modifier.fillMaxWidth()) {
                    Text(
                        "常见问题",
                        style = MaterialTheme.typography.titleMedium,
                        fontWeight = FontWeight.SemiBold,
                        modifier = Modifier.padding(bottom = 10.dp)
                    )
                    Column(verticalArrangement = Arrangement.spacedBy(10.dp)) {
                        quickQuestions.forEach { q ->
                            Row(
                                modifier = Modifier
                                    .fillMaxWidth()
                                    .background(Color.White, RoundedCornerShape(14.dp))
                                    .clickable { sendMessage(q, messages, appState, scope, onTyping = { isTyping = it }) }
                                    .padding(14.dp),
                                verticalAlignment = Alignment.CenterVertically
                            ) {
                                Text("💡", fontSize = 18.sp)
                                Spacer(modifier = Modifier.width(10.dp))
                                Text(
                                    q,
                                    style = MaterialTheme.typography.bodyMedium,
                                    modifier = Modifier.weight(1f)
                                )
                                Text("›", color = Color(0xFF9CA3AF), fontSize = 20.sp)
                            }
                        }
                    }
                }
            }
        } else {
            LazyColumn(
                state = listState,
                modifier = Modifier
                    .fillMaxWidth()
                    .weight(1f)
                    .padding(12.dp),
                verticalArrangement = Arrangement.spacedBy(12.dp)
            ) {
                items(messages, key = { it.id }) { msg ->
                    MessageBubble(message = msg)
                }
                if (isTyping) {
                    item {
                        TypingIndicator()
                    }
                }
                item { Spacer(modifier = Modifier.height(8.dp)) }
            }
        }

        Column(
            modifier = Modifier
                .fillMaxWidth()
                .background(Color.White)
                .padding(12.dp)
        ) {
            if (messages.isNotEmpty() && !isTyping && inputText.isBlank()) {
                val suggestions = listOf("请解释一下", "有什么建议？", "还需要注意什么？")
                Row(
                    horizontalArrangement = Arrangement.spacedBy(8.dp),
                    modifier = Modifier.padding(bottom = 10.dp)
                ) {
                    suggestions.take(3).forEach { suggestion ->
                        Box(
                            modifier = Modifier
                                .background(
                                    MaterialTheme.colorScheme.primary.copy(alpha = 0.08f),
                                    RoundedCornerShape(16.dp)
                                )
                                .clickable { inputText = suggestion }
                                .padding(horizontal = 12.dp, vertical = 8.dp)
                        ) {
                            Text(
                                suggestion,
                                color = MaterialTheme.colorScheme.primary,
                                fontSize = 12.sp,
                                fontWeight = FontWeight.Medium
                            )
                        }
                    }
                }
            }

            Row(
                verticalAlignment = Alignment.CenterVertically,
                horizontalArrangement = Arrangement.spacedBy(8.dp)
            ) {
                OutlinedTextField(
                    value = inputText,
                    onValueChange = { inputText = it },
                    placeholder = {
                        Text(
                            "请输入您的法律问题...",
                            color = MaterialTheme.colorScheme.onSurfaceVariant,
                            fontSize = 13.sp
                        )
                    },
                    modifier = Modifier
                        .weight(1f)
                        .height(48.dp),
                    shape = RoundedCornerShape(24.dp),
                    colors = OutlinedTextFieldDefaults.colors(
                        focusedBorderColor = MaterialTheme.colorScheme.primary,
                        unfocusedBorderColor = Color(0xFFE5E7EB)
                    ),
                    keyboardOptions = KeyboardOptions(imeAction = ImeAction.Send),
                    keyboardActions = KeyboardActions(
                        onSend = {
                            if (inputText.isNotBlank()) {
                                sendMessage(
                                    inputText, messages, appState, scope,
                                    onTyping = { isTyping = it }
                                )
                                inputText = ""
                            }
                        }
                    )
                )
                Button(
                    onClick = {
                        if (inputText.isNotBlank()) {
                            sendMessage(
                                inputText, messages, appState, scope,
                                onTyping = { isTyping = it }
                            )
                            inputText = ""
                        }
                    },
                    modifier = Modifier
                        .size(48.dp),
                    shape = RoundedCornerShape(24.dp),
                    enabled = inputText.isNotBlank(),
                    colors = ButtonDefaults.buttonColors(
                        containerColor = MaterialTheme.colorScheme.primary,
                        disabledContainerColor = Color(0xFFE5E7EB)
                    ),
                    contentPadding = androidx.compose.foundation.layout.PaddingValues(0.dp)
                ) {
                    Text("→", fontSize = 22.sp, fontWeight = FontWeight.Bold)
                }
            }
            Text(
                "AI回复仅供参考，具体情况请咨询专业律师",
                modifier = Modifier
                    .fillMaxWidth()
                    .padding(top = 8.dp),
                fontSize = 11.sp,
                color = Color(0xFF9CA3AF),
                textAlign = androidx.compose.ui.text.style.TextAlign.Center
            )
        }
    }
}

@Composable
private fun TopBarAI(title: String, onBack: () -> Unit, onClear: () -> Unit) {
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
            fontWeight = FontWeight.SemiBold,
            modifier = Modifier.weight(1f)
        )
        Text(
            text = "🗑️",
            fontSize = 20.sp,
            modifier = Modifier.clickable(onClick = onClear)
        )
    }
}

@Composable
private fun MessageBubble(message: AIMessage) {
    val isUser = message.role == AIMessage.Role.USER
    Row(
        modifier = Modifier.fillMaxWidth(),
        horizontalArrangement = if (isUser) Arrangement.End else Arrangement.Start
    ) {
        if (!isUser) {
            Box(
                modifier = Modifier
                    .size(36.dp)
                    .padding(end = 8.dp)
                    .background(
                        androidx.compose.ui.graphics.Brush.linearGradient(
                            listOf(Color(0xFF1D6CFF), Color(0xFF6366F1))
                        ),
                        RoundedCornerShape(10.dp)
                    ),
                contentAlignment = Alignment.Center
            ) {
                Text("🤖", fontSize = 18.sp)
            }
        }
        Box(
            modifier = Modifier
                .then(
                    if (isUser) Modifier.padding(start = 48.dp)
                    else Modifier.padding(end = 48.dp)
                )
                .background(
                    if (isUser) MaterialTheme.colorScheme.primary
                    else Color.White,
                    RoundedCornerShape(16.dp)
                )
                .padding(horizontal = 14.dp, vertical = 12.dp)
        ) {
            Column {
                Text(
                    message.content,
                    style = MaterialTheme.typography.bodyMedium,
                    color = if (isUser) Color.White else MaterialTheme.colorScheme.onSurface,
                    lineHeight = 20.sp
                )
                Text(
                    message.timestamp.substringAfter(" ").take(5),
                    modifier = Modifier.padding(top = 6.dp),
                    fontSize = 10.sp,
                    color = if (isUser) Color.White.copy(alpha = 0.7f) else Color(0xFF9CA3AF)
                )
            }
        }
        if (isUser) {
            Spacer(modifier = Modifier.width(8.dp))
        }
    }
}

@Composable
private fun TypingIndicator() {
    Row(
        modifier = Modifier
            .padding(start = 44.dp)
            .background(Color.White, RoundedCornerShape(16.dp))
            .padding(horizontal = 16.dp, vertical = 14.dp),
        verticalAlignment = Alignment.CenterVertically
    ) {
        Text("AI正在思考", fontSize = 13.sp, color = MaterialTheme.colorScheme.onSurfaceVariant)
        Spacer(modifier = Modifier.width(6.dp))
        repeat(3) { i ->
            val delay = i * 0.15f
            Box(
                modifier = Modifier
                    .size(6.dp)
                    .background(MaterialTheme.colorScheme.primary, RoundedCornerShape(3.dp))
            )
        }
    }
}

private fun sendMessage(
    text: String,
    messages: MutableList<AIMessage>,
    appState: com.dispute.app.AppState,
    scope: androidx.compose.runtime.CoroutineScope,
    onTyping: (Boolean) -> Unit
) {
    if (text.isBlank()) return

    val now = Clock.System.now().toLocalDateTime(TimeZone.currentSystemDefault())
    val timestamp = "${now.date} ${now.hour.toString().padStart(2, '0')}:${now.minute.toString().padStart(2, '0')}:${now.second.toString().padStart(2, '0')}"

    val userMsg = AIMessage(
        id = java.util.UUID.randomUUID().toString(),
        role = AIMessage.Role.USER,
        content = text,
        timestamp = timestamp
    )
    messages.add(userMsg)

    scope.launch {
        onTyping(true)
        kotlinx.coroutines.delay(800 + (500..1500).random().toLong())
        onTyping(false)

        val reply = getMockReply(text)
        val replyTimestamp = kotlinx.datetime.Clock.System.now()
            .toLocalDateTime(kotlinx.datetime.TimeZone.currentSystemDefault())
            .let { "${it.date} ${it.hour.toString().padStart(2, '0')}:${it.minute.toString().padStart(2, '0')}:${it.second.toString().padStart(2, '0')}" }

        messages.add(
            AIMessage(
                id = java.util.UUID.randomUUID().toString(),
                role = AIMessage.Role.ASSISTANT,
                content = reply,
                timestamp = replyTimestamp
            )
        )
    }
}

private fun getMockReply(question: String): String {
    val replies = mapOf(
        "公司拖欠工资怎么办？" to """您好，针对拖欠工资问题，建议您按以下步骤处理：

1️⃣ **收集证据**：劳动合同、工资条、考勤记录、聊天记录等
2️⃣ **协商解决**：先与公司管理层沟通，要求限期支付
3️⃣ **劳动监察投诉**：拨打12333或到人社局投诉
4️⃣ **劳动仲裁**：到劳动仲裁委员会申请（免费，时效1年）
5️⃣ **法院起诉**：对仲裁结果不服可起诉

💡 建议同时到调解中心登记，我们的调解员可协助沟通。""",
        "邻里噪音纠纷如何处理？" to """您好，噪音扰民可通过以下途径解决：

1️⃣ **友好沟通**：先礼貌地与邻居说明情况
2️⃣ **收集证据**：在扰民时段录音录像，记录时间频次
3️⃣ **物业调解**：联系物业或居委会协助调解
4️⃣ **报警处理**：夜间10点后可拨打110，警方可警告处罚
5️⃣ **法律诉讼**：严重扰民可起诉要求停止侵害并赔偿

🏛️ 建议到调解中心登记，调解员可帮助双方达成谅解。""",
        "合同到期不续签有赔偿吗？" to """您好，劳动合同到期补偿问题：

✅ **单位不续签**：有经济补偿
  - 每满1年支付1个月工资
  - 6个月以上按1年算，不足6个月付半月工资

❌ **员工主动不续签**：无补偿（单位降低条件除外）

📌 **特别提示**：
  • 连续签2次固定期合同后，可要求签无固定期合同
  • 建议到期前30天确认意向
  • 如单位未提前通知可能需付代通知金

需要帮助请登记纠纷，专业律师为您解答。""",
        "交通事故责任怎么认定？" to """您好，交通事故责任认定流程：

1️⃣ **保护现场**：拍照录像后移车至安全地带
2️⃣ **报警处理**：交警勘察后出具《事故认定书》
3️⃣ **责任类型**：全部/主要/同等/次要/无责任
4️⃣ **复核申请**：对认定不服可3日内向上级申请复核

💰 赔偿顺序：交强险 → 商业险 → 责任人承担

📋 常见责任划分：
  • 追尾：后车全责
  • 闯红灯：违规方全责
  • 转弯让直行：转弯车主要责任

建议保管好认定书，如赔偿纠纷可到调解中心申请调解。""",
        "借贷没有借条能起诉吗？" to """您好，没有借条也可以起诉，但需要其他证据：

✅ **可作为证据的材料**：
  • 转账记录（银行/微信/支付宝流水）
  • 聊天记录（承认借款的对话）
  • 录音录像（对方承认借款的内容）
  • 证人证言
  • 还款承诺或计划

⚠️ **注意时效**：一般诉讼时效3年

💡 **建议步骤**：
  1. 先与对方沟通并录音取证
  2. 微信/短信确认借款事实
  3. 补写借条或还款计划
  4. 证据充足后可起诉或调解

建议携带证据到调解中心咨询，调解员可协助取证和调解。"""
    )
    return replies[question] ?: """感谢您的咨询！

针对您的问题，建议：

1️⃣ **收集证据**：保留所有相关书面材料、聊天记录、凭证
2️⃣ **尝试协商**：先与对方友好沟通解决方案
3️⃣ **申请调解**：到调解中心登记，专业调解员免费协助
4️⃣ **法律途径**：调解不成可申请仲裁或起诉

您可以详细描述一下具体情况吗？例如：
  • 事情发生的时间和地点
  • 涉及的金额或诉求
  • 目前已有的证据材料

这样我可以为您提供更具体的建议。如需帮助，可登记纠纷寻求专业调解服务。"""
}
