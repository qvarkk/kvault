package routes

import (
	"qvarkk/kvault/internal/handlers"
	"qvarkk/kvault/internal/middleware"
	"qvarkk/kvault/internal/repo"

	"github.com/gin-gonic/gin"
)

type Repos struct {
	UserRepo *repo.UserRepo
}

type Services struct {
	AuthService handlers.AuthService
}

func SetupRouter(repos *Repos, services *Services) *gin.Engine {
	r := gin.Default()
	r.Use(middleware.ErrorHandlingMiddleware())

	userHandler := handlers.NewUserHandler(repos.UserRepo)
	authHandler := handlers.NewAuthHandler(services.AuthService)

	apiGroup := r.Group("/api")
	{
		apiGroup.POST("/auth/register", authHandler.RegisterUser)
		apiGroup.POST("/auth/login", authHandler.AuthenticateUser)

		authRequired := apiGroup.Group("/", middleware.AuthRequired(repos.UserRepo))
		{
			authRequired.GET("/auth/me", authHandler.GetAuthenticatedUser)
			authRequired.POST("/auth/refresh", authHandler.RotateApiKey)

			// TODO: RBAC, fix the idea that /users route only gets user by email lol
			authRequired.GET("/users", userHandler.GetUserByEmail)
		}
	}

	return r
}
