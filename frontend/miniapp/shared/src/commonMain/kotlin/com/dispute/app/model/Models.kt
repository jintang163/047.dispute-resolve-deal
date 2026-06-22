package com.dispute.app.model

import kotlinx.serialization.Serializable

@Serializable
data class User(
    val id: String,
    val nickname: String,
    val avatar: String? = null,
    val phone: String? = null,
    val realName: String? = null,
    val idNumber: String? = null,
    val gender: String? = null,
    val email: String? = null,
    val address: String? = null,
    val wechatOpenId: String? = null,
    val isVerified: Boolean = false,
    val createdAt: String? = null
) {
    val displayName: String
        get() = nickname.ifBlank { realName ?: "用户$id" }

    val maskedPhone: String?
        get() = phone?.let {
            if (it.length >= 11) it.substring(0, 3) + "****" + it.substring(7)
            else it
        }

    val maskedIdNumber: String?
        get() = idNumber?.let {
            if (it.length >= 14) it.substring(0, 6) + "********" + it.substring(14)
            else it
        }
}

@Serializable
data class DisputeType(
    val id: String,
    val name: String,
    val icon: String? = null,
    val children: List<DisputeType>? = null
) {
    val hasChildren: Boolean get() = !children.isNullOrEmpty()
}

@Serializable
data class Case(
    val id: String,
    val caseNumber: String,
    val userId: String,
    val applicantName: String,
    val applicantPhone: String,
    val disputeTypePath: List<String> = emptyList(),
    val disputeTypeName: String,
    val opponentName: String,
    val opponentPhone: String? = null,
    val description: String,
    val expectedResolution: String? = null,
    val status: Status,
    val statusText: String,
    val priority: Priority = Priority.NORMAL,
    val mediatorId: String? = null,
    val mediatorName: String? = null,
    val mediatorPhone: String? = null,
    val evidenceList: List<Evidence> = emptyList(),
    val progressList: List<MediationProgress> = emptyList(),
    val satisfactionRating: Int? = null,
    val satisfactionComment: String? = null,
    val submitTime: String,
    val lastUpdateTime: String? = null,
    val estimatedDays: Int? = null,
    val receiptUrl: String? = null
) {
    enum class Status(val displayName: String, val color: Long) {
        PENDING_REVIEW("待审核", 0xFFF59E0B),
        REVIEWING("审核中", 0xFFF59E0B),
        ASSIGNED("已分配调解员", 0xFF6366F1),
        MEDIATING("调解中", 0xFF1D6CFF),
        PENDING_CONFIRMATION("待确认结果", 0xFF6366F1),
        SUCCESSFUL("调解成功", 0xFF22C55E),
        UNSUCCESSFUL("调解未达成", 0xFFEF4444),
        CLOSED("已结案", 0xFF9CA3AF),
        CANCELLED("已撤销", 0xFF9CA3AF)
    }

    enum class Priority(val displayName: String) {
        LOW("普通"),
        NORMAL("一般"),
        HIGH("紧急"),
        URGENT("特急")
    }

    val fullTypeName: String
        get() = if (disputeTypePath.isNotEmpty()) disputeTypePath.joinToString(" > ") else disputeTypeName

    val statusColor: Long
        get() = status.color
}

@Serializable
data class Evidence(
    val id: String,
    val name: String,
    val type: EvidenceType,
    val url: String,
    val size: Long = 0,
    val uploadTime: String,
    val description: String? = null
) {
    enum class EvidenceType(val displayName: String, val icon: String) {
        IMAGE("图片", "🖼️"),
        DOCUMENT("文档", "📄"),
        VIDEO("视频", "🎬"),
        AUDIO("音频", "🎵"),
        OTHER("其他", "📁")
    }

    val displaySize: String
        get() = when {
            size < 1024 -> "$size B"
            size < 1024 * 1024 -> "${(size / 1024).toInt()} KB"
            else -> "${(size / (1024 * 1024)).toInt()} MB"
        }
}

@Serializable
data class MediationProgress(
    val id: String,
    val caseNumber: String,
    val stage: ProgressStage,
    val title: String,
    val description: String? = null,
    val operatorName: String? = null,
    val operatorRole: String? = null,
    val timestamp: String,
    val attachments: List<String> = emptyList()
) {
    enum class ProgressStage(val displayName: String, val order: Int) {
        SUBMITTED("提交申请", 1),
        ACCEPTED("受理登记", 2),
        ASSIGNED("分配调解员", 3),
        CONTACT_PARTIES("联系当事人", 4),
        INVESTIGATION("调查取证", 5),
        MEDIATION_MEETING("调解会议", 6),
        AGREEMENT_DRAFT("拟定协议", 7),
        SIGNING("签订协议", 8),
        FOLLOW_UP("后续跟进", 9),
        CLOSED("结案归档", 10);
    }

    val isCompleted: Boolean = true
}

@Serializable
data class AIMessage(
    val id: String,
    val role: Role,
    val content: String,
    val timestamp: String,
    val conversationId: String? = null,
    val suggestedQuestions: List<String>? = null
) {
    enum class Role(val value: String) {
        USER("user"),
        ASSISTANT("assistant")
    }
}

