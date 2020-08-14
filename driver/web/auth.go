/*
 *   Copyright (c) 2020 Board of Trustees of the University of Illinois.
 *   All rights reserved.

 *   Licensed under the Apache License, Version 2.0 (the "License");
 *   you may not use this file except in compliance with the License.
 *   You may obtain a copy of the License at

 *   http://www.apache.org/licenses/LICENSE-2.0

 *   Unless required by applicable law or agreed to in writing, software
 *   distributed under the License is distributed on an "AS IS" BASIS,
 *   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *   See the License for the specific language governing permissions and
 *   limitations under the License.
 */

package web

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"
	"talent-chooser/core"
	"talent-chooser/utils"
	"time"

	"github.com/dgrijalva/jwt-go"
	"golang.org/x/oauth2"
	"gopkg.in/ericchiang/go-oidc.v2"
)

type state struct {
	value   string
	created int64
}

//Auth handler
type Auth struct {
	app *core.Application

	host         string
	oidcProvider *oidc.Provider
	oauth2Config oauth2.Config
	statesLock   *sync.Mutex
	states       []*state

	apiKeysAuth APIKeysAuth
	jwtAuth     *JWTAuth
}

//Start starts the auth module
func (auth *Auth) Start() error {
	go auth.checkForHangingStates()

	return nil
}

func (auth *Auth) checkForHangingStates() {
	log.Println("checkForHangingStates -> start")

	toRemove := []state{}

	//find all hanging items - more than 5 minutes period from their creation
	now := time.Now().Unix()
	for _, item := range auth.states {
		difference := now - item.created

		//5 minutes
		if difference > 300 {
			toRemove = append(toRemove, *item)
		}
	}

	//remove all hanging items
	if len(toRemove) > 0 {
		for _, item := range toRemove {
			auth.removeState(item)
		}
	} else {
		log.Println("checkForHangingStates -> nothing to remove")
	}

	//log the current states
	log.Printf("checkForHangingStates -> current states %v\n", auth.states)

	nextLoad := time.Minute * 10
	log.Printf("checkForHangingStates() -> next exec after %s\n", nextLoad)
	timer := time.NewTimer(nextLoad)
	<-timer.C
	log.Println("checkForHangingStates() -> timer expired")

	auth.checkForHangingStates()
}

func (auth *Auth) login(w http.ResponseWriter, r *http.Request) {

	//generate random state
	value := utils.RandSeq(25)
	log.Printf("login() -> generated state - %s\n", utils.GetLogValue(value))
	created := time.Now().Unix()
	state := state{value, created}

	//add the generate state
	auth.addState(state)

	//generate url
	url := auth.oauth2Config.AuthCodeURL(state.value)

	//redirect
	http.Redirect(w, r, url, http.StatusFound)
}

func (auth *Auth) addState(state state) {
	auth.statesLock.Lock()
	auth.states = append(auth.states, &state)
	auth.statesLock.Unlock()
}

func (auth *Auth) removeState(state state) {
	auth.statesLock.Lock()
	log.Printf("remove - %s %d\n", state.value, state.created)

	//find the index
	index := -1
	for i, item := range auth.states {
		if state.value == item.value {
			index = i
			break
		}
	}

	//remove it
	if index != -1 {
		auth.states[index] = auth.states[len(auth.states)-1] // Copy last element to index i.
		auth.states[len(auth.states)-1] = nil                // Erase last element (write zero value).
		auth.states = auth.states[:len(auth.states)-1]       // Truncate slice.
	}

	auth.statesLock.Unlock()
}

func (auth *Auth) callback(w http.ResponseWriter, r *http.Request) {
	// Verify state and code
	qValues := r.URL.Query()

	queryState := qValues.Get("state")
	state := auth.containsState(queryState)
	if state == nil {
		log.Printf("401 - Unauthorized for state %s", queryState)

		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Unauthorized"))
		return
	}
	//remove the accepted state
	auth.removeState(*state)

	code := qValues.Get("code")
	if len(code) <= 0 {
		log.Println("401 - Unauthorized for empty code")

		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Unauthorized"))
		return
	}

	//Get token
	oauth2Token, err := auth.oauth2Config.Exchange(context.Background(), code)
	if err != nil {
		log.Printf("Error getting token %s\n", err)

		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error"))
		return
	}
	tokenSource := auth.oauth2Config.TokenSource(context.Background(), oauth2Token)

	//Get user info
	userInfo, err := auth.oidcProvider.UserInfo(context.Background(), tokenSource)
	if err != nil {
		log.Printf("Error getting user info %s\n", err)

		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error"))
		return
	}

	//Get custom claims
	var claims struct {
		Subject       string `json:"sub"`
		Profile       string `json:"profile"`
		Email         string `json:"email"`
		EmailVerified bool   `json:"email_verified"`

		Username   string `json:"preferred_username"`
		FirstName  string `json:"given_name"`
		MiddleName string `json:"middle_name"`
		FamilyName string `json:"family_name"`

		UIuceduIsMemberOf *[]string `json:"uiucedu_is_member_of"`
	}
	err = userInfo.Claims(&claims)
	if err != nil {
		log.Printf("Error getting custom claims %s\n", err)

		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error"))
		return
	}

	//Check if member of urn:mace:uiuc.edu:urbana:authman:app-rokwire-service-policy-rokwire admin app
	isMember := auth.isMemberOf(claims.UIuceduIsMemberOf)
	if !isMember {
		log.Printf("403 - Forbidden access for user %s\n", claims.Username)

		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("Access controll error"))
		return
	}

	//Ready..
	auth.onOIDCLoginSuccess(claims.Username, w, r)
}

