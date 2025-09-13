package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/SrabanMondal/SecureStore/internal/services"
	"github.com/SrabanMondal/SecureStore/internal/utils"
)

type ShareHandler struct {
	ShareSvc *services.ShareService
}

func NewShareHandler(shareSvc *services.ShareService) *ShareHandler {
	return &ShareHandler{ShareSvc: shareSvc}
}

func (h *ShareHandler) CreateShareLink(c echo.Context) error {
	type reqBody struct {
		FileID   string `json:"file_id" validate:"required"`
		Expiry   int    `json:"expiry_hours" validate:"required,min=1"`
		Password string `json:"password"`
	}

	var body reqBody
	if err := c.Bind(&body); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid request"})
	}

	ctx := c.Request().Context()
	share, err := h.ShareSvc.CreateShareLink(ctx, body.FileID, time.Duration(body.Expiry)*time.Hour, body.Password)
	if err != nil {
		utils.Error.Err(err).Msg("create share link failed")
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "could not create share link"})
	}

	return c.JSON(http.StatusOK, echo.Map{
		"token":      share.ShareToken,
		"expires_at": share.ExpiresAt,
	})
}

func (h *ShareHandler) AccessShareLink(c echo.Context) error {
	token := c.Param("token")
	ctx := c.Request().Context()

	file, err := h.ShareSvc.ValidateShareLink(ctx, token, "")
	if err != nil {
		if err.Error() == "password_required" {
			return c.JSON(http.StatusUnauthorized, echo.Map{"error": "password_required"})
		}
		return c.JSON(http.StatusForbidden, echo.Map{"error": err.Error()})
	}

	content, err := h.ShareSvc.GetDownloadContent(ctx, file)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "download failed"})
	}

	switch v := content.(type) {
	case []byte:
		return c.Blob(http.StatusOK, "application/octet-stream", v)
	case string:
		return c.JSON(http.StatusOK, echo.Map{"download_url": v})
	default:
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "unexpected content type"})
	}
}

func (h *ShareHandler) ValidatePassword(c echo.Context) error {
	token := c.Param("token")

	type reqBody struct {
		Password string `json:"password" validate:"required"`
	}
	var body reqBody
	if err := c.Bind(&body); err != nil || body.Password == "" {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "password is required"})
	}

	ctx := c.Request().Context()
	file, err := h.ShareSvc.ValidateShareLink(ctx, token, body.Password)
	if err != nil {
		return c.JSON(http.StatusForbidden, echo.Map{"error": err.Error()})
	}

	content, err := h.ShareSvc.GetDownloadContent(ctx, file)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "download failed"})
	}

	switch v := content.(type) {
	case []byte:
		return c.Blob(http.StatusOK, "application/octet-stream", v)
	case string:
		return c.JSON(http.StatusOK, echo.Map{"download_url": v})
	default:
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "unexpected content type"})
	}
}

func (h* ShareHandler) DeleteLink(c echo.Context) error {
	token := c.Param("id")
	err:= h.ShareSvc.DeleteLink(context.Background(),token)
	if err!=nil{
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, echo.Map{"message":"deleted"})
}