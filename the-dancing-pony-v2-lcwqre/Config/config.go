package config

import (
	"fmt"
	"log"
	utils "the-dancing-pony-v2-lcwqre/Utils"
	"the-dancing-pony-v2-lcwqre/model"
	"the-dancing-pony-v2-lcwqre/repository"
	"the-dancing-pony-v2-lcwqre/service"

	"github.com/go-playground/validator/v10"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func SetupDatabase(user string, password string, dbName string) (*gorm.DB, error) {
	// Construct the Data Source Name (DSN) for PostgreSQL
	dsn := fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable", user, password, dbName)

	// Open a connection to the database
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to the database: %v", err)

	}

	// Get the underlying SQL DB instance
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("Failed to get SQL DB instance: %v", err)
	}

	// Ensure the uuid-ossp extension is installed
	if _, err := sqlDB.Exec(`CREATE EXTENSION IF NOT EXISTS "uuid-ossp";`); err != nil {
		log.Fatalf("Failed to create uuid-ossp extension: %v", err)
	}

	// Perform database migrations
	err = db.AutoMigrate(&model.Dish{}, &model.Rating{}, &model.User{}, &model.Permission{}, &model.Restaurant{})
	if err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	fmt.Println("Database migrated successfully")
	initializePermissions(db)
	return db, nil
}
func initializePermissions(db *gorm.DB) {
	permissions := []model.Permission{
		{Name: "customer"},
		{Name: "restaurant"},
		{Name: "admin"},
	}

	for _, perm := range permissions {
		// Check if permission already exists
		var existingPerm model.Permission
		err := db.Where("name = ?", perm.Name).First(&existingPerm).Error
		if err != nil && err != gorm.ErrRecordNotFound {
			log.Fatalf("Failed to check permission %s: %v", perm.Name, err)
		}

		if err == gorm.ErrRecordNotFound {
			// Permission does not exist, create it
			if err := db.Create(&perm).Error; err != nil {
				log.Fatalf("Failed to create permission %s: %v", perm.Name, err)
			} else {
				fmt.Printf("Created permission: %s\n", perm.Name)
			}
		} else {
			fmt.Printf("Permission already exists: %s\n", perm.Name)
		}
	}
}

// initializeServices sets up the dish and auth services
func InitializeServices(db *gorm.DB, validate *validator.Validate, s3uploader *utils.S3Uploader) (service.DishesService, service.AuthService, service.RestaurantsService) {
	dishRepository := repository.NewDishesRepositoryImpl(db)
	resturantRepository := repository.NewRestaurantsRepositoryImpl(db)
	userRepo := repository.NewUserRepository(db)
	dishService := service.NewDishesServiceImpl(dishRepository, validate, s3uploader)
	resturantService := service.NewRestaurantsServiceImpl(resturantRepository, validate)
	authService := service.NewAuthService(userRepo)
	return dishService, authService, resturantService
}
