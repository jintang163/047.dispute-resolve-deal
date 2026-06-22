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

    suspend fun getMyTasks(status: Int? = null, page: Int = 1, pageSize: Int = 20): PageResponse<GridTask> {
        val params = buildMap<String, String?> {
            status?.let { put("status", it.toString()) }
            put("page", page.toString())
            put("pageSize", pageSize.toString())
        }
        val response: ApiResponse<PageResponse<GridTask>> = client.get("/patrol/my/tasks", params)
        return response.getOrThrow()
    }

    suspend fun getTaskList(
        status: Int? = null,
        assigneeId: Long? = null,
        taskType: Int? = null,
        priority: Int? = null,
        keyword: String? = null,
        orgId: Long? = null,
        page: Int = 1,
        pageSize: Int = 20
    ): PageResponse<GridTask> {
        val params = buildMap<String, String?> {
            status?.let { put("status", it.toString()) }
            assigneeId?.let { put("assigneeId", it.toString()) }
            taskType?.let { put("taskType", it.toString()) }
            priority?.let { put("priority", it.toString()) }
            keyword?.let { put("keyword", it) }
            orgId?.let { put("orgId", it.toString()) }
            put("page", page.toString())
            put("pageSize", pageSize.toString())
        }
        val response: ApiResponse<PageResponse<GridTask>> = client.get("/patrol/task", params)
        return response.getOrThrow()
    }

    suspend fun getTaskDetail(taskId: Long): GridTask {
        val response: ApiResponse<GridTask> = client.get("/patrol/task/$taskId")
        return response.getOrThrow()
    }

    suspend fun createTask(request: CreatePatrolTaskRequest): Long {
        val response: ApiResponse<Long> = client.post("/patrol/task", request)
        return response.getOrThrow()
    }

    suspend fun updateTask(taskId: Long, request: UpdatePatrolTaskRequest) {
        val response: ApiResponse<Unit> = client.put("/patrol/task/$taskId", request)
        if (!response.isSuccess) {
            throw ApiException.BusinessError(response.code, response.message)
        }
    }

    suspend fun deleteTask(taskId: Long) {
        val response: ApiResponse<Unit> = client.delete("/patrol/task/$taskId")
        if (!response.isSuccess) {
            throw ApiException.BusinessError(response.code, response.message)
        }
    }

    suspend fun cancelTask(taskId: Long, reason: String? = null) {
        val request = mapOf("reason" to reason)
        val response: ApiResponse<Unit> = client.post("/patrol/task/$taskId/cancel", request)
        if (!response.isSuccess) {
            throw ApiException.BusinessError(response.code, response.message)
        }
    }

    suspend fun startTask(taskId: Long) {
        val response: ApiResponse<Unit> = client.post("/patrol/task/$taskId/start", emptyMap<String, String>())
        if (!response.isSuccess) {
            throw ApiException.BusinessError(response.code, response.message)
        }
    }

    suspend fun completeTask(taskId: Long) {
        val response: ApiResponse<Unit> = client.post("/patrol/task/$taskId/complete", emptyMap<String, String>())
        if (!response.isSuccess) {
            throw ApiException.BusinessError(response.code, response.message)
        }
    }

    suspend fun getTaskPoints(taskId: Long): List<TaskPoint> {
        val response: ApiResponse<List<TaskPoint>> = client.get("/patrol/task/$taskId/points")
        return response.getOrThrow()
    }

    suspend fun planRoute(request: PlanRouteRequest): MapRoute {
        val response: ApiResponse<MapRoute> = client.post("/patrol/route/plan", request)
        return response.getOrThrow()
    }

    suspend fun checkIn(request: CheckInRequest): CheckInResponse {
        val response: ApiResponse<CheckInResponse> = client.post("/patrol/checkin", request)
        return response.getOrThrow()
    }

    suspend fun getCheckInRecords(page: Int = 1, pageSize: Int = 20): PageResponse<CheckInRecord> {
        val params = mapOf("page" to page.toString(), "pageSize" to pageSize.toString())
        val response: ApiResponse<PageResponse<CheckInRecord>> = client.get("/patrol/checkin/records", params)
        return response.getOrThrow()
    }

    suspend fun getCheckInStatistics(): Map<String, Any> {
        val response: ApiResponse<Map<String, Any>> = client.get("/patrol/checkin/statistics")
        return response.getOrThrow()
    }

    suspend fun getVisitRecords(
        memberId: Long? = null,
        status: Int? = null,
        visitType: Int? = null,
        startDate: String? = null,
        endDate: String? = null,
        page: Int = 1,
        pageSize: Int = 20
    ): PageResponse<VisitRecord> {
        val params = buildMap<String, String?> {
            memberId?.let { put("memberId", it.toString()) }
            status?.let { put("status", it.toString()) }
            visitType?.let { put("visitType", it.toString()) }
            startDate?.let { put("startDate", it) }
            endDate?.let { put("endDate", it) }
            put("page", page.toString())
            put("pageSize", pageSize.toString())
        }
        val response: ApiResponse<PageResponse<VisitRecord>> = client.get("/patrol/visit", params)
        return response.getOrThrow()
    }

    suspend fun getVisitRecordDetail(id: Long): VisitRecord {
        val response: ApiResponse<VisitRecord> = client.get("/patrol/visit/$id")
        return response.getOrThrow()
    }

    suspend fun createVisitRecord(request: CreateVisitRecordRequest): Long {
        val response: ApiResponse<Long> = client.post("/patrol/visit", request)
        return response.getOrThrow()
    }

    suspend fun updateVisitRecord(id: Long, request: UpdateVisitRecordRequest) {
        val response: ApiResponse<Unit> = client.put("/patrol/visit/$id", request)
        if (!response.isSuccess) {
            throw ApiException.BusinessError(response.code, response.message)
        }
    }

    suspend fun auditVisitRecord(id: Long, status: Int, remark: String? = null) {
        val request = mapOf("status" to status, "remark" to remark)
        val response: ApiResponse<Unit> = client.post("/patrol/visit/$id/audit", request)
        if (!response.isSuccess) {
            throw ApiException.BusinessError(response.code, response.message)
        }
    }

    suspend fun deleteVisitRecord(id: Long) {
        val response: ApiResponse<Unit> = client.delete("/patrol/visit/$id")
        if (!response.isSuccess) {
            throw ApiException.BusinessError(response.code, response.message)
        }
    }

    suspend fun getVisitStatistics(): Map<String, Any> {
        val response: ApiResponse<Map<String, Any>> = client.get("/patrol/visit/statistics")
        return response.getOrThrow()
    }

    suspend fun getDangerList(
        reporterId: Long? = null,
        status: Int? = null,
        dangerType: Int? = null,
        level: Int? = null,
        keyword: String? = null,
        page: Int = 1,
        pageSize: Int = 20
    ): PageResponse<HazardReport> {
        val params = buildMap<String, String?> {
            reporterId?.let { put("reporterId", it.toString()) }
            status?.let { put("status", it.toString()) }
            dangerType?.let { put("dangerType", it.toString()) }
            level?.let { put("level", it.toString()) }
            keyword?.let { put("keyword", it) }
            put("page", page.toString())
            put("pageSize", pageSize.toString())
        }
        val response: ApiResponse<PageResponse<HazardReport>> = client.get("/patrol/danger", params)
        return response.getOrThrow()
    }

    suspend fun getDangerDetail(id: Long): HazardReport {
        val response: ApiResponse<HazardReport> = client.get("/patrol/danger/$id")
        return response.getOrThrow()
    }

    suspend fun reportDanger(request: ReportDangerRequest): Long {
        val response: ApiResponse<Long> = client.post("/patrol/danger", request)
        return response.getOrThrow()
    }

    suspend fun handleDanger(id: Long, status: Int, result: String? = null) {
        val request = mapOf("status" to status, "result" to result)
        val response: ApiResponse<Unit> = client.post("/patrol/danger/$id/handle", request)
        if (!response.isSuccess) {
            throw ApiException.BusinessError(response.code, response.message)
        }
    }

    suspend fun getDangerStatistics(): Map<String, Any> {
        val response: ApiResponse<Map<String, Any>> = client.get("/patrol/danger/statistics")
        return response.getOrThrow()
    }

    suspend fun getMemberList(
        orgId: Long? = null,
        status: Int? = null,
        keyword: String? = null,
        page: Int = 1,
        pageSize: Int = 20
    ): PageResponse<GridWorker> {
        val params = buildMap<String, String?> {
            orgId?.let { put("orgId", it.toString()) }
            status?.let { put("status", it.toString()) }
            keyword?.let { put("keyword", it) }
            put("page", page.toString())
            put("pageSize", pageSize.toString())
        }
        val response: ApiResponse<PageResponse<GridWorker>> = client.get("/patrol/member", params)
        return response.getOrThrow()
    }

    suspend fun getMemberDetail(id: Long): GridWorker {
        val response: ApiResponse<GridWorker> = client.get("/patrol/member/$id")
        return response.getOrThrow()
    }

    suspend fun getCurrentMember(): GridWorker? {
        val response: ApiResponse<GridWorker?> = client.get("/patrol/member/me")
        return if (response.isSuccess) response.data else null
    }

    suspend fun createMember(request: CreateMemberRequest): Long {
        val response: ApiResponse<Long> = client.post("/patrol/member", request)
        return response.getOrThrow()
    }

    suspend fun updateMember(id: Long, request: UpdateMemberRequest) {
        val response: ApiResponse<Unit> = client.put("/patrol/member/$id", request)
        if (!response.isSuccess) {
            throw ApiException.BusinessError(response.code, response.message)
        }
    }

    suspend fun deleteMember(id: Long) {
        val response: ApiResponse<Unit> = client.delete("/patrol/member/$id")
        if (!response.isSuccess) {
            throw ApiException.BusinessError(response.code, response.message)
        }
    }

    suspend fun getPointsSummary(): Map<String, Any> {
        val response: ApiResponse<Map<String, Any>> = client.get("/points/summary")
        return response.getOrThrow()
    }

    suspend fun addPoints(request: PointsOperationRequest) {
        val response: ApiResponse<Unit> = client.post("/points/add", request)
        if (!response.isSuccess) {
            throw ApiException.BusinessError(response.code, response.message)
        }
    }

    suspend fun deductPoints(request: PointsOperationRequest) {
        val response: ApiResponse<Unit> = client.post("/points/deduct", request)
        if (!response.isSuccess) {
            throw ApiException.BusinessError(response.code, response.message)
        }
    }

    suspend fun getPointsRecords(page: Int = 1, pageSize: Int = 20): PageResponse<PointRecord> {
        val params = mapOf("page" to page.toString(), "pageSize" to pageSize.toString())
        val response: ApiResponse<PageResponse<PointRecord>> = client.get("/points/records", params)
        return response.getOrThrow()
    }

    suspend fun getPointsRules(ruleType: String? = null): List<PointRule> {
        val params = ruleType?.let { mapOf("ruleType" to it) } ?: emptyMap()
        val response: ApiResponse<List<PointRule>> = client.get("/points/rules", params)
        return response.getOrThrow()
    }

    suspend fun createPointsRule(request: CreatePointsRuleRequest): Long {
        val response: ApiResponse<Long> = client.post("/points/rules", request)
        return response.getOrThrow()
    }

    suspend fun updatePointsRule(id: Long, request: UpdatePointsRuleRequest) {
        val response: ApiResponse<Unit> = client.put("/points/rules/$id", request)
        if (!response.isSuccess) {
            throw ApiException.BusinessError(response.code, response.message)
        }
    }

    suspend fun deletePointsRule(id: Long) {
        val response: ApiResponse<Unit> = client.delete("/points/rules/$id")
        if (!response.isSuccess) {
            throw ApiException.BusinessError(response.code, response.message)
        }
    }

    suspend fun exchangeGift(request: ExchangeGiftRequest): Long {
        val response: ApiResponse<Long> = client.post("/points/exchange", request)
        return response.getOrThrow()
    }

    suspend fun getGiftList(
        categoryId: Long? = null,
        status: Int? = null,
        keyword: String? = null,
        isHot: Int? = null,
        isNew: Int? = null,
        minPoints: Int? = null,
        maxPoints: Int? = null,
        sortBy: String? = null,
        page: Int = 1,
        pageSize: Int = 20
    ): PageResponse<Gift> {
        val params = buildMap<String, String?> {
            categoryId?.let { put("categoryId", it.toString()) }
            status?.let { put("status", it.toString()) }
            keyword?.let { put("keyword", it) }
            isHot?.let { put("isHot", it.toString()) }
            isNew?.let { put("isNew", it.toString()) }
            minPoints?.let { put("minPoints", it.toString()) }
            maxPoints?.let { put("maxPoints", it.toString()) }
            sortBy?.let { put("sortBy", it) }
            put("page", page.toString())
            put("pageSize", pageSize.toString())
        }
        val response: ApiResponse<PageResponse<Gift>> = client.get("/gift", params)
        return response.getOrThrow()
    }

    suspend fun getGiftDetail(id: Long): Gift {
        val response: ApiResponse<Gift> = client.get("/gift/$id")
        return response.getOrThrow()
    }

    suspend fun getGiftCategories(): List<GiftCategory> {
        val response: ApiResponse<List<GiftCategory>> = client.get("/gift/categories")
        return response.getOrThrow()
    }

    suspend fun getExchangeList(
        memberId: Long? = null,
        status: Int? = null,
        giftId: Long? = null,
        startDate: String? = null,
        endDate: String? = null,
        keyword: String? = null,
        page: Int = 1,
        pageSize: Int = 20
    ): PageResponse<GiftExchangeRecord> {
        val params = buildMap<String, String?> {
            memberId?.let { put("memberId", it.toString()) }
            status?.let { put("status", it.toString()) }
            giftId?.let { put("giftId", it.toString()) }
            startDate?.let { put("startDate", it) }
            endDate?.let { put("endDate", it) }
            keyword?.let { put("keyword", it) }
            put("page", page.toString())
            put("pageSize", pageSize.toString())
        }
        val response: ApiResponse<PageResponse<GiftExchangeRecord>> = client.get("/gift/exchange", params)
        return response.getOrThrow()
    }

    suspend fun getExchangeDetail(id: Long): GiftExchangeRecord {
        val response: ApiResponse<GiftExchangeRecord> = client.get("/gift/exchange/$id")
        return response.getOrThrow()
    }

    suspend fun getMyExchanges(page: Int = 1, pageSize: Int = 20): PageResponse<GiftExchangeRecord> {
        val params = mapOf("page" to page.toString(), "pageSize" to pageSize.toString())
        val response: ApiResponse<PageResponse<GiftExchangeRecord>> = client.get("/gift/my/exchanges", params)
        return response.getOrThrow()
    }

    suspend fun auditExchange(id: Long, status: Int, remark: String? = null) {
        val request = mapOf("status" to status, "remark" to remark)
        val response: ApiResponse<Unit> = client.post("/gift/exchange/$id/audit", request)
        if (!response.isSuccess) {
            throw ApiException.BusinessError(response.code, response.message)
        }
    }

    suspend fun shipExchange(id: Long, expressCompany: String, expressNo: String) {
        val request = mapOf("expressCompany" to expressCompany, "expressNo" to expressNo)
        val response: ApiResponse<Unit> = client.post("/gift/exchange/$id/ship", request)
        if (!response.isSuccess) {
            throw ApiException.BusinessError(response.code, response.message)
        }
    }

    suspend fun receiveExchange(id: Long) {
        val response: ApiResponse<Unit> = client.post("/gift/exchange/$id/receive", emptyMap<String, String>())
        if (!response.isSuccess) {
            throw ApiException.BusinessError(response.code, response.message)
        }
    }

    suspend fun cancelExchange(id: Long, reason: String? = null) {
        val request = mapOf("reason" to reason)
        val response: ApiResponse<Unit> = client.post("/gift/exchange/$id/cancel", request)
        if (!response.isSuccess) {
            throw ApiException.BusinessError(response.code, response.message)
        }
    }

    suspend fun getGiftStatistics(): Map<String, Any> {
        val response: ApiResponse<Map<String, Any>> = client.get("/gift/statistics")
        return response.getOrThrow()
    }
}

