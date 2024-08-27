package router

import (
	"net/http"
	"time"

	"the-dancing-pony-v2-lcwqre/controller"
	"the-dancing-pony-v2-lcwqre/middleware"
	"the-dancing-pony-v2-lcwqre/repository"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func NewRouter(
	dishController *controller.DishesController,
	authController *controller.AuthController,
	resturantController *controller.RestaurantsController,
	userRepo repository.UserRepository,
) *gin.Engine {
	router := gin.New()

	// Global middleware
	router.Use(
		gin.Recovery(), // Recovery middleware provided by Gin for handling panics
		middleware.MetricsMiddleware(),
	)

	// Initialize user based rate limiter
	rateLimiter := middleware.NewRateLimiter(10*time.Second, 5)

	// Initialize user based rate limiter
	iPrateLimiter := middleware.NewIPRateLimiter(10*time.Second, 5)

	// Prometheus metrics endpoint
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Welcome route
	router.GET("/", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, "Welcome home")
	})

	// API route group
	apiRouter := router.Group("/api")
	apiRouter.Use(middleware.RequestUniqueId())

	// Authentication routes
	// Authentication routes with restaurantId
	authRouter := apiRouter.Group("/restaurants/:restaurantId/auth")
	authRouter.Use(middleware.MultiTenantRouting(), iPrateLimiter.Limit())
	{
		authRouter.POST("/register", authController.Register)
		authRouter.POST("/login", authController.Login)
	}

	// Customer and admin routes
	dishesRouter := apiRouter.Group("/restaurants/:restaurantId/dishes")
	dishesRouter.Use(middleware.AuthMiddleware(userRepo), middleware.PermissionMiddleware("customer", "admin", "restaurant"), middleware.MultiTenantRouting(), rateLimiter.Limit())
	{
		dishesRouter.GET("", dishController.FindAll)
		dishesRouter.GET("/:dishId", dishController.FindById)
		dishesRouter.GET("/search", dishController.Search)
		dishesRouter.POST("/:dishId/rate", dishController.RateDish)
	}

	// Admin-only routes
	adminDishesRouter := apiRouter.Group("/restaurants/:restaurantId/dishes/admin")
	adminDishesRouter.Use(middleware.AuthMiddleware(userRepo), middleware.PermissionMiddleware("restaurant", "admin"), middleware.MultiTenantRouting(), rateLimiter.Limit())
	{
		adminDishesRouter.POST("/", dishController.Create)
		adminDishesRouter.PATCH("/:dishId", dishController.Update)
		adminDishesRouter.DELETE("/:dishId", dishController.Delete)
	}

	restaurantRouter := apiRouter.Group("/restaurants")
	restaurantRouter.POST("/", resturantController.Create)
	restaurantRouter.GET("/:restaurantId", resturantController.FindById)
	restaurantRouter.GET("", resturantController.FindAll)
	restaurantRouter.PATCH("/:restaurantId", resturantController.Update)
	restaurantRouter.DELETE("/:restaurantId", resturantController.Delete)

	return router
}
