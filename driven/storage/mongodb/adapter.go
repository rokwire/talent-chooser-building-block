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

package mongodb

import (
	"encoding/json"
	"errors"
	"log"
	"strconv"
	"sync"
	"talent-chooser/core"
	"talent-chooser/core/model"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

//TODO refactor implementation...

//TODO
type dataItem struct {
	Version string `bson:"version"`
	Data    string `bson:"data"`
}

type data struct {
	LastUpdated   string `json:"last_updated"`
	LastUpdatedBy string `json:"last_updated_by"`

	ContentItems        []contentItem       `json:"content_items"`
	ContentItemsUIItems []contentItemUIItem `json:"content_items_ui_items"`
	RuleTypes           []ruleType          `json:"rule_types"`
	Rules               []rule              `json:"rules"`
	RulesUIItems        []ruleUIItem        `json:"rules_ui_items"`
	UIItems             []uiItem            `json:"ui_items"`
}

type storageItem interface {
	GetID() int
}

type contentItem struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func (ca contentItem) GetID() int {
	return ca.ID
}

type uiItem struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Order int    `json:"order"`
}

func (ua uiItem) GetID() int {
	return ua.ID
}

type contentItemUIItem struct {
	ID            int `json:"id"`
	ContentItemID int `json:"content_item_id"`
	UIItemID      int `json:"ui_item_id"`
}

func (caua contentItemUIItem) GetID() int {
	return caua.ID
}

type ruleType struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type rule struct {
	ID         int         `json:"id"`
	RuleTypeID int         `json:"rule_type_id"`
	Value      interface{} `json:"value"`
}

func (r rule) GetID() int {
	return r.ID
}

type ruleUIItem struct {
	ID       int `json:"id"`
	UIItemID int `json:"ui_item_id"`
	RuleID   int `json:"rule_id"`
}

func (r ruleUIItem) GetID() int {
	return r.ID
}

//Adapter implements the Storage interface
type Adapter struct {
	db *database

	mu *sync.Mutex //TODO
}

//Start starts the storage
func (a *Adapter) Start() error {
	err := a.db.start()
	return err
}

//SetStorageListener sets listener for the storage
func (a *Adapter) SetStorageListener(storageListener core.StorageListener) {
	a.db.listener = storageListener
}

//ReadConfig reads the configuration from the storage
func (a *Adapter) ReadConfig() (model.Config, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	return model.Config{Flag1: true}, nil
}

//ReadUIContent reads the UI content from the storage
func (a *Adapter) ReadUIContent() (map[string]*model.UIContent, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	uiContents, err := a.readFullUIContent()
	if err != nil {
		log.Printf("ReadUIContent -> Error reading the content items %s\n", err.Error())
		return nil, err
	}

	result := map[string]*model.UIContent{}
	for version, item := range uiContents {
		uiContent := model.NewUIContent(item)
		result[version] = &uiContent
	}
	return result, nil
}

//ReadContentItems reads the content items from the storage
func (a *Adapter) ReadContentItems(dataVersion string) ([]model.ContentItem, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	contentItems, err := a.readContentItems(dataVersion)
	if err != nil {
		log.Printf("ReadContentItems -> Error reading the content items %s\n", err.Error())
		return nil, err
	}

	return contentItems, nil
}

//ReadContentItem reads a content item from the storage
func (a *Adapter) ReadContentItem(dataVersion string, ID int) (*model.ContentItem, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	data, err := a.readData(dataVersion)
	if err != nil {
		log.Print(err.Error())
		return nil, err
	}
	if data == nil {
		log.Println("ReadContentItem - data is nil")
		return nil, errors.New("ReadContentItem - data is nil")
	}

	uiItemsList := data.UIItems
	contentItemsList := data.ContentItems
	contentItemsUIItemsList := data.ContentItemsUIItems

	for _, contentItem := range contentItemsList {
		if ID == contentItem.ID {
			name := contentItem.Name
			ciuiItems := a.getContentItemUIItems(ID, contentItemsUIItemsList)

			//add ui items
			uiItems := make([]model.UIItem, len(ciuiItems))
			if ciuiItems != nil {
				for index, ciuiItem := range ciuiItems {
					uiItem, _ := a.findUIItem(ciuiItem.UIItemID, uiItemsList)
					uiItems[index] = model.UIItem{ID: uiItem.ID, Name: uiItem.Name, Order: uiItem.Order, Rules: nil}
				}
			}
			return &model.ContentItem{ID: ID, Name: name, UIItems: uiItems}, nil
		}
	}
	return nil, errors.New("There is no a content item with the provided id")
}

