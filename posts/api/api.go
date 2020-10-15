package api

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"database/sql"
	"strconv"
)


func RegisterRoutes(router *mux.Router) error {
	router.HandleFunc("/api/posts/{startIndex}", getFeed).Methods(http.MethodGet, http.MethodOptions)
	router.HandleFunc("/api/posts/{uuid}/{startIndex}", getPosts).Methods(http.MethodGet, http.MethodOptions)
	router.HandleFunc("/api/posts/create", createPost).Methods(http.MethodPost, http.MethodOptions)
	router.HandleFunc("/api/posts/delete/{postID}", deletePost).Methods(http.MethodDelete, http.MethodOptions)

	return nil
}

func getUUID (w http.ResponseWriter, r *http.Request) (uuid string) {
	cookie, err := r.Cookie("access_token")
	
	//Check and log error
	// YOUR CODE HERE


	//validate the cookie
	claims, err := ValidateToken(cookie.Value)
	//Check and log error
	// YOUR CODE HERE

	return claims["UserID"].(string)
}

func getPosts(w http.ResponseWriter, r *http.Request) {
	
	uuid := mux.Vars(r)["uuid"]
	startIndex := mux.Vars(r)["startIndex"]
  	//check auth
	isAuthorized := (getUUID(w, r) == uuid)
	// Error and return if theyre user is not authorized


  	//fetch public vs private depending on if user is accessing own profile
	var posts *sql.Rows
	var err error
	posts, err = DB.Query("YOUR CODE HERE", /*YOUR CODE HERE*/, /* YOUR CODE HERE*/)
	
	//Check and log error
	// YOUR CODE HERE


	var (
		content string
		postID string
		userid string
		postTime time.Time
	)
	numPosts := 0

	//Create an empty array of 25 Posts
	postsArray := make([]Post, 25)
	//iterate through the rows
	for i := 0; i < 25 && posts.Next(); i++ {
		// Scan the columns of the array into the specific variables
		// Remember that the columns read from the database in order. Check the database schema!
		// Hint: Do we pass in the variables or their addresses? Check the doc example! https://golang.org/pkg/database/sql/#DB.Query+
		err = posts.Scan(/*YOUR CODE HERE*/, /*YOUR CODE HERE*/, /*YOUR CODE HERE*/, /*YOUR CODE HERE*/)
		//Check and log error
		// YOUR CODE HERE

		// Create a new Post object and initialize it with the values we just scanned
		// YOUR CODE HERE

		// Set the ith index of the postsArray to the Post object we just created
		// YOUR CODE HERE
	
		postsArray[i] = Post{content, postID, userid, postTime}
		numPosts++
	}

	posts.Close()
	err = posts.Err()
	
	// Check and log error
	// YOUR CODE HERE

	//encode fetched data as json and serve to client
	json.NewEncoder(w).Encode(postsArray[:numPosts])
	return
}

func createPost(w http.ResponseWriter, r *http.Request) {
	userID := getUUID(w, r)
	var post Post
	json.NewDecoder(r.Body).Decode(&post)
	postID := uuid.New()
	pst, err := time.LoadLocation("America/Los_Angeles")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	//Insert the post content and id, userID, and current time into the database
	currTime := time.Now().In(pst)
	_, err = DB.Exec("YOUR CODE HERE", /*YOUR CODE HERE*/, /*YOUR CODE HERE*/, /*YOUR CODE HERE*/, /*YOUR CODE HERE*/)
	
	// Check and log error
	// YOUR CODE HERE

	return
}

func deletePost(w http.ResponseWriter, r *http.Request) {
	postID := mux.Vars(r)["postID"]
	//fetch cookie
	uuid := getUUID(w, r)
	log.Println(uuid)
	var exists bool
	// check if post exists
	// USE THE postID
	err := DB.QueryRow("YOUR CODE HERE", postID).Scan(/*YOUR CODE HERE*/)
	
	// Check and log error
	// YOUR CODE HERE

	// Error if the post does not exist
	// YOUR CODE HERE

	var postUUID string
	//Obtain the authorID of the post with the give postID
	err = DB.QueryRow("YOUR CODE HERE", postID).Scan(/*YOUR CODE HERE*/)
	// Check and log error
	// YOUR CODE HERE

	// Error if our given UUID does not match the post's UUID
	// YOUR CODE HERE

	//Delete the row with the given postID
	_, err = DB.Exec("YOUR CODE HERE", /*YOUR CODE HERE*/)

	// Check and log error
	// YOUR CODE HERE
}

func getFeed(w http.ResponseWriter, r *http.Request) {
	//get the start index
	startIndex := mux.Vars(r)["startIndex"]
	//convert to int
	intStartIndex, err := strconv.Atoi(startIndex)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	//fetch cookie
	uuid := getUUID(w, r)

	// Get the next 25 posts where the authorID of each post does not match our uuID
	// The offset is given in startIndex, but you will always be getting the next 25 rows (or less if there's less than 25 total posts)
	// You did something very similar in the SQL homework!  
	posts, err := DB.Query("YOUR CODE HERE", /*YOUR CODE HERE*/, /*YOUR CODE HERE*/)
	// Check and log error
	// YOUR CODE HERE

	var (
		content string
		postID string
		userid string
		postTime time.Time
	)
	numPosts := 0

	//Create an empty array of 25 Posts
	postsArray := make([]Post, 25)
	//iterate through the rows
	for i := 0; i < 25 && posts.Next(); i++ {
		// Scan the columns of the array into the specific variables
		// Remember that the columns read from the database in order. Check the database schema!
		// Hint: Do we pass in the variables or their addresses? Check the doc example! https://golang.org/pkg/database/sql/#DB.Query+

		err = posts.Scan(/*YOUR CODE HERE*/, /*YOUR CODE HERE*/, /*YOUR CODE HERE*/, /*YOUR CODE HERE*/)
		//Check and log error
		// YOUR CODE HERE

		// Create a new Post object and initialize it with the values we just scanned
		// YOUR CODE HERE

		// Set the ith index of the postsArray to the Post object we just created
		// YOUR CODE HERE
	
		postsArray[i] = Post{content, postID, userid, postTime}
		numPosts++
	}

	posts.Close()
	err = posts.Err()
	
	// Check and log error
	// YOUR CODE HERE

	// encode fetched data as json and serve to client
	json.NewEncoder(w).Encode(postsArray[:numPosts])
	return
}
