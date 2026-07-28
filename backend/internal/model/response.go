package model

import "github.com/gin-gonic/gin"

// Response is the unified API response format.
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func OK(c *gin.Context, data interface{}) {
	c.JSON(200, Response{Code: 0, Message: "ok", Data: data})
}

func Created(c *gin.Context, data interface{}) {
	c.JSON(201, Response{Code: 0, Message: "created", Data: data})
}

func Accepted(c *gin.Context, data interface{}) {
	c.JSON(202, Response{Code: 0, Message: "accepted", Data: data})
}

func BadRequest(c *gin.Context, msg string) {
	c.JSON(400, Response{Code: 400, Message: msg})
}

func NotFound(c *gin.Context, msg string) {
	c.JSON(404, Response{Code: 404, Message: msg})
}

func InternalError(c *gin.Context, msg string) {
	c.JSON(500, Response{Code: 500, Message: msg})
}
