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
import androidx.compose.runtime.setValue
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
import com.dispute.app.components.TagCard
import com.dispute.app.audio.AudioRecorder
import com.dispute.app.audio.isAudioRecordingSupported
import com.dispute.app.audio.toBase64
import com.dispute.app.model.Case
import com.dispute.app.model.DisputeType
import com.dispute.app.model.Evidence
import com.dispute.app.model.MockData
import kotlinx.coroutines.launch

private data class DraftCase(
    val id: String = "",
    var disputeTypePath: MutableList<String> = mutableStateListOf(),
    var disputeTypeName: String = "",
    var opponentName: String = "",
    var opponentPhone: String = "",
    var opponentAddress: String = "",
    var description: String = "",
    var expectedResolution: MutableList<String> = mutableStateListOf()
)

@Composable
fun RegisterCasePage() = RegisterContent()

@Composable
private fun RegisterContent() {
    val appState = LocalAppState.current
    val router = LocalRouter.current

    var currentStep by remember { mutableStateOf(0) }
    val draft = remember { DraftCase() }
    val user by appState.currentUser
    val disputeTypes = MockData.mockDisputeTypes
    val selectedTypePath = remember { mutableStateListOf<DisputeType>() }
    var currentTypeLevel by remember { mutableStateOf(disputeTypes) }

    val totalSteps = 4

    val canProceed = when (currentStep) {
        0 -> selectedTypePath.isNotEmpty() && !selectedTypePath.last().hasChildren
        1 -> draft.opponentName.isNotBlank() && draft.opponentPhone.isNotBlank()
        2 -> draft.description.length >= 20
        3 -> true
        else -> false
    }

    Column(
        modifier = Modifier
            .fillMaxSize()
            .background(MaterialTheme.colorScheme.background)
    ) {
        TopBarWithBack(
            title = "纠纷登记",
            onBack = {
                if (currentStep > 0) currentStep-- else router.back()
            }
        )

        StepIndicator(
            currentStep = currentStep,
            totalSteps = totalSteps,
            labels = listOf("纠纷类型", "对方信息", "情况描述", "确认提交")
        )

        LazyColumn(
            modifier = Modifier
                .fillMaxWidth()
                .weight(1f)
                .padding(horizontal = 16.dp),
            verticalArrangement = Arrangement.spacedBy(16.dp)
        ) {
            when (currentStep) {
                0 -> item {
                    DisputeTypeSelection(
                        path = selectedTypePath,
                        currentLevel = currentTypeLevel,
                        onSelect = { type ->
                            if (type.hasChildren) {
                                selectedTypePath.add(type)
                                currentTypeLevel = type.children!!
                            } else {
                                if (selectedTypePath.isEmpty() || selectedTypePath.last().id != type.id) {
                                    val lastParent = if (selectedTypePath.isNotEmpty()) {
                                        selectedTypePath.dropLast(1)
                                    } else {
                                        selectedTypePath.toList()
                                    }
                                    selectedTypePath.clear()
                                    selectedTypePath.addAll(lastParent)
                                    selectedTypePath.add(type)
                                }
                                draft.disputeTypePath.clear()
                                draft.disputeTypePath.addAll(selectedTypePath.map { it.name })
                                draft.disputeTypeName = type.name
                            }
                        },
                        onBack = {
                            if (selectedTypePath.isNotEmpty()) {
                                selectedTypePath.removeLast()
                                currentTypeLevel = if (selectedTypePath.isEmpty()) {
                                    disputeTypes
                                } else {
                                    selectedTypePath.last().children!!
                                }
                            }
                        }
                    )
                }

                1 -> item {
                    OpponentInfoSection(
                        name = draft.opponentName,
                        onNameChange = { draft.opponentName = it },
                        phone = draft.opponentPhone,
                        onPhoneChange = { draft.opponentPhone = it },
                        address = draft.opponentAddress,
                        onAddressChange = { draft.opponentAddress = it }
                    )
                }

                2 -> item {
                    DescriptionSection(
                        description = draft.description,
                        onDescriptionChange = { draft.description = it },
                        expectations = draft.expectedResolution,
                        onExpectationsChange = { draft.expectedResolution.clear(); draft.expectedResolution.addAll(it) }
                    )
                }

                3 -> item {
                    ConfirmSection(
                        user = user,
                        draft = draft,
                        typePath = selectedTypePath
                    )
                }
            }

            item { Spacer(modifier = Modifier.height(16.dp)) }
        }

        BottomActionButtons(
            onBack = {
                if (currentStep > 0) {
                    currentStep--
                    if (currentStep == 0) {
                        currentTypeLevel = if (selectedTypePath.isNotEmpty()) {
                            selectedTypePath.last().children ?: disputeTypes
                        } else disputeTypes
                    }
                } else router.back()
            },
            onNext = {
                if (currentStep < totalSteps - 1) {
                    currentStep++
                } else {
                    submitCase(appState, router, draft, selectedTypePath, user)
                }
            },
            nextText = if (currentStep == totalSteps - 1) "提交登记" else "下一步",
            canProceed = canProceed,
            showBack = currentStep > 0
        )
    }
}

