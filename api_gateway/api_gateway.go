package main

import (
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"http://127.0.0.1:8080", "http://localhost:8080"}
	config.AllowMethods = []string{"POST", "GET"}
	config.AllowHeaders = []string{"Authorization", "Origin", "Connection", "Access-Control-Allow-Origin", "Content-Type"}
	config.AllowCredentials = true

	g := NewGatewayModule()
	g.Init()

	r := gin.Default()
	r.Use(cors.New(config))
	r.POST("/user-login", g.UserLogin)
	r.POST("/user-create", g.UserCreate)
	r.POST("/user-refresh", g.RefreshHandler)
	r.POST("/add-favorite", g.AddFavorite)
	user_beh := r.Group("/user/")
	user_beh.Use(g.AuthMiddleware())
	r.GET("/protected", g.AuthMiddleware(), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Protected Route Accessed"})
	})

	user_router := r.Group("/user_router")
	user_router.POST("/user-protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Protected Route Accessed"})
	})

	r.Run(":8899")
}
