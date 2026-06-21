package com.dispute.app

import androidx.compose.runtime.Composable
import androidx.compose.runtime.CompositionLocalProvider
import androidx.compose.runtime.ProvidableCompositionLocal
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.setValue
import androidx.compose.runtime.staticCompositionLocalOf
import androidx.compose.ui.platform.LocalDensity
import androidx.compose.ui.text.TextStyle
import androidx.compose.ui.text.font.FontFamily
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.sp
import com.dispute.app.model.Case
import com.dispute.app.model.User
import androidx.compose.runtime.State
import androidx.compose.runtime.MutableState

sealed class Route {
    object Login : Route()
    object Home : Route()
    object RegisterCase : Route()
    object CaseList : Route()
    data class CaseDetail(val caseNumber: String) : Route()
    object Progress : Route()
    data class ScanProgress(val caseNo: String, val phone: String? = null) : Route()
    object AIConsult : Route()
    data class Satisfaction(val caseNumber: String) : Route()
    object Profile : Route()
    object JudicialList : Route()
    data class JudicialDetail(val id: Long) : Route()
    object JudicialApply : Route()
    object JudicialQuery : Route()
}

val LocalRouter: ProvidableCompositionLocal<Router> = staticCompositionLocalOf {
    error("Router not provided")
}

val LocalAppState: ProvidableCompositionLocal<AppState> = staticCompositionLocalOf {
    error("AppState not provided")
}

val LocalApiClient: ProvidableCompositionLocal<com.dispute.app.api.ApiClient> = staticCompositionLocalOf {
    error("ApiClient not provided")
}

class Router(private val appState: AppState) {
    private val _currentRoute = mutableStateOf<Route>(Route.Login)
    val currentRoute: State<Route> = _currentRoute

    private val backStack = mutableListOf<Route>()

    fun navigate(route: Route) {
        backStack.add(_currentRoute.value)
        _currentRoute.value = route
    }

    fun navigateWithReplace(route: Route) {
        backStack.clear()
        _currentRoute.value = route
    }

    fun back(): Boolean {
        if (backStack.isNotEmpty()) {
            _currentRoute.value = backStack.removeLast()
            return true
        }
        return false
    }

    fun navigateToHome() {
        backStack.clear()
        _currentRoute.value = Route.Home
    }

    fun navigateToLogin() {
        backStack.clear()
        _currentRoute.value = Route.Login
    }
}

val AppTypography = androidx.compose.material3.Typography(
    displayLarge = TextStyle(
        fontFamily = FontFamily.Default,
        fontWeight = FontWeight.Bold,
        fontSize = 36.sp
    ),
    displayMedium = TextStyle(
        fontFamily = FontFamily.Default,
        fontWeight = FontWeight.Bold,
        fontSize = 30.sp
    ),
    displaySmall = TextStyle(
        fontFamily = FontFamily.Default,
        fontWeight = FontWeight.SemiBold,
        fontSize = 24.sp
    ),
    headlineLarge = TextStyle(
        fontFamily = FontFamily.Default,
        fontWeight = FontWeight.Bold,
        fontSize = 22.sp
    ),
    headlineMedium = TextStyle(
        fontFamily = FontFamily.Default,
        fontWeight = FontWeight.SemiBold,
        fontSize = 20.sp
    ),
    titleLarge = TextStyle(
        fontFamily = FontFamily.Default,
        fontWeight = FontWeight.SemiBold,
        fontSize = 18.sp
    ),
    titleMedium = TextStyle(
        fontFamily = FontFamily.Default,
        fontWeight = FontWeight.Medium,
        fontSize = 16.sp
    ),
    bodyLarge = TextStyle(
        fontFamily = FontFamily.Default,
        fontWeight = FontWeight.Normal,
        fontSize = 16.sp
    ),
    bodyMedium = TextStyle(
        fontFamily = FontFamily.Default,
        fontWeight = FontWeight.Normal,
        fontSize = 14.sp
    ),
    labelLarge = TextStyle(
        fontFamily = FontFamily.Default,
        fontWeight = FontWeight.Medium,
        fontSize = 14.sp
    ),
    labelMedium = TextStyle(
        fontFamily = FontFamily.Default,
        fontWeight = FontWeight.Medium,
        fontSize = 12.sp
    )
)
