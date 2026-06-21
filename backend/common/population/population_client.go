package population

import (
	"bytes"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/dispute-resolve/common/config"
	"github.com/dispute-resolve/common/logger"
	"go.uber.org/zap"
)

type PopulationInfo struct {
	Name        string `json:"name"`
	Gender      int    `json:"gender"`
	GenderName  string `json:"genderName"`
	Age         int    `json:"age"`
	Nation      string `json:"nation"`
	BirthDate   string `json:"birthDate"`
	IDCard      string `json:"idcard"`
	Address     string `json:"address"`
	Phone       string `json:"phone"`
	Household   string `json:"household"`
	EthnicCode  string `json:"ethnicCode"`
	Issuer      string `json:"issuer"`
	ValidPeriod string `json:"validPeriod"`
}

type PopulationClient struct {
	baseURL    string
	appID      string
	appSecret  string
	mockMode   bool
	httpClient *http.Client
}

var populationClient *PopulationClient

func NewPopulationClient(cfg *config.PopulationConfig) *PopulationClient {
	timeout := 30
	if cfg.Timeout > 0 {
		timeout = cfg.Timeout
	}
	return &PopulationClient{
		baseURL:   cfg.APIEndpoint,
		appID:     cfg.AppID,
		appSecret: cfg.AppSecret,
		mockMode:  cfg.MockMode,
		httpClient: &http.Client{
			Timeout: time.Duration(timeout) * time.Second,
		},
	}
}

func InitPopulation() {
	cfg := config.GetConfig()
	if cfg.Population.APIEndpoint == "" && !cfg.Population.MockMode {
		logger.Warn("Population API not configured, skipping initialization")
		return
	}
	populationClient = NewPopulationClient(&cfg.Population)
	logger.Info("Population client initialized",
		zap.String("baseURL", populationClient.baseURL),
		zap.Bool("mockMode", populationClient.mockMode),
	)
}

func GetPopulationClient() *PopulationClient {
	if populationClient == nil {
		InitPopulation()
	}
	return populationClient
}

func (c *PopulationClient) generateSign(params map[string]interface{}, timestamp string) string {
	h := md5.New()
	h.Write([]byte(c.appSecret + timestamp))
	for k, v := range params {
		h.Write([]byte(fmt.Sprintf("%s=%v", k, v)))
	}
	return fmt.Sprintf("%x", h.Sum(nil))
}

func (c *PopulationClient) QueryByIDCard(idCard string) (*PopulationInfo, error) {
	if c.mockMode {
		return mockQueryByIDCard(idCard)
	}
	return c.queryByIDCardReal(idCard)
}

func (c *PopulationClient) queryByIDCardReal(idCard string) (*PopulationInfo, error) {
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	params := map[string]interface{}{
		"app_id":    c.appID,
		"idcard":    idCard,
		"timestamp": timestamp,
	}
	sign := c.generateSign(params, timestamp)
	params["sign"] = sign

	url := fmt.Sprintf("%s/api/population/query", c.baseURL)
	body, _ := json.Marshal(params)

	req, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request failed: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	var result struct {
		Code    int             `json:"code"`
		Message string          `json:"message"`
		Data    *PopulationInfo `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response failed: %v", err)
	}

	if result.Code != 0 {
		return nil, fmt.Errorf("population query failed: %s", result.Message)
	}

	return result.Data, nil
}

func mockQueryByIDCard(idCard string) (*PopulationInfo, error) {
	if len(idCard) != 18 {
		return nil, fmt.Errorf("身份证号格式不正确")
	}

	genderCode := idCard[16:17]
	gender := 1
	genderName := "男"
	if num, err := strconv.Atoi(genderCode); err == nil && num%2 == 0 {
		gender = 2
		genderName = "女"
	}

	birthYear, _ := strconv.Atoi(idCard[6:10])
	birthMonth, _ := strconv.Atoi(idCard[10:12])
	birthDay, _ := strconv.Atoi(idCard[12:14])
	birthDate := fmt.Sprintf("%04d-%02d-%02d", birthYear, birthMonth, birthDay)

	now := time.Now()
	age := now.Year() - birthYear
	if now.YearDay() < dayOfYear(birthYear, birthMonth, birthDay) {
		age--
	}

	regionCode := idCard[0:6]
	regionMap := map[string]string{
		"110101": "北京市东城区",
		"110102": "北京市西城区",
		"110105": "北京市朝阳区",
		"110106": "北京市丰台区",
		"110108": "北京市海淀区",
		"310101": "上海市黄浦区",
		"310104": "上海市徐汇区",
		"310105": "上海市长宁区",
		"330102": "浙江省杭州市上城区",
		"330103": "浙江省杭州市下城区",
		"330104": "浙江省杭州市江干区",
		"320102": "江苏省南京市玄武区",
		"320104": "江苏省南京市秦淮区",
		"440103": "广东省广州市荔湾区",
		"440104": "广东省广州市越秀区",
		"440303": "广东省深圳市罗湖区",
		"440304": "广东省深圳市福田区",
		"420102": "湖北省武汉市江岸区",
		"420103": "湖北省武汉市江汉区",
		"510104": "四川省成都市锦江区",
		"510105": "四川省成都市青羊区",
	}
	region := regionMap[regionCode]
	if region == "" {
		region = "某市某区某街道"
	}

	nations := []string{"汉", "满", "回", "壮", "蒙古"}
	nation := nations[rand.Intn(len(nations))]

	surname := ""
	if gender == 1 {
		surnames := []string{"张伟", "王伟", "李强", "王刚", "刘洋", "陈明", "杨光", "黄伟", "周杰", "吴刚"}
		surname = surnames[rand.Intn(len(surnames))]
	} else {
		surnames := []string{"王芳", "李娜", "张敏", "刘静", "陈丽", "杨燕", "赵敏", "黄婷", "周雪", "吴霞"}
		surname = surnames[rand.Intn(len(surnames))]
	}

	phonePrefix := []string{"138", "139", "158", "159", "188", "189", "136", "137"}
	prefix := phonePrefix[rand.Intn(len(phonePrefix))]
	phone := prefix + fmt.Sprintf("%08d", rand.Intn(100000000))

	return &PopulationInfo{
		Name:        surname,
		Gender:      gender,
		GenderName:  genderName,
		Age:         age,
		Nation:      nation,
		BirthDate:   birthDate,
		IDCard:      idCard,
		Address:     fmt.Sprintf("%s某某路XX号XX小区X栋X单元XXX室", region),
		Phone:       phone,
		Household:   fmt.Sprintf("%s某某派出所", region),
		EthnicCode:  nation,
		Issuer:      fmt.Sprintf("%s公安分局", region),
		ValidPeriod: fmt.Sprintf("%04d.%02d.%02d-%04d.%02d.%02d", birthYear+16, birthMonth, birthDay, birthYear+36, birthMonth, birthDay),
	}, nil
}

func dayOfYear(year, month, day int) int {
	t := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.Local)
	return t.YearDay()
}
