package gatewaymodule

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

	"useraccountmodule"
)

type GatewayModule struct {
	JwtSecret string
	Network   string
}

func NewGatewayModule() *GatewayModule {
	return &GatewayModule{}
}

func (g *GatewayModule) Init() {
	vp := viper.New()
	vp.AddConfigPath("../services/modules/configs/")
	// todo here, change config name & type
	vp.SetConfigName("gateway.yaml")
	vp.SetConfigType("yaml")
	vp.ReadInConfig()
	vp.Unmarshal(&g)
}

func (g *GatewayModule) UserCreate(c *gin.Context) {
	user_info := useraccountmodule.User{}
	if err := c.ShouldBind(&user_info); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	user_JSON, _ := json.Marshal(user_info)
	resp, _ := http.Post(g.Network+":8900"+"/UserCreate", "application/json", bytes.NewBuffer(user_JSON))
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
	login_content := &useraccountmodule.User{}

	if err := c.ShouldBind(&login_content); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	login_JSON, _ := json.Marshal(login_content)

	resp, _ := http.Post(g.Network+":8900"+"/UserLogin", "application/json", bytes.NewBuffer(login_JSON))

	body, _ := io.ReadAll(resp.Body)
	tokens := struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
	}{}
	err := json.Unmarshal(body, &tokens)
	if err != nil {
		print("error\n")
	}

	c.SetCookie("rt", tokens.RefreshToken, int(time.Hour*24*7), "/", g.Network, false, true)

	c.JSON(http.StatusOK, gin.H{
		"login_result": true,
		"access_token": tokens.AccessToken,
	})
}

func (g *GatewayModule) VerifyToken(tokenString string) (*useraccountmodule.CustomClaims, error) {
	claims := &useraccountmodule.CustomClaims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return g.JwtSecret, nil
	})

	if err != nil || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return claims, nil
}

func (g *GatewayModule) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Retrieve the access token from the Authorization header
		accessToken := strings.Split(c.GetHeader("Authorization"), " ")[1]
		print(accessToken)

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
		print("\nPass Verify\n")

		// Attach the user ID to the context for use in the protected route
		c.Set("username", claims.Username)

		c.Next()
	}
}
