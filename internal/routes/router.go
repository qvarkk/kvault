package routes

import (
	"qvarkk/kvault/internal/handlers"
	"qvarkk/kvault/internal/middleware"

	"github.com/gin-gonic/gin"
)

type Services struct {
	AuthService     handlers.AuthService
	AuthUserService handlers.AuthUserService
	MwUserService   middleware.UserService
	UserService     handlers.UserService
}

func SetupRouter(services *Services) *gin.Engine {
	r := gin.Default()
	r.Use(middleware.ErrorHandlingMiddleware())

	userHandler := handlers.NewUserHandler(services.UserService)
	authHandler := handlers.NewAuthHandler(services.AuthService, services.AuthUserService)

	apiGroup := r.Group("/api")
	{
		apiGroup.POST("/auth/register", authHandler.RegisterUser)
		apiGroup.POST("/auth/login", authHandler.AuthenticateUser)

		authRequired := apiGroup.Group("/", middleware.AuthRequired(services.MwUserService))
		{
			authRequired.GET("/auth/me", authHandler.GetAuthenticatedUser)
			authRequired.POST("/auth/refresh", authHandler.RotateApiKey)

			// TODO: RBAC, fix the idea that /users route only gets user by email lol
			authRequired.GET("/users", userHandler.GetByEmail)
		}
	}

	return r
}
