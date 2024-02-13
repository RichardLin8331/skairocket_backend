package useraccountmodule

import (
	"fmt"
	"net/http"
	"time"

	"gatewaymodule"

	"github.com/dgrijalva/jwt-go"
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
	MySQL_Port     int64
	MySQL_Datbase  string
}

type UserAccountModule struct {
	MySQL_Info      mysql_info_struct
	jwt_secret      string
	jwt_secret_byte []byte
	user_account_DB *gorm.DB
	iscontainer     bool
}

type User struct {
	Username string `json:"username" gorm:"primary_key"`
	Password string `json:"password"`
	Email    string `json:"email"`
	UserID   int64  `json:"userid" gorm:"auto_increase"`
}

func NewUserAccountModule() *UserAccountModule {
	return &UserAccountModule{iscontainer: false}

}

func (ua *UserAccountModule) Init() {
	vp := viper.New()
	vp.AddConfigPath("../modules/configs/")
	// todo here, change config name & type
	vp.SetConfigName("user_account.yaml")
	vp.SetConfigType("yaml")
	vp.ReadInConfig()
	vp.Unmarshal(&ua)
	ua.jwt_secret_byte = []byte(ua.jwt_secret)
	if !ua.iscontainer {
		ua.MySQL_Info.MySQL_IP = "127.0.0.1"
	}
	dsn := fmt.Sprintf("%s:%s@%s(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local", ua.MySQL_Info.MySQL_Username, ua.MySQL_Info.MySQL_Password, ua.MySQL_Info.MySQL_Network, ua.MySQL_Info.MySQL_IP, ua.MySQL_Info.MySQL_Port, ua.MySQL_Info.MySQL_Datbase)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("Error occurs while gorm working, " + err.Error())
	}

	if err := db.AutoMigrate(new(User)); err != nil {
		panic("Database error, " + err.Error())
	}
	ua.user_account_DB = db
	fmt.Println("MySQL Connected")
}

func (ua *UserAccountModule) LoginHandler(c *gin.Context) {
	login_content := &User{}

	if err := c.ShouldBind(&login_content); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	userinfo, err := findUser(ua.user_account_DB, login_content.Username)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Incorrect Username or Password"})
		return
	}

	// Check the credentials (this is a simplified example)
	if login_content.Username != userinfo.Username || login_content.Password != userinfo.Password {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Create and sign the JWT

	accessToken, err := ua.createAccessToken(userinfo.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create access token"})
		return
	}

	refreshToken, err := ua.createRefreshToken(userinfo.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create refresh token"})
		return
	}

	// Set the refresh token as an HttpOnly cookie
	c.SetCookie("rt", refreshToken, int(time.Hour*24*7), "/", "localhost", false, true)

	c.JSON(http.StatusOK, gin.H{
		"access_token": accessToken,
	})
}

func (ua *UserAccountModule) UserCreateHandler(c *gin.Context) {
	usercreate_content := &User{}

	if err := c.ShouldBind(&usercreate_content); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	_, err := findUser(ua.user_account_DB, usercreate_content.Username)
	if err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Exist Username"})
		return
	}

	ua.user_account_DB.Create(usercreate_content)
}

func (ua *UserAccountModule) createAccessToken(username string) (string, error) {
	claims := gatewaymodule.CustomClaims{
		UserName: username,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Second * 15).Unix(), // Access token expires in 15 minutes
			IssuedAt:  time.Now().Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString(ua.jwt_secret_byte)
	if err != nil {
		return "", err
	}

	return signedToken, nil
}

func (ua *UserAccountModule) createRefreshToken(username string) (string, error) {
	claims := gatewaymodule.CustomClaims{
		UserName: username,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour * 24 * 7).Unix(), // Refresh token expires in 7 days
			IssuedAt:  time.Now().Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	rftoken, err := token.SignedString(ua.jwt_secret_byte)
	if err != nil {
		return "", err
	}

	return rftoken, nil
}

// Not belong to user_account_module

func findUser(db *gorm.DB, username string) (*User, error) {
	user := new(User)
	user.Username = username
	err := db.First(&user).Error
	return user, err
}
