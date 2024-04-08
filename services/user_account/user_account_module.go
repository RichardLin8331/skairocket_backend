package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type db_info_struct struct {
	MySQL_Username string
	MySQL_Password string
	MySQL_Network  string
	MySQL_IP       string
	MySQL_Port     int
	MySQL_Datbase  string

	Mongo_IP       string
	Mongo_port     string
	Mongo_Username string
	Mongo_Password string
}

type UserAccountModule struct {
	DB_Info                  db_info_struct
	Jwt_secret               string
	jwt_secret_byte          []byte
	user_account_DB          *gorm.DB
	iscontainer              bool
	user_mongo_client        *mongo.Client
	user_favorite_collection *mongo.Collection
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

type StockFavorite struct {
	UserID        int      `bson:"userid"`
	FavoriteStock []string `bson:"favorite_stock"`
}

func NewUserAccountModule() *UserAccountModule {
	return &UserAccountModule{iscontainer: true}
}

func (ua *UserAccountModule) Init() {
	vp := viper.New()
	vp.AddConfigPath("./configs/")
	// todo here, change config name & type
	vp.SetConfigName("user_account.yaml")
	vp.SetConfigType("yaml")
	vp.ReadInConfig()
	vp.Unmarshal(&ua)
	ua.jwt_secret_byte = []byte(ua.Jwt_secret)
	if !ua.iscontainer {
		ua.DB_Info.MySQL_IP = "127.0.0.1"
	}
	dsn := fmt.Sprintf("%s:%s@%s(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local", ua.DB_Info.MySQL_Username, ua.DB_Info.MySQL_Password, ua.DB_Info.MySQL_Network, ua.DB_Info.MySQL_IP, ua.DB_Info.MySQL_Port, ua.DB_Info.MySQL_Datbase)
	print(dsn, " jwt ", ua.Jwt_secret, "\n")
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("Error occurs while gorm working, " + err.Error())
	}

	if err := db.AutoMigrate(new(User)); err != nil {
		panic("Database error, " + err.Error())
	}
	ua.user_account_DB = db
	fmt.Println("MySQL Connected")
	mongo_credential := options.Credential{
		Username: ua.DB_Info.Mongo_Username,
		Password: ua.DB_Info.Mongo_Password,
	}
	clientOptions := options.Client().ApplyURI("mongodb://" + ua.DB_Info.Mongo_IP + ua.DB_Info.Mongo_port).SetAuth(mongo_credential)
	ua.user_mongo_client, err = mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		log.Fatal(err.Error())
	}

	// Choose the database and collection
	dbName := "skAI_MONGO"
	collectionName := "user_stock_favorites"
	ua.user_favorite_collection = ua.user_mongo_client.Database(dbName).Collection(collectionName)

	indexModel := mongo.IndexModel{
		Keys: bson.D{{Key: "userid", Value: 1}},
	}
	name, err := ua.user_favorite_collection.Indexes().CreateOne(context.TODO(), indexModel)
	if err != nil {
		panic(err)
	}
	fmt.Println("Name of Index Created: " + name)
	fmt.Println("Databases Connection Complete")

}

func (ua *UserAccountModule) LoginHandler(c *gin.Context) {
	login_content := &User{}

	if err := c.ShouldBind(&login_content); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	userinfo, err := ua.findUser(login_content.Username)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Incorrect Username or Password"})
		return
	}

	// Check the credentials (this is a simplified example)
	if login_content.Username != userinfo.Username || login_content.Password != userinfo.Password {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	user_favorate, err2 := ua.getUserFavoriteStocks(userinfo.UserID)
	if err2 != nil {
		panic(err2.Error())
	}

	c.JSON(http.StatusOK, gin.H{
		"success":         true,
		"profile_picture": userinfo.ProfilePicture,
		"favorite_list":   user_favorate,
	})
}

