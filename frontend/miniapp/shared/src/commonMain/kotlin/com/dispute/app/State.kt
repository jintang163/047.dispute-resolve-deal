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
import com.dispute.app.model.Gift
import com.dispute.app.model.GiftCategory
import com.dispute.app.model.GiftExchangeRecord
import com.dispute.app.model.GridTask
import com.dispute.app.model.GridWorker
import com.dispute.app.model.HazardReport
import com.dispute.app.model.JudicialConfirmation
import com.dispute.app.model.PointRecord
import com.dispute.app.model.PointRule
import com.dispute.app.model.TaskPoint
import com.dispute.app.model.User
import com.dispute.app.model.VisitRecord
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

    private val _judicialList = mutableStateOf<List<JudicialConfirmation>>(emptyList())
    val judicialList: State<List<JudicialConfirmation>> = _judicialList

    private val _selectedJudicial = mutableStateOf<JudicialConfirmation?>(null)
    val selectedJudicial: State<JudicialConfirmation?> = _selectedJudicial

    private val _judicialStatusFilter = mutableStateOf<JudicialConfirmation.Status?>(null)
    val judicialStatusFilter: State<JudicialConfirmation.Status?> = _judicialStatusFilter

    private val _gridWorker = mutableStateOf<GridWorker?>(null)
    val gridWorker: State<GridWorker?> = _gridWorker

    private val _gridTaskList = mutableStateOf<List<GridTask>>(emptyList())
    val gridTaskList: State<List<GridTask>> = _gridTaskList

    private val _selectedGridTask = mutableStateOf<GridTask?>(null)
    val selectedGridTask: State<GridTask?> = _selectedGridTask

    private val _gridTaskStatusFilter = mutableStateOf<GridTask.TaskStatus?>(null)
    val gridTaskStatusFilter: State<GridTask.TaskStatus?> = _gridTaskStatusFilter

    private val _visitRecordList = mutableStateOf<List<VisitRecord>>(emptyList())
    val visitRecordList: State<List<VisitRecord>> = _visitRecordList

    private val _hazardReportList = mutableStateOf<List<HazardReport>>(emptyList())
    val hazardReportList: State<List<HazardReport>> = _hazardReportList

    private val _pointRecordList = mutableStateOf<List<PointRecord>>(emptyList())
    val pointRecordList: State<List<PointRecord>> = _pointRecordList

    private val _pointRuleList = mutableStateOf<List<PointRule>>(emptyList())
    val pointRuleList: State<List<PointRule>> = _pointRuleList

    private val _giftCategoryList = mutableStateOf<List<GiftCategory>>(emptyList())
    val giftCategoryList: State<List<GiftCategory>> = _giftCategoryList

    private val _giftList = mutableStateOf<List<Gift>>(emptyList())
    val giftList: State<List<Gift>> = _giftList

    private val _selectedGift = mutableStateOf<Gift?>(null)
    val selectedGift: State<Gift?> = _selectedGift

    private val _giftCategoryFilter = mutableStateOf<String?>(null)
    val giftCategoryFilter: State<String?> = _giftCategoryFilter

    private val _exchangeRecordList = mutableStateOf<List<GiftExchangeRecord>>(emptyList())
    val exchangeRecordList: State<List<GiftExchangeRecord>> = _exchangeRecordList

    private val _checkInPoint = mutableStateOf<TaskPoint?>(null)
    val checkInPoint: State<TaskPoint?> = _checkInPoint

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

    fun setGridWorker(worker: GridWorker?) {
        _gridWorker.value = worker
    }

    fun setGridTaskList(tasks: List<GridTask>) {
        _gridTaskList.value = tasks
    }

    fun addGridTask(task: GridTask) {
        _gridTaskList.value = listOf(task) + _gridTaskList.value
    }

    fun updateGridTask(taskId: String, update: (GridTask) -> GridTask) {
        _gridTaskList.value = _gridTaskList.value.map {
            if (it.id == taskId) update(it) else it
        }
        _selectedGridTask.value?.let {
            if (it.id == taskId) {
                _selectedGridTask.value = update(it)
            }
        }
    }

    fun setSelectedGridTask(task: GridTask?) {
        _selectedGridTask.value = task
    }

    fun setGridTaskStatusFilter(status: GridTask.TaskStatus?) {
        _gridTaskStatusFilter.value = status
    }

    fun findGridTask(taskId: String): GridTask? {
        return _gridTaskList.value.find { it.id == taskId }
            ?: _selectedGridTask.value?.takeIf { it.id == taskId }
    }

    fun setVisitRecordList(records: List<VisitRecord>) {
        _visitRecordList.value = records
    }

    fun addVisitRecord(record: VisitRecord) {
        _visitRecordList.value = listOf(record) + _visitRecordList.value
    }

    fun setHazardReportList(reports: List<HazardReport>) {
        _hazardReportList.value = reports
    }

    fun addHazardReport(report: HazardReport) {
        _hazardReportList.value = listOf(report) + _hazardReportList.value
    }

    fun setPointRecordList(records: List<PointRecord>) {
        _pointRecordList.value = records
    }

    fun setPointRuleList(rules: List<PointRule>) {
        _pointRuleList.value = rules
    }

    fun setGiftCategoryList(categories: List<GiftCategory>) {
        _giftCategoryList.value = categories
    }

    fun setGiftList(gifts: List<Gift>) {
        _giftList.value = gifts
    }

    fun setSelectedGift(gift: Gift?) {
        _selectedGift.value = gift
    }

    fun setGiftCategoryFilter(categoryId: String?) {
        _giftCategoryFilter.value = categoryId
    }

    fun findGift(giftId: String): Gift? {
        return _giftList.value.find { it.id == giftId }
            ?: _selectedGift.value?.takeIf { it.id == giftId }
    }

    fun setExchangeRecordList(records: List<GiftExchangeRecord>) {
        _exchangeRecordList.value = records
    }

    fun addExchangeRecord(record: GiftExchangeRecord) {
        _exchangeRecordList.value = listOf(record) + _exchangeRecordList.value
    }

    fun setCheckInPoint(point: TaskPoint?) {
        _checkInPoint.value = point
    }

    fun updateCheckInPoint(pointId: String, update: (TaskPoint) -> TaskPoint) {
        _checkInPoint.value?.let {
            if (it.id == pointId) {
                _checkInPoint.value = update(it)
            }
        }
        _selectedGridTask.value?.let { task ->
            val updatedPoints = task.pointList.map { point ->
                if (point.id == pointId) update(point) else point
            }
            _selectedGridTask.value = task.copy(pointList = updatedPoints)
        }
    }

    fun isLoggedIn(): Boolean = _currentUser.value != null

    fun logout() {
        _currentUser.value = null
        _caseList.value = emptyList()
        _selectedCase.value = null
        _aiConversationId.value = null
        _caseStatusFilter.value = null
        _judicialList.value = emptyList()
        _selectedJudicial.value = null
        _judicialStatusFilter.value = null
        _gridWorker.value = null
        _gridTaskList.value = emptyList()
        _selectedGridTask.value = null
        _gridTaskStatusFilter.value = null
        _visitRecordList.value = emptyList()
        _hazardReportList.value = emptyList()
        _pointRecordList.value = emptyList()
        _pointRuleList.value = emptyList()
        _giftCategoryList.value = emptyList()
        _giftList.value = emptyList()
        _selectedGift.value = null
        _giftCategoryFilter.value = null
        _exchangeRecordList.value = emptyList()
        _checkInPoint.value = null
    }

    fun setJudicialList(list: List<JudicialConfirmation>) {
        _judicialList.value = list
    }

    fun addJudicialConfirmation(confirmation: JudicialConfirmation) {
        _judicialList.value = listOf(confirmation) + _judicialList.value
    }

    fun updateJudicialConfirmation(id: Long, update: (JudicialConfirmation) -> JudicialConfirmation) {
        _judicialList.value = _judicialList.value.map {
            if (it.id == id) update(it) else it
        }
    }

    fun setSelectedJudicial(confirmation: JudicialConfirmation?) {
        _selectedJudicial.value = confirmation
    }

    fun setJudicialStatusFilter(status: JudicialConfirmation.Status?) {
        _judicialStatusFilter.value = status
    }

    fun getFilteredJudicialList(): List<JudicialConfirmation> {
        val status = _judicialStatusFilter.value
        return if (status == null) {
            _judicialList.value
        } else {
            _judicialList.value.filter { it.status == status }
        }
    }

    fun findJudicialConfirmation(id: Long): JudicialConfirmation? {
        return _judicialList.value.find { it.id == id }
            ?: _selectedJudicial.value?.takeIf { it.id == id }
    }

    fun findJudicialConfirmationByNo(confirmNo: String): JudicialConfirmation? {
        return _judicialList.value.find { it.confirmNo == confirmNo }
            ?: _selectedJudicial.value?.takeIf { it.confirmNo == confirmNo }
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
