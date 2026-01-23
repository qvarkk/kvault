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
	r.POST("/users", userHandler.CreateUser)

	return r
}