@Serializable
data class PageResponse<T>(
    val list: List<T>,
    val total: Long,
    val page: Int,
    val pageSize: Int
)

@Serializable
data class CreatePatrolTaskRequest(
    val title: String,
    val description: String,
    val taskType: Int,
    val priority: Int,
    val assigneeId: Long,
    val startTime: String? = null,
    val endTime: String? = null,
    val orgId: Long? = null,
    val gridCodes: String? = null,
    val points: List<TaskPointRequest> = emptyList()
)

@Serializable
data class TaskPointRequest(
    val pointName: String,
    val address: String,
    val longitude: Double,
    val latitude: Double,
    val pointType: Int,
    val checkinType: Int,
    val checkinRadius: Double = 200.0,
    val requiredPhotos: Int = 3
)

@Serializable
data class UpdatePatrolTaskRequest(
    val title: String? = null,
    val description: String? = null,
    val taskType: Int? = null,
    val priority: Int? = null,
    val assigneeId: Long? = null,
    val startTime: String? = null,
    val endTime: String? = null,
    val status: Int? = null
)

@Serializable
data class PlanRouteRequest(
    val startLng: Double,
    val startLat: Double,
    val points: List<RoutePointRequest>,
    val strategy: Int = 10
)

@Serializable
data class RoutePointRequest(
    val pointName: String,
    val address: String,
    val longitude: Double,
    val latitude: Double
)