@Serializable
data class QuickQuestion(
    val id: String,
    val question: String,
    val category: String? = null
)

@Serializable
data class Category(
    val id: String,
    val name: String,
    val icon: String,
    val description: String
)

object MockData {
    val mockCases: List<Case> = listOf(
        Case(
            id = "1",
            caseNumber = "JF202412010001",
            userId = "user001",
            applicantName = "张三",
            applicantPhone = "13800138000",
            disputeTypePath = listOf("劳动争议", "工资报酬", "拖欠工资"),
            disputeTypeName = "拖欠工资",
            opponentName = "某某科技有限公司",
            opponentPhone = "010-12345678",
            description = "本人于2023年1月入职该公司，担任软件工程师职务。自2024年8月起，公司以经营困难为由，拖欠本人3个月工资共计54000元，多次协商未果。",
            expectedResolution = "支付拖欠工资并解除劳动合同补偿",
            status = Case.Status.MEDIATING,
            statusText = "调解中",
            mediatorName = "李调解",
            mediatorPhone = "13900139000",
            submitTime = "2024-12-01 10:30:00",
            lastUpdateTime = "2024-12-10 14:20:00",
            estimatedDays = 15
        ),
        Case(
            id = "2",
            caseNumber = "JF202411280002",
            userId = "user001",
            applicantName = "张三",
            applicantPhone = "13800138000",
            disputeTypePath = listOf("物业邻里", "邻里纠纷", "噪音扰民"),
            disputeTypeName = "噪音扰民",
            opponentName = "王某某（楼上邻居）",
            description = "楼上住户经常性在晚上10点后制造噪音，包括敲打声、拖拽家具等，严重影响本人正常休息，多次沟通无效。",
            status = Case.Status.PENDING_REVIEW,
            statusText = "待审核",
            submitTime = "2024-11-28 09:15:00"
        ),
        Case(
            id = "3",
            caseNumber = "JF202411150003",
            userId = "user001",
            applicantName = "张三",
            applicantPhone = "13800138000",
            disputeTypePath = listOf("消费维权", "商品质量", "退换货纠纷"),
            disputeTypeName = "退换货纠纷",
            opponentName = "某某电子产品店",
            description = "购买的手机使用一周后出现屏幕故障，商家以'人为损坏'为由拒绝退换，要求检测后处理。",
            status = Case.Status.SUCCESSFUL,
            statusText = "调解成功",
            mediatorName = "赵调解",
            submitTime = "2024-11-15 16:45:00",
            lastUpdateTime = "2024-11-25 10:00:00",
            satisfactionRating = 5
        )
    )

    val mockDisputeTypes: List<DisputeType> = listOf(
        DisputeType(
            id = "1",
            name = "劳动争议",
            icon = "💼",
            children = listOf(
                DisputeType(
                    id = "1-1",
                    name = "劳动合同",
                    children = listOf(
                        DisputeType(id = "1-1-1", name = "合同签订纠纷"),
                        DisputeType(id = "1-1-2", name = "合同解除纠纷"),
                        DisputeType(id = "1-1-3", name = "合同续签纠纷"),
                        DisputeType(id = "1-1-4", name = "试用期纠纷")
                    )
                ),
                DisputeType(
                    id = "1-2",
                    name = "工资报酬",
                    children = listOf(
                        DisputeType(id = "1-2-1", name = "拖欠工资"),
                        DisputeType(id = "1-2-2", name = "加班费争议"),
                        DisputeType(id = "1-2-3", name = "奖金提成纠纷"),
                        DisputeType(id = "1-2-4", name = "社保公积金争议")
                    )
                ),
                DisputeType(
                    id = "1-3",
                    name = "工伤赔偿",
                    children = listOf(
                        DisputeType(id = "1-3-1", name = "工伤认定争议"),
                        DisputeType(id = "1-3-2", name = "工伤赔偿纠纷")
                    )
                )
            )
        ),
        DisputeType(
            id = "2",
            name = "婚姻家庭",
            icon = "👨‍👩‍👧",
            children = listOf(
                DisputeType(
                    id = "2-1",
                    name = "婚姻关系",
                    children = listOf(
                        DisputeType(id = "2-1-1", name = "离婚纠纷"),
                        DisputeType(id = "2-1-2", name = "财产分割"),
                        DisputeType(id = "2-1-3", name = "彩礼返还")
                    )
                ),
                DisputeType(
                    id = "2-2",
                    name = "子女抚养",
                    children = listOf(
                        DisputeType(id = "2-2-1", name = "抚养权争议"),
                        DisputeType(id = "2-2-2", name = "抚养费纠纷"),
                        DisputeType(id = "2-2-3", name = "探视权纠纷")
                    )
                )
            )
        ),
        DisputeType(
            id = "3",
            name = "物业邻里",
            icon = "🏢",
            children = listOf(
                DisputeType(id = "3-1", name = "物业服务"),
                DisputeType(id = "3-2", name = "邻里纠纷")
            )
        ),
        DisputeType(
            id = "4",
            name = "消费维权",
            icon = "🛒"
        ),
        DisputeType(
            id = "5",
            name = "民间借贷",
            icon = "💳"
        ),
        DisputeType(
            id = "6",
            name = "交通事故",
            icon = "🚗"
        )
    )

