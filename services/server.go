package services

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	config "github.com/mohammedrefaat/hamber/Config"
	db "github.com/mohammedrefaat/hamber/Db"
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
	db, err := db.OpenDbConn(dsn)
	if err != nil {
		return nil, err
	}
	stStore := stores.NewDbStore(db)
	if err != nil {
		return nil, err
	}
	router, err := GetRouter()

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
		ReadTimeout:    time.Duration(c.config.Server.ReadTimeout) * time.Second,
		WriteTimeout:   time.Duration(c.config.Server.WriteTimeout) * time.Second,
		MaxHeaderBytes: c.config.Server.MaxHeaderBytes,
	}

	// Start the server
	s.ListenAndServe()
}
