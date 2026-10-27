package v1

import (
	"stitchforce-server/internal/controller/http/middleware"
	"stitchforce-server/internal/usecase"
	"stitchforce-server/pkg/jwt"
	"stitchforce-server/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

func NewRouter(
	users usecase.UserUseCase,
	garmentWorkers usecase.GarmentWorkerUseCase,
	log *logger.Logger,
	jwtManager *jwt.Manager,
	validate *validator.Validate,
) *gin.Engine {
	r := gin.New()
	r.Use(
		middleware.Recovery(log),
		middleware.RequestLogger(log),
	)

	userHandler := NewUserHandler(users, validate, jwtManager)
	garmentWorkerHandler := NewGarmentWorkerHandler(
		garmentWorkers, validate,
	)

	v1 := r.Group("/v1")
	{
		auth := v1.Group("/auth")
		{
			auth.POST("/register", userHandler.Register)
			auth.POST("/login", userHandler.Login)
		}

		protected := v1.Group("")
		protected.Use(middleware.Auth(jwtManager))
		{
			usersRoute := protected.Group("/users")
			{
				usersRoute.GET("", userHandler.List)
				usersRoute.GET("/:id", userHandler.Get)
				usersRoute.PUT("/:id", userHandler.Update)
				usersRoute.DELETE("/:id", userHandler.Delete)
			}
			organizations := protected.Group("/organizations/:organizationID")
			{
				workers := organizations.Group("/garment-workers")
				{
					workers.POST("", garmentWorkerHandler.Create)
					workers.GET("", garmentWorkerHandler.List)
					workers.GET("/:id", garmentWorkerHandler.Get)
					workers.PUT("/:id", garmentWorkerHandler.Update)
					workers.DELETE("/:id", garmentWorkerHandler.Delete)
				}
			}
		}
	}

	return r
}
