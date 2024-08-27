package main

import (
	"net/http"
	"os"

	"github.com/joho/godotenv"

	"github.com/go-playground/validator/v10"
	"github.com/rs/zerolog/log"

	config "the-dancing-pony-v2-lcwqre/Config"
	utils "the-dancing-pony-v2-lcwqre/Utils"
	cache "the-dancing-pony-v2-lcwqre/caching"
	"the-dancing-pony-v2-lcwqre/controller"
	"the-dancing-pony-v2-lcwqre/repository"
	"the-dancing-pony-v2-lcwqre/router"
)

func main() {
	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		log.Error().Err(err).Msg("Error loading .env file")
	}

	// Retrieve environment variables
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Default port
	}

	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")

	redisAddr := os.Getenv("REDIS_ADDR")
	redisPassword := os.Getenv("REDIS_PASSWORD")

	log.Info().Msgf("Starting server on port %s", port)

	// Initialize database
	db, err := config.SetupDatabase(dbUser, dbPassword, dbName)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to set up database")
	}
	log.Info().Msgf("Database succesfully set up")
	// Initialize Redis client
	if err := cache.InitRedisClient(redisAddr, redisPassword); err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize Redis client")
	}
	log.Info().Msgf("Redis client initialized")

	accessKey := os.Getenv("AWS_ACCESS_KEY_ID")
	secretKey := os.Getenv("AWS_SECRET_ACCESS_KEY")
	region := os.Getenv("AWS_REGION")
	// Initialize the S3 uploader
	uploader, err := utils.NewS3Uploader(accessKey, secretKey, region)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize S3 uploader:")

	}

	// Create validator instance
	validate := validator.New()

	// Initialize services
	dishService, authService, resturantService := config.InitializeServices(db, validate, uploader)

	// Initialize controllers
	dishController := controller.NewDishesController(dishService)
	authController := controller.NewAuthController(authService)
	restaurantController := controller.NewRestaurantsController(resturantService)

	// Setup router
	routes := router.NewRouter(dishController, authController, restaurantController, repository.NewUserRepository(db))

	// Start server
	server := &http.Server{
		Addr:    ":" + port,
		Handler: routes,
	}

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal().Err(err).Msg("Server failed")
	}
}
