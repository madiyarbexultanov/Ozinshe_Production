package admin

import (
	"net/http"
	"ozinshe_production/models"
	"ozinshe_production/repositories"

	"strconv"

	"github.com/gin-gonic/gin"
)

type CategoriesHandler struct {
	categoriesRepo *repositories.CategoriesRepository
}

func NewCategoriesHandler(repo *repositories.CategoriesRepository) *CategoriesHandler {
	return &CategoriesHandler{categoriesRepo: repo}
}

type createCategoryRequest struct {
	Title  string                `form:"title"`
}

func (h *CategoriesHandler) FindAll(c *gin.Context) {
	categories, err := h.categoriesRepo.FindAll(c)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}
	c.JSON(http.StatusOK, categories)
}

func (h *CategoriesHandler) FindById(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)

	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewApiError("Invalid category id"))
		return
	}

	category, err := h.categoriesRepo.FindById(c, id)
	if err != nil {
		c.JSON(http.StatusNotFound, models.NewApiError(err.Error()))
		return
	}

	c.JSON(http.StatusOK, category)
}

func (h *CategoriesHandler) Create(c *gin.Context) {
	var createCategory createCategoryRequest

	err := c.Bind(&createCategory)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewApiError("Couldn't bind json"))
		return
	}

	categpry := models.Category{
		Title:     createCategory.Title,
	}

	id, err := h.categoriesRepo.Create(c, categpry)
	if err != nil {
		c.JSON(http.StatusNotFound, models.NewApiError(err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id": id,
	})
}

func (h *CategoriesHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)

	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewApiError("Invalid category id"))
		return
	}

	_, err = h.categoriesRepo.FindById(c, id)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewApiError(err.Error()))
		return
	}

	h.categoriesRepo.Delete(c, id)
	c.Status(http.StatusOK)
}
