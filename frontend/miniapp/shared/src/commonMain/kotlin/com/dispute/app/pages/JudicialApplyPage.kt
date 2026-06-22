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
import com.dispute.app.Route
import com.dispute.app.components.InputField
import com.dispute.app.components.LargeButton
import com.dispute.app.components.PickerField
import com.dispute.app.components.TextAreaField
import com.dispute.app.components.TopBarWithBackList
import com.dispute.app.model.CourtOption
import com.dispute.app.model.MockData
import androidx.compose.material3.Divider
import androidx.compose.foundation.clickable
import androidx.compose.runtime.LaunchedEffect

@Composable
fun JudicialApplyPage() = JudicialApplyContent()

@Composable
private fun JudicialApplyContent() {
    val appState = LocalAppState.current
    val router = LocalRouter.current

    var caseNo by remember { mutableStateOf("") }
    var caseTitle by remember { mutableStateOf("") }
    var applicantName by remember { mutableStateOf("") }
    var applicantIdCard by remember { mutableStateOf("") }
    var applicantPhone by remember { mutableStateOf("") }
    var respondentName by remember { mutableStateOf("") }
    var respondentPhone by remember { mutableStateOf("") }
    var agreementContent by remember { mutableStateOf("") }
    var confirmAmount by remember { mutableStateOf("") }
    var performanceDeadline by remember { mutableStateOf("") }
    var courtId by remember { mutableStateOf<Long?>(null) }
    var courtName by remember { mutableStateOf("") }
    var remark by remember { mutableStateOf("") }

    var courtOptions by remember { mutableStateOf<List<CourtOption>>(emptyList()) }
    var showCourtPicker by remember { mutableStateOf(false) }
    var showDatePicker by remember { mutableStateOf(false) }
    var isSubmitting by remember { mutableStateOf(false) }
    var validationErrors by remember { mutableStateOf<Map<String, String>>(emptyMap()) }

    LaunchedEffect(Unit) {
        courtOptions = MockData.mockCourtOptions
    }

    fun validate(): Boolean {
        val errors = mutableMapOf<String, String>()
        if (caseNo.isBlank()) errors["caseNo"] = "请输入案件编号"
        if (caseTitle.isBlank()) errors["caseTitle"] = "请输入案件标题"
        if (applicantName.isBlank()) errors["applicantName"] = "请输入申请人姓名"
        if (applicantPhone.isBlank()) errors["applicantPhone"] = "请输入申请人手机号"
        if (respondentName.isBlank()) errors["respondentName"] = "请输入被申请人姓名"
        if (respondentPhone.isBlank()) errors["respondentPhone"] = "请输入被申请人手机号"
        if (agreementContent.isBlank()) errors["agreementContent"] = "请输入协议内容"
        if (courtId == null) errors["courtId"] = "请选择管辖法院"
        if (performanceDeadline.isBlank()) errors["performanceDeadline"] = "请选择履行期限"
        validationErrors = errors
        return errors.isEmpty()
    }

    fun handleSubmit() {
        if (!validate()) return
        isSubmitting = true

        appState.appScope.launch {
            try {
                val request = com.dispute.app.model.CreateJudicialRequest(
                    caseNo = caseNo,
                    caseTitle = caseTitle,
                    applicantName = applicantName,
                    applicantIdCard = applicantIdCard.takeIf { it.isNotBlank() },
                    applicantPhone = applicantPhone,
                    respondentName = respondentName,
                    respondentPhone = respondentPhone,
                    agreementContent = agreementContent,
                    confirmAmount = confirmAmount.toDoubleOrNull(),
                    performanceDeadline = performanceDeadline,
                    courtId = courtId!!,
                    courtName = courtName,
                    remark = remark.takeIf { it.isNotBlank() }
                )

                val result = com.dispute.app.api.JudicialApi.createJudicialConfirmation(request)
                if (result.success) {
                    result.data?.let {
                        appState.addJudicialConfirmation(it)
                        appState.setSelectedJudicial(it)
                    }
                    router.back()
                } else {
                    validationErrors = mapOf("submit" to (result.message ?: "提交失败，请重试"))
                }
            } catch (e: Exception) {
                validationErrors = mapOf("submit" to "网络异常，请检查网络后重试")
            } finally {
                isSubmitting = false
            }
        }
    }

    Column(
        modifier = Modifier
            .fillMaxSize()
            .background(MaterialTheme.colorScheme.background)
    ) {
        TopBarWithBackList(
            title = "申请司法确认",
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
                NoticeCard()
            }

            item {
                SectionHeader("案件信息")
            }

            item {
                InputField(
                    label = "案件编号",
                    placeholder = "请输入调解案件编号",
                    value = caseNo,
                    onValueChange = { caseNo = it },
                    error = validationErrors["caseNo"],
                    required = true
                )
            }

            item {
                InputField(
                    label = "案件标题",
                    placeholder = "请输入案件标题",
                    value = caseTitle,
                    onValueChange = { caseTitle = it },
                    error = validationErrors["caseTitle"],
                    required = true
                )
            }

            item {
                SectionHeader("申请人信息")
            }

            item {
                Row(
                    modifier = Modifier.fillMaxWidth(),
                    horizontalArrangement = Arrangement.spacedBy(12.dp)
                ) {
                    InputField(
                        modifier = Modifier.weight(1f),
                        label = "姓名",
                        placeholder = "请输入姓名",
                        value = applicantName,
                        onValueChange = { applicantName = it },
                        error = validationErrors["applicantName"],
                        required = true
                    )
                    InputField(
                        modifier = Modifier.weight(1.5f),
                        label = "手机号",
                        placeholder = "请输入手机号",
                        value = applicantPhone,
                        onValueChange = { applicantPhone = it },
                        error = validationErrors["applicantPhone"],
                        required = true
                    )
                }
            }

            item {
                InputField(
                    label = "身份证号",
                    placeholder = "请输入身份证号（选填）",
                    value = applicantIdCard,
                    onValueChange = { applicantIdCard = it },
                    error = validationErrors["applicantIdCard"],
                    required = false
                )
            }

            item {
                SectionHeader("被申请人信息")
            }

            item {
                Row(
                    modifier = Modifier.fillMaxWidth(),
                    horizontalArrangement = Arrangement.spacedBy(12.dp)
                ) {
                    InputField(
                        modifier = Modifier.weight(1f),
                        label = "姓名",
                        placeholder = "请输入姓名",
                        value = respondentName,
                        onValueChange = { respondentName = it },
                        error = validationErrors["respondentName"],
                        required = true
                    )
                    InputField(
                        modifier = Modifier.weight(1.5f),
                        label = "手机号",
                        placeholder = "请输入手机号",
                        value = respondentPhone,
                        onValueChange = { respondentPhone = it },
                        error = validationErrors["respondentPhone"],
                        required = true
                    )
                }
            }

            item {
                SectionHeader("协议信息")
            }

            item {
                PickerField(
                    label = "管辖法院",
                    placeholder = "请选择管辖法院",
                    value = courtName,
                    onValueChange = { },
                    onClick = { showCourtPicker = true },
                    error = validationErrors["courtId"],
                    required = true
                )
            }

            item {
                TextAreaField(
                    label = "调解协议内容",
                    placeholder = "请详细描述调解协议内容",
                    value = agreementContent,
                    onValueChange = { agreementContent = it },
                    error = validationErrors["agreementContent"],
                    required = true,
                    minLines = 6
                )
            }

            item {
                Row(
                    modifier = Modifier.fillMaxWidth(),
                    horizontalArrangement = Arrangement.spacedBy(12.dp)
                ) {
                    InputField(
                        modifier = Modifier.weight(1f),
                        label = "确认金额（元）",
                        placeholder = "选填",
                        value = confirmAmount,
                        onValueChange = { confirmAmount = it },
                        keyboardType = androidx.compose.ui.text.input.KeyboardType.Decimal,
                        required = false
                    )
                    PickerField(
                        modifier = Modifier.weight(1f),
                        label = "履行期限",
                        placeholder = "请选择",
                        value = performanceDeadline,
                        onValueChange = { },
                        onClick = { showDatePicker = true },
                        error = validationErrors["performanceDeadline"],
                        required = true
                    )
                }
            }

            item {
                TextAreaField(
                    label = "备注",
                    placeholder = "请输入其他需要说明的信息（选填）",
                    value = remark,
                    onValueChange = { remark = it },
                    required = false,
                    minLines = 3
                )
            }

            item {
                validationErrors["submit"]?.let {
                    Box(
                        modifier = Modifier
                            .fillMaxWidth()
                            .background(Color(0xFFFEF2F2), RoundedCornerShape(8.dp))
                            .padding(12.dp)
                    ) {
                        Text(
                            text = it,
                            color = Color(0xFFDC2626),
                            style = MaterialTheme.typography.bodyMedium
                        )
                    }
                }
            }

            item {
                Spacer(modifier = Modifier.height(16.dp))
            }

            item {
                LargeButton(
                    text = "提交申请",
                    onClick = { handleSubmit() },
                    loading = isSubmitting,
                    enabled = !isSubmitting
                )
            }

            item {
                Spacer(modifier = Modifier.height(32.dp))
            }
        }
    }

    if (showCourtPicker) {
        CourtPickerBottomSheet(
            options = courtOptions,
            selectedId = courtId,
            onDismiss = { showCourtPicker = false },
            onSelected = { option ->
                courtId = option.id
                courtName = option.name
                showCourtPicker = false
            }
        )
    }
}