//CreateContentItem creates a content item
func (a *Adapter) CreateContentItem(dataVersion string, name string) (*model.ContentItem, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	data, err := a.readData(dataVersion)
	if err != nil {
		log.Print(err.Error())
		return nil, err
	}
	if data == nil {
		log.Println("CreateContentItem - data is nil")
		return nil, errors.New("CreateContentItem - data is nil")
	}

	//1. get content items
	contentItemsList := data.ContentItems

	//2. find biggest id
	storageItems := make([]storageItem, len(contentItemsList))
	for index, item := range contentItemsList {
		storageItems[index] = item
	}
	biggestID, err := a.findBiggestID(storageItems)
	if err != nil {
		log.Println(err.Error())
		return nil, err
	}

	//3. create a new item
	newItem := contentItem{ID: biggestID + 1, Name: name}
	//4. add it to the list
	contentItemsList = append(contentItemsList, newItem)
	//5. write the list
	data.ContentItems = contentItemsList
	err = a.saveData(dataVersion, data)
	if err != nil {
		return nil, err
	}
	//6. return the new created content item
	return &model.ContentItem{ID: newItem.ID, Name: newItem.Name}, nil
}

//UpdateContentItem updates the content item
func (a *Adapter) UpdateContentItem(dataVersion string, ID int, name string) (*model.ContentItem, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	data, err := a.readData(dataVersion)
	if err != nil {
		log.Print(err.Error())
		return nil, err
	}
	if data == nil {
		log.Println("UpdateContentItem - data is nil")
		return nil, errors.New("UpdateContentItem - data is nil")
	}

	//1. get content items
	contentItemsList := data.ContentItems

	//2. find the item
	founded, index := a.findContentItem(ID, contentItemsList)
	if founded == nil {
		return nil, errors.New("there is no an item with the provided id")
	}

	//3. update the item
	founded.Name = name

	//4. replace the updated item in the list
	contentItemsList[index] = *founded

	//5. write the list
	data.ContentItems = contentItemsList
	err = a.saveData(dataVersion, data)
	if err != nil {
		return nil, err
	}

	return &model.ContentItem{ID: founded.ID, Name: founded.Name}, nil
}

//DeleteContentItem deletes the content item
func (a *Adapter) DeleteContentItem(dataVersion string, ID int) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	data, err := a.readData(dataVersion)
	if err != nil {
		log.Print(err.Error())
		return err
	}
	if data == nil {
		log.Println("DeleteContentItem - data is nil")
		return errors.New("DeleteContentItem - data is nil")
	}

	//1. check if there is ui items associated with this content item. We should allow deleting in this case.
	contentItemsUIItemsList := data.ContentItemsUIItems
	hasUIItems := a.hasUIItems(ID, contentItemsUIItemsList)
	if hasUIItems {
		return errors.New("cannot be deleted because there is associated ui items with it")
	}

	//2. get content items
	contentItemsList := data.ContentItems

	//3. find the item
	founded, index := a.findContentItem(ID, contentItemsList)
	if founded == nil {
		return errors.New("there is no an item with the provided id")
	}

	//4. remove it from the list
	contentItemsList = append(contentItemsList[:index], contentItemsList[index+1:]...)

	//5. write the list
	data.ContentItems = contentItemsList
	err = a.saveData(dataVersion, data)
	if err != nil {
		return err
	}

	return nil
}

