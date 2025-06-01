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
	// // SortBy: field name, empty string means no sort
	// SortBy string `json:"sortBy" form:"sortBy"`
	// // SortType: "asc" or "desc".
	// SortType string `json:"sortType" form:"sortType" binding:"omitempty,oneof=asc desc"`
}

type PaginatedData[T any] struct {
	List     []T   `json:"list"`
	Total    int64 `json:"total"`
	Page     int   `json:"page"`
	PageSize int   `json:"pageSize"`
}

type KeysetPage struct {
	// How many items to return per page.
	Limit int `json:"limit" form:"limit" binding:"required,min=1,max=100"`
	// Field name to sort by (e.g., "created_at", "name").
	// Must be a field suitable for keyset pagination (indexed, relatively unique).
	SortBy string `json:"sortBy" form:"sortBy" binding:"required"`
	// default is ASC, if true, then DESC
	DESC bool `json:"desc" form:"desc"`
	// An opaque cursor indicating the last item from the previous page (for forward pagination).
	// The client should send the 'nextCursor' received from the previous response here.
	After any `json:"after" form:"after"`
	// An opaque cursor indicating the first item from the previous page (for backward pagination).
	// The client should send the 'previousCursor' received from the previous response here.
	Before any `json:"before" form:"before"`
}

type KeysetPaginatedData[T any] struct {
	List []T `json:"list"`
	// Opaque cursor to fetch the next page. Null if no next page.
	NextCursor *string `json:"nextCursor"`
	// Opaque cursor to fetch the previous page. Null if no previous page.
	PreviousCursor *string `json:"previousCursor"`
	// Note: Total count is typically omitted in keyset pagination.
	// Total int64 `json:"-"`
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
