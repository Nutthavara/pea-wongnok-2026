package middleware

import (
	"crypto/subtle"
	"net/http"
	"wongnok/internal/httputil"

	"github.com/gin-gonic/gin"
)

const (
	basicAuthUser = "admin"
	basicAuthPass = "secret"
)

func BasicAuthMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		user, pass, ok := ctx.Request.BasicAuth()

		// ใข้ ConstantTimeCompare เพื่อป้อง attacker จับเวลาในการ compare ข้อมูล
		// เพื่อถ้าใช้ if user == basicAuthUser ระยะเวลาในการ compare ของข้อมูลที่ยาวไม่เท่ากันจะไม่ตรงกัน
		// เป็นช่องโว่ให้ attacker สามารถคาดเดาความยาวของ password และ brute force ได้ง่ายขึ้น
		// เพราะงั้น ConstantTimeCompare จะมาช่วยตรงนี้ โดยไม่ว่าความยาวข้อมูลที่ส่งเข้ามาจะเท่าไร่
		// มันจะใช้อะไร compare เท่าเดิมเสมอ
		validUser := subtle.ConstantTimeCompare([]byte(user), []byte(basicAuthUser)) == 1
		validPassword := subtle.ConstantTimeCompare([]byte(pass), []byte(basicAuthPass)) == 1

		if !ok || !validUser || !validPassword {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, httputil.ErrorResponse{
				Message: http.StatusText(http.StatusUnauthorized),
			})
			return
		}

		ctx.Next()
	}
}
