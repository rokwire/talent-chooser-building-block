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

package utils

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"
)

//IsString checks if object is string type
func IsString(object interface{}) bool {
	_, ok := object.(string)
	return ok
}

//IsSlice checks if object is slice
func IsSlice(object interface{}) bool {
	switch object.(type) {
	case []interface{}:
		return true
	default:
		return false
	}
}

//Contains checks if list contains item
func Contains(list []string, item string) bool {
	for _, current := range list {
		if current == item {
			return true
		}
	}
	return false
}

//Value gives the value for string pointer
func Value(value *string) string {
	if value != nil {
		return *value
	}
	return "nil"
}

//RandSeq generates a random string
func RandSeq(n int) string {
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

	rand.Seed(time.Now().UnixNano())

	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

//LogRequest logs the request as hide some header fields because of security reasons
func LogRequest(req *http.Request) {
	if req == nil {
		return
	}

	method := req.Method
	path := req.URL.Path

	header := make(map[string][]string)
	for key, value := range req.Header {
		var logValue []string
		//do not log api key and cookies
		if key == "Rokwire-Api-Key" || key == "Cookie" {
			logValue = append(logValue, "---")
		} else {
			logValue = value
		}
		header[key] = logValue
	}
	log.Printf("%s %s %s", method, path, header)
}

//GetLogValue prepares a sensitive data to be logged.
func GetLogValue(value string) string {
	if len(value) <= 3 {
		return "***"
	}
	last3 := value[len(value)-3:]
	return fmt.Sprintf("***%s", last3)
}
