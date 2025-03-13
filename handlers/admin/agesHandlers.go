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

type AgesHandler struct {
	agesRepo *repositories.AgesRepository
}

func NewAgesHandler(repo *repositories.AgesRepository) *AgesHandler {
	return &AgesHandler{agesRepo: repo}
}

type createAgesRequest struct {
	Title  string                `form:"title"`
	Poster *multipart.FileHeader `form:"poster"`
}

func (h *AgesHandler) FindAll(c *gin.Context) {
	logger := logger.GetLogger()

	ages, err := h.agesRepo.FindAll(c)
	if err != nil {
		logger.Error("Failed to load ages", zap.Error(err))
		c.Status(http.StatusInternalServerError)
		return
	}

	logger.Info("Ages loaded successfully", zap.Int("count", len(ages)))
	c.JSON(http.StatusOK, ages)
}

func (h *AgesHandler) FindById(c *gin.Context) {
	logger := logger.GetLogger()

	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		logger.Error("Invalid ages id", zap.String("id", idStr), zap.Error(err))
		c.JSON(http.StatusBadRequest, models.NewApiError("Invalid ages id"))
		return
	}

	ages, err := h.agesRepo.FindById(c, id)
	if err != nil {
		logger.Error("Ages not found", zap.Int("id", id), zap.Error(err))
		c.JSON(http.StatusNotFound, models.NewApiError(err.Error()))
		return
	}

	logger.Info("Ages found", zap.Int("id", id))
	c.JSON(http.StatusOK, ages)
}

func (h *AgesHandler) Create(c *gin.Context) {
	logger := logger.GetLogger()

	var createAges createAgesRequest
	err := c.Bind(&createAges)
	if err != nil {
		logger.Error("Failed to bind request", zap.Error(err))
		c.JSON(http.StatusBadRequest, models.NewApiError("Couldn't bind json"))
		return
	}

	logger.Info("Received create ages request", zap.String("title", createAges.Title))

	filename, err := savePoster(c, createAges.Poster)
	if err != nil {
		logger.Error("Failed to save poster", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewApiError(err.Error()))
		return
	}

	ages := models.Ages{
		Title:     createAges.Title,
		PosterUrl: filename,
	}

	id, err := h.agesRepo.Create(c, ages)
	if err != nil {
		logger.Error("Failed to create ages", zap.String("title", createAges.Title), zap.Error(err))
		c.JSON(http.StatusNotFound, models.NewApiError(err.Error()))
		return
	}

	logger.Info("Ages created successfully", zap.Int("id", id))
	c.JSON(http.StatusOK, gin.H{
		"id": id,
	})
}

func (h *AgesHandler) Update(c *gin.Context) {
	logger := logger.GetLogger()

	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		logger.Error("Invalid ages id", zap.String("id", idStr), zap.Error(err))
		c.JSON(http.StatusBadRequest, models.NewApiError("Invalid ages id"))
		return
	}

	_, err = h.agesRepo.FindById(c, id)
	if err != nil {
		logger.Error("Ages not found", zap.Int("id", id), zap.Error(err))
		c.JSON(http.StatusBadRequest, models.NewApiError(err.Error()))
		return
	}

	var updateAges models.Ages
	err = c.BindJSON(&updateAges)
	if err != nil {
		logger.Error("Failed to bind update request", zap.Error(err))
		c.JSON(http.StatusBadRequest, models.NewApiError("Couldn't bind json"))
		return
	}

	err = h.agesRepo.Update(c, id, updateAges)
	if err != nil {
		logger.Error("Failed to update ages", zap.Int("id", id), zap.Error(err))
		c.JSON(http.StatusNotFound, models.NewApiError(err.Error()))
		return
	}

	logger.Info("Ages updated successfully", zap.Int("id", id))
	c.Status(http.StatusOK)
}

func (h *AgesHandler) Delete(c *gin.Context) {
	logger := logger.GetLogger()

	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		logger.Error("Invalid ages id", zap.String("id", idStr), zap.Error(err))
		c.JSON(http.StatusBadRequest, models.NewApiError("Invalid ages id"))
		return
	}

	_, err = h.agesRepo.FindById(c, id)
	if err != nil {
		logger.Error("Ages not found", zap.Int("id", id), zap.Error(err))
		c.JSON(http.StatusBadRequest, models.NewApiError(err.Error()))
		return
	}

	err = h.agesRepo.Delete(c, id)
	if err != nil {
		logger.Error("Failed to delete ages", zap.Int("id", id), zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewApiError("Failed to delete ages"))
		return
	}

	logger.Info("Ages deleted successfully", zap.Int("id", id))
	c.Status(http.StatusOK)
}