@Serializable
data class CheckInRequest(
    val taskId: String,
    val pointId: String,
    val longitude: Double,
    val latitude: Double,
    val locationAccuracy: Double? = null,
    val address: String? = null,
    val photoUrl: String? = null,
    val livePhotoUrl: String? = null,
    val workerId: String? = null,
    val livenessVerified: Boolean = false,
    val deviceInfo: String? = null,
    val remark: String? = null
) {
    @Deprecated("Use pointId instead", ReplaceWith("pointId"))
    @kotlinx.serialization.Transient
    val taskPointId: Long? = pointId.toLongOrNull()

    @kotlinx.serialization.Transient
    val isLiveVerified: Int? = if (livenessVerified) 1 else 0
}

@Serializable
data class CheckInResponse(
    val id: Long,
    val checkinNo: String,
    val isValid: Int,
    val distance: Double,
    val pointsEarned: Int,
    val isLiveVerified: Int,
    val invalidReason: String? = null,
    val checkinTime: String
)

@Serializable
data class CreateVisitRecordRequest(
    val taskId: Long? = null,
    val taskPointId: Long? = null,
    val visitType: Int,
    val visitObject: String,
    val visitContent: String,
    val visitResult: String? = null,
    val longitude: Double? = null,
    val latitude: Double? = null,
    val address: String? = null,
    val photoUrls: String? = null,
    val residentId: Long? = null,
    val disputeCaseId: Long? = null,
    val remark: String? = null
)

