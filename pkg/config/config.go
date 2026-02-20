package config

import (
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Env         string
	Port        string
	DatabaseURL string
	JWTSecret   string
	JWTExpiry   time.Duration
	//  S3/MinIO fields
	S3Endpoint     string
	S3Region       string
	S3AccessKey    string
	S3SecretKey    string
	S3BucketName   string
	S3PublicURL    string
	S3UsePathStyle bool
}

// load reads environment variables and returns a Config struct
func Load() (*Config, error) {
	err := godotenv.Load()
	if err != nil {
		log.Println(".env file not found, relying on OS environment variables")
	}
	env := os.Getenv("APP_ENV")
	if env == "" {
		return nil, ErrMissingEnv("APP_ENV")
	}

	port := os.Getenv("APP_PORT")
	if port == "" {
		return nil, ErrMissingEnv("APP_PORT")
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		return nil, ErrMissingEnv("DATABASE_URL")
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return nil, ErrMissingEnv("JWT_SECRET")
	}

	// JWT expiry from env, default 24h
	jwtExpiryStr := os.Getenv("JWT_EXPIRY_HOURS")
	jwtExpiry := 24 * time.Hour
	if jwtExpiryStr != "" {
		if hours, err := time.ParseDuration(jwtExpiryStr + "h"); err == nil {
			jwtExpiry = hours
		} else {
			log.Println("invalid JWT_EXPIRY_HOURS, using default 24h")
		}
	}

	cfg := &Config{
		Env:            env,
		Port:           port,
		DatabaseURL:    dbURL,
		JWTSecret:      jwtSecret,
		JWTExpiry:      jwtExpiry,
		S3Endpoint:     os.Getenv("S3_ENDPOINT"),
		S3Region:       os.Getenv("S3_REGION"),
		S3AccessKey:    os.Getenv("S3_ACCESS_KEY"),
		S3SecretKey:    os.Getenv("S3_SECRET_KEY"),
		S3BucketName:   os.Getenv("S3_BUCKET_NAME"),
		S3PublicURL:    os.Getenv("S3_PUBLIC_URL"),
		S3UsePathStyle: os.Getenv("S3_USE_PATH_STYLE") == "true",
	}

	return cfg, nil
}

// custom error type for missing env variables
type ErrMissingEnv string

func (e ErrMissingEnv) Error() string {
	return "missing required environment variable: " + string(e)
}
