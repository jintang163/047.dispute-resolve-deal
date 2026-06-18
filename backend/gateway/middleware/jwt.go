package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/dispute-resolve/common/auth"
	"github.com/dispute-resolve/common/cache"
	"github.com/dispute-resolve/common/constants"
	"github.com/dispute-resolve/common/response"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/hlog"
)

type UserContextKey string

const (
	UserInfoKey UserContextKey = "userInfo"
)

func JWTAuthMiddleware() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, response.Unauthorized("未提供认证Token"))
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			c.JSON(http.StatusUnauthorized, response.Unauthorized("认证格式错误，请使用 Bearer {token} 格式"))
			c.Abort()
			return
		}

		tokenString := parts[1]

		claims, err := auth.ParseToken(tokenString)
		if err != nil {
			hlog.CtxError(ctx, "Parse token failed: %v", err)
			c.JSON(http.StatusUnauthorized, response.Unauthorized(err.Error()))
			c.Abort()
			return
		}

		cacheKey := constants.RedisKeyPrefixToken + tokenString
		exists, _ := cache.Exists(ctx, cacheKey)
		if !exists {
			c.JSON(http.StatusUnauthorized, response.Unauthorized("Token已失效，请重新登录"))
			c.Abort()
			return
		}

		userInfo := &auth.UserInfo{
			UserID:         claims.UserID,
			Username:       claims.Username,
			RealName:       claims.RealName,
			Role:           claims.Role,
			OrganizationID: claims.OrganizationID,
		}

		ctx = context.WithValue(ctx, UserInfoKey, userInfo)
		c.Set("userInfo", userInfo)

		c.Next(ctx)
	}
}

func WSJWTAuthMiddleware() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		token := c.Query("token")
		if token == "" {
			token = c.GetHeader("Authorization")
			if token != "" {
				parts := strings.SplitN(token, " ", 2)
				if len(parts) == 2 && parts[0] == "Bearer" {
					token = parts[1]
				}
			}
		}

		if token == "" {
			c.JSON(http.StatusUnauthorized, response.Unauthorized("未提供认证Token"))
			c.Abort()
			return
		}

		claims, err := auth.ParseToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, response.Unauthorized(err.Error()))
			c.Abort()
			return
		}

		userInfo := &auth.UserInfo{
			UserID:         claims.UserID,
			Username:       claims.Username,
			RealName:       claims.RealName,
			Role:           claims.Role,
			OrganizationID: claims.OrganizationID,
		}

		ctx = context.WithValue(ctx, UserInfoKey, userInfo)
		c.Set("userInfo", userInfo)

		c.Next(ctx)
	}
}

func AdminRequiredMiddleware() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		userInfo := GetUserInfo(c)
		if userInfo == nil {
			c.JSON(http.StatusUnauthorized, response.Unauthorized(""))
			c.Abort()
			return
		}

		if userInfo.Role != constants.RoleAdmin {
			c.JSON(http.StatusForbidden, response.Forbidden("需要管理员权限"))
			c.Abort()
			return
		}

		c.Next(ctx)
	}
}

func RoleRequiredMiddleware(roles ...int32) app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		userInfo := GetUserInfo(c)
		if userInfo == nil {
			c.JSON(http.StatusUnauthorized, response.Unauthorized(""))
			c.Abort()
			return
		}

		hasRole := false
		for _, role := range roles {
			if userInfo.Role == role {
				hasRole = true
				break
			}
		}

		if !hasRole {
			c.JSON(http.StatusForbidden, response.Forbidden("权限不足"))
			c.Abort()
			return
		}

		c.Next(ctx)
	}
}

func GetUserInfo(c *app.RequestContext) *auth.UserInfo {
	if userInfo, ok := c.Get("userInfo"); ok {
		if u, ok := userInfo.(*auth.UserInfo); ok {
			return u
		}
	}
	return nil
}

func GetUserID(c *app.RequestContext) int64 {
	if userInfo := GetUserInfo(c); userInfo != nil {
		return userInfo.UserID
	}
	return 0
}

func GetOrgID(c *app.RequestContext) int64 {
	if userInfo := GetUserInfo(c); userInfo != nil {
		return userInfo.OrganizationID
	}
	return 0
}
