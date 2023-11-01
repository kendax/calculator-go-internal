package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/kendax/calculator_go_internal/controllers"
)

//Define a function to handle routing for the application
func SetupRoutes() *gin.Engine {
	//Define a gin engine instance that will be used to manage the application's routes
	r := gin.Default()

	//Configure the gin engine to load HTML templates
	r.LoadHTMLGlob("templates/*/*.html")

	//Configure a static file server to serve files located in the "assets" directory
	r.Static("/assets", "./assets")

	//Define a route for handling HTTP GET requests to the root URL ("/")
	r.GET("/", controllers.Display)

	//Define a route for handling HTTP POST requests to the "/postinput" URL
	r.POST("/postinput", controllers.InputSave)

	return r
}