//ReadUIItem gets ui item for a specific content item
func (a *Adapter) ReadUIItem(dataVersion string, contentItemID int, ID int) (*model.UIItem, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	data, err := a.readData(dataVersion)
	if err != nil {
		log.Print(err.Error())
		return nil, err
	}
	if data == nil {
		log.Println("ReadUIItem - data is nil")
		return nil, errors.New("ReadUIItem - data is nil")
	}

	//1. check if thre is a content item with the provided id
	contentItemsList := data.ContentItems
	contentItem, _ := a.findContentItem(contentItemID, contentItemsList)
	if contentItem == nil {
		return nil, errors.New("there is no a content item with the provided id")
	}

	//2. check if thre is an ui item with the provided id
	uiItemsList := data.UIItems
	uiItem, _ := a.findUIItem(ID, uiItemsList)
	if uiItem == nil {
		return nil, errors.New("there is no a ui item with the provided id")
	}

	//3. read rules
	ruleTypesList := data.RuleTypes
	rulesList := data.Rules
	rulesUIItems := data.RulesUIItems

	rules := a.getRules(uiItem.ID, rulesList, ruleTypesList, rulesUIItems)

	return &model.UIItem{ID: uiItem.ID, Name: uiItem.Name, Order: uiItem.Order, Rules: rules}, nil
}

//CreateUIItem create ui item for a specific content item
func (a *Adapter) CreateUIItem(dataVersion string, contentItemID int, name string, order int) (*model.UIItem, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	data, err := a.readData(dataVersion)
	if err != nil {
		log.Print(err.Error())
		return nil, err
	}
	if data == nil {
		log.Println("CreateUIItem - data is nil")
		return nil, errors.New("CreateUIItem - data is nil")
	}

	//1. check if thre is a content item with the provided id
	contentItemsList := data.ContentItems
	contentItem, _ := a.findContentItem(contentItemID, contentItemsList)
	if contentItem == nil {
		return nil, errors.New("there is no a content item with the provided id")
	}

	//2. download the ui items and content items ui items files
	uiItemsList := data.UIItems
	contentItemsUIItemsList := data.ContentItemsUIItems

	//3. add the new ui item in the ui items list
	uiStorageItems := make([]storageItem, len(uiItemsList))
	for index, item := range uiItemsList {
		uiStorageItems[index] = item
	}
	uiItemBiggestID, err := a.findBiggestID(uiStorageItems)
	if err != nil {
		log.Println(err.Error())
		return nil, err
	}
	uiItemID := uiItemBiggestID + 1
	newItem := uiItem{ID: uiItemID, Name: name, Order: order}
	uiItemsList = append(uiItemsList, newItem)

	//4. add a record in the relation file
	relStorageItems := make([]storageItem, len(contentItemsUIItemsList))
	for index, item := range contentItemsUIItemsList {
		relStorageItems[index] = item
	}
	relBiggestID, err := a.findBiggestID(relStorageItems)
	if err != nil {
		log.Println(err.Error())
		return nil, err
	}
	relItem := contentItemUIItem{ID: relBiggestID + 1, ContentItemID: contentItemID, UIItemID: uiItemID}
	contentItemsUIItemsList = append(contentItemsUIItemsList, relItem)

	//5. upload the files
	data.UIItems = uiItemsList
	data.ContentItemsUIItems = contentItemsUIItemsList
	err = a.saveData(dataVersion, data)
	if err != nil {
		return nil, err
	}

	return &model.UIItem{ID: newItem.ID, Name: newItem.Name, Order: newItem.Order}, nil
}

//UpdateUIItem updates ui item for a specific content item
func (a *Adapter) UpdateUIItem(dataVersion string, contentItemID int, ID int, name string, order int) (*model.UIItem, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	data, err := a.readData(dataVersion)
	if err != nil {
		log.Print(err.Error())
		return nil, err
	}
	if data == nil {
		log.Println("UpdateUIItem - data is nil")
		return nil, errors.New("UpdateUIItem - data is nil")
	}

	//1. check if thre is a content item with the provided id
	contentItemsList := data.ContentItems
	contentItem, _ := a.findContentItem(contentItemID, contentItemsList)
	if contentItem == nil {
		return nil, errors.New("there is no a content item with the provided id")
	}

	//2. download the content items ui items files
	contentItemsUIItemsList := data.ContentItemsUIItems
	uiItemsList := data.UIItems

	//3. find if there is a such ui item for the provided content item
	foundedRelItem, _ := a.findContentItemUIItemRel(contentItemID, ID, contentItemsUIItemsList)
	if foundedRelItem == nil {
		return nil, errors.New("there is no associated ui item with the provided content item id")
	}
	//4. find if there is ui item for the provided id
	foundedUIItem, uiItemIndex := a.findUIItem(ID, uiItemsList)
	if foundedUIItem == nil {
		return nil, errors.New("there is no ui item for the provided id")
	}

	//5. update the item
	foundedUIItem.Name = name
	foundedUIItem.Order = order

	//6. replace the updated item in the list
	uiItemsList[uiItemIndex] = *foundedUIItem

	//7. write the list
	data.UIItems = uiItemsList
	err = a.saveData(dataVersion, data)
	if err != nil {
		return nil, err
	}

	return &model.UIItem{ID: foundedUIItem.ID, Name: foundedUIItem.Name, Order: foundedUIItem.Order}, nil
}

