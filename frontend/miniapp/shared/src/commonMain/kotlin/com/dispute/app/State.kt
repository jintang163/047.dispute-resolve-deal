package com.dispute.app

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.MutableState
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.setValue
import androidx.compose.runtime.State
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.dispute.app.model.Case
import com.dispute.app.model.User
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.SupervisorJob
import kotlinx.coroutines.launch

class AppState {
    private val _currentUser = mutableStateOf<User?>(null)
    val currentUser: State<User?> = _currentUser

    private val _caseList = mutableStateOf<List<Case>>(emptyList())
    val caseList: State<List<Case>> = _caseList

    private val _isLoading = mutableStateOf(false)
    val isLoading: State<Boolean> = _isLoading

    private val _toastMessage = mutableStateOf<String?>(null)
    val toastMessage: State<String?> = _toastMessage

    private val _selectedCase = mutableStateOf<Case?>(null)
    val selectedCase: State<Case?> = _selectedCase

    private val _caseStatusFilter = mutableStateOf<Case.Status?>(null)
    val caseStatusFilter: State<Case.Status?> = _caseStatusFilter

    private val _aiConversationId = mutableStateOf<String?>(null)
    val aiConversationId: State<String?> = _aiConversationId

    val appScope = CoroutineScope(SupervisorJob() + Dispatchers.Main)

    fun setUser(user: User?) {
        _currentUser.value = user
    }

    fun updateUserInfo(update: (User) -> User) {
        _currentUser.value?.let {
            _currentUser.value = update(it)
        }
    }

    fun setCaseList(cases: List<Case>) {
        _caseList.value = cases
    }

    fun addCase(case: Case) {
        _caseList.value = listOf(case) + _caseList.value
    }

    fun updateCase(caseNumber: String, update: (Case) -> Case) {
        _caseList.value = _caseList.value.map {
            if (it.caseNumber == caseNumber) update(it) else it
        }
    }

    fun setCaseStatusFilter(status: Case.Status?) {
        _caseStatusFilter.value = status
    }

    fun getFilteredCases(): List<Case> {
        val status = _caseStatusFilter.value
        return if (status == null) {
            _caseList.value
        } else {
            _caseList.value.filter { it.status == status }
        }
    }

    fun setSelectedCase(case: Case?) {
        _selectedCase.value = case
    }

    fun findCase(caseNumber: String): Case? {
        return _caseList.value.find { it.caseNumber == caseNumber }
            ?: _selectedCase.value?.takeIf { it.caseNumber == caseNumber }
    }

    fun setAIConversationId(id: String?) {
        _aiConversationId.value = id
    }

    fun showLoading() {
        _isLoading.value = true
    }

    fun hideLoading() {
        _isLoading.value = false
    }

    fun <T> withLoading(block: suspend () -> T): suspend T? {
        return try {
            showLoading()
            block()
        } catch (e: Exception) {
            showToast("操作失败: ${e.message}")
            null
        } finally {
            hideLoading()
        }
    }

    fun launchWithLoading(block: suspend () -> Unit) {
        appScope.launch {
            withLoading { block() }
        }
    }

    fun showToast(message: String, durationMs: Long = 2000) {
        _toastMessage.value = message
        appScope.launch {
            kotlinx.coroutines.delay(durationMs)
            if (_toastMessage.value == message) {
                _toastMessage.value = null
            }
        }
    }

    fun clearToast() {
        _toastMessage.value = null
    }

    fun isLoggedIn(): Boolean = _currentUser.value != null

    fun logout() {
        _currentUser.value = null
        _caseList.value = emptyList()
        _selectedCase.value = null
        _aiConversationId.value = null
        _caseStatusFilter.value = null
    }
}

@Composable
fun LoadingOverlay() {
    Box(
        modifier = Modifier
            .fillMaxSize()
            .background(Color.Black.copy(alpha = 0.4f)),
        contentAlignment = Alignment.Center
    ) {
        androidx.compose.foundation.layout.Column(
            horizontalAlignment = Alignment.CenterHorizontally,
            modifier = Modifier
                .background(Color.White, RoundedCornerShape(16.dp))
                .padding(32.dp)
        ) {
            CircularProgressIndicator()
            Text(
                text = "加载中...",
                fontSize = 14.sp,
                modifier = Modifier.padding(top = 16.dp)
            )
        }
    }
}

@Composable
fun Toast(message: String, onDismiss: () -> Unit) {
    Box(
        modifier = Modifier
            .fillMaxSize()
            .padding(bottom = 100.dp),
        contentAlignment = Alignment.BottomCenter
    ) {
        androidx.compose.foundation.layout.Column(
            horizontalAlignment = Alignment.CenterHorizontally,
            modifier = Modifier
                .background(Color(0xCC000000.toInt()), RoundedCornerShape(24.dp))
                .padding(horizontal = 24.dp, vertical = 12.dp)
        ) {
            Text(
                text = message,
                color = Color.White,
                fontSize = 14.sp
            )
        }
    }
}
