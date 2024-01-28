// It is the code flow entrance

package main

import (
	"GoProject/routes"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"net/http"
	"os"
)

// code starts here
func main() {
	// loading the .env file
	err := godotenv.Load(".env")
	if err != nil {
		fmt.Println("can't load the .env file please retry")
		return
	}
	// getting the host port number
	port := os.Getenv("PORT")
	if port == "" {
		port = "8010"
		fmt.Println("setting it to default port ", port)
	}
	router := gin.New()
	router.Use(gin.Logger())

	routes.AuthRoutes(router)
	routes.UserRoutes(router)
	// get request
	router.GET("/api-1", func(context *gin.Context) {
		context.JSON(200, gin.H{
			"code":    http.StatusOK,
			"Message": "Connected successfully to api-1",
		})
	})

	// get request
	router.GET("/api-2", func(context *gin.Context) {
		context.JSON(200, gin.H{
			"code":    http.StatusOK,
			"Message": "Connected successfully to api-2",
		})
	})

	err = router.Run(":" + port)
	if err != nil {
		fmt.Println("Some error occurred ", err)
		return
	}
}
