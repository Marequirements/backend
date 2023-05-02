package token

import (
	"log"
	"math/rand"
	"sync"
	"time"
)

type TokenStorage struct {
	tokens map[string]string
	mutex  sync.Mutex
}

var tokenStorageInstance *TokenStorage
var once sync.Once

func GetTokenStorageInstance() *TokenStorage {
	once.Do(func() {
		tokenStorageInstance = &TokenStorage{
			tokens: make(map[string]string),
		}
	})
	return tokenStorageInstance
}

func (t *TokenStorage) AddToken(key string, value string) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	t.tokens[key] = value
}

func (t *TokenStorage) CheckToken(key string, value string) bool {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	log.Println("checking user: " + key + " and token: " + value)
	if t.tokens[key] == value {
		log.Println("found token and value")
		return true
	}
	return false
}

func (t *TokenStorage) DeleteToken(key string, value string) bool {
	response := false
	t.mutex.Lock()
	defer t.mutex.Unlock()

	log.Println("Deleting token for key:", key, "with value:", value)
	delete(t.tokens, key)
	log.Println("Token deleted for key:", key, "with value:", value)
	response = true
	return response
}

func (*TokenStorage) GenerateToken() string {
	rand.Seed(time.Now().UnixNano())

	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	var result string
	for i := 0; i < 40; i++ {
		result += string(charset[rand.Intn(len(charset))])
	}
	return result
}
