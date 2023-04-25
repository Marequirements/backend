package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func main() {
	router := chi.NewRouter()

	err := http.ListenAndServe(":3000", router)
	if err != nil {
		log.Println(err)
	}

	router.Get("/login", func(writer http.ResponseWriter, request *http.Request) {
		_, err := writer.Write([]byte("Hello World"))
		if err != nil {
			log.Println(err)
		}
	})

}
