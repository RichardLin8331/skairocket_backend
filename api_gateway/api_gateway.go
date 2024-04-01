package main

import (
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
	r.POST("/search-stock", g.SearchStock)
	r.POST("/refresh", g.RefreshHandler)
	user_auth_router := r.Group("/user")
	user_auth_router.Use(g.AuthMiddleware())
	user_auth_router.POST("/add-favorite", g.AddFavorite)
	user_auth_router.POST("/delete-favorite", g.DeleteFavorite)

	r.Run(":8899")
}
