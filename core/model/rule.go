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

package model

import (
	"log"
	"talent-chooser/utils"
)

//Rule represents rule entity
type Rule struct {
	ID       int         `json:"id"`
	RuleType RuleType    `json:"rule-type"`
	Value    interface{} `json:"value"`
}

//Match matches if the input paramters match with the rule value
func (role Rule) Match(inputData InputRulesParameters) bool {
	return role.RuleType.Match(inputData, role.Value)
}

//RuleType represents rule type interface
type RuleType interface {
	GetID() int
	GetName() string
	ValidData(data interface{}) bool
	Match(inputData InputRulesParameters, ruleValue interface{}) bool
}

//InputRulesParameters wraps all input rules parameters to be passed on the match function
type InputRulesParameters struct {
	User *User

	Auth        *Auth
	AuthV2      *AuthV2
	AuthV3      *AuthV3
	AuthVersion string //1 or 2 or 3

	IlliniCash *IlliniCash
	Platform   *Platform
}

//RolesRuleType represents roles rule type entity
type RolesRuleType struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

//GetID gives the rule id
func (rr RolesRuleType) GetID() int {
	return rr.ID
}

//GetName gives the rule name
func (rr RolesRuleType) GetName() string {
	return rr.Name
}

//ValidData checks if the input data is valid for the rule type
func (rr RolesRuleType) ValidData(data interface{}) bool {
	if data == nil {
		return false
	}
	return true
}

//Match match if the input paramters match with the value rules
func (rr RolesRuleType) Match(inputData InputRulesParameters, ruleValue interface{}) bool {
	return rr.match(ruleValue, inputData.User)
}

func (rr RolesRuleType) match(roleRule interface{}, user *User) bool {

	if utils.IsString(roleRule) {
		return user != nil && user.Roles != nil && utils.Contains(*user.Roles, roleRule.(string))
	}

	if utils.IsSlice(roleRule) {
		list := roleRule.([]interface{})
		length := len(list)

		if length == 1 {
			return rr.match(list[0], user)
		}

		if length == 2 {
			operation := list[0]
			argument := list[1]
			if utils.IsString(operation) {
				if operation == "NOT" {
					return !rr.match(argument, user)
				}
			}
		}

		if length > 2 {
			result := rr.match(list[0], user)
			for index := 1; (index + 1) < length; index += 2 {
				operation := list[index]
				argument := list[index+1]
				if utils.IsString(operation) {
					opr := operation.(string)
					if opr == "AND" {
						result = result && rr.match(argument, user)
					} else if opr == "OR" {
						result = result || rr.match(argument, user)
					}
				}
			}
			return result
		}
	}

	return true // allow everything that is not defined or we do not understand
}

//NewRolesRuleType creates roles rule type instance
func NewRolesRuleType(id int, name string) RolesRuleType {
	return RolesRuleType{ID: id, Name: name}
}

//PrivacyRuleType represents privacy rule type entity
type PrivacyRuleType struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

//GetID gives the rule id
func (rr PrivacyRuleType) GetID() int {
	return rr.ID
}

//GetName gives the rule name
func (rr PrivacyRuleType) GetName() string {
	return rr.Name
}

//ValidData checks if the input data is valid for the rule type
func (rr PrivacyRuleType) ValidData(data interface{}) bool {
	if data == nil {
		return false
	}
	_, ok := data.(float64)
	if !ok {
		return false
	}
	return true
}

//Match match if the input paramters match with the value rules
func (rr PrivacyRuleType) Match(inputData InputRulesParameters, ruleValue interface{}) bool {
	//wanted min level
	minLevel := ruleValue.(float64)

	user := inputData.User
	if user == nil {
		return false // it does not match
	}
	return user.PrivacySettings.Level >= int(minLevel)
}

//NewPrivacyRuleType creates privacy rule type instance
func NewPrivacyRuleType(id int, name string) PrivacyRuleType {
	return PrivacyRuleType{ID: id, Name: name}
}

//AuthRuleType represents auth rule type entity
type AuthRuleType struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

//GetID gives the rule id
func (rr AuthRuleType) GetID() int {
	return rr.ID
}

//GetName gives the rule name
func (rr AuthRuleType) GetName() string {
	return rr.Name
}

//ValidData checks if the input data is valid for the rule type
func (rr AuthRuleType) ValidData(data interface{}) bool {
	if data == nil {
		return false
	}
	_, ok := data.(map[string]interface{})
	if !ok {
		return false
	}
	return true
}

//Match match if the input paramters match with the value rules
func (rr AuthRuleType) Match(inputData InputRulesParameters, ruleValue interface{}) bool {
	version := inputData.AuthVersion
	if version == "1" {
		return rr.matchV1(inputData, ruleValue)
	}
	if version == "2" {
		return rr.matchV2(inputData, ruleValue)
	}
	if version == "3" {
		return rr.matchV3(inputData, ruleValue)
	}
	return false
}

