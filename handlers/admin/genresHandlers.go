package admin

import (
	"mime/multipart"
	"net/http"
	"ozinshe_production/models"
	"ozinshe_production/repositories"
	"strconv"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"ozinshe_production/logger"
)

type GenresHandler struct {
	genresRepo *repositories.GenresRepository
}

func NewGenresHandler(repo *repositories.GenresRepository) *GenresHandler {
	return &GenresHandler{genresRepo: repo}
}

type createGenresRequest struct {
	Title  string                `form:"title"`
	Poster *multipart.FileHeader `form:"poster"`
}

// FindAll godoc
// @Summary Get all genres
// @Description Retrieve all genres from the database
// @Tags genres
// @Accept json
// @Produce json
// @Success 200 {array} models.Genre
// @Failure 500 {object} models.ApiError
// @Router /admin/genres [get]
func (h *GenresHandler) FindAll(c *gin.Context) {
	logger := logger.GetLogger()

	genres, err := h.genresRepo.FindAll(c)
	if err != nil {
		logger.Error("Failed to load genres", zap.Error(err))
		c.Status(http.StatusInternalServerError)
		return
	}

	logger.Info("Genres loaded successfully", zap.Int("count", len(genres)))
	c.JSON(http.StatusOK, genres)
}

// FindById godoc
// @Summary Get genre by ID
// @Description Retrieve a genre by its ID
// @Tags genres
// @Accept json
// @Produce json
// @Param id path int true "Genre ID"
// @Success 200 {object} models.Genre
// @Failure 400 {object} models.ApiError
// @Failure 404 {object} models.ApiError
// @Router /admin/genres/{id} [get]
func (h *GenresHandler) FindById(c *gin.Context) {
	logger := logger.GetLogger()

	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		logger.Error("Invalid genre id", zap.String("id", idStr))
		c.JSON(http.StatusBadRequest, models.NewApiError("Invalid genres id"))
		return
	}

	genre, err := h.genresRepo.FindById(c, id)
	if err != nil {
		logger.Error("Failed to find genre", zap.Int("id", id), zap.Error(err))
		c.JSON(http.StatusNotFound, models.NewApiError(err.Error()))
		return
	}

	logger.Info("Genre found", zap.Int("id", id), zap.String("title", genre.Title))
	c.JSON(http.StatusOK, genre)
}

// Create godoc
// @Summary Create a new genre
// @Description Create a new genre and upload a poster
// @Tags genres
// @Accept multipart/form-data
// @Produce json
// @Param title formData string true "Title of the genre"
// @Param poster formData file true "Poster image"
// @Success 200 {object} map[string]int "id of the created genre"
// @Failure 400 {object} models.ApiError
// @Failure 500 {object} models.ApiError
// @Router /admin/genres [post]
func (h *GenresHandler) Create(c *gin.Context) {
	logger := logger.GetLogger()

	var createGenre createGenresRequest
	err := c.Bind(&createGenre)
	if err != nil {
		logger.Error("Couldn't bind json", zap.Error(err))
		c.JSON(http.StatusBadRequest, models.NewApiError("Couldn't bind json"))
		return
	}

	filename, err := savePoster(c, createGenre.Poster)
	if err != nil {
		logger.Error("Failed to save poster", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewApiError(err.Error()))
		return
	}

	genre := models.Genre{
		Title:     createGenre.Title,
		PosterUrl: filename,
	}

	id, err := h.genresRepo.Create(c, genre)
	if err != nil {
		logger.Error("Failed to create genre", zap.Error(err))
		c.JSON(http.StatusNotFound, models.NewApiError(err.Error()))
		return
	}

	logger.Info("Genre created successfully", zap.Int("id", id))
	c.JSON(http.StatusOK, gin.H{
		"id": id,
	})
}

// Update godoc
// @Summary Update a genre by ID
// @Description Update a genre's details, including its title and poster
// @Tags genres
// @Accept json
// @Produce json
// @Param id path int true "Genre ID"
// @Param genre body models.Genre true "Updated genre data"
// @Success 200 {string} string "Genre updated successfully"
// @Failure 400 {object} models.ApiError
// @Failure 404 {object} models.ApiError
// @Router /admin/genres/{id} [put]
func (h *GenresHandler) Update(c *gin.Context) {
	logger := logger.GetLogger()

	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		logger.Error("Invalid genre id", zap.String("id", idStr))
		c.JSON(http.StatusBadRequest, models.NewApiError("Invalid genre id"))
		return
	}

	_, err = h.genresRepo.FindById(c, id)
	if err != nil {
		logger.Error("Failed to find genre", zap.Int("id", id), zap.Error(err))
		c.JSON(http.StatusNotFound, models.NewApiError(err.Error()))
		return
	}

	var updateGenre models.Genre
	err = c.BindJSON(&updateGenre)
	if err != nil {
		logger.Error("Couldn't bind json", zap.Error(err))
		c.JSON(http.StatusBadRequest, models.NewApiError("Couldn't bind json"))
		return
	}

	err = h.genresRepo.Update(c, id, updateGenre)
	if err != nil {
		logger.Error("Failed to update genre", zap.Int("id", id), zap.Error(err))
		c.JSON(http.StatusNotFound, models.NewApiError(err.Error()))
		return
	}

	logger.Info("Genre updated successfully", zap.Int("id", id))
	c.Status(http.StatusOK)
}

// Delete godoc
// @Summary Delete a genre by ID
// @Description Delete a genre from the database by its ID
// @Tags genres
// @Accept json
// @Produce json
// @Param id path int true "Genre ID"
// @Success 200 {string} string "Genre deleted successfully"
// @Failure 400 {object} models.ApiError
// @Failure 404 {object} models.ApiError
// @Router /admin/genres/{id} [delete]
func (h *GenresHandler) Delete(c *gin.Context) {
	logger := logger.GetLogger()

	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		logger.Error("Invalid genre id", zap.String("id", idStr))
		c.JSON(http.StatusBadRequest, models.NewApiError("Invalid genre id"))
		return
	}

	_, err = h.genresRepo.FindById(c, id)
	if err != nil {
		logger.Error("Failed to find genre", zap.Int("id", id), zap.Error(err))
		c.JSON(http.StatusBadRequest, models.NewApiError(err.Error()))
		return
	}

	err = h.genresRepo.Delete(c, id)
	if err != nil {
		logger.Error("Failed to delete genre", zap.Int("id", id), zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewApiError(err.Error()))
		return
	}

	logger.Info("Genre deleted successfully", zap.Int("id", id))
	c.Status(http.StatusOK)
}
