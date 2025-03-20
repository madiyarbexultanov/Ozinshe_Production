package admin

import (
	"net/http"
	"ozinshe_production/models"
	"ozinshe_production/repositories"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"ozinshe_production/logger"
)

type MovieTypesHandler struct {
	movieTypesRepo *repositories.MovieTypesRepository
}

func NewMovieTypesHandler(repo *repositories.MovieTypesRepository) *MovieTypesHandler {
	return &MovieTypesHandler{movieTypesRepo: repo}
}

type createMovieTypeRequest struct {
	Title string `json:"name"`
}

// FindAll godoc
// @Summary Get all movie types
// @Description Get a list of all movie types
// @Tags MovieTypes
// @Accept json
// @Produce json
// @Success 200 {array} models.MovieType
// @Failure 500 {object} models.ApiError
// @Router /admin/movieTypes [get]
func (h *MovieTypesHandler) FindAll(c *gin.Context) {
	logger := logger.GetLogger()

	movieTypes, err := h.movieTypesRepo.FindAll(c)
	if err != nil {
		logger.Error("Failed to load movie types", zap.Error(err))
		c.Status(http.StatusInternalServerError)
		return
	}

	logger.Info("Movie types loaded successfully", zap.Int("count", len(movieTypes)))
	c.JSON(http.StatusOK, movieTypes)
}

// FindById godoc
// @Summary Get movie type by ID
// @Description Get a movie type by its ID
// @Tags MovieTypes
// @Accept json
// @Produce json
// @Param id path int true "Movie Type ID"
// @Success 200 {object} models.MovieType
// @Failure 400 {object} models.ApiError
// @Failure 404 {object} models.ApiError
// @Router /admin/movieTypes/{id} [get]
func (h *MovieTypesHandler) FindById(c *gin.Context) {
	logger := logger.GetLogger()

	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		logger.Error("Invalid movie type id", zap.String("id", idStr))
		c.JSON(http.StatusBadRequest, models.NewApiError("Invalid movie type id"))
		return
	}

	movieType, err := h.movieTypesRepo.FindById(c, id)
	if err != nil {
		logger.Error("Failed to find movie type", zap.Int("id", id), zap.Error(err))
		c.JSON(http.StatusNotFound, models.NewApiError(err.Error()))
		return
	}

	logger.Info("Movie type found", zap.Int("id", id), zap.String("name", movieType.Title))
	c.JSON(http.StatusOK, movieType)
}

// Create godoc
// @Summary Create a new movie type
// @Description Create a new movie type with the specified name
// @Tags MovieTypes
// @Accept json
// @Produce json
// @Param movieType body createMovieTypeRequest true "Movie Type Information"
// @Success 200 {object} map[string]int "Movie Type ID"
// @Failure 400 {object} models.ApiError
// @Failure 500 {object} models.ApiError
// @Router /admin/movieTypes [post]
func (h *MovieTypesHandler) Create(c *gin.Context) {
	logger := logger.GetLogger()

	var createMovieType createMovieTypeRequest
	err := c.Bind(&createMovieType)
	if err != nil {
		logger.Error("Couldn't bind json", zap.Error(err))
		c.JSON(http.StatusBadRequest, models.NewApiError("Couldn't bind json"))
		return
	}

	movieType := models.MovieType{
		Title: createMovieType.Title,
	}

	id, err := h.movieTypesRepo.Create(c, movieType)
	if err != nil {
		logger.Error("Failed to create movie type", zap.Error(err))
		c.JSON(http.StatusNotFound, models.NewApiError(err.Error()))
		return
	}

	logger.Info("Movie type created successfully", zap.Int("id", id))
	c.JSON(http.StatusOK, gin.H{
		"id": id,
	})
}

// Update godoc
// @Summary Update an existing movie type
// @Description Update an existing movie type by its ID
// @Tags MovieTypes
// @Accept json
// @Produce json
// @Param id path int true "Movie Type ID"
// @Param movieType body models.MovieType true "Updated Movie Type Information"
// @Success 200 {object} map[string]string "Movie type updated successfully"
// @Failure 400 {object} models.ApiError
// @Failure 404 {object} models.ApiError
// @Failure 500 {object} models.ApiError
// @Router /admin/movieTypes/{id} [put]
func (h *MovieTypesHandler) Update(c *gin.Context) {
	logger := logger.GetLogger()

	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		logger.Error("Invalid movie type id", zap.String("id", idStr))
		c.JSON(http.StatusBadRequest, models.NewApiError("Invalid movie type id"))
		return
	}

	_, err = h.movieTypesRepo.FindById(c, id)
	if err != nil {
		logger.Error("Failed to find movie type", zap.Int("id", id), zap.Error(err))
		c.JSON(http.StatusNotFound, models.NewApiError(err.Error()))
		return
	}

	var updateMovieType models.MovieType
	err = c.BindJSON(&updateMovieType)
	if err != nil {
		logger.Error("Couldn't bind json", zap.Error(err))
		c.JSON(http.StatusBadRequest, models.NewApiError("Couldn't bind json"))
		return
	}

	err = h.movieTypesRepo.Update(c, id, updateMovieType)
	if err != nil {
		logger.Error("Failed to update movie type", zap.Int("id", id), zap.Error(err))
		c.JSON(http.StatusNotFound, models.NewApiError(err.Error()))
		return
	}

	logger.Info("Movie type updated successfully", zap.Int("id", id))
	c.Status(http.StatusOK)
}

// Delete godoc
// @Summary Delete a movie type
// @Description Delete a movie type by its ID
// @Tags MovieTypes
// @Accept json
// @Produce json
// @Param id path int true "Movie Type ID"
// @Success 200 {object} map[string]string "Movie type deleted successfully"
// @Failure 400 {object} models.ApiError
// @Failure 500 {object} models.ApiError
// @Router /admin/movieTypes/{id} [delete]
func (h *MovieTypesHandler) Delete(c *gin.Context) {
	logger := logger.GetLogger()

	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		logger.Error("Invalid movie type id", zap.String("id", idStr))
		c.JSON(http.StatusBadRequest, models.NewApiError("Invalid movie type id"))
		return
	}

	_, err = h.movieTypesRepo.FindById(c, id)
	if err != nil {
		logger.Error("Failed to find movie type", zap.Int("id", id), zap.Error(err))
		c.JSON(http.StatusBadRequest, models.NewApiError(err.Error()))
		return
	}

	err = h.movieTypesRepo.Delete(c, id)
	if err != nil {
		logger.Error("Failed to delete movie type", zap.Int("id", id), zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewApiError(err.Error()))
		return
	}

	logger.Info("Movie type deleted successfully", zap.Int("id", id))
	c.Status(http.StatusOK)
}
