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

func SetupRouter(repos *Repos) *gin.Engine {
	r := gin.Default()
	r.Use(middleware.ErrorHandlingMiddleware())

	userHandler := handlers.NewUserHandler(repos.UserRepo)
	authHandler := handlers.NewAuthHandler(repos.UserRepo)

	apiGroup := r.Group("/api")
	{
		apiGroup.POST("/auth/register", authHandler.RegisterUser)
		apiGroup.POST("/auth/login", authHandler.AuthenticateUser)

		authRequired := apiGroup.Group("/", middleware.AuthRequired(repos.UserRepo))
		{
			authRequired.GET("/auth/me", authHandler.GetAuthenticatedUser)
			authRequired.POST("/auth/refresh", authHandler.RefreshApiKey)

			authRequired.GET("/users", userHandler.GetUserByEmail) // TODO: RBAC, fix the idea that /users route only gets user by email lol
		}
	}

	return r
}
