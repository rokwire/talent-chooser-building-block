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

package mongodb

import (
	"context"
	"errors"
	"log"
	"talent-chooser/core"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type database struct {
	mongoDBAuth  string
	mongoDBName  string
	mongoTimeout time.Duration

	db       *mongo.Database
	dbClient *mongo.Client

	tchdata *collectionWrapper

	listener core.StorageListener
}

func (m *database) start() error {
	log.Println("database -> start")

	//connect to the database
	clientOptions := options.Client().ApplyURI(m.mongoDBAuth)
	connectContext, cancel := context.WithTimeout(context.Background(), m.mongoTimeout)
	client, err := mongo.Connect(connectContext, clientOptions)
	cancel()
	if err != nil {
		return err
	}

	//ping the database
	pingContext, cancel := context.WithTimeout(context.Background(), m.mongoTimeout)
	err = client.Ping(pingContext, nil)
	cancel()
	if err != nil {
		return err
	}

	//apply checks
	db := client.Database(m.mongoDBName)

	tchdata := &collectionWrapper{database: m, coll: db.Collection("tchdata")}
	err = m.applyTChDataChecks(tchdata)
	if err != nil {
		return err
	}

	//asign the db, db client and the collections
	m.db = db
	m.dbClient = client

	m.tchdata = tchdata

	//watch for tchdata changes
	go m.tchdata.Watch(nil)

	return nil
}

func (m *database) applyTChDataChecks(tchdata *collectionWrapper) error {
	log.Println("apply tchdata checks.....")

	//add version index - unique
	err := tchdata.AddIndex(bson.D{primitive.E{Key: "version", Value: 1}}, true)
	if err != nil {
		return err
	}

	// check if there is 2.4 data
	filter := bson.D{primitive.E{Key: "version", Value: "2.4"}}
	var result []*dataItem
	err = tchdata.Find(filter, &result, nil)
	if err != nil {
		return err
	}
	if result == nil || len(result) == 0 {
		//there is no 2.4 data, so insert it
		log.Println("there is no 2.4 data, so insert it")

		//get the initial 2.4 data from 2.3
		filter2p3 := bson.D{primitive.E{Key: "version", Value: "2.3"}}
		var result2p3 []*dataItem
		err = tchdata.Find(filter2p3, &result2p3, nil)
		if err != nil {
			return err
		}
		if result2p3 == nil || len(result2p3) == 0 {
			return errors.New("there is no 2.3 for some reasons")
		}
		dataItem2p3 := result2p3[0]

		//insert data for 2.4
		dataItem2p4 := dataItem{Version: "2.4", Data: dataItem2p3.Data}
		_, err = tchdata.InsertOne(dataItem2p4)
		if err != nil {
			return err
		}
	} else {
		//there is 2.4, nothing to do
		log.Println("there is 2.4, nothing to do")
	}

	log.Println("tchdata checks passed")
	return nil
}

func (m *database) onDataChanged(changeDoc map[string]interface{}) {
	if changeDoc == nil {
		return
	}
	log.Printf("onDataChanged: %+v\n", changeDoc)
	ns := changeDoc["ns"]
	if ns == nil {
		return
	}
	nsMap := ns.(map[string]interface{})
	coll := nsMap["coll"]

	if "tchdata" == coll {
		log.Println("tchdata collection changed")

		if m.listener != nil {
			m.listener.OnDataChanged()
		}
	} else {
		log.Println("other collection changed")
	}
}
