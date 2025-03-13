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

type RolesHandler struct {
	rolesRepo *repositories.RolesRepository
}

func NewRolesHandler(repo *repositories.RolesRepository) *RolesHandler {
	return &RolesHandler{rolesRepo: repo}
}

type createRoleRequest struct {
	Name              string `json:"name" binding:"required"`
	CanEditProjects   bool   `json:"can_edit_projects"`
	CanEditCategories bool   `json:"can_edit_categories"`
	CanEditUsers      bool   `json:"can_edit_users"`
	CanEditRoles      bool   `json:"can_edit_roles"`
	CanEditGenres     bool   `json:"can_edit_genres"`
	CanEditAges       bool   `json:"can_edit_ages"`
}

func (h *RolesHandler) FindAll(c *gin.Context) {
	logger := logger.GetLogger()

	roles, err := h.rolesRepo.FindAll(c)
	if err != nil {
		logger.Error("Failed to load roles", zap.Error(err))
		c.Status(http.StatusInternalServerError)
		return
	}

	logger.Info("Roles loaded successfully", zap.Int("count", len(roles)))
	c.JSON(http.StatusOK, roles)
}

func (h *RolesHandler) FindById(c *gin.Context) {
	logger := logger.GetLogger()

	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		logger.Error("Invalid role id", zap.String("id", idStr))
		c.JSON(http.StatusBadRequest, models.NewApiError("Invalid role id"))
		return
	}

	role, err := h.rolesRepo.FindById(c, id)
	if err != nil {
		logger.Error("Failed to find role", zap.Int("id", id), zap.Error(err))
		c.JSON(http.StatusNotFound, models.NewApiError(err.Error()))
		return
	}

	logger.Info("Role found", zap.Int("id", id), zap.String("name", role.Name))
	c.JSON(http.StatusOK, role)
}

func (h *RolesHandler) Create(c *gin.Context) {
	logger := logger.GetLogger()

	var createRole createRoleRequest
	err := c.BindJSON(&createRole)
	if err != nil {
		logger.Error("Couldn't bind JSON", zap.Error(err))
		c.JSON(http.StatusBadRequest, models.NewApiError("Couldn't bind JSON"))
		return
	}

	role := models.Role{
		Name:              createRole.Name,
		CanEditProjects:   createRole.CanEditProjects,
		CanEditCategories: createRole.CanEditCategories,
		CanEditUsers:      createRole.CanEditUsers,
		CanEditRoles:      createRole.CanEditRoles,
		CanEditGenres:     createRole.CanEditGenres,
		CanEditAges:       createRole.CanEditAges,
	}

	id, err := h.rolesRepo.Create(c, role)
	if err != nil {
		logger.Error("Failed to create role", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewApiError(err.Error()))
		return
	}

	logger.Info("Role created successfully", zap.Int("id", id), zap.String("name", createRole.Name))
	c.JSON(http.StatusOK, gin.H{"id": id})
}

func (h *RolesHandler) Update(c *gin.Context) {
	logger := logger.GetLogger()

	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		logger.Error("Invalid role id", zap.String("id", idStr))
		c.JSON(http.StatusBadRequest, models.NewApiError("Invalid role id"))
		return
	}

	_, err = h.rolesRepo.FindById(c, id)
	if err != nil {
		logger.Error("Failed to find role", zap.Int("id", id), zap.Error(err))
		c.JSON(http.StatusNotFound, models.NewApiError(err.Error()))
		return
	}

	var updateRole models.Role
	err = c.BindJSON(&updateRole)
	if err != nil {
		logger.Error("Couldn't bind JSON", zap.Error(err))
		c.JSON(http.StatusBadRequest, models.NewApiError("Couldn't bind JSON"))
		return
	}

	err = h.rolesRepo.Update(c, id, updateRole)
	if err != nil {
		logger.Error("Failed to update role", zap.Int("id", id), zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewApiError(err.Error()))
		return
	}

	logger.Info("Role updated successfully", zap.Int("id", id))
	c.Status(http.StatusOK)
}

func (h *RolesHandler) Delete(c *gin.Context) {
	logger := logger.GetLogger()

	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		logger.Error("Invalid role id", zap.String("id", idStr))
		c.JSON(http.StatusBadRequest, models.NewApiError("Invalid role id"))
		return
	}

	_, err = h.rolesRepo.FindById(c, id)
	if err != nil {
		logger.Error("Failed to find role", zap.Int("id", id), zap.Error(err))
		c.JSON(http.StatusNotFound, models.NewApiError(err.Error()))
		return
	}

	err = h.rolesRepo.Delete(c, id)
	if err != nil {
		logger.Error("Failed to delete role", zap.Int("id", id), zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewApiError(err.Error()))
		return
	}

	logger.Info("Role deleted successfully", zap.Int("id", id))
	c.Status(http.StatusOK)
}
