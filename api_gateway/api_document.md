# API Document

+ `router` means this request can be sent without loggined in status
+ `user-router` means this request shoud be sent with loggined in status

## User API
+ Login
```golang
router.post("/login", {username string, password string}) {
    ctx.JSON(
        http.StatusOK,
        gin.H{
            "LoginStatus": "true",
            "access_token": accessToken
    })
}
```

+ Refresh
```golang
user-router.post("/refresh", {NULL}) {
    c.JSON(
        http.StatusOK, 
        gin.H{
            "access_token": accessToken
    })
}
```

+ Register
```golang
router.post("/register", {username, password, email}) {
    ctx.JSON(
        http.StatusOK,
        gin.H{
            "RegisterResult": "true",
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
            "RegsetPassword": "true",
    })
}
```

+ Change Password


## Stock API
+ Search 
```golang
router.post("/stock_search", {stock_num}) {
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
router.post("/stock_trending", {NULL}) {
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
            "AddStock": "true",
        }
)}

```

+ Remove from Favorite
```golang
user-router.post("/del-fav", {username string, stork_num int}) {
    ctx.JSON(
        http.StatusOK,
        gin.H{
            "DelStock": "true",
        }
)}
```
