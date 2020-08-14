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
	"io/ioutil"
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

	//1. check if there is 1.2 data
	data1p2filter := bson.D{primitive.E{Key: "version", Value: "1.2"}}
	var data1p2Result []*dataItem
	err = tchdata.Find(data1p2filter, &data1p2Result, nil)
	if err != nil {
		return err
	}
	if data1p2Result == nil || len(data1p2Result) == 0 {
		//there is no 1.2 data, so insert it
		log.Println("there is no 1.2 data, so insert it")

		data, err := ioutil.ReadFile("./driven/storage/mongodb/1.2.json")
		if err != nil {
			return err
		}
		dataItem := dataItem{"1.2", string(data)}
		_, err = tchdata.InsertOne(dataItem)
		if err != nil {
			return err
		}

	} else {
		//there is 1.2, nothing to do
		log.Println("there is 1.2, nothing to do")
	}

	//2. check if there is 2.0 data
	data2p0filter := bson.D{primitive.E{Key: "version", Value: "2.0"}}
	var data2p0Result []*dataItem
	err = tchdata.Find(data2p0filter, &data2p0Result, nil)
	if err != nil {
		return err
	}
	if data2p0Result == nil || len(data2p0Result) == 0 {
		//there is no 2.0 data, so insert it
		log.Println("there is no 2.0 data, so insert it")

		data, err := ioutil.ReadFile("./driven/storage/mongodb/2.0.json")
		if err != nil {
			return err
		}
		dataItem := dataItem{"2.0", string(data)}
		_, err = tchdata.InsertOne(dataItem)
		if err != nil {
			return err
		}

	} else {
		//there is 2.0, nothing to do
		log.Println("there is 2.0, nothing to do")
	}

	//3. check if there is 2.1 data
	data2p1filter := bson.D{primitive.E{Key: "version", Value: "2.1"}}
	var data2p1Result []*dataItem
	err = tchdata.Find(data2p1filter, &data2p1Result, nil)
	if err != nil {
		return err
	}
	if data2p1Result == nil || len(data2p1Result) == 0 {
		//there is no 2.1 data, so insert it
		log.Println("there is no 2.1 data, so insert it")

		data, err := ioutil.ReadFile("./driven/storage/mongodb/2.1.json")
		if err != nil {
			return err
		}
		dataItem := dataItem{"2.1", string(data)}
		_, err = tchdata.InsertOne(dataItem)
		if err != nil {
			return err
		}

	} else {
		//there is 2.1, nothing to do
		log.Println("there is 2.1, nothing to do")
	}

	//4. check if there is 2.2 data
	data2p2filter := bson.D{primitive.E{Key: "version", Value: "2.2"}}
	var data2p2Result []*dataItem
	err = tchdata.Find(data2p2filter, &data2p2Result, nil)
	if err != nil {
		return err
	}
	if data2p2Result == nil || len(data2p2Result) == 0 {
		//there is no 2.2 data, so insert it
		log.Println("there is no 2.2 data, so insert it")

		data, err := ioutil.ReadFile("./driven/storage/mongodb/2.2.json")
		if err != nil {
			return err
		}
		dataItem := dataItem{"2.2", string(data)}
		_, err = tchdata.InsertOne(dataItem)
		if err != nil {
			return err
		}

	} else {
		//there is 2.2, nothing to do
		log.Println("there is 2.2, nothing to do")
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
