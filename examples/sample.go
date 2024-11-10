package main

import (
	"fmt"
	"log"
	"net/http"
	"encoding/json"
	"strconv"

	"github.com/sys-apps-go/gorouter/pkg/router"
)

type Post struct {
	ID      int    `json:"id"`
	Title   string `json:"title"`
	Content string `json:"content"`
}

var PostList []Post
var lastPostID int

type User struct {
	Name  string `json:"name"`
	EMail string `json:"email"`
}

var UserList []User
var semaphore chan struct{}
var getCount int

func main() {
	UserList = make([]User, 0)
	UserList = make([]User, 0)
	PostList = make([]Post, 0)
	lastPostID = 0
	routerTest := router.NewRouter()

	// Root route
	routerTest.GET("/", func(c *router.Context) {
		c.String(http.StatusOK, "Welcome to the API!")
	})

	// API group
	api := routerTest.Group("/api")
	{
		// Users routes
		users := api.Group("/users")
		{
			users.GET("", listUsers)
			users.POST("", createUser)
			users.GET("/:id", getUser)
			users.PUT("/:id", updateUser)
			users.DELETE("/:id", deleteUser)
		}

		// Posts routes
		posts := api.Group("/posts")
		{
			posts.GET("", listPosts)
			posts.POST("", createPost)
			posts.GET("/:id", getPost)
			posts.PUT("/:id", updatePost)
			posts.DELETE("/:id", deletePost)
		}
	}

	routerTest.PrintRoutes()
	u := User{
		Name: "John Doe",
		EMail: "john.doe@gmail.com",
	}
	UserList = append(UserList, u)
	UserList = append(UserList, u)
	UserList = append(UserList, u)
	UserList = append(UserList, u)
	UserList = append(UserList, u)
	UserList = append(UserList, u)

	// Start the server
	fmt.Println("Server is running on http://localhost:50051")
	log.Fatal(http.ListenAndServe(":50051", routerTest))
}

// Handler functions
func listUsers(c *router.Context) {
	getCount++
	if len(UserList) == 0 {
		c.JSON(http.StatusOK, map[string]string{"message": "No users found"})
		return
	}

	c.JSON(http.StatusOK, map[string]interface{}{
		"message": "List of users",
		"users":   UserList,
	})
}

func createUser(c *router.Context) {
	var newUser User

	// Parse the JSON body
	err := json.NewDecoder(c.Request.Body).Decode(&newUser)
	if err != nil {
		c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid JSON"})
		return
	}

	// Validate the new user data
	if newUser.Name == "" || newUser.EMail == "" {
		c.JSON(http.StatusBadRequest, map[string]string{"error": "Name and email are required"})
		return
	}

	// Append the new user to UserList
	UserList = append(UserList, newUser)

	// Respond with success message
	c.JSON(http.StatusCreated, map[string]string{"message": "User created successfully"})
}

func getUser(c *router.Context) {
	id := c.Param("id")
	for _, user := range UserList {
		if user.EMail == id { // Assuming email is unique and used as ID
			c.JSON(http.StatusOK, user)
			return
		}
	}
	c.JSON(http.StatusNotFound, map[string]string{"message": fmt.Sprintf("User %s not found", id)})
}

func updateUser(c *router.Context) {
	id := c.Param("id")
	var updatedUser User
	if err := json.NewDecoder(c.Request.Body).Decode(&updatedUser); err != nil {
		c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid JSON"})
		return
	}
	for i, user := range UserList {
		if user.EMail == id {
			UserList[i] = updatedUser
			c.JSON(http.StatusOK, map[string]string{"message": fmt.Sprintf("User %s updated", id)})
			return
		}
	}
	c.JSON(http.StatusNotFound, map[string]string{"message": fmt.Sprintf("User %s not found", id)})
}

func deleteUser(c *router.Context) {
	id := c.Param("id")
	for i, user := range UserList {
		if user.EMail == id {
			UserList = append(UserList[:i], UserList[i+1:]...)
			c.JSON(http.StatusOK, map[string]string{"message": fmt.Sprintf("User %s deleted", id)})
			return
		}
	}
	c.JSON(http.StatusNotFound, map[string]string{"message": fmt.Sprintf("User %s not found", id)})
}

func listPosts(c *router.Context) {
	c.JSON(http.StatusOK, map[string]interface{}{
		"message": "List of posts",
		"posts":   PostList,
	})
}

func createPost(c *router.Context) {
	var newPost Post
	if err := json.NewDecoder(c.Request.Body).Decode(&newPost); err != nil {
		c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid JSON"})
		return
	}
	lastPostID++
	newPost.ID = lastPostID
	PostList = append(PostList, newPost)
	c.JSON(http.StatusCreated, map[string]string{"message": "Post created", "id": fmt.Sprintf("%d", newPost.ID)})
}

func getPost(c *router.Context) {
	id := c.Param("id")
	postID, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid post ID"})
		return
	}
	for _, post := range PostList {
		if post.ID == postID {
			c.JSON(http.StatusOK, post)
			return
		}
	}
	c.JSON(http.StatusNotFound, map[string]string{"message": fmt.Sprintf("Post %s not found", id)})
}

func updatePost(c *router.Context) {
	id := c.Param("id")
	postID, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid post ID"})
		return
	}
	var updatedPost Post
	if err := json.NewDecoder(c.Request.Body).Decode(&updatedPost); err != nil {
		c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid JSON"})
		return
	}
	for i, post := range PostList {
		if post.ID == postID {
			updatedPost.ID = postID
			PostList[i] = updatedPost
			c.JSON(http.StatusOK, map[string]string{"message": fmt.Sprintf("Post %s updated", id)})
			return
		}
	}
	c.JSON(http.StatusNotFound, map[string]string{"message": fmt.Sprintf("Post %s not found", id)})
}

func deletePost(c *router.Context) {
	id := c.Param("id")
	postID, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid post ID"})
		return
	}
	for i, post := range PostList {
		if post.ID == postID {
			PostList = append(PostList[:i], PostList[i+1:]...)
			c.JSON(http.StatusOK, map[string]string{"message": fmt.Sprintf("Post %s deleted", id)})
			return
		}
	}
	c.JSON(http.StatusNotFound, map[string]string{"message": fmt.Sprintf("Post %s not found", id)})
}
