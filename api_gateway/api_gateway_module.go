package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

type GatewayModule struct {
	JwtSecret         string
	Network           string
	SkAI_user_account string
	SkAI_stock_center string
	jwt_secret_byte   []byte
}

func NewGatewayModule() *GatewayModule {
	return &GatewayModule{}
}

type User struct {
	UserID         int    `json:"userid" gorm:"autoIncrement;primaryKey;not null"`
	Username       string `json:"username"`
	Password       string `json:"password"`
	Email          string `json:"email"`
	ProfilePicture string `json:"profile_picture"`
}

type CustomClaims struct {
	Username string `json:"username"`
	jwt.StandardClaims
}

func (g *GatewayModule) Init() {
	vp := viper.New()
	vp.AddConfigPath("./configs/")
	// todo here, change config name & type
	vp.SetConfigName("gateway.yaml")
	vp.SetConfigType("yaml")
	vp.ReadInConfig()
	vp.Unmarshal(&g)
	g.jwt_secret_byte = []byte(g.JwtSecret)
}

func (g *GatewayModule) UserCreate(c *gin.Context) {
	user_info := &User{}
	if err := c.ShouldBind(&user_info); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	user_JSON, _ := json.Marshal(user_info)
	resp, _ := http.Post("http://"+g.SkAI_user_account+":8900"+"/UserCreate", "application/json", bytes.NewBuffer(user_JSON))
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode == 200 {
		create_res := struct {
			Result bool `json:"create_result"`
		}{}
		err := json.Unmarshal(body, &create_res)
		if err != nil {
			print("error\n")
		}

		if create_res.Result {
			c.JSON(http.StatusOK, gin.H{"user_create_result": "true"})
		}
	} else {
		create_res := struct {
			Result bool   `json:"create_result"`
			Error  string `json:"error"`
		}{}
		err := json.Unmarshal(body, &create_res)
		if err != nil {
			print("error\n")
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": create_res.Error})

	}

}

func (g *GatewayModule) UserLogin(c *gin.Context) {
	login_content := &User{}

	if err := c.ShouldBind(&login_content); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	login_JSON, _ := json.Marshal(login_content)

	resp, err1 := http.Post("http://"+g.SkAI_user_account+":8900"+"/UserLogin", "application/json", bytes.NewBuffer(login_JSON))
	if err1 != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	body, _ := io.ReadAll(resp.Body)
	login_result := struct {
		Success        bool     `json:"success"`
		ProfilePicture string   `json:"profile_picture"`
		FavoriteList   []string `json:"favorite_list"`
	}{}
	err2 := json.Unmarshal(body, &login_result)
	if err2 != nil {
		print("error\n")
	}
	userAccessToken, _ := g.createAccessToken(login_content.Username)
	userRefreshToken, err3 := g.createRefreshToken(login_content.Username)
	if err3 != nil {
		print(err3.Error())
	}

	c.SetCookie("refreshtoken", userRefreshToken, int(time.Hour*24*7), "/", "127.0.0.1", false, true)

	c.JSON(http.StatusOK, gin.H{
		"login_result":    true,
		"accesstoken":     userAccessToken,
		"profile_picture": login_result.ProfilePicture,
		"favorite_list":   login_result.FavoriteList,
	})
}

func (g *GatewayModule) AddFavorite(c *gin.Context) {
	postcontent := struct {
		Username string   `json:"username"`
		Stocknum []string `json:"stocknum"`
	}{}

	if err := c.ShouldBind(&postcontent); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}
	post_JSON, _ := json.Marshal(postcontent)
	_, err1 := http.Post("http://"+g.SkAI_user_account+":8900"+"/AddFavorite", "application/json", bytes.NewBuffer(post_JSON))
	if err1 == nil {
		c.JSON(http.StatusOK, gin.H{"add_result": true})
	}
}

func (g *GatewayModule) DeleteFavorite(c *gin.Context) {
	postcontent := struct {
		Username string   `json:"username"`
		Stocknum []string `json:"stocknum"`
	}{}

	if err := c.ShouldBind(&postcontent); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}
	post_JSON, _ := json.Marshal(postcontent)
	_, err1 := http.Post("http://"+g.SkAI_user_account+":8900"+"/DeleteFavorite", "application/json", bytes.NewBuffer(post_JSON))
	if err1 == nil {
		c.JSON(http.StatusOK, gin.H{"delete_result": true})
	}
}

func (g *GatewayModule) SearchStock(c *gin.Context) {
	postcontent := struct {
		Stocknum string `json:"stocknum"`
	}{}

	if err := c.ShouldBind(&postcontent); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}
	post_JSON, _ := json.Marshal(postcontent)
	resp, err := http.Post("http://"+g.SkAI_stock_center+":8901"+"/SearchStock", "application/json", bytes.NewBuffer(post_JSON))
	if err != nil {
		print(err.Error())
	}
	body, _ := io.ReadAll(resp.Body)
	pred_resp := struct {
		Resultprice      string `json:"predicted_price"`
		Resultconfidence string `json:"prediction_confidence"`
	}{}
	err2 := json.Unmarshal(body, &pred_resp)
	if err2 != nil {
		print("error\n")
	}
	c.JSON(http.StatusOK, pred_resp)

}

func (g *GatewayModule) createAccessToken(username string) (string, error) {
	claims := CustomClaims{
		Username: username,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Second * 15).Unix(), // Access token expires in 15 minutes
			IssuedAt:  time.Now().Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString(g.jwt_secret_byte)
	if err != nil {
		return "", err
	}

	return signedToken, nil
}

func (g *GatewayModule) createRefreshToken(username string) (string, error) {
	claims := CustomClaims{
		Username: username,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour * 24 * 7).Unix(), // Refresh token expires in 7 days
			IssuedAt:  time.Now().Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	rftoken, err := token.SignedString(g.jwt_secret_byte)
	if err != nil {
		return "", err
	}

	return rftoken, nil
}

func (g *GatewayModule) VerifyToken(tokenString string) (*CustomClaims, error) {
	claims := &CustomClaims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return g.jwt_secret_byte, nil
	})

	if err != nil || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return claims, nil
}

func (g *GatewayModule) RefreshHandler(c *gin.Context) {
	// Retrieve the refresh token from the HttpOnly cookie
	refreshToken, err := c.Cookie("refreshtoken")
	if err != nil {
		print("No Token\n")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
		return
	}

	// Verify the refresh token
	claims, err := g.VerifyToken(refreshToken)
	if err != nil {
		print("bad token\n")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
		return
	}

	// Create and sign a new access token
	accessToken, err := g.createAccessToken(claims.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create access token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"accesstoken": accessToken})
}

func (g *GatewayModule) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Retrieve the access token from the Authorization header
		accessToken := strings.Split(c.GetHeader("Authorization"), " ")[1]

		if accessToken == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Access token missing"})
			c.Abort()
			return
		}

		// Verify the access token
		claims, err := g.VerifyToken(accessToken)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid access token"})
			c.Abort()
			return
		}

		// Attach the user ID to the context for use in the protected route
		c.Set("username", claims.Username)

		c.Next()
	}
}