@Composable
private fun TopBarWithBack(title: String, onBack: () -> Unit) {
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
private fun StepIndicator(
    currentStep: Int,
    totalSteps: Int,
    labels: List<String>
) {
    Column(
        modifier = Modifier
            .fillMaxWidth()
            .padding(horizontal = 16.dp, vertical = 12.dp)
    ) {
        Row(
            modifier = Modifier.fillMaxWidth(),
            horizontalArrangement = Arrangement.SpaceBetween
        ) {
            labels.forEachIndexed { index, label ->
                Column(
                    horizontalAlignment = Alignment.CenterHorizontally,
                    modifier = Modifier.weight(1f)
                ) {
                    Box(
                        modifier = Modifier
                            .size(if (index == currentStep) 30.dp else 24.dp)
                            .background(
                                when {
                                    index < currentStep -> Color(0xFF22C55E)
                                    index == currentStep -> MaterialTheme.colorScheme.primary
                                    else -> Color(0xFFE5E7EB)
                                },
                                androidx.compose.foundation.shape.CircleShape
                            ),
                        contentAlignment = Alignment.Center
                    ) {
                        Text(
                            text = if (index < currentStep) "✓" else (index + 1).toString(),
                            color = if (index <= currentStep) Color.White else Color(0xFF9CA3AF),
                            fontSize = if (index == currentStep) 13.sp else 12.sp,
                            fontWeight = FontWeight.Bold
                        )
                    }
                    Text(
                        text = label,
                        fontSize = 11.sp,
                        color = if (index <= currentStep) MaterialTheme.colorScheme.primary else Color(0xFF9CA3AF),
                        fontWeight = if (index == currentStep) FontWeight.SemiBold else FontWeight.Normal,
                        modifier = Modifier.padding(top = 4.dp)
                    )
                }
            }
        }
    }
}

@Composable
private fun DisputeTypeSelection(
    path: List<DisputeType>,
    currentLevel: List<DisputeType>,
    onSelect: (DisputeType) -> Unit,
    onBack: () -> Unit
) {
    AppCard(
        title = "请选择纠纷类型",
        subtitle = if (path.isNotEmpty()) path.joinToString(" > ") { it.name } else null
    ) {
        Column {
            if (path.isNotEmpty()) {
                Row(
                    verticalAlignment = Alignment.CenterVertically,
                    modifier = Modifier
                        .fillMaxWidth()
                        .background(
                            MaterialTheme.colorScheme.primary.copy(alpha = 0.1f),
                            RoundedCornerShape(8.dp)
                        )
                        .clickable(onClick = onBack)
                        .padding(12.dp)
                ) {
                    Text("← 返回上一级", fontSize = 13.sp, color = MaterialTheme.colorScheme.primary)
                }
                Spacer(modifier = Modifier.height(12.dp))
            }

            Column(verticalArrangement = Arrangement.spacedBy(10.dp)) {
                currentLevel.forEach { type ->
                    val isSelected = path.any { it.id == type.id }
                    Row(
                        modifier = Modifier
                            .fillMaxWidth()
                            .background(
                                if (isSelected) MaterialTheme.colorScheme.primary.copy(alpha = 0.08f)
                                else MaterialTheme.colorScheme.surfaceVariant.copy(alpha = 0.3f),
                                RoundedCornerShape(12.dp)
                            )
                            .then(
                                if (isSelected) Modifier.background(
                                    MaterialTheme.colorScheme.primary.copy(alpha = 0.08f),
                                    RoundedCornerShape(12.dp)
                                ) else Modifier
                            )
                            .clickable { onSelect(type) }
                            .padding(horizontal = 16.dp, vertical = 14.dp),
                        verticalAlignment = Alignment.CenterVertically
                    ) {
                        Text(
                            text = type.icon ?: "📋",
                            fontSize = 24.sp,
                            modifier = Modifier.padding(end = 12.dp)
                        )
                        Text(
                            text = type.name,
                            style = MaterialTheme.typography.titleSmall,
                            fontWeight = FontWeight.SemiBold,
                            modifier = Modifier.weight(1f)
                        )
                        Text(
                            text = if (type.hasChildren) "›" else "✓",
                            color = if (isSelected) MaterialTheme.colorScheme.primary else Color(0xFF9CA3AF),
                            fontSize = 22.sp,
                            fontWeight = FontWeight.Bold
                        )
                    }
                }
            }
        }
    }
}

@Composable
private fun OpponentInfoSection(
    name: String,
    onNameChange: (String) -> Unit,
    phone: String,
    onPhoneChange: (String) -> Unit,
    address: String,
    onAddressChange: (String) -> Unit
) {
    Column(verticalArrangement = Arrangement.spacedBy(16.dp)) {
        AppCard(title = "对方当事人信息", subtitle = "请准确填写以便调解员联系") {
            Column(verticalArrangement = Arrangement.spacedBy(14.dp)) {
                LabeledTextField(
                    label = "对方姓名/单位名称",
                    value = name,
                    onValueChange = onNameChange,
                    placeholder = "请输入姓名或单位名称"
                )
                LabeledTextField(
                    label = "联系电话",
                    value = phone,
                    onValueChange = { if (it.length <= 11) onPhoneChange(it) },
                    placeholder = "请输入对方联系电话",
                    keyboardType = androidx.compose.foundation.text.KeyboardOptions(
                        keyboardType = androidx.compose.ui.text.input.KeyboardType.Phone
                    )
                )
                LabeledTextField(
                    label = "联系地址",
                    value = address,
                    onValueChange = onAddressChange,
                    placeholder = "请输入详细住址或单位地址（选填）",
                    minLines = 2
                )
            }
        }

        AppCard(
            backgroundColor = Color(0xFFEFF6FF),
            borderColor = Color(0xFFBFDBFE)
        ) {
            Row(verticalAlignment = Alignment.Top) {
                Text("💡", fontSize = 20.sp, modifier = Modifier.padding(end = 10.dp))
                Text(
                    text = "如无法获取对方完整信息，可先填写已知信息，调解员将协助补充。",
                    style = MaterialTheme.typography.bodyMedium,
                    color = Color(0xFF1E40AF)
                )
            }
        }
    }
}

@Composable
private fun DescriptionSection(
    description: String,
    onDescriptionChange: (String) -> Unit,
    expectations: List<String>,
    onExpectationsChange: (List<String>) -> Unit
) {
    val options = listOf(
        "调解解决" to "由调解员帮助双方协商",
        "要求退款" to "要求退还相关款项",
        "要求赔偿" to "要求赔偿实际损失",
        "要求道歉" to "要求对方承认错误",
        "要求整改" to "纠正不当行为",
        "法律咨询" to "了解相关法律途径"
    )

    val selected = remember { mutableStateListOf<String>() }
    selected.addAll(expectations)

    val appState = LocalAppState.current
    val apiClient = LocalApiClient.current

    val isRecording = remember { mutableStateOf(false) }
    val isRecognizing = remember { mutableStateOf(false) }
    val audioRecorder = remember { AudioRecorder() }
    val recordingSupported = remember { isAudioRecordingSupported() }

    fun handleVoiceInput() {
        if (isRecording.value) {
            audioRecorder.stopRecording()
        } else {
            if (!recordingSupported) {
                appState.showToast("当前环境不支持语音输入")
                return
            }

            audioRecorder.setOnRecordingStart {
                isRecording.value = true
            }

            audioRecorder.setOnRecordingStop { audioData, fileName, format ->
                isRecording.value = false
                isRecognizing.value = true

                appState.launchWithLoading {
                    try {
                        val result = apiClient.voice.recognizeSpeech(
                            fileName = fileName,
                            fileBase64 = audioData.toBase64(),
                            format = format
                        )

                        if (result.text.isNotBlank()) {
                            val newText = if (description.isNotBlank()) {
                                description + "\n" + result.text
                            } else {
                                result.text
                            }
                            onDescriptionChange(newText)
                            appState.showToast("语音识别成功，已填入描述")
                        } else {
                            appState.showToast("未识别到语音内容")
                        }
                    } catch (e: Exception) {
                        appState.showToast("语音识别失败: ${e.message}")
                    } finally {
                        isRecognizing.value = false
                    }
                }
            }

            audioRecorder.setOnError { message ->
                isRecording.value = false
                isRecognizing.value = false
                appState.showToast(message)
            }

            audioRecorder.startRecording()
        }
    }

    Column(verticalArrangement = Arrangement.spacedBy(16.dp)) {
        AppCard(title = "纠纷情况描述", subtitle = "请详细描述，有助于调解员了解情况") {
            Column {
                OutlinedTextField(
                    value = description,
                    onValueChange = onDescriptionChange,
                    placeholder = {
                        Text(
                            "请详细描述纠纷情况：\n1. 纠纷发生的时间、地点\n2. 事情的起因和经过\n3. 您受到的损失或影响\n4. 双方沟通过程",
                            color = MaterialTheme.colorScheme.onSurfaceVariant,
                            fontSize = 13.sp
                        )
                    },
                    modifier = Modifier
                        .fillMaxWidth()
                        .height(180.dp),
                    shape = RoundedCornerShape(12.dp),
                    colors = OutlinedTextFieldDefaults.colors(
                        focusedBorderColor = MaterialTheme.colorScheme.primary,
                        unfocusedBorderColor = Color(0xFFE5E7EB)
                    )
                )
                Row(
                    modifier = Modifier.fillMaxWidth().padding(top = 8.dp),
                    horizontalArrangement = Arrangement.SpaceBetween,
                    verticalAlignment = Alignment.CenterVertically
                ) {
                    Row(verticalAlignment = Alignment.CenterVertically) {
                        Text(
                            text = "建议至少20字",
                            style = MaterialTheme.typography.labelMedium,
                            color = if (description.length >= 20) Color(0xFF22C55E) else Color(0xFFF59E0B)
                        )
                        if (recordingSupported) {
                            Spacer(modifier = Modifier.width(12.dp))
                            VoiceInputButton(
                                isRecording = isRecording.value,
                                isRecognizing = isRecognizing.value,
                                onClick = { handleVoiceInput() }
                            )
                        }
                    }
                    Text(
                        text = "${description.length}/2000",
                        style = MaterialTheme.typography.labelMedium,
                        color = MaterialTheme.colorScheme.onSurfaceVariant
                    )
                }
            }
        }

        AppCard(title = "期望解决方式", subtitle = "可多选") {
            Column(verticalArrangement = Arrangement.spacedBy(10.dp)) {
                options.forEach { (label, desc) ->
                    val isSelected = selected.contains(label)
                    Row(
                        modifier = Modifier
                            .fillMaxWidth()
                            .background(
                                if (isSelected) MaterialTheme.colorScheme.primary.copy(alpha = 0.08f)
                                else MaterialTheme.colorScheme.surfaceVariant.copy(alpha = 0.3f),
                                RoundedCornerShape(12.dp)
                            )
                            .clickable {
                                val newList = if (isSelected) {
                                    selected.filter { it != label }
                                } else {
                                    selected + label
                                }
                                selected.clear()
                                selected.addAll(newList)
                                onExpectationsChange(selected.toList())
                            }
                            .padding(14.dp),
                        verticalAlignment = Alignment.CenterVertically
                    ) {
                        Box(
                            modifier = Modifier
                                .size(22.dp)
                                .background(
                                    if (isSelected) MaterialTheme.colorScheme.primary else Color.White,
                                    RoundedCornerShape(6.dp)
                                )
                                .then(
                                    if (!isSelected) Modifier.background(
                                        Color.White,
                                        RoundedCornerShape(6.dp)
                                    ) else Modifier
                                ),
                            contentAlignment = Alignment.Center
                        ) {
                            if (isSelected) {
                                Text("✓", color = Color.White, fontSize = 14.sp, fontWeight = FontWeight.Bold)
                            }
                        }
                        Spacer(modifier = Modifier.width(12.dp))
                        Column(modifier = Modifier.weight(1f)) {
                            Text(
                                text = label,
                                style = MaterialTheme.typography.titleSmall,
                                fontWeight = FontWeight.SemiBold
                            )
                            Text(
                                text = desc,
                                style = MaterialTheme.typography.labelMedium,
                                color = MaterialTheme.colorScheme.onSurfaceVariant
                            )
                        }
                    }
                }
            }
        }
    }
}

@Composable
private fun ConfirmSection(
    user: com.dispute.app.model.User?,
    draft: DraftCase,
    typePath: List<DisputeType>
) {
    Column(verticalArrangement = Arrangement.spacedBy(16.dp)) {
        AppCard(
            backgroundColor = Color(0xFFFEF3C7),
            borderColor = Color(0xFFFCD34D)
        ) {
            Text(
                text = "⚠️ 请仔细核对以下信息，提交后将无法修改。如发现错误请返回上一步修改。",
                style = MaterialTheme.typography.bodyMedium,
                color = Color(0xFF92400E)
            )
        }

        InfoCard(title = "申请人", value = user?.displayName ?: "—")
        InfoCard(title = "联系电话", value = user?.maskedPhone ?: "—")

        AppCard(title = "纠纷类型") {
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.spacedBy(6.dp)
            ) {
                typePath.forEachIndexed { index, type ->
                    Text(
                        text = type.name,
                        style = MaterialTheme.typography.bodyMedium,
                        color = MaterialTheme.colorScheme.primary,
                        fontWeight = FontWeight.SemiBold
                    )
                    if (index < typePath.lastIndex) {
                        Text("›", color = Color(0xFF9CA3AF))
                    }
                }
            }
        }

        AppCard(title = "对方信息") {
            Column(verticalArrangement = Arrangement.spacedBy(10.dp)) {
                InfoCard(title = "姓名/名称", value = draft.opponentName)
                InfoCard(title = "联系电话", value = draft.opponentPhone)
                if (draft.opponentAddress.isNotBlank()) {
                    InfoCard(title = "联系地址", value = draft.opponentAddress)
                }
            }
        }

        AppCard(title = "纠纷描述") {
            Text(
                text = draft.description,
                style = MaterialTheme.typography.bodyMedium,
                lineHeight = 20.sp
            )
        }

        if (draft.expectedResolution.isNotEmpty()) {
            AppCard(title = "期望解决方式") {
                Row(
                    modifier = Modifier.fillMaxWidth(),
                    horizontalArrangement = Arrangement.spacedBy(8.dp)
                ) {
                    draft.expectedResolution.forEach {
                        TagCard(text = it, selected = true) {}
                    }
                }
            }
        }
    }
}