    val mockProgress: List<MediationProgress> = listOf(
        MediationProgress(
            id = "p1",
            caseNumber = "JF202412010001",
            stage = MediationProgress.ProgressStage.SUBMITTED,
            title = "提交调解申请",
            description = "申请人通过自助终端提交纠纷登记",
            operatorName = "系统",
            timestamp = "2024-12-01 10:30:00"
        ),
        MediationProgress(
            id = "p2",
            caseNumber = "JF202412010001",
            stage = MediationProgress.ProgressStage.ACCEPTED,
            title = "受理登记",
            description = "案件已受理，案件编号 JF202412010001",
            operatorName = "王工作人员",
            operatorRole = "立案员",
            timestamp = "2024-12-01 15:20:00"
        ),
        MediationProgress(
            id = "p3",
            caseNumber = "JF202412010001",
            stage = MediationProgress.ProgressStage.ASSIGNED,
            title = "分配调解员",
            description = "已分配李调解同志负责本案",
            operatorName = "管理员",
            timestamp = "2024-12-02 09:15:00"
        ),
        MediationProgress(
            id = "p4",
            caseNumber = "JF202412010001",
            stage = MediationProgress.ProgressStage.CONTACT_PARTIES,
            title = "联系当事人",
            description = "调解员已与双方电话沟通，定于12月12日进行调解会议",
            operatorName = "李调解",
            operatorRole = "调解员",
            timestamp = "2024-12-03 14:00:00"
        )
    )

    val mockJudicialConfirmations: List<JudicialConfirmation> = listOf(
        JudicialConfirmation(
            id = 1,
            confirmNo = "SF000120241215000001",
            caseId = 1,
            caseNo = "JF202412010001",
            caseTitle = "张三与某某科技有限公司工资报酬纠纷",
            status = JudicialConfirmation.Status.CONFIRMED,
            statusText = "已确认",
            applicantName = "张三",
            applicantPhone = "13800138000",
            respondentName = "某某科技有限公司",
            respondentPhone = "010-12345678",
            courtId = 1,
            courtName = "某某区人民法院",
            agreementContent = "1. 被申请人于2024年12月31日前支付申请人工资54000元；\n2. 双方解除劳动关系，申请人不再主张其他权利。",
            performanceDeadline = "2024-12-31",
            confirmAmount = 54000.00,
            documentUrl = "https://example.com/doc/confirm1.pdf",
            sealTime = "2024-12-15 14:30:00",
            createTime = "2024-12-15 10:00:00",
            daysLeft = 10
        ),
        JudicialConfirmation(
            id = 2,
            confirmNo = "SF000120241210000002",
            caseId = 3,
            caseNo = "JF202411150003",
            caseTitle = "张三与某某电子产品店退换货纠纷",
            status = JudicialConfirmation.Status.REVIEWING,
            statusText = "审核中",
            applicantName = "张三",
            applicantPhone = "13800138000",
            respondentName = "某某电子产品店",
            respondentPhone = "010-87654321",
            courtId = 1,
            courtName = "某某区人民法院",
            agreementContent = "1. 被申请人为申请人更换同款手机一部；\n2. 申请人放弃其他赔偿请求。",
            performanceDeadline = "2024-12-25",
            confirmAmount = 0.00,
            courtCaseNo = "2024京0101确调00123号",
            createTime = "2024-12-10 15:20:00",
            daysLeft = 4
        )
    )

    val mockCourtOptions: List<CourtOption> = listOf(
        CourtOption(
            id = 1,
            name = "北京市朝阳区人民法院",
            address = "北京市朝阳区朝阳公园南路甲2号",
            phone = "010-85999888"
        ),
        CourtOption(
            id = 2,
            name = "北京市海淀区人民法院",
            address = "北京市海淀区丹棱街12号",
            phone = "010-62697000"
        ),
        CourtOption(
            id = 3,
            name = "北京市西城区人民法院",
            address = "北京市西城区后英房胡同1号",
            phone = "010-82299277"
        ),
        CourtOption(
            id = 4,
            name = "北京市东城区人民法院",
            address = "北京市东城区交道口东大街1号",
            phone = "010-64031381"
        ),
        CourtOption(
            id = 5,
            name = "北京市丰台区人民法院",
            address = "北京市丰台区近园路9号",
            phone = "010-83836068"
        ),
        CourtOption(
            id = 6,
            name = "北京市石景山区人民法院",
            address = "北京市石景山区阜石路169号",
            phone = "010-68899888"
        ),
        CourtOption(
            id = 7,
            name = "北京市通州区人民法院",
            address = "北京市通州区梨园北街187号",
            phone = "010-81553500"
        ),
        CourtOption(
            id = 8,
            name = "北京市大兴区人民法院",
            address = "北京市大兴区黄村镇金星西路8号",
            phone = "010-57362870"
        )
    )
}

