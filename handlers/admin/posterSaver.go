package admin

import (
	"path/filepath"
	"github.com/google/uuid"
	"fmt"
	"mime/multipart"
	"github.com/gin-gonic/gin"
)


func savePoster(c *gin.Context, poster *multipart.FileHeader) (string, error) {
	filename := fmt.Sprintf("%s%s", uuid.NewString(), filepath.Ext(poster.Filename))
	filepath := fmt.Sprintf("images/%s", filename)
	err := c.SaveUploadedFile(poster, filepath)

	return filename, err
}