//DeleteUIItem deltes ui item for a specific content item
func (a *Adapter) DeleteUIItem(dataVersion string, contentItemID int, ID int) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	data, err := a.readData(dataVersion)
	if err != nil {
		log.Print(err.Error())
		return err
	}
	if data == nil {
		log.Println("CreateRule - data is nil")
		return errors.New("CreateRule - data is nil")
	}

	//1. check if thre is a content item with the provided id
	contentItemsList := data.ContentItems
	contentItem, _ := a.findContentItem(contentItemID, contentItemsList)
	if contentItem == nil {
		return errors.New("there is no a content item with the provided id")
	}

	//2. download the content items ui items files
	contentItemsUIItemsList := data.ContentItemsUIItems
	uiItemsList := data.UIItems

	//3. find if there is a such ui item for the provided content item
	foundedRelItem, relIndex := a.findContentItemUIItemRel(contentItemID, ID, contentItemsUIItemsList)
	if foundedRelItem == nil {
		return errors.New("there is no associated ui item with the provided content item id")
	}
	//4. find if there is ui item for the provided id
	foundedUIItem, uiItemIndex := a.findUIItem(ID, uiItemsList)
	if foundedUIItem == nil {
		return errors.New("there is no ui item for the provided id")
	}

	//5. check if we can delete it from the ui items file and the rel files
	ruiList := data.RulesUIItems
	canDelete, reason := a.canDeleteUIItem(contentItemID, ID, contentItemsUIItemsList, ruiList)
	if !canDelete {
		return errors.New(reason)
	}

	//6. remove it from the ui items and the rel files
	contentItemsUIItemsList = append(contentItemsUIItemsList[:relIndex], contentItemsUIItemsList[relIndex+1:]...)
	uiItemsList = append(uiItemsList[:uiItemIndex], uiItemsList[uiItemIndex+1:]...)

	//7. upload the files
	data.UIItems = uiItemsList
	data.ContentItemsUIItems = contentItemsUIItemsList
	err = a.saveData(dataVersion, data)
	if err != nil {
		return err
	}
	return nil
}

//ReadRule reads a rule for a specific ui item
func (a *Adapter) ReadRule(dataVersion string, uiItemID int, ID int) (*model.Rule, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	data, err := a.readData(dataVersion)
	if err != nil {
		log.Print(err.Error())
		return nil, err
	}
	if data == nil {
		log.Println("ReadRule - data is nil")
		return nil, errors.New("ReadRule - data is nil")
	}

	//1. check if there is a ui item with the provided id
	uiItemsList := data.UIItems
	uiItem, _ := a.findUIItem(uiItemID, uiItemsList)
	if uiItem == nil {
		return nil, errors.New("there is no a ui item with the provided id")
	}

	//2. check if there is a rule with the provided id
	rulesList := data.Rules
	rule, _ := a.findRule(ID, rulesList)
	if rule == nil {
		return nil, errors.New("there is no a rule with the provided id")
	}

	//3. read rule types
	ruleTypesList := data.RuleTypes
	rType := a.findRuleType(rule.RuleTypeID, ruleTypesList)
	ruleType := *model.NewRuleType(rType.ID, rType.Name)

	return &model.Rule{ID: rule.ID, RuleType: ruleType, Value: rule.Value}, nil
}