@Composable
private fun BottomActionButtons(
    onBack: () -> Unit,
    onNext: () -> Unit,
    nextText: String,
    canProceed: Boolean,
    showBack: Boolean
) {
    Row(
        modifier = Modifier
            .fillMaxWidth()
            .padding(16.dp),
        horizontalArrangement = Arrangement.spacedBy(12.dp)
    ) {
        if (showBack) {
            Button(
                onClick = onBack,
                modifier = Modifier
                    .weight(1f)
                    .height(50.dp),
                shape = RoundedCornerShape(14.dp),
                colors = ButtonDefaults.buttonColors(
                    containerColor = MaterialTheme.colorScheme.surfaceVariant,
                    contentColor = MaterialTheme.colorScheme.onSurface
                )
            ) {
                Text(text = "上一步", fontWeight = FontWeight.SemiBold)
            }
        }
        Button(
            onClick = onNext,
            modifier = Modifier
                .weight(2f)
                .height(50.dp),
            shape = RoundedCornerShape(14.dp),
            enabled = canProceed,
            colors = ButtonDefaults.buttonColors(
                containerColor = MaterialTheme.colorScheme.primary,
                disabledContainerColor = MaterialTheme.colorScheme.primary.copy(alpha = 0.35f)
            )
        ) {
            Text(text = nextText, fontWeight = FontWeight.SemiBold, fontSize = 15.sp)
        }
    }
}

