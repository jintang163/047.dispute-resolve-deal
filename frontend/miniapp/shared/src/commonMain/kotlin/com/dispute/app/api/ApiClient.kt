package com.dispute.app.api

import io.ktor.client.HttpClient
import io.ktor.client.call.body
import io.ktor.client.engine.HttpClientEngine
import io.ktor.client.plugins.contentnegotiation.ContentNegotiation
import io.ktor.client.plugins.logging.Logger
import io.ktor.client.plugins.logging.LogLevel
import io.ktor.client.plugins.logging.Logging
import io.ktor.client.plugins.logging.SIMPLE
import io.ktor.client.request.HttpRequestBuilder
import io.ktor.client.request.delete
import io.ktor.client.request.get
import io.ktor.client.request.header
import io.ktor.client.request.post
import io.ktor.client.request.put
import io.ktor.client.request.setBody
import io.ktor.http.ContentType
import io.ktor.http.HttpHeaders
import io.ktor.http.contentType
import io.ktor.serialization.kotlinx.json.json
import kotlinx.serialization.json.Json

expect fun createHttpClientEngine(): HttpClientEngine

interface PlatformApiConfig {
    val baseUrl: String
    var authToken: String?
    var deviceId: String?
    var platform: String
}

expect object PlatformApiConfigImpl : PlatformApiConfig

class ApiClient(
    private val config: PlatformApiConfig = PlatformApiConfigImpl
) {
    private val json = Json {
        prettyPrint = true
        isLenient = true
        ignoreUnknownKeys = true
        encodeDefaults = true
    }

    val client: HttpClient = HttpClient(createHttpClientEngine()) {
        install(ContentNegotiation) {
            json(json)
        }
        install(Logging) {
            logger = Logger.SIMPLE
            level = LogLevel.INFO
        }
    }

    val dispute = DisputeApi(this)
    val user = UserApi(this)
    val ai = AIApi(this)

    fun baseUrl(): String = config.baseUrl

    fun setAuthToken(token: String?) {
        config.authToken = token
    }

    fun clearAuthToken() {
        config.authToken = null
    }

    fun getAuthToken(): String? = config.authToken

    suspend fun <T> get(
        path: String,
        parameters: Map<String, String?> = emptyMap(),
        block: HttpRequestBuilder.() -> Unit = {}
    ): T {
        val url = buildUrl(path, parameters)
        return client.get(url) {
            applyDefaults()
            block()
        }.body()
    }

    suspend fun <T, B> post(
        path: String,
        body: B,
        block: HttpRequestBuilder.() -> Unit = {}
    ): T {
        return client.post(buildUrl(path)) {
            applyDefaults()
            setBody(body)
            block()
        }.body()
    }

    suspend fun <T, B> put(
        path: String,
        body: B,
        block: HttpRequestBuilder.() -> Unit = {}
    ): T {
        return client.put(buildUrl(path)) {
            applyDefaults()
            setBody(body)
            block()
        }.body()
    }

    suspend fun <T> delete(
        path: String,
        block: HttpRequestBuilder.() -> Unit = {}
    ): T {
        return client.delete(buildUrl(path)) {
            applyDefaults()
            block()
        }.body()
    }

    private fun HttpRequestBuilder.applyDefaults() {
        contentType(ContentType.Application.Json)
        header(HttpHeaders.UserAgent, "DisputeResolveApp/1.0.0/${config.platform}")
        config.deviceId?.let { header("X-Device-ID", it) }
        config.authToken?.let { header(HttpHeaders.Authorization, "Bearer $it") }
    }

    private fun buildUrl(path: String, parameters: Map<String, String?> = emptyMap()): String {
        val base = config.baseUrl.trimEnd('/')
        val cleanPath = if (path.startsWith('/')) path else "/$path"
        val queryString = parameters
            .filterValues { it != null }
            .entries
            .joinToString(separator = "&", prefix = "?") { (k, v) ->
                "$k=${v?.let { encodeUrlParam(it) }}"
            }
        return if (parameters.isNotEmpty() && queryString != "?") {
            "$base$cleanPath$queryString"
        } else {
            "$base$cleanPath"
        }
    }

    private fun encodeUrlParam(value: String): String {
        return value
            .replace(" ", "%20")
            .replace("?", "%3F")
            .replace("&", "%26")
            .replace("=", "%3D")
            .replace("+", "%2B")
            .replace("/", "%2F")
            .replace("%", "%25")
    }

    suspend fun safeCall(block: suspend () -> Unit) {
        try {
            block()
        } catch (e: io.ktor.client.plugins.ClientRequestException) {
            throw ApiException.ClientError(e.response.status.value, e.message ?: "客户端错误")
        } catch (e: io.ktor.client.plugins.ServerResponseException) {
            throw ApiException.ServerError(e.response.status.value, e.message ?: "服务器错误")
        } catch (e: io.ktor.client.network.sockets.SocketTimeoutException) {
            throw ApiException.Timeout("请求超时，请检查网络")
        } catch (e: Exception) {
            throw ApiException.Unknown(e.message ?: "未知错误")
        }
    }
}

sealed class ApiException(message: String) : Exception(message) {
    data class ClientError(val statusCode: Int, override val message: String) : ApiException(message)
    data class ServerError(val statusCode: Int, override val message: String) : ApiException(message)
    data class Timeout(override val message: String) : ApiException(message)
    data class Unknown(override val message: String) : ApiException(message)
    data class BusinessError(val code: Int, override val message: String) : ApiException(message)
}

@kotlinx.serialization.Serializable
data class ApiResponse<T>(
    val code: Int = 0,
    val message: String = "",
    val data: T? = null
) {
    val isSuccess: Boolean get() = code == 200 || code == 0

    fun getOrThrow(): T {
        if (!isSuccess) throw ApiException.BusinessError(code, message)
        return data ?: throw ApiException.Unknown("返回数据为空")
    }

    fun getOrNull(): T? = if (isSuccess) data else null
}
