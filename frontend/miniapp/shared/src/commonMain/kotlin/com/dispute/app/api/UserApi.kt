package com.dispute.app.api

import com.dispute.app.model.User
import kotlinx.serialization.Serializable

class UserApi(private val client: ApiClient) {

    suspend fun loginByWechat(code: String, encryptedData: String?, iv: String?): User {
        val request = WechatLoginRequest(code, encryptedData, iv)
        val response: ApiResponse<LoginResponse> = client.post("/api/user/wechat-login", request)
        val loginResp = response.getOrThrow()
        client.setAuthToken(loginResp.token)
        return loginResp.user
    }

    suspend fun loginByPhone(phone: String, smsCode: String): User {
        val request = PhoneLoginRequest(phone, smsCode)
        val response: ApiResponse<LoginResponse> = client.post("/api/user/phone-login", request)
        val loginResp = response.getOrThrow()
        client.setAuthToken(loginResp.token)
        return loginResp.user
    }

    suspend fun sendSmsCode(phone: String) {
        val response: ApiResponse<Unit> = client.post("/api/user/send-sms", mapOf("phone" to phone))
        if (!response.isSuccess) {
            throw ApiException.BusinessError(response.code, response.message)
        }
    }

    suspend fun logout() {
        val response: ApiResponse<Unit> = client.post("/api/user/logout", emptyMap<String, String>())
        client.clearAuthToken()
        if (!response.isSuccess) {
            throw ApiException.BusinessError(response.code, response.message)
        }
    }

    suspend fun getCurrentUser(): User {
        val response: ApiResponse<User> = client.get("/api/user/profile")
        return response.getOrThrow()
    }

    suspend fun updateProfile(request: UpdateProfileRequest): User {
        val response: ApiResponse<User> = client.put("/api/user/profile", request)
        return response.getOrThrow()
    }

    suspend fun verifyIdCard(name: String, idNumber: String): Boolean {
        val request = mapOf("name" to name, "idNumber" to idNumber)
        val response: ApiResponse<Boolean> = client.post("/api/user/verify-idcard", request)
        return response.getOrThrow()
    }

    suspend fun bindPhone(phone: String, smsCode: String) {
        val request = mapOf("phone" to phone, "smsCode" to smsCode)
        val response: ApiResponse<Unit> = client.post("/api/user/bind-phone", request)
        if (!response.isSuccess) {
            throw ApiException.BusinessError(response.code, response.message)
        }
    }
}

@Serializable
data class WechatLoginRequest(
    val code: String,
    val encryptedData: String?,
    val iv: String?
)

@Serializable
data class PhoneLoginRequest(
    val phone: String,
    val smsCode: String
)

@Serializable
data class LoginResponse(
    val token: String,
    val user: User,
    val expiresIn: Long = 86400
)

@Serializable
data class UpdateProfileRequest(
    val nickname: String?,
    val avatar: String?,
    val gender: String?,
    val email: String?,
    val address: String?
)
