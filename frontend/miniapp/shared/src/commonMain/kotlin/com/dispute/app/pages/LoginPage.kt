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
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.text.KeyboardOptions
import androidx.compose.material3.Button
import androidx.compose.material3.ButtonDefaults
import androidx.compose.material3.Checkbox
import androidx.compose.material3.CheckboxDefaults
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.OutlinedTextField
import androidx.compose.material3.OutlinedTextFieldDefaults
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
import androidx.compose.ui.text.input.KeyboardType
import androidx.compose.ui.text.input.PasswordVisualTransformation
import androidx.compose.ui.text.input.VisualTransformation
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.dispute.app.LocalAppState
import com.dispute.app.LocalApiClient
import com.dispute.app.LocalRouter
import com.dispute.app.Route
import com.dispute.app.model.MockData
import com.dispute.app.model.User
import kotlinx.coroutines.launch

enum class LoginMode {
    WECHAT, PHONE
}

@Composable
fun LoginPage() = LoginContent()

@Composable
private fun LoginContent() {
    val appState = LocalAppState.current
    val router = LocalRouter.current
    val apiClient = LocalApiClient.current

    var loginMode by remember { mutableStateOf(LoginMode.WECHAT) }
    var phone by remember { mutableStateOf("") }
    var smsCode by remember { mutableStateOf("") }
    var agreeTerms by remember { mutableStateOf(false) }
    var showPassword by remember { mutableStateOf(false) }
    var countdown by remember { mutableStateOf(0) }

    if (countdown > 0) {
        androidx.compose.runtime.LaunchedEffect(countdown) {
            kotlinx.coroutines.delay(1000)
            countdown--
        }
    }

    val phoneValid = remember(phone) { phone.matches(Regex("^1[3-9]\\d{9}$")) }
    val canSubmit = agreeTerms && when (loginMode) {
        LoginMode.WECHAT -> true
        LoginMode.PHONE -> phoneValid && smsCode.length == 6
    }

    Column(
        modifier = Modifier
            .fillMaxSize()
            .background(MaterialTheme.colorScheme.background)
            .padding(horizontal = 24.dp)
    ) {
        Spacer(modifier = Modifier.height(60.dp))

        Column(
            horizontalAlignment = Alignment.CenterHorizontally,
            modifier = Modifier.fillMaxWidth()
        ) {
            Box(
                modifier = Modifier
                    .size(80.dp)
                    .background(
                        androidx.compose.ui.graphics.Brush.linearGradient(
                            colors = listOf(Color(0xFF1D6CFF), Color(0xFF4D8CFF))
                        ),
                        RoundedCornerShape(24.dp)
                    ),
                contentAlignment = Alignment.Center
            ) {
                Text("⚖️", fontSize = 44.sp)
            }

            Spacer(modifier = Modifier.height(20.dp))

            Text(
                text = "纠纷调解服务平台",
                style = MaterialTheme.typography.headlineMedium,
                fontWeight = FontWeight.Bold
            )

            Text(
                text = "公正 · 高效 · 便民",
                style = MaterialTheme.typography.bodyMedium,
                color = MaterialTheme.colorScheme.onSurfaceVariant,
                modifier = Modifier.padding(top = 8.dp)
            )
        }

        Spacer(modifier = Modifier.height(48.dp))

        Row(
            modifier = Modifier
                .fillMaxWidth()
                .background(
                    MaterialTheme.colorScheme.surfaceVariant.copy(alpha = 0.5f),
                    RoundedCornerShape(12.dp)
                )
                .padding(4.dp),
            horizontalArrangement = Arrangement.spacedBy(4.dp)
        ) {
            LoginModeTab(
                title = "微信登录",
                selected = loginMode == LoginMode.WECHAT,
                onClick = { loginMode = LoginMode.WECHAT },
                modifier = Modifier.weight(1f)
            )
            LoginModeTab(
                title = "手机号登录",
                selected = loginMode == LoginMode.PHONE,
                onClick = { loginMode = LoginMode.PHONE },
                modifier = Modifier.weight(1f)
            )
        }

        Spacer(modifier = Modifier.height(32.dp))

        when (loginMode) {
            LoginMode.WECHAT -> WechatLoginSection(
                onWechatLogin = {
                    appState.launchWithLoading {
                        val mockUser = User(
                            id = "user001",
                            nickname = "微信用户",
                            avatar = null,
                            phone = "138****8000",
                            realName = "张三",
                            isVerified = true,
                            wechatOpenId = "wx_openid_001"
                        )
                        appState.setUser(mockUser)
                        appState.setCaseList(MockData.mockCases)
                        appState.loadCurrentGridWorker(apiClient)
                        router.navigateToHome()
                    }
                }
            )

            LoginMode.PHONE -> PhoneLoginSection(
                phone = phone,
                onPhoneChange = { phone = it },
                smsCode = smsCode,
                onSmsCodeChange = { smsCode = it },
                showPassword = showPassword,
                onTogglePassword = { showPassword = !showPassword },
                countdown = countdown,
                onSendSms = {
                    if (!phoneValid) {
                        appState.showToast("请输入正确的手机号")
                    } else if (countdown == 0) {
                        countdown = 60
                        appState.showToast("验证码已发送")
                    }
                },
                phoneValid = phoneValid
            )
        }

        Spacer(modifier = Modifier.height(32.dp))

        Row(
            modifier = Modifier.fillMaxWidth(),
            verticalAlignment = Alignment.CenterVertically
        ) {
            Checkbox(
                checked = agreeTerms,
                onCheckedChange = { agreeTerms = it },
                colors = CheckboxDefaults.colors(
                    checkedColor = MaterialTheme.colorScheme.primary
                )
            )
            Text(
                text = "我已阅读并同意",
                style = MaterialTheme.typography.bodySmall,
                color = MaterialTheme.colorScheme.onSurfaceVariant
            )
            Text(
                text = "《用户协议》",
                style = MaterialTheme.typography.bodySmall,
                color = MaterialTheme.colorScheme.primary,
                modifier = Modifier.clickable { }
            )
            Text(
                text = "和",
                style = MaterialTheme.typography.bodySmall,
                color = MaterialTheme.colorScheme.onSurfaceVariant
            )
            Text(
                text = "《隐私政策》",
                style = MaterialTheme.typography.bodySmall,
                color = MaterialTheme.colorScheme.primary,
                modifier = Modifier.clickable { }
            )
        }

        Spacer(modifier = Modifier.height(24.dp))

        Button(
            onClick = {
                if (!agreeTerms) {
                    appState.showToast("请先同意用户协议和隐私政策")
                    return@Button
                }

                appState.launchWithLoading {
                    val mockUser = User(
                        id = "user001",
                        nickname = "用户$phone",
                        avatar = null,
                        phone = phone.ifBlank { "13800138000" },
                        realName = "张三",
                        isVerified = true
                    )
                    appState.setUser(mockUser)
                    appState.setCaseList(MockData.mockCases)
                    appState.loadCurrentGridWorker(apiClient)
                    router.navigateToHome()
                }
            },
            modifier = Modifier
                .fillMaxWidth()
                .height(52.dp),
            shape = RoundedCornerShape(16.dp),
            enabled = canSubmit,
            colors = ButtonDefaults.buttonColors(
                containerColor = MaterialTheme.colorScheme.primary,
                disabledContainerColor = MaterialTheme.colorScheme.primary.copy(alpha = 0.35f)
            )
        ) {
            Text(
                text = when (loginMode) {
                    LoginMode.WECHAT -> "微信登录"
                    LoginMode.PHONE -> "登录"
                },
                fontSize = 16.sp,
                fontWeight = FontWeight.SemiBold
            )
        }

        Spacer(modifier = Modifier.weight(1f))

        Text(
            text = "登录即代表您已了解服务说明及相关权利",
            style = MaterialTheme.typography.labelMedium,
            color = MaterialTheme.colorScheme.onSurfaceVariant,
            modifier = Modifier
                .align(Alignment.CenterHorizontally)
                .padding(bottom = 24.dp)
        )
    }
}