@Serializable
data class JudicialConfirmation(
    val id: Long,
    val confirmNo: String,
    val caseId: Long,
    val caseNo: String,
    val caseTitle: String,
    val status: Status,
    val statusText: String,
    val applicantName: String,
    val applicantPhone: String,
    val respondentName: String,
    val respondentPhone: String,
    val courtId: Long,
    val courtName: String,
    val agreementContent: String,
    val performanceDeadline: String? = null,
    val confirmAmount: Double? = null,
    val courtCaseNo: String? = null,
    val documentNo: String? = null,
    val documentUrl: String? = null,
    val sealTime: String? = null,
    val performanceRemindTime: String? = null,
    val expirationRemindTime: String? = null,
    val fulfilledTime: String? = null,
    val remark: String? = null,
    val createTime: String,
    val updateTime: String? = null,
    val daysLeft: Int? = null
) {
    enum class Status(val displayName: String, val color: Long) {
        SUBMITTED("已提交", 0xFF9CA3AF),
        REVIEWING("审核中", 0xFFF59E0B),
        CONFIRMED("已确认", 0xFF22C55E),
        REJECTED("已驳回", 0xFFEF4444),
        EXPIRED("已失效", 0xFFF97316)
    }

    val statusColor: Long
        get() = status.color

    val isExpired: Boolean
        get() = daysLeft != null && daysLeft <= 0

    val isWarning: Boolean
        get() = daysLeft != null && daysLeft in 1..7
}

@Serializable
data class JudicialConfirmLog(
    val id: Long,
    val confirmId: Long,
    val confirmNo: String,
    val actionType: Int,
    val actionTypeName: String,
    val operatorId: Long? = null,
    val operatorName: String? = null,
    val operatorType: Int,
    val operatorTypeName: String,
    val remark: String? = null,
    val detail: String? = null,
    val createTime: String
)

@Serializable
data class CourtOption(
    val id: Long,
    val courtName: String
)

@Serializable
data class CreateJudicialRequest(
    val caseId: Long,
    val caseNo: String? = null,
    val caseTitle: String? = null,
    val mediationRecordId: Long? = null,
    val protocolId: Long? = null,
    val applicantName: String,
    val applicantPhone: String,
    val applicantIdCard: String? = null,
    val applicantAddress: String? = null,
    val respondentName: String,
    val respondentPhone: String,
    val respondentIdCard: String? = null,
    val respondentAddress: String? = null,
    val courtId: Long,
    val courtName: String? = null,
    val agreementContent: String,
    val performanceDeadline: String? = null,
    val confirmAmount: Double? = null,
    val remark: String? = null
)

@Serializable
data class CreateJudicialResponse(
    val confirmNo: String
)

@Serializable
data class CourtOption(
    val id: Long,
    val name: String,
    val address: String? = null,
    val phone: String? = null
)

@Serializable
data class ReceiptQRCodeResult(
    val caseNo: String,
    val token: String,
    val qrCodeUrl: String,
    val miniAppUrl: String? = null,
    val expiredAt: String
)

@Serializable
data class GridWorker(
    val id: String,
    val userId: String,
    val realName: String,
    val phone: String,
    val gridCode: String,
    val gridName: String,
    val area: String? = null,
    val avatar: String? = null,
    val level: String = "初级网格员",
    val points: Int = 0,
    val totalTasks: Int = 0,
    val completedTasks: Int = 0,
    val joinDate: String? = null
)

@Serializable
data class GridTask(
    val id: String,
    val taskNo: String,
    val title: String,
    val description: String,
    val type: TaskType,
    val status: TaskStatus,
    val priority: TaskPriority,
    val gridCode: String,
    val gridName: String,
    val assignedTo: String? = null,
    val assignedName: String? = null,
    val pointList: List<TaskPoint> = emptyList(),
    val deadline: String? = null,
    val expectedPoints: Int = 10,
    val createTime: String,
    val updateTime: String? = null,
    val remark: String? = null
) {
    enum class TaskType(val displayName: String) {
        PATROL("日常巡逻"),
        DISPUTE("纠纷调解"),
        INSPECTION("安全检查"),
        VISIT("入户走访"),
        HAZARD("隐患排查"),
        PUBLICITY("政策宣传"),
        OTHER("其他任务")
    }

    enum class TaskStatus(val displayName: String, val color: Long) {
        PENDING("待执行", 0xFFF59E0B),
        IN_PROGRESS("进行中", 0xFF1D6CFF),
        COMPLETED("已完成", 0xFF22C55E),
        CANCELLED("已取消", 0xFF9CA3AF)
    }

    enum class TaskPriority(val displayName: String, val color: Long) {
        LOW("普通", 0xFF9CA3AF),
        NORMAL("一般", 0xFF1D6CFF),
        HIGH("紧急", 0xFFF59E0B),
        URGENT("特急", 0xFFEF4444)
    }

    val statusColor: Long
        get() = status.color

    val priorityColor: Long
        get() = priority.color
}

