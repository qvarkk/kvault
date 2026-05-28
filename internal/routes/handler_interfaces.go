package routes

import "github.com/gin-gonic/gin"

type AuthHandler interface {
	RegisterUser(*gin.Context) error
	AuthenticateUser(*gin.Context) error
	GetAuthenticatedUser(*gin.Context) error
	RotateApiKey(*gin.Context) error
	ChangePassword(*gin.Context) error
	DeleteAccount(*gin.Context) error
}

type ItemHandler interface {
	Create(*gin.Context) error
	List(*gin.Context) error
	ListDeleted(*gin.Context) error
	Get(*gin.Context) error
	Update(*gin.Context) error
	Delete(*gin.Context) error
	Restore(*gin.Context) error
	PermanentlyDelete(*gin.Context) error
	ClearTrash(*gin.Context) error
	BindTag(*gin.Context) error
	UnbindTag(*gin.Context) error
	Autotag(*gin.Context) error
	Refetch(*gin.Context) error
}

type FileHandler interface {
	UploadFile(*gin.Context) error
	List(*gin.Context) error
	ListDeleted(*gin.Context) error
	Download(*gin.Context) error
	GetViewURL(*gin.Context) error
	Delete(*gin.Context) error
	Restore(*gin.Context) error
	PermanentlyDelete(*gin.Context) error
	ClearTrash(*gin.Context) error
}

type StopwordHandler interface {
	Create(*gin.Context) error
	List(*gin.Context) error
	Enable(*gin.Context) error
	Disable(*gin.Context) error
	Delete(*gin.Context) error
}

type TagHandler interface {
	Create(*gin.Context) error
	List(*gin.Context) error
	Update(*gin.Context) error
	Delete(*gin.Context) error
}