func (auth *Auth) isMemberOf(groups *[]string) bool {
	if groups == nil {
		return false
	}
	for _, group := range *groups {
		if group == "urn:mace:uiuc.edu:urbana:authman:app-rokwire-service-policy-rokwire admin app" {
			return true
		}
	}
	return false
}

func (auth *Auth) containsState(stateValue string) *state {
	if auth.states == nil {
		return nil
	}
	for _, item := range auth.states {
		if item.value == stateValue {
			return item
		}
	}
	return nil
}

func (auth *Auth) onOIDCLoginSuccess(username string, w http.ResponseWriter, r *http.Request) {
	//now jwt
	jwtToken, expires, err := auth.jwtAuth.createToken(username)
	if err != nil {
		log.Printf("Error on creating token for user %s %s\n", username, err.Error())
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "tch-token",
		Value:    jwtToken,
		Expires:  *expires,
		Secure:   true,
		HttpOnly: true,
		Path:     "/",
	})

	redirectURL := auth.host + "/talent-chooser/home"
	http.Redirect(w, r, redirectURL, http.StatusMovedPermanently)
}

func (auth *Auth) jwtCheck(w http.ResponseWriter, r *http.Request) bool {
	return auth.jwtAuth.check(w, r)
}

func (auth *Auth) apiKeyCheck(w http.ResponseWriter, r *http.Request) bool {
	return auth.apiKeysAuth.check(w, r)
}

//NewAuth creates new auth handler
func NewAuth(app *core.Application, host string, oidcProvider string, oidcClientID string,
	oidcClientSecret string, redirectURL string, jwtKey string, appKeys []string) *Auth {

	provider, err := oidc.NewProvider(context.Background(), oidcProvider)
	if err != nil {
		log.Fatalln(err)
	}

	// Configure an OpenID Connect aware OAuth2 client.
	oauth2Config := oauth2.Config{
		ClientID:     oidcClientID,
		ClientSecret: oidcClientSecret,
		RedirectURL:  redirectURL,

		// Discovery returns the OAuth2 endpoints.
		Endpoint: provider.Endpoint(),

		// "openid" is a required scope for OpenID Connect flows.
		Scopes: []string{oidc.ScopeOpenID, "profile", "email", "offline_access"},
	}

	statesLock := &sync.Mutex{}
	states := []*state{}

	auth := Auth{app: app, host: host, oidcProvider: provider, oauth2Config: oauth2Config,
		states: states, statesLock: statesLock,
		apiKeysAuth: newAPIKeysAuth(appKeys), jwtAuth: newJWTAuth(jwtKey)}
	return &auth
}

/////////////////////////////////////

//APIKeysAuth entity
type APIKeysAuth struct {
	appKeys []string
}

func (auth APIKeysAuth) check(w http.ResponseWriter, r *http.Request) bool {
	apiKey := r.Header.Get("ROKWIRE-API-KEY")
	//check if there is api key in the header
	if len(apiKey) == 0 {
		//no key, so return 400
		log.Println(fmt.Sprintf("400 - Bad Request"))

		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Bad Request"))
		return false
	}

	//check if the api key is one of the listed
	appKeys := auth.appKeys
	exist := false
	for _, element := range appKeys {
		if element == apiKey {
			exist = true
			break
		}
	}
	if !exist {
		//not exist, so return 401
		log.Println(fmt.Sprintf("401 - Unauthorized for key %s", apiKey))

		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Unauthorized"))
		return false
	}
	return true
}

//NewAPIKeysAuth creates new api keys auth
func newAPIKeysAuth(appKeys []string) APIKeysAuth {
	auth := APIKeysAuth{appKeys}
	return auth
}

//Claims represents jwt claim
type Claims struct {
	Username string `json:"username"`
	jwt.StandardClaims
}

//JWTAuth jwt authentication
type JWTAuth struct {
	jwtKey []byte
}

func (jwtAuth *JWTAuth) createToken(username string) (string, *time.Time, error) {
	expirationTime := time.Now().Add(30 * time.Minute) //30 minutes
	claims := &Claims{
		Username: username,
		StandardClaims: jwt.StandardClaims{
			// In JWT, the expiry time is expressed as unix milliseconds
			ExpiresAt: expirationTime.Unix(),
		},
	}

	// Declare the token with the algorithm used for signing, and the claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	// Create the JWT string
	tokenString, err := token.SignedString(jwtAuth.jwtKey)
	if err != nil {
		return "", nil, err
	}

	return tokenString, &expirationTime, nil
}

func (jwtAuth *JWTAuth) check(w http.ResponseWriter, r *http.Request) bool {
	//check if there is a cookie
	token, err := r.Cookie("tch-token")
	if token == nil || err != nil {
		//no cookie, so return 400
		log.Println("400 - Bad Request")

		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Bad Request"))
		return false
	}
	log.Printf("Got cookie:%s\n", utils.GetLogValue(token.Value))

	// Initialize a new instance of `Claims`
	claims := &Claims{}

	// Parse the JWT string and store the result in `claims`.
	// Note that we are passing the key in this method as well. This method will return an error
	// if the token is invalid (if it has expired according to the expiry time we set on sign in),
	// or if the signature does not match
	tkn, err := jwt.ParseWithClaims(token.Value, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtAuth.jwtKey, nil
	})
	if err != nil {
		log.Println(err)
		if err == jwt.ErrSignatureInvalid {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("Unauthorized"))
			return false
		}
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Bad Request"))
		return false
	}
	if !tkn.Valid {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Unauthorized"))
		return false
	}
	return true
}

//NewJWTAuth creates new jwt auth
func newJWTAuth(jwtKey string) *JWTAuth {
	jwtAuth := JWTAuth{jwtKey: []byte(jwtKey)}
	return &jwtAuth
}