@Composable
private fun LabeledTextField(
    label: String,
    value: String,
    onValueChange: (String) -> Unit,
    placeholder: String,
    keyboardOptions: androidx.compose.foundation.text.KeyboardOptions = androidx.compose.foundation.text.KeyboardOptions.Default,
    minLines: Int = 1
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
                Text(
                    placeholder,
                    color = MaterialTheme.colorScheme.onSurfaceVariant,
                    fontSize = 13.sp
                )
            },
            modifier = Modifier.fillMaxWidth(),
            shape = RoundedCornerShape(12.dp),
            minLines = minLines,
            keyboardOptions = keyboardOptions,
            colors = OutlinedTextFieldDefaults.colors(
                focusedBorderColor = MaterialTheme.colorScheme.primary,
                unfocusedBorderColor = Color(0xFFE5E7EB)
            )
        )
    }
}

@Composable
private fun VoiceInputButton(
    isRecording: Boolean,
    isRecognizing: Boolean,
    onClick: () -> Unit
) {
    val backgroundColor = when {
        isRecording -> Color(0xFFEF4444)
        isRecognizing -> Color(0xFFF59E0B)
        else -> MaterialTheme.colorScheme.primary
    }

    val contentColor = Color.White

    Box(
        modifier = Modifier
            .size(36.dp)
            .background(
                backgroundColor.copy(alpha = 0.1f),
                RoundedCornerShape(50)
            )
            .clickable {
                if (!isRecognizing) {
                    onClick()
                }
            },
        contentAlignment = Alignment.Center
    ) {
        when {
            isRecognizing -> {
                Text(
                    text = "⏳",
                    fontSize = 16.sp
                )
            }
            isRecording -> {
                Text(
                    text = "🔴",
                    fontSize = 16.sp
                )
            }
            else -> {
                Text(
                    text = "🎤",
                    fontSize = 16.sp
                )
            }
        }
    }
}

