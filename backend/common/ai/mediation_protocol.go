package ai

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/dispute-resolve/common/logger"
	"go.uber.org/zap"
)

const (
	MediationProtocolSystemPrompt = `你是一名经验丰富的人民调解员和法律文书专家，精通《人民调解法》《民法典》等相关法律法规。请根据调解员提供的关键要素，生成一份规范、完整、具备法律效力的人民调解协议书。

协议书必须严格按照以下标准结构输出（Markdown格式）：

# 人民调解协议书

**编号：** [自动生成规则：地区简称+调+年份+序号，例如：京调字2025第00123号]
**签订地点：** [填写]
**签订时间：** [YYYY年MM月DD日]

---

## 一、当事人基本信息

### 申请人（甲方）
- 姓名：
- 性别：
- 民族：
- 身份证号：
- 住址/联系地址：
- 联系电话：

### 被申请人（乙方）
- 姓名：
- 性别：
- 民族：
- 身份证号：
- 住址/联系地址：
- 联系电话：

---

## 二、纠纷简要情况

[清晰、客观地描述纠纷发生的时间、地点、原因、经过、双方争议焦点和各自主张]

---

## 三、责任划分

[明确双方在纠纷中的责任比例和具体责任内容，如：甲方承担70%主要责任，乙方承担30%次要责任，理由如下：...]

---

## 四、协议内容

经人民调解委员会调解，双方当事人自愿、平等协商，达成如下协议：

### （一）赔偿/补偿事项
1. **赔偿金额：** 人民币XX元（大写：XX圆整）
2. **支付方式：** [一次性支付/分期支付]
3. **履行期限：** [具体日期，如：2025年X月X日前全部支付完毕]
4. **支付账户/方式：** [现金/银行转账等]

### （二）其他履行事项
1. [具体行为履行要求，如：乙方于X月X日前拆除违建围墙]
2. [赔礼道歉/恢复名誉等非财产性履行]
3. ...

### （三）双方权利义务
1. 甲方权利：[按时足额获得赔偿/补偿款；监督乙方履行协议等]
2. 甲方义务：[配合乙方履行；不得就同一纠纷再次主张权利；保密等]
3. 乙方权利：[要求甲方按约配合履行；协议履行完毕后免责等]
4. 乙方义务：[按时足额支付赔偿/补偿款；按约履行其他义务；保密等]

---

## 五、违约责任

1. 若甲方违约：[具体违约责任，如：双倍返还已收款项，承担违约金XX元]
2. 若乙方违约：[具体违约责任，如：逾期支付按每日万分之五支付违约金，逾期超过15日甲方可就剩余全部款项一并主张]
3. 守约方有权就违约部分向人民法院申请强制执行司法确认后的调解协议。

---

## 六、法律依据

本协议依据以下法律法规订立：
1. 《中华人民共和国人民调解法》第X条、第X条
2. 《中华人民共和国民法典》第X编第X章（如：侵权责任编、合同编等）第X条
3. [适用的其他行政法规、地方性法规或司法解释]

---

## 七、其他约定

1. 本协议为双方真实意思表示，不存在欺诈、胁迫、重大误解等可撤销情形。
2. 本协议自双方当事人签字（捺印）、调解员签名并加盖人民调解委员会印章之日起生效。
3. 本协议一式X份，双方当事人各执一份，人民调解委员会留存一份，[人民法院/司法所]备案一份，具有同等法律效力。
4. 双方自愿自本协议生效之日起30日内共同向人民法院申请司法确认。
5. 本协议履行完毕后，双方就本次纠纷互不追究。

---

## 八、签署

**申请人（甲方）签字/捺印：** __________________  日期：________

**被申请人（乙方）签字/捺印：** __________________  日期：________

**调解员签名：** __________________

**人民调解委员会（盖章）：**

日期：________年____月____日`
)

