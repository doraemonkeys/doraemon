package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type StatusCode int

type Status struct {
	Code StatusCode `json:"code"`
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
	CodeErrorForbidden           StatusCode = 2008
)

var (
	StatusSuccess             = Status{Code: CodeSuccess, Msg: "success"}
	StatusInvalidParams       = Status{Code: CodeErrorInvalidParams, Msg: "invalid params"}
	StatusDatabase            = Status{Code: CodeErrorDatabase, Msg: "database error"}
	StatusUserNotExist        = Status{Code: CodeErrorUserNotExist, Msg: "user not exist"}
	StatusPassword            = Status{Code: CodeErrorPassword, Msg: "password error"}
	StatusTokenExpired        = Status{Code: CodeErrorTokenExpired, Msg: "token expired"}
	StatusUnauthorized        = Status{Code: CodeUnauthorized, Msg: "unauthorized"}
	StatusForbidden           = Status{Code: CodeErrorForbidden, Msg: "forbidden"}
	StatusInternalServerError = Status{Code: CodeErrorInternalServerError, Msg: "internal server error"}
)

type PageInfo struct {
	// To retrive all items, just set the page very large
	Page     int `json:"page" form:"page" binding:"required,min=1"`
	PageSize int `json:"pageSize" form:"pageSize" binding:"required,min=1,max=100"`
}

type PaginatedData[T any] struct {
	List     []T   `json:"list"`
	Total    int64 `json:"total"`
	Page     int   `json:"page"`
	PageSize int   `json:"pageSize"`
}

type Response struct {
	Status
	Data any `json:"data"`
}

func result(c *gin.Context, status Status, data any) {
	resultWithCode(c, http.StatusOK, status, data)
}

func resultWithCode(c *gin.Context, code int, status Status, data any) {
	c.JSON(code, Response{
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

func Fail(c *gin.Context, status Status) {
	result(c, status, nil)
}

func FailHttpCode(c *gin.Context, code int, status Status) {
	resultWithCode(c, code, status, nil)
}
