package handlers

import (
	"context"
	//"io"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/SrabanMondal/SecureStore/internal/services"
	"github.com/SrabanMondal/SecureStore/internal/repository"
	//"github.com/SrabanMondal/SecureStore/internal/models"
)

type FileHandler struct {
	FileService *services.FileService
	FileRepo    *repositories.FileRepository
}

func NewFileHandler(fs *services.FileService, repo *repositories.FileRepository) *FileHandler {
	return &FileHandler{
		FileService: fs,
		FileRepo:    repo,
	}
}


func (h *FileHandler) UploadPresigned(c echo.Context) error {
	userID := c.Get("userID").(string)

	req := struct {
		FilePath string `json:"file_path"`
		Size     int64  `json:"size"`
	}{}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid request"})
	}

	url, file, err := h.FileService.GeneratePresignedUpload(context.Background(), userID, req.FilePath, req.Size)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, echo.Map{
		"upload_url": url,
		"file_id":    file.ID,
	})
}

func (h *FileHandler) FinalizeUpload(c echo.Context) error {
	fileID := c.Param("id")

	if err := h.FileService.MarkUploaded(context.Background(), fileID); err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, echo.Map{"status": "uploaded"})
}

func (h *FileHandler) UploadEncrypted(c echo.Context) error {
	userID := c.Get("userID").(string)
	filePath := c.FormValue("file_path")

	fileHeader, err := c.FormFile("file")
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "file missing"})
	}

	src, err := fileHeader.Open()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "failed to open file"})
	}
	defer src.Close()

	if err := h.FileService.UploadEncrypted(context.Background(), userID, filePath, src, fileHeader.Size); err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, echo.Map{"status": "uploaded"})
}

func (h *FileHandler) Download(c echo.Context) error {
	fileID := c.Param("id")

	file, err := h.FileRepo.GetFileByID(context.Background(), fileID)
	if err != nil {
		return c.JSON(http.StatusNotFound, echo.Map{"error": "file not found"})
	}

	if file.IsEncrypted {
		data, err := h.FileService.DownloadDecrypt(context.Background(), file)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
		}
		return c.Blob(http.StatusOK, "application/octet-stream", data)
	}

	url, err := h.FileService.GetDownloadURL(context.Background(), file)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}
	return c.Redirect(http.StatusFound, url)
}

func (h *FileHandler) Delete(c echo.Context) error {
	fileID := c.Param("id")

	if err := h.FileService.DeleteFile(context.Background(), fileID); err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, echo.Map{"status": "deleted"})
}

func (h* FileHandler) ListFiles(c echo.Context) error {
	userID := c.Get("userID").(string)
	files, err := h.FileRepo.ListFilesByUser(context.Background(), userID)
	if(err!=nil){
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, echo.Map{"files":files})
}