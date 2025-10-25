package internal

import (
	"github.com/gin-gonic/gin"
)

func NewServer() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)

	engine := gin.Default()

	v1Group := engine.Group("/v1")
	{
		v1Group.GET("/models", ModelsHandler)
		v1Group.POST("/chat/completions", ChatHandler)
	}

	return engine
}
