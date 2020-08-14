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
	"log"
	"os"
	"strings"

	"talent-chooser/core"
	"talent-chooser/driven/storage/mongodb"
	web "talent-chooser/driver/web"
)

var (
	// Version : version of this executable
	Version string
	// Build : build date of this executable
	Build string
)

func main() {
	if len(Version) == 0 {
		Version = "dev"
	}

	//mongoDB adapter
	mongoDBAuth := getEnvKey("TCH_MONGO_AUTH", true)
	mongoDBName := getEnvKey("TCH_MONGO_DATABASE", true)
	mongoTimeout := getEnvKey("TCH_MONGO_TIMEOUT", false)
	storageAdapter := mongodb.NewStorageAdapter(mongoDBAuth, mongoDBName, mongoTimeout)
	err := storageAdapter.Start()
	if err != nil {
		log.Fatal("Cannot start the mongoDB adapter - " + err.Error())
	}

	application := core.NewApplication(Version, Build, storageAdapter)
	application.Start()

	//APIkeys
	apiKeys := getAPIKeys()
	jwtKey := getEnvKey("TCH_JWT_KEY", true)
	host := getEnvKey("TCH_HOST", true)
	oidcProvider := getEnvKey("TCH_OIDC_PROVIDER", true)
	oidcClientID := getEnvKey("TCH_OIDC_CLIENT_ID", true)
	oidcClientSecret := getEnvKey("TCH_OIDC_CLIENT_SECRET", true)
	redirectURL := getEnvKey("TCH_OIDC_REDIRECT_URL", true)
	webAdapter := web.NewWebAdapter(apiKeys, jwtKey, application, host, oidcProvider, oidcClientID, oidcClientSecret, redirectURL)
	webAdapter.Start()
}

func getAPIKeys() []string {
	//get from the environment
	rokwireAPIKeys := getEnvKey("ROKWIRE_API_KEYS", true)

	//it is comma separated format
	rokwireAPIKeysList := strings.Split(rokwireAPIKeys, ",")
	if len(rokwireAPIKeysList) <= 0 {
		log.Fatal("For some reasons the apis keys list is empty")
	}

	return rokwireAPIKeysList
}

func getEnvKey(key string, required bool) string {
	//get from the environment
	value, exist := os.LookupEnv(key)
	if !exist {
		if required {
			log.Fatal("No provided environment variable for " + key)
		} else {
			log.Printf("No provided environment variable for " + key)
		}
	}
	printEnvVar(key, value)
	return value
}

func printEnvVar(name string, value string) {
	if Version == "dev" {
		log.Printf("%s=%s", name, value)
	}
}
