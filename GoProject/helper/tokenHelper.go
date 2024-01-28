package helper

import (
	"GoProject/database"
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"fmt"
	jwt "github.com/dgrijalva/jwt-go"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"time"
)

type SignedDetails struct {
	Email     string
	FirstName string
	LastName  string
	UserType  string
	UserId    string
	jwt.StandardClaims
}

var userCollection = database.OpenCollection(database.DBInstance(), "user")

// GenerateRandomSecretKey var secretKey = os.Getenv("SECRET_KEY")
// GenerateRandomSecretKey ... generate random secret key
func GenerateRandomSecretKey() (secretKey *ecdsa.PrivateKey, err error) {
	// Generate ECDSA private key
	secretKey, err = ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		fmt.Println("Failed to generate ECDSA private key:", err)
		return nil, err
	}
	return secretKey, nil
}

func GenerateAllTokens(email, firstName, lastName, userType, userId string) (singedToken, signedRefreshToken string, err error) {
	claims := &SignedDetails{
		Email:     email,
		FirstName: firstName,
		LastName:  lastName,
		UserType:  userType,
		UserId:    userId,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Local().Add(time.Hour * time.Duration(1)).Unix(),
		},
	}
	refreshClaim := &SignedDetails{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Local().Add(time.Hour * time.Duration(1)).Unix(),
		},
	}
	secretKey, err := GenerateRandomSecretKey()
	if err != nil {
		fmt.Println("GenerateAllTokens GenerateRandomSecretKey error here ")
		return "", "", err
	}
	signedTokenString, err := jwt.NewWithClaims(jwt.SigningMethodES256, claims).SignedString(secretKey)
	if err != nil {
		fmt.Println("error is here ", err)
		return "", "", err
	}
	refreshTokenString, err := jwt.NewWithClaims(jwt.SigningMethodES256, refreshClaim).SignedString(secretKey)
	if err != nil {
		fmt.Println("error is here2 ", err)
		return "", "", err
	}
	return signedTokenString, refreshTokenString, nil
}
func UpdateAllToken(signedToken, signedRefToken, uid string) {
	var updateObj primitive.D
	updateObj = append(updateObj, bson.E{
		"token", signedToken,
	})
	updateObj = append(updateObj, bson.E{
		"refresh_token", signedRefToken,
	})
	updatedAt := time.Now().UTC()
	updateObj = append(updateObj, bson.E{
		"updated_at", updatedAt,
	})

	upsert := true
	filter := bson.M{"user_id": uid}
	opt := options.UpdateOptions{
		Upsert: &upsert,
	}
	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()
	_, err := userCollection.UpdateOne(
		ctx,
		filter,
		bson.D{
			{"$set", updateObj},
		},
		&opt)
	if err != nil {
		log.Panic("Chan bjbkhhmb ", err)
		return
	}
}

func ValidateToken(signedToken string) (claim *SignedDetails, msg string) {
	secretKey, err := GenerateRandomSecretKey()
	if err != nil {
		fmt.Println(" ValidateToken GenerateRandomSecretKey error here ")
		return
	}
	token, err := jwt.ParseWithClaims(
		signedToken,
		&SignedDetails{},
		func(token *jwt.Token) (interface{}, error) {
			return secretKey, nil
		})
	if err != nil {
		msg = err.Error()
		return
	}
	claims, ok := token.Claims.(*SignedDetails)
	if !ok {
		msg = fmt.Sprint("invalid token")
		return
	}
	if claims.ExpiresAt < time.Now().Local().Unix() {
		msg = fmt.Sprintf("token is expired !! ")
	}
	return claims, msg
}