@Serializable
data class TaskPoint(
    val id: String,
    val taskId: String,
    val name: String,
    val address: String,
    val longitude: Double,
    val latitude: Double,
    val sortOrder: Int = 0,
    val checkInStatus: CheckInStatus = CheckInStatus.PENDING,
    val checkInTime: String? = null,
    val checkInPhoto: String? = null,
    val remark: String? = null
) {
    enum class CheckInStatus(val displayName: String) {
        PENDING("未签到"),
        CHECKED_IN("已签到"),
        SKIPPED("已跳过")
    }
}

@Serializable
data class CheckInRecord(
    val id: String,
    val taskId: String,
    val pointId: String,
    val gridWorkerId: String,
    val checkInTime: String,
    val longitude: Double,
    val latitude: Double,
    val address: String,
    val photoUrl: String? = null,
    val livenessVerified: Boolean = false,
    val remark: String? = null
)

@Serializable
data class VisitRecord(
    val id: String,
    val visitNo: String,
    val gridWorkerId: String,
    val residentName: String,
    val residentPhone: String? = null,
    val residentAddress: String,
    val visitType: VisitType,
    val visitContent: String,
    val visitResult: String? = null,
    val photoUrls: List<String> = emptyList(),
    val longitude: Double? = null,
    val latitude: Double? = null,
    val createTime: String,
    val updateTime: String? = null
) {
    enum class VisitType(val displayName: String) {
        REGULAR("常规走访"),
        DISPUTE("纠纷回访"),
        HELP("帮扶走访"),
        SPECIAL("特殊群体走访"),
        INVESTIGATION("问卷调查"),
        OTHER("其他")
    }
}

@Serializable
data class HazardReport(
    val id: String,
    val reportNo: String,
    val reporterId: String,
    val reporterName: String,
    val type: HazardType,
    val level: HazardLevel,
    val title: String,
    val description: String,
    val address: String,
    val longitude: Double? = null,
    val latitude: Double? = null,
    val photoUrls: List<String> = emptyList(),
    val status: ReportStatus = ReportStatus.PENDING,
    val handlerId: String? = null,
    val handlerName: String? = null,
    val handleResult: String? = null,
    val handleTime: String? = null,
    val createTime: String,
    val updateTime: String? = null
) {
    enum class HazardType(val displayName: String) {
        FIRE("消防安全"),
        TRAFFIC("交通安全"),
        PUBLIC("公共安全"),
        ENVIRONMENT("环境卫生"),
        FACILITY("设施损坏"),
        DISPUTE("矛盾纠纷"),
        OTHER("其他隐患")
    }

    enum class HazardLevel(val displayName: String, val color: Long) {
        LOW("一般", 0xFF22C55E),
        MEDIUM("较大", 0xFFF59E0B),
        HIGH("重大", 0xFFEF4444),
        EXTREME("特别重大", 0xFF7C0000)
    }

    enum class ReportStatus(val displayName: String) {
        PENDING("待处理"),
        PROCESSING("处理中"),
        RESOLVED("已解决"),
        CLOSED("已关闭")
    }
}

@Serializable
data class PointRecord(
    val id: String,
    val gridWorkerId: String,
    val type: PointType,
    val amount: Int,
    val balance: Int,
    val description: String,
    val relatedId: String? = null,
    val relatedType: String? = null,
    val createTime: String
) {
    enum class PointType(val displayName: String, val isIncome: Boolean) {
        TASK_COMPLETE("完成任务", true),
        CHECK_IN("签到奖励", true),
        VISIT("走访奖励", true),
        HAZARD_REPORT("隐患上报", true),
        EXCHANGE("礼品兑换", false),
        DEDUCT("积分扣除", false);

        val sign: String
            get() = if (isIncome) "+" else "-"
    }
}

@Serializable
data class PointRule(
    val id: String,
    val name: String,
    val description: String,
    val type: String,
    val points: Int,
    val maxDaily: Int? = null,
    val sortOrder: Int = 0
)

@Serializable
data class GiftCategory(
    val id: String,
    val name: String,
    val icon: String? = null,
    val sortOrder: Int = 0
)

@Serializable
data class Gift(
    val id: String,
    val name: String,
    val categoryId: String,
    val categoryName: String,
    val description: String,
    val points: Int,
    val originalPrice: Double? = null,
    val imageUrl: String? = null,
    val stock: Int = 0,
    val salesCount: Int = 0,
    val isHot: Boolean = false,
    val isNew: Boolean = false,
    val status: GiftStatus = GiftStatus.ON_SALE
) {
    enum class GiftStatus {
        ON_SALE, OFF_SALE, SOLD_OUT
    }
}

@Serializable
data class GiftExchangeRecord(
    val id: String,
    val exchangeNo: String,
    val gridWorkerId: String,
    val giftId: String,
    val giftName: String,
    val giftImage: String? = null,
    val points: Int,
    val quantity: Int = 1,
    val status: ExchangeStatus = ExchangeStatus.PENDING,
    val receiverName: String? = null,
    val receiverPhone: String? = null,
    val receiverAddress: String? = null,
    val logisticsNo: String? = null,
    val logisticsCompany: String? = null,
    val createTime: String,
    val updateTime: String? = null
) {
    enum class ExchangeStatus(val displayName: String) {
        PENDING("待发货"),
        SHIPPED("已发货"),
        DELIVERED("已收货"),
        CANCELLED("已取消")
    }
}

