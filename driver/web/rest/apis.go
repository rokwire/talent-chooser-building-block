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

package rest

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"talent-chooser/core"
	"talent-chooser/core/model"
)

//ApisHandler handles the rest APIs implementation
type ApisHandler struct {
	app *core.Application
}

type getUIContentData struct {
	User       *model.User       `json:"user"`
	Auth       *model.Auth       `json:"auth"`
	IlliniCash *model.IlliniCash `json:"illini_cash"`
}

type getUIContentDataV2 struct {
	User       *model.User       `json:"user"`
	Auth       *model.AuthV2     `json:"auth"`
	IlliniCash *model.IlliniCash `json:"illini_cash"`
}

type getUIContentDataV3 struct {
	User       *userV3       `json:"user"`
	AuthToken  *authTokenV3  `json:"auth_token"`
	AuthUser   *authUserV3   `json:"auth_user"`
	AuthCard   *authCardV3   `json:"card"`
	IlliniCash *illiniCashV3 `json:"illini_cash"`
	Platform   *platformV3   `json:"platform"`
	Pii        *piiV3        `json:"pii"`
} // @name getUIContentRequest

type userV3 struct {
	UUID            *string           `json:"uuid"`
	PrivacySettings privacySettingsV3 `json:"privacySettings"`
	Roles           *[]string         `json:"roles"`
} // @name User

type privacySettingsV3 struct {
	Level int `json:"level"`
} // @name PrivacySettings

type authTokenV3 struct {
	IDToken      *string `json:"id_token"`
	AccessToken  *string `json:"access_token"`
	RefreshToken *string `json:"refresh_token"`
	PhoneNumber  *string `json:"phone"`
} // @name Token

type authUserV3 struct {
	Name              *string   `json:"name"`
	GivenName         *string   `json:"given_name"`
	FamilyName        *string   `json:"family_name"`
	UIuceduUIN        *string   `json:"uiucedu_uin"`
	PreferredUsername *string   `json:"preferred_username"`
	Sub               *string   `json:"sub"`
	Email             *string   `json:"email"`
	UIuceduIsMemberOf *[]string `json:"uiucedu_is_member_of"`
} // @name UserInfo

type authCardV3 struct {
	CardNumber    *string `json:"card_number"`
	LibraryNumber *string `json:"library_number"`
} // @name Card

type illiniCashV3 struct {
	HousingResidentStatus bool `json:"HousingResidentStatus"`
} // @name IliniCash

type platformV3 struct {
	OS *string `json:"os"`
} // @name Platform

type piiV3 struct {
	DocumentType *string `json:"documentType"`
} // @name Pii

//Version gives the service version
// @Description Gives the service version.
// @Tags APIs
// @ID Version
// @Produce plain
// @Success 200 {string} v1.1.0
// @Router /api/version [get]
func (h ApisHandler) Version(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(h.app.Services.GetVersion()))
}

