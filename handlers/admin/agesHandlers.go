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

// FindAll godoc
// @Summary      Get all ages
// @Description  Retrieves a list of all ages
// @Tags         ages
// @Produce      json
// @Success      200 {array} models.Ages "A list of ages"
// @Failure      500 {object} models.ApiError "Internal server error"
// @Router       /admin/ages [get]
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

// FindById godoc
// @Summary      Get age by ID
// @Description  Retrieves an age by its ID
// @Tags         ages
// @Produce      json
// @Param        id path int true "Age ID"
// @Success      200 {object} models.Ages "Age details"
// @Failure      400 {object} models.ApiError "Invalid age ID"
// @Failure      404 {object} models.ApiError "Age not found"
// @Router       /admin/ages/{id} [get]
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

// Create godoc
// @Summary      Create new age
// @Description  Creates a new age entry
// @Tags         ages
// @Accept       json
// @Produce      json
// @Param        request body createAgesRequest true "Create age request"
// @Success      200 {object} gin.H "ID of the created age"
// @Failure      400 {object} models.ApiError "Invalid input"
// @Failure      500 {object} models.ApiError "Failed to save poster or create age"
// @Router       /admin/ages [post]
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

// Update godoc
// @Summary      Update an existing age
// @Description  Updates an existing age by ID
// @Tags         ages
// @Accept       json
// @Produce      json
// @Param        id path int true "Age ID"
// @Param        request body models.Ages true "Update age request"
// @Success      200 {object} string "Success message"
// @Failure      400 {object} models.ApiError "Invalid age ID or input"
// @Failure      404 {object} models.ApiError "Age not found"
// @Router       /admin/ages/{id} [put]
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

// Delete godoc
// @Summary      Delete an age by ID
// @Description  Deletes an age by its ID
// @Tags         ages
// @Param        id path int true "Age ID"
// @Success      200 {object} string "Success message"
// @Failure      400 {object} models.ApiError "Invalid age ID"
// @Failure      500 {object} models.ApiError "Failed to delete age"
// @Router       /admin/ages/{id} [delete]
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