@Serializable
data class MapRoute(
    val distance: Double,
    val duration: Int,
    val startPoint: TaskPoint,
    val endPoint: TaskPoint,
    val wayPoints: List<TaskPoint> = emptyList(),
    val polyline: String? = null
)

object GridWorkerMockData {
    val mockGridWorker = GridWorker(
        id = "gw001",
        userId = "user001",
        realName = "李网格",
        phone = "13900139000",
        gridCode = "GRID001",
        gridName = "阳光社区第一网格",
        area = "朝阳区",
        level = "中级网格员",
        points = 2580,
        totalTasks = 156,
        completedTasks = 142,
        joinDate = "2023-01-15"
    )

    val mockTasks: List<GridTask> = listOf(
        GridTask(
            id = "task001",
            taskNo = "RW202412010001",
            title = "12月第一周日常巡逻",
            description = "对网格内的重点区域进行日常巡逻，检查安全隐患",
            type = GridTask.TaskType.PATROL,
            status = GridTask.TaskStatus.PENDING,
            priority = GridTask.TaskPriority.NORMAL,
            gridCode = "GRID001",
            gridName = "阳光社区第一网格",
            expectedPoints = 20,
            deadline = "2024-12-07 18:00:00",
            createTime = "2024-12-01 09:00:00",
            pointList = listOf(
                TaskPoint(
                    id = "p001",
                    taskId = "task001",
                    name = "阳光小区北门",
                    address = "北京市朝阳区阳光路1号",
                    longitude = 116.4074,
                    latitude = 39.9042,
                    sortOrder = 1
                ),
                TaskPoint(
                    id = "p002",
                    taskId = "task001",
                    name = "阳光小区中心花园",
                    address = "北京市朝阳区阳光路1号院内",
                    longitude = 116.4084,
                    latitude = 39.9052,
                    sortOrder = 2
                ),
                TaskPoint(
                    id = "p003",
                    taskId = "task001",
                    name = "阳光小区南门",
                    address = "北京市朝阳区阳光路1号南门",
                    longitude = 116.4094,
                    latitude = 39.9062,
                    sortOrder = 3
                )
            )
        ),
        GridTask(
            id = "task002",
            taskNo = "RW202412010002",
            title = "张三与李四邻里纠纷调解",
            description = "阳光小区3号楼2单元501和502住户因噪音问题产生纠纷，需要入户调解",
            type = GridTask.TaskType.DISPUTE,
            status = GridTask.TaskStatus.IN_PROGRESS,
            priority = GridTask.TaskPriority.HIGH,
            gridCode = "GRID001",
            gridName = "阳光社区第一网格",
            assignedTo = "gw001",
            assignedName = "李网格",
            expectedPoints = 50,
            deadline = "2024-12-05 18:00:00",
            createTime = "2024-12-02 10:30:00",
            updateTime = "2024-12-02 14:00:00",
            pointList = listOf(
                TaskPoint(
                    id = "p004",
                    taskId = "task002",
                    name = "阳光小区3号楼2单元501",
                    address = "北京市朝阳区阳光路1号3号楼2单元501",
                    longitude = 116.4064,
                    latitude = 39.9032,
                    sortOrder = 1,
                    checkInStatus = TaskPoint.CheckInStatus.CHECKED_IN,
                    checkInTime = "2024-12-02 15:00:00"
                ),
                TaskPoint(
                    id = "p005",
                    taskId = "task002",
                    name = "阳光小区3号楼2单元502",
                    address = "北京市朝阳区阳光路1号3号楼2单元502",
                    longitude = 116.4065,
                    latitude = 39.9033,
                    sortOrder = 2
                )
            )
        ),
        GridTask(
            id = "task003",
            taskNo = "RW202411280003",
            title = "冬季消防安全检查",
            description = "对网格内商户进行消防安全专项检查",
            type = GridTask.TaskType.INSPECTION,
            status = GridTask.TaskStatus.COMPLETED,
            priority = GridTask.TaskPriority.HIGH,
            gridCode = "GRID001",
            gridName = "阳光社区第一网格",
            assignedTo = "gw001",
            assignedName = "李网格",
            expectedPoints = 30,
            createTime = "2024-11-28 09:00:00",
            updateTime = "2024-11-29 17:30:00",
            pointList = listOf(
                TaskPoint(
                    id = "p006",
                    taskId = "task003",
                    name = "阳光便利店",
                    address = "北京市朝阳区阳光路2号",
                    longitude = 116.4054,
                    latitude = 39.9022,
                    sortOrder = 1,
                    checkInStatus = TaskPoint.CheckInStatus.CHECKED_IN
                ),
                TaskPoint(
                    id = "p007",
                    taskId = "task003",
                    name = "阳光餐厅",
                    address = "北京市朝阳区阳光路3号",
                    longitude = 116.4044,
                    latitude = 39.9012,
                    sortOrder = 2,
                    checkInStatus = TaskPoint.CheckInStatus.CHECKED_IN
                )
            )
        ),
        GridTask(
            id = "task004",
            taskNo = "RW202412030004",
            title = "独居老人入户走访",
            description = "对网格内5户独居老人进行走访慰问，了解生活状况",
            type = GridTask.TaskType.VISIT,
            status = GridTask.TaskStatus.PENDING,
            priority = GridTask.TaskPriority.NORMAL,
            gridCode = "GRID001",
            gridName = "阳光社区第一网格",
            expectedPoints = 25,
            deadline = "2024-12-10 18:00:00",
            createTime = "2024-12-03 08:30:00",
            pointList = listOf(
                TaskPoint(
                    id = "p008",
                    taskId = "task004",
                    name = "王奶奶家",
                    address = "北京市朝阳区阳光路1号1号楼1单元101",
                    longitude = 116.4034,
                    latitude = 39.9002,
                    sortOrder = 1
                ),
                TaskPoint(
                    id = "p009",
                    taskId = "task004",
                    name = "李爷爷家",
                    address = "北京市朝阳区阳光路1号2号楼3单元202",
                    longitude = 116.4024,
                    latitude = 39.8992,
                    sortOrder = 2
                )
            )
        )
    )

