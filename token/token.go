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
	log.Println("Function AddToken called")

	ts.mutex.Lock()
	defer ts.mutex.Unlock()

	ts.tokens[key] = Token{Key: key, Value: value, Role: role}
	log.Println("AddToken: Added a token ", ts.tokens[key], " to the key ", key)
}

func (ts *Storage) CheckToken(key string, value string) bool {
	log.Println("Function CheckToken called")

	ts.mutex.Lock()
	defer ts.mutex.Unlock()

	log.Println("CheckToken: checking user " + key + " and token " + value)
	token, ok := ts.tokens[key]
	if !ok {
		log.Println("CheckToken: The key ", key, " does not exist")
		return false
	}
	if token.Value == value {
		log.Println("CheckToken: Key value ", token.Value, " and provided value ", value, " are correct")
		return true
	}
	log.Println("CheckToken: Failed Key value ", token.Value, " and provided value ", value, " are not the same")
	return false
}

func (ts *Storage) DeleteToken(key string, value string) error {
	log.Println("Function DeleteToken called")

	ts.mutex.Lock()
	defer ts.mutex.Unlock()

	log.Println("DeleteToken: Deleting token for key:", key, "with value:", value)
	storedValue, ok := ts.tokens[key]
	if !ok {
		log.Println("DeleteToken: token not in storage")
		return fmt.Errorf("token not in storage")
	}
	log.Println("DeleteToken: Token: ", value, "found in token storage")
	log.Println("DeleteToken: checking the value ", value, " with the value belonging to the key ", key)
	if storedValue.Value != value {
		log.Println("DeleteToken: Value ", value, " does not match with the key value: ", storedValue.Value)
		return fmt.Errorf("value does not match")
	}
	log.Println("DeleteToken: Value ", value, "  is correct with the key value: ", storedValue.Value)
	delete(ts.tokens, key)
	log.Println("DeleteToken: Token deleted for key:", key, "with value:", value)
	return nil
}

func (*Storage) GenerateToken() string {
	log.Println("Function GenerateToken called")

	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	const length = 40

	src := rand.NewSource(time.Now().UnixNano())
	rnd := rand.New(src)

	result := make([]byte, length)
	for i := range result {
		result[i] = charset[rnd.Intn(len(charset))]
	}
	log.Println("GenerateToken: Token generated ", result)
	return string(result)
}

func (ts *Storage) GetUsernameByToken(token string) (string, error) {
	log.Println("Function GetUsernameByToken called")

	ts.mutex.RLock() // Use ts.mutex instead of ts.mu
	defer ts.mutex.RUnlock()

	log.Println("GetUsernameByToken: Searching for username/key with token= ", token, " in token storage")
	for _, mytoken := range ts.tokens {
		if mytoken.Value == token {
			log.Println("GetUsernameByToken: Found key/username = ", mytoken.Key)
			return mytoken.Key, nil
		}
	}
	log.Println("GetUsernameByToken: Token not found")
	return "", errors.New("token not found")
}

func (ts *Storage) GetRoleByToken(token string) (string, error) {
	log.Println("Function GetRoleByToken called")

	ts.mutex.RLock() // Use ts.mutex instead of ts.mu
	defer ts.mutex.RUnlock()
	log.Println("GetRoleByToken: Searching for role with token= ", token, " in token storage")
	for _, mytoken := range ts.tokens {
		if mytoken.Value == token {
			log.Println("GetRoleToken: Found role = ", mytoken.Role)
			return mytoken.Role, nil
		}
	}

	log.Println("GetRoleByToken: Token not found")
	return "", errors.New("token not found")
}
