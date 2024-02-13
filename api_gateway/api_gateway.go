package main

import (
	"net/http"

	"gatewaymodule"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"http://localhost:8080"}
	config.AllowMethods = []string{"POST", "GET"}
	config.AllowHeaders = []string{"Authorization", "Origin", "Connection"}
	config.AllowCredentials = true

	g := gatewaymodule.NewGatewayModule()

	r := gin.Default()
	r.Use(cors.New(config))
	r.GET("/protected", g.AuthMiddleware(), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Protected Route Accessed"})
	})

	r.Run(":8899")
}
