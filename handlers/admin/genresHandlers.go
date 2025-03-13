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
