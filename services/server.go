// services/server.go - Updated version

package services

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	config "github.com/mohammedrefaat/hamber/Config"
	db "github.com/mohammedrefaat/hamber/Db"
	"github.com/mohammedrefaat/hamber/controllers"
	"github.com/mohammedrefaat/hamber/notification"
	"github.com/mohammedrefaat/hamber/stores"
	"github.com/mohammedrefaat/hamber/utils"
)

type Service struct {
	Router       *gin.Engine
	StStore      *stores.DbStore
	config       *config.Config
	photosrv     *db.PhotoSrv
	notifService *notification.NotificationService
}

func NewServer() (*Service, error) {
	config, err := config.LoadConfig("config.yaml")
	if err != nil {
		return nil, err
	}

	// Set config for JWT utilities
	utils.SetConfig(config)

	// Get the DSN from config
	dsn := config.GetDSN()

	// Initialize photo service
	err = InitPhotoService(config)
	if err != nil {
		return nil, err
	}

	// Connect to the database
	database, err := db.OpenDbConn(dsn)
	if err != nil {
		return nil, err
	}
	StStore, err := stores.NewDbStore(database)
	if err != nil {
		return nil, err
	}

	// Seed the database
	/*if err := dbmodels.SeedDatabase(database); err != nil {
		log.Fatal("Seeding failed:", err)
	}*/
	// Initialize notification service
	var notifService *notification.NotificationService
	if config.IsRabbitMQEnabled() {
		rabbitMQURL := config.GetRabbitMQURL()
		notifService, err = notification.NewNotificationService(rabbitMQURL, StStore)
		if err != nil {
			log.Printf("⚠️ Warning: Failed to initialize notification service: %v", err)
			log.Println("⚠️ Continuing without notification service...")
		} else {
			log.Println("✓ Notification service initialized successfully")
		}
	} else {
		log.Println("ℹ️ RabbitMQ is disabled in configuration")
	}

	// Set the global store for controllers
	controllers.SetStore(&controllers.GlobalService{
		StStore:      StStore,
		Config:       config,
		PhotoSrv:     GetPhotoService(),
		NotifService: notifService,
	})

	router, err := GetRouter(config)
	if err != nil {
		return nil, err
	}

	serv := Service{
		StStore:      StStore,
		Router:       router,
		config:       config,
		photosrv:     GetPhotoService(),
		notifService: notifService,
	}

	return &serv, nil
}

func (c *Service) Run() {
	// Create a custom HTTP server using the config values
	s := &http.Server{
		Addr:           c.config.Server.Port,
		Handler:        c.Router,
		ReadTimeout:    c.config.GetReadTimeout(),
		WriteTimeout:   c.config.GetWriteTimeout(),
		MaxHeaderBytes: c.config.Server.MaxHeaderBytes,
	}

	log.Printf("🚀 Server starting on port %s", c.config.Server.Port)
	log.Printf("📚 Swagger documentation available at: http://localhost%s/swagger/index.html", c.config.Server.Port)

	// Start the server
	if err := s.ListenAndServe(); err != nil {
		log.Fatal("❌ Server failed to start:", err)
	}
}

func (c *Service) Shutdown() {
	if c.notifService != nil {
		c.notifService.Close()
	}
	log.Println("🛑 Server shutdown complete")
}

var GlobalPhotoService *db.PhotoSrv

// InitPhotoService initializes the global photo service
func InitPhotoService(cfg *config.Config) error {
	// Initialize MinIO photo service
	photoService, err := db.NewPhotoService(
		cfg.Storage.MinIO.Endpoint,
		cfg.Storage.MinIO.AccessKey,
		cfg.Storage.MinIO.SecretKey,
		cfg.Storage.MinIO.Bucket,
		cfg.Storage.MinIO.UseSSL,
	)
	if err != nil {
		return err
	}

	GlobalPhotoService = photoService
	log.Println("✓ Photo service initialized successfully")
	return nil
}

// GetPhotoService returns the global photo service instance
func GetPhotoService() *db.PhotoSrv {
	if GlobalPhotoService == nil {
		log.Fatal("❌ Photo service not initialized. Call InitPhotoService first.")
	}
	return GlobalPhotoService
}
