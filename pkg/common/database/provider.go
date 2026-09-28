package database

import (
	"banking-microservices/pkg/common/config"

	"github.com/google/wire"
	"gorm.io/gorm"
)

// ProviderSet provides database dependencies
var ProviderSet = wire.NewSet(ProvideDB)

// ProvideDB initializes database using configuration
func ProvideDB(cfg *config.Config) (*gorm.DB, error) {
	return Connect(cfg.DBDriver, cfg.DBDSN)
}
