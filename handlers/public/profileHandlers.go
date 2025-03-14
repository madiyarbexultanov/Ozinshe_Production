package public

import (
	"net/http"
	"ozinshe_production/logger"
	"ozinshe_production/models"
	"ozinshe_production/repositories"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"

	"go.uber.org/zap"
)

type ProfilesHandler struct {
	userRepo *repositories.UsersRepository
}

func NewProfilesHandler(repo *repositories.UsersRepository) *ProfilesHandler {
	return &ProfilesHandler{userRepo: repo}
}

type profileResponse struct {
	Name  		string `json:"name"`
	Email 		string `json:"email"`
	Phone 		string `json:"phone_number"`
	Birthday 	string `json:"birthday"`
}

type updateRequest struct {
	Name  		string `json:"name"`
	Email 		string `json:"email"`
	Phone 		string `json:"phone_number"`
	Birthday 	string `json:"birthday"`
}

type ResetPasswordRequest struct {
	Password 		string `json:"password" binding:"required,min=8"`
	PasswordCheck 	string `json:"passwordCheck" binding:"required,min=8"`
}

func (h *ProfilesHandler) UserProfile(c *gin.Context) {
	logger := logger.GetLogger()

	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)

	if err != nil {
		logger.Error("Invalid user id", zap.String("id", idStr))
		c.JSON(http.StatusBadRequest, models.NewApiError("Invalid user id"))
		return
	}

	user, err := h.userRepo.UserProfile(c, id)
	if err != nil {
		logger.Error("Failed to find user", zap.Int("id", id), zap.Error(err))
		c.JSON(http.StatusNotFound, models.NewApiError(err.Error()))
		return
	}

	response := profileResponse{
		Name:     user.Name,
		Email:    user.Email,
		Phone:    user.Phone,
		Birthday: user.Birthday.Format("2006-01-02"),
	}

	logger.Info("User found", zap.Int("id", id), zap.String("name", user.Name))
	c.JSON(http.StatusOK, response)
}

func (h *ProfilesHandler) Update(c *gin.Context) {
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

	var updateUser updateRequest
	err = c.BindJSON(&updateUser)
	if err != nil {
		logger.Error("Couldn't bind JSON", zap.Error(err))
		c.JSON(http.StatusBadRequest, models.NewApiError("Couldn't bind json"))
		return
	}

	userToUpdate := models.User{
		Email:    updateUser.Email,
		Name:     updateUser.Name,
		Phone:    updateUser.Phone,
	}
	
	// Парсим дату
	if updateUser.Birthday != "" {
		birthDate, err := time.Parse("2006-01-02", updateUser.Birthday)
		if err != nil {
			logger.Error("Invalid birthday format", zap.Error(err))
			c.JSON(http.StatusBadRequest, models.NewApiError("Invalid birthday format, use YYYY-MM-DD"))
			return
		}
		userToUpdate.Birthday = birthDate
	}

	err = h.userRepo.Update(c, id, userToUpdate)
	if err != nil {
		logger.Error("Failed to update user", zap.Int("id", id), zap.Error(err))
		c.JSON(http.StatusNotFound, models.NewApiError(err.Error()))
		return
	}

	logger.Info("User updated successfully", zap.Int("id", id))
	c.Status(http.StatusOK)
}

func (h *ProfilesHandler) ChangePassword(c *gin.Context) {
	userId, _ := c.Get("userId")

	id, ok := userId.(int)
	if !ok {
		c.JSON(http.StatusUnauthorized, models.NewApiError("Invalid user ID"))
		return
	}

	var request ResetPasswordRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, models.NewApiError("Invalid payload"))
		return
	}

	if request.Password != request.PasswordCheck {
		c.JSON(http.StatusBadRequest, models.NewApiError("Passwords do not match"))
		return
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(request.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewApiError("Failed to hash password"))
		return
	}

	if err := h.userRepo.ChangePasswordHash(c, id, string(passwordHash)); err != nil {
		c.JSON(http.StatusInternalServerError, models.NewApiError("Failed to update password"))
		return
	}

	c.Status(http.StatusOK)
}