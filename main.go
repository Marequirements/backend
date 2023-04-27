package main

import (
	"back-end/token"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func main() {
	router := chi.NewRouter()

	ts := token.GetTokenStorageInstance()
	token := ts.GenerateToken()
	log.Println("generetad token" + token)
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