//CreateRule creates a rule for a specific ui item
func (a *Adapter) CreateRule(dataVersion string, uiItemID int, ruleTypeID int, value interface{}) (*model.Rule, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	data, err := a.readData(dataVersion)
	if err != nil {
		log.Print(err.Error())
		return nil, err
	}
	if data == nil {
		log.Println("CreateRule - data is nil")
		return nil, errors.New("CreateRule - data is nil")
	}

	//1. first check if there is ui item for the provided ui item id
	uiItemsList := data.UIItems
	uiItem, _ := a.findUIItem(uiItemID, uiItemsList)
	if uiItem == nil {
		return nil, errors.New("there is no a ui item with the provided id")
	}

	//2. check if there is rule type for the provided rule type id
	ruleTypesList := data.RuleTypes
	rType := a.findRuleType(ruleTypeID, ruleTypesList)
	if rType == nil {
		return nil, errors.New("there is no a rule type with the provided id")
	}

	//3. Validate the value data for the rule type
	ruleType := *model.NewRuleType(rType.ID, rType.Name)
	valid := ruleType.ValidData(value)
	if !valid {
		return nil, errors.New("the provided data is not valid for this rule type")
	}

	//4. Add a record in the rules
	rulesList := data.Rules
	if err != nil {
		log.Println(err.Error())
		return nil, err
	}
	uiStorageItems := make([]storageItem, len(rulesList))
	for index, item := range rulesList {
		uiStorageItems[index] = item
	}
	rulesListBiggestID, err := a.findBiggestID(uiStorageItems)
	if err != nil {
		log.Println(err.Error())
		return nil, err
	}
	ruleID := rulesListBiggestID + 1
	newRule := rule{ID: ruleID, RuleTypeID: ruleTypeID, Value: value}
	rulesList = append(rulesList, newRule)

	//5. Add a record in the relations file
	rulesUIItemsList := data.RulesUIItems
	relStorageItems := make([]storageItem, len(rulesUIItemsList))
	for index, item := range rulesUIItemsList {
		relStorageItems[index] = item
	}
	relBiggestID, err := a.findBiggestID(relStorageItems)
	if err != nil {
		log.Println(err.Error())
		return nil, err
	}
	relItem := ruleUIItem{ID: relBiggestID + 1, UIItemID: uiItemID, RuleID: ruleID}
	rulesUIItemsList = append(rulesUIItemsList, relItem)

	//6. upload the files
	data.Rules = rulesList
	data.RulesUIItems = rulesUIItemsList

	err = a.saveData(dataVersion, data)
	if err != nil {
		return nil, err
	}

	rule := model.Rule{ID: ruleID, RuleType: ruleType, Value: value}
	return &rule, nil
}

//UpdateRule creates a rule for a specific ui item
func (a *Adapter) UpdateRule(dataVersion string, ID int, uiItemID int, ruleTypeID int, value interface{}) (*model.Rule, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	data, err := a.readData(dataVersion)
	if err != nil {
		log.Print(err.Error())
		return nil, err
	}
	if data == nil {
		log.Println("UpdateRule - data is nil")
		return nil, errors.New("UpdateRule - data is nil")
	}

	//1. check if there is a ui item with the provided id
	uiItemsList := data.UIItems
	contentItem, _ := a.findUIItem(uiItemID, uiItemsList)
	if contentItem == nil {
		return nil, errors.New("there is no a ui item with the provided id")
	}

	//2. download the rules ui items files
	rulesUIItemsList := data.RulesUIItems
	rulesList := data.Rules

	//3. find if there is a such rule for the provided ui item
	foundedRelItem, _ := a.findRuleUIItemRel(ID, uiItemID, rulesUIItemsList)
	if foundedRelItem == nil {
		return nil, errors.New("there is no associated rule with the provided ui item id")
	}
	//4. find if there is rule for the provided id
	foundedRule, ruleIndex := a.findRule(ID, rulesList)
	if foundedRule == nil {
		return nil, errors.New("there is no rule for the provided id")
	}

	//5. check if there is rule type for the provided rule type id
	ruleTypesList := data.RuleTypes
	rType := a.findRuleType(ruleTypeID, ruleTypesList)
	if rType == nil {
		return nil, errors.New("there is no a rule type with the provided id")
	}

	//6. Validate the value data for the rule type
	ruleType := *model.NewRuleType(rType.ID, rType.Name)
	valid := ruleType.ValidData(value)
	if !valid {
		return nil, errors.New("the provided data is not valid for this rule type")
	}

	//7. update the item
	foundedRule.RuleTypeID = ruleTypeID
	foundedRule.Value = value

	//8. replace the updated item in the list
	rulesList[ruleIndex] = *foundedRule

	//9. write the list
	data.Rules = rulesList
	err = a.saveData(dataVersion, data)
	if err != nil {
		return nil, err
	}

	return &model.Rule{ID: ID, RuleType: ruleType, Value: value}, nil
}

