package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
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
	MySQL_Info      mysql_info_struct
	Stock_Info_IP   string
	Stock_Info_port string
	stock_info_DB   *gorm.DB
	iscontainer     bool
}

type StockInfo struct {
	Stocknum   string `json:"stocknum" gorm:"primaryKey;not null"`
	UpdateDate string `json:"updatesate"`
	//LastDayOpen       float64 `json:"lastdayopen"`
	//LastDayClose      float64 `json:"lastdayclose"`
	PredictedPrice    string `json:"predicted_price"`
	PredictConfidence string `json:"prediction_confidence"`
}

type post_content struct {
	Stocknum   string `json:"stocknum"`
	Stockmonth int    `json:"stockmonth"`
}

type resp_content struct {
	Resultprice      float64 `json:"predictedprice"`
	Resultconfidence float64 `json:"predictionconfidence"`
}

type reurn_stock_pred struct {
	Resultprice      string `json:"predicted_price"`
	Resultconfidence string `json:"prediction_confidence"`
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
	now := time.Now()
	today_date := now.Format("2006-01-02")
	today_month, _ := strconv.Atoi(strings.Split(today_date, "-")[1])
	pc := post_content{Stockmonth: today_month}
	if err := c.ShouldBind(&pc); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	si, stock_status := sc.checkStockStatus(pc.Stocknum, today_date)

	if stock_status == 1 {
		c.JSON(http.StatusOK, si)
		return
	}

	pc_json, _ := json.Marshal(pc)

	resp, err2 := http.Post("http://"+sc.Stock_Info_IP+sc.Stock_Info_port+"/StockPredict", "application/json", bytes.NewBuffer(pc_json))

	if err2 != nil {
		panic(err2.Error())
	}

	body, _ := io.ReadAll(resp.Body)

	pred_result := resp_content{}
	err3 := json.Unmarshal(body, &pred_result)
	if err3 != nil {
		print("error\n")
	}

	stock_pred := reurn_stock_pred{Resultprice: fmt.Sprintf("%.0f", pred_result.Resultprice), Resultconfidence: fmt.Sprintf("%.2f", pred_result.Resultconfidence)}
	si.PredictedPrice = stock_pred.Resultprice
	si.PredictConfidence = stock_pred.Resultconfidence

	sc.stock_info_DB.Save(si)
	c.JSON(http.StatusOK, stock_pred)
}

func (sc *StockCenterModule) checkStockStatus(stocknum string, today_date string) (*StockInfo, int) {
	si := &StockInfo{}
	var stock_status int

	err := sc.stock_info_DB.First(&si, "stocknum = ?", stocknum).Error
	if err != nil {
		stock_status = 2
		si.Stocknum = stocknum
	} else if si.UpdateDate != today_date {
		stock_status = 3
	} else {
		stock_status = 1

	}
	si.UpdateDate = today_date
	return si, stock_status
}
