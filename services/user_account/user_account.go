package main

import (
	"context"

	"github.com/gin-gonic/gin"
)

func main() {
	uam := NewUserAccountModule()
	defer uam.user_mongo_client.Disconnect(context.Background())
	uam.Init()
	r := gin.Default()
	r.POST("/UserLogin", uam.LoginHandler)
	r.POST("/UserCreate", uam.UserCreateHandler)
	r.POST("/AddFavorite", uam.AddFavoriteHandler)
	r.POST("/DeleteFavorite", uam.DelFavoriteHandler)

	r.Run(":8900")

}
