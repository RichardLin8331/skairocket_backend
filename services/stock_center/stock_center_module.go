package main

import (
	"fmt"

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
	UpdateDate        string  `json:"updatesate"`
	LastDayOpen       float64 `json:"lastdayopen"`
	LastDayClose      float64 `json:"lastdayclose"`
	PredictedPrice    float64 `json:"predictedprice"`
	PredictConfidence float64 `json:"predicctionconfidence"`
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

func (sc *StockCenterModule) SearchStockHandler() {

}
