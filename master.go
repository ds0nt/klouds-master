package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/ant0ine/go-json-rest/rest"
	"github.com/fortytw2/abdi"
	"github.com/grayj/go-json-rest-middleware-tokenauth"
	"github.com/parnurzeal/gorequest"
)

var (
	configFile = "./config.yml"
	agentToken string
	// Configuration object
)

// Serve a json api
func startAPI() {
	api := rest.NewApi()
	api.Use(AuthMiddleware())
	api.Use(rest.DefaultDevStack...)
	router, err := rest.MakeRouter(
		rest.Post("/login", login),
		rest.Post("/register", register),
		rest.Get("/apps", list),
		rest.Get("/demo/:id", demo),
		rest.Get("/claim/:id", claim),
	)
	if err != nil {
		log.Fatal(err)
	}
	api.SetApp(router)

	log.Printf("Port %s\n", Config.Port)
	log.Fatal(http.ListenAndServe(":"+Config.Port, api.MakeHandler()))
}

func main() {

	// verify files exist
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		log.Fatalf("klouds config file: %s\n", err)
	}

	err := ParseConfig(configFile)
	if err != nil {
		log.Fatal(err)
	}

	newPool()

	agentToken, _ = connectAgent()
	abdi.Key = []byte(Config.HmacKey)

	startAPI()
}

type agentResponse struct {
	Token string `json:"access_token"`
}

func connectAgent() (string, error) {
	request := gorequest.New()
	_, body, err := request.Post(Config.AgentURL+"/login").SetBasicAuth(Config.MasterUser, Config.MasterPass).End()
	if err != nil {
		return "", err[0]
	}

	data := agentResponse{}
	json.Unmarshal([]byte(body), &data)
	return data.Token, nil
}

func login(w rest.ResponseWriter, r *rest.Request) {
	// this route is protected by basic auth
	token, err := tokenauth.New()
	if err != nil {
		rest.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	/// Create a token
	rd := pool.Get()
	defer rd.Close()
	_, err = rd.Do("SET", Config.TokenNamespace+tokenauth.Hash(token), r.Env["REMOTE_USER"].(string), "EX", 604800)

	if err != nil {
		log.Panicln("Internal Server Error", err)
		rest.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	// Give token.
	w.WriteJson(map[string]string{
		"access_token": token,
	})
}

type RegisterResponse struct {
	Email string
}

func register(w rest.ResponseWriter, r *rest.Request) {
	email := r.Form["email"][0]
	password := r.Form["password"][0]

	user, err := CreateUser(email, password)
	if err != nil {
		rest.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.WriteJson(user)
}

type ListResponse struct {
	containers []string
}

func list(w rest.ResponseWriter, r *rest.Request) {
	request := gorequest.New()
	_, body, err := request.Get(Config.AgentURL+"/containers").Set("Authorization", "Token "+agentToken).End()
	if err != nil {
		rest.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	list := []string{}
	json.Unmarshal([]byte(body), &list)
	w.WriteJson(list)
}

func demo(w rest.ResponseWriter, r *rest.Request) {
	w.WriteJson("unimplemented")

}
func claim(w rest.ResponseWriter, r *rest.Request) {
	w.WriteJson("unimplemented")
}

//https://github.com/stripe/stripe-go/blob/a6e40c8d67e2563657721d2ab50ef57888b4af7b/example_test.go
// "github.com/stripe/stripe-go/client"
// "github.com/stripe/stripe-go"
// func newCustomer()  {
// 	params := &stripe.CustomerParams{
//     Balance: -123,
//     Desc:  "Stripe Developer",
//     Email: "gostripe@stripe.com",
// 		Token: token
// 	}
//
// 	customer, err := customer.New(params)
// }