//GetUIContent gives the ui content based on the parameters
func (h ApisHandler) GetUIContent(w http.ResponseWriter, r *http.Request) {
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error on marshal the ui flat data - %s\n", err.Error())
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	var requestData getUIContentData
	//handle the params data if available
	if len(data) > 0 {
		err = json.Unmarshal(data, &requestData)
		if err != nil {
			log.Printf("Error on unmarshal the request data - %s\n", err.Error())
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

	uiContent := h.app.Services.GetUIContent(requestData.User, "1.2", requestData.Auth, requestData.IlliniCash)
	data, err = json.Marshal(uiContent)
	if err != nil {
		log.Println("Error on marshal the ui flat data")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

//GetUIContentV2 gives the ui content based on the parameters - V2
func (h ApisHandler) GetUIContentV2(w http.ResponseWriter, r *http.Request) {
	dataVersionKeys, ok := r.URL.Query()["data-version"]
	var dataVersion string
	if !ok || len(dataVersionKeys[0]) < 1 {
		//in this case we set 1.2
		dataVersion = "1.2"
	} else {
		dataVersion = dataVersionKeys[0]
	}

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("GetUIContentV2 -> error on marshal the ui flat data - %s\n", err.Error())
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	var requestData getUIContentDataV2
	//handle the params data if available
	if len(data) > 0 {
		err = json.Unmarshal(data, &requestData)
		if err != nil {
			log.Printf("GetUIContentV2 -> error on unmarshal the request data - %s\n", err.Error())
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

	uiContent := h.app.Services.GetUIContentV2(requestData.User, dataVersion, requestData.Auth, requestData.IlliniCash)
	data, err = json.Marshal(uiContent)
	if err != nil {
		log.Println("GetUIContentV2 -> error on marshal the ui flat data")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

func (h ApisHandler) getDataVersion(r *http.Request) string {
	dataVersionKeys, ok := r.URL.Query()["data-version"]
	if !ok || len(dataVersionKeys[0]) < 1 {
		//in case there is no valid input then return the latest version
		return h.getLatestVersion()
	}
	dataVersion := dataVersionKeys[0]
	if h.isSupportedVersion(dataVersion) {
		return dataVersion
	}

	return h.getLatestVersion()
}

func (h ApisHandler) getLatestVersion() string {
	return "2.2"
}

func (h ApisHandler) isSupportedVersion(v string) bool {
	return v == "1.2" || v == "2.0" || v == "2.1" || v == "2.2"
}

//Swag does not support map!
type getUIContentV3SwagReturn map[string][]string // @name UIContent

//GetUIContentV3 gives the ui content based on the parameters - V3
// @Description Gives the ui content based on the parameters.
// @Tags APIs
// @ID GetUIContentV3
// @Accept json
// @Produce json
// @Param data-version query string false "for example '2.2'"
// @Param data body getUIContentDataV3 true "body data"
// @Success 200 {object} getUIContentV3SwagReturn
// @Security RokwireAuth
// @Router /api/v3/ui-content [get]
func (h ApisHandler) GetUIContentV3(w http.ResponseWriter, r *http.Request) {
	dataVersion := h.getDataVersion(r)
	log.Println(dataVersion)

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("GetUIContentV3 -> error on marshal the ui flat data - %s\n", err.Error())
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	var requestData getUIContentDataV3
	//handle the params data if available
	if len(data) > 0 {
		err = json.Unmarshal(data, &requestData)
		if err != nil {
			log.Printf("GetUIContentV3 -> error on unmarshal the request data - %s\n", err.Error())
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

	//user
	var user *model.User
	reqUser := requestData.User
	if reqUser != nil {
		privacySettings := model.PrivacySettings{Level: reqUser.PrivacySettings.Level}
		user = &model.User{UUID: *reqUser.UUID, PrivacySettings: privacySettings, Roles: reqUser.Roles}
	}

	//auth
	auth := &model.AuthV3{}
	reqToken := requestData.AuthToken
	if reqToken != nil {
		token := &model.AuthToken{IDToken: reqToken.IDToken, AccessToken: reqToken.AccessToken,
			RefreshToken: reqToken.RefreshToken, PhoneNumber: reqToken.PhoneNumber}
		auth.Token = token
	}
	reqAuthUser := requestData.AuthUser
	if reqAuthUser != nil {
		authUser := &model.AuthUser{Name: reqAuthUser.Name, GivenName: reqAuthUser.GivenName,
			FamilyName: reqAuthUser.FamilyName, UIuceduUIN: reqAuthUser.UIuceduUIN,
			PreferredUsername: reqAuthUser.PreferredUsername, Sub: reqAuthUser.Sub,
			Email: reqAuthUser.Email, UIuceduIsMemberOf: reqAuthUser.UIuceduIsMemberOf}
		auth.User = authUser
	}
	reqAuthCard := requestData.AuthCard
	if reqAuthCard != nil {
		card := &model.AuthCard{CardNumber: reqAuthCard.CardNumber, LibraryNumber: reqAuthCard.LibraryNumber}
		auth.Card = card
	}
	reqAuthPii := requestData.Pii
	if reqAuthPii != nil {
		pii := &model.Pii{DocumentType: reqAuthPii.DocumentType}
		auth.Pii = pii
	}

	//illini cash
	var illiniCash *model.IlliniCash
	reqIlliniCash := requestData.IlliniCash
	if reqIlliniCash != nil {
		illiniCash = &model.IlliniCash{HousingResidentStatus: reqIlliniCash.HousingResidentStatus}
	}

	//platform
	var platform *model.Platform
	reqPlatform := requestData.Platform
	if reqPlatform != nil {
		platform = &model.Platform{OS: reqPlatform.OS}
	}

	uiContent := h.app.Services.GetUIContentV3(user, dataVersion, auth, illiniCash, platform)
	data, err = json.Marshal(uiContent)
	if err != nil {
		log.Println("GetUIContentV3 -> error on marshal the ui flat data")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

//NewApisHandler creates new rest Handler instance
func NewApisHandler(app *core.Application) ApisHandler {
	return ApisHandler{app: app}
}
