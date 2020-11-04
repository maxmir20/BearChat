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
	"fmt"
)


func RegisterRoutes(router *mux.Router) error {
	// Why don't we put options here? Check main.go :)

	router.HandleFunc("/api/posts/{startIndex}", getFeed).Methods(http.MethodGet) 
	router.HandleFunc("/api/posts/{uuid}/{startIndex}", getPosts).Methods(http.MethodGet)
	router.HandleFunc("/api/posts/create", createPost).Methods(http.MethodPost, http.MethodOptions)
	router.HandleFunc("/api/posts/delete/{postID}", deletePost).Methods(http.MethodDelete, http.MethodOptions)

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

func getPosts(w http.ResponseWriter, r *http.Request) {

	// Load the uuid and startIndex from the url paramater into their own variables
	// Look at mux.Vars() ... -> https://godoc.org/github.com/gorilla/mux#Vars
	// make sure to use "strconv" to convert the startIndex to an integer!
	uuid := mux.Vars(r)["uuid"]
	start,err := strconv.Atoi(mux.Vars(r)["startIndex"])




	// Check if the user is authorized
	// First get the uuid from the access_token (see getUUID())
	// Compare that to the uuid we got from the url parameters, if they're not the same, return an error http.StatusUnauthorized
	// YOUR CODE HERE
	userID := getUUID(w, r)
	if userID != uuid {
		http.Error(w, errors.New("uuid does not match").Error(), http.StatusUnauthorized)
	}


	
	var posts *sql.Rows
	// var err error

	/* 
		-Get all that posts that matches our userID (or uuid)
		-Sort them chronologically (the database has a "postTime" field), hint: ORDER BY
		-Make sure to always get up to 25, and start with an offset of {startIndex} (look at the previous SQL homework for hints)\
		-As indicated by the "posts" variable, this query returns multiple rows
	*/
	//posts, err = DB.Query("SELECT * FROM posts WHERE AuthorID = ?, PostID >= ? ORDER BY PostTime LIMIT 25", uuid , start)
	posts, err = DB.Query("SELECT * FROM posts WHERE authorID = ? ORDER BY postTime LIMIT 25 OFFSET ?", &uuid, &start)


	// Check for errors from the query
	if err != nil {
		http.Error(w, errors.New("error obtainting rows: " + err.Error()).Error(), http.StatusUnauthorized)
		log.Print(err.Error())
	}

	


	var (
		content string
		postID string
		userid string
		postTime time.Time
	)
	numPosts := 0
	// Create "postsArray", which is a slice (array) of Posts. Make sure it has size 25
	// Hint: https://tour.golang.org/moretypes/13
	postsArray := make([]Post,25)

	for i := 0; i < 25 && posts.Next(); i++ {
		// Every time we call posts.Next() we get access to the next row returned from our query
		// Question: How many columns did we return
		// Reminder: Scan() scans the rows in order of their columns. See the variables defined up above for your convenience
		err = posts.Scan(&content, &postID, &userid, &postTime) //&postAuthor /* missing 5th column, decide about later*/)
		
		// Check for errors in scanning
		if err != nil {
			http.Error(w, errors.New("error scanning columns: " + err.Error()).Error(), http.StatusUnauthorized)
			log.Print(err.Error())
		}

		// Set the i-th index of postsArray to a new Post with values directly from the variables you just scanned into
		// Check post.go for the structure of a Post
		// Hint: https://gobyexample.com/structs 
		
		//YOUR CODE HERE
		postsArray[i] = Post{content, postID, userid, postTime, ""}//postAuthor /* add final column?*/}
		numPosts++
	}

	posts.Close()
	err = posts.Err()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Print(err.Error())
	}
  //encode fetched data as json and serve to client
  // Up until now, we've actually been counting the number of posts (numPosts)
  // We will always have *up to* 25 posts, but we can have less
  // However, we already allocated 25 spots in oru postsArray
  // Return the subarray that contains all of our values (which may be a subsection of our array or the entire array)
  json.NewEncoder(w).Encode(postsArray[:numPosts])
  return;
}

func createPost(w http.ResponseWriter, r *http.Request) {
	// Obtain the userID from the JSON Web Token
	// See getUUID(...)
	userID := getUUID(w, r)

	// Create a Post object and then Decode the JSON Body (which has the structure of a Post) into that object
	post := Post{}
	err := json.NewDecoder(r.Body).Decode(&post)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	// Use the uuid library to generate a post ID
	// Hint: https://godoc.org/github.com/google/uuid#New
	postID := uuid.New()

	//Load our location in PST
	pst, err := time.LoadLocation("America/Los_Angeles")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	currPST := time.Now().In(pst)

	// Insert the post into the database
	// Look at /db-server/initdb.sql for a better understanding of what you need to insert
	result , err := DB.Exec("INSERT INTO posts (content, postID, authorID, postTime) VALUES (?,?,?,?)", post.PostBody, postID, userID, currPST)
	
	// Check errors with executing the query
	if err != nil {
		http.Error(w, errors.New("error inserting the post into the database").Error(), http.StatusInternalServerError)
		log.Print(err.Error())
		return
	}

	// Make sure at least one row was affected, otherwise return an InternalServerError
	// You did something very similar in Checkpoint 2
	row_count, err := result.RowsAffected()
	if row_count == 0 {
		http.Error(w, errors.New("Cannot find the post in the database").Error(), http.StatusInternalServerError)
		log.Print(err.Error())
		return
	} 

	// What kind of HTTP header should we return since we created something?
	// Check your signup from Checkpoint 2!
	w.WriteHeader(201)

	return
}

