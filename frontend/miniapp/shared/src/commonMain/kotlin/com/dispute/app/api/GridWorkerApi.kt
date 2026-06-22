package com.dispute.app.api

import com.dispute.app.model.CheckInRecord
import com.dispute.app.model.Gift
import com.dispute.app.model.GiftCategory
import com.dispute.app.model.GiftExchangeRecord
import com.dispute.app.model.GridTask
import com.dispute.app.model.GridWorker
import com.dispute.app.model.HazardReport
import com.dispute.app.model.MapRoute
import com.dispute.app.model.PointRecord
import com.dispute.app.model.PointRule
import com.dispute.app.model.TaskPoint
import com.dispute.app.model.VisitRecord
import kotlinx.serialization.Serializable

class GridWorkerApi(private val client: ApiClient) {

    suspend fun getGridWorkerInfo(workerId: String): GridWorker {
        val response: ApiResponse<GridWorker> = client.get("/api/gridworker/info/$workerId")
        return response.getOrThrow()
    }

    suspend fun getTaskList(
        workerId: String,
        status: GridTask.TaskStatus? = null
    ): List<GridTask> {
        val params = buildMap {
            put("workerId", workerId)
            status?.let { put("status", it.name) }
        }
        val response: ApiResponse<List<GridTask>> = client.get("/api/gridworker/tasks", params)
        return response.getOrThrow()
    }

    suspend fun getTaskDetail(taskId: String): GridTask {
        val response: ApiResponse<GridTask> = client.get("/api/gridworker/task/$taskId")
        return response.getOrThrow()
    }

    suspend fun startTask(taskId: String, workerId: String) {
        val request = mapOf("taskId" to taskId, "workerId" to workerId)
        val response: ApiResponse<Unit> = client.post("/api/gridworker/task/start", request)
        if (!response.isSuccess) {
            throw ApiException.BusinessError(response.code, response.message)
        }
    }

    suspend fun completeTask(taskId: String, workerId: String) {
        val request = mapOf("taskId" to taskId, "workerId" to workerId)
        val response: ApiResponse<Unit> = client.post("/api/gridworker/task/complete", request)
        if (!response.isSuccess) {
            throw ApiException.BusinessError(response.code, response.message)
        }
    }

    suspend fun checkIn(request: CheckInRequest): CheckInRecord {
        val response: ApiResponse<CheckInRecord> = client.post("/api/gridworker/checkin", request)
        return response.getOrThrow()
    }

    suspend fun getTodayCheckInStatus(workerId: String): CheckInStatusResponse {
        val params = mapOf("workerId" to workerId)
        val response: ApiResponse<CheckInStatusResponse> =
            client.get("/api/gridworker/checkin/today", params)
        return response.getOrThrow()
    }

    suspend fun getVisitRecords(workerId: String): List<VisitRecord> {
        val params = mapOf("workerId" to workerId)
        val response: ApiResponse<List<VisitRecord>> =
            client.get("/api/gridworker/visits", params)
        return response.getOrThrow()
    }

    suspend fun createVisitRecord(request: CreateVisitRequest): VisitRecord {
        val response: ApiResponse<VisitRecord> =
            client.post("/api/gridworker/visit/create", request)
        return response.getOrThrow()
    }

    suspend fun getHazardReports(workerId: String): List<HazardReport> {
        val params = mapOf("workerId" to workerId)
        val response: ApiResponse<List<HazardReport>> =
            client.get("/api/gridworker/hazards", params)
        return response.getOrThrow()
    }

    suspend fun reportHazard(request: ReportHazardRequest): HazardReport {
        val response: ApiResponse<HazardReport> =
            client.post("/api/gridworker/hazard/report", request)
        return response.getOrThrow()
    }

    suspend fun planRoute(taskId: String): MapRoute {
        val params = mapOf("taskId" to taskId)
        val response: ApiResponse<MapRoute> = client.get("/api/gridworker/route/plan", params)
        return response.getOrThrow()
    }

    suspend fun getPointBalance(workerId: String): Int {
        val response: ApiResponse<Int> = client.get("/api/gridworker/points/balance/$workerId")
        return response.getOrThrow()
    }

    suspend fun getPointRecords(workerId: String): List<PointRecord> {
        val params = mapOf("workerId" to workerId)
        val response: ApiResponse<List<PointRecord>> =
            client.get("/api/gridworker/points/records", params)
        return response.getOrThrow()
    }

    suspend fun getPointRules(): List<PointRule> {
        val response: ApiResponse<List<PointRule>> = client.get("/api/gridworker/points/rules")
        return response.getOrThrow()
    }

    suspend fun getGiftCategories(): List<GiftCategory> {
        val response: ApiResponse<List<GiftCategory>> =
            client.get("/api/gridworker/gifts/categories")
        return response.getOrThrow()
    }

    suspend fun getGiftList(categoryId: String? = null): List<Gift> {
        val params = categoryId?.let { mapOf("categoryId" to it) } ?: emptyMap()
        val response: ApiResponse<List<Gift>> = client.get("/api/gridworker/gifts", params)
        return response.getOrThrow()
    }

    suspend fun getGiftDetail(giftId: String): Gift {
        val response: ApiResponse<Gift> = client.get("/api/gridworker/gift/$giftId")
        return response.getOrThrow()
    }

    suspend fun exchangeGift(request: ExchangeGiftRequest): GiftExchangeRecord {
        val response: ApiResponse<GiftExchangeRecord> =
            client.post("/api/gridworker/gift/exchange", request)
        return response.getOrThrow()
    }

    suspend fun getExchangeRecords(workerId: String): List<GiftExchangeRecord> {
        val params = mapOf("workerId" to workerId)
        val response: ApiResponse<List<GiftExchangeRecord>> =
            client.get("/api/gridworker/gifts/exchange-records", params)
        return response.getOrThrow()
    }
}

@Serializable
data class CheckInRequest(
    val taskId: String,
    val pointId: String,
    val workerId: String,
    val longitude: Double,
    val latitude: Double,
    val address: String,
    val photoBase64: String? = null,
    val livenessVerified: Boolean = false,
    val remark: String? = null
)

@Serializable
data class CheckInStatusResponse(
    val hasCheckedIn: Boolean,
    val checkInTime: String? = null,
    val checkInCount: Int = 0,
    val totalPoints: Int = 0
)

@Serializable
data class CreateVisitRequest(
    val workerId: String,
    val residentName: String,
    val residentPhone: String? = null,
    val residentAddress: String,
    val visitType: VisitRecord.VisitType,
    val visitContent: String,
    val visitResult: String? = null,
    val photoBase64List: List<String> = emptyList(),
    val longitude: Double? = null,
    val latitude: Double? = null
)

@Serializable
data class ReportHazardRequest(
    val reporterId: String,
    val reporterName: String,
    val type: HazardReport.HazardType,
    val level: HazardReport.HazardLevel,
    val title: String,
    val description: String,
    val address: String,
    val longitude: Double? = null,
    val latitude: Double? = null,
    val photoBase64List: List<String> = emptyList()
)

@Serializable
data class ExchangeGiftRequest(
    val workerId: String,
    val giftId: String,
    val quantity: Int = 1,
    val receiverName: String? = null,
    val receiverPhone: String? = null,
    val receiverAddress: String? = null
)
