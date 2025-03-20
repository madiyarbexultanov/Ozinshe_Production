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

type UsersHandler struct {
	userRepo *repositories.UsersRepository
}

func NewUsersHandler(repo *repositories.UsersRepository) *UsersHandler {
	return &UsersHandler{userRepo: repo}
}

type userResponse struct {
	Id    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type AssignRoleRequest struct {
	RoleID int `json:"role_id" binding:"required"`
}

// FindAll retrieves all users.
// @Summary Get all users
// @Description Retrieve a list of all users
// @Tags Users
// @Accept json
// @Produce json
// @Success 200 {array} userResponse
// @Failure 500 {object} models.ApiError
// @Router /admin/users [get]
func (h *UsersHandler) FindAll(c *gin.Context) {
	logger := logger.GetLogger()

	filter := models.Userfilters {
		Sort: 		c.Query("sort"),
	}

	users, err := h.userRepo.FindAll(c, filter)
	if err != nil {
		logger.Error("Failed to load users", zap.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, models.NewApiError("couldn't load users"))
		return
	}

	dtos := make([]userResponse, 0, len(users))
	for _, u := range users {
		dtos = append(dtos, userResponse{Id: u.Id, Name: u.Name, Email: u.Email})
	}

	logger.Info("Users loaded successfully", zap.Int("count", len(users)))
	c.JSON(http.StatusOK, dtos)
}

// FindById retrieves a user by ID.
// @Summary Get a user by ID
// @Description Retrieve a user by their ID
// @Tags Users
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Success 200 {object} models.User
// @Failure 400 {object} models.ApiError
// @Failure 404 {object} models.ApiError
// @Failure 500 {object} models.ApiError
// @Router /admin/users/{id} [get]
func (h *UsersHandler) FindById(c *gin.Context) {
	logger := logger.GetLogger()

	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)

	if err != nil {
		logger.Error("Invalid user id", zap.String("id", idStr))
		c.JSON(http.StatusBadRequest, models.NewApiError("Invalid user id"))
		return
	}

	user, err := h.userRepo.FindById(c, id)
	if err != nil {
		logger.Error("Failed to find user", zap.Int("id", id), zap.Error(err))
		c.JSON(http.StatusNotFound, models.NewApiError(err.Error()))
		return
	}

	logger.Info("User found", zap.Int("id", id), zap.String("name", user.Name))
	c.JSON(http.StatusOK, user)
}

// Delete deletes a user by ID.
// @Summary Delete a user by ID
// @Description Delete a user by their ID
// @Tags Users
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Success 200 {string} string "User deleted successfully"
// @Failure 400 {object} models.ApiError
// @Failure 500 {object} models.ApiError
// @Router /admin/users/{id} [delete]
func (h *UsersHandler) Delete(c *gin.Context) {
	logger := logger.GetLogger()

	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)

	if err != nil {
		logger.Error("Invalid user id", zap.String("id", idStr))
		c.JSON(http.StatusBadRequest, models.NewApiError("Invalid user id"))
		return
	}

	_, err = h.userRepo.FindById(c, id)
	if err != nil {
		logger.Error("Failed to find user", zap.Int("id", id), zap.Error(err))
		c.JSON(http.StatusBadRequest, models.NewApiError(err.Error()))
		return
	}

	err = h.userRepo.Delete(c, id)
	if err != nil {
		logger.Error("Failed to delete user", zap.Int("id", id), zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewApiError(err.Error()))
		return
	}

	logger.Info("User deleted successfully", zap.Int("id", id))
	c.Status(http.StatusOK)
}

// AssignRole assigns a role to a user.
// @Summary Assign a role to a user
// @Description Assign a role to a user by their ID
// @Tags Users
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Param role body AssignRoleRequest true "Role ID"
// @Success 200 {object} map[string]interface{} "Role assigned successfully"
// @Failure 400 {object} models.ApiError
// @Failure 500 {object} models.ApiError
// @Router /admin/users/{id}/getRole [put]
func (h *UsersHandler) AssignRole(c *gin.Context) {
	logger := logger.GetLogger()

	idStr := c.Param("id")
	userID, err := strconv.Atoi(idStr)
	if err != nil {
		logger.Error("Invalid user id", zap.String("id", idStr))
		c.JSON(http.StatusBadRequest, models.NewApiError("Invalid user id"))
		return
	}

	var req AssignRoleRequest
	if err := c.BindJSON(&req); err != nil {
		logger.Error("Invalid request for assigning role", zap.Error(err))
		c.JSON(http.StatusBadRequest, models.NewApiError("Invalid request"))
		return
	}

	err = h.userRepo.AssignRole(c, userID, req.RoleID)
	if err != nil {
		logger.Error("Failed to assign role", zap.Int("user_id", userID), zap.Int("role_id", req.RoleID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewApiError("Failed to assign role"))
		return
	}

	logger.Info("Role assigned successfully", zap.Int("user_id", userID), zap.Int("role_id", req.RoleID))
	c.JSON(http.StatusOK, gin.H{"message": "Role assigned successfully"})
}
