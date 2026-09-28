package config

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	ServiceName string
	Port        int
	Env         string
	DBDriver    string
	DBDSN       string
	JWTSecret   string
	AccessTTL   time.Duration
	RefreshTTL  time.Duration

	// Microservices Endpoints (for API Gateway & inter-service client)
	AuthServiceURL         string
	CustomerServiceURL     string
	AccountServiceURL      string
	TransactionServiceURL  string
	AuditServiceURL        string
	NotificationServiceURL string

	// gRPC Configuration
	GRPCPort             int
	NotificationGRPCAddr string
}

func init() {
	// Auto-load .env file if present in working directory or parent
	loadDotEnv(".env")
}

func LoadConfig(serviceName string, defaultPort int) *Config {
	port := defaultPort
	if p := os.Getenv("PORT"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil {
			port = parsed
		}
	}

	grpcPort := 50051
	if gp := os.Getenv("GRPC_PORT"); gp != "" {
		if parsed, err := strconv.Atoi(gp); err == nil {
			grpcPort = parsed
		}
	}

	driver := getEnv("DB_DRIVER", "mysql")

	dbUser := getEnv("DB_USER", "root")
	dbPass := getEnv("DB_PASSWORD", "password")
	dbHost := getEnv("DB_HOST", "127.0.0.1")
	dbPort := getEnv("DB_PORT", "3306")
	dbName := getEnv("DB_NAME", "banking")

	defaultDSN := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", dbUser, dbPass, dbHost, dbPort, dbName)
	if driver == "sqlite" {
		defaultDSN = "banking.db"
	}
	dsn := getEnv("DB_DSN", defaultDSN)

	return &Config{
		ServiceName:            serviceName,
		Port:                   port,
		GRPCPort:               grpcPort,
		Env:                    getEnv("APP_ENV", "development"),
		DBDriver:               driver,
		DBDSN:                  dsn,
		JWTSecret:              getEnv("JWT_SECRET", "super-secret-banking-jwt-key-2026-production-grade"),
		AccessTTL:              parseDuration(getEnv("JWT_ACCESS_EXPIRY", "60m"), 60*time.Minute),
		RefreshTTL:             parseDuration(getEnv("JWT_REFRESH_EXPIRY", "168h"), 7*24*time.Hour),
		AuthServiceURL:         getEnv("AUTH_SERVICE_URL", "http://localhost:8001"),
		CustomerServiceURL:     getEnv("CUSTOMER_SERVICE_URL", "http://localhost:8002"),
		AccountServiceURL:      getEnv("ACCOUNT_SERVICE_URL", "http://localhost:8003"),
		TransactionServiceURL:  getEnv("TRANSACTION_SERVICE_URL", "http://localhost:8004"),
		AuditServiceURL:        getEnv("AUDIT_SERVICE_URL", "http://localhost:8005"),
		NotificationServiceURL: getEnv("NOTIFICATION_SERVICE_URL", "http://localhost:8006"),
		NotificationGRPCAddr:   getEnv("NOTIFICATION_GRPC_ADDR", "localhost:50051"),
	}
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func parseDuration(val string, defaultVal time.Duration) time.Duration {
	d, err := time.ParseDuration(val)
	if err != nil {
		return defaultVal
	}
	return d
}

// loadDotEnv reads key=value pairs from a file into os.Setenv if not already set
func loadDotEnv(filepath string) {
	file, err := os.Open(filepath)
	if err != nil {
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			k := strings.TrimSpace(parts[0])
			v := strings.Trim(strings.TrimSpace(parts[1]), "\"'")
			if os.Getenv(k) == "" {
				_ = os.Setenv(k, v)
			}
		}
	}
}
