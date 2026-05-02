package routes

import "github.com/gin-gonic/gin"

type AuthHandler interface {
	RegisterUser(*gin.Context) error
	AuthenticateUser(*gin.Context) error
	GetAuthenticatedUser(*gin.Context) error
	RotateApiKey(*gin.Context) error
}

type UserHandler interface {
	GetByEmail(*gin.Context) error
}

type ItemHandler interface {
	Create(*gin.Context) error
	List(*gin.Context) error
	Get(*gin.Context) error
	Update(*gin.Context) error
	Delete(*gin.Context) error
	Restore(*gin.Context) error
}

type FileHandler interface {
	UploadFile(*gin.Context) error
	List(*gin.Context) error
	Download(*gin.Context) error
	Delete(*gin.Context) error
	Restore(*gin.Context) error
}

type StopwordHandler interface {
	Create(*gin.Context) error
	List(*gin.Context) error
	Enable(*gin.Context) error
	Disable(*gin.Context) error
	Delete(*gin.Context) error
}

type TagHandler interface {
	List(ctx *gin.Context) error
}
