package routes

import (
	"qvarkk/kvault/internal/handlers/web"
	"qvarkk/kvault/internal/middleware"

	_ "qvarkk/kvault/docs" // swagger routes

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type HandlerServices struct {
	Auth     web.AuthService
	AuthUser web.AuthUserService
	Item     web.ItemService
	File     web.FileService
	Tag      web.TagService
}

type MiddlewareServices struct {
	User middleware.UserService
}

func SetupRouter(hs *HandlerServices, ms *MiddlewareServices, corsOrigins []string) *gin.Engine {
	r := gin.Default()
	r.Use(cors.New(cors.Config{
		AllowOrigins:     corsOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "Accept-Language"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))
	r.Use(middleware.ErrorHandlingMiddleware())

	api := r.Group("/api/v1")
	api.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	auth := middleware.AuthRequired(ms.User)

	registerAuthRoutes(api, auth, web.NewAuthHandler(hs.Auth, hs.AuthUser, hs.File))
	registerItemRoutes(api, auth, web.NewItemHandler(hs.Item))
	registerFileRoutes(api, auth, web.NewFileHandler(hs.File))
	registerTagRoutes(api, auth, web.NewTagHandler(hs.Tag))

	return r
}

func registerAuthRoutes(api *gin.RouterGroup, auth gin.HandlerFunc, h AuthHandler) {
	group := api.Group("/auth")

	// Throttle credential endpoints per IP to blunt brute force / signup spam.
	authLimit := middleware.AuthRateLimit()
	group.POST("/register", authLimit, web.APIWrap(h.RegisterUser))
	group.POST("/login", authLimit, web.APIWrap(h.AuthenticateUser))

	protected := group.Group("/", auth)
	protected.GET("/me", web.APIWrap(h.GetAuthenticatedUser))
	protected.PATCH("/me/password", web.APIWrap(h.ChangePassword))
	protected.POST("/me/delete", web.APIWrap(h.DeleteAccount))

	protected.GET("/keys", web.APIWrap(h.ListKeys))
	protected.PATCH("/keys/:id", web.APIWrap(h.RenameKey))
	protected.DELETE("/keys/:id", web.APIWrap(h.DeleteKey))
	protected.POST("/logout", web.APIWrap(h.Logout))
	protected.POST("/logout-others", web.APIWrap(h.LogoutOthers))
}

func registerItemRoutes(api *gin.RouterGroup, auth gin.HandlerFunc, h ItemHandler) {
	group := api.Group("/items", auth)
	group.POST("", web.APIWrap(h.Create))
	group.GET("", web.APIWrap(h.List))
	group.GET("/deleted", web.APIWrap(h.ListDeleted))
	group.DELETE("/deleted", web.APIWrap(h.ClearSoftDeleted))
	group.DELETE("/deleted/:id", web.APIWrap(h.PermanentlyDelete))
	group.GET("/:id", web.APIWrap(h.Get))
	group.PATCH("/:id", web.APIWrap(h.Update))
	group.DELETE("/:id", web.APIWrap(h.SoftDelete))
	group.POST("/:id/restore", web.APIWrap(h.Restore))

	group.POST("/:id/tags", web.APIWrap(h.AttachTag))
	group.DELETE("/:id/tags/:tag_id", web.APIWrap(h.DetachTag))
}

func registerFileRoutes(api *gin.RouterGroup, auth gin.HandlerFunc, h FileHandler) {
	group := api.Group("/files", auth)
	group.POST("/upload", web.APIWrap(h.UploadFile))
	group.GET("", web.APIWrap(h.List))
	group.GET("/deleted", web.APIWrap(h.ListDeleted))
	group.DELETE("/deleted", web.APIWrap(h.ClearTrash))
	group.DELETE("/deleted/:id", web.APIWrap(h.PermanentlyDelete))
	group.GET("/:id", web.APIWrap(h.Download))
	group.GET("/:id/view", web.APIWrap(h.GetViewURL))
	group.GET("/:id/info", web.APIWrap(h.GetInfo))
	group.DELETE("/:id", web.APIWrap(h.Delete))
	group.POST("/:id/restore", web.APIWrap(h.Restore))
}

func registerTagRoutes(api *gin.RouterGroup, auth gin.HandlerFunc, h TagHandler) {
	group := api.Group("/tags", auth)
	group.POST("", web.APIWrap(h.Create))
	group.GET("", web.APIWrap(h.List))
	group.PATCH("/:id", web.APIWrap(h.Update))
	group.DELETE("/:id", web.APIWrap(h.Delete))
}