    val mockVisitRecords: List<VisitRecord> = listOf(
        VisitRecord(
            id = "visit001",
            visitNo = "ZF202412010001",
            gridWorkerId = "gw001",
            residentName = "王奶奶",
            residentPhone = "13800138001",
            residentAddress = "阳光小区1号楼1单元101",
            visitType = VisitRecord.VisitType.HELP,
            visitContent = "了解老人生活情况，帮助购买生活用品",
            visitResult = "老人身体状况良好，已帮助购买米、面、油等生活用品",
            createTime = "2024-12-01 10:30:00"
        ),
        VisitRecord(
            id = "visit002",
            visitNo = "ZF202411280002",
            gridWorkerId = "gw001",
            residentName = "张三",
            residentAddress = "阳光小区3号楼2单元501",
            visitType = VisitRecord.VisitType.DISPUTE,
            visitContent = "噪音纠纷调解后回访",
            visitResult = "双方已达成谅解，邻里关系恢复正常",
            createTime = "2024-11-28 15:00:00"
        ),
        VisitRecord(
            id = "visit003",
            visitNo = "ZF202411250003",
            gridWorkerId = "gw001",
            residentName = "赵大爷",
            residentAddress = "阳光小区5号楼1单元303",
            visitType = VisitRecord.VisitType.SPECIAL,
            visitContent = "低保户季度走访",
            createTime = "2024-11-25 14:00:00"
        )
    )

    val mockHazardReports: List<HazardReport> = listOf(
        HazardReport(
            id = "hazard001",
            reportNo = "YH202412020001",
            reporterId = "gw001",
            reporterName = "李网格",
            type = HazardReport.HazardType.FIRE,
            level = HazardReport.HazardLevel.MEDIUM,
            title = "阳光小区2号楼消防通道堵塞",
            description = "2号楼单元门口堆放大量杂物，堵塞消防通道，存在严重消防安全隐患",
            address = "北京市朝阳区阳光路1号2号楼",
            longitude = 116.4084,
            latitude = 39.9042,
            status = HazardReport.ReportStatus.PROCESSING,
            handlerId = "admin001",
            handlerName = "物业王经理",
            handleResult = "已通知物业清理，预计3日内完成",
            handleTime = "2024-12-02 16:00:00",
            createTime = "2024-12-02 14:30:00"
        ),
        HazardReport(
            id = "hazard002",
            reportNo = "YH202411300002",
            reporterId = "gw001",
            reporterName = "李网格",
            type = HazardReport.HazardType.FACILITY,
            level = HazardReport.HazardLevel.LOW,
            title = "健身器材损坏",
            description = "小区中心花园的健身器材跑步机出现故障，存在安全隐患",
            address = "北京市朝阳区阳光路1号中心花园",
            status = HazardReport.ReportStatus.RESOLVED,
            handlerId = "admin002",
            handlerName = "社区刘主任",
            handleResult = "已联系维修人员修复完成",
            handleTime = "2024-12-01 10:00:00",
            createTime = "2024-11-30 09:00:00"
        )
    )

    val mockPointRecords: List<PointRecord> = listOf(
        PointRecord(
            id = "pt001",
            gridWorkerId = "gw001",
            type = PointRecord.PointType.TASK_COMPLETE,
            amount = 30,
            balance = 2580,
            description = "完成任务：冬季消防安全检查",
            relatedId = "task003",
            relatedType = "TASK",
            createTime = "2024-11-29 17:30:00"
        ),
        PointRecord(
            id = "pt002",
            gridWorkerId = "gw001",
            type = PointRecord.PointType.CHECK_IN,
            amount = 5,
            balance = 2550,
            description = "签到奖励：阳光小区北门",
            relatedId = "p006",
            relatedType = "CHECKIN",
            createTime = "2024-11-29 09:00:00"
        ),
        PointRecord(
            id = "pt003",
            gridWorkerId = "gw001",
            type = PointRecord.PointType.HAZARD_REPORT,
            amount = 20,
            balance = 2545,
            description = "隐患上报：健身器材损坏",
            relatedId = "hazard002",
            relatedType = "HAZARD",
            createTime = "2024-11-30 09:00:00"
        ),
        PointRecord(
            id = "pt004",
            gridWorkerId = "gw001",
            type = PointRecord.PointType.EXCHANGE,
            amount = 100,
            balance = 2525,
            description = "礼品兑换：5kg大米一袋",
            relatedId = "exchange001",
            relatedType = "EXCHANGE",
            createTime = "2024-11-20 14:00:00"
        )
    )

