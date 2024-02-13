package gatewaymodule

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

type CustomClaims struct {
	UserName string `json:"user_name"`
	jwt.StandardClaims
}

type GatewayModule struct {
	jwtSecret []byte
	Network   string
}

func NewGatewayModule() *GatewayModule {
	return &GatewayModule{}
}

func (g *GatewayModule) init() {
	vp := viper.New()
	vp.AddConfigPath("../modules/configs/")
	// todo here, change config name & type
	vp.SetConfigName("gateway.yaml")
	vp.SetConfigType("yaml")
	vp.ReadInConfig()
	vp.Unmarshal(&g)
}

func (g *GatewayModule) VerifyToken(tokenString string) (*CustomClaims, error) {
	claims := &CustomClaims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return g.jwtSecret, nil
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
		c.Set("username", claims.UserName)

		c.Next()
	}
}
