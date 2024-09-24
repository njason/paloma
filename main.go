package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
	"time"
)

// Secret represents a secret stored in memory.
type Secret struct {
	Value     string
	ExpiresAt time.Time
}

// MemoryStore represents an in-memory store for secrets with thread-safety.
type MemoryStore struct {
	mu      sync.Mutex
	secrets map[string]Secret
}

// NewMemoryStore initializes a new memory store.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		secrets: make(map[string]Secret),
	}
}

// StoreSecret stores a secret and returns its unique key.
func (ms *MemoryStore) StoreSecret(secret string, ttl time.Duration) string {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	// Generate a unique key for the secret.
	key := generateKey(32)
	ms.secrets[key] = Secret{
		Value:     secret,
		ExpiresAt: time.Now().Add(ttl),
	}
	return key
}

// GetSecret retrieves and deletes a secret by its key.
func (ms *MemoryStore) GetSecret(key string) (string, bool) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	secret, exists := ms.secrets[key]
	if !exists || time.Now().After(secret.ExpiresAt) {
		// If secret doesn't exist or has expired, delete it.
		delete(ms.secrets, key)
		return "", false
	}

	// Delete the secret after retrieval (one-time use).
	delete(ms.secrets, key)
	return secret.Value, true
}

// generateKey generates a random base64-encoded string.
func generateKey(length int) string {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		log.Fatal("Error generating random key:", err)
	}
	return base64.URLEncoding.EncodeToString(bytes)
}

// memoryStore is the in-memory storage for secrets.
var memoryStore = NewMemoryStore()

// StoreSecretHandler handles the storage of secrets.
func StoreSecretHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	secret, err := ioutil.ReadAll(r.Body)
	if err != nil || len(secret) == 0 {
		http.Error(w, "Secret cannot be empty", http.StatusBadRequest)
		return
	}

	// Store the secret with a TTL of 1 hour.
	key := memoryStore.StoreSecret(string(secret), time.Hour)

	// Return the unique URL where the secret can be retrieved.
	secretURL := fmt.Sprintf("%s/secret/%s", r.Host, key)
	fmt.Fprintf(w, "Secret stored. Access it at: http://%s\n", secretURL)
}

// GetSecretHandler handles the retrieval of secrets.
func GetSecretHandler(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Path[len("/secret/"):]

	if key == "" {
		http.Error(w, "Secret key is required", http.StatusBadRequest)
		return
	}

	secret, exists := memoryStore.GetSecret(key)
	if !exists {
		http.Error(w, "Secret not found or has expired", http.StatusNotFound)
		return
	}

	// Return the secret and mark it as retrieved.
	fmt.Fprintf(w, "Your secret is: %s\n", secret)
}

func main() {
	http.HandleFunc("/store", StoreSecretHandler)
	http.HandleFunc("/secret/", GetSecretHandler)
	http.Handle("/", http.FileServer(http.Dir("./")))

	fmt.Println("Starting server on :8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