func deletePost(w http.ResponseWriter, r *http.Request) {
	// Get the postID to delete
	// Look at mux.Vars() ... -> https://godoc.org/github.com/gorilla/mux#Vars
	postID := mux.Vars(r)["postID"]

	fmt.Sprintf("postID: %s", postID)

	// Get the uuid from the access token, see getUUID(...)
	uuid := getUUID(w, r)

	fmt.Sprintf("UUID: %s", uuid)

	var exists bool
	//check if post exists
	err := DB.QueryRow("SELECT EXISTS (SELECT postID FROM posts WHERE postID = ?)", postID).Scan(&exists)

	// Check for errors in executing the query 

	if err != nil {
		http.Error(w, errors.New("error checking if post/postID exists").Error(), http.StatusInternalServerError)
		log.Print(err.Error())
		return
	}

	// Check if the post actually exists, otherwise return an http.StatusNotFound
	if exists != true {
		http.Error(w, errors.New("the post cannot be found/doesn't exists").Error(), http.StatusNotFound)
		return
	}

	// Get the authorID of the post with the specified postID
	var authorID string
	err = DB.QueryRow("SELECT authorID FROM posts WHERE postID = ?", postID).Scan(&authorID)

	fmt.Sprintf("authorID: %s", authorID)
	
	// Check for errors in executing the query
	if err != nil {
		http.Error(w, errors.New("error getting the authorID of the post with the specified postID").Error(), http.StatusInternalServerError)
		log.Print(err.Error())
		return
	}

	// Check if the uuid from the access token is the same as the authorID from the query
	// If not, return http.StatusUnauthorized
	if uuid != authorID {
		http.Error(w, errors.New("requested source doesn't match uuid in database").Error(), http.StatusUnauthorized)
		return
	}

	// Delete the post since by now we're authorized to do so
	_, err = DB.Exec("DELETE FROM posts WHERE postID = ?", postID)
	
	// Check for errors in executing the query
	if err != nil {
		http.Error(w, errors.New("error getting the authorID of the post with the specified postID").Error(), http.StatusInternalServerError)
		log.Print(err.Error())
		return
	}
	return
}

func getFeed(w http.ResponseWriter, r *http.Request) {
	// get the start index from the url paramaters
	// based on the previous functions, you should be familiar with how to do so
	start, err := strconv.Atoi(mux.Vars(r)["startIndex"])

	// convert startIndex to int
	// YOUR CODE HERE
	
	// Check for errors in converting
	// If error, return http.StatusBadRequest
	if err != nil {
		http.Error(w, errors.New("error converting startIndex to integer").Error(), http.StatusBadRequest)
		log.Print(err.Error())
		return
	}

	// Get the userID from the access_token
	// You should now be familiar with how to do so
	userID := getUUID(w, r)
	if err != nil {
		http.Error(w, errors.New("error retrieving").Error(), http.StatusBadRequest)
		log.Print(err.Error())
		return
	}

	var posts *sql.Rows
	  
	// Obtain all of the posts where the authorID is *NOT* the current authorID
	// Sort chronologically
	// Always limit to 25 queries
	// Always start at an offset of startIndex
	//posts, err := DB.Query("SELECT * FROM posts WHERE AuthorID != ?, PostID >= ? ORDER BY postTime LIMIT 25", &userID, &start)
	posts, err = DB.Query("SELECT * FROM posts WHERE authorID != ? ORDER BY postTime LIMIT 25 OFFSET ?", &userID, &start)
	
	// Check for errors in executing the query
	if err != nil {
		http.Error(w, errors.New("error obtaining rows").Error(), http.StatusBadRequest)
		log.Print(err.Error())
		return
	}

	var (
		content string
		postID string
		userid string
		postTime time.Time
	)

	numPosts := 0
	// Put all the posts into an array of Max Size 25 and return all the filled spots
	// Almost exaclty like getPosts()
	postsArray := make([]Post, 25)
	for i :=0; i< len(postsArray) && posts.Next(); i++ {
		err = posts.Scan(&content, &postID, &userID, &postTime)
		if err != nil {
			http.Error(w, errors.New("error scanning columns: " + err.Error()).Error(), http.StatusUnauthorized)
			log.Print(err.Error())
		}

		postsArray[i] = Post{content, postID, userid, postTime, ""}
		numPosts++

	}

	posts.Close()
	err = posts.Err()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Print(err.Error())
	}
  //encode fetched data as json and serve to client
  // Up until now, we've actually been counting the number of posts (numPosts)
  // We will always have *up to* 25 posts, but we can have less
  // However, we already allocated 25 spots in oru postsArray
  // Return the subarray that contains all of our values (which may be a subsection of our array or the entire array)
  //json.NewEncoder(w).Encode(postsArray[:numPosts])

	// postJson, err := json.Marshal(postsArray[:numPosts])
 //    if err != nil {
 //        log.Fatal("Cannot encode to JSON ", err)
 //    }

	//json.NewEncoder(w).Encode(postJson)
	json.NewEncoder(w).Encode(postsArray[:numPosts])

	//return postsArray[:numPosts]

	return
}