//DeleteRule deletes a rule for a specific ui item
func (a *Adapter) DeleteRule(dataVersion string, uiItemID int, ID int) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	data, err := a.readData(dataVersion)
	if err != nil {
		log.Print(err.Error())
		return err
	}
	if data == nil {
		log.Println("DeleteRule - data is nil")
		return errors.New("DeleteRule - data is nil")
	}

	//1. check if thre is a ui item with the provided id
	uiItemsList := data.UIItems
	uiItem, _ := a.findUIItem(uiItemID, uiItemsList)
	if uiItem == nil {
		return errors.New("there is no a ui item with the provided id")
	}

	//2. download the rules ui items files
	rulesUIItemsList := data.RulesUIItems
	rulesList := data.Rules

	//3. find if there is a such rule for the provided ui item
	foundedRelItem, relIndex := a.findRuleUIItemRel(ID, uiItemID, rulesUIItemsList)
	if foundedRelItem == nil {
		return errors.New("there is no associated rule with the provided ui item id")
	}
	//4. find if there is rule for the provided id
	foundedRule, ruleIndex := a.findRule(ID, rulesList)
	if foundedRule == nil {
		return errors.New("there is no rule for the provided id")
	}

	//5. remove it from the rules and the rel files
	rulesUIItemsList = append(rulesUIItemsList[:relIndex], rulesUIItemsList[relIndex+1:]...)
	rulesList = append(rulesList[:ruleIndex], rulesList[ruleIndex+1:]...)

	//6. upload the files
	data.Rules = rulesList
	data.RulesUIItems = rulesUIItemsList

	err = a.saveData(dataVersion, data)
	if err != nil {
		return err
	}
	return nil
}

//ReadRuleTypes reads all rule types
func (a *Adapter) ReadRuleTypes(dataVersion string) ([]model.RuleType, error) {
	data, err := a.readData(dataVersion)
	if err != nil {
		log.Print(err.Error())
		return nil, err
	}
	if data == nil {
		log.Println("ReadRuleTypes - data is nil")
		return nil, errors.New("ReadRuleTypes - data is nil")
	}

	ruleTypesList := data.RuleTypes

	list := make([]model.RuleType, len(ruleTypesList))

	for i, ruleType := range ruleTypesList {
		ruleTypeEntity := model.NewRuleType(ruleType.ID, ruleType.Name)
		list[i] = *ruleTypeEntity
	}
	return list, nil
}

func (a *Adapter) canDeleteUIItem(contentItemID int, uiItemID int, ciuiList []contentItemUIItem, ruiList []ruleUIItem) (bool, string) {
	//1. check if there is associated rules
	for _, rui := range ruiList {
		if rui.UIItemID == uiItemID {
			return false, "there is associated rules with this ui item"
		}
	}

	//2. check if there is another associated content items except the one we need to delete to
	for _, ciui := range ciuiList {
		if ciui.UIItemID == uiItemID && ciui.ContentItemID != contentItemID {
			return false, "there is associated another content items with this ui item"
		}
	}
	return true, ""
}

func (a *Adapter) findBiggestID(list []storageItem) (int, error) {
	length := len(list)
	if length == 0 {
		return 1, nil
	}

	biggest := list[0].GetID()
	for index := 1; index < length; index++ {
		current := list[index].GetID()
		if current > biggest {
			biggest = current
		}
	}
	return biggest, nil
}

