package responses

import (
	"errors"

	"github.com/gin-gonic/gin"
)

var ErrNotFound = errors.New("product not found")

type Error struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}

type Success struct {
	Status  int         `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

func SendError(ctx *gin.Context, status int, message string, err error) {
	errorMessage := ""
	if err != nil {
		errorMessage = err.Error()
	}
	ctx.JSON(status, Error{
		Status:  status,
		Message: message,
		Error:   errorMessage,
	})
}

func SendSuccess(ctx *gin.Context, status int, message string, data interface{}) {
	ctx.JSON(status, Success{
		Status:  status,
		Message: message,
		Data:    data,
	})
}
