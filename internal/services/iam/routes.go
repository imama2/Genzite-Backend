package iam

import "github.com/gin-gonic/gin"

func (m *Module) RegisterRoutes(router *gin.RouterGroup, middlewares ...gin.HandlerFunc) {
	public := router.Group("/auth")
	public.POST("/register", m.controller.Register)
	public.POST("/login", m.controller.Login)
	public.GET("/google/login", m.controller.GoogleLogin)
	public.GET("/google/callback", m.controller.GoogleCallback)

	protected := router.Group("/auth", middlewares...)
	protected.GET("/me", m.controller.Me)
}
