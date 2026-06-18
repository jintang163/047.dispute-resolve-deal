package auth

import (
	"errors"
	"time"

	"github.com/dispute-resolve/common/config"
	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID         int64  `json:"userId"`
	Username       string `json:"username"`
	RealName       string `json:"realName"`
	Role           int32  `json:"role"`
	OrganizationID int64  `json:"organizationId"`
	jwt.RegisteredClaims
}

type UserInfo struct {
	UserID         int64  `json:"userId"`
	Username       string `json:"username"`
	RealName       string `json:"realName"`
	Role           int32  `json:"role"`
	OrganizationID int64  `json:"organizationId"`
}

var (
	ErrTokenExpired     = errors.New("token is expired")
	ErrTokenNotValidYet = errors.New("token not active yet")
	ErrTokenMalformed   = errors.New("that's not even a token")
	ErrTokenInvalid     = errors.New("couldn't handle this token")
)

func GenerateToken(user *UserInfo) (string, error) {
	cfg := config.GetConfig().JWT
	claims := Claims{
		UserID:         user.UserID,
		Username:       user.Username,
		RealName:       user.RealName,
		Role:           user.Role,
		OrganizationID: user.OrganizationID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(cfg.ExpireTime) * time.Second)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    cfg.Issuer,
			Subject:   user.Username,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(cfg.Secret))
}

func ParseToken(tokenString string) (*Claims, error) {
	cfg := config.GetConfig().JWT
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(cfg.Secret), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenMalformed) {
			return nil, ErrTokenMalformed
		} else if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		} else if errors.Is(err, jwt.ErrTokenNotValidYet) {
			return nil, ErrTokenNotValidYet
		}
		return nil, ErrTokenInvalid
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, ErrTokenInvalid
}

func RefreshToken(tokenString string) (string, error) {
	claims, err := ParseToken(tokenString)
	if err != nil && !errors.Is(err, ErrTokenExpired) {
		return "", err
	}

	return GenerateToken(&UserInfo{
		UserID:         claims.UserID,
		Username:       claims.Username,
		RealName:       claims.RealName,
		Role:           claims.Role,
		OrganizationID: claims.OrganizationID,
	})
}