func (rr AuthRuleType) matchV1(inputData InputRulesParameters, ruleValue interface{}) bool {
	// supported checks
	//{ "shibbolethLoggedIn": true }
	mapData := ruleValue.(map[string]interface{})

	//it supports only shibbolethLoggedIn check, return false otherwise
	shibboVal := mapData["shibbolethLoggedIn"]
	if shibboVal == nil {
		return false
	}

	authData := inputData.Auth

	//wanted
	shibbolethLoggedIn := mapData["shibbolethLoggedIn"].(bool)
	if shibbolethLoggedIn {
		//if we want to be true

		if authData == nil {
			return false //it does not matches
		}
		return authData.UIuceduUIN != nil
	}
	//if we want to be false

	if authData == nil {
		return true //it matches
	}
	return authData.UIuceduUIN == nil
}

func (rr AuthRuleType) matchV2(inputData InputRulesParameters, ruleValue interface{}) bool {
	mapData := ruleValue.(map[string]interface{})

	//it supports the following checks
	shibbolethLoggedIn := mapData["shibbolethLoggedIn"]
	loggedIn := mapData["loggedIn"]
	phoneLoggedIn := mapData["phoneLoggedIn"]
	eventEditor := mapData["eventEditor"]
	shibbolethMemberOf := mapData["shibbolethMemberOf"]

	if shibbolethLoggedIn == nil && loggedIn == nil && phoneLoggedIn == nil && eventEditor == nil && shibbolethMemberOf == nil {
		return false //not supported
	}

	if shibbolethLoggedIn != nil {
		wantedValue := shibbolethLoggedIn.(bool)
		value := rr.isShibbolethLoggedIn(inputData.AuthV2)
		return value == wantedValue
	}

	if loggedIn != nil {
		wantedValue := loggedIn.(bool)
		value := rr.isLoggedIn(inputData.AuthV2)
		return value == wantedValue
	}

	if phoneLoggedIn != nil {
		wantedValue := phoneLoggedIn.(bool)
		value := rr.isPhoneLoggedIn(inputData.AuthV2)
		return value == wantedValue
	}

	if eventEditor != nil {
		wantedValue := eventEditor.(bool)
		value := rr.isEventEditor(inputData.AuthV2)
		return value == wantedValue
	}

	if shibbolethMemberOf != nil {
		wantedValue := shibbolethMemberOf.(string)
		return rr.shibbolethMemberOf(inputData.AuthV2, wantedValue)
	}
	return false
}

func (rr AuthRuleType) matchV3(inputData InputRulesParameters, ruleValue interface{}) bool {
	mapData := ruleValue.(map[string]interface{})

	//it supports the following checks
	shibbolethLoggedIn := mapData["shibbolethLoggedIn"]
	loggedIn := mapData["loggedIn"]
	phoneLoggedIn := mapData["phoneLoggedIn"]
	eventEditor := mapData["eventEditor"]
	shibbolethMemberOf := mapData["shibbolethMemberOf"]
	iCardNum := mapData["iCardNum"]
	iCardLibraryNum := mapData["iCardLibraryNum"]
	documentType := mapData["documentType"]

	if shibbolethLoggedIn == nil && loggedIn == nil && phoneLoggedIn == nil && eventEditor == nil &&
		shibbolethMemberOf == nil && iCardNum == nil && iCardLibraryNum == nil && documentType == nil {
		return false //not supported
	}

	if shibbolethLoggedIn != nil {
		wantedValue := shibbolethLoggedIn.(bool)
		value := rr.isShibbolethLoggedInV3(inputData.AuthV3)
		return value == wantedValue
	}

	if loggedIn != nil {
		wantedValue := loggedIn.(bool)
		value := rr.isLoggedInV3(inputData.AuthV3)
		return value == wantedValue
	}

	if phoneLoggedIn != nil {
		wantedValue := phoneLoggedIn.(bool)
		value := rr.isPhoneLoggedInV3(inputData.AuthV3)
		return value == wantedValue
	}

	if eventEditor != nil {
		wantedValue := eventEditor.(bool)
		value := rr.isEventEditorV3(inputData.AuthV3)
		return value == wantedValue
	}

	if shibbolethMemberOf != nil {
		wantedValue := shibbolethMemberOf.(string)
		return rr.shibbolethMemberOfV3(inputData.AuthV3, wantedValue)
	}

	if iCardNum != nil {
		wantedValue := iCardNum.(bool)
		value := rr.isICardNumV3(inputData.AuthV3)
		return value == wantedValue

	}

	if iCardLibraryNum != nil {
		wantedValue := iCardLibraryNum.(bool)
		value := rr.isICardLibraryNumV3(inputData.AuthV3)
		return value == wantedValue
	}

	if documentType != nil {
		wantedValue := documentType.(string)
		log.Printf("wanted:%s", wantedValue)
		value := rr.getPiiDocumentType(inputData.AuthV3)
		if value == nil {
			return false
		}
		return *value == wantedValue
	}

	return false
}

