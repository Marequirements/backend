package token

import (
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

func getTokenStorageInstance() *TokenStorage {
	once.Do(func() {
		tokenStorageInstance = &TokenStorage{
			tokens: make(map[string]string),
		}
	})
	return tokenStorageInstance
}

func (t *TokenStorage) addToken(key string, value string) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	t.tokens[key] = value
}

func (t *TokenStorage) checkToken(key string, value string) bool {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	if t.tokens[key] == value {
		return true
	}
	return false
}

func (t *TokenStorage) deleteToken(key string, value string) bool {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	if t.checkToken(key, value) {
		delete(t.tokens, key)
		return true
	}
	return false
}

func (*TokenStorage) generateToken() string {
	rand.Seed(time.Now().UnixNano())

	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	var result string
	for i := 0; i < 40; i++ {
		result += string(charset[rand.Intn(len(charset))])
	}
	return result
}
