package admin

import (
	"mime/multipart"
	"net/http"
	"ozinshe_production/models"
	"ozinshe_production/repositories"

	"strconv"

	"github.com/gin-gonic/gin"
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
	genres, err := h.genresRepo.FindAll(c)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}
	c.JSON(http.StatusOK, genres)
}

func (h *GenresHandler) FindById(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)

	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewApiError("Invalid genres id"))
		return
	}

	genre, err := h.genresRepo.FindById(c, id)
	if err != nil {
		c.JSON(http.StatusNotFound, models.NewApiError(err.Error()))
		return
	}

	c.JSON(http.StatusOK, genre)
}

func (h *GenresHandler) Create(c *gin.Context) {
	var createGenre createAgesRequest

	err := c.Bind(&createGenre)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewApiError("Couldn't bind json"))
		return
	}

	filename, err := savePoster(c, createGenre.Poster)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewApiError(err.Error()))
		return
	}

	genre := models.Genre{
		Title:     createGenre.Title,
		PosterUrl: filename,
	}

	id, err := h.genresRepo.Create(c, genre)
	if err != nil {
		c.JSON(http.StatusNotFound, models.NewApiError(err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id": id,
	})
}

func (h *GenresHandler) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)

	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewApiError("Invalid genre id"))
		return
	}

	_, err = h.genresRepo.FindById(c, id)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewApiError(err.Error()))
		return
	}

	var updateGenre models.Genre
	err = c.BindJSON(&updateGenre)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewApiError("Couldn't bind json"))
		return
	}


	err = h.genresRepo.Update(c, id, updateGenre)
	if err != nil {
        c.JSON(http.StatusNotFound, models.NewApiError(err.Error()))
        return
    }

	c.Status(http.StatusOK)
}

func (h *GenresHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)

	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewApiError("Invalid genre id"))
		return
	}

	_, err = h.genresRepo.FindById(c, id)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewApiError(err.Error()))
		return
	}

	h.genresRepo.Delete(c, id)
	c.Status(http.StatusOK)
}
