package helper

import (
	"errors"
	"github.com/gin-gonic/gin"
)

func CheckUserType(ctx *gin.Context, role string) error {
	userType := ctx.GetString("user_type")
	if userType != role {
		return errors.New("Unauthorized to access this resource :)")
	}
	return nil
}
func MatchUserTypeToUserId(c *gin.Context, userId string) error {
	userType := c.GetString("user_type")
	uId := c.GetString("uid")

	if userType == "USER" && uId != userId {
		return errors.New("Unauthorized to access this resource :)")
	}
	return CheckUserType(c, userType)
}
