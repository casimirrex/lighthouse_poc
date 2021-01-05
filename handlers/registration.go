package handler

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func HandleRequests() {

	myRouter := mux.NewRouter().StrictSlash(true)

	log.Println("Listening ...")

	myRouter.HandleFunc("/api/v1/metrics/{dataSource}", getData).Methods("GET")
	log.Fatal(http.ListenAndServe(":8084", myRouter))
}
