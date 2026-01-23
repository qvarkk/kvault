package routes

import (
	"qvarkk/kvault/internal/handlers"
	"qvarkk/kvault/internal/repo"

	"github.com/gin-gonic/gin"
)

type Repos struct {
	UserRepo *repo.UserRepo
}

func SetupRouter(repos *Repos) *gin.Engine {
	r := gin.Default()

	userHandler := handlers.NewUserHandler(repos.UserRepo)

	// r.GET("/users", userHandler.GetUserByEmail) TODO: RBAC
	r.POST("/auth/register", userHandler.CreateUser)
	r.POST("/auth/login", userHandler.AuthenticateUser)

	return r
}
