package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/sendgrid/sendgrid-go"
	"golang.org/x/crypto/bcrypt"
)

const (
	verifyTokenSize = 6
	resetTokenSize  = 6
)

// RegisterRoutes initializes the api endpoints and maps the requests to specific functions
func RegisterRoutes(router *mux.Router) error {
	router.HandleFunc("/api/auth/signup", signup).Methods(http.MethodPost, http.MethodOptions)
	router.HandleFunc("/api/auth/signin", signin).Methods(http.MethodPost, http.MethodOptions)
	router.HandleFunc("/api/auth/logout", logout).Methods(http.MethodPost, http.MethodOptions)
	router.HandleFunc("/api/auth/verify", verify).Methods(http.MethodPost, http.MethodOptions)
	router.HandleFunc("/api/auth/sendreset", sendReset).Methods(http.MethodPost, http.MethodOptions)
	router.HandleFunc("/api/auth/resetpw", resetPassword).Methods(http.MethodPost, http.MethodOptions)

	// Load sendgrid credentials
	err := godotenv.Load()
	if err != nil {
		return err
	}

	sendgridKey = os.Getenv("SENDGRID_KEY")
	sendgridClient = sendgrid.NewSendClient(sendgridKey)
	return nil
}

func signup(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
	w.Header().Set("Access-Control-Allow-Headers", "content-type")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Credentials", "true")

	if (*r).Method == "OPTIONS" {
		return
	}

	//Obtain the credentials from the request body
	credential := Credentials{}
	err := json.NewDecoder(r.Body).Decode(&credential)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	if credential.Username == "" || credential.Password == "" || credential.Email == "" {
		w.WriteHeader(400)
		return
	}


	//Check if the username already exists
	var exists bool
	err = DB.QueryRow("SELECT EXISTS (SELECT username FROM users WHERE username=?)", credential.Username).Scan(&exists)
	
	//Check for error
	if err != nil {
		http.Error(w, errors.New("error checking if username exists").Error(), http.StatusInternalServerError)
		log.Print(err.Error())
		return
	}

	//Check boolean returned from query
	if exists == true {
		http.Error(w, errors.New("this username is taken").Error(), http.StatusConflict)
		return
	}

	//Check if the email already exists
	err = DB.QueryRow("SELECT EXISTS (SELECT username FROM users WHERE email=?)", credential.Email).Scan(&exists)
	
	//Check for error
	if err != nil {
		http.Error(w, errors.New("error checking if email exists").Error(), http.StatusInternalServerError)
		log.Print(err.Error())
		return
	}

	//Check boolean returned from query
	//max notes: should be != ?
	if exists == true {
		http.Error(w, errors.New("this email is taken").Error(), http.StatusConflict)
		return
	}

	//Hash the password using bcrypt and store the hashed password in a variable
	//max notes: do I need my compared hash to have the same cost?
	hashed_password, err := bcrypt.GenerateFromPassword([]byte(credential.Password), 14)

	//Check for errors during hashing process
	if err != nil {
		http.Error(w, errors.New("error during hashing process").Error(), http.StatusInternalServerError)
		log.Print(err.Error())
		return
	}

	//Create a new user UUID, convert it to string, and store it within a variable
	new_uuid = string(uuid.New())

	//Create new verification token with the default token size (look at GetRandomBase62 and our constants)
	verify_token := GetRandomBase62(verifyTokenSize)

	//Store credentials in database
	_, err = DB.Query("INSERT INTO users (username, email, hashedPassword, verified, resetToken, verifiedToken, userID) VALUES (?,?,?, True, NULL, ?, ?)", 
					   credential.Username, credential.Email, hashed_password, verify_token, new_uuid)
	
	//Check for errors in storing the credentials
	if err != nil {
		http.Error(w, errors.New("error in storing the credentials").Error(), http.StatusInternalServerError)
		log.Print(err.Error())
		return
	}

	//Generate an access token, expiry dates are in Unix time
	accessExpiresAt := time.Now().Add(time.Minute * 15).Unix() //set for 15 minutes
	var accessToken string
	accessToken, err = setClaims(AuthClaims{
		UserID: credential.Username,
		StandardClaims: jwt.StandardClaims{
			Subject:   "access",
			ExpiresAt: accessExpiresAt,
			Issuer:    defaultJWTIssuer,
			IssuedAt:  time.Now().Unix(),
		},
	})
	
	//Check for error in generating an access token
	if err != nil {
		http.Error(w, errors.New("error in generating an access token").Error(), http.StatusInternalServerError)
		log.Print(err.Error())
		return
	}


	//Set the cookie, name it "access_token"
	http.SetCookie(w, &http.Cookie{
		Name:    "access_token",
		Value:   accessToken,
		Expires: accessExpiresAt,
		// Leave these next three values commented for now
		// Secure: true,
		// HttpOnly: true,
		// SameSite: http.SameSiteNoneMode,
		Path: "/",
	})

	//Generate refresh token
	var refreshExpiresAt = time.Now().Add(DefaultRefreshJWTExpiry)
	var refreshToken string
	refreshToken, err = setClaims(AuthClaims{
		UserID: userID,
		StandardClaims: jwt.StandardClaims{
			Subject:   "refresh",
			ExpiresAt: refreshExpiresAt,
			Issuer:    defaultJWTIssuer,
			IssuedAt: time.Now().Unix(),
		},
	})

	if err != nil {
		http.Error(w, errors.New("error creating refreshToken").Error(), http.StatusInternalServerError)
		log.Print(err.Error())
		return
	}

	//set the refresh token ("refresh_token") as a cookie
	http.SetCookie(w, &http.Cookie{
		Name:    "refresh_token",
		Value:   refreshToken,
		Expires: refreshExpiresAt,
		Path: "/",
	})

	// Send verification email
	err = SendEmail(credential.Email, "Email Verification", "user-signup.html", map[string]interface{}{"Token": verificationToken})
	if err != nil {
		http.Error(w, errors.New("error sending verification email").Error(), http.StatusInternalServerError)
		log.Print(err.Error())
		return
	}

	w.WriteHeader(200)
	return
}

