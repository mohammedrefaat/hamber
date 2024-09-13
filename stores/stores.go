package stores

import (
	"gorm.io/gorm"
)

type DbStore struct {
	db *gorm.DB
	//Router *gin.Engine
}

func NewDbStore(db *gorm.DB) *DbStore {
	return &DbStore{db: db}
}
