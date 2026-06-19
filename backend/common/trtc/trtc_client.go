package trtc

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/dispute-resolve/common/config"
	"github.com/dispute-resolve/common/logger"
	"go.uber.org/zap"
)

type TRTCClient struct {
	sdkAppID  uint32
	secretKey string
	adminUser string
}

var trtcClient *TRTCClient

func InitTRTC() {
	cfg := config.GetConfig()
	trtcClient = &TRTCClient{
		sdkAppID:  cfg.TRTC.SdkAppID,
		secretKey: cfg.TRTC.SecretKey,
		adminUser: cfg.TRTC.AdminUserID,
	}
	logger.Info("TRTC client initialized",
		zap.Uint32("sdkAppID", trtcClient.sdkAppID),
		zap.String("adminUser", trtcClient.adminUser),
	)
}

func GetTRTCClient() *TRTCClient {
	if trtcClient == nil {
		InitTRTC()
	}
	return trtcClient
}

func (c *TRTCClient) GetSdkAppID() uint32 {
	return c.sdkAppID
}

func (c *TRTCClient) GetAdminUserID() string {
	return c.adminUser
}

func (c *TRTCClient) GenUserSig(userID string, expireSeconds uint32) string {
	if expireSeconds == 0 {
		expireSeconds = 86400
	}
	currTime := uint32(time.Now().Unix())
	return c.genUserSigImpl(c.sdkAppID, c.secretKey, userID, expireSeconds, currTime, nil)
}

func (c *TRTCClient) genUserSigImpl(sdkAppID uint32, secretKey string, userID string, expire uint32, currTime uint32, base64UserBuf *string) string {
	sigContent := fmt.Sprintf("TLS.identifier=%s&TLS.sdkappid=%d&TLS.expire=%d&TLS.time=%d&TLS.accounttype=0",
		userID, sdkAppID, expire, currTime)

	var userBuf string
	if base64UserBuf != nil {
		userBuf = *base64UserBuf
		sigContent += fmt.Sprintf("&TLS.userbuf=%s", userBuf)
	}

	mac := hmacSha256String(secretKey, sigContent)

	sig := map[string]interface{}{
		"TLS.ver":        "2.0",
		"TLS.identifier": userID,
		"TLS.sdkappid":   sdkAppID,
		"TLS.expire":     expire,
		"TLS.time":       currTime,
		"TLS.sig":        mac,
	}
	if userBuf != "" {
		sig["TLS.userbuf"] = userBuf
	}

	jsonBytes, _ := json.Marshal(sig)

	base64Sig := base64StdEncode(jsonBytes)
	compressed := compressStr(base64Sig)

	return compressed
}

func (c *TRTCClient) GenPrivateMapKey(userID string, expire uint32, roomID uint32) string {
	currTime := uint32(time.Now().Unix())

	privMapStr := fmt.Sprintf("TLS.identifier=%s&TLS.sdkappid=%d&TLS.expire=%d&TLS.time=%d&TLS.roomid=%d&TLS.privilege=255&usersig="+
		"TLS.accounttype=0&TLS.identifier=%s&TLS.sdkappid=%d&TLS.expire=%d&TLS.time=%d&TLS.accounttype=0",
		userID, c.sdkAppID, expire, currTime, roomID,
		userID, c.sdkAppID, expire, currTime)

	privMap := base64StdEncode([]byte(privMapStr))
	userSig := c.GenUserSig(userID, expire)
	sigContent := fmt.Sprintf("%s&userbuf=%s", userSig, privMap)
	_ = sigContent

	return userSig
}

func hmacSha256String(key string, data string) string {
	h := hmac.New(sha256.New, []byte(key))
	io.WriteString(h, data)
	return hex.EncodeToString(h.Sum(nil))
}

func base64StdEncode(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}

func compressStr(source string) string {
	if len(source) == 0 {
		return source
	}

	result := strings.Builder{}
	count := 1
	lastChar := source[0]

	for i := 1; i < len(source); i++ {
		if source[i] == lastChar {
			count++
		} else {
			result.WriteByte(lastChar)
			if count > 1 {
				result.WriteString(strconv.Itoa(count))
			}
			lastChar = source[i]
			count = 1
		}
	}

	result.WriteByte(lastChar)
	if count > 1 {
		result.WriteString(strconv.Itoa(count))
	}

	compressed := result.String()
	if len(compressed) < len(source) {
		return compressed
	}
	return source
}

func (c *TRTCClient) generateRandomNumber() uint32 {
	return uint32(rand.Uint32())
}

func hmacsha256(key, content string) string {
	mac := hmac.New(sha256.New, []byte(key))
	mac.Write([]byte(content))
	return fmt.Sprintf("%x", mac.Sum(nil))
}
