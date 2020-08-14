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

package file

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"talent-chooser/core/model"
)

type contentItem struct {
	ID   int
	Name string
}

type uiItem struct {
	ID    int
	Name  string
	Order int
}

type contentItemUIItem struct {
	ID            int
	ContentItemID int `json:"content_item_id"`
	UIItemID      int `json:"ui_item_id"`
}

type ruleType struct {
	ID   int
	Name string
}

type rule struct {
	ID         int
	RuleTypeID int `json:"rule_type_id"`
	Value      interface{}
}

type ruleUIItem struct {
	ID       int
	UIItemID int `json:"ui_item_id"`
	RuleID   int `json:"rule_id"`
}

//Adapter implements the Storage interface as file system
type Adapter struct {
	version string
}

//Start starts the storage
func (fa Adapter) Start() error {
	return nil
}

//ReadConfig reads the configuration from the storage
func (fa Adapter) ReadConfig() (model.Config, error) {
	return model.Config{Flag1: true}, nil
}

//ReadUIContent reads the whole UI content from the storage
func (fa Adapter) ReadUIContent() (*model.UIContent, error) {
	uiItemsList := fa.readUIItemsList()
	contentItemsList := fa.readContentItemsList()
	contentItemsUIItemsList := fa.readContentItemsUIItemsList()
	ruleTypesList := fa.readRuleTypesList()
	rulesList := fa.readRulesList()
	rulesUIItems := fa.readRulesUIItemsList()

	uiData := make([]model.ContentItem, len(contentItemsList))

	for _, contentItem := range contentItemsList {
		id := contentItem.ID
		name := contentItem.Name
		ciuiItems := fa.getContentItemUIItems(id, contentItemsUIItemsList)

		//add ui items
		uiItems := make([]model.UIItem, len(ciuiItems))
		if ciuiItems != nil {
			for index, ciuiItem := range ciuiItems {
				uiItem := fa.findUIItem(ciuiItem.UIItemID, uiItemsList)
				rules := fa.getRules(uiItem.ID, rulesList, ruleTypesList, rulesUIItems)
				uiItems[index] = model.UIItem{ID: uiItem.ID, Name: uiItem.Name, Order: uiItem.Order, Rules: rules}
			}
		}

		uiData[id] = model.ContentItem{Name: name, UIItems: uiItems}
	}
	uiContent := model.NewUIContent(uiData)
	return &uiContent, nil
}

func (fa Adapter) getContentItemUIItems(contentItemID int, list []contentItemUIItem) []contentItemUIItem {
	var result []contentItemUIItem
	for _, item := range list {
		if contentItemID == item.ContentItemID {
			result = append(result, item)
		}
	}
	return result
}

func (fa Adapter) getRules(uiItemID int, rules []rule, rulesTypes []ruleType, rulesUIItems []ruleUIItem) *[]model.Rule {
	var rulesResult []model.Rule
	for _, ruleItemID := range rulesUIItems {
		if uiItemID == ruleItemID.UIItemID {
			rule := fa.findRule(ruleItemID.RuleID, rules)
			ruleType := fa.findRuleType(rule.RuleTypeID, rulesTypes)

			var ruleTypeEntity model.RuleType
			switch ruleType.Name {
			case "roles":
				ruleTypeEntity = model.NewRolesRuleType(ruleType.ID, ruleType.Name)
			case "privacy":
				ruleTypeEntity = model.NewPrivacyRuleType(ruleType.ID, ruleType.Name)
			case "auth":
				ruleTypeEntity = model.NewAuthRuleType(ruleType.ID, ruleType.Name)
			case "illini_cash":
				ruleTypeEntity = model.NewIlliniCashRuleType(ruleType.ID, ruleType.Name)
			case "enable":
				ruleTypeEntity = model.NewEnableRuleType(ruleType.ID, ruleType.Name)
			}

			ruleEntity := model.Rule{ID: rule.ID, RuleType: ruleTypeEntity, Value: rule.Value}
			rulesResult = append(rulesResult, ruleEntity)
		}
	}
	return &rulesResult
}

func (fa Adapter) readRuleTypesList() []ruleType {
	rulesData, err := ioutil.ReadFile("./driven/storage/data_rule_types_2.json")
	if err != nil {
		log.Fatal("Cannot read the rule types data file")
	}

	var ruleTypes []ruleType
	err = json.Unmarshal(rulesData, &ruleTypes)
	if err != nil {
		log.Fatal("Cannot unmarshal the rule types data")
	}
	return ruleTypes
}

func (fa Adapter) readRulesList() []rule {
	rulesData, err := ioutil.ReadFile("./driven/storage/data_rules_2.json")
	if err != nil {
		log.Fatal("Cannot read the rules data file")
	}

	var rules []rule
	err = json.Unmarshal(rulesData, &rules)
	if err != nil {
		log.Fatal("Cannot unmarshal the rules data")
	}
	return rules
}

func (fa Adapter) readRulesUIItemsList() []ruleUIItem {
	rulesUIItemsData, err := ioutil.ReadFile("./driven/storage/data_rules_ui_items_2.json")
	if err != nil {
		log.Fatal("Cannot read the rules ui items data file")
	}

	var rulesUIItems []ruleUIItem
	err = json.Unmarshal(rulesUIItemsData, &rulesUIItems)
	if err != nil {
		log.Fatal("Cannot unmarshal the rules in items data")
	}
	return rulesUIItems
}

func (fa Adapter) findUIItem(id int, uiItems []uiItem) *uiItem {
	for _, uiItem := range uiItems {
		if id == uiItem.ID {
			return &uiItem
		}
	}
	return nil
}

func (fa Adapter) findRule(id int, rules []rule) *rule {
	for _, rule := range rules {
		if id == rule.ID {
			return &rule
		}
	}
	return nil
}

func (fa Adapter) findRuleType(id int, ruleTypes []ruleType) *ruleType {
	for _, ruleType := range ruleTypes {
		if id == ruleType.ID {
			return &ruleType
		}
	}
	return nil
}

func (fa Adapter) readUIItemsList() []uiItem {
	uiItemsData, err := ioutil.ReadFile("./driven/storage/data_ui_items_2.json")
	if err != nil {
		log.Fatal("Cannot read the ui items data file")
	}

	var uiItems []uiItem
	err = json.Unmarshal(uiItemsData, &uiItems)
	if err != nil {
		log.Fatal("Cannot unmarshal the ui items data")
	}
	return uiItems
}

func (fa Adapter) readContentItemsList() []contentItem {
	contentItemData, err := ioutil.ReadFile("./driven/storage/data_content_items_2.json")
	if err != nil {
		log.Fatal("Cannot read the ui items data file")
	}

	var contentItems []contentItem
	err = json.Unmarshal(contentItemData, &contentItems)
	if err != nil {
		log.Fatal("Cannot unmarshal the content items data")
	}
	return contentItems
}

func (fa Adapter) readContentItemsUIItemsList() []contentItemUIItem {
	data, err := ioutil.ReadFile("./driven/storage/data_content_items_ui_items_2.json")
	if err != nil {
		log.Fatal("Cannot read the content items ui items data file")
	}

	var items []contentItemUIItem
	err = json.Unmarshal(data, &items)
	if err != nil {
		log.Fatal("Cannot unmarshal the content items ui items data")
	}
	return items
}

//NewAdapter creates a new file adapter instance
func NewAdapter() Adapter {
	version := "2"
	return Adapter{version: version}
}
