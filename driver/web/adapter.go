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
	"fmt"
	"log"
	"net/http"
	"talent-chooser/core"
	"talent-chooser/driver/web/rest"
	"talent-chooser/utils"

	"github.com/gorilla/mux"

	httpSwagger "github.com/swaggo/http-swagger"
)

//Adapter entity
type Adapter struct {
	host string
	auth *Auth

	apisHandler      rest.ApisHandler
	adminApisHandler rest.AdminApisHandler

	app *core.Application
}

// @title Rokwire Talent Chooser Building Block API
// @description Rokwire Talent Chooser Building Block API Documentation.
// @version 1.8.0
// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html
// @host localhost
// @BasePath /talent-chooser
// @schemes https

// @securityDefinitions.apikey RokwireAuth
// @in header
// @name ROKWIRE-API-KEY

//Start starts the web server
func (we Adapter) Start() {
	//start the auth module
	err := we.auth.Start()
	if err != nil {
		log.Fatal(err)
	}

	router := mux.NewRouter().StrictSlash(true)

	// handle ui content
	router.HandleFunc("/talent-chooser/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./driver/web/ui/login.html")
	})
	router.HandleFunc("/talent-chooser/home", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./driver/web/ui/home.html")
	})
	router.HandleFunc("/talent-chooser/content-items", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./driver/web/ui/content-items.html")
	})
	router.HandleFunc("/talent-chooser/new-content-item", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./driver/web/ui/new-content-item.html")
	})
	router.HandleFunc("/talent-chooser/edit-content-item", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./driver/web/ui/edit-content-item.html")
	})
	router.HandleFunc("/talent-chooser/new-ui-item", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./driver/web/ui/new-ui-item.html")
	})
	router.HandleFunc("/talent-chooser/edit-ui-item", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./driver/web/ui/edit-ui-item.html")
	})
	router.HandleFunc("/talent-chooser/new-rule", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./driver/web/ui/new-rule.html")
	})
	router.HandleFunc("/talent-chooser/edit-rule", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./driver/web/ui/edit-rule.html")
	})

	//handle doc apis
	router.PathPrefix("/talent-chooser/doc/ui").Handler(we.serveDocUI())
	router.HandleFunc("/talent-chooser/doc", we.serveDoc)

	// handle login apis
	authSubrouter := router.PathPrefix("/talent-chooser/auth").Subrouter()
	authSubrouter.HandleFunc("/login", we.wrapFunc(we.auth.login)).Methods("GET")
	authSubrouter.HandleFunc("/callback", we.wrapFunc(we.auth.callback)).Methods("GET")

	// handle rest apis
	restSubrouter := router.PathPrefix("/talent-chooser/api").Subrouter()
	restSubrouter.HandleFunc("/version", we.wrapFunc(we.apisHandler.Version)).Methods("GET")
	restSubrouter.HandleFunc("/ui-content", we.apiKeysAuthWrapFunc(we.apisHandler.GetUIContent)).Methods("GET")
	restSubrouter.HandleFunc("/v2/ui-content", we.apiKeysAuthWrapFunc(we.apisHandler.GetUIContentV2)).Methods("GET")
	restSubrouter.HandleFunc("/v3/ui-content", we.apiKeysAuthWrapFunc(we.apisHandler.GetUIContentV3)).Methods("GET")

	// handle admin rest apis
	adminrestSubrouter := router.PathPrefix("/talent-chooser/admin").Subrouter()

	adminrestSubrouter.HandleFunc("/data-version", we.jwtAuthWrapFunc(we.adminApisHandler.SetDataVersion)).Methods("PUT")
	adminrestSubrouter.HandleFunc("/data-version", we.jwtAuthWrapFunc(we.adminApisHandler.GetDataVersion)).Methods("GET")

	adminrestSubrouter.HandleFunc("/config", we.jwtAuthWrapFunc(we.adminApisHandler.GetConfig)).Methods("GET")
	adminrestSubrouter.HandleFunc("/ui-content", we.jwtAuthWrapFunc(we.adminApisHandler.GetFullUIContent)).Methods("GET")
	adminrestSubrouter.HandleFunc("/ui-content/reload", we.jwtAuthWrapFunc(we.adminApisHandler.ReloadUIContent)).Methods("GET")

	adminrestSubrouter.HandleFunc("/content-items", we.jwtAuthWrapFunc(we.adminApisHandler.GetContentItems)).Methods("GET")
	adminrestSubrouter.HandleFunc("/content-items/{id}", we.jwtAuthWrapFunc(we.adminApisHandler.GetContentItem)).Methods("GET")
	adminrestSubrouter.HandleFunc("/content-items", we.jwtAuthWrapFunc(we.adminApisHandler.CreateContentItem)).Methods("POST")
	adminrestSubrouter.HandleFunc("/content-items/{id}", we.jwtAuthWrapFunc(we.adminApisHandler.UpdateContentItem)).Methods("PUT")
	adminrestSubrouter.HandleFunc("/content-items/{id}", we.jwtAuthWrapFunc(we.adminApisHandler.DeleteContentItem)).Methods("DELETE")

	adminrestSubrouter.HandleFunc("/content-items/{content-item-id}/ui-items/{id}", we.jwtAuthWrapFunc(we.adminApisHandler.GetUIItem)).Methods("GET")
	adminrestSubrouter.HandleFunc("/content-items/{content-item-id}/ui-items", we.jwtAuthWrapFunc(we.adminApisHandler.CreateUIItem)).Methods("POST")
	adminrestSubrouter.HandleFunc("/content-items/{content-item-id}/ui-items/{id}", we.jwtAuthWrapFunc(we.adminApisHandler.UpdateUIItem)).Methods("PUT")
	adminrestSubrouter.HandleFunc("/content-items/{content-item-id}/ui-items/{id}", we.jwtAuthWrapFunc(we.adminApisHandler.DeleteUIItem)).Methods("DELETE")

	adminrestSubrouter.HandleFunc("/ui-items/{ui-item-id}/rules/{id}", we.jwtAuthWrapFunc(we.adminApisHandler.GetRule)).Methods("GET")
	adminrestSubrouter.HandleFunc("/ui-items/{ui-item-id}/rules", we.jwtAuthWrapFunc(we.adminApisHandler.CreateRule)).Methods("POST")
	adminrestSubrouter.HandleFunc("/ui-items/{ui-item-id}/rules/{id}", we.jwtAuthWrapFunc(we.adminApisHandler.UpdateRule)).Methods("PUT")
	adminrestSubrouter.HandleFunc("/ui-items/{ui-item-id}/rules/{id}", we.jwtAuthWrapFunc(we.adminApisHandler.DeleteRule)).Methods("DELETE")

	adminrestSubrouter.HandleFunc("/rule-types", we.jwtAuthWrapFunc(we.adminApisHandler.GetRuleTypes)).Methods("GET")

	log.Fatal(http.ListenAndServe(":80", router))
}

