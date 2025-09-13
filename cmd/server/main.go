package main

import (
	"context"

	"github.com/SrabanMondal/SecureStore/internal/config"
	"github.com/SrabanMondal/SecureStore/internal/utils"
	"github.com/labstack/echo/v4"
)

func main() {
	ctx := context.Background()
	utils.InitLogger()
	cfg := config.LoadConfig(ctx)

	e := echo.New()
	e.Logger.Fatal(e.Start(cfg.AppPort))
	
}