package admin

import (
	"net/http"
	"ozinshe_production/models"
	"ozinshe_production/repositories"

	"github.com/gin-gonic/gin"
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

func (h *UsersHandler) FindAll(c *gin.Context) {
	users, err := h.userRepo.FindAll(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewApiError("couldn't load users"))
		return
	}
	dtos := make([]userResponse, 0, len(users))
	for _, u := range users {
		dtos = append(dtos, userResponse{Id: u.Id, Name: u.Name, Email: u.Email})
	}
	c.JSON(http.StatusOK, dtos)
}
