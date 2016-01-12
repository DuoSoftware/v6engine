package endpoints

import (
	//"duov6.com/duonotifier/messaging"
	//"duov6.com/duonotifier/repositories"
	"fmt"
	"github.com/go-martini/martini"
	"github.com/martini-contrib/cors"
	"net/http"
)

type HTTPService struct {
}

func (h *HTTPService) Start() {
	fmt.Println("DuoNotifier Listening on Port : 7000")
	m := martini.Classic()
	m.Use(cors.Allow(&cors.Options{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},
		AllowHeaders:     []string{"securityToken", "Content-Type"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	m.Post("/:namespace", handleRequest)
	fmt.Println("huehuehue")
	m.Run()
}

func handleRequest(params martini.Params, w http.ResponseWriter, r *http.Request) {
	fmt.Println("FUCKERS")
}
