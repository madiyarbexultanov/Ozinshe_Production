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

// UserProfile godoc
// @Summary Get user profile
// @Description Retrieves a user profile by ID
// @Tags Public Profile
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Success 200 {object} profileResponse
// @Failure 400 {object} models.ApiError "Invalid user id"
// @Failure 404 {object} models.ApiError "User not found"
// @Router /public/profile/{id} [get]
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

	logger.Info("User profile retrieved successfully", zap.Int("id", id), zap.String("name", user.Name))
	c.JSON(http.StatusOK, response)
}

// Update godoc
// @Summary Update user profile
// @Description Updates the details of a user profile by ID
// @Tags Public Profile
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Param body body updateRequest true "Update Profile Data"
// @Success 200 {string} string "Profile updated successfully"
// @Failure 400 {object} models.ApiError "Invalid input data"
// @Failure 404 {object} models.ApiError "User not found"
// @Router /public/profile/{id} [put]
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
		logger.Error("Failed to find user for update", zap.Int("id", id), zap.Error(err))
		c.JSON(http.StatusBadRequest, models.NewApiError(err.Error()))
		return
	}

	var updateUser updateRequest
	err = c.BindJSON(&updateUser)
	if err != nil {
		logger.Error("Couldn't bind JSON for update", zap.Error(err))
		c.JSON(http.StatusBadRequest, models.NewApiError("Couldn't bind json"))
		return
	}

	userToUpdate := models.User{
		Email: updateUser.Email,
		Name:  updateUser.Name,
		Phone: updateUser.Phone,
	}

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

// ChangePassword godoc
// @Summary Change user password
// @Description Allows a user to change their password
// @Tags Public Profile
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Param body body ResetPasswordRequest true "Password Reset Data"
// @Success 200 {string} string "Password changed successfully"
// @Failure 400 {object} models.ApiError "Invalid input data"
// @Failure 401 {object} models.ApiError "Unauthorized"
// @Failure 500 {object} models.ApiError "Internal server error"
// @Router /public/profile/changepassword/{id} [put]
func (h *ProfilesHandler) ChangePassword(c *gin.Context) {
	logger := logger.GetLogger()

	userId, exists := c.Get("userId")
	if !exists {
		logger.Error("Missing user ID in context")
		c.JSON(http.StatusUnauthorized, models.NewApiError("Invalid user ID"))
		return
	}

	id, ok := userId.(int)
	if !ok {
		logger.Error("Invalid user ID type", zap.Any("userId", userId))
		c.JSON(http.StatusUnauthorized, models.NewApiError("Invalid user ID"))
		return
	}

	var request ResetPasswordRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		logger.Error("Invalid JSON payload for password change", zap.Error(err))
		c.JSON(http.StatusBadRequest, models.NewApiError("Invalid payload"))
		return
	}

	if request.Password != request.PasswordCheck {
		logger.Warn("Passwords do not match", zap.Int("id", id))
		c.JSON(http.StatusBadRequest, models.NewApiError("Passwords do not match"))
		return
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(request.Password), bcrypt.DefaultCost)
	if err != nil {
		logger.Error("Failed to hash password", zap.Int("id", id), zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewApiError("Failed to hash password"))
		return
	}

	err = h.userRepo.ChangePasswordHash(c, id, string(passwordHash))
	if err != nil {
		logger.Error("Failed to update password", zap.Int("id", id), zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewApiError("Failed to update password"))
		return
	}

	logger.Info("Password changed successfully", zap.Int("id", id))
	c.Status(http.StatusOK)
}