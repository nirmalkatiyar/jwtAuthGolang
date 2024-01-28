package controller

import (
	"GoProject/database"
	"GoProject/helper"
	"GoProject/model"
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
	"log"
	"net/http"
	"strconv"
	"time"
)

var (
	userCollection = database.OpenCollection(database.DBInstance(), "user")
	validate       = validator.New()
)

func HashPassword(password string) string {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		log.Panic(err)
		return ""
	}
	return string(hashedPassword)
}

func VerifyPassword(userPassword, actualPassword string) (isPasswordMatched bool, message string) {
	err := bcrypt.CompareHashAndPassword([]byte(actualPassword), []byte(userPassword))
	if err != nil {
		return false, " email or password is incorrect kindly check :)"
	}
	return true, ""
}

func SignUp() gin.HandlerFunc {
	return func(c *gin.Context) {
		var user model.User
		if err := c.BindJSON(&user); err != nil {
			fmt.Println("BindJson Error")
			c.JSON(http.StatusInternalServerError, gin.H{" error ": err})
			return
		}
		validationError := validate.Struct(user)
		if validationError != nil {
			fmt.Println("validate struct  Error")
			c.JSON(http.StatusInternalServerError, gin.H{" error ": validationError})
			return
		}
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		count, err := userCollection.CountDocuments(ctx, bson.M{"email": user.Email})
		defer cancel()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{" something went wrong during email addition ": err})
			return
		}
		password := HashPassword(*user.Password)
		user.Password = &password
		count, err = userCollection.CountDocuments(ctx, bson.M{"phone": user.Phone})
		defer cancel()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{" something went wrong during phone addition ": err})
			return
		}
		if count > 0 {
			c.JSON(http.StatusOK, gin.H{" message: ": " This email or phone no already exists please try with different one "})
			return
		}
		user.CreatedAt = time.Now().UTC()
		user.UpdatedAt = time.Now().UTC()
		user.Id = primitive.NewObjectID()
		user.UserId = user.Id.Hex()
		token, refreshToken, err := helper.GenerateAllTokens(*user.Email, *user.FirstName, *user.LastName, *user.UserType, user.UserId)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error message ": "token is not generated ",
			})
			return
		}
		user.Token = &token
		user.RefreshToken = &refreshToken
		result, insertError := userCollection.InsertOne(ctx, user)
		log.Println("inserted one :1 ")
		if insertError != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error message ": "unable to add user",
			})
			return
		}
		defer cancel()
		c.JSON(http.StatusOK, result)
		return

	}
}

func Login() gin.HandlerFunc {
	return func(c *gin.Context) {
		var user, foundUser model.User
		if err := c.BindJSON(&user); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error message ": " something went wrong in Login"})
			return
		}
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		err := userCollection.FindOne(ctx, bson.M{"email": user.Email}).Decode(&foundUser)
		defer cancel()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error message ": " something went wrong in login Decode"})
			return
		}
		isPasswordMatched, errorMessage := VerifyPassword(*user.Password, *foundUser.Password)
		if isPasswordMatched != true {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error message ": errorMessage,
			})
			return
		}
		defer cancel()
		if foundUser.Email == nil {
			c.JSON(http.StatusOK, gin.H{
				"error message: ": "User not found for given emailId",
			})
			return
		}
		tokens, refToken, err := helper.GenerateAllTokens(*foundUser.Email, *foundUser.FirstName, *foundUser.LastName, *foundUser.UserType, foundUser.UserId)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error message :": "unable to generate token",
			})
			return
		}
		helper.UpdateAllToken(tokens, refToken, foundUser.UserId)
		err = userCollection.FindOne(ctx, bson.M{"user_id": foundUser.UserId}).Decode(&foundUser)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error message :": "No user found :)",
			})
			return
		}
		c.JSON(http.StatusOK, foundUser)
		return
	}
}

func GetUsers() gin.HandlerFunc {
	return func(c *gin.Context) {
		err := helper.CheckUserType(c, "ADMIN")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error message ": "You don't have access to the resource",
			})
			return
		}
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()
		recordPerPage, err := strconv.Atoi(c.Query("recordPerPage"))
		// set the default pages
		if err != nil || recordPerPage < 1 {
			recordPerPage = 10
		}

		page, err1 := strconv.Atoi(c.Query("page"))
		if err1 != nil || page < 1 {
			page = 1
		}

		startIndex := (page - 1) * recordPerPage
		startIndex, err = strconv.Atoi(c.Query("startIndex"))
		if err != nil {
			return
		}

		matchStage := bson.D{{"$match", bson.D{{}}}}
		groupStage := bson.D{{"$group", bson.D{
			{"_id", bson.D{{"_id", "null"}}},
			{"total_sum", bson.D{{"$sum", 1}}},
			{"data", bson.D{{"$push", "$$ROOT"}}},
		}}}
		projectStage := bson.D{
			{"$project", bson.D{
				{"_id", 0},
				{"total_count", 1},
				{"user_items", bson.D{{"$slice", []interface{}{"$data", startIndex}}}},
			}},
		}
		result, err := userCollection.Aggregate(ctx, mongo.Pipeline{
			matchStage, groupStage, projectStage,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error message ": "something went wrong while aggregating GetUsers",
			})
		}
		defer cancel()
		var allUsers []bson.M
		if err := result.All(ctx, &allUsers); err != nil {
			log.Fatal(err)
		}
		c.JSON(http.StatusOK, allUsers[0])
		return
	}
}

func GetUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		userId := c.Param("user_id")
		if err := helper.MatchUserTypeToUserId(c, userId); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{" error ": err.Error()})
			return
		}
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()
		var user model.User
		err := userCollection.FindOne(ctx, bson.M{"user_id": userId}).Decode(&user)
		if err != nil {
			fmt.Println("Unable to decode : ", err)
			c.JSON(http.StatusInternalServerError, gin.H{" error ": err})
			return
		}
		c.JSON(http.StatusOK, user)
	}
}
