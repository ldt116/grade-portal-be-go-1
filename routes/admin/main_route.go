package routes_admin

import (
	"LearnGo/config"
	middlewares_admin "LearnGo/middlewares/admin"

	"github.com/gin-gonic/gin"
)

func MainRoute(r *gin.Engine) {
	prefixAdmin := config.PrefixAdmin()
	// /admin/api
	// vo được mà không cần đăng nhập
	AuthRoute(r.Group(prefixAdmin))
	// middleware đảm bảo rằng đã đăng nhập trước khi vào web
	protectedGroup := r.Group(prefixAdmin)
	protectedGroup.Use(middlewares_admin.RequireAuth)
	// tạo các group để chạy các api sau khi đã đăng nhập thành công
	TeacherRoute(protectedGroup.Group("/teacher"))
	StudentRoute(protectedGroup.Group("/student"))
	ResultScoreRoute(protectedGroup.Group("/resultScore"))
	// tthem account vao database
	AccountRoute(protectedGroup.Group("/account"))
}
