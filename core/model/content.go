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

//UIContent represents content ui entity
type UIContent struct {
	Data []ContentItem `json:"data"`
}

//Print prints the ui content strucute
func (uiContent *UIContent) Print() {
	for key, contentItem := range uiContent.Data {
		fmt.Printf("[\n\tkey/id:%d\n\t%s]\n", key, contentItem.String())
	}
}

//NewUIContent creates new ui content instance
func NewUIContent(data []ContentItem) UIContent {
	return UIContent{Data: data}
}

//ContentItem represents content item entity
type ContentItem struct {
	ID      int      `json:"id"`
	Name    string   `json:"name"`
	UIItems []UIItem `json:"ui-items"`
}

//String give the string representation of the content item
func (cItem *ContentItem) String() string {
	var children string
	if cItem.UIItems != nil {
		for _, uiItem := range cItem.UIItems {
			children = fmt.Sprintf("%s\n\t\t%s", children, uiItem.String())
		}
	}
	return fmt.Sprintf("ui item:[\n\t\t%s\n\t]\n\tchildren:[\n\t\t%s\n\t]\t\n", cItem.Name, children)
}
