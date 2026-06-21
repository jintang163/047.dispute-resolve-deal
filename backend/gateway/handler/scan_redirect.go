package handler

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/dispute-resolve-deal/backend/common/config"
	"github.com/dispute-resolve-deal/backend/common/database"
	"github.com/dispute-resolve-deal/backend/common/response"
	"github.com/dispute-resolve-deal/backend/common/utils"
)

type ReceiptQRCodeRequest struct {
	CaseNo    string `json:"caseNo"`
	Phone     string `json:"phone,omitempty"`
	ExpiredIn int    `json:"expiredIn,omitempty"`
}

type ReceiptQRCodeResult struct {
	CaseNo      string `json:"caseNo"`
	Token       string `json:"token"`
	QRCodeURL   string `json:"qrCodeUrl"`
	MiniAppURL  string `json:"miniAppUrl"`
	ExpiredAt   string `json:"expiredAt"`
}

type ScanTokenPayload struct {
	CaseNo    string `json:"c"`
	Phone     string `json:"p,omitempty"`
	ExpiredAt int64  `json:"e"`
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

	payload := ScanTokenPayload{
		CaseNo:    req.CaseNo,
		Phone:     req.Phone,
		ExpiredAt: expiredAt.Unix(),
	}

	token, err := encryptScanToken(payload)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ServerError("生成扫码Token失败"))
		return
	}

	cfg := config.GetConfig()
	scanBaseURL := cfg.WechatMiniApp.ScanBaseURL
	if scanBaseURL == "" {
		scanBaseURL = fmt.Sprintf("https://%s/api/v1/public", cfg.Server.Host)
	}

	qrCodeURL := fmt.Sprintf("%s/scan/%s", scanBaseURL, token)

	miniAppURL := fmt.Sprintf("weixin://dl/business/?appid=%s&path=%s&query=caseNo=%s",
		cfg.WechatMiniApp.AppID,
		cfg.WechatMiniApp.QRCodePath,
		req.CaseNo,
	)
	if cfg.WechatMiniApp.AppID == "" {
		miniAppURL = ""
	}

	c.JSON(http.StatusOK, response.Success(ReceiptQRCodeResult{
		CaseNo:     req.CaseNo,
		Token:      token,
		QRCodeURL:  qrCodeURL,
		MiniAppURL: miniAppURL,
		ExpiredAt:  expiredAt.Format("2006-01-02 15:04:05"),
	}))
}

func ScanRedirect(ctx context.Context, c *app.RequestContext) {
	token := c.Param("token")
	if token == "" {
		c.HTML(http.StatusBadRequest, "error.html", "无效的扫码链接")
		return
	}

	payload, err := decryptScanToken(token)
	if err != nil {
		c.HTML(http.StatusBadRequest, "error.html", "二维码已失效或格式错误")
		return
	}

	if time.Now().Unix() > payload.ExpiredAt {
		c.HTML(http.StatusGone, "expired.html", "该二维码已过期，请重新获取")
		return
	}

	cfg := config.GetConfig()
	appID := cfg.WechatMiniApp.AppID
	originalID := cfg.WechatMiniApp.OriginalID
	qrCodePath := cfg.WechatMiniApp.QRCodePath

	if appID == "" {
		appID = "wx_placeholder_appid"
	}
	if originalID == "" {
		originalID = "gh_placeholder_originalid"
	}
	if qrCodePath == "" {
		qrCodePath = "pages/progress/index"
	}

	queryStr := fmt.Sprintf("caseNo=%s", payload.CaseNo)
	if payload.Phone != "" {
		queryStr += fmt.Sprintf("&phone=%s", payload.Phone)
	}

	encodedQuery := urlEncode(queryStr)
	launchURL := fmt.Sprintf("weixin://dl/business/?appid=%s&path=%s&query=%s",
		appID, urlEncode(qrCodePath), encodedQuery)

	html := buildScanRedirectHTML(payload.CaseNo, appID, originalID, qrCodePath, queryStr, launchURL)
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(html))
}

