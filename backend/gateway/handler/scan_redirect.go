package handler

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/dispute-resolve-deal/backend/common/config"
	"github.com/dispute-resolve-deal/backend/common/database"
	"github.com/dispute-resolve-deal/backend/common/response"
)

type ReceiptQRCodeRequest struct {
	CaseNo    string `json:"caseNo" vd:"len($)>0"`
	Phone     string `json:"phone,omitempty"`
	ExpiredIn int    `json:"expiredIn,omitempty"`
}

type ReceiptQRCodeResult struct {
	CaseNo     string `json:"caseNo"`
	Token      string `json:"token"`
	QRCodeURL  string `json:"qrCodeUrl"`
	MiniAppURL string `json:"miniAppUrl"`
	ExpiredAt  string `json:"expiredAt"`
	ExpireDays int    `json:"expireDays"`
}

type ScanTokenPayload struct {
	CaseNo    string `json:"c"`
	Phone     string `json:"p,omitempty"`
	ExpiredAt int64  `json:"e"`
	Nonce     string `json:"n,omitempty"`
}

func GenerateReceiptQRCode(ctx context.Context, c *app.RequestContext) {
	var req ReceiptQRCodeRequest
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest(err.Error()))
		return
	}

	if req.CaseNo == "" {
		c.JSON(http.StatusBadRequest, response.BadRequest("案件编号不能为空"))
		return
	}

	var count int64
	database.GetDB().Table("dispute_case").Where("case_no = ?", req.CaseNo).Count(&count)
	if count == 0 {
		c.JSON(http.StatusNotFound, response.NotFound("案件不存在"))
		return
	}

	expiredIn := req.ExpiredIn
	if expiredIn <= 0 {
		expiredIn = 7 * 24
	}
	expiredAt := time.Now().Add(time.Duration(expiredIn) * time.Hour)
	expireDays := expiredIn / 24
	if expireDays <= 0 {
		expireDays = 1
	}

	nonce := generateNonce(8)
	payload := ScanTokenPayload{
		CaseNo:    req.CaseNo,
		Phone:     req.Phone,
		ExpiredAt: expiredAt.Unix(),
		Nonce:     nonce,
	}

	token, err := encryptScanToken(payload)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ServerError("生成扫码Token失败"))
		return
	}

	cfg := config.GetConfig()
	miniAppCfg := cfg.WechatMiniApp
	scanBaseURL := miniAppCfg.ScanBaseURL
	if scanBaseURL == "" {
		scanBaseURL = fmt.Sprintf("https://%s/api/v1/public", cfg.Server.Host)
	}

	qrCodeURL := fmt.Sprintf("%s/scan/%s", scanBaseURL, token)

	var miniAppURL string
	if miniAppCfg.AppID != "" {
		qrCodePath := miniAppCfg.QRCodePath
		if qrCodePath == "" {
			qrCodePath = "pages/progress/index"
		}
		queryStr := fmt.Sprintf("caseNo=%s", url.QueryEscape(req.CaseNo))
		if req.Phone != "" {
			queryStr += fmt.Sprintf("&phone=%s", url.QueryEscape(req.Phone))
		}
		miniAppURL = fmt.Sprintf(
			"weixin://dl/business/?appid=%s&path=%s&query=%s",
			miniAppCfg.AppID,
			url.QueryEscape(qrCodePath),
			queryStr,
		)
	}

	c.JSON(http.StatusOK, response.Success(ReceiptQRCodeResult{
		CaseNo:     req.CaseNo,
		Token:      token,
		QRCodeURL:  qrCodeURL,
		MiniAppURL: miniAppURL,
		ExpiredAt:  expiredAt.Format("2006-01-02 15:04:05"),
		ExpireDays: expireDays,
	}))
}

func ScanRedirect(ctx context.Context, c *app.RequestContext) {
	token := c.Param("token")
	if token == "" {
		renderErrorHTML(c, http.StatusBadRequest, "无效的扫码链接", "请检查二维码是否完整")
		return
	}

	payload, err := decryptScanToken(token)
	if err != nil {
		renderErrorHTML(c, http.StatusBadRequest, "二维码已失效或格式错误", "请重新获取案件回执获取最新二维码")
		return
	}

	if time.Now().Unix() > payload.ExpiredAt {
		renderErrorHTML(c, http.StatusGone, "该二维码已过期", "请重新到自助终端或小程序获取最新的案件回执二维码")
		return
	}

	cfg := config.GetConfig()
	miniAppCfg := cfg.WechatMiniApp
	appID := miniAppCfg.AppID
	originalID := miniAppCfg.OriginalID
	qrCodePath := miniAppCfg.QRCodePath

	if appID == "" {
		appID = "wx1234567890abcdef"
	}
	if originalID == "" {
		originalID = "gh_1234567890ab"
	}
	if qrCodePath == "" {
		qrCodePath = "pages/progress/index"
	}

	queryMap := map[string]string{"caseNo": payload.CaseNo}
	if payload.Phone != "" {
		queryMap["phone"] = payload.Phone
	}

	queryStr := buildQueryString(queryMap)
	encodedQuery := url.QueryEscape(queryStr)
	launchURL := fmt.Sprintf(
		"weixin://dl/business/?appid=%s&path=%s&query=%s",
		appID,
		url.QueryEscape(qrCodePath),
		encodedQuery,
	)

	html := buildScanRedirectHTML(payload.CaseNo, appID, originalID, qrCodePath, queryStr, launchURL)
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(html))
}

