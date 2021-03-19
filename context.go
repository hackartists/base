package base

import "github.com/gin-gonic/gin"

type Context struct {
	*gin.Context
	User string
}