@Composable
private fun LoginModeTab(
    title: String,
    selected: Boolean,
    onClick: () -> Unit,
    modifier: Modifier = Modifier
) {
    val bgColor = if (selected) Color.White else Color.Transparent
    val textColor = if (selected) {
        MaterialTheme.colorScheme.primary
    } else {
        MaterialTheme.colorScheme.onSurfaceVariant
    }

    Box(
        modifier = modifier
            .background(bgColor, RoundedCornerShape(10.dp))
            .clickable(onClick = onClick)
            .padding(vertical = 12.dp),
        contentAlignment = Alignment.Center
    ) {
        Text(
            text = title,
            style = MaterialTheme.typography.titleMedium,
            color = textColor,
            fontWeight = if (selected) FontWeight.SemiBold else FontWeight.Normal
        )
    }
}

@Composable
private fun WechatLoginSection(onWechatLogin: () -> Unit) {
    Column(
        horizontalAlignment = Alignment.CenterHorizontally,
        modifier = Modifier.fillMaxWidth()
    ) {
        Box(
            modifier = Modifier
                .size(140.dp)
                .background(
                    Color(0xFF07C160).copy(alpha = 0.1f),
                    RoundedCornerShape(24.dp)
                ),
            contentAlignment = Alignment.Center
        ) {
            Text("💬", fontSize = 72.sp)
        }

        Spacer(modifier = Modifier.height(24.dp))

        Text(
            text = "使用微信一键登录",
            style = MaterialTheme.typography.titleMedium,
            fontWeight = FontWeight.SemiBold
        )

        Text(
            text = "快捷、安全，无需注册",
            style = MaterialTheme.typography.bodyMedium,
            color = MaterialTheme.colorScheme.onSurfaceVariant,
            modifier = Modifier.padding(top = 8.dp)
        )

        Spacer(modifier = Modifier.height(24.dp))

        Button(
            onClick = onWechatLogin,
            modifier = Modifier
                .fillMaxWidth()
                .height(52.dp),
            shape = RoundedCornerShape(16.dp),
            colors = ButtonDefaults.buttonColors(
                containerColor = Color(0xFF07C160)
            )
        ) {
            Text(
                text = "微信一键登录",
                fontSize = 16.sp,
                fontWeight = FontWeight.SemiBold
            )
        }
    }
}

