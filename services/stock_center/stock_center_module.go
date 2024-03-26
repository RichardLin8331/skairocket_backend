package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type mysql_info_struct struct {
	MySQL_Username string
	MySQL_Password string
	MySQL_Network  string
	MySQL_IP       string
	MySQL_Port     int
	MySQL_Datbase  string
}

type StockCenterModule struct {
	MySQL_Info    mysql_info_struct
	stock_info_DB *gorm.DB
	iscontainer   bool
}

type StockInfo struct {
	Stocknum   int    `json:"stocknum" gorm:"primaryKey;not null"`
	UpdateDate string `json:"updatesate"`
	//LastDayOpen       float64 `json:"lastdayopen"`
	//LastDayClose      float64 `json:"lastdayclose"`
	PredictedPrice    float64 `json:"predictedprice"`
	PredictConfidence float64 `json:"predictionconfidence"`
}

type post_content struct {
	Stocknum   int    `json:"stocknum"`
	Stockmonth string `json:"stockmonth"`
}

type resp_content struct {
	Resultprice      float64 `json:"predictedprice"`
	Resultconfidence float64 `json:"predictionconfidence"`
}

func (sc *StockCenterModule) Init() {
	vp := viper.New()
	vp.AddConfigPath("./configs/")
	// todo here, change config name & type
	vp.SetConfigName("stock_center.yaml")
	vp.SetConfigType("yaml")
	vp.ReadInConfig()
	vp.Unmarshal(&sc)
	if !sc.iscontainer {
		sc.MySQL_Info.MySQL_IP = "127.0.0.1"
	}
	dsn := fmt.Sprintf("%s:%s@%s(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local", sc.MySQL_Info.MySQL_Username, sc.MySQL_Info.MySQL_Password, sc.MySQL_Info.MySQL_Network, sc.MySQL_Info.MySQL_IP, sc.MySQL_Info.MySQL_Port, sc.MySQL_Info.MySQL_Datbase)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("Error occurs while gorm working, " + err.Error())
	}

	if err := db.AutoMigrate(new(StockInfo)); err != nil {
		panic("Database error, " + err.Error())
	}
	sc.stock_info_DB = db
	fmt.Println("MySQL Connected")
}

func (sc *StockCenterModule) SearchStockHandler(c *gin.Context) {

	pc := post_content{}
	if err := c.ShouldBind(&pc); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	si, stock_status := sc.checkStockStatus(pc.Stocknum)

	if stock_status == 1 {
		c.JSON(http.StatusOK, si)
		return
	}

	pc_json, _ := json.Marshal(pc)

	resp, err2 := http.Post("http://localhost"+":19982"+"/StockPredict", "application/json", bytes.NewBuffer(pc_json))

	if err2 != nil {
		panic(err2.Error())
	}

	body, _ := io.ReadAll(resp.Body)

	pred_result := resp_content{}
	err3 := json.Unmarshal(body, &pred_result)
	if err3 != nil {
		print("error\n")
	}
	si.PredictedPrice = pred_result.Resultprice
	si.PredictConfidence = pred_result.Resultconfidence

	sc.stock_info_DB.Save(si)
	c.JSON(http.StatusOK, pred_result)
	print("Successfully Update ")
}

func (sc *StockCenterModule) checkStockStatus(stocknum int) (*StockInfo, int) {
	si := &StockInfo{}
	var stock_status int
	now := time.Now()
	today_date := now.Format("2006-01-02")

	err := sc.stock_info_DB.First(&si, "stocknum = ?", stocknum).Error
	if err != nil {
		stock_status = 2
		si.Stocknum = stocknum
	} else if si.UpdateDate != today_date {
		stock_status = 3
	} else {
		stock_status = 1
		print(si.PredictedPrice, "\n")

	}
	si.UpdateDate = today_date
	return si, stock_status
}
