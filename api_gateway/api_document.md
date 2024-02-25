# API Document

+ Routers
    + `router` means this request can be sent without loggined in status
    + `user-router` means this request shoud be sent with loggined in status
+ Naming Principles
    + `Camel-Case Naming` for golang structs
    + All lower case letters with `_` for JSON keys
    + All lower case letters with `-` for API Gateway handlers
+ Config
    + All config files are palced in ./services/modules/configs/
    + Env of docker-compose is ./.env


## User API

+ User Create
```golang
router.post("/user-create", {username, password, email}) {
    ctx.JSON(
        http.StatusOK,
        gin.H{
            "create_result": true,
    })
}
```

+ Login
```golang
router.post("/user-login", {username string, password string}) {
    ctx.JSON(
        http.StatusOK,
        gin.H{
            "login_result": "true",
            "access_token": AccessToken
    })
}
```

+ Refresh
```golang
user-router.post("/user-refresh", {NULL}) {
    c.JSON(
        http.StatusOK, 
        gin.H{
            "access_token": AccessToken
    })
}
```

+ Forget Password  

**Editing**
```golang
user-router.post("/forget-pwd", {username, email}) {
    ctx.JSON(
        http.StatusOK,
        gin.H{
            "user_valid": true,
    })
}
```

+ Change Password


## Stock API
+ Search 
```golang
router.post("/stock-search", {stock_num}) {
    ctx.JSON(
        http.StatusOK,
        StockInfo
)}

type StockInfo struct {
    StockName string `json: "stock_name"`
    CurrentPrice float64 `json: "current_price"`
    PredictingValue float64 `json: "predicting_price"`
    PredictionConfidence float64 `json: "prediction_confidence"`

}
```

+ Trending
```golang
router.post("/stock-trending", {NULL}) {
    ctx.JSON(
        http.StatusOK,
        StockList
)}

type StockList struct {
    StockInfoList []StockInfo `json:"stock_info_list"`
}

```

## User-Stock API
+ Show Favorite
```golang
user-router.post("/show-fav", {username string}) {
    ctx.JSON(
        http.StatusOK,
        StockList
)}


```
+ Add to Favorite
```golang
user-router.post("/add-fav", {username string, stork_num int}) {
    ctx.JSON(
        http.StatusOK,
        gin.H{
            "add_stock": "true",
        }
)}

```

+ Remove from Favorite
```golang
user-router.post("/del-fav", {username string, stork_num int}) {
    ctx.JSON(
        http.StatusOK,
        gin.H{
            "del_stock": "true",
        }
)}
```
