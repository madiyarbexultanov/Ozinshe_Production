package admin

import (
	"mime/multipart"
	"net/http"
	"ozinshe_production/models"
	"ozinshe_production/repositories"

	"strconv"

	"github.com/gin-gonic/gin"
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
	ages, err := h.agesRepo.FindAll(c)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}
	c.JSON(http.StatusOK, ages)
}

func (h *AgesHandler) FindById(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)

	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewApiError("Invalid ages id"))
		return
	}

	ages, err := h.agesRepo.FindById(c, id)
	if err != nil {
		c.JSON(http.StatusNotFound, models.NewApiError(err.Error()))
		return
	}

	c.JSON(http.StatusOK, ages)
}

func (h *AgesHandler) Create(c *gin.Context) {
	var createAges createAgesRequest

	err := c.Bind(&createAges)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewApiError("Couldn't bind json"))
		return
	}

	filename, err := savePoster(c, createAges.Poster)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewApiError(err.Error()))
		return
	}

	ages := models.Ages{
		Title:     createAges.Title,
		PosterUrl: filename,
	}

	id, err := h.agesRepo.Create(c, ages)
	if err != nil {
		c.JSON(http.StatusNotFound, models.NewApiError(err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id": id,
	})
}

func (h *AgesHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)

	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewApiError("Invalid ages id"))
		return
	}

	_, err = h.agesRepo.FindById(c, id)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewApiError(err.Error()))
		return
	}

	h.agesRepo.Delete(c, id)
	c.Status(http.StatusOK)
}