type MediationProtocolParams struct {
	CaseID           int64    `json:"caseId"`
	CaseNo           string   `json:"caseNo"`
	CaseTitle        string   `json:"caseTitle"`
	DisputeType      string   `json:"disputeType"`
	PartyAName       string   `json:"partyAName"`
	PartyAGender     string   `json:"partyAGender,omitempty"`
	PartyAIDCard     string   `json:"partyAIDCard,omitempty"`
	PartyAAddress    string   `json:"partyAAddress,omitempty"`
	PartyAPhone      string   `json:"partyAPhone,omitempty"`
	PartyBName       string   `json:"partyBName"`
	PartyBGender     string   `json:"partyBGender,omitempty"`
	PartyBIDCard     string   `json:"partyBIDCard,omitempty"`
	PartyBAddress    string   `json:"partyBAddress,omitempty"`
	PartyBPhone      string   `json:"partyBPhone,omitempty"`
	DisputeSummary   string   `json:"disputeSummary"`
	LiabilityParty   string   `json:"liabilityParty"`
	LiabilityRatioA  int      `json:"liabilityRatioA"`
	LiabilityRatioB  int      `json:"liabilityRatioB"`
	LiabilityReason  string   `json:"liabilityReason"`
	CompensationAmount float64 `json:"compensationAmount"`
	CompensationType string   `json:"compensationType"`
	PaymentMethod    string   `json:"paymentMethod"`
	PerformanceDate  string   `json:"performanceDate"`
	PaymentAccount   string   `json:"paymentAccount,omitempty"`
	OtherTerms       []string `json:"otherTerms,omitempty"`
	BreachClause     string   `json:"breachClause,omitempty"`
	MediatorName     string   `json:"mediatorName"`
	SignPlace        string   `json:"signPlace,omitempty"`
	SignDate         string   `json:"signDate,omitempty"`
	RegionPrefix     string   `json:"regionPrefix,omitempty"`
	ProtocolYear     int      `json:"protocolYear,omitempty"`
	ProtocolSeq      int      `json:"protocolSeq,omitempty"`
}

type MediationProtocolResult struct {
	ProtocolNo     string   `json:"protocolNo"`
	Title          string   `json:"title"`
	Content        string   `json:"content"`
	PartyAName     string   `json:"partyAName"`
	PartyBName     string   `json:"partyBName"`
	MediatorName   string   `json:"mediatorName"`
	AgreementItems string   `json:"agreementItems"`
	BreachClause   string   `json:"breachClause"`
	LegalBasis     []string `json:"legalBasis"`
	GeneratedAt    string   `json:"generatedAt"`
}

func GenerateMediationProtocol(params *MediationProtocolParams) (*MediationProtocolResult, error) {
	if params == nil {
		return nil, fmt.Errorf("params is nil")
	}

	client := GetDeepSeekClient()

	userContent := buildProtocolUserPrompt(params)

	messages := []ChatMessage{
		{Role: "user", Content: userContent},
	}

	rawContent, err := client.ChatCompletion(messages, MediationProtocolSystemPrompt)
	if err != nil {
		logger.Error("Generate mediation protocol failed", logger.Error(err))
		return nil, fmt.Errorf("调用AI生成协议失败: %w", err)
	}

	result := parseProtocolResult(rawContent, params)

	return result, nil
}

func buildProtocolUserPrompt(params *MediationProtocolParams) string {
	ratioText := ""
	if params.LiabilityRatioA > 0 || params.LiabilityRatioB > 0 {
		ratioText = fmt.Sprintf("，其中甲方承担%d%%，乙方承担%d%%", params.LiabilityRatioA, params.LiabilityRatioB)
	}

	otherTermsText := ""
	if len(params.OtherTerms) > 0 {
		for i, term := range params.OtherTerms {
			otherTermsText += fmt.Sprintf("%d. %s\n", i+1, term)
		}
	}

	breachText := params.BreachClause
	if breachText == "" {
		breachText = "按《民法典》及相关法律规定，违约方应承担继续履行、赔偿损失等违约责任，逾期付款按每日万分之五支付违约金。"
	}

	return fmt.Sprintf(`请根据以下关键要素生成一份完整规范的人民调解协议书：

## 案件基本信息
- 案件编号：%s
- 案件名称：%s
- 纠纷类型：%s

## 当事人信息
【甲方（申请人）】
- 姓名：%s
- 性别：%s
- 身份证号：%s
- 住址：%s
- 联系电话：%s

【乙方（被申请人）】
- 姓名：%s
- 性别：%s
- 身份证号：%s
- 住址：%s
- 联系电话：%s

## 纠纷简要情况
%s

## 责任划分
主要责任方：%s%s
责任理由：%s

## 赔偿/补偿方案
- 金额：人民币%.2f元
- 类型：%s（赔偿金/补偿金/医疗费/误工费等）
- 支付方式：%s
- 履行期限：%s之前
- 支付方式/账户：%s

## 其他履行事项
%s

## 违约责任
%s

## 调解员与签署信息
- 调解员姓名：%s
- 签订地点：%s
- 签订时间：%s
- 协议编号前缀：%s
- 协议年份：%d
- 协议序号：%d

请严格输出Markdown格式的完整协议书，协议编号按"前缀+调+年份+第+序号+号"格式自动生成。条款中的法律依据请根据纠纷类型引用《民法典》对应编章和《人民调解法》的具体条款。`,
		params.CaseNo,
		params.CaseTitle,
		params.DisputeType,
		params.PartyAName, params.PartyAGender, params.PartyAIDCard, params.PartyAAddress, params.PartyAPhone,
		params.PartyBName, params.PartyBGender, params.PartyBIDCard, params.PartyBAddress, params.PartyBPhone,
		params.DisputeSummary,
		params.LiabilityParty, ratioText,
		params.LiabilityReason,
		params.CompensationAmount, params.CompensationType,
		params.PaymentMethod, params.PerformanceDate, params.PaymentAccount,
		otherTermsText,
		breachText,
		params.MediatorName,
		params.SignPlace, params.SignDate,
		params.RegionPrefix, params.ProtocolYear, params.ProtocolSeq,
	)
}

