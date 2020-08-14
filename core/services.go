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
	"fmt"
	"log"
	"sort"
	"talent-chooser/core/model"
)

func (app *Application) getVersion() string {
	return app.version
}

func (app *Application) getUIContent(user *model.User, dataVersion string, auth *model.Auth, illiniCash *model.IlliniCash) map[string][]string {
	app.printGetUIContentParameters(user, dataVersion, auth, illiniCash)

	inputRulesparameters := model.InputRulesParameters{
		User: user, Auth: auth, AuthV2: nil, AuthV3: nil, AuthVersion: "1", IlliniCash: illiniCash, Platform: nil}
	readyData := app.prepareData(dataVersion, inputRulesparameters)
	return readyData
}

func (app *Application) getUIContentV2(user *model.User, dataVersion string, auth *model.AuthV2, illiniCash *model.IlliniCash) map[string][]string {
	app.printGetUIContentV2Parameters(user, dataVersion, auth, illiniCash)

	inputRulesparameters := model.InputRulesParameters{
		User: user, Auth: nil, AuthV2: auth, AuthV3: nil, AuthVersion: "2", IlliniCash: illiniCash, Platform: nil}
	readyData := app.prepareData(dataVersion, inputRulesparameters)
	return readyData
}

func (app *Application) getUIContentV3(user *model.User, dataVersion string, auth *model.AuthV3, illiniCash *model.IlliniCash, platform *model.Platform) map[string][]string {
	app.printGetUIContentV3Parameters(user, dataVersion, auth, illiniCash, platform)

	inputRulesparameters := model.InputRulesParameters{
		User: user, Auth: nil, AuthV2: nil, AuthV3: auth, AuthVersion: "3", IlliniCash: illiniCash, Platform: platform}
	readyData := app.prepareData(dataVersion, inputRulesparameters)
	return readyData
}

//apply rules and sort
func (app *Application) prepareData(dataVersion string, inputRulesParameters model.InputRulesParameters) map[string][]string {
	result := make(map[string][]string)
	allData := app.getData()
	data := allData[dataVersion]
	for _, item := range data.Data {
		name := item.Name
		uiItems := item.UIItems

		if uiItems != nil {
			//sort
			sort.Slice(uiItems, func(i, j int) bool {
				return uiItems[i].Order < uiItems[j].Order
			})

			//apply rules
			var uiItemsList []string
			for _, uiItem := range uiItems {
				if app.matchRules(uiItem, inputRulesParameters) {
					uiItemsList = append(uiItemsList, uiItem.Name)
				}
			}

			if len(uiItemsList) > 0 {
				result[name] = uiItemsList
			}
		}
	}
	return result
}

func (app *Application) matchRules(uiItem model.UIItem, inputRulesParams model.InputRulesParameters) bool {
	itemRules := uiItem.Rules
	if itemRules == nil {
		return true //if no rules it matches
	}

	for _, rule := range *itemRules {
		matches := rule.Match(inputRulesParams)
		if !matches {
			return false //does not match if any of them does not match
		}
	}
	return true
}

func (app *Application) printGetUIContentParameters(user *model.User, dataVersion string, auth *model.Auth, illiniCash *model.IlliniCash) {
	userData := "nil"
	if user != nil {
		userData = "1"
	}
	authData := "nil"
	if auth != nil {
		authData = "1"
	}
	illiniCashData := "nil"
	if illiniCash != nil {
		illiniCashData = "1"
	}
	log.Printf("getUIContent -> input data [\n\tdata-version:%s,\n\tuser:%s,\n\tauth:%s,\n\tillini cash:%s\n]\n",
		dataVersion, userData, authData, illiniCashData)
}

func (app *Application) printGetUIContentV2Parameters(user *model.User, dataVersion string, auth *model.AuthV2, illiniCash *model.IlliniCash) {
	userData := "nil"
	if user != nil {
		userData = "1"
	}
	authData := "nil"
	if auth != nil {
		authData = "1"
	}
	illiniCashData := "nil"
	if illiniCash != nil {
		illiniCashData = "1"
	}
	log.Printf("printGetUIContentV2Parameters -> input data [\n\tdata-version:%s,\n\tuser:%s,\n\tauth:%s,\n\tillini cash:%s\n]\n",
		dataVersion, userData, authData, illiniCashData)
}

func (app *Application) printGetUIContentV3Parameters(user *model.User, dataVersion string, auth *model.AuthV3,
	illiniCash *model.IlliniCash, platform *model.Platform) {
	userData := "nil"
	if user != nil {
		userData = "1"
	}
	authData := "nil"
	if auth != nil {
		authData = "1"

		token := "nil"
		if auth.Token != nil {
			token = "1"
		}
		user := "nil"
		if auth.User != nil {
			user = "1"
		}
		card := "nil"
		if auth.Card != nil {
			card = "1"
		}
		pii := "nil"
		if auth.Pii != nil {
			pii = "1"
		}
		authData = fmt.Sprintf("[token:%s\tuser:%s\tcard:%s\tpii:%s]", token, user, card, pii)
	}
	illiniCashData := "nil"
	if illiniCash != nil {
		illiniCashData = "1"
	}
	platformData := "nil"
	if platform != nil {
		platformData = "1"
	}
	log.Printf("printGetUIContentV3Parameters -> input data [\n\tdata-version:%s,\n\tuser:%s,\n\tauth:%s,\n\tillini cash:%s\n\tplatform:%s\n]\n",
		dataVersion, userData, authData, illiniCashData, platformData)
}