@Composable
private fun NoticeCard() {
    androidx.compose.material3.Card(
        modifier = Modifier.fillMaxWidth(),
        shape = RoundedCornerShape(12.dp),
        colors = androidx.compose.material3.CardDefaults.cardColors(
            containerColor = Color(0xFFFEF3C7)
        ),
        elevation = androidx.compose.material3.CardDefaults.cardElevation(defaultElevation = 0.dp)
    ) {
        Column(
            modifier = Modifier
                .fillMaxWidth()
                .padding(16.dp)
        ) {
            Row(
                verticalAlignment = Alignment.Top
            ) {
                Text("📋", fontSize = 18.sp)
                Spacer(modifier = Modifier.width(12.dp))
                Column(modifier = Modifier.weight(1f)) {
                    Text(
                        text = "申请须知",
                        style = MaterialTheme.typography.titleMedium,
                        fontWeight = FontWeight.SemiBold,
                        color = Color(0xFF92400E)
                    )
                    Spacer(modifier = Modifier.height(8.dp))
                    Text(
                        text = "1. 司法确认是指人民法院对调解协议进行审查，依法确认调解协议效力的程序\n" +
                               "2. 经人民法院确认有效的调解协议，具有强制执行力\n" +
                               "3. 申请司法确认不收取诉讼费用\n" +
                               "4. 请如实填写信息，提供虚假信息将承担相应法律责任",
                        style = MaterialTheme.typography.bodyMedium,
                        color = Color(0xFF78350F),
                        lineHeight = 22.sp
                    )
                }
            }
        }
    }
}