func renderErrorHTML(c *app.RequestContext, status int, title, message string) {
	html := fmt.Sprintf(`<!DOCTYPE html>
<html lang="zh-CN">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0, maximum-scale=1.0, user-scalable=no">
<title>%s</title>
<style>
*{margin:0;padding:0;box-sizing:border-box}
body{font-family:-apple-system,BlinkMacSystemFont,"Segoe UI","PingFang SC","Hiragino Sans GB","Microsoft YaHei",sans-serif;background:#fff5f5;min-height:100vh;display:flex;align-items:center;justify-content:center}
.container{background:#fff;border-radius:16px;padding:40px 24px;max-width:400px;width:90%%;text-align:center;box-shadow:0 4px 24px rgba(0,0,0,0.08)}
.icon{font-size:64px;margin-bottom:16px}
.title{font-size:20px;font-weight:700;color:#ff4d4f;margin-bottom:12px}
.message{font-size:14px;color:#666;margin-bottom:24px;line-height:1.6}
.btn{display:block;width:100%%;padding:14px;background:#1d6cff;color:#fff;border:none;border-radius:12px;font-size:16px;font-weight:600;cursor:pointer;text-decoration:none}
</style>
</head>
<body>
<div class="container">
  <div class="icon">⚠️</div>
  <div class="title">%s</div>
  <div class="message">%s</div>
  <a href="javascript:history.back()" class="btn">返回</a>
</div>
</body>
</html>`, title, title, message)
	c.Data(status, "text/html; charset=utf-8", []byte(html))
}

