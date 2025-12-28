package main

import (
	"log"
	"math/rand"
)

type User struct {
	name  string
	email string
	id    int
}

type DataStore struct {
	users map[int]User
}

func CreateDataStore() *DataStore {
	return &DataStore{users: make(map[int]User)}
}

func (ds *DataStore) CreateUser() int {
	id := ds.GenerateUserID()
	for _, exists := ds.users[id]; exists; {
		id = ds.GenerateUserID()
	}
	user := User{
		id: id,
	}
	ds.users[id] = user
	log.Printf("User %d created", id)
	return id
}

func (ds *DataStore) GetUser(id int) (User, bool) {
	user, exists := ds.users[id]
	return user, exists
}

func (ds *DataStore) DeleteUser(id int) bool {
	_, exists := ds.users[id]
	delete(ds.users, id)
	return exists
}

func (ds *DataStore) UpdateUser(id int, name string, email string) bool {
	user, exists := ds.users[id]
	if !exists {
		return false
	}
	if name == "" {
		name = user.name
	}
	if email == "" {
		email = user.email
	}
	user.name = name
	user.email = email
	ds.users[id] = user
	return true
}

func (ds *DataStore) GenerateUserID() int {
	return rand.Int()
}
