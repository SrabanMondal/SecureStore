package main

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	//"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	//"github.com/minio/minio-go/v7"
	//"github.com/minio/minio-go/v7/pkg/credentials"

	"github.com/SrabanMondal/SecureStore/internal/config"
	"github.com/SrabanMondal/SecureStore/internal/handler"
	"github.com/SrabanMondal/SecureStore/internal/repository"
	"github.com/SrabanMondal/SecureStore/internal/services"
	"github.com/SrabanMondal/SecureStore/internal/utils"
)

func JWTMiddleware(secret string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return c.JSON(http.StatusUnauthorized, echo.Map{"error": "missing token"})
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				return c.JSON(http.StatusUnauthorized, echo.Map{"error": "invalid token format"})
			}

			tokenStr := parts[1]
			claims := jwt.MapClaims{}
			token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
				return []byte(secret), nil
			})
			if err != nil || !token.Valid {
				return c.JSON(http.StatusUnauthorized, echo.Map{"error": "invalid or expired token"})
			}

			userID, ok := claims["user_id"].(string)
			if !ok {
				return c.JSON(http.StatusUnauthorized, echo.Map{"error": "invalid token payload"})
			}

			c.Set("userID", userID)
			return next(c)
		}
	}
}

func main() {
	ctx := context.Background()
	utils.InitLogger()
	cfg := config.LoadConfig(ctx)

	defer cfg.DB.Close()

	userRepo := repositories.NewUserRepository(cfg.DB)
	fileRepo := repositories.NewFileRepository(cfg.DB)
	shareRepo := repositories.NewShareRepository(cfg.DB)

	authSvc := services.NewAuthService(userRepo, cfg.JWTKey, 24 * time.Hour)
	fileSvc := services.NewFileService(fileRepo, cfg.Minio, "uploads", cfg.FileKey)
	shareSvc := services.NewShareService(shareRepo, fileRepo, fileSvc)

	authHandler := handlers.NewAuthHandler(authSvc)
	fileHandler := handlers.NewFileHandler(fileSvc, fileRepo)
	shareHandler := handlers.NewShareHandler(shareSvc)

	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.POST("/api/register", authHandler.Register)
	e.POST("/api/login", authHandler.Login)

	api := e.Group("/api")
	api.Use(JWTMiddleware(cfg.JWTKey))

	api.POST("/files/presigned", fileHandler.UploadPresigned)
	api.POST("/files/encrypted", fileHandler.UploadEncrypted)
	api.POST("/files/:id/finalize", fileHandler.FinalizeUpload)
	api.GET("/files/:id/download", fileHandler.Download)
	api.DELETE("/files/:id", fileHandler.Delete)
	api.GET("/files", fileHandler.ListFiles)

	api.POST("/shares", shareHandler.CreateShareLink)
	api.DELETE("/shares/:id",shareHandler.DeleteLink)
	e.GET("/api/shares/:token", shareHandler.AccessShareLink)        
	e.POST("/api/shares/:token/validate", shareHandler.ValidatePassword)


	go func() {
		ticker := time.NewTicker(10 * time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				if err := fileSvc.CleanupDeletedFiles(ctx); err != nil {
					utils.Error.Err(err).Msg("cleanup deleted files failed")
				} else {
					utils.Info.Info().Msg("deleted files cleanup done")
				}
			}
		}
	}()

	go func() {
		ticker := time.NewTicker(15 * time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				if err := fileSvc.ReconcilePendingFiles(ctx); err != nil {
					utils.Error.Err(err).Msg("reconcile pending files failed")
				} else {
					utils.Info.Info().Msg("pending files reconciled")
				}
			}
		}
	}()

	go func() {
		ticker := time.NewTicker(30 * time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				if err := shareRepo.DeleteExpiredShareLinks(ctx); err != nil {
					utils.Error.Err(err).Msg("expired share cleanup failed")
				} else {
					utils.Info.Info().Msg("expired share links cleaned")
				}
			}
		}
	}()

	utils.Info.Info().Msgf("Server running on %s", cfg.AppPort)
	e.Logger.Fatal(e.Start(cfg.AppPort))
}
