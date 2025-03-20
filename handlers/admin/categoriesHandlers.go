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

// FindAll godoc
// @Summary      Get all categories
// @Description  Retrieves a list of all categories
// @Tags         categories
// @Produce      json
// @Success      200 {array} models.Category "A list of categories"
// @Failure      500 {object} models.ApiError "Internal server error"
// @Router       /categories [get]
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

// FindById godoc
// @Summary      Get category by ID
// @Description  Retrieves a category by its ID
// @Tags         categories
// @Produce      json
// @Param        id path int true "Category ID"
// @Success      200 {object} models.Category "Category details"
// @Failure      400 {object} models.ApiError "Invalid category ID"
// @Failure      404 {object} models.ApiError "Category not found"
// @Router       /admin/categories/{id} [get]
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

// Create godoc
// @Summary      Create new category
// @Description  Creates a new category entry
// @Tags         categories
// @Accept       json
// @Produce      json
// @Param        request body createCategoryRequest true "Create category request"
// @Success      200 {object} gin.H "ID of the created category"
// @Failure      400 {object} models.ApiError "Invalid input"
// @Failure      500 {object} models.ApiError "Failed to create category"
// @Router       /admin/categories [post]
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


// Update godoc
// @Summary      Update an existing category
// @Description  Updates an existing category by ID
// @Tags         categories
// @Accept       json
// @Produce      json
// @Param        id path int true "Category ID"
// @Param        request body models.Category true "Update category request"
// @Success      200 {object} string "Success message"
// @Failure      400 {object} models.ApiError "Invalid category ID or input"
// @Failure      404 {object} models.ApiError "Category not found"
// @Router       /admin/categories/{id} [put]
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

// Delete godoc
// @Summary      Delete a category by ID
// @Description  Deletes a category by its ID
// @Tags         categories
// @Param        id path int true "Category ID"
// @Success      200 {object} string "Success message"
// @Failure      400 {object} models.ApiError "Invalid category ID"
// @Failure      500 {object} models.ApiError "Failed to delete category"
// @Router       /admin/categories/{id} [delete]
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