func (a *Adapter) readFullUIContent() (map[string][]model.ContentItem, error) {
	data, err := a.readFullData()
	if err != nil {
		log.Print(err.Error())
		return nil, err
	}
	if data == nil {
		log.Println("readFullUIContent - data is nil")
		return nil, errors.New("readFullUIContent - data is nil")
	}

	result := map[string][]model.ContentItem{}
	for version, item := range data {
		uiItemsList := item.UIItems
		contentItemsList := item.ContentItems
		contentItemsUIItemsList := item.ContentItemsUIItems
		ruleTypesList := item.RuleTypes
		rulesList := item.Rules
		rulesUIItems := item.RulesUIItems

		contentItems := make([]model.ContentItem, len(contentItemsList))

		for i, contentItem := range contentItemsList {
			id := contentItem.ID
			name := contentItem.Name
			ciuiItems := a.getContentItemUIItems(id, contentItemsUIItemsList)

			//add ui items
			uiItems := make([]model.UIItem, len(ciuiItems))
			if ciuiItems != nil {
				for index, ciuiItem := range ciuiItems {
					uiItem, _ := a.findUIItem(ciuiItem.UIItemID, uiItemsList)
					rules := a.getRules(uiItem.ID, rulesList, ruleTypesList, rulesUIItems)
					uiItems[index] = model.UIItem{ID: uiItem.ID, Name: uiItem.Name, Order: uiItem.Order, Rules: rules}
				}
			}
			contentItems[i] = model.ContentItem{ID: id, Name: name, UIItems: uiItems}
		}
		result[version] = contentItems
	}
	return result, nil
}

func (a *Adapter) readContentItems(dataVersion string) ([]model.ContentItem, error) {
	data, err := a.readData(dataVersion)
	if err != nil {
		log.Print(err.Error())
		return nil, err
	}
	if data == nil {
		log.Println("readContentItems - data is nil")
		return nil, errors.New("readContentItems - data is nil")
	}

	uiItemsList := data.UIItems
	contentItemsList := data.ContentItems
	contentItemsUIItemsList := data.ContentItemsUIItems

	contentItems := make([]model.ContentItem, len(contentItemsList))

	for i, contentItem := range contentItemsList {
		id := contentItem.ID
		name := contentItem.Name
		ciuiItems := a.getContentItemUIItems(id, contentItemsUIItemsList)

		//add ui items
		uiItems := make([]model.UIItem, len(ciuiItems))
		if ciuiItems != nil {
			for index, ciuiItem := range ciuiItems {
				uiItem, _ := a.findUIItem(ciuiItem.UIItemID, uiItemsList)
				uiItems[index] = model.UIItem{ID: uiItem.ID, Name: uiItem.Name, Order: uiItem.Order, Rules: nil}
			}
		}
		contentItems[i] = model.ContentItem{ID: id, Name: name, UIItems: uiItems}
	}
	return contentItems, nil
}

func (a *Adapter) getContentItemUIItems(contentItemID int, list []contentItemUIItem) []contentItemUIItem {
	var result []contentItemUIItem
	for _, item := range list {
		if contentItemID == item.ContentItemID {
			result = append(result, item)
		}
	}
	return result
}

func (a *Adapter) findUIItem(id int, uiItems []uiItem) (*uiItem, int) {
	for index, uiItem := range uiItems {
		if id == uiItem.ID {
			return &uiItem, index
		}
	}
	return nil, -1
}

func (a *Adapter) findContentItem(ID int, list []contentItem) (*contentItem, int) {
	for index, item := range list {
		if ID == item.ID {
			return &item, index
		}
	}
	return nil, -1
}

func (a *Adapter) hasUIItems(contentItemID int, list []contentItemUIItem) bool {
	for _, item := range list {
		if contentItemID == item.ContentItemID {
			return true
		}
	}
	return false
}

func (a *Adapter) findContentItemUIItemRel(contentItemID int, uiItem int, list []contentItemUIItem) (*contentItemUIItem, int) {
	for index, item := range list {
		if contentItemID == item.ContentItemID && uiItem == item.UIItemID {
			return &item, index
		}
	}
	return nil, -1
}

func (a *Adapter) findRuleUIItemRel(ruleID int, uiItemID int, list []ruleUIItem) (*ruleUIItem, int) {
	for index, item := range list {
		if ruleID == item.RuleID && uiItemID == item.UIItemID {
			return &item, index
		}
	}
	return nil, -1
}

