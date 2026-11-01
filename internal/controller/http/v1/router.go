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
	organizationUseCase usecase.OrganizationUseCase,
	styleUseCase usecase.StyleUseCase,
	styleOperationUseCase usecase.StyleOperationUseCase,
	sizeUseCase usecase.SizeUseCase,
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
	garmentWorkerHandler := NewGarmentWorkerHandler(garmentWorkers, validate)
	styleHandler := NewStyleHandler(styleUseCase, validate)
	styleOperationHandler := NewStyleOperationHandler(styleOperationUseCase, validate)
	sizeHandler := NewSizeHandler(sizeUseCase, validate)

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
				usersRoute.GET("/me", userHandler.GetMe)
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
					workers.PATCH("/:id/status", garmentWorkerHandler.SetStatus)
					workers.DELETE("/:id", garmentWorkerHandler.Delete)
				}

				styles := organizations.Group("/styles")
				{
					styles.POST("", styleHandler.Create)
					styles.GET("", styleHandler.List)
					styles.GET("/:styleID", styleHandler.Get)
					styles.PUT("/:styleID", styleHandler.Update)
					styles.DELETE("/:styleID", styleHandler.Delete)

					operations := styles.Group("/:styleID/operations")
					{
						operations.POST("", styleOperationHandler.Create)
						operations.GET("", styleOperationHandler.List)
						operations.GET("/:id", styleOperationHandler.Get)
						operations.PUT("/:id", styleOperationHandler.Update)
						operations.DELETE("/:id", styleOperationHandler.Delete)
					}
				}

				sizes := organizations.Group("/sizes")
				{
					sizes.POST("", sizeHandler.Create)
					sizes.GET("", sizeHandler.List)
					sizes.GET("/:id", sizeHandler.Get)
					sizes.PUT("/:id", sizeHandler.Update)
					sizes.DELETE("/:id", sizeHandler.Delete)
				}
			}
		}
	}

	return r
}
