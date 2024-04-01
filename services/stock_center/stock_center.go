package main

import "github.com/gin-gonic/gin"

/*
	type stock_basic struct {
		Stocknum   string `json:"stocknum"`
		Stockmonth string `json:"stockmonth"`
	}
*/
func main() {
	scm := &StockCenterModule{iscontainer: true}
	scm.Init()

	r := gin.Default()
	r.POST("/SearchStock", scm.SearchStockHandler)

	r.Run(":8901")

}