func buildScanRedirectHTML(caseNo, appID, originalID, path, query, launchURL string) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="zh-CN">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0, maximum-scale=1.0, user-scalable=no">
<title>纠纷调解进度查询</title>
<style>
*{margin:0;padding:0;box-sizing:border-box}
body{font-family:-apple-system,BlinkMacSystemFont,"Segoe UI","PingFang SC","Hiragino Sans GB","Microsoft YaHei",sans-serif;background:#f0f5ff;min-height:100vh;display:flex;align-items:center;justify-content:center}
.container{background:#fff;border-radius:16px;padding:40px 24px;max-width:400px;width:90%%;text-align:center;box-shadow:0 4px 24px rgba(0,0,0,0.08)}
.icon{font-size:64px;margin-bottom:16px}
.title{font-size:22px;font-weight:700;color:#1a1a1a;margin-bottom:8px}
.subtitle{font-size:14px;color:#666;margin-bottom:24px;line-height:1.6}
.case-no{font-family:"Courier New",monospace;font-size:20px;font-weight:800;color:#1d6cff;letter-spacing:2px;background:#f0f5ff;padding:12px 20px;border-radius:8px;margin-bottom:24px;display:inline-block}
.btn{display:block;width:100%%;padding:14px;background:#07c160;color:#fff;border:none;border-radius:12px;font-size:17px;font-weight:600;cursor:pointer;text-decoration:none;margin-bottom:12px}
.btn:active{background:#06ad56}
.btn-secondary{background:#1d6cff}
.btn-secondary:active{background:#1557d6}
.tip{font-size:12px;color:#999;margin-top:16px;line-height:1.6}
.auto-tip{color:#1d6cff;font-weight:500;margin-bottom:16px;font-size:14px}
</style>
</head>
<body>
<div class="container">
  <div class="icon">📋</div>
  <div class="title">纠纷调解进度查询</div>
  <div class="subtitle">您正在查询案件调解进度</div>
  <div class="case-no">%s</div>
  <div class="auto-tip">正在自动跳转小程序...</div>
  <a href="%s" class="btn" id="openBtn">打开小程序查询进度</a>
  <a href="%s" class="btn btn-secondary" id="copyBtn">复制案件编号</a>
  <div class="tip">
    如未自动跳转，请点击上方按钮<br>
    也可在微信中搜索小程序手动输入案件编号查询<br>
    案件编号有效期7天，过期后请重新获取
  </div>
</div>
<script>
(function(){
  var caseNo = "%s";
  var launchUrl = "%s";

  setTimeout(function(){
    window.location.href = launchUrl;
  }, 800);

  document.getElementById("copyBtn").addEventListener("click", function(e){
    e.preventDefault();
    if(navigator.clipboard && navigator.clipboard.writeText){
      navigator.clipboard.writeText(caseNo).then(function(){
        alert("案件编号已复制: " + caseNo);
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
    try{ document.execCommand("copy"); alert("案件编号已复制: " + text); }
    catch(e){ alert("复制失败，请手动复制: " + text); }
    document.body.removeChild(ta);
  }
})();
</script>
</body>
</html>`, caseNo, launchURL, "#", caseNo, launchURL)
}

func encryptScanToken(payload ScanTokenPayload) (string, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	keyHex := getScanTokenKey()
	key, err := hex.DecodeString(keyHex)
	if err != nil {
		return "", err
	}

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

	keyHex := getScanTokenKey()
	key, err := hex.DecodeString(keyHex)
	if err != nil {
		return nil, err
	}

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

func getScanTokenKey() string {
	cfg := config.GetConfig()
	jwtSecret := cfg.JWT.Secret
	if len(jwtSecret) >= 64 {
		return jwtSecret[:64]
	}
	for len(jwtSecret) < 64 {
		jwtSecret = jwtSecret + jwtSecret
	}
	return jwtSecret[:64]
}

func urlEncode(s string) string {
	return strings.ReplaceAll(
		strings.ReplaceAll(
			strings.ReplaceAll(s, " ", "%20"),
			"=", "%3D"),
		"&", "%26")
}

func init() {
	_ = utils.GenerateID
}
