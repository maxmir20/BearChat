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
	router.HandleFunc("/api/profile/{uuid}", updateProfile).Methods(http.MethodPut)

	return nil
}

func getUUID (w http.ResponseWriter, r *http.Request) (uuid string) {
	cookie, err := r.Cookie("access_token")
	if err != nil {
		http.Error(w, errors.New("error obtaining cookie: " + err.Error()).Error(), http.StatusBadRequest)
		log.Print(err.Error())
	}
	//validate the cookie
	claims, err := ValidateToken(cookie.Value)
	if err != nil {
		http.Error(w, errors.New("error validating token: " + err.Error()).Error(), http.StatusUnauthorized)
		log.Print(err.Error())
	}
	log.Println(claims)

	return claims["UserID"].(string)
}

func getProfile(w http.ResponseWriter, r *http.Request) {

	// Obtain the uuid from the url path and store it in a `uuid` variable
	// Hint: mux.Vars()
	uuid := mux.Vars(r)["uuid"]


	// Initialize a new Profile variable
	prof := Profile{}

	// Obtain all the information associated with the requested uuid
	// Scan the information into the profile struct's variables
	// Remember to pass in the address!
	err := DB.QueryRow("SELECT * FROM users WHERE uuid = ?", uuid).Scan(&prof.Firstname, &prof.Lastname, &prof.Email, &prof.UUID)
	
	/*  Check for errors with querying the database
		Return an Internal Server Error if such an error occurs
	*/
	if err != nil{
		http.Error(w, errors.New("error validating token: " + err.Error()).Error(), http.StatusInternalServerError)
	}

  	//encode fetched data as json and serve to client
	json.NewEncoder(w).Encode(prof)
	return
}

func updateProfile(w http.ResponseWriter, r *http.Request) {
	
	// Obtain the requested uuid from the url path and store it in a `uuid` variable
	uuid := mux.Vars(r)["uuid"]

	// Obtain the userID from the cookie
	userID := getUUID(w, r)

	// If the two ID's don't match, return a StatusUnauthorized
	if userID != uuid {
		http.Error(w, errors.New("uuid does not match").Error(), http.StatusUnauthorized)
	}

	// Decode the Request Body's JSON data into a profile variable
	updated_profile := Profile{}
	err := json.NewDecoder(r.Body).Decode(&updated_profile)

	// Return an InternalServerError if there is an error decoding the request body
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}


	// Insert the profile data into the users table
	// Check db-server/initdb.sql for the scheme
	// Make sure to use REPLACE INTO (as covered in the SQL homework)
	_, err = DB.Exec("REPLACE INTO users (firstName, lastName, email, uuid) VALUES(?,?,?,?)", updated_profile.Firstname, updated_profile.Lastname, updated_profile.Email, updated_profile.UUID)

	
	// Return an internal server error if any errors occur when querying the database.
	if err != nil {
		http.Error(w, errors.New("error retrieving").Error(), http.StatusInternalServerError)
	}

	return
}
