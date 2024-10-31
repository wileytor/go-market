package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/wileytor/go-market/auth/internal/server"
)

func SetupAuthRoutes(s *server.Server) *gin.Engine {
	r := gin.Default()

	userGroup := r.Group("/users")
	{
		userGroup.GET(":id", s.GetUserProfileHandler)
		userGroup.POST("/register", s.RegisterUserHandler)
		userGroup.POST("/login", s.LoginUserHandler)
	}
	return r
}
