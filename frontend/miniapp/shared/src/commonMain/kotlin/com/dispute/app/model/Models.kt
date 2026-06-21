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