func buildScanRedirectHTML(caseNo, appID, originalID, path, query, launchURL string) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="zh-CN">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0, maximum-scale=1.0, user-scalable=no">
<title>纠纷调解进度查询</title>
<meta name="apple-mobile-web-app-capable" content="yes">
<meta name="apple-mobile-web-app-status-bar-style" content="default">
<style>
*{margin:0;padding:0;box-sizing:border-box}
body{font-family:-apple-system,BlinkMacSystemFont,"Segoe UI","PingFang SC","Hiragino Sans GB","Microsoft YaHei",sans-serif;background:linear-gradient(135deg,#667eea 0%%,#764ba2 100%%);min-height:100vh;display:flex;align-items:center;justify-content:center;padding:20px}
.container{background:#fff;border-radius:20px;padding:36px 24px;max-width:400px;width:100%%;text-align:center;box-shadow:0 20px 60px rgba(0,0,0,0.15)}
.icon-wrap{width:80px;height:80px;margin:0 auto 20px;border-radius:50%%;background:linear-gradient(135deg,#667eea,#764ba2);display:flex;align-items:center;justify-content:center;font-size:40px;color:#fff}
.title{font-size:22px;font-weight:700;color:#1a1a1a;margin-bottom:8px}
.subtitle{font-size:14px;color:#888;margin-bottom:24px;line-height:1.6}
.case-card{background:#f7f9fc;border-radius:12px;padding:16px;margin-bottom:24px}
.case-label{font-size:12px;color:#999;margin-bottom:6px}
.case-no{font-family:"SF Mono,Monaco,"Courier New",monospace;font-size:22px;font-weight:800;color:#1d6cff;letter-spacing:1px;word-break:break-all}
.auto-tip{color:#07c160;font-weight:600;margin-bottom:20px;font-size:14px;background:#f0fdf4;padding:10px 16px;border-radius:8px;display:inline-block}
.btn{display:flex;align-items:center;justify-content:center;width:100%%;padding:14px 20px;border:none;border-radius:12px;font-size:16px;font-weight:600;cursor:pointer;text-decoration:none;margin-bottom:12px;transition:transform .2s}
.btn:active{transform:scale(0.97)}
.btn-primary{background:#07c160;color:#fff}
.btn-primary:active{background:#06ad56}
.btn-secondary{background:#1d6cff;color:#fff}
.btn-secondary:active{background:#1557d6}
.btn-outline{background:transparent;color:#1d6cff;border:1px solid #1d6cff}
.tip{font-size:12px;color:#aaa;margin-top:20px;line-height:1.8;text-align:left;background:#fafafa;padding:12px;border-radius:8px}
.tip-item{margin-bottom:4px}
.divider{height:1px;background:#f0f0f0;margin:16px 0}
.qr-hint{font-size:12px;color:#999;margin-top:8px}
</style>
</head>
<body>
<div class="container">
  <div class="icon-wrap">📋</div>
  <div class="title">纠纷调解进度查询</div>
  <div class="subtitle">扫码成功，正在为您跳转小程序查询</div>
  <div class="case-card">
    <div class="case-label">案件编号</div>
    <div class="case-no">%s</div>
  </div>
  <div class="auto-tip">✓ 正在自动跳转小程序...</div>
  <a href="%s" class="btn btn-primary" id="openBtn">
    <span style="margin-right:8px">💚</span>
    打开微信小程序查询
  </a>
  <button class="btn btn-outline" id="copyBtn">复制案件编号</button>
  <div class="tip">
    <div class="tip-item">• 如未自动跳转，请点击上方「打开微信小程序</div>
    <div class="tip-item">• 也可在微信搜索小程序手动输入编号查询</div>
    <div class="tip-item">• 案件二维码7天内有效，请及时查询</div>
  </div>
  <div class="divider"></div>
  <div class="qr-hint">
    微信内打开可直接跳转小程序</div>
</div>
<script>
(function(){
  var caseNo = "%s";
  var launchUrl = "%s";
  var appId = "%s";
  var path = "%s";
  var query = "%s";

  setTimeout(function(){
    try {
      window.location.href = launchUrl;
    } catch(e) {}
  }, 1000);

  document.getElementById("copyBtn").addEventListener("click", function(e){
    e.preventDefault();
    if(navigator.clipboard && navigator.clipboard.writeText){
      navigator.clipboard.writeText(caseNo).then(function(){
        showToast("案件编号已复制");
      }).catch(function(){
        fallbackCopy(caseNo);
      });
    } else {
      fallbackCopy(caseNo);
    }
  });

  function fallbackCopy(text){
    var ta = document.createElement("textarea");
    ta.value = text;
    ta.style.position = "fixed";
    ta.style.left = "-9999px";
    document.body.appendChild(ta);
    ta.select();
    try{ document.execCommand("copy"); showToast("案件编号已复制"); }
    catch(e){ alert("复制失败，请手动复制: " + text); }
    document.body.removeChild(ta);
  }

  function showToast(msg){
    var toast = document.createElement("div");
    toast.style.cssText = "position:fixed;top:50%%;left:50%%;transform:translate(-50%%,-50%%);background:rgba(0,0,0,0.75);color:#fff;padding:12px 24px;border-radius:8px;font-size:14px;z-index:9999;";
    toast.textContent = msg;
    document.body.appendChild(toast);
    setTimeout(function(){ document.body.removeChild(toast); }, 1500);
  }

  if (typeof wx !== "undefined" && wx.miniProgram) {
    try {
      wx.miniProgram.getEnv(function(res) {
        if (res.miniprogram) {
        }
      });
    } catch(e) {}
  }
})();
</script>
<script src="https://res.wx.qq.com/open/js/jweixin-1.6.0.js"></script>
</body>
</html>`, caseNo, launchURL, caseNo, launchURL, appID, path, query)
}

func encryptScanToken(payload ScanTokenPayload) (string, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	key := deriveAESKey()
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, aesGCM.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := aesGCM.Seal(nonce, nonce, data, nil)
	return base64.URLEncoding.EncodeToString(ciphertext), nil
}

func decryptScanToken(token string) (*ScanTokenPayload, error) {
	ciphertext, err := base64.URLEncoding.DecodeString(token)
	if err != nil {
		return nil, fmt.Errorf("token decode failed: %w", err)
	}

	key := deriveAESKey()
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := aesGCM.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, fmt.Errorf("token too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("token decrypt failed: %w", err)
	}

	var payload ScanTokenPayload
	if err := json.Unmarshal(plaintext, &payload); err != nil {
		return nil, fmt.Errorf("token parse failed: %w", err)
	}

	return &payload, nil
}

func deriveAESKey() []byte {
	cfg := config.GetConfig()
	secret := cfg.JWT.Secret
	if secret == "" {
		secret = "dispute-resolve-default-secret-key-2024"
	}
	salt := "scan-token-salt-v1"
	combined := secret + ":" + salt
	hash := sha256.Sum256([]byte(combined))
	return hash[:]
}

func generateNonce(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return ""
	}
	for i := range b {
		b[i] = charset[int(b[i])%len(charset)]
	}
	return string(b)
}

func buildQueryString(params map[string]string) string {
	var parts []string
	for k, v := range params {
		parts = append(parts, k+"="+v)
	}
	return strings.Join(parts, "&")
}