func (rr AuthRuleType) isLoggedIn(auth *AuthV2) bool {
	if auth != nil && auth.IDToken != nil {
		return true
	}
	return false
}

func (rr AuthRuleType) isLoggedInV3(auth *AuthV3) bool {
	if auth != nil && auth.Token != nil && auth.Token.IDToken != nil {
		return true
	}
	return false
}

func (rr AuthRuleType) isShibbolethLoggedIn(auth *AuthV2) bool {
	if auth != nil && auth.IDToken != nil && auth.AccessToken != nil && auth.RefreshToken != nil {
		return true
	}
	return false
}

func (rr AuthRuleType) isShibbolethLoggedInV3(auth *AuthV3) bool {
	if auth == nil {
		return false
	}
	token := auth.Token
	if token != nil && token.IDToken != nil && token.AccessToken != nil && token.RefreshToken != nil {
		return true
	}
	return false
}

func (rr AuthRuleType) isPhoneLoggedIn(auth *AuthV2) bool {
	if auth == nil {
		return false
	}

	if auth.IDToken != nil &&
		auth.AccessToken == nil && auth.RefreshToken == nil &&
		auth.PhoneNumber != nil && len(*auth.PhoneNumber) > 0 {
		return true
	}
	return false
}

func (rr AuthRuleType) isPhoneLoggedInV3(auth *AuthV3) bool {
	if auth == nil {
		return false
	}
	token := auth.Token
	if token == nil {
		return false
	}

	if token.IDToken != nil &&
		token.AccessToken == nil && token.RefreshToken == nil &&
		token.PhoneNumber != nil && len(*token.PhoneNumber) > 0 {
		return true
	}
	return false
}

func (rr AuthRuleType) isEventEditor(auth *AuthV2) bool {
	if auth == nil {
		return false
	}
	userInfo := auth.UserInfo
	if userInfo == nil {
		return false
	}
	memberOfList := userInfo.UIuceduIsMemberOf
	if memberOfList == nil {
		return false
	}
	for _, item := range *memberOfList {
		if item == "urn:mace:uiuc.edu:urbana:authman:app-rokwire-service-policy-rokwire event approvers" {
			return true
		}
	}
	return false
}

func (rr AuthRuleType) isEventEditorV3(auth *AuthV3) bool {
	if auth == nil {
		return false
	}
	userInfo := auth.User
	if userInfo == nil {
		return false
	}
	memberOfList := userInfo.UIuceduIsMemberOf
	if memberOfList == nil {
		return false
	}
	for _, item := range *memberOfList {
		if item == "urn:mace:uiuc.edu:urbana:authman:app-rokwire-service-policy-rokwire event approvers" {
			return true
		}
	}
	return false
}

func (rr AuthRuleType) shibbolethMemberOf(auth *AuthV2, wantedValue string) bool {
	if auth == nil {
		return false
	}
	userInfo := auth.UserInfo
	if userInfo == nil {
		return false
	}
	memberOfList := userInfo.UIuceduIsMemberOf
	if memberOfList == nil {
		return false
	}
	for _, item := range *memberOfList {
		if item == wantedValue {
			return true
		}
	}
	return false
}

func (rr AuthRuleType) shibbolethMemberOfV3(auth *AuthV3, wantedValue string) bool {
	if auth == nil {
		return false
	}
	userInfo := auth.User
	if userInfo == nil {
		return false
	}
	memberOfList := userInfo.UIuceduIsMemberOf
	if memberOfList == nil {
		return false
	}
	for _, item := range *memberOfList {
		if item == wantedValue {
			return true
		}
	}
	return false
}

func (rr AuthRuleType) isICardNumV3(auth *AuthV3) bool {
	if auth == nil {
		return false
	}
	card := auth.Card
	if card == nil {
		return false
	}

	if card.CardNumber != nil && len(*card.CardNumber) > 0 {
		return true
	}
	return false
}

func (rr AuthRuleType) isICardLibraryNumV3(auth *AuthV3) bool {
	if auth == nil {
		return false
	}
	card := auth.Card
	if card == nil {
		return false
	}

	if card.LibraryNumber != nil && len(*card.LibraryNumber) > 0 {
		return true
	}
	return false
}

