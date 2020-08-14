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

package main

import (
	"os"
	"testing"
)

func TestGetApiKeys(t *testing.T) {
	os.Setenv("ROKWIRE_API_KEYS", "1234,567,8910;67")

	keys := getAPIKeys()
	if len(keys) != 3 {
		t.Errorf("Wrong keys list size %d", len(keys))
	}
	key1 := keys[0]
	key2 := keys[1]
	key3 := keys[2]

	if key1 != "1234" {
		t.Errorf("Key1 is wrong %s", key1)
	}
	if key2 != "567" {
		t.Errorf("Key2 is wrong %s", key2)
	}
	if key3 != "8910;67" {
		t.Errorf("Key3 is wrong %s", key3)
	}
}
