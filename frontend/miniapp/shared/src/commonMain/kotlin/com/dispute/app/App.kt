package com.dispute.app

import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Surface
import androidx.compose.material3.darkColorScheme
import androidx.compose.material3.lightColorScheme
import androidx.compose.runtime.Composable
import androidx.compose.runtime.CompositionLocalProvider
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.graphics.Color
import com.dispute.app.pages.AIConsultPage
import com.dispute.app.pages.CaseDetailPage
import com.dispute.app.pages.CaseListPage
import com.dispute.app.pages.HomePage
import com.dispute.app.pages.JudicialApplyPage
import com.dispute.app.pages.JudicialDetailPage
import com.dispute.app.pages.JudicialListPage
import com.dispute.app.pages.JudicialQueryPage
import com.dispute.app.pages.LoginPage
import com.dispute.app.pages.ProfilePage
import com.dispute.app.pages.ProgressPage
import com.dispute.app.pages.RegisterCasePage
import com.dispute.app.pages.SatisfactionPage
import com.dispute.app.api.ApiClient
import com.dispute.app.pages.GridWorkerHomePage
import com.dispute.app.pages.GridTaskListPage
import com.dispute.app.pages.GridTaskDetailPage
import com.dispute.app.pages.CheckInPage
import com.dispute.app.pages.VisitRecordListPage
import com.dispute.app.pages.VisitRecordAddPage
import com.dispute.app.pages.HazardReportPage
import com.dispute.app.pages.MapRoutePage
import com.dispute.app.pages.PointCenterPage
import com.dispute.app.pages.GiftMallPage
import com.dispute.app.pages.GiftDetailPage

val PrimaryColor = Color(0xFF1D6CFF)
val PrimaryLightColor = Color(0xFF4D8CFF)
val SuccessColor = Color(0xFF22C55E)
val WarningColor = Color(0xFFF59E0B)
val DangerColor = Color(0xFFEF4444)
val InfoColor = Color(0xFF6366F1)

object LaunchParams {
    var caseNo: String = ""
    var phone: String = ""

    fun setFromQuery(query: Map<String, String>) {
        query["caseNo"]?.let { caseNo = it }
        query["phone"]?.let { phone = it }
    }

    fun hasScanData(): Boolean = caseNo.isNotBlank()
}

val LightColors = lightColorScheme(
    primary = PrimaryColor,
    secondary = InfoColor,
    tertiary = SuccessColor,
    error = DangerColor,
    background = Color(0xFFF0F7FF),
    surface = Color(0xFFFFFFFF),
    onPrimary = Color.White,
    onSecondary = Color.White,
    onBackground = Color(0xFF1A1A1A),
    onSurface = Color(0xFF1A1A1A)
)

val DarkColors = darkColorScheme(
    primary = PrimaryLightColor,
    secondary = InfoColor,
    tertiary = SuccessColor,
    error = DangerColor,
    background = Color(0xFF0F172A),
    surface = Color(0xFF1E293B),
    onPrimary = Color.White,
    onSecondary = Color.White,
    onBackground = Color.White,
    onSurface = Color.White
)

@Composable
fun App() {
    val appState = remember { AppState() }
    val router = remember { Router(appState) }
    val apiClient = remember { ApiClient() }

    var scanHandled by remember { mutableStateOf(false) }

    LaunchedEffect(Unit) {
        if (LaunchParams.hasScanData() && !scanHandled) {
            scanHandled = true
            router.navigate(Route.ScanProgress(LaunchParams.caseNo, LaunchParams.phone))
        }
    }

    val apiClientRef = apiClient
    val appStateRef = appState
    LaunchedEffect(Unit) {
        if (appStateRef.isLoggedIn()) {
            appStateRef.loadCurrentGridWorker(apiClientRef)
        }
    }

    CompositionLocalProvider(
        LocalAppState provides appState,
        LocalRouter provides router,
        LocalApiClient provides apiClient
    ) {
        MaterialTheme(
            colorScheme = LightColors,
            typography = AppTypography
        ) {
            Surface(color = MaterialTheme.colorScheme.background) {
                AppContent(router, appState)
            }
        }
    }
}

@Composable
private fun AppContent(router: Router, appState: AppState) {
    when (val currentRoute = router.currentRoute.value) {
        is Route.Login -> LoginPage()
        is Route.Home -> HomePage()
        is Route.RegisterCase -> RegisterCasePage()
        is Route.CaseList -> CaseListPage()
        is Route.CaseDetail -> CaseDetailPage(currentRoute.caseNumber)
        is Route.Progress -> ProgressPage()
        is Route.ScanProgress -> ProgressPage(initialCaseNo = currentRoute.caseNo, initialPhone = currentRoute.phone)
        is Route.AIConsult -> AIConsultPage()
        is Route.Satisfaction -> SatisfactionPage(currentRoute.caseNumber)
        is Route.Profile -> ProfilePage()
        is Route.JudicialList -> JudicialListPage()
        is Route.JudicialDetail -> JudicialDetailPage(currentRoute.id)
        is Route.JudicialApply -> JudicialApplyPage()
        is Route.JudicialQuery -> JudicialQueryPage()
        is Route.GridWorkerHome -> GridWorkerHomePage()
        is Route.GridTaskList -> GridTaskListPage()
        is Route.GridTaskDetail -> GridTaskDetailPage(currentRoute.taskId)
        is Route.CheckIn -> CheckInPage(currentRoute.taskId, currentRoute.pointId)
        is Route.VisitRecordList -> VisitRecordListPage()
        is Route.VisitRecordAdd -> VisitRecordAddPage()
        is Route.HazardReport -> HazardReportPage()
        is Route.MapRoute -> MapRoutePage(currentRoute.taskId)
        is Route.PointCenter -> PointCenterPage()
        is Route.GiftMall -> GiftMallPage()
        is Route.GiftDetail -> GiftDetailPage(currentRoute.giftId)
    }

    if (appState.isLoading.value) {
        LoadingOverlay()
    }

    appState.toastMessage.value?.let { message ->
        Toast(message, onDismiss = { appState.clearToast() })
    }
}
