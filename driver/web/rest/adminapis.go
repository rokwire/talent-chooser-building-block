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
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"talent-chooser/core"

	"github.com/gorilla/mux"
)

type setDataVersion struct {
	DataVersion string `json:"data-version"`
}

type createContentItem struct {
	Name string `json:"name"`
}

type updateContentItem struct {
	Name string `json:"name"`
}

type createUIItem struct {
	Name  string `json:"name"`
	Order int    `json:"order"`
}

type updateUIItem struct {
	Name  string `json:"name"`
	Order int    `json:"order"`
}

type createRule struct {
	RuleTypeID int         `json:"rule-type-id"`
	Value      interface{} `json:"value"`
}

type updateRule struct {
	RuleTypeID int         `json:"rule-type-id"`
	Value      interface{} `json:"value"`
}

//AdminApisHandler handles the admin rest APIs implementation
type AdminApisHandler struct {
	app *core.Application
}

//SetDataVersion sets the passes data version in a cookie
func (h AdminApisHandler) SetDataVersion(w http.ResponseWriter, r *http.Request) {
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error on marshal set data version - %s\n", err.Error())
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	var requestData setDataVersion
	err = json.Unmarshal(data, &requestData)
	if err != nil {
		log.Printf("Error on unmarshal the create ui item request data - %s\n", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	dataVersion := requestData.DataVersion
	if !(dataVersion == "1.2" || dataVersion == "2.0" || dataVersion == "2.1" || dataVersion == "2.2" ||
		dataVersion == "2.3" || dataVersion == "2.4" || dataVersion == "2.5" || dataVersion == "2.6" ||
		dataVersion == "2.7" || dataVersion == "3.0") {
		log.Println("Not valid data version")
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	//set cookie
	http.SetCookie(w, &http.Cookie{
		Name:  "tch-data-version",
		Value: dataVersion})

	w.WriteHeader(http.StatusOK)
}

//GetDataVersion gets the sent cookie and return it
func (h AdminApisHandler) GetDataVersion(w http.ResponseWriter, r *http.Request) {
	versionCookie := getDataVersionCookie(r)
	if versionCookie == nil {
		log.Println("Error getting data version cookie")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(*versionCookie))
}

//GetConfig gets the config
func (h AdminApisHandler) GetConfig(w http.ResponseWriter, r *http.Request) {
	config, err := h.app.Administration.GetConfig()
	if err != nil {
		log.Println("Error getting the config")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	fmt.Fprint(w, config.Flag1)
}

//GetFullUIContent gives the full ui content (no rules applied)
func (h AdminApisHandler) GetFullUIContent(w http.ResponseWriter, r *http.Request) {
	uiContent := h.app.Administration.GetFullUIContent()
	data, err := json.Marshal(uiContent)
	if err != nil {
		log.Println("Error on marshal the full ui flat data")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

//ReloadUIContent reloads the ui content data from the storage
func (h AdminApisHandler) ReloadUIContent(w http.ResponseWriter, r *http.Request) {
	err := h.app.Administration.ReloadUIContent()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Successfully reloaded"))
}

//GetContentItems gets all content items
func (h AdminApisHandler) GetContentItems(w http.ResponseWriter, r *http.Request) {
	versionCookie := getDataVersionCookie(r)
	if versionCookie == nil {
		log.Println("Version cookie error")
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	contentItems, err := h.app.Administration.GetContentItems(*versionCookie)
	if err != nil {
		log.Println("Error on getting the content items")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	data, err := json.Marshal(contentItems)
	if err != nil {
		log.Println("Error on marshal the content items")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

//GetContentItem gets content item by id
func (h AdminApisHandler) GetContentItem(w http.ResponseWriter, r *http.Request) {
	versionCookie := getDataVersionCookie(r)
	if versionCookie == nil {
		log.Println("Version cookie error")
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	params := mux.Vars(r)
	ID := params["id"]
	if len(ID) <= 0 {
		log.Println("Content item id is required")
		http.Error(w, "Content item id is required", http.StatusBadRequest)
		return
	}
	numberID, err := strconv.Atoi(ID)
	if err != nil {
		log.Println("The id must be number")
		http.Error(w, "The id must be number", http.StatusBadRequest)
		return
	}
	contentItem, err := h.app.Administration.GetContentItem(*versionCookie, numberID)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data, err := json.Marshal(contentItem)
	if err != nil {
		log.Println("Error on marshal the content item when get")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

//CreateContentItem creates a content item
func (h AdminApisHandler) CreateContentItem(w http.ResponseWriter, r *http.Request) {
	versionCookie := getDataVersionCookie(r)
	if versionCookie == nil {
		log.Println("Version cookie error")
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error on marshal the create content item - %s\n", err.Error())
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	var requestData createContentItem
	err = json.Unmarshal(data, &requestData)
	if err != nil {
		log.Printf("Error on unmarshal the create content item request data - %s\n", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	name := requestData.Name
	if len(name) == 0 {
		http.Error(w, "Name cannot be empty", http.StatusBadRequest)
		return
	}

	contentItem, err := h.app.Administration.CreateContentItem(*versionCookie, name)
	if err != nil {
		log.Println("Error on creating the content item")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	data, err = json.Marshal(contentItem)
	if err != nil {
		log.Println("Error on marshal the content item")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

//UpdateContentItem updates a content item
func (h AdminApisHandler) UpdateContentItem(w http.ResponseWriter, r *http.Request) {
	versionCookie := getDataVersionCookie(r)
	if versionCookie == nil {
		log.Println("Version cookie error")
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	params := mux.Vars(r)
	ID := params["id"]
	if len(ID) <= 0 {
		log.Println("Content item id is required")
		http.Error(w, "Content item id is required", http.StatusBadRequest)
		return
	}
	numberID, err := strconv.Atoi(ID)
	if err != nil {
		log.Println("The id must be number")
		http.Error(w, "The id must be number", http.StatusBadRequest)
		return
	}
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error on marshal the update content item - %s\n", err.Error())
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	var requestData updateContentItem
	err = json.Unmarshal(data, &requestData)
	if err != nil {
		log.Printf("Error on unmarshal the update content item request data - %s\n", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	name := requestData.Name
	if len(name) == 0 {
		http.Error(w, "Name cannot be empty", http.StatusBadRequest)
		return
	}

	contentItem, err := h.app.Administration.UpdateContentItem(*versionCookie, numberID, name)
	if err != nil {
		log.Println("Error on updating the content item")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	data, err = json.Marshal(contentItem)
	if err != nil {
		log.Println("Error on marshal the content item")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

//DeleteContentItem deletes a content item
func (h AdminApisHandler) DeleteContentItem(w http.ResponseWriter, r *http.Request) {
	versionCookie := getDataVersionCookie(r)
	if versionCookie == nil {
		log.Println("Version cookie error")
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	params := mux.Vars(r)
	ID := params["id"]
	if len(ID) <= 0 {
		log.Println("Content item id is required")
		http.Error(w, "Content item id is required", http.StatusBadRequest)
		return
	}
	numberID, err := strconv.Atoi(ID)
	if err != nil {
		log.Println("The id must be number")
		http.Error(w, "The id must be number", http.StatusBadRequest)
		return
	}
	err = h.app.Administration.DeleteContentItem(*versionCookie, numberID)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Successfully deleted an item"))
}

//GetUIItem gets ui item for a specific content item
func (h AdminApisHandler) GetUIItem(w http.ResponseWriter, r *http.Request) {
	versionCookie := getDataVersionCookie(r)
	if versionCookie == nil {
		log.Println("Version cookie error")
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	params := mux.Vars(r)
	contentItemID := params["content-item-id"]
	ID := params["id"]
	if len(contentItemID) <= 0 || len(ID) <= 0 {
		log.Println("Content item id and id are required")
		http.Error(w, "Content item id and id are required", http.StatusBadRequest)
		return
	}
	contentItemNumberID, err := strconv.Atoi(contentItemID)
	if err != nil {
		log.Println("The content item id must be number")
		http.Error(w, "The content item id must be number", http.StatusBadRequest)
		return
	}
	numberID, err := strconv.Atoi(ID)
	if err != nil {
		log.Println("The id must be number")
		http.Error(w, "The id must be number", http.StatusBadRequest)
		return
	}

	contentItem, err := h.app.Administration.GetUIItem(*versionCookie, contentItemNumberID, numberID)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data, err := json.Marshal(contentItem)
	if err != nil {
		log.Println("Error on marshal the ui item when get")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

//CreateUIItem creates ui item for a specific content item
func (h AdminApisHandler) CreateUIItem(w http.ResponseWriter, r *http.Request) {
	versionCookie := getDataVersionCookie(r)
	if versionCookie == nil {
		log.Println("Version cookie error")
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	params := mux.Vars(r)
	contentItemID := params["content-item-id"]
	if len(contentItemID) <= 0 {
		log.Println("Content item id is required")
		http.Error(w, "Content item id is required", http.StatusBadRequest)
		return
	}
	contentItemNumberID, err := strconv.Atoi(contentItemID)
	if err != nil {
		log.Println("The content item id must be number")
		http.Error(w, "The content item id must be number", http.StatusBadRequest)
		return
	}
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error on marshal the create ui item - %s\n", err.Error())
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	var requestData createUIItem
	err = json.Unmarshal(data, &requestData)
	if err != nil {
		log.Printf("Error on unmarshal the create ui item request data - %s\n", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	name := requestData.Name
	if len(name) == 0 {
		http.Error(w, "Name cannot be empty", http.StatusBadRequest)
		return
	}
	order := requestData.Order
	if order < 1 {
		http.Error(w, "Order must be positive", http.StatusBadRequest)
		return
	}

	uiItem, err := h.app.Administration.CreateUIItem(*versionCookie, contentItemNumberID, name, order)
	if err != nil {
		log.Println("Error on creating the ui item")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	data, err = json.Marshal(uiItem)
	if err != nil {
		log.Println("Error on marshal the ui item")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

//UpdateUIItem updates ui item for a specific content item
func (h AdminApisHandler) UpdateUIItem(w http.ResponseWriter, r *http.Request) {
	versionCookie := getDataVersionCookie(r)
	if versionCookie == nil {
		log.Println("Version cookie error")
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	params := mux.Vars(r)
	contentItemID := params["content-item-id"]
	ID := params["id"]
	if len(contentItemID) <= 0 || len(ID) <= 0 {
		log.Println("Content item id and id are required")
		http.Error(w, "Content item id and id are required", http.StatusBadRequest)
		return
	}
	contentItemNumberID, err := strconv.Atoi(contentItemID)
	if err != nil {
		log.Println("The content item id must be number")
		http.Error(w, "The content item id must be number", http.StatusBadRequest)
		return
	}
	numberID, err := strconv.Atoi(ID)
	if err != nil {
		log.Println("The id must be number")
		http.Error(w, "The id must be number", http.StatusBadRequest)
		return
	}
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error on marshal the update ui item - %s\n", err.Error())
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	var requestData updateUIItem
	err = json.Unmarshal(data, &requestData)
	if err != nil {
		log.Printf("Error on unmarshal the update ui item request data - %s\n", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	name := requestData.Name
	if len(name) == 0 {
		http.Error(w, "Name cannot be empty", http.StatusBadRequest)
		return
	}
	order := requestData.Order
	if order < 1 {
		http.Error(w, "Name must be positive", http.StatusBadRequest)
		return
	}

	uiItem, err := h.app.Administration.UpdateUIItem(*versionCookie, contentItemNumberID, numberID, name, order)
	if err != nil {
		log.Printf("Error on updating the ui item %s", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	data, err = json.Marshal(uiItem)
	if err != nil {
		log.Println("Error on marshal the ui item")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

//DeleteUIItem deletes ui item for a specific content item
func (h AdminApisHandler) DeleteUIItem(w http.ResponseWriter, r *http.Request) {
	versionCookie := getDataVersionCookie(r)
	if versionCookie == nil {
		log.Println("Version cookie error")
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	params := mux.Vars(r)
	contentItemID := params["content-item-id"]
	ID := params["id"]
	if len(contentItemID) <= 0 || len(ID) <= 0 {
		log.Println("Content item id and id are required")
		http.Error(w, "Content item id and id are required", http.StatusBadRequest)
		return
	}
	contentItemNumberID, err := strconv.Atoi(contentItemID)
	if err != nil {
		log.Println("The content item id must be number")
		http.Error(w, "The content item id must be number", http.StatusBadRequest)
		return
	}
	numberID, err := strconv.Atoi(ID)
	if err != nil {
		log.Println("The id must be number")
		http.Error(w, "The id must be number", http.StatusBadRequest)
		return
	}

	err = h.app.Administration.DeleteUIItem(*versionCookie, contentItemNumberID, numberID)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Successfully deleted an item"))
}

//GetRule gets a rule for a specific ui item
func (h AdminApisHandler) GetRule(w http.ResponseWriter, r *http.Request) {
	versionCookie := getDataVersionCookie(r)
	if versionCookie == nil {
		log.Println("Version cookie error")
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	params := mux.Vars(r)
	uiItemID := params["ui-item-id"]
	ID := params["id"]
	if len(uiItemID) <= 0 || len(ID) <= 0 {
		log.Println("UI item id and id are required")
		http.Error(w, "UI item id and id are required", http.StatusBadRequest)
		return
	}
	uiItemNumberID, err := strconv.Atoi(uiItemID)
	if err != nil {
		log.Println("The ui item id must be number")
		http.Error(w, "The ui item id must be number", http.StatusBadRequest)
		return
	}
	numberID, err := strconv.Atoi(ID)
	if err != nil {
		log.Println("The id must be number")
		http.Error(w, "The id must be number", http.StatusBadRequest)
		return
	}

	rule, err := h.app.Administration.GetRule(*versionCookie, uiItemNumberID, numberID)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data, err := json.Marshal(rule)
	if err != nil {
		log.Println("Error on marshal the rule item when get")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

//CreateRule creates a rule for a specific ui item
func (h AdminApisHandler) CreateRule(w http.ResponseWriter, r *http.Request) {
	versionCookie := getDataVersionCookie(r)
	if versionCookie == nil {
		log.Println("Version cookie error")
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	params := mux.Vars(r)
	uiItemID := params["ui-item-id"]
	if len(uiItemID) <= 0 {
		log.Println("ui item id is required")
		http.Error(w, "ui item id is required", http.StatusBadRequest)
		return
	}
	uiItemNumberID, err := strconv.Atoi(uiItemID)
	if err != nil {
		log.Println("The ui item id must be number")
		http.Error(w, "The ui item id must be number", http.StatusBadRequest)
		return
	}
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error on marshal the create rule item - %s\n", err.Error())
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	var requestData createRule
	err = json.Unmarshal(data, &requestData)
	if err != nil {
		log.Printf("Error on unmarshal the create rule item request data - %s\n", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	ruleTypeID := requestData.RuleTypeID
	if ruleTypeID <= 0 {
		http.Error(w, "Rule type should be possitive", http.StatusBadRequest)
		return
	}
	value := requestData.Value
	if value == nil {
		http.Error(w, "Value should not be empty", http.StatusBadRequest)
		return
	}

	rule, err := h.app.Administration.CreateRule(*versionCookie, uiItemNumberID, ruleTypeID, value)
	if err != nil {
		log.Println("Error on creating the rule item")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data, err = json.Marshal(rule)
	if err != nil {
		log.Println("Error on marshal the rule item")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

//UpdateRule updates a rule for a specific ui item
func (h AdminApisHandler) UpdateRule(w http.ResponseWriter, r *http.Request) {
	versionCookie := getDataVersionCookie(r)
	if versionCookie == nil {
		log.Println("Version cookie error")
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	params := mux.Vars(r)
	uiItemID := params["ui-item-id"]
	ID := params["id"]
	if len(uiItemID) <= 0 || len(ID) <= 0 {
		log.Println("ui item id and id are required")
		http.Error(w, "ui item id and id are required", http.StatusBadRequest)
		return
	}
	uiItemNumberID, err := strconv.Atoi(uiItemID)
	if err != nil {
		log.Println("The ui item id must be number")
		http.Error(w, "The ui item id must be number", http.StatusBadRequest)
		return
	}
	numberID, err := strconv.Atoi(ID)
	if err != nil {
		log.Println("The id must be number")
		http.Error(w, "The id must be number", http.StatusBadRequest)
		return
	}
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error on marshal the update rule item - %s\n", err.Error())
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	var requestData updateRule
	err = json.Unmarshal(data, &requestData)
	if err != nil {
		log.Printf("Error on unmarshal the create rule item request data - %s\n", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	ruleTypeID := requestData.RuleTypeID
	if ruleTypeID <= 0 {
		http.Error(w, "Rule type should be possitive", http.StatusBadRequest)
		return
	}
	value := requestData.Value
	if value == nil {
		http.Error(w, "Value should not be empty", http.StatusBadRequest)
		return
	}

	rule, err := h.app.Administration.UpdateRule(*versionCookie, numberID, uiItemNumberID, ruleTypeID, value)
	if err != nil {
		log.Println("Error on updating the rule item")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data, err = json.Marshal(rule)
	if err != nil {
		log.Println("Error on marshal the rule item")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

//DeleteRule deletes a rule for a specific ui item
func (h AdminApisHandler) DeleteRule(w http.ResponseWriter, r *http.Request) {
	versionCookie := getDataVersionCookie(r)
	if versionCookie == nil {
		log.Println("Version cookie error")
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	params := mux.Vars(r)
	uiItemID := params["ui-item-id"]
	ID := params["id"]
	if len(uiItemID) <= 0 || len(ID) <= 0 {
		log.Println("UI item id and id are required")
		http.Error(w, "UI item id and id are required", http.StatusBadRequest)
		return
	}
	uiItemNumberID, err := strconv.Atoi(uiItemID)
	if err != nil {
		log.Println("The ui item id must be number")
		http.Error(w, "The ui item id must be number", http.StatusBadRequest)
		return
	}
	numberID, err := strconv.Atoi(ID)
	if err != nil {
		log.Println("The id must be number")
		http.Error(w, "The id must be number", http.StatusBadRequest)
		return
	}

	err = h.app.Administration.DeleteRule(*versionCookie, uiItemNumberID, numberID)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Successfully deleted an item"))
}

//GetRuleTypes gets the rule types
func (h AdminApisHandler) GetRuleTypes(w http.ResponseWriter, r *http.Request) {
	versionCookie := getDataVersionCookie(r)
	if versionCookie == nil {
		log.Println("Version cookie error")
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	ruleTypes, err := h.app.Administration.GetRuleTypes(*versionCookie)
	if err != nil {
		log.Println("Error on getting the rule types")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	data, err := json.Marshal(ruleTypes)
	if err != nil {
		log.Println("Error on marshal the rule types")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

//NewAdminApisHandler creates new admin rest Handler instance
func NewAdminApisHandler(app *core.Application) AdminApisHandler {
	return AdminApisHandler{app: app}
}

func getDataVersionCookie(r *http.Request) *string {
	versionCookie, err := r.Cookie("tch-data-version")
	if versionCookie == nil || err != nil {
		return nil
	}
	return &versionCookie.Value
}