func parseProtocolResult(raw string, params *MediationProtocolParams) *MediationProtocolResult {
	raw = strings.TrimSpace(raw)

	protocolNo := fmt.Sprintf("%s调%d第%05d号", params.RegionPrefix, params.ProtocolYear, params.ProtocolSeq)
	if params.RegionPrefix == "" {
		protocolNo = fmt.Sprintf("调%d第%05d号", params.ProtocolYear, params.ProtocolSeq)
	}

	title := fmt.Sprintf("人民调解协议书（%s）", params.CaseNo)

	agreementItems := fmt.Sprintf("1. %s向%s支付%s人民币%.2f元；\n2. 履行期限：%s；\n3. 支付方式：%s",
		params.LiabilityParty,
		getCounterParty(params.LiabilityParty, params.PartyAName, params.PartyBName),
		params.CompensationType,
		params.CompensationAmount,
		params.PerformanceDate,
		params.PaymentMethod,
	)
	if len(params.OtherTerms) > 0 {
		for i, term := range params.OtherTerms {
			agreementItems += fmt.Sprintf("；\n%d. %s", i+4, term)
		}
	}

	breachClause := params.BreachClause
	if breachClause == "" {
		breachClause = "违约方应承担继续履行、赔偿损失等违约责任，逾期付款按每日万分之五支付违约金，逾期超过15日守约方可就剩余款项一并主张权利。"
	}

	legalBasis := inferLegalBasis(params.DisputeType)

	generatedAt := time.Now().Format("2006-01-02 15:04:05")

	_ = json.Unmarshal([]byte("{}"), &map[string]interface{}{})

	return &MediationProtocolResult{
		ProtocolNo:     protocolNo,
		Title:          title,
		Content:        raw,
		PartyAName:     params.PartyAName,
		PartyBName:     params.PartyBName,
		MediatorName:   params.MediatorName,
		AgreementItems: agreementItems,
		BreachClause:   breachClause,
		LegalBasis:     legalBasis,
		GeneratedAt:    generatedAt,
	}
}

func getCounterParty(liabilityParty, partyA, partyB string) string {
	if liabilityParty == partyA {
		return partyB
	}
	return partyA
}

func inferLegalBasis(disputeType string) []string {
	basis := []string{
		"《中华人民共和国人民调解法》第二条、第二十八条、第二十九条、第三十一条",
	}

	dt := strings.ToLower(disputeType)
	switch {
	case strings.Contains(dt, "侵权") || strings.Contains(dt, "伤害") || strings.Contains(dt, "损"):
		basis = append(basis, "《中华人民共和国民法典》第七编 侵权责任 第一千一百六十五条、第一千一百七十九条、第一千一百八十四条")
	case strings.Contains(dt, "合同") || strings.Contains(dt, "违约"):
		basis = append(basis, "《中华人民共和国民法典》第三编 合同 第一分编 通则 第五百零九条、第五百七十七条、第五百八十四条")
	case strings.Contains(dt, "劳动") || strings.Contains(dt, "工资") || strings.Contains(dt, "工伤"):
		basis = append(basis, "《中华人民共和国劳动法》第七十七条、第七十九条", "《中华人民共和国劳动合同法》第三条、第二十九条")
	case strings.Contains(dt, "婚姻") || strings.Contains(dt, "家庭") || strings.Contains(dt, "抚养") || strings.Contains(dt, "赡养"):
		basis = append(basis, "《中华人民共和国民法典》第五编 婚姻家庭")
	case strings.Contains(dt, "物权") || strings.Contains(dt, "相邻") || strings.Contains(dt, "土地") || strings.Contains(dt, "房产"):
		basis = append(basis, "《中华人民共和国民法典》第二编 物权 第二分编 所有权")
	case strings.Contains(dt, "邻里") || strings.Contains(dt, "物业"):
		basis = append(basis, "《中华人民共和国民法典》第二编 物权 第七章 相邻关系、第二十四章 物业服务合同")
	default:
		basis = append(basis, "《中华人民共和国民法典》第一编 总则 第六章 民事法律行为、第一百七十六条 民事责任")
	}

	logger.Debug("Inferred legal basis for mediation protocol",
		zap.String("disputeType", disputeType),
		zap.Strings("basis", basis),
	)

	return basis
}
