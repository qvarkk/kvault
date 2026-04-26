package routes

import (
	"qvarkk/kvault/internal/handlers"
	"qvarkk/kvault/internal/middleware"

	_ "qvarkk/kvault/docs"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type HandlerServices struct {
	Auth     handlers.AuthService
	AuthUser handlers.AuthUserService
	User     handlers.UserService
	Item     handlers.ItemService
	File     handlers.FileService
}

type MiddlewareServices struct {
	User middleware.UserService
}

func SetupRouter(hs *HandlerServices, ms *MiddlewareServices) *gin.Engine {
	r := gin.Default()
	r.Use(middleware.ErrorHandlingMiddleware())

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	api := r.Group("/api/v1")
	auth := middleware.AuthRequired(ms.User)

	registerAuthRoutes(api, auth, handlers.NewAuthHandler(hs.Auth, hs.AuthUser))
	registerUserRoutes(api, auth, handlers.NewUserHandler(hs.User))
	registerItemRoutes(api, auth, handlers.NewItemHandler(hs.Item))
	registerFileRoutes(api, auth, handlers.NewFileHandler(hs.File))

	return r
}

func registerAuthRoutes(api *gin.RouterGroup, auth gin.HandlerFunc, h *handlers.AuthHandler) {
	group := api.Group("/auth")
	group.POST("/register", handlers.APIWrap(h.RegisterUser))
	group.POST("/login", handlers.APIWrap(h.AuthenticateUser))

	protected := group.Group("/", auth)
	protected.GET("/me", handlers.APIWrap(h.GetAuthenticatedUser))
	protected.POST("/refresh", handlers.APIWrap(h.RotateApiKey))
}

func registerUserRoutes(api *gin.RouterGroup, auth gin.HandlerFunc, h *handlers.UserHandler) {
	group := api.Group("/users", auth)
	// TODO: RBAC, fix the idea that /users route only gets user by email lol
	group.GET("", handlers.APIWrap(h.GetByEmail))
}

func registerItemRoutes(api *gin.RouterGroup, auth gin.HandlerFunc, h *handlers.ItemHandler) {
	group := api.Group("/items", auth)
	group.POST("", handlers.APIWrap(h.Create))
	group.GET("", handlers.APIWrap(h.List))
	group.GET("/:id", handlers.APIWrap(h.Get))
}

func registerFileRoutes(api *gin.RouterGroup, auth gin.HandlerFunc, h *handlers.FileHandler) {
	group := api.Group("/files", auth)
	group.POST("/upload", handlers.APIWrap(h.UploadFile))
	group.GET("", handlers.APIWrap(h.List))
}
