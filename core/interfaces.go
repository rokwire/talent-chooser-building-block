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
	"log"
	"talent-chooser/core/model"
)

//Services exposes APIs for the driver adapters
type Services interface {
	GetVersion() string
	GetUIContent(user *model.User, dataVersion string, auth *model.Auth, illiniCash *model.IlliniCash) map[string][]string
	GetUIContentV2(user *model.User, dataVersion string, auth *model.AuthV2, illiniCash *model.IlliniCash) map[string][]string
	GetUIContentV3(user *model.User, dataVersion string, auth *model.AuthV3, illiniCash *model.IlliniCash, platform *model.Platform) map[string][]string
}

type servicesImpl struct {
	app *Application
}

func (s *servicesImpl) GetVersion() string {
	return s.app.getVersion()
}

func (s *servicesImpl) GetUIContent(user *model.User, dataVersion string, auth *model.Auth, illiniCash *model.IlliniCash) map[string][]string {
	return s.app.getUIContent(user, dataVersion, auth, illiniCash)
}

func (s *servicesImpl) GetUIContentV2(user *model.User, dataVersion string, auth *model.AuthV2, illiniCash *model.IlliniCash) map[string][]string {
	return s.app.getUIContentV2(user, dataVersion, auth, illiniCash)
}

func (s *servicesImpl) GetUIContentV3(user *model.User, dataVersion string, auth *model.AuthV3, illiniCash *model.IlliniCash, platform *model.Platform) map[string][]string {
	return s.app.getUIContentV3(user, dataVersion, auth, illiniCash, platform)
}

//Administration exposes administration APIs for the driver adapters
type Administration interface {
	GetConfig() (model.Config, error)
	GetFullUIContent() map[string]*model.UIContent
	ReloadUIContent() error

	GetContentItems(dataVersion string) ([]model.ContentItem, error)
	GetContentItem(dataVersion string, ID int) (*model.ContentItem, error)
	CreateContentItem(dataVersion string, name string) (*model.ContentItem, error)
	UpdateContentItem(dataVersion string, ID int, name string) (*model.ContentItem, error)
	DeleteContentItem(dataVersion string, ID int) error

	GetUIItem(dataVersion string, contentItemID int, ID int) (*model.UIItem, error)
	CreateUIItem(dataVersion string, contentItemID int, name string, order int) (*model.UIItem, error)
	UpdateUIItem(dataVersion string, contentItemID int, ID int, name string, order int) (*model.UIItem, error)
	DeleteUIItem(dataVersion string, contentItemID int, ID int) error

	GetRule(dataVersion string, uiItemID int, ID int) (*model.Rule, error)
	CreateRule(dataVersion string, uiItemID int, ruleTypeID int, value interface{}) (*model.Rule, error)
	UpdateRule(dataVersion string, ID int, uiItemID int, ruleTypeID int, value interface{}) (*model.Rule, error)
	DeleteRule(dataVersion string, uiItemID int, ID int) error

	GetRuleTypes(dataVersion string) ([]model.RuleType, error)
}

type administrationImpl struct {
	app *Application
}

func (a *administrationImpl) GetConfig() (model.Config, error) {
	return a.app.getConfig()
}

func (a *administrationImpl) GetFullUIContent() map[string]*model.UIContent {
	return a.app.getFullUIContent()
}

func (a *administrationImpl) ReloadUIContent() error {
	return a.app.loadData()
}

func (a *administrationImpl) GetContentItems(dataVersion string) ([]model.ContentItem, error) {
	return a.app.getContentItems(dataVersion)
}

func (a *administrationImpl) GetContentItem(dataVersion string, ID int) (*model.ContentItem, error) {
	return a.app.getContentItem(dataVersion, ID)
}

func (a *administrationImpl) CreateContentItem(dataVersion string, name string) (*model.ContentItem, error) {
	return a.app.createContentItem(dataVersion, name)
}