@Serializable
data class UpdateVisitRecordRequest(
    val visitType: Int? = null,
    val visitObject: String? = null,
    val visitContent: String? = null,
    val visitResult: String? = null,
    val address: String? = null,
    val photoUrls: String? = null,
    val remark: String? = null
)

@Serializable
data class ReportDangerRequest(
    val taskId: Long? = null,
    val taskPointId: Long? = null,
    val dangerType: Int,
    val level: Int,
    val title: String,
    val description: String,
    val longitude: Double? = null,
    val latitude: Double? = null,
    val address: String? = null,
    val photoUrls: String? = null,
    val videoUrl: String? = null,
    val involvedPerson: String? = null
)

@Serializable
data class CreateMemberRequest(
    val userId: Long,
    val realName: String,
    val phone: String? = null,
    val orgId: Long? = null,
    val gridCodes: String? = null,
    val status: Int = 1
)

@Serializable
data class UpdateMemberRequest(
    val realName: String? = null,
    val phone: String? = null,
    val orgId: Long? = null,
    val gridCodes: String? = null,
    val status: Int? = null
)

@Serializable
data class PointsOperationRequest(
    val memberId: Long,
    val points: Int,
    val businessType: String? = null,
    val businessNo: String? = null,
    val description: String? = null
)

@Serializable
data class CreatePointsRuleRequest(
    val ruleCode: String,
    val ruleName: String,
    val ruleType: String,
    val points: Int,
    val maxPointsPerDay: Int = 0,
    val maxPointsPerMonth: Int = 0,
    val isActive: Int = 1,
    val description: String? = null,
    val expireDays: Int = 365,
    val sortOrder: Int = 0
)

@Serializable
data class UpdatePointsRuleRequest(
    val ruleName: String? = null,
    val ruleType: String? = null,
    val points: Int? = null,
    val maxPointsPerDay: Int? = null,
    val maxPointsPerMonth: Int? = null,
    val isActive: Int? = null,
    val description: String? = null,
    val expireDays: Int? = null,
    val sortOrder: Int? = null
)

@Serializable
data class ExchangeGiftRequest(
    val giftId: Long,
    val quantity: Int = 1,
    val receiverName: String? = null,
    val receiverPhone: String? = null,
    val receiverAddress: String? = null,
    val remark: String? = null
)
