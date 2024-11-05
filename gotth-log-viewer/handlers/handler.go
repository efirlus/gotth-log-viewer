package handlers

import (
	"encoding/json"
	"gotthlogviewer/views/index"
	"net/http"
	"strconv"
	"sync"
)

type User struct {
	Name string `json:"name"`
}

var userCache = make(map[int]User)

var cacheMutex sync.RWMutex

func HandleRoot(w http.ResponseWriter, r *http.Request) error {
	return Render(w, r, index.Index())
}

func CreateUser(w http.ResponseWriter, r *http.Request) {
	var user User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if user.Name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}

	cacheMutex.Lock()
	userCache[len(userCache)+1] = user
	cacheMutex.Unlock()

	w.WriteHeader(http.StatusNoContent)
}

func GetUser(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	cacheMutex.RLock()
	user, ok := userCache[id]
	cacheMutex.RUnlock()
	if !ok {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	j, err := json.Marshal(user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(j)
}

func DeleteUser(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if _, ok := userCache[id]; !ok {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}

	cacheMutex.Lock()
	delete(userCache, id)
	cacheMutex.Unlock()

	w.WriteHeader(http.StatusNoContent)
}