@Composable
private fun SectionHeader(title: String) {
    Column(modifier = Modifier.padding(top = 8.dp)) {
        Text(
            text = title,
            style = MaterialTheme.typography.titleMedium,
            fontWeight = FontWeight.SemiBold,
            color = MaterialTheme.colorScheme.onSurface
        )
        Spacer(modifier = Modifier.height(12.dp))
    }
}

@Composable
private fun CourtPickerBottomSheet(
    options: List<CourtOption>,
    selectedId: Long?,
    onDismiss: () -> Unit,
    onSelected: (CourtOption) -> Unit
) {
    androidx.compose.material3.ModalBottomSheet(
        onDismissRequest = onDismiss,
        containerColor = MaterialTheme.colorScheme.surface,
        shape = RoundedCornerShape(topStart = 20.dp, topEnd = 20.dp)
    ) {
        Column(
            modifier = Modifier
                .fillMaxWidth()
                .padding(20.dp)
        ) {
            Text(
                text = "选择管辖法院",
                style = MaterialTheme.typography.titleLarge,
                fontWeight = FontWeight.Bold,
                color = MaterialTheme.colorScheme.onSurface
            )

            Spacer(modifier = Modifier.height(16.dp))

            LazyColumn(
                verticalArrangement = Arrangement.spacedBy(4.dp),
                modifier = Modifier.height(400.dp)
            ) {
                items(options) { option ->
                    val isSelected = option.id == selectedId
                    Row(
                        modifier = Modifier
                            .fillMaxWidth()
                            .background(
                                if (isSelected) MaterialTheme.colorScheme.primary.copy(alpha = 0.1f)
                                else Color.Transparent,
                                RoundedCornerShape(8.dp)
                            )
                            .clickable { onSelected(option) }
                            .padding(16.dp),
                        horizontalArrangement = Arrangement.SpaceBetween,
                        verticalAlignment = Alignment.CenterVertically
                    ) {
                        Column(modifier = Modifier.weight(1f)) {
                            Text(
                                text = option.name,
                                style = MaterialTheme.typography.bodyLarge,
                                fontWeight = if (isSelected) FontWeight.SemiBold else FontWeight.Normal,
                                color = if (isSelected) MaterialTheme.colorScheme.primary
                                else MaterialTheme.colorScheme.onSurface
                            )
                            if (option.address != null) {
                                Spacer(modifier = Modifier.height(4.dp))
                                Text(
                                    text = option.address!!,
                                    style = MaterialTheme.typography.bodyMedium,
                                    color = MaterialTheme.colorScheme.onSurfaceVariant
                                )
                            }
                        }
                        if (isSelected) {
                            Text(
                                text = "✓",
                                style = MaterialTheme.typography.titleLarge,
                                color = MaterialTheme.colorScheme.primary,
                                fontWeight = FontWeight.Bold
                            )
                        }
                    }
                    Divider(
                        color = MaterialTheme.colorScheme.outlineVariant,
                        thickness = 0.5.dp
                    )
                }
            }

            Spacer(modifier = Modifier.height(16.dp))

            LargeButton(
                text = "取消",
                onClick = onDismiss,
                filled = false
            )
        }
    }
}

private fun items(options: List<CourtOption>, block: @Composable (CourtOption) -> Unit) {
    options.forEach { option ->
        block(option)
    }
}

private fun Modifier.height(value: Int) = this.then(Modifier.height(value.dp))
