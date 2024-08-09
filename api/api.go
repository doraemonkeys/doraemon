package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type StatusCode int

type Status struct {
	Code StatusCode `json:"status"`
	Msg  string     `json:"msg"`
}

func (s Status) Send(c *gin.Context) {
	result(c, s, nil)
}

const (
	CodeSuccess                  StatusCode = 1
	CodeError                    StatusCode = 0
	CodeErrorInvalidParams       StatusCode = 2001
	CodeErrorDatabase            StatusCode = 2002
	CodeErrorUserNotExist        StatusCode = 2003
	CodeErrorPassword            StatusCode = 2004
	CodeErrorTokenExpired        StatusCode = 2005
	CodeUnauthorized             StatusCode = 2006
	CodeErrorInternalServerError StatusCode = 2007
)

var (
	StatusSuccess             = Status{Code: CodeSuccess, Msg: "success"}
	StatusInvalidParams       = Status{Code: CodeErrorInvalidParams, Msg: "invalid params"}
	StatusDatabase            = Status{Code: CodeErrorDatabase, Msg: "database error"}
	StatusUserNotExist        = Status{Code: CodeErrorUserNotExist, Msg: "user not exist"}
	StatusPassword            = Status{Code: CodeErrorPassword, Msg: "password error"}
	StatusTokenExpired        = Status{Code: CodeErrorTokenExpired, Msg: "token expired"}
	StatusUnauthorized        = Status{Code: CodeUnauthorized, Msg: "unauthorized"}
	StatusInternalServerError = Status{Code: CodeErrorInternalServerError, Msg: "internal server error"}
)

type Body struct {
	Status
	Data any `json:"data"`
}

func result(c *gin.Context, status Status, data any) {
	c.JSON(http.StatusOK, Body{
		status,
		data,
	})
}

func Ok(c *gin.Context) {
	result(c, StatusSuccess, nil)
}

func OkWithData(c *gin.Context, data any) {
	result(c, StatusSuccess, data)
}

func OkWithMsg(c *gin.Context, msg string) {
	result(c, Status{Code: CodeSuccess, Msg: msg}, nil)
}

func FailInternalError(c *gin.Context) {
	result(c, StatusInternalServerError, nil)
}

func FailWithMsg(c *gin.Context, msg string) {
	result(c, Status{Code: CodeError, Msg: msg}, nil)
}
