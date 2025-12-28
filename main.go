package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

type BudgetPalServer struct {
	dataStore  *DataStore
	httpServer *http.Server
}

type UserResponse struct {
	ID    int     `json:"id"`
	Name  string  `json:"name,omitempty"`
	Email string  `json:"email,omitempty"`
	Links []Links `json:"links,omitempty"`
}

type Links struct {
	Rel  string `json:"rel"`
	Href string `json:"href"`
}

func (server *BudgetPalServer) handlePostUser(w http.ResponseWriter, r *http.Request) {
	id := server.dataStore.CreateUser()

	response := UserResponse{
		ID: id,
		Links: []Links{
			{Rel: "self", Href: fmt.Sprintf("/users/%d", id)},
			{Rel: "delete", Href: fmt.Sprintf("/users/%d", id)},
			{Rel: "users", Href: "/users/"},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

func (server *BudgetPalServer) handleGetUser(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	user, exists := server.dataStore.GetUser(id)
	if !exists {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	response := UserResponse{
		ID:    user.id,
		Name:  user.name,
		Email: user.email,
		Links: []Links{
			{Rel: "self", Href: fmt.Sprintf("/users/%d", user.id)},
			{Rel: "delete", Href: fmt.Sprintf("/users/%d", user.id)},
			{Rel: "users", Href: "/users/"},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (server *BudgetPalServer) handleGetAllUsers(w http.ResponseWriter, r *http.Request) {
	users := make([]UserResponse, 0)

	for _, user := range server.dataStore.users {
		userResponse := UserResponse{
			ID:    user.id,
			Name:  user.name,
			Email: user.email,
			Links: []Links{
				{Rel: "self", Href: fmt.Sprintf("/users/%d", user.id)},
				{Rel: "update", Href: fmt.Sprintf("/users/%d", user.id)},
				{Rel: "delete", Href: fmt.Sprintf("/users/%d", user.id)},
			},
		}
		users = append(users, userResponse)
	}

	response := struct {
		Users []UserResponse `json:"users"`
		Links []Links        `json:"links"`
	}{
		Users: users,
		Links: []Links{
			{Rel: "self", Href: "/users/"},
			{Rel: "create", Href: "/users/"},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (server *BudgetPalServer) handlePutUser(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	_, exists := server.dataStore.GetUser(id)
	if !exists {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	var requestBody struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	success := server.dataStore.UpdateUser(id, requestBody.Name, requestBody.Email)
	if !success {
		http.Error(w, "Failed to update user", http.StatusInternalServerError)
		return
	}

	user, _ := server.dataStore.GetUser(id)
	response := UserResponse{
		ID:    user.id,
		Name:  user.name,
		Email: user.email,
		Links: []Links{
			{Rel: "self", Href: fmt.Sprintf("/users/%d", id)},
			{Rel: "delete", Href: fmt.Sprintf("/users/%d", id)},
			{Rel: "users", Href: "/users/"},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (server *BudgetPalServer) handlePatchUser(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	user, exists := server.dataStore.GetUser(id)
	if !exists {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	var requestBody struct {
		Name  *string `json:"name,omitempty"`
		Email *string `json:"email,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	name := user.name
	email := user.email

	success := server.dataStore.UpdateUser(id, name, email)
	if !success {
		http.Error(w, "Failed to update user", http.StatusInternalServerError)
		return
	}

	updatedUser, _ := server.dataStore.GetUser(id)
	response := UserResponse{
		ID:    updatedUser.id,
		Name:  updatedUser.name,
		Email: updatedUser.email,
		Links: []Links{
			{Rel: "self", Href: fmt.Sprintf("/users/%d", id)},
			{Rel: "delete", Href: fmt.Sprintf("/users/%d", id)},
			{Rel: "users", Href: "/users/"},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (server *BudgetPalServer) handleDeleteUser(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	exists := server.dataStore.DeleteUser(id)
	if !exists {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	response := struct {
		Message string  `json:"message"`
		Links   []Links `json:"links"`
	}{
		Message: "User deleted successfully",
		Links: []Links{
			{Rel: "users", Href: "/users/"},
			{Rel: "create", Href: "/users/"},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func initServer() *BudgetPalServer {
	server := &BudgetPalServer{
		dataStore: CreateDataStore(),
	}
	mux := http.NewServeMux()
	mux.HandleFunc("POST /users/", server.handlePostUser)
	mux.HandleFunc("GET /users/{id}", server.handleGetUser)
	mux.HandleFunc("GET /users/", server.handleGetAllUsers)
	mux.HandleFunc("PUT /users/{id}", server.handlePutUser)
	mux.HandleFunc("PATCH /users/{id}", server.handlePatchUser)
	mux.HandleFunc("DELETE /users/{id}", server.handleDeleteUser)

	server.httpServer = &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}
	return server
}

func serverShutdown(server *http.Server, serverError chan error) {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverError:
		log.Fatalf("Server error: %v", err)
	case <-stop:
		log.Printf("Received stop signal %v", stop)
	}

	log.Println("Shutting down the server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exiting")
}

func main() {
	server := initServer()
	serverError := make(chan error, 1)

	go func() {
		log.Printf("Budget-Pal RESTful API-Server is running on http://localhost%s", server.httpServer.Addr)
		if err := server.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverError <- err
		}
	}()

	serverShutdown(server.httpServer, serverError)
}
