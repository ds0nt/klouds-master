package main

import (
	"encoding/json"
	"fmt"

	"github.com/ant0ine/go-json-rest/rest"
	"github.com/fortytw2/abdi"
	"github.com/garyburd/redigo/redis"
	"github.com/grayj/go-json-rest-middleware-tokenauth"
)

// User as seen in redis
type User struct {
	Email      string
	HashedPass string
}

func CreateUser(email, password string) (*User, error) {
	rd := pool.Get()
	defer rd.Close()

	exists, err := redis.Bool(rd.Do("EXISTS", Config.UserNamespace+email))

	if exists {
		return nil, fmt.Errorf("that account is already registered")
	} else if err != nil {
		return nil, err
	}

	// Create a token
	hash, err := abdi.Hash(password)
	if err != nil {
		return nil, err
	}
	user := User{
		Email:      email,
		HashedPass: hash,
	}

	data, err := json.Marshal(user)
	if err != nil {
		return nil, err
	}

	_, err = rd.Do("SET", Config.UserNamespace+email, data)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// FindUser gets a user redis
func FindUser(email string) (*User, error) {
	rd := pool.Get()
	defer rd.Close()
	data, _ := redis.String(rd.Do("GET", Config.UserNamespace+email))

	user := User{}
	err := json.Unmarshal([]byte(data), &user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

//AuthMiddleware is the authorization middleware
func AuthMiddleware() rest.Middleware {
	return &rest.IfMiddleware{
		Condition: func(request *rest.Request) bool {
			return request.URL.Path != "/login"
		},
		IfTrue: &tokenauth.AuthTokenMiddleware{
			Realm: "token-auth",
			Authenticator: func(token string) string {
				rd := pool.Get()
				defer rd.Close()
				user, _ := redis.String(rd.Do("GET", Config.TokenNamespace+tokenauth.Hash(token)))
				return user
			},
		},
		IfFalse: &rest.AuthBasicMiddleware{
			Realm: "basic-auth",
			Authenticator: func(auth string, password string) bool {
				user, err := FindUser(auth)
				if err != nil {
					panic(err)
				}

				if err = abdi.Check(password, user.HashedPass); err == nil {
					fmt.Println("logged in ", user)
					return true
				}
				return false
			},
		},
	}
}
