package main

import (
	"useraccountmodule"

	"github.com/gin-gonic/gin"
)

func main() {
	uam := useraccountmodule.NewUserAccountModule()
	uam.Init()
	r := gin.Default()
	r.POST("/UserLogin", uam.LoginHandler)
	r.POST("/UserCreate", uam.UserCreateHandler)

	r.Run(":8900")

}
