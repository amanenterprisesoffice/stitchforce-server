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
	bundleUseCase usecase.BundleUseCase,
	productionEntryUseCase usecase.ProductionEntryUseCase,
	log *logger.Logger,
	jwtManager *jwt.Manager,
	validate *validator.Validate,
) *gin.Engine {
	r := gin.New()
	r.Use(
		middleware.Recovery(log),
		middleware.RequestLogger(log),
	)

	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Welcome to StitchForce API",
		})
	})
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "ok",
		})
	})

	userHandler := NewUserHandler(users, validate, jwtManager)
	styleHandler := NewStyleHandler(styleUseCase, validate)
	styleOperationHandler := NewStyleOperationHandler(styleOperationUseCase, validate)
	sizeHandler := NewSizeHandler(sizeUseCase, validate)
	bundleHandler := NewBundleHandler(bundleUseCase, validate)
	garmentWorkerHandler := NewGarmentWorkerHandler(garmentWorkers, validate)
	productionEntryHandler := NewProductionEntryHandler(productionEntryUseCase, validate)

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

				bundles := organizations.Group("/bundles")
				{
					bundles.POST("", bundleHandler.Create)
					bundles.GET("", bundleHandler.List)
					bundles.GET("/:id", bundleHandler.Get)
					bundles.PUT("/:id", bundleHandler.Update)
					bundles.DELETE("/:id", bundleHandler.Delete)
				}

				workers := organizations.Group("/garment-workers")
				{
					workers.POST("", garmentWorkerHandler.Create)
					workers.GET("", garmentWorkerHandler.List)
					workers.GET("/:id", garmentWorkerHandler.Get)
					workers.PUT("/:id", garmentWorkerHandler.Update)
					workers.PATCH("/:id/status", garmentWorkerHandler.SetStatus)
					workers.DELETE("/:id", garmentWorkerHandler.Delete)
				}
				productionEntries := organizations.Group("/production-entries")
				{
					productionEntries.POST("", productionEntryHandler.Create)
					productionEntries.GET("", productionEntryHandler.List)
					productionEntries.GET("/:id", productionEntryHandler.Get)
					productionEntries.PUT("/:id", productionEntryHandler.Update)
					productionEntries.DELETE("/:id", productionEntryHandler.Delete)
				}
			}
		}
	}

	return r
}