func (a *administrationImpl) UpdateContentItem(dataVersion string, ID int, name string) (*model.ContentItem, error) {
	return a.app.updateContentItem(dataVersion, ID, name)
}

func (a *administrationImpl) DeleteContentItem(dataVersion string, ID int) error {
	return a.app.deleteContentItem(dataVersion, ID)
}

func (a *administrationImpl) GetUIItem(dataVersion string, contentItemID int, ID int) (*model.UIItem, error) {
	return a.app.getUIItem(dataVersion, contentItemID, ID)
}

func (a *administrationImpl) CreateUIItem(dataVersion string, contentItemID int, name string, order int) (*model.UIItem, error) {
	return a.app.createUIItem(dataVersion, contentItemID, name, order)
}

func (a *administrationImpl) UpdateUIItem(dataVersion string, contentItemID int, ID int, name string, order int) (*model.UIItem, error) {
	return a.app.updateUIItem(dataVersion, contentItemID, ID, name, order)
}

func (a *administrationImpl) DeleteUIItem(dataVersion string, contentItemID int, ID int) error {
	return a.app.deleteUIItem(dataVersion, contentItemID, ID)
}

func (a *administrationImpl) GetRule(dataVersion string, uiItemID int, ID int) (*model.Rule, error) {
	return a.app.getRule(dataVersion, uiItemID, ID)
}

func (a *administrationImpl) CreateRule(dataVersion string, uiItemID int, ruleTypeID int, value interface{}) (*model.Rule, error) {
	return a.app.createRule(dataVersion, uiItemID, ruleTypeID, value)
}

func (a *administrationImpl) UpdateRule(dataVersion string, ID int, uiItemID int, ruleTypeID int, value interface{}) (*model.Rule, error) {
	return a.app.updateRule(dataVersion, ID, uiItemID, ruleTypeID, value)
}

func (a *administrationImpl) DeleteRule(dataVersion string, uiItemID int, ID int) error {
	return a.app.deleteRule(dataVersion, uiItemID, ID)
}

func (a *administrationImpl) GetRuleTypes(dataVersion string) ([]model.RuleType, error) {
	return a.app.getRuleTypes(dataVersion)
}

//Storage is used by core to storage data - DB storage adapter, file storage adapter etc
type Storage interface {
	Start() error
	SetStorageListener(storageListener StorageListener)

	ReadConfig() (model.Config, error)
	ReadUIContent() (map[string]*model.UIContent, error)

	ReadContentItems(dataVersion string) ([]model.ContentItem, error)
	ReadContentItem(dataVersion string, ID int) (*model.ContentItem, error)
	CreateContentItem(dataVersion string, name string) (*model.ContentItem, error)
	UpdateContentItem(dataVersion string, ID int, name string) (*model.ContentItem, error)
	DeleteContentItem(dataVersion string, ID int) error

	ReadUIItem(dataVersion string, contentItemID int, ID int) (*model.UIItem, error)
	CreateUIItem(dataVersion string, contentItemID int, name string, order int) (*model.UIItem, error)
	UpdateUIItem(dataVersion string, contentItemID int, ID int, name string, order int) (*model.UIItem, error)
	DeleteUIItem(dataVersion string, contentItemID int, ID int) error

	ReadRule(dataVersion string, uiItemID int, ID int) (*model.Rule, error)
	CreateRule(dataVersion string, uiItemID int, ruleTypeID int, value interface{}) (*model.Rule, error)
	UpdateRule(dataVersion string, ID int, uiItemID int, ruleTypeID int, value interface{}) (*model.Rule, error)
	DeleteRule(dataVersion string, uiItemID int, ID int) error

	ReadRuleTypes(dataVersion string) ([]model.RuleType, error)
}

//StorageListener listenes for change data storage events
type StorageListener interface {
	OnDataChanged()
}

type storageListenerImpl struct {
	app *Application
}

func (a *storageListenerImpl) OnDataChanged() {
	log.Println("OnDataChanged")

	//reload the cache data
	go a.app.loadData()
}
