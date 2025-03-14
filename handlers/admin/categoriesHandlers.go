package admin

import (
	"net/http"
	"ozinshe_production/logger"
	"ozinshe_production/models"
	"ozinshe_production/repositories"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type CategoriesHandler struct {
	categoriesRepo *repositories.CategoriesRepository
}

func NewCategoriesHandler(repo *repositories.CategoriesRepository) *CategoriesHandler {
	return &CategoriesHandler{categoriesRepo: repo}
}

type createCategoryRequest struct {
	Title  string `form:"title"`
}

func (h *CategoriesHandler) FindAll(c *gin.Context) {
	logger := logger.GetLogger()

	categories, err := h.categoriesRepo.FindAll(c)
	if err != nil {
		logger.Error("Failed to load categories", zap.Error(err))
		c.Status(http.StatusInternalServerError)
		return
	}

	logger.Info("Categories loaded successfully", zap.Int("count", len(categories)))
	c.JSON(http.StatusOK, categories)
}

func (h *CategoriesHandler) FindById(c *gin.Context) {
	logger := logger.GetLogger()

	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)

	if err != nil {
		logger.Error("Invalid category id", zap.String("id", idStr), zap.Error(err))
		c.JSON(http.StatusBadRequest, models.NewApiError("Invalid category id"))
		return
	}

	category, err := h.categoriesRepo.FindById(c, id)
	if err != nil {
		logger.Error("Category not found", zap.Int("id", id), zap.Error(err))
		c.JSON(http.StatusNotFound, models.NewApiError(err.Error()))
		return
	}

	logger.Info("Category found", zap.Int("id", id))
	c.JSON(http.StatusOK, category)
}

func (h *CategoriesHandler) Create(c *gin.Context) {
	logger := logger.GetLogger()

	var createCategory createCategoryRequest
	err := c.Bind(&createCategory)
	if err != nil {
		logger.Error("Failed to bind request", zap.Error(err))
		c.JSON(http.StatusBadRequest, models.NewApiError("Couldn't bind json"))
		return
	}

	logger.Info("Received create category request", zap.String("title", createCategory.Title))

	category := models.Category{
		Title: createCategory.Title,
	}

	id, err := h.categoriesRepo.Create(c, category)
	if err != nil {
		logger.Error("Failed to create category", zap.String("title", createCategory.Title), zap.Error(err))
		c.JSON(http.StatusNotFound, models.NewApiError(err.Error()))
		return
	}

	logger.Info("Category created successfully", zap.Int("id", id))
	c.JSON(http.StatusOK, gin.H{
		"id": id,
	})
}

func (h *CategoriesHandler) Update(c *gin.Context) {
	logger := logger.GetLogger()

	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)

	if err != nil {
		logger.Error("Invalid category id", zap.String("id", idStr), zap.Error(err))
		c.JSON(http.StatusBadRequest, models.NewApiError("Invalid genre id"))
		return
	}

	_, err = h.categoriesRepo.FindById(c, id)
	if err != nil {
		logger.Error("Category not found", zap.Int("id", id), zap.Error(err))
		c.JSON(http.StatusBadRequest, models.NewApiError(err.Error()))
		return
	}

	var updateCategory models.Category
	err = c.BindJSON(&updateCategory)
	if err != nil {
		logger.Error("Failed to bind update request", zap.Error(err))
		c.JSON(http.StatusBadRequest, models.NewApiError("Couldn't bind json"))
		return
	}

	err = h.categoriesRepo.Update(c, id, updateCategory)
	if err != nil {
		logger.Error("Failed to update category", zap.Int("id", id), zap.Error(err))
		c.JSON(http.StatusNotFound, models.NewApiError(err.Error()))
		return
	}

	logger.Info("Category updated successfully", zap.Int("id", id))
	c.Status(http.StatusOK)
}

func (h *CategoriesHandler) Delete(c *gin.Context) {
	logger := logger.GetLogger()

	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)

	if err != nil {
		logger.Error("Invalid category id", zap.String("id", idStr), zap.Error(err))
		c.JSON(http.StatusBadRequest, models.NewApiError("Invalid category id"))
		return
	}

	_, err = h.categoriesRepo.FindById(c, id)
	if err != nil {
		logger.Error("Category not found", zap.Int("id", id), zap.Error(err))
		c.JSON(http.StatusBadRequest, models.NewApiError(err.Error()))
		return
	}

	err = h.categoriesRepo.Delete(c, id)
	if err != nil {
		logger.Error("Failed to delete category", zap.Int("id", id), zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewApiError("Failed to delete category"))
		return
	}

	logger.Info("Category deleted successfully", zap.Int("id", id))
	c.Status(http.StatusOK)
}
