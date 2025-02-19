package public

import (

	"net/http"
	"ozinshe_production/config"
	"ozinshe_production/models"
	"ozinshe_production/repositories"
	"ozinshe_production/logger"

	"strconv"
	"time"
	"net/mail"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"golang.org/x/crypto/bcrypt"
	"go.uber.org/zap"
)

type AuthHandlers struct {
	userRepo *repositories.UsersRepository
}

func NewAuthHandlers(userRepo *repositories.UsersRepository) *AuthHandlers {
	return &AuthHandlers{userRepo: userRepo}
}


type SignUpRequest struct {
	Email    		string `json:"email" binding:"required,email"`
	Password 		string `json:"password" binding:"required,min=8"`
	PasswordCheck 	string `json:"passwordCheck" binding:"required,min=8"`
}


type SignInRequest struct {
	Email 			string
	Password 		string
}

type ResetPasswordRequest struct {
	Password 		string `json:"password" binding:"required,min=8"`
	PasswordCheck 	string `json:"passwordCheck" binding:"required,min=8"`
}

// SignUp godoc
// @Summary      User Registration
// @Description  Registers a new user by providing an email, password, and password confirmation
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body handlers.SignUpRequest true "User registration request"
// @Success      200 {object} object{id=int} "User successfully created"
// @Failure      400 {object} models.ApiError "Validation error: invalid email, password mismatch, or weak password"
// @Failure      500 {object} models.ApiError "Server error: failed to hash password or create user"
// @Router       /auth/signUp [post]
func (h *AuthHandlers) SignUp(c *gin.Context) {
	logger := logger.GetLogger()
	var request SignUpRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		logger.Error("Invalid request data for sign-up", zap.String("error", err.Error()))
		c.JSON(http.StatusBadRequest, models.NewApiError(err.Error()))
		return
	}

	if _, err := mail.ParseAddress(request.Email); err != nil {
		logger.Error("Invalid email format", zap.String("email", request.Email))
		c.JSON(http.StatusBadRequest, models.NewApiError("Invalid email format"))
		return
	}

	if request.Password != request.PasswordCheck {
		logger.Warn("Passwords do not match", zap.String("email", request.Email))
		c.JSON(http.StatusBadRequest, models.NewApiError("Passwords do not match"))
		return
	}

	user, err := h.userRepo.FindByEmail(c, request.Email)
	if err != nil {
		logger.Error("Error checking email existence", zap.String("email", request.Email), zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewApiError("Server error: unable to check email"))
		return
	}

	if user.Email != "" {
		logger.Warn("Email already exists", zap.String("email", request.Email))
		c.JSON(http.StatusBadRequest, models.NewApiError("Email already exists"))
		return
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(request.Password), bcrypt.DefaultCost)
	if err != nil {
		logger.Error("Failed to hash password", zap.String("email", request.Email), zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewApiError("Failed to hash password"))
		return
	}

	id, err := h.userRepo.SignUp(c, models.User{
		Email: request.Email, PasswordHash: string(passwordHash),
	})
	if err != nil {
		logger.Error("Error creating user", zap.String("email", request.Email), zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewApiError("Couldn't create user"))
		return
	}

	claims := jwt.RegisteredClaims {
		Subject: strconv.Itoa(id),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(config.Config.JwtExpiresIn)),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(config.Config.JwtSecretKey))
	if err != nil {
		logger.Error("Couldn't generate JWT token", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewApiError("Couldn't generate JWT token"))
		return
	}

	logger.Info("User successfully registered", zap.String("email", request.Email), zap.Int("user_id", id))
	c.JSON(http.StatusOK, gin.H{"token": tokenString})
}



// SignIn godoc
// @Summary      User Sign In
// @Description  Authenticates a user by verifying the email and password, and returns a JWT token
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body handlers.SignInRequest true "User sign-in request"
// @Success      200 {object} object{token=string} "JWT token successfully generated"
// @Failure      400 {object} models.ApiError "Invalid payload"
// @Failure      401 {object} models.ApiError "Invalid credentials: wrong email or password"
// @Failure      500 {object} models.ApiError "Internal server error: failed to generate JWT token"
// @Router       /auth/signIn [post]
func (h *AuthHandlers) SignIn(c *gin.Context) {
	logger := logger.GetLogger()
	var request SignInRequest
	if err := c.BindJSON(&request); err != nil {
		logger.Error("Invalid sign-in request", zap.String("error", err.Error()))
		c.JSON(http.StatusBadRequest, models.NewApiError("Invalid payload"))
		return
	}

	user, err := h.userRepo.FindByEmail(c, request.Email)
	if err != nil {
		logger.Error("Error finding user by email", zap.String("email", request.Email), zap.Error(err))
		c.JSON(http.StatusUnauthorized, models.NewApiError("Invalid credentials"))
		return
	}
	
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(request.Password))
	if err != nil {
		logger.Warn("Invalid password", zap.String("email", request.Email))
		c.JSON(http.StatusUnauthorized, models.NewApiError("Invalid credentials"))
		return
	}

	claims := jwt.RegisteredClaims {
		Subject: strconv.Itoa(user.Id),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(config.Config.JwtExpiresIn)),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(config.Config.JwtSecretKey))
	if err != nil {
		logger.Error("Error generating JWT token", zap.String("email", request.Email), zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewApiError("Couldn't generate JWT token"))
		return
	}

	logger.Info("User successfully signed in", zap.String("email", request.Email))
	c.JSON(http.StatusOK, gin.H{"token": tokenString})
}

// SignOut godoc
// @Summary      User Sign Out
// @Description  Invalidates the user's current session (requires a valid JWT token)
// @Tags         auth
// @Accept       json
// @Produce      json
// @Success      200 "OK"
// @Failure      401 {object} models.ApiError "Authorization header required"
// @Router       /auth/signOut [post]
// @Security Bearer
func (h *AuthHandlers) SignOut(c *gin.Context) {
	logger := logger.GetLogger()
	logger.Info("User successfully signed out")

	c.Status(http.StatusOK)
}


