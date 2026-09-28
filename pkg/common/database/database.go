package database

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/glebarez/sqlite"
	_ "github.com/go-sql-driver/mysql"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Connect initializes a GORM database connection supporting MySQL and SQLite
func Connect(driver, dsn string) (*gorm.DB, error) {
	var dialector gorm.Dialector

	switch driver {
	case "mysql":
		if err := ensureMySQLDatabase(dsn); err != nil {
			log.Printf("[Database] Note on ensuring MySQL database: %v", err)
		}
		dialector = mysql.Open(dsn)
	case "sqlite":
		dialector = sqlite.Open(dsn)
	default:
		return nil, fmt.Errorf("unsupported database driver: %s", driver)
	}

	db, err := gorm.Open(dialector, &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database (%s): %w", driver, err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get sql.DB: %w", err)
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(50)
	sqlDB.SetConnMaxLifetime(time.Hour)
	stats := sqlDB.Stats()

	log.Printf("Max Open Connections: %d", stats.MaxOpenConnections)
	log.Printf("Open Connections: %d", stats.OpenConnections)
	log.Printf("In Use: %d", stats.InUse)
	log.Printf("Idle: %d", stats.Idle)
	log.Printf("[Database] Connected successfully using %s driver", driver)
	return db, nil
}

// ensureMySQLDatabase automatically creates the MySQL database if it does not already exist
func ensureMySQLDatabase(dsn string) error {
	slashIdx := strings.LastIndex(dsn, "/")
	if slashIdx == -1 {
		return nil
	}

	qIdx := strings.Index(dsn[slashIdx:], "?")
	var dbName string
	var serverDSN string

	if qIdx == -1 {
		dbName = dsn[slashIdx+1:]
		serverDSN = dsn[:slashIdx+1]
	} else {
		dbName = dsn[slashIdx+1 : slashIdx+qIdx]
		serverDSN = dsn[:slashIdx+1] + dsn[slashIdx+qIdx:]
	}

	if dbName == "" {
		return nil
	}

	serverDB, err := sql.Open("mysql", serverDSN)
	if err != nil {
		return err
	}
	defer serverDB.Close()

	query := fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci", dbName)
	_, err = serverDB.Exec(query)
	if err != nil {
		return fmt.Errorf("unable to execute CREATE DATABASE: %w", err)
	}

	log.Printf("[Database] Verified MySQL database `%s` exists", dbName)
	return nil
}
