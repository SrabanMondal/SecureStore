package config

import (
	"context"
	"os"
	"strconv"
	"time"

	"github.com/SrabanMondal/SecureStore/internal/utils"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type Config struct {
	DB      *pgxpool.Pool
	Minio   *minio.Client
	JWTKey  string
	AppPort string
}

func LoadConfig(ctx context.Context) *Config {

	if err := godotenv.Load(); err != nil {
		utils.Warn.Warn().Msg("No .env file found")
	}

	// ========== POSTGRES ==========
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		utils.Error.Error().Msg("DATABASE_URL not set in env")
		os.Exit(1)
	}

	dbpool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		utils.Error.Error().Err(err).Msg("Unable to connect to database")
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := dbpool.Ping(ctx); err != nil {
		utils.Error.Error().Err(err).Msg("Cannot ping database")
		os.Exit(1)
	}

	// ========== MINIO ==========
	minioEndpoint := os.Getenv("MINIO_ENDPOINT")
	minioAccessKey := os.Getenv("MINIO_ACCESS_KEY")
	minioSecretKey := os.Getenv("MINIO_SECRET_KEY")
	minioSSL, _ := strconv.ParseBool(os.Getenv("MINIO_USE_SSL"))

	minioClient, err := minio.New(minioEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(minioAccessKey, minioSecretKey, ""),
		Secure: minioSSL,
	})
	if err != nil {
		utils.Error.Error().Err(err).Msg("Failed to connect to MinIO")
		os.Exit(1)
	}

	// ========== JWT ==========
	jwtKey := os.Getenv("JWT_SECRET")
	if jwtKey == "" {
		utils.Error.Error().Msg("JWT_SECRET not set in env")
		os.Exit(1)
	}

	// ========== APP PORT ==========
	appPort := os.Getenv("APP_PORT")
	if appPort == "" {
		appPort = "8080"
	}
	appPort = ":"+appPort
	utils.Info.Info().Msg("Config loaded successfully")

	return &Config{
		DB:      dbpool,
		Minio:   minioClient,
		JWTKey:  jwtKey,
		AppPort: appPort,
	}
}
