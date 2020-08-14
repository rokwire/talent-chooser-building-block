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

package core

import (
	"errors"
	"log"
	"talent-chooser/core/model"
)

func (app *Application) getConfig() (model.Config, error) {
	//load from storage
	config, err := app.storage.ReadConfig()
	return config, err
}

func (app *Application) getFullUIContent() map[string]*model.UIContent {
	return app.getData()
}

func (app *Application) getContentItems(dataVersion string) ([]model.ContentItem, error) {
	//read it from the storage
	contentItems, err := app.storage.ReadContentItems(dataVersion)
	if err != nil {
		log.Printf("getContentItems -> Error reading the content items from the storage %s\n", err.Error())
		return nil, err
	}
	return contentItems, nil
}

func (app *Application) getContentItem(dataVersion string, ID int) (*model.ContentItem, error) {
	//read it from the storage
	contentItem, err := app.storage.ReadContentItem(dataVersion, ID)
	if err != nil {
		log.Printf("getContentItem -> Error reading a content item from the storage %s\n", err.Error())
		return nil, err
	}
	return contentItem, nil
}

func (app *Application) createContentItem(dataVersion string, name string) (*model.ContentItem, error) {
	if len(name) == 0 {
		return nil, errors.New("Name cannot be empty")
	}
	contentItem, err := app.storage.CreateContentItem(dataVersion, name)
	if err != nil {
		return nil, err
	}

	return contentItem, nil
}

func (app *Application) updateContentItem(dataVersion string, ID int, name string) (*model.ContentItem, error) {
	if ID <= 0 {
		return nil, errors.New("The ID must be positive")
	}
	if len(name) == 0 {
		return nil, errors.New("Name cannot be empty")
	}
	contentItem, err := app.storage.UpdateContentItem(dataVersion, ID, name)
	if err != nil {
		return nil, err
	}

	return contentItem, nil
}

func (app *Application) deleteContentItem(dataVersion string, ID int) error {
	if ID <= 0 {
		return errors.New("The ID must be positive")
	}
	err := app.storage.DeleteContentItem(dataVersion, ID)
	if err != nil {
		return err
	}

	return nil
}

func (app *Application) getUIItem(dataVersion string, contentItemID int, ID int) (*model.UIItem, error) {
	//read it from the storage
	uiItem, err := app.storage.ReadUIItem(dataVersion, contentItemID, ID)
	if err != nil {
		log.Printf("getUIItem -> Error reading a ui item from the storage %s\n", err.Error())
		return nil, err
	}
	return uiItem, nil
}

func (app *Application) createUIItem(dataVersion string, contentItemID int, name string, order int) (*model.UIItem, error) {
	if contentItemID == 0 || len(name) == 0 || order == 0 {
		return nil, errors.New("Bad params")
	}
	uiItem, err := app.storage.CreateUIItem(dataVersion, contentItemID, name, order)
	if err != nil {
		return nil, err
	}

	return uiItem, nil
}

func (app *Application) updateUIItem(dataVersion string, contentItemID int, ID int, name string, order int) (*model.UIItem, error) {
	if ID <= 0 {
		return nil, errors.New("The ID must be positive")
	}
	if len(name) == 0 {
		return nil, errors.New("Name cannot be empty")
	}
	uiItem, err := app.storage.UpdateUIItem(dataVersion, contentItemID, ID, name, order)
	if err != nil {
		return nil, err
	}

	return uiItem, nil
}

func (app *Application) deleteUIItem(dataVersion string, contentItemID int, ID int) error {
	if ID <= 0 || contentItemID <= 0 {
		return errors.New("The IDs must be positive")
	}
	err := app.storage.DeleteUIItem(dataVersion, contentItemID, ID)
	if err != nil {
		return err
	}

	return nil
}

func (app *Application) getRule(dataVersion string, uiItemID int, ID int) (*model.Rule, error) {
	//read it from the storage
	rule, err := app.storage.ReadRule(dataVersion, uiItemID, ID)
	if err != nil {
		log.Printf("getRule -> Error reading a rule from the storage %s\n", err.Error())
		return nil, err
	}
	return rule, nil
}

func (app *Application) createRule(dataVersion string, uiItemID int, ruleTypeID int, value interface{}) (*model.Rule, error) {
	if uiItemID <= 0 {
		return nil, errors.New("UI item id should be possitive")
	}
	if ruleTypeID <= 0 {
		return nil, errors.New("Rule type id should be possitive")
	}

	rule, err := app.storage.CreateRule(dataVersion, uiItemID, ruleTypeID, value)
	if err != nil {
		return nil, err
	}

	return rule, nil
}

func (app *Application) updateRule(dataVersion string, ID int, uiItemID int, ruleTypeID int, value interface{}) (*model.Rule, error) {
	if ID <= 0 {
		return nil, errors.New("The ID must be positive")
	}

	rule, err := app.storage.UpdateRule(dataVersion, ID, uiItemID, ruleTypeID, value)
	if err != nil {
		return nil, err
	}

	return rule, nil
}

func (app *Application) deleteRule(dataVersion string, uiItemD int, ID int) error {
	if ID <= 0 || uiItemD <= 0 {
		return errors.New("The IDs must be positive")
	}
	err := app.storage.DeleteRule(dataVersion, uiItemD, ID)
	if err != nil {
		return err
	}

	return nil
}

func (app *Application) getRuleTypes(dataVersion string) ([]model.RuleType, error) {
	//read it from the storage
	ruleTypes, err := app.storage.ReadRuleTypes(dataVersion)
	if err != nil {
		log.Printf("getRuleTypes -> Error reading the rule types from the storage %s\n", err.Error())
		return nil, err
	}
	return ruleTypes, nil
}
