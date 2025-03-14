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


func (h *UsersHandler) FindAll(c *gin.Context) {
	logger := logger.GetLogger()

	users, err := h.userRepo.FindAll(c)
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
