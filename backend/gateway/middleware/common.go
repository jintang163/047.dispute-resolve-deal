package middleware

import (
	"context"
	"fmt"
	"time"

	"github.com/dispute-resolve/common/cache"
	"github.com/dispute-resolve/common/constants"
	"github.com/dispute-resolve/common/response"
	"github.com/dispute-resolve/common/utils"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/hlog"
)

func RequestIDMiddleware() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = utils.GenerateUUID()
		}
		c.Set("requestID", requestID)
		c.Header("X-Request-ID", requestID)
		ctx = context.WithValue(ctx, "requestID", requestID)
		c.Next(ctx)
	}
}

func LoggerMiddleware() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		start := time.Now()
		path := c.FullPath()
		method := string(c.Method())
		requestID := c.GetString("requestID")

		hlog.CtxInfof(ctx, "Request started | method=%s path=%s requestID=%s ip=%s",
			method, path, requestID, c.ClientIP())

		c.Next(ctx)

		latency := time.Since(start)
		statusCode := c.Response.StatusCode()

		hlog.CtxInfof(ctx, "Request completed | method=%s path=%s status=%d latency=%v requestID=%s",
			method, path, statusCode, latency, requestID)
	}
}

func RateLimitMiddleware() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		ip := c.ClientIP()
		path := c.FullPath()
		key := fmt.Sprintf("%s%s:%s", constants.RedisKeyPrefixRateLimit, ip, path)

		count, err := cache.Incr(ctx, key)
		if err != nil {
			hlog.CtxWarnf(ctx, "Rate limit check failed: %v", err)
			c.Next(ctx)
			return
		}

		if count == 1 {
			cache.Expire(ctx, key, time.Duration(constants.RedisExpireRateLimit)*time.Second)
		}

		limit := 100
		if count > int64(limit) {
			hlog.CtxWarnf(ctx, "Rate limit exceeded | ip=%s path=%s count=%d", ip, path, count)
			c.JSON(429, response.Fail(429, "请求过于频繁，请稍后再试"))
			c.Abort()
			return
		}

		c.Next(ctx)
	}
}

func OperationLogMiddleware(module string) app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		userInfo := GetUserInfo(c)
		if userInfo == nil {
			c.Next(ctx)
			return
		}

		start := time.Now()
		params := string(c.Request.Body())

		c.Next(ctx)

		latency := time.Since(start)
		status := 1
		if c.Response.StatusCode() >= 400 {
			status = 0
		}

		hlog.CtxInfof(ctx, "Operation log | user=%d module=%s operation=%s status=%d latency=%v",
			userInfo.UserID, module, string(c.Method())+" "+c.FullPath(), status, latency)

		go func() {
			_ = params
		}()
	}
}

func DataPermissionMiddleware() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		userInfo := GetUserInfo(c)
		if userInfo == nil {
			c.Next(ctx)
			return
		}

		c.Set("userOrgID", userInfo.OrganizationID)
		c.Set("userRole", userInfo.Role)

		switch userInfo.Role {
		case constants.RoleAdmin, constants.RoleDirector:
		case constants.RoleLeader:
			c.Set("dataScope", "org")
		case constants.RoleMediator:
			c.Set("dataScope", "self")
		default:
			c.Set("dataScope", "self")
		}

		c.Next(ctx)
	}
}
