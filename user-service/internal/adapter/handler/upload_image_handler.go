package handler

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"time"
	"user-service/config"
	"user-service/internal/adapter"
	"user-service/internal/adapter/handler/response"
	"user-service/internal/adapter/storage"
	"user-service/internal/core/service"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
)

type IUploadImageHandler interface {
	UploadImage(c echo.Context) error
}

type uploadImageHandler struct {
	storage storage.ISupabase
}

// UploadImage implements IUploadImageHandler.
func (u *uploadImageHandler) UploadImage(c echo.Context) error {
	var resp = response.DefaultResponse{}

	file, err := c.FormFile("photo")
	if err != nil {
		log.Errorf("[UploadImageHandler-1] UploadImage: %v", err)
		resp.Message = err.Error()
		resp.Data = nil
		return c.JSON(http.StatusBadRequest, resp)
	}

	src, err := file.Open()
	if err != nil {
		log.Errorf("[UploadImageHandler-2] UploadImage: %v", err)
		resp.Message = err.Error()
		resp.Data = nil
		return c.JSON(http.StatusBadRequest, resp)
	}

	defer src.Close()
	fileBuffer := new(bytes.Buffer)
	_, err = io.Copy(fileBuffer, src)
	if err != nil {
		log.Errorf("[UploadImage-3] UploadImage: %v", err)
		resp.Message = err.Error()
		resp.Data = nil
		return c.JSON(http.StatusInternalServerError, resp)
	}

	newFileName := fmt.Sprintf("%s_%d%s", uuid.New().String(), time.Now().Unix(), getExtension(file.Filename))
	uploadPath := fmt.Sprintf("public/uploads/%s", newFileName)
	url, err := u.storage.UploadFile(uploadPath, fileBuffer)
	
	if err != nil {
		log.Errorf("[UploadImage-4] UploadImage: %v", err)
		resp.Message = err.Error()
		resp.Data = nil
		return c.JSON(http.StatusInternalServerError, resp)
	}

	resp.Message = "Success"
	resp.Data = map[string]string{"image_url": url}
	
	return c.JSON(http.StatusOK, resp)
}

func getExtension(fileName string) string {
	ext := "." + fileName[len(fileName)-3:] // Ambil 3 karakter terakhir untuk ekstensi
	if len(fileName) > 4 && fileName[len(fileName)-4] == '.' {
		ext = "." + fileName[len(fileName)-4:]
	}
	return ext
}

func NewUploadImageHandler(e *echo.Echo, cfg *config.Config,storage storage.ISupabase, jwtService service.IJWTService) IUploadImageHandler {
	res := &uploadImageHandler{
		storage: storage,
	}

	mid := adapter.NewMiddlewareAdapter(cfg, jwtService)
	e.POST("/auth/profile/image-upload", res.UploadImage, mid.CheckToken())

	return res
}