func (ua *UserAccountModule) UserCreateHandler(c *gin.Context) {

	usercreate_content := &User{}

	if err := c.ShouldBind(&usercreate_content); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"create_result": false,
			"error":         "Bad User Input",
		})
		return
	}

	_, err := ua.findUser(usercreate_content.Username)
	if err == nil {
		fmt.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{
			"create_result": false,
			"error":         "Existing Username",
		})
		return
	}

	ua.user_account_DB.Create(usercreate_content)

	userID := usercreate_content.UserID
	stocksToAdd := []string{}
	err2 := ua.addStocksToFavorite(userID, stocksToAdd)
	if err2 != nil {
		log.Fatal(err2.Error())
	}

	c.JSON(http.StatusOK, gin.H{"create_result": true})
}

// Not belong to user_account_module

func (ua *UserAccountModule) findUser(username string) (*User, error) {
	user := new(User)
	err := ua.user_account_DB.First(&user, "username = ?", username).Error
	return user, err
}

func (ua *UserAccountModule) AddFavoriteHandler(c *gin.Context) {
	post_content := struct {
		Username string   `json:"username"`
		Stocknum []string `json:"stocknum"`
	}{}

	if err := c.ShouldBind(&post_content); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	userinfo, err := ua.findUser(post_content.Username)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Incorrect Username or Password"})
		return
	}

	err2 := ua.addStocksToFavorite(userinfo.UserID, post_content.Stocknum)
	if err2 == nil {
		c.JSON(http.StatusOK, gin.H{"add_result": true})
	}

}

func (ua *UserAccountModule) DelFavoriteHandler(c *gin.Context) {
	post_content := struct {
		Username string   `json:"username"`
		Stocknum []string `json:"stocknum"`
	}{}

	if err := c.ShouldBind(&post_content); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	userinfo, err := ua.findUser(post_content.Username)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Incorrect Username or Password"})
		return
	}

	err2 := ua.deleteStocksFromFavorite(userinfo.UserID, post_content.Stocknum)
	if err2 == nil {
		c.JSON(http.StatusOK, gin.H{"delete_result": true})
	}

}

func (ua *UserAccountModule) GetFavoriteHandler(c *gin.Context) {
	login_content := &User{}

	if err := c.ShouldBind(&login_content); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	userinfo, err := ua.findUser(login_content.Username)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Incorrect Username or Password"})
		return
	}

	fav_list, err2 := ua.getUserFavoriteStocks(userinfo.UserID)
	if err2 == nil {
		c.JSON(http.StatusOK, gin.H{"get_favorite": fav_list})
	}
}

func (ua *UserAccountModule) addStocksToFavorite(userID int, stocks []string) error {
	ctx := context.Background()

	var result StockFavorite
	err := ua.user_favorite_collection.FindOne(ctx, bson.M{"userid": userID}).Decode(&result)
	if err == nil {
		//print("Not New User\n")
		// Update operation to add stocks to favorite list
		update := bson.M{"$addToSet": bson.M{"favorite_stock": bson.M{"$each": stocks}}}
		_, err := ua.user_favorite_collection.UpdateOne(ctx, bson.M{"userid": userID}, update)
		if err != nil {
			return err
		}
	} else {
		//print("New User\n")
		inserttodb := bson.M{"userid": userID, "favorite_stock": stocks}
		_, err := ua.user_favorite_collection.InsertOne(ctx, inserttodb)
		if err != nil {
			return err
		}
	}

	fmt.Println("Stocks added to favorite successfully")
	return nil
}

func (ua *UserAccountModule) deleteStocksFromFavorite(userID int, stocks []string) error {
	ctx := context.Background()

	// Update operation to remove stocks from favorite list
	update := bson.M{"$pull": bson.M{"favorite_stock": bson.M{"$in": stocks}}}
	_, err := ua.user_favorite_collection.UpdateOne(ctx, bson.M{"userid": userID}, update)
	if err != nil {
		return err
	}

	fmt.Println("Stocks deleted from favorite successfully")
	return nil
}

func (ua *UserAccountModule) getUserFavoriteStocks(userID int) ([]string, error) {
	ctx := context.Background()

	// Query user's favorite stocks
	var result StockFavorite
	err := ua.user_favorite_collection.FindOne(ctx, bson.M{"userid": userID}).Decode(&result)
	if err != nil {
		return nil, err
	}

	return result.FavoriteStock, nil
}