func signin(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Credentials", "true")

	if (*r).Method == "OPTIONS" {
		return
	}

	//Store the credentials in a instance of Credentials + //Check for errors in storing credntials
	credential := Credentials{}
	err := json.NewDecoder(request.Body).Decode(&credential)
	if err != nil {
		http.Error(response, err.Error(), http.StatusBadRequest)
	}

	if credential.Username == "" || credential.Password == "" || credential.Email == ""{
		response.WriteHeader(400)
		return
	}



	//Get the hashedPassword and userId of the user
	var hashedPassword, userID string
	err = DB.QueryRow("SELECT hashedPassword, userId FROM users WHERE username = ?, email = ?", credential.username, credential.email).Scan(&hashedPassword, &userID)
	// process errors associated with emails
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, errors.New("this email is not associated with an account").Error(), http.StatusNotFound)
		} else {
			http.Error(w, errors.New("error retrieving information with this email").Error(), http.StatusInternalServerError)
			log.Print(err.Error())
		}
		return
	}

	// Check if hashed password matches the one corresponding to the email + Check error in comparing hashed passwords

	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(credential.Password))
	if err != nil {
		http.Error(w, errors.New("incorrect password").Error(), http.StatusUnauthorized)
		log.Print(err.Error())
	}

	//Generate an access token  and set it as a cookie (Look at signup and feel free to copy paste!)
	//Generate an access token, expiry dates are in Unix time
	accessExpiresAt := time.Now().Add(time.Minute * 15).Unix() //set for 15 minutes
	var accessToken string
	accessToken, err = setClaims(AuthClaims{
		UserID: credential.Username,
		StandardClaims: jwt.StandardClaims{
			Subject:   "access",
			ExpiresAt: accessExpiresAt,
			Issuer:    defaultJWTIssuer,
			IssuedAt:  time.Now().Unix(),
		},
	})
	
	//Check for error in generating an access token
	if err != nil {
		http.Error(w, errors.New("error in generating an access token").Error(), http.StatusInternalServerError)
		log.Print(err.Error())
		return
	}

	//Set the cookie, name it "access_token"
	http.SetCookie(w, &http.Cookie{
		Name:    "access_token",
		Value:   accessToken,
		Expires: accessExpiresAt,
		// Leave these next three values commented for now
		// Secure: true,
		// HttpOnly: true,
		// SameSite: http.SameSiteNoneMode,
		Path: "/",
	})

	//Generate a refresh token and set it as a cookie (Look at signup and feel free to copy paste!)
	//Generate refresh token
	var refreshExpiresAt = time.Now().Add(DefaultRefreshJWTExpiry)
	var refreshToken string 
	refreshToken, err = setClaims(AuthClaims{
		UserID: userID,
		StandardClaims: jwt.StandardClaims{
			Subject:   "refresh",
			ExpiresAt: refreshExpiresAt,
			Issuer:    defaultJWTIssuer,
			IssuedAt: time.Now().Unix(),
		},
	})

	if err != nil {
		http.Error(w, errors.New("error creating refreshToken").Error(), http.StatusInternalServerError)
		log.Print(err.Error())
		return
	}

	//set the refresh token ("refresh_token") as a cookie
	http.SetCookie(w, &http.Cookie{
		Name:    "refresh_token",
		Value:   refreshToken,
		Expires: refreshExpiresAt,
		Path: "/",
	})
	//max notes: add header?
	w.WriteHeader(200)
	return
}

