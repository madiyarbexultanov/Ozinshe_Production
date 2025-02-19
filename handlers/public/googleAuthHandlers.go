package public


import (
	"context"
	"encoding/json"
	"net/http"
	"ozinshe_production/config"
	"ozinshe_production/models"
	"strconv"

	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/oauth2"
)

type GoogleUser struct {
	Id            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Picture       string `json:"picture"`
	Name          string `json:"name"`
}

func (h *AuthHandlers) GoogleLogin(c *gin.Context) {
	url := config.GoogleOAuthConfig.AuthCodeURL("state", oauth2.AccessTypeOffline)
	c.Redirect(http.StatusFound, url)
}


func (h *AuthHandlers) GoogleCallback(c *gin.Context) {
	code := c.Query("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, models.NewApiError("Authorization code not provided"))
		return
	}


	token, err := config.GoogleOAuthConfig.Exchange(context.Background(), code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewApiError("Failed to exchange token"))
		return
	}


	client := config.GoogleOAuthConfig.Client(context.Background(), token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewApiError("Failed to get user info"))
		return
	}
	defer resp.Body.Close()

	var googleUser GoogleUser
	if err := json.NewDecoder(resp.Body).Decode(&googleUser); err != nil {
		c.JSON(http.StatusInternalServerError, models.NewApiError("Failed to parse user info"))
		return
	}


	user, err := h.userRepo.FindByEmail(c, googleUser.Email)
	if err != nil {

		id, err := h.userRepo.SignUp(c, models.User{
			Email:        googleUser.Email,
			PasswordHash: "", 
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.NewApiError("Failed to create user"))
			return
		}
		user = models.User{Id: id, Email: googleUser.Email}
	}


	tokenString, err := generateJWT(user.Id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewApiError("Failed to generate token"))
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": tokenString, "user": googleUser})
}


func generateJWT(userId int) (string, error) {
	claims := jwt.RegisteredClaims{
		Subject: strconv.Itoa(userId),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(config.Config.JwtExpiresIn)),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(config.Config.JwtSecretKey))
}