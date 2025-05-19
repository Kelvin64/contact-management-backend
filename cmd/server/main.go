package main

import (
	"log"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"contactmanagement/internal/config"
	"contactmanagement/internal/handlers"
	"contactmanagement/internal/repository"
	"contactmanagement/internal/types"
)

func initDB(cfg *config.Config) *gorm.DB {
	db, err := gorm.Open(postgres.Open(cfg.Database.GetDatabaseURL()), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Auto migrate the schema
	err = db.AutoMigrate(&types.Contact{}, &types.Phone{})
	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	return db
}

func setupRouter(cfg *config.Config, contactHandler *handlers.ContactHandler) *gin.Engine {
	// Set Gin mode
	gin.SetMode(cfg.Server.GinMode)
	
	r := gin.Default()

	// Configure CORS
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowOrigins = cfg.CORS.AllowedOrigins
	corsConfig.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	corsConfig.AllowHeaders = []string{"Origin", "Content-Type", "Accept"}
	r.Use(cors.New(corsConfig))

	// API routes
	api := r.Group("/api")
	{
		contacts := api.Group("/contacts")
		{
			contacts.POST("", contactHandler.CreateContact)
			contacts.GET("", contactHandler.ListContacts)
			contacts.GET("/:id", contactHandler.GetContact)
			contacts.PUT("/:id", contactHandler.UpdateContact)
			contacts.DELETE("/:id", contactHandler.DeleteContact)
			contacts.POST("/import", contactHandler.ImportContacts)
		}
	}

	return r
}

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load configuration:", err)
	}

	// Initialize database
	db := initDB(cfg)

	// Initialize repository
	contactRepo := repository.NewContactRepository(db)

	// Initialize handler with repository
	contactHandler := handlers.NewContactHandler(contactRepo)

	// Setup router with handler
	r := setupRouter(cfg, contactHandler)
	
	// Start server
	log.Printf("Server starting on port %s", cfg.Server.Port)
	if err := r.Run(":" + cfg.Server.Port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
} 