private fun submitCase(
    appState: com.dispute.app.AppState,
    router: com.dispute.app.Router,
    draft: DraftCase,
    typePath: List<DisputeType>,
    user: com.dispute.app.model.User?
) {
    appState.launchWithLoading {
        val caseNumber = generateCaseNumber()
        val now = java.text.SimpleDateFormat("yyyy-MM-dd HH:mm:ss", java.util.Locale.getDefault())
            .format(java.util.Date())

        val newCase = Case(
            id = generateId(),
            caseNumber = caseNumber,
            userId = user?.id ?: "",
            applicantName = user?.displayName ?: "",
            applicantPhone = user?.phone ?: "",
            disputeTypePath = typePath.map { it.name },
            disputeTypeName = draft.disputeTypeName,
            opponentName = draft.opponentName,
            opponentPhone = draft.opponentPhone,
            description = draft.description,
            expectedResolution = draft.expectedResolution.joinToString("、"),
            status = Case.Status.PENDING_REVIEW,
            statusText = Case.Status.PENDING_REVIEW.displayName,
            submitTime = now,
            lastUpdateTime = now
        )

        appState.addCase(newCase)
        appState.showToast("登记成功！案件编号：$caseNumber")
        kotlinx.coroutines.delay(500)
        router.navigateWithReplace(Route.CaseDetail(caseNumber))
    }
}

private fun generateCaseNumber(): String {
    val now = java.util.Calendar.getInstance()
    val year = now.get(java.util.Calendar.YEAR)
    val month = String.format("%02d", now.get(java.util.Calendar.MONTH) + 1)
    val day = String.format("%02d", now.get(java.util.Calendar.DAY_OF_MONTH))
    val random = (1000..9999).random()
    return "JF${year}${month}${day}${random}"
}

private fun generateId(): String =
    java.util.UUID.randomUUID().toString().replace("-", "").substring(0, 16)
