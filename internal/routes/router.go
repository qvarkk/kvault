package routes

import (
	"qvarkk/kvault/internal/handlers/web"
	"qvarkk/kvault/internal/middleware"

	_ "qvarkk/kvault/docs"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type HandlerServices struct {
	Auth     web.AuthService
	AuthUser web.AuthUserService
	User     web.UserService
	Item     web.ItemService
	File     web.FileService
	Stopword web.StopwordService
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

	registerAuthRoutes(api, auth, web.NewAuthHandler(hs.Auth, hs.AuthUser))
	registerUserRoutes(api, auth, web.NewUserHandler(hs.User))
	registerItemRoutes(api, auth, web.NewItemHandler(hs.Item))
	registerFileRoutes(api, auth, web.NewFileHandler(hs.File))
	registerStopwordRoutes(api, auth, web.NewStopwordHandler(hs.Stopword))

	return r
}

func registerAuthRoutes(api *gin.RouterGroup, auth gin.HandlerFunc, h *web.AuthHandler) {
	group := api.Group("/auth")
	group.POST("/register", web.APIWrap(h.RegisterUser))
	group.POST("/login", web.APIWrap(h.AuthenticateUser))

	protected := group.Group("/", auth)
	protected.GET("/me", web.APIWrap(h.GetAuthenticatedUser))
	protected.POST("/refresh", web.APIWrap(h.RotateApiKey))
}

func registerUserRoutes(api *gin.RouterGroup, auth gin.HandlerFunc, h *web.UserHandler) {
	group := api.Group("/users", auth)
	// TODO: RBAC, fix the idea that /users route only gets user by email lol
	group.GET("", web.APIWrap(h.GetByEmail))
}

func registerItemRoutes(api *gin.RouterGroup, auth gin.HandlerFunc, h *web.ItemHandler) {
	group := api.Group("/items", auth)
	group.POST("", web.APIWrap(h.Create))
	group.GET("", web.APIWrap(h.List))
	group.GET("/:id", web.APIWrap(h.Get))
	group.PATCH("/:id", web.APIWrap(h.Update))
	group.DELETE("/:id", web.APIWrap(h.Delete))
	group.POST("/:id/restore", web.APIWrap(h.Restore))
}

func registerFileRoutes(api *gin.RouterGroup, auth gin.HandlerFunc, h *web.FileHandler) {
	group := api.Group("/files", auth)
	group.POST("/upload", web.APIWrap(h.UploadFile))
	group.GET("", web.APIWrap(h.List))
	group.GET("/:id", web.APIWrap(h.Download))
	group.DELETE("/:id", web.APIWrap(h.Delete))
	group.POST("/:id/restore", web.APIWrap(h.Restore))
}

func registerStopwordRoutes(api *gin.RouterGroup, auth gin.HandlerFunc, h *web.StopwordHandler) {
	group := api.Group("/stopwords", auth)
	group.POST("", web.APIWrap(h.Create))
	group.GET("", web.APIWrap(h.List))
	group.POST("/:word/enable", web.APIWrap(h.Enable))
	group.POST("/:word/disable", web.APIWrap(h.Disable))
}