@Composable
private fun PhoneLoginSection(
    phone: String,
    onPhoneChange: (String) -> Unit,
    smsCode: String,
    onSmsCodeChange: (String) -> Unit,
    showPassword: Boolean,
    onTogglePassword: () -> Unit,
    countdown: Int,
    onSendSms: () -> Unit,
    phoneValid: Boolean
) {
    Column {
        OutlinedTextField(
            value = phone,
            onValueChange = { if (it.length <= 11) onPhoneChange(it) },
            placeholder = {
                Text(
                    "请输入手机号",
                    color = MaterialTheme.colorScheme.onSurfaceVariant
                )
            },
            leadingIcon = {
                Box(
                    modifier = Modifier
                        .padding(start = 8.dp)
                        .background(
                            MaterialTheme.colorScheme.primary.copy(alpha = 0.1f),
                            RoundedCornerShape(8.dp)
                        )
                        .padding(horizontal = 12.dp, vertical = 8.dp)
                ) {
                    Text("🇨🇳 +86", fontSize = 14.sp, fontWeight = FontWeight.Medium)
                }
            },
            modifier = Modifier
                .fillMaxWidth()
                .height(56.dp),
            shape = RoundedCornerShape(14.dp),
            keyboardOptions = KeyboardOptions(keyboardType = KeyboardType.Phone),
            colors = OutlinedTextFieldDefaults.colors(
                focusedBorderColor = MaterialTheme.colorScheme.primary,
                unfocusedBorderColor = Color(0xFFE5E7EB)
            ),
            isError = phone.isNotEmpty() && !phoneValid
        )

        if (phone.isNotEmpty() && !phoneValid) {
            Text(
                text = "请输入正确的手机号格式",
                color = MaterialTheme.colorScheme.error,
                style = MaterialTheme.typography.labelMedium,
                modifier = Modifier.padding(start = 4.dp, top = 6.dp)
            )
        }

        Spacer(modifier = Modifier.height(16.dp))

        Row(
            modifier = Modifier.fillMaxWidth(),
            horizontalArrangement = Arrangement.spacedBy(12.dp),
            verticalAlignment = Alignment.CenterVertically
        ) {
            OutlinedTextField(
                value = smsCode,
                onValueChange = { if (it.length <= 6) onSmsCodeChange(it) },
                placeholder = {
                    Text(
                        "请输入验证码",
                        color = MaterialTheme.colorScheme.onSurfaceVariant
                    )
                },
                modifier = Modifier
                    .weight(1f)
                    .height(56.dp),
                shape = RoundedCornerShape(14.dp),
                keyboardOptions = KeyboardOptions(keyboardType = KeyboardType.NumberPassword),
                visualTransformation = if (showPassword) VisualTransformation.None else PasswordVisualTransformation(),
                trailingIcon = {
                    Text(
                        text = if (showPassword) "👁️" else "👁️‍🗨️",
                        fontSize = 20.sp,
                        modifier = Modifier
                            .padding(end = 12.dp)
                            .clickable(onClick = onTogglePassword)
                    )
                },
                colors = OutlinedTextFieldDefaults.colors(
                    focusedBorderColor = MaterialTheme.colorScheme.primary,
                    unfocusedBorderColor = Color(0xFFE5E7EB)
                )
            )

            Button(
                onClick = onSendSms,
                modifier = Modifier
                    .height(56.dp)
                    .width(130.dp),
                shape = RoundedCornerShape(14.dp),
                enabled = phoneValid && countdown == 0,
                colors = ButtonDefaults.buttonColors(
                    containerColor = MaterialTheme.colorScheme.primary.copy(alpha = 0.15f),
                    contentColor = MaterialTheme.colorScheme.primary,
                    disabledContainerColor = MaterialTheme.colorScheme.surfaceVariant.copy(alpha = 0.5f),
                    disabledContentColor = MaterialTheme.colorScheme.onSurfaceVariant
                )
            ) {
                Text(
                    text = if (countdown > 0) "${countdown}s" else "获取验证码",
                    fontSize = 13.sp,
                    fontWeight = FontWeight.SemiBold
                )
            }
        }
    }
}