func (we Adapter) serveDoc(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("access-control-allow-origin", "*")
	http.ServeFile(w, r, "./docs/swagger.yaml")
}

func (we Adapter) serveDocUI() http.Handler {
	url := fmt.Sprintf("%s/talent-chooser/doc", we.host)
	return httpSwagger.Handler(httpSwagger.URL(url))
}

func (we Adapter) wrapFunc(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		utils.LogRequest(req)

		handler(w, req)
	}
}

func (we Adapter) apiKeysAuthWrapFunc(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		utils.LogRequest(req)

		authenticated := we.auth.apiKeyCheck(w, req)
		if !authenticated {
			return
		}

		handler(w, req)
	}
}

func (we Adapter) jwtAuthWrapFunc(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		utils.LogRequest(req)

		authenticated := we.auth.jwtCheck(w, req)
		if !authenticated {
			return
		}

		handler(w, req)
	}
}

//NewWebAdapter creates new WebAdapter instance
func NewWebAdapter(appKeys []string, jwtKey string, app *core.Application,
	host string, oidcProvider string, oidcClientID string, oidcClientSecret string,
	redirectURL string) Adapter {

	auth := NewAuth(app, host, oidcProvider, oidcClientID, oidcClientSecret, redirectURL, jwtKey, appKeys)

	apisHandler := rest.NewApisHandler(app)
	adminApisHandler := rest.NewAdminApisHandler(app)

	return Adapter{host: host, auth: auth, apisHandler: apisHandler, adminApisHandler: adminApisHandler, app: app}
}
