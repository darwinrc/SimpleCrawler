package main

import (
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"server/internal/handler"
	"server/internal/service"
)

func main() {
	router := mux.NewRouter()

	crawlerService := service.NewCrawlerService()
	wsHandler := handler.NewWsHandler(crawlerService)
	wsHandler.Attach(router)

	allowedOrigins := handlers.AllowedOrigins([]string{"*"})
	allowedMethods := handlers.AllowedMethods([]string{"GET", "POST"})
	allowedHeaders := handlers.AllowedHeaders([]string{"Content-Type"})

	if err := http.ListenAndServe(":5000", handlers.CORS(allowedOrigins, allowedMethods, allowedHeaders)(router)); err != nil {
		log.Fatal("ListenAndServe", err)
	}

	log.Println("Parser web crawler server listening on port 5000")
}
