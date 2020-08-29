package api

import (
	"github.com/fahribaharudin/edukacloud-toolkit/session"
	"github.com/labstack/echo/v4"
	"net/http"
)

const (
	AppSession = "App_Session"
	GeneralError = "error"
	GeneralSuccess = "success"
)

type ApplicationContext struct {
	echo.Context
	Session session.Session
}

func ParseContext(c echo.Context) *ApplicationContext {
	data := c.Get(AppSession)
	if data == nil {
		data = session.Session{}
	}
	return &ApplicationContext{
		Context: c, Session: data.(session.Session),
	}
}

func (c ApplicationContext) OK(data interface{}, message string) error {
	if data == nil {
		data = struct{}{}
	}

	response := DefaultResponse{
		Response: Response{
			Status:  GeneralSuccess,
			Message: message,
		},
		Data: data,
	}

	return c.Context.JSON(http.StatusOK, response)
}

func (c ApplicationContext) Error(err error, statusCode int, data interface{}) error {
	if data == nil {
		data = struct{}{}
	}

	response := DefaultResponse{
		Response: Response{
			Status:  GeneralError,
			Message: err.Error(),
		},
		Data: data,
	}

	return c.Context.JSON(statusCode, response)
}
