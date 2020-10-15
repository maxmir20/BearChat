package api

import (
	"net/http"
	"github.com/go-gremlin/gremlin"
	"github.com/gorilla/mux"
	"log"
	"encoding/json"
)

func RegisterRoutes(router *mux.Router) error {
	router.HandleFunc("/api/friends/{uuid}", areFriends).Methods(// YOUR CODE HERE)
	router.HandleFunc("/api/friends/{uuid}", addFriend).Methods(// YOUR CODE HERE)
	router.HandleFunc("/api/friends/{uuid}", deleteFriend).Methods(// YOUR CODE HERE)
	router.HandleFunc("/api/friends/{uuid}/mutual", mutualFriends).Methods(// YOUR CODE HERE)
	router.HandleFunc("/api/friends", addUser).Methods(// YOUR CODE HERE)

	return nil
}

func getUUID (w http.ResponseWriter, r *http.Request) (uuid string) {
	cookie, err := r.Cookie("YOUR CODE HERE")
	// Check and log the error
	// YOUR CODE HERE


	//validate the cookie
	claims, err := ValidateToken(cookie.Value)
	// Check and log the error
	// YOUR CODE HERE

	return claims["UserID"].(string)
}

func addUser (w http.ResponseWriter, r *http.Request) {
	uuid := getUUID(w, r)
	_, err := DB.Exec(gremlin.Query(`YOUR CODE HERE`).Bindings(gremlin.Bind{"userID": uuid}))
	
	// Check and log the error
	// YOUR CODE HERE
}

func areFriends(w http.ResponseWriter, r *http.Request) {
	otherUUID := mux.Vars(r)["uuid"]
	uuid := getUUID(w, r)
	isFriend, err := DB.Exec(gremlin.Query(`// YOUR CODE HERE`).Bindings(gremlin.Bind{"userID": uuid, "otherUUID": otherUUID}))
	
	// Check and log the error
	// YOUR CODE HERE
	
	json.NewEncoder(w).Encode(string(isFriend))
}

func addFriend(w http.ResponseWriter, r *http.Request) {
	otherUUID := mux.Vars(r)["uuid"]
	uuid := getUUID(w, r)
	_, err := DB.Exec(gremlin.Query(`YOUR CODE HERE`).Bindings(gremlin.Bind{"userID": uuid, "otherUUID": otherUUID}))
	
	// Check and log the error
	// YOUR CODE HERE
}

func deleteFriend(w http.ResponseWriter, r *http.Request) {
	otherUUID := mux.Vars(r)["uuid"]
	uuid := getUUID(w, r)
	_, err := DB.Exec(gremlin.Query(`YOUR CODE HERE`).Bindings(gremlin.Bind{"userID": uuid, "otherUUID": otherUUID}))
	
	// Check and log the error
	// YOUR CODE HERE
}

func mutualFriends(w http.ResponseWriter, r *http.Request) {
	otherUUID := mux.Vars(r)["uuid"]
	uuid := getUUID(w, r)
	isFriend, err := DB.Exec(gremlin.Query(`YOUR CODE HERE`).Bindings(gremlin.Bind{"userID": uuid, "otherUUID": otherUUID}))

	// Check and log the error
	// YOUR CODE HERE

	json.NewEncoder(w).Encode(isFriend)
}
