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
	"sync"
	"talent-chooser/core/model"
)

//Application represents the core application code based on hexagonal architecture
type Application struct {
	version string
	build   string

	Services       Services       //expose to the drivers adapters
	Administration Administration //expose to the drivers adapters

	storage Storage

	//data cache
	dataLock       *sync.RWMutex
	data           map[string]*model.UIContent // version - data
	dataStatusLock *sync.RWMutex
	dataStatus     bool
}

//Start starts the core part of the application
func (app *Application) Start() {
	storageListener := storageListenerImpl{app: app}
	app.storage.SetStorageListener(&storageListener)

	app.loadData()
}

func (app *Application) loadData() error {
	log.Println("Start loading data")

	app.setDataStatus(false)

	data, err := app.storage.ReadUIContent()
	if err != nil {
		log.Printf("Error on loading data... %s\n", err.Error())

		app.setDataStatus(true)
		return err
	}
	app.setData(data)
	app.setDataStatus(true)

	log.Println("Successfully loaded data")

	return nil
}

func (app *Application) setData(data map[string]*model.UIContent) {
	app.dataLock.RLock()
	app.data = data

	log.Println("Set data...")

	app.dataLock.RUnlock()
}

func (app *Application) getData() map[string]*model.UIContent {
	//wait until the data is ready
	for !app.getDataStatus() {
	}

	app.dataLock.RLock()
	defer app.dataLock.RUnlock()

	return app.data
}

func (app *Application) setDataStatus(status bool) {
	app.dataStatusLock.RLock()
	app.dataStatus = status
	app.dataStatusLock.RUnlock()
}

func (app *Application) getDataStatus() bool {
	app.dataStatusLock.RLock()
	defer app.dataStatusLock.RUnlock()

	return app.dataStatus
}

//NewApplication creates new Application
func NewApplication(version string, build string, storage Storage) *Application {
	dataLock := &sync.RWMutex{}
	dataStatusLock := &sync.RWMutex{}
	data := map[string]*model.UIContent{}
	application := Application{version: version, build: build, storage: storage,
		dataLock: dataLock, dataStatusLock: dataStatusLock, data: data}

	//add the drivers ports/interfaces
	application.Services = &servicesImpl{app: &application}
	application.Administration = &administrationImpl{app: &application}

	return &application
}
