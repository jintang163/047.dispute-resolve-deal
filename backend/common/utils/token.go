package utils

import (
	"github.com/dispute-resolve/common/auth"
)

func GenerateToken(userID int64, username string, role int, orgID int64) (string, error) {
	return auth.GenerateToken(&auth.UserInfo{
		UserID:         userID,
		Username:       username,
		RealName:       username,
		Role:           int32(role),
		OrganizationID: orgID,
	})
}
