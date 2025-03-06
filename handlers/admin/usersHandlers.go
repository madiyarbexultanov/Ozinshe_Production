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
	c.JSON(http.StatusOK, dtos)
}

func (h *UsersHandler) FindById(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)

	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewApiError("Invalid user id"))
		return
	}

	user, err := h.userRepo.FindById(c, id)
	if err != nil {
		c.JSON(http.StatusNotFound, models.NewApiError(err.Error()))
		return
	}

	c.JSON(http.StatusOK, user)
}

func (h *UsersHandler) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)

	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewApiError("Invalid user id"))
		return
	}

	_, err = h.userRepo.FindById(c, id)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewApiError(err.Error()))
		return
	}

	var updateUser models.User
	err = c.BindJSON(&updateUser)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewApiError("Couldn't bind json"))
		return
	}


	err = h.userRepo.Update(c, id, updateUser)
	if err != nil {
        c.JSON(http.StatusNotFound, models.NewApiError(err.Error()))
        return
    }

	c.Status(http.StatusOK)
}

func (h *UsersHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)

	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewApiError("Invalid user id"))
		return
	}

	_, err = h.userRepo.FindById(c, id)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewApiError(err.Error()))
		return
	}

	h.userRepo.Delete(c, id)
	c.Status(http.StatusOK)
}

func (h *UsersHandler) AssignRole(c *gin.Context) {
	idStr := c.Param("id")
	userID, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewApiError("Invalid user id"))
		return
	}

	var req AssignRoleRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewApiError("Invalid request"))
		return
	}

	err = h.userRepo.AssignRole(c, userID, req.RoleID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewApiError("Failed to assign role"))
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Role assigned successfully"})
}
