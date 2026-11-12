package v1

import (
	"stitchforce-server/internal/authorization"
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
	licenseUseCase usecase.LicenseUseCase,
	styleUseCase usecase.StyleUseCase,
	styleOperationUseCase usecase.StyleOperationUseCase,
	sizeUseCase usecase.SizeUseCase,
	bundleUseCase usecase.BundleUseCase,
	productionEntryUseCase usecase.ProductionEntryUseCase,
	workerPaymentUseCase usecase.WorkerPaymentUseCase,
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
	organizationHandler := NewOrganizationHandler(organizationUseCase, validate)
	userHandler := NewUserHandler(users, validate, jwtManager)
	styleHandler := NewStyleHandler(styleUseCase, validate)
	styleOperationHandler := NewStyleOperationHandler(styleOperationUseCase, validate)
	sizeHandler := NewSizeHandler(sizeUseCase, validate)
	bundleHandler := NewBundleHandler(bundleUseCase, validate)
	garmentWorkerHandler := NewGarmentWorkerHandler(garmentWorkers, validate)
	productionEntryHandler := NewProductionEntryHandler(productionEntryUseCase, validate)
	workerPaymentHandler := NewWorkerPaymentHandler(workerPaymentUseCase, validate)
	licenseHandler := NewLicenseHandler(licenseUseCase, validate)
	v1 := r.Group("/v1")
	{
		auth := v1.Group("/auth")
		{
			auth.POST("/register", userHandler.Register)
			auth.POST("/login", userHandler.Login)
			auth.POST("/logout", userHandler.Logout)
			auth.POST("/verify-email", userHandler.VerifyEmail)
			auth.POST("/accept-invitation", organizationHandler.AcceptInvitation)
		}

		protected := v1.Group("")
		protected.Use(middleware.Auth(jwtManager))
		{
			usersRoute := protected.Group("/users")
			{
				usersRoute.GET("/me", userHandler.GetMe)
				usersRoute.PATCH("/me", userHandler.Update)

				usersRoute.GET("/:id",
					middleware.RequirePermission(authorization.PermissionUsersView),
					userHandler.Get,
				)
				usersRoute.DELETE("/:id",
					middleware.RequirePermission(authorization.PermissionUsersManage),
					userHandler.Delete,
				)
			}

			organizations := protected.Group("/organizations/:organizationID")
			{
				// License activation: membership required, license can be expired.
				organizations.POST(
					"/license/activate",
					middleware.OrganizationAccess(organizationUseCase, false),
					licenseHandler.Activate,
				)

				organizations.GET("",
					middleware.OrganizationAccess(organizationUseCase, false),
					middleware.RequirePermission(authorization.PermissionOrganizationView),
					organizationHandler.GetByID,
				)

				// All normal organization routes require an active license and membership.
				licensed := organizations.Group("")
				licensed.Use(middleware.OrganizationAccess(organizationUseCase, true))
				{
					// Organization
					licensed.PUT("",
						middleware.RequirePermission(authorization.PermissionOrganizationUpdate),
						organizationHandler.Update,
					)

					licensed.DELETE("",
						middleware.RequirePermission(authorization.PermissionOrganizationUpdate),
						organizationHandler.Delete,
					)

					// Users and invitations
					licensed.GET("/users",
						middleware.RequirePermission(authorization.PermissionUsersView),
						userHandler.List,
					)

					licensed.POST("/invitations",
						middleware.RequirePermission(authorization.PermissionUsersManage),
						organizationHandler.Invite,
					)
					licensed.GET("/invitations",
						middleware.RequirePermission(authorization.PermissionUsersView),
						organizationHandler.ListInvitations,
					)
					licensed.POST("/invitations/:id/resend",
						middleware.RequirePermission(authorization.PermissionUsersManage),
						organizationHandler.ResendInvitation,
					)
					licensed.DELETE("/invitations/:id",
						middleware.RequirePermission(authorization.PermissionUsersManage),
						organizationHandler.CancelInvitation,
					)

					// Dashboard
					licensed.GET("/dashboard",
						middleware.RequirePermission(authorization.PermissionDashboardView),
						organizationHandler.GetDashboardOverview,
					)
					licensed.GET("/dashboard-trend",
						middleware.RequirePermission(authorization.PermissionDashboardView),
						organizationHandler.GetDashboardTrend,
					)

					// Styles
					styles := licensed.Group("/styles")
					{
						view := styles.Group("")
						view.Use(middleware.RequirePermission(authorization.PermissionStylesView))
						{
							view.GET("", styleHandler.List)
							view.GET("/:styleID", styleHandler.Get)
						}

						create := styles.Group("")
						create.Use(middleware.RequirePermission(authorization.PermissionStylesCreate))
						create.POST("", styleHandler.Create)

						update := styles.Group("")
						update.Use(middleware.RequirePermission(authorization.PermissionStylesUpdate))
						update.PUT("/:styleID", styleHandler.Update)

						delete := styles.Group("")
						delete.Use(middleware.RequirePermission(authorization.PermissionStylesDelete))
						delete.DELETE("/:styleID", styleHandler.Delete)

						// Style operations
						operations := styles.Group("/:styleID/operations")
						{
							view := operations.Group("")
							view.Use(middleware.RequirePermission(authorization.PermissionStylesView))
							{
								view.GET("", styleOperationHandler.List)
								view.GET("/:id", styleOperationHandler.Get)
							}

							create := operations.Group("")
							create.Use(middleware.RequirePermission(authorization.PermissionStylesCreate))
							create.POST("", styleOperationHandler.Create)

							update := operations.Group("")
							update.Use(middleware.RequirePermission(authorization.PermissionStylesUpdate))
							update.PUT("/:id", styleOperationHandler.Update)

							delete := operations.Group("")
							delete.Use(middleware.RequirePermission(authorization.PermissionStylesDelete))
							delete.DELETE("/:id", styleOperationHandler.Delete)
						}
					}

					// Sizes
					sizes := licensed.Group("/sizes")
					{
						view := sizes.Group("")
						view.Use(middleware.RequirePermission(authorization.PermissionStylesView))
						{
							view.GET("", sizeHandler.List)
							view.GET("/:id", sizeHandler.Get)
						}

						create := sizes.Group("")
						create.Use(middleware.RequirePermission(authorization.PermissionStylesCreate))
						create.POST("", sizeHandler.Create)

						update := sizes.Group("")
						update.Use(middleware.RequirePermission(authorization.PermissionStylesUpdate))
						update.PUT("/:id", sizeHandler.Update)

						delete := sizes.Group("")
						delete.Use(middleware.RequirePermission(authorization.PermissionStylesDelete))
						delete.DELETE("/:id", sizeHandler.Delete)
					}

					// Bundles
					bundles := licensed.Group("/bundles")
					{
						view := bundles.Group("")
						view.Use(middleware.RequirePermission(authorization.PermissionBundlesView))
						{
							view.GET("", bundleHandler.List)
							view.GET("/:id", bundleHandler.Get)
						}

						create := bundles.Group("")
						create.Use(middleware.RequirePermission(authorization.PermissionBundlesCreate))
						create.POST("", bundleHandler.Create)

						update := bundles.Group("")
						update.Use(middleware.RequirePermission(authorization.PermissionBundlesUpdate))
						update.PUT("/:id", bundleHandler.Update)

						delete := bundles.Group("")
						delete.Use(middleware.RequirePermission(authorization.PermissionBundlesDelete))
						delete.DELETE("/:id", bundleHandler.Delete)
					}

					// Garment workers
					workers := licensed.Group("/garment-workers")
					{
						view := workers.Group("")
						view.Use(middleware.RequirePermission(authorization.PermissionWorkersView))
						{
							view.GET("", garmentWorkerHandler.List)
							view.GET("/:id", garmentWorkerHandler.Get)
							view.GET("/:id/payment-summary", workerPaymentHandler.GetWorkerSummary)
						}

						create := workers.Group("")
						create.Use(middleware.RequirePermission(authorization.PermissionWorkersCreate))
						create.POST("", garmentWorkerHandler.Create)

						update := workers.Group("")
						update.Use(middleware.RequirePermission(authorization.PermissionWorkersUpdate))
						update.PUT("/:id", garmentWorkerHandler.Update)

						delete := workers.Group("")
						delete.Use(middleware.RequirePermission(authorization.PermissionWorkersDelete))
						delete.DELETE("/:id", garmentWorkerHandler.Delete)
					}

					// Worker payments
					workerPayments := licensed.Group("/worker-payments")
					{
						view := workerPayments.Group("")
						view.Use(middleware.RequirePermission(authorization.PermissionProductionView))
						{
							view.GET("/balances", workerPaymentHandler.ListWorkerBalances)
							view.GET("/summary", workerPaymentHandler.GetSummary)
							view.GET("", workerPaymentHandler.List)
							view.GET("/export", workerPaymentHandler.Export)
							view.GET("/:id", workerPaymentHandler.Get)
						}

						create := workerPayments.Group("")
						create.Use(middleware.RequirePermission(authorization.PermissionProductionCreate))
						create.POST("", workerPaymentHandler.Create)

						update := workerPayments.Group("")
						update.Use(middleware.RequirePermission(authorization.PermissionProductionUpdate))
						update.PATCH("/:id/void", workerPaymentHandler.Void)
					}

					// Production entries
					productionEntries := licensed.Group("/production-entries")
					{
						view := productionEntries.Group("")
						view.Use(middleware.RequirePermission(authorization.PermissionProductionView))
						{
							view.GET("/export", productionEntryHandler.Export)
							view.GET("", productionEntryHandler.List)
							view.GET("/:id", productionEntryHandler.Get)
						}

						create := productionEntries.Group("")
						create.Use(middleware.RequirePermission(authorization.PermissionProductionCreate))
						create.POST("", productionEntryHandler.Create)

						update := productionEntries.Group("")
						update.Use(middleware.RequirePermission(authorization.PermissionProductionUpdate))
						update.PUT("/:id", productionEntryHandler.Update)

						delete := productionEntries.Group("")
						delete.Use(middleware.RequirePermission(authorization.PermissionProductionDelete))
						delete.DELETE("/:id", productionEntryHandler.Delete)
					}
				}
			}
		}
	}

	return r
}