func (rr AuthRuleType) getPiiDocumentType(auth *AuthV3) *string {
	if auth == nil {
		return nil
	}
	pii := auth.Pii
	if pii == nil {
		return nil
	}
	return pii.DocumentType
}

//NewAuthRuleType creates privacy rule type instance
func NewAuthRuleType(id int, name string) AuthRuleType {
	return AuthRuleType{ID: id, Name: name}
}

//IlliniCashRuleType represents illini cash rule type entity
type IlliniCashRuleType struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

//GetID gives the rule id
func (rr IlliniCashRuleType) GetID() int {
	return rr.ID
}

//GetName gives the rule name
func (rr IlliniCashRuleType) GetName() string {
	return rr.Name
}

//ValidData checks if the input data is valid for the rule type
func (rr IlliniCashRuleType) ValidData(data interface{}) bool {
	if data == nil {
		return false
	}
	_, ok := data.(map[string]interface{})
	if !ok {
		return false
	}
	return true
}

//Match match if the input paramters match with the value rules
func (rr IlliniCashRuleType) Match(inputData InputRulesParameters, ruleValue interface{}) bool {
	//value { "housingResidenceStatus" : true }

	mapData := ruleValue.(map[string]interface{})

	illiniCashInputData := inputData.IlliniCash

	//wanted
	housingResidenceStatus := mapData["housingResidenceStatus"].(bool)
	if housingResidenceStatus {
		//if we want to be true

		if illiniCashInputData == nil {
			return false //it does not matches
		}
		return illiniCashInputData.HousingResidentStatus == true
	}
	//if we want to be false

	if illiniCashInputData == nil {
		return true //it matches
	}
	return illiniCashInputData.HousingResidentStatus == false
}

//NewIlliniCashRuleType creates privacy rule type instance
func NewIlliniCashRuleType(id int, name string) IlliniCashRuleType {
	return IlliniCashRuleType{ID: id, Name: name}
}

//EnableRuleType represents enable rule type entity
type EnableRuleType struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

//GetID gives the rule id
func (rr EnableRuleType) GetID() int {
	return rr.ID
}

//GetName gives the rule name
func (rr EnableRuleType) GetName() string {
	return rr.Name
}

//ValidData checks if the input data is valid for the rule type
func (rr EnableRuleType) ValidData(data interface{}) bool {
	if data == nil {
		return false
	}
	_, ok := data.(bool)
	if !ok {
		return false
	}
	return true
}

//Match match if the input paramters match with the value rules
func (rr EnableRuleType) Match(inputData InputRulesParameters, ruleValue interface{}) bool {
	//it does not rely on any input parameters
	return ruleValue.(bool)
}

//NewEnableRuleType creates privacy rule instance
func NewEnableRuleType(id int, name string) EnableRuleType {
	return EnableRuleType{ID: id, Name: name}
}

//PlatformRuleType represents platform rule type entity
type PlatformRuleType struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

//GetID gives the rule id
func (rr PlatformRuleType) GetID() int {
	return rr.ID
}

//GetName gives the rule name
func (rr PlatformRuleType) GetName() string {
	return rr.Name
}

//ValidData checks if the input data is valid for the rule type
func (rr PlatformRuleType) ValidData(data interface{}) bool {
	if data == nil {
		return false
	}
	_, ok := data.(map[string]interface{})
	if !ok {
		return false
	}
	return true
}

//Match match if the input paramters match with the value rules
func (rr PlatformRuleType) Match(inputData InputRulesParameters, ruleValue interface{}) bool {
	mapData := ruleValue.(map[string]interface{})

	//it supports the following checks
	os := mapData["os"]

	if os == nil {
		return false //not supported
	}

	if os != nil {
		value := rr.getOSValue(inputData.Platform)
		if value == nil {
			return false
		}

		wantedValue := os.(string)
		return *value == wantedValue
	}

	return false
}

func (rr PlatformRuleType) getOSValue(platform *Platform) *string {
	if platform == nil {
		return nil
	}
	return platform.OS
}

//NewPlatformRuleType creates platform rule type instance
func NewPlatformRuleType(id int, name string) PlatformRuleType {
	return PlatformRuleType{ID: id, Name: name}
}

//NewRuleType creates a new rule type
func NewRuleType(id int, name string) *RuleType {
	var ruleType RuleType
	switch name {
	case "roles":
		ruleType = NewRolesRuleType(id, name)
	case "privacy":
		ruleType = NewPrivacyRuleType(id, name)
	case "auth":
		ruleType = NewAuthRuleType(id, name)
	case "illini_cash":
		ruleType = NewIlliniCashRuleType(id, name)
	case "enable":
		ruleType = NewEnableRuleType(id, name)
	case "platform":
		ruleType = NewPlatformRuleType(id, name)
	}
	return &ruleType
}
