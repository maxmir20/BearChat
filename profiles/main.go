package main

import (
	"log"
	_ "log"
	"net/http"
	_ "net/http"

	"github.com/BearCloud/fa20-project-dev/backend/profile/api"
	"github.com/gorilla/mux"
)

func main() {
	//init db
	DB := api.InitDB()
	defer DB.Close()

	//ping the database to make sure it's up
	err := DB.Ping()
	if err != nil {
		panic(err.Error())
	}
	//Create a new mux for routing api calls
	router := mux.NewRouter()
	router.Use(CORS)

	err = api.RegisterRoutes(router)
	if err != nil {
		log.Fatal("Error registering API endpoints")
	}

	http.ListenAndServe(":80", router)
}

func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// Set headers
		w.Header().Set("Access-Control-Allow-Headers:", "Content-Type")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, PUT, OPTIONS")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Next
		next.ServeHTTP(w, r)
		return
	})
}