package http

import "github.com/gin-gonic/gin"

type KaizenServer struct {
	GinEngine *gin.Engine
}

func (s KaizenServer) Start() error {
	return s.GinEngine.Run()
}

func (s KaizenServer) RegisterRoutes() {
	s.GinEngine.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
}
