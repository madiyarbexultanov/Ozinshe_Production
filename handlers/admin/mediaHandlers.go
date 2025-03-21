package admin

import (
	"mime/multipart"
	"net/http"
	"ozinshe_production/logger"
	"ozinshe_production/models"
	"ozinshe_production/repositories"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type MediaHandler struct {
	mediaRepo *repositories.MediaRepository
}

func NewMediaHandler(repo *repositories.MediaRepository) *MediaHandler {
	return &MediaHandler{mediaRepo: repo}
}


type addMediaRequest struct {
	Cover  		*multipart.FileHeader       `form:"cover"`
	Screenshots []*multipart.FileHeader 	`form:"screenshots"`
}

func (h *MediaHandler) GetMovieMedia(c *gin.Context) {
    logger := logger.GetLogger()

    movieIdStr := c.Param("id")
    movieId, err := strconv.Atoi(movieIdStr)
    if err != nil {
        logger.Error("Invalid movie ID", zap.Error(err))
        c.JSON(http.StatusBadRequest, models.NewApiError("Invalid movie ID"))
        return
    }

    media, err := h.mediaRepo.GetMovieMedia(c, movieId)
    if err != nil {
        c.JSON(http.StatusInternalServerError, models.NewApiError("Couldn't get movie media"))
        return
    }

    c.JSON(http.StatusOK, media)
}

func (h *MediaHandler) UploadMovieMedia(c *gin.Context) {
    var request addMediaRequest
    logger := logger.GetLogger()

    // Привязка данных формы к структуре request
    if err := c.ShouldBind(&request); err != nil {
        logger.Error("Failed to bind multipart form", zap.Error(err))
        c.JSON(http.StatusBadRequest, models.NewApiError("Failed to bind form data"))
        return
    }

    movieIdStr := c.Param("id")
    movieId, err := strconv.Atoi(movieIdStr)
    if err != nil {
        logger.Error("Invalid movie ID", zap.Error(err))
        c.JSON(http.StatusBadRequest, models.NewApiError("Invalid movie ID"))
        return
    }

    var coverFilename *string
    var screenshotFilenames []string

    // Обработка обложки
    if request.Cover != nil {
        filename, err := savePoster(c, request.Cover)
        if err != nil {
            c.JSON(http.StatusInternalServerError, models.NewApiError("Couldn't save cover"))
            return
        }
        coverFilename = &filename
    }

    // Обработка скриншотов
    for _, file := range request.Screenshots {
        filename, err := savePoster(c, file)
        if err != nil {
            c.JSON(http.StatusInternalServerError, models.NewApiError("Couldn't save screenshot"))
            return
        }
        screenshotFilenames = append(screenshotFilenames, filename)
    }

    // Обновление в БД
    err = h.mediaRepo.UpdateMovieMedia(c, movieId, coverFilename, screenshotFilenames)
    if err != nil {
        c.JSON(http.StatusInternalServerError, models.NewApiError("Couldn't update movie media"))
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "Media uploaded successfully"})
}


func (h *MediaHandler) UploadSingleMovieMedia(c *gin.Context) {
    logger := logger.GetLogger()

    movieIdStr := c.Param("id")
    movieId, err := strconv.Atoi(movieIdStr)
    if err != nil {
        logger.Error("Invalid movie ID", zap.Error(err))
        c.JSON(http.StatusBadRequest, models.NewApiError("Invalid movie ID"))
        return
    }

    mediaType := c.DefaultPostForm("type", "")
    if mediaType != "cover" && mediaType != "screenshot" {
        logger.Error("Invalid media type", zap.String("mediaType", mediaType))
        c.JSON(http.StatusBadRequest, models.NewApiError("Invalid media type. Expected 'cover' or 'screenshot'"))
        return
    }

    file, err := c.FormFile("file")
    if err != nil {
        c.JSON(http.StatusBadRequest, models.NewApiError("No file uploaded"))
        return
    }

    filename, err := savePoster(c, file)
    if err != nil {
        c.JSON(http.StatusInternalServerError, models.NewApiError("Couldn't save media file"))
        return
    }

    err = h.mediaRepo.UpdateSingleMovieMedia(c, movieId, mediaType, filename)
    if err != nil {
        c.JSON(http.StatusInternalServerError, models.NewApiError("Couldn't update movie media"))
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "Media uploaded successfully"})
}


func (h *MediaHandler) DeleteMovieMedia(c *gin.Context) {
    logger := logger.GetLogger()

    movieIdStr := c.Param("id")
    movieId, err := strconv.Atoi(movieIdStr)
    if err != nil {
        logger.Error("Invalid movie ID", zap.Error(err))
        c.JSON(http.StatusBadRequest, models.NewApiError("Invalid movie ID"))
        return
    }

    var req struct {
        ImageURL string `json:"image_url"`
    }
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, models.NewApiError("Invalid request"))
        return
    }

    err = h.mediaRepo.DeleteMovieMedia(c, movieId, req.ImageURL)
    if err != nil {
        c.JSON(http.StatusInternalServerError, models.NewApiError("Couldn't delete movie media"))
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "Media deleted successfully"})
}

