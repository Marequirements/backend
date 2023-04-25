package token

import (
	"math/rand"
	"time"
)

func generateToken() string {
	rand.Seed(time.Now().UnixNano())

	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	var result string
	for i := 0; i < 40; i++ {
		result += string(charset[rand.Intn(len(charset))])
	}
	return result
}

// package token

// import (
//     "sync"
// )

// type Token struct {
//     // Add any fields you need for the token struct
// }

// type TokenSingleton struct {
//     tokens map[string]*Token
//     mutex  sync.Mutex
// }

// var instance *TokenSingleton

// func GetInstance() *TokenSingleton {
//     if instance == nil {
//         instance = &TokenSingleton{
//             tokens: make(map[string]*Token),
//         }
//     }
//     return instance
// }

// func (ts *TokenSingleton) GetToken(key string) *Token {
//     ts.mutex.Lock()
//     defer ts.mutex.Unlock()

//     if token, ok := ts.tokens[key]; ok {
//         return token
//     }

//     // If the token doesn't exist, create a new one and add it to the map
//     token := &Token{}
//     ts.tokens[key] = token
//     return token
// }
