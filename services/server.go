package services

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	config "github.com/mohammedrefaat/hamber/Config"
	db "github.com/mohammedrefaat/hamber/Db"
	"github.com/mohammedrefaat/hamber/controllers"
	"github.com/mohammedrefaat/hamber/stores"
)

type Service struct {
	Router  *gin.Engine
	stStore *stores.DbStore
	config  *config.Config
}

func NewServer() (*Service, error) {
	config, err := config.LoadConfig("config.yaml")
	if err != nil {
		return nil, err
	}

	// Get the DSN from config
	dsn := config.GetDSN()

	// Connect to the database
	database, err := db.OpenDbConn(dsn)
	if err != nil {
		return nil, err
	}

	stStore, err := stores.NewDbStore(database)
	if err != nil {
		return nil, err
	}

	// Set the global store for controllers
	controllers.SetStore(stStore)

	router, err := GetRouter()
	if err != nil {
		return nil, err
	}

	serv := Service{
		stStore: stStore,
		Router:  router,
		config:  config,
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

	log.Printf("Server starting on port %s", c.config.Server.Port)
	// Start the server
	if err := s.ListenAndServe(); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}
