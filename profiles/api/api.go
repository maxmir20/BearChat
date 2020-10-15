package api

import (
	"log"
	"net/http"
	"encoding/json"
	"errors"
	"github.com/gorilla/mux"
)

func RegisterRoutes(router *mux.Router) error {
	router.HandleFunc("/api/profile/{uuid}", getProfile).Methods(http.MethodGet)
	router.HandleFunc("/api/profile/{uuid}", setProfile).Methods(http.MethodPut)

	return nil
}

func getProfile(w http.ResponseWriter, r *http.Request) {
  uuid := mux.Vars(r)["uuid"]
  //check auth
	//fetch cookie
	cookie, err := r.Cookie("access_token")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Print(err.Error())
	}
	//validate the cookie
	claims, err := ValidateToken(cookie.Value)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		log.Print(err.Error())
	}
	log.Println(claims)
	auth := (claims["UserID"] == uuid)
  //fetch public vs private depending on if user is accessing own profile
	var (
		first string
		last string
		email string
		userid string
	)
	var profile Profile
	if !auth {
		// Obtain the first name and last name of the user with the given UUID
		err := DB.QueryRow("YOUR CODE HERE", uuid).Scan(/*YOUR CODE HERE*/, /*YOUR CODE HERE*/)
		// Check and log error
		// YOUR CODE HERE
		profile = Profile{first, last, "", ""}
	} else {
		//Get all the columns for the user with the given uuid
		err := DB.QueryRow("YOUR CODE HERE", /*YOUR CODE HERE*/).Scan/*YOUR CODE HERE*/, /*YOUR CODE HERE*/, /*YOUR CODE HERE*/, /*YOUR CODE HERE*/)
		// Check and log error
		// YOUR CODE HERE
		profile = Profile{first, last, email, userid}
	}

	//to add later - more data if friends

	//encode fetched data as json and serve to client
	json.NewEncoder(w).Encode(profile)
	return
}

func setProfile(w http.ResponseWriter, r *http.Request) {
	uuid := mux.Vars(r)["uuid"]
	//check auth - should also check if profile exists cause token is invalidated when profile deleted
	//fetch cookie
	cookie, err := r.Cookie("access_token")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Print(err.Error())
	}
	//validate the cookie
	claims, err := ValidateToken(cookie.Value)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		log.Print(err.Error())
	}
	log.Println(claims)
	auth := (claims["UserID"] == uuid)

	//Checks if the profile you're trying to edit belongs to you
	if !auth {
		http.Error(w, errors.New("you are not authorized to edit this profile").Error(), http.StatusUnauthorized)
		log.Print(err.Error())
		return
	}

	//store new profile data if auth correct
	profile := Profile{}
	err = json.NewDecoder(r.Body).Decode(&profile)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Print(err.Error())
		return
	}

	// REPLACE the existing profile data with the new data
	// You did something very similar in the SQL homework!
	err = DB.Exec("YOUR CODE HERE", /*YOUR CODE HERE*/, /*YOUR CODE HERE*/, /*YOUR CODE HERE*/, /*YOUR CODE HERE*/)
	
	// Check and log errors
	// YOUR CODE HERE

	return
}
