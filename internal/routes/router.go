package routes

import (
	"qvarkk/kvault/internal/handlers"
	"qvarkk/kvault/internal/middleware"

	_ "qvarkk/kvault/docs"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type Services struct {
	AuthService     handlers.AuthService
	AuthUserService handlers.AuthUserService
	MwUserService   middleware.UserService
	UserService     handlers.UserService
	ItemService     handlers.ItemService
	FileService     handlers.FileService
}

func SetupRouter(services *Services) *gin.Engine {
	r := gin.Default()
	r.Use(middleware.ErrorHandlingMiddleware())

	authHandler := handlers.NewAuthHandler(services.AuthService, services.AuthUserService)
	userHandler := handlers.NewUserHandler(services.UserService)
	itemHandler := handlers.NewItemHandler(services.ItemService)
	fileHandler := handlers.NewFileHandler(services.FileService)

	apiGroup := r.Group("/api/v1")
	{
		apiGroup.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

		apiGroup.POST("/auth/register", authHandler.RegisterUser)
		apiGroup.POST("/auth/login", authHandler.AuthenticateUser)

		authRequired := apiGroup.Group("/", middleware.AuthRequired(services.MwUserService))
		{
			authRequired.GET("/auth/me", authHandler.GetAuthenticatedUser)
			authRequired.POST("/auth/refresh", authHandler.RotateApiKey)

			// TODO: RBAC, fix the idea that /users route only gets user by email lol
			authRequired.GET("/users", userHandler.GetByEmail)

			authRequired.POST("/items", itemHandler.Create)

			authRequired.POST("/files/upload", fileHandler.UploadFile)
		}
	}

	return r
}