func logout(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")

	if (*r).Method == "OPTIONS" {
		return
	}

	// logging out causes expiration time of cookie to be set to now

	//Set the access_token and refresh_token to have an empty value and set their expiration date to anytime in the past

	//max notes: debating whether I should set the value at nil or 0 (refer to verify function for comparison)
	var expiresAt = time.Now() - 1
	http.SetCookie(w, &http.Cookie{Name: "access_token", Value: nil, Expires: expiresAt})
	http.SetCookie(w, &http.Cookie{Name: "refresh_token", Value: nil , Expires: expiresAt})
	return
}

func verify(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "localhost:3000")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Credentials", "true")

	if (*r).Method == "OPTIONS" {
		return
	}

	//max notes: will this return the verified token? it's Token in the verification email and token here
	token, ok := r.URL.Query()["token"]
	// check that valid token exists
	if !ok || len(token[0]) < 1 {
		http.Error(w, errors.New("Url Param 'token' is missing").Error(), http.StatusInternalServerError)
		log.Print(errors.New("Url Param 'token' is missing").Error())
		return
	}

	//Obtain the user with the verifiedToken from the query parameter and set their verification status to the integer "1"
	_, err := DB.Exec("UPDATE users SET verified = 1 WHERE verifiedToken = ?", token)

	//Check for errors in executing the previous query
	//max notes: probably haven't covered all use cases yet
	if err != nil {
		http.Error(w, errors.New("Cannot find that verification token").Error(), StatusNotFound)
		log.Print(err.Error())
		return
	}

	return
}


func sendReset(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "localhost:3000")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Credentials", "true")

	if (*r).Method == "OPTIONS" {
		return
	}

	//Get the email from the body (decode into an instance of Credentials)
	credential := Credentials{}
	err := json.NewDecoder(request.Body).Decode(&credential)
	
	//check for errors decoding the object
	if err != nil {
		http.Error(response, err.Error(), http.StatusBadRequest)
	}

	//check for other miscallenous errors that may occur
	//what is considered an invalid input for an email?
	//max notes: are there other invalid inputs?
	if credential.Email == ""{
		response.WriteHeader(400)
		return
	}

	//generate reset token
	token := GetRandomBase62(resetTokenSize)

	//Obtain the user with the specified email and set their resetToken to the token we generated
	_, err = DB.Exec("UPDATE users SET resetToken = ? WHERE email = ?", token, credential.Email)
	// _, err = DB.Query("YOUR CODE HERE", /*YOUR CODE HERE*/, /*YOUR CODE HERE*/)
	
	//Check for errors executing the queries
	//max notes: right error?
	if err != nil {
		http.Error(w, errors.New("error finding user").Error(), http.StatusInternalServerError)
		log.Print(err.Error())
		return
	}

	// Send verification email
	err = SendEmail(credential.Email, "BearChat Password Reset", "password-reset.html", map[string]interface{}{"Token": token})
	if err != nil {
		http.Error(w, errors.New("error sending verification email").Error(), http.StatusInternalServerError)
		log.Print(err.Error())
		return
	}
	return
}

func resetPassword(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "localhost:3000")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Credentials", "true")

	if (*r).Method == "OPTIONS" {
		return
	}
	
	//get token from query params
	token := r.URL.Query().Get("token")

	//get the username, email, and password from the body
	credential = Credentials{}
	err := json.NewDecoder(request.Body).Decode(&credential)


	//Check for errors decoding the body
	if err != nil {
		http.Error(response, err.Error(), http.StatusBadRequest)
	}

	//Check for invalid inputs, return an error if input is invalid
	if credential.Username == "" || credential.Password == "" || credential.Email == "" {
		response.WriteHeader(400)
		return
	}


	email := credential.Email;
	username := credential.Username;
	password := credential.Password
	var exists bool
	//check if the username and token pair exist
	err = DB.QueryRow("SELECT EXISTS (SELECT username, resetToken FROM users WHERE username = ?, resetToken = ?", username, token).Scan(&exists)

	//Check for errors executing the query	
	if err != nil {
		http.Error(w, errors.New("error checking if username exists").Error(), http.StatusInternalServerError)
		log.Print(err.Error())
		return
	}

	//Check exists boolean. Call an error if the username-token pair doesn't exist
	if exists != true {
		http.Error(w, errors.New("this username is taken").Error(), http.StatusConflict)
		return
	}



	//Hash the new password
	hashed_password, err := bcrypt.GenerateFromPassword([]byte(credential.Password), 14)

	//Check for errors in hashing the new password
	if err != nil {
		http.Error(w, errors.New("error during hashing process").Error(), http.StatusInternalServerError)
		log.Print(err.Error())
		return
	}


	//input new password and clear the reset token (set the token equal to empty string)
	_, err = DB.Exec("UPDATE users SET hashedPassword = ?, resetToken = '' WHERE username = ?, email = ?", hashed_password, username, email)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Print(err.Error())
	}

	//put the user in the redis cache to invalidate all current sessions (NOT IN SCOPE FOR PROJECT), leave this comment for future reference

	return
}
