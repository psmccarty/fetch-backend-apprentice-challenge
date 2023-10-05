package main

import (
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/psmccarty/fetch-backend-apprentice-challenge/handler"
)

const Port = ":8080"

func main() {
	router := httprouter.New()
	rHandler := handler.NewReceiptsHandler()

	rHandler.RegisterRoutes(router)
	log.Println("Listening on localhost" + Port)
	log.Fatal(http.ListenAndServe(Port, router))
}
