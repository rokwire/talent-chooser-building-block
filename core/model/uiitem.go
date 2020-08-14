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
	"fmt"
)

//UIItem represents ui item entity
type UIItem struct {
	ID    int     `json:"id"`
	Name  string  `json:"name"`
	Order int     `json:"order"`
	Rules *[]Rule `json:"rules"`
}

//String gives the string representation of the ui item
func (uiItem UIItem) String() string {
	var rules string
	if uiItem.Rules != nil {
		for _, rule := range *uiItem.Rules {
			rules = fmt.Sprintf("%s %d %s %s\n\t\t\t", rules, rule.ID, rule.RuleType.GetName(), rule.Value)
		}
	}
	return fmt.Sprintf("name:%s\n\t\trules:\n\t\t[\n\t\t\t%s\n\t\t]", uiItem.Name, rules)
}
