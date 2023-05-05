package token

import (
	"errors"
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"
)

type Storage struct {
	tokens map[string]Token
	mutex  sync.RWMutex
}

type Token struct {
	Key   string
	Value string
	Role  string
}

var tokenStorageInstance *Storage
var once sync.Once

func GetTokenStorageInstance() *Storage {
	once.Do(func() {
		tokenStorageInstance = &Storage{
			tokens: make(map[string]Token),
		}
	})
	return tokenStorageInstance
}

func (ts *Storage) AddToken(key string, value string, role string) {
	ts.mutex.Lock()
	defer ts.mutex.Unlock()

	ts.tokens[key] = Token{Key: key, Value: value, Role: role}
}

func (ts *Storage) CheckToken(key string, value string) bool {
	ts.mutex.Lock()
	defer ts.mutex.Unlock()

	log.Println("checking user: " + key + " and token: " + value)
	token, ok := ts.tokens[key]
	if !ok {
		return false
	}
	if token.Value == value {
		log.Println("found token and value")
		return true
	}
	return false
}

func (ts *Storage) DeleteToken(key string, value string) error {
	ts.mutex.Lock()
	defer ts.mutex.Unlock()

	log.Println("Deleting token for key:", key, "with value:", value)
	storedValue, ok := ts.tokens[key]
	if !ok {
		return fmt.Errorf("token not in storage")
	}
	if storedValue.Value != value {
		return fmt.Errorf("value does not match")
	}
	delete(ts.tokens, key)
	log.Println("Token deleted for key:", key, "with value:", value)
	return nil
}

func (*Storage) GenerateToken() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	const length = 40

	src := rand.NewSource(time.Now().UnixNano())
	rnd := rand.New(src)

	result := make([]byte, length)
	for i := range result {
		result[i] = charset[rnd.Intn(len(charset))]
	}
	return string(result)
}

func (ts *Storage) GetUsernameByToken(token string) (string, error) {
	ts.mutex.RLock() // Use ts.mutex instead of ts.mu
	defer ts.mutex.RUnlock()

	for _, mytoken := range ts.tokens {
		if mytoken.Value == token {
			return mytoken.Key, nil
		}
	}

	return "", errors.New("token not found")
}

func (ts *Storage) GetRoleByToken(token string) (string, error) {
	ts.mutex.RLock() // Use ts.mutex instead of ts.mu
	defer ts.mutex.RUnlock()

	for _, mytoken := range ts.tokens {
		if mytoken.Value == token {
			return mytoken.Role, nil
		}
	}

	return "", errors.New("token not found")
}