    val mockPointRules: List<PointRule> = listOf(
        PointRule(id = "rule001", name = "完成日常任务", description = "每完成一个日常巡逻任务", type = "TASK", points = 20, sortOrder = 1),
        PointRule(id = "rule002", name = "完成调解任务", description = "每完成一个纠纷调解任务", type = "TASK", points = 50, sortOrder = 2),
        PointRule(id = "rule003", name = "完成检查任务", description = "每完成一个安全检查任务", type = "TASK", points = 30, sortOrder = 3),
        PointRule(id = "rule004", name = "点位签到", description = "每个点位签到", type = "CHECKIN", points = 5, maxDaily = 20, sortOrder = 4),
        PointRule(id = "rule005", name = "走访记录", description = "每条有效走访记录", type = "VISIT", points = 15, sortOrder = 5),
        PointRule(id = "rule006", name = "隐患上报", description = "每条有效隐患上报", type = "HAZARD", points = 20, sortOrder = 6)
    )

    val mockGiftCategories: List<GiftCategory> = listOf(
        GiftCategory(id = "cat001", name = "粮油食品", icon = "🍚", sortOrder = 1),
        GiftCategory(id = "cat002", name = "生活用品", icon = "🧴", sortOrder = 2),
        GiftCategory(id = "cat003", name = "家居用品", icon = "🏠", sortOrder = 3),
        GiftCategory(id = "cat004", name = "电子产品", icon = "📱", sortOrder = 4),
        GiftCategory(id = "cat005", name = "优惠券", icon = "🎫", sortOrder = 5)
    )

    val mockGifts: List<Gift> = listOf(
        Gift(
            id = "gift001",
            name = "5kg大米一袋",
            categoryId = "cat001",
            categoryName = "粮油食品",
            description = "东北优质大米，5kg装",
            points = 100,
            originalPrice = 35.0,
            imageUrl = null,
            stock = 50,
            salesCount = 128,
            isHot = true,
            status = Gift.GiftStatus.ON_SALE
        ),
        Gift(
            id = "gift002",
            name = "1.8L食用油",
            categoryId = "cat001",
            categoryName = "粮油食品",
            description = "非转基因大豆油，1.8L装",
            points = 80,
            originalPrice = 28.0,
            stock = 60,
            salesCount = 95,
            status = Gift.GiftStatus.ON_SALE
        ),
        Gift(
            id = "gift003",
            name = "洗衣液2kg",
            categoryId = "cat002",
            categoryName = "生活用品",
            description = "薰衣草香洗衣液，2kg装",
            points = 60,
            originalPrice = 20.0,
            stock = 100,
            salesCount = 200,
            isNew = true,
            status = Gift.GiftStatus.ON_SALE
        ),
        Gift(
            id = "gift004",
            name = "保温杯500ml",
            categoryId = "cat003",
            categoryName = "家居用品",
            description = "304不锈钢保温杯，500ml",
            points = 150,
            originalPrice = 50.0,
            stock = 30,
            salesCount = 76,
            isHot = true,
            status = Gift.GiftStatus.ON_SALE
        ),
        Gift(
            id = "gift005",
            name = "蓝牙耳机",
            categoryId = "cat004",
            categoryName = "电子产品",
            description = "无线蓝牙耳机，续航8小时",
            points = 500,
            originalPrice = 168.0,
            stock = 10,
            salesCount = 25,
            isNew = true,
            status = Gift.GiftStatus.ON_SALE
        ),
        Gift(
            id = "gift006",
            name = "超市优惠券20元",
            categoryId = "cat005",
            categoryName = "优惠券",
            description = "华联超市20元代金券，满100可用",
            points = 30,
            originalPrice = 20.0,
            stock = 200,
            salesCount = 350,
            status = Gift.GiftStatus.ON_SALE
        )
    )

    val mockExchangeRecords: List<GiftExchangeRecord> = listOf(
        GiftExchangeRecord(
            id = "ex001",
            exchangeNo = "DH202411200001",
            gridWorkerId = "gw001",
            giftId = "gift001",
            giftName = "5kg大米一袋",
            points = 100,
            quantity = 1,
            status = GiftExchangeRecord.ExchangeStatus.DELIVERED,
            receiverName = "李网格",
            receiverPhone = "13900139000",
            receiverAddress = "北京市朝阳区阳光社区办公室",
            logisticsNo = "SF1234567890",
            logisticsCompany = "顺丰速运",
            createTime = "2024-11-20 14:00:00",
            updateTime = "2024-11-22 10:30:00"
        )
    )
}