func (a *Adapter) getRules(uiItemID int, rules []rule, rulesTypes []ruleType, rulesUIItems []ruleUIItem) *[]model.Rule {
	var rulesResult []model.Rule
	for _, ruleItemID := range rulesUIItems {
		if uiItemID == ruleItemID.UIItemID {
			rule, _ := a.findRule(ruleItemID.RuleID, rules)
			ruleType := a.findRuleType(rule.RuleTypeID, rulesTypes)

			ruleTypeEntity := *model.NewRuleType(ruleType.ID, ruleType.Name)
			ruleEntity := model.Rule{ID: rule.ID, RuleType: ruleTypeEntity, Value: rule.Value}
			rulesResult = append(rulesResult, ruleEntity)
		}
	}
	return &rulesResult
}

func (a *Adapter) findRule(id int, rules []rule) (*rule, int) {
	for index, rule := range rules {
		if id == rule.ID {
			return &rule, index
		}
	}
	return nil, -1
}

func (a *Adapter) findRuleType(id int, ruleTypes []ruleType) *ruleType {
	for _, ruleType := range ruleTypes {
		if id == ruleType.ID {
			return &ruleType
		}
	}
	return nil
}

func (a *Adapter) readData(dataVersion string) (*data, error) {
	filter := bson.D{primitive.E{Key: "version", Value: dataVersion}}
	var dataItems []*dataItem
	err := a.db.tchdata.Find(filter, &dataItems, nil)
	if err != nil {
		log.Printf("Cannot find data item for %s - %s\n", dataVersion, err)
		return nil, err
	}
	if dataItems == nil || len(dataItems) == 0 {
		return nil, errors.New("Cannot find data item for " + dataVersion)
	}
	dataItem := dataItems[0]

	var data data
	err = json.Unmarshal([]byte(dataItem.Data), &data)
	if err != nil {
		log.Printf("Cannot unmarshal the data %s\n", err)
		return nil, err
	}
	return &data, nil
}

func (a *Adapter) readFullData() (map[string]*data, error) {
	filter := bson.D{}
	var results []*dataItem
	err := a.db.tchdata.Find(filter, &results, nil)
	if err != nil {
		return nil, err
	}

	resultMap := map[string]*data{}

	for _, item := range results {
		var data data
		err = json.Unmarshal([]byte(item.Data), &data)
		if err != nil {
			log.Printf("Cannot unmarshal the data %s\n", err)
			return nil, err
		}
		resultMap[item.Version] = &data
	}
	return resultMap, nil
}

func (a *Adapter) saveData(dataVersion string, data *data) error {
	//1. find the item
	filter := bson.D{primitive.E{Key: "version", Value: dataVersion}}
	var dataItems []*dataItem
	err := a.db.tchdata.Find(filter, &dataItems, nil)
	if err != nil {
		log.Printf("Cannot find data item for %s - %s\n", dataVersion, err)
		return err
	}
	if dataItems == nil || len(dataItems) == 0 {
		return errors.New("Cannot find data item for " + dataVersion)
	}
	dataItem := dataItems[0]

	//2. prepare the data
	data.LastUpdated = time.Now().UTC().String()
	d, err := json.MarshalIndent(data, "", "    ")
	if err != nil {
		log.Printf("Cannot marshal the data %s\n", err)
		return err
	}

	//3. update the value
	dataItem.Data = string(d)

	//4. save it
	err = a.db.tchdata.ReplaceOne(filter, dataItem, nil)
	if err != nil {
		return err
	}
	return nil
}

//NewStorageAdapter creates a new storage adapter instance
func NewStorageAdapter(mongoDBAuth string, mongoDBName string, mongoTimeout string) *Adapter {
	timeout, err := strconv.Atoi(mongoTimeout)
	if err != nil {
		log.Println("Set default timeout - 500")
		timeout = 500
	}
	timeoutMS := time.Millisecond * time.Duration(timeout)

	db := &database{mongoDBAuth: mongoDBAuth, mongoDBName: mongoDBName, mongoTimeout: timeoutMS}
	mu := &sync.Mutex{}
	return &Adapter{db: db, mu: mu}
}
