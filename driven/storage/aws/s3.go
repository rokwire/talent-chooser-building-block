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

package aws

import (
	"bytes"
	"encoding/json"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

func (a *Adapter) createS3Session() (*session.Session, error) {
	var s *session.Session
	var err error
	if len(a.S3AccessKeyID) > 0 && len(a.S3SecretAccessKey) > 0 {
		//we have provided access id and secret key - for dev environment
		s, err = session.NewSession(&aws.Config{
			Region: aws.String(a.S3Region),
			Credentials: credentials.NewStaticCredentials(
				a.S3AccessKeyID,
				a.S3SecretAccessKey,
				""),
		})
		log.Println("Static S3 credentials")
	} else {
		//we do have provided access id and secret key, so rely on the aws infrastructure
		s, err = session.NewSession(&aws.Config{
			Region: aws.String(a.S3Region),
		})
		log.Println("AWS S3 infrastrucutre credentials")
	}

	if err != nil {
		log.Printf("Error creating S3 session %s", err.Error())
		return nil, err
	}
	return s, nil
}

func (a *Adapter) downloadData() (*data, error) {
	downloadedData, err := a.downloadFile("data")
	if err != nil {
		log.Printf("Cannot download the data file %s\n", err)
		return nil, err
	}

	var data data
	err = json.Unmarshal(downloadedData, &data)
	if err != nil {
		log.Printf("Cannot unmarshal the data %s\n", err)
		return nil, err
	}
	return &data, nil
}

func (a *Adapter) uploadData(data *data) error {
	data.LastUpdated = time.Now().UTC().String()

	d, err := json.MarshalIndent(data, "", "    ")
	if err != nil {
		log.Printf("Cannot marshal the data %s\n", err)
		return err
	}
	err = a.uploadFile("data", d)
	if err != nil {
		log.Printf("Cannot upload the data file %s\n", err)
		return err
	}
	return nil
}

func (a *Adapter) downloadFile(fileName string) ([]byte, error) {
	fileName = fileName + "_" + a.version + ".json"

	buff := &aws.WriteAtBuffer{}
	downloader := s3manager.NewDownloader(a.awsSession)
	_, err := downloader.Download(buff,
		&s3.GetObjectInput{
			Bucket: aws.String(a.S3Bucket),
			Key:    aws.String(fileName),
		})
	if err != nil {
		log.Printf("Error downloading file %s with error %s", fileName, err.Error())
		return nil, err
	}
	data := buff.Bytes()
	return data, nil
}

func (a *Adapter) uploadFile(fileName string, data []byte) error {
	fileName = fileName + "_" + a.version + ".json"

	preparedData := bytes.NewReader(data)
	uploader := s3manager.NewUploader(a.awsSession)
	_, err := uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(a.S3Bucket),
		Key:    aws.String(fileName),
		Body:   preparedData,
	})
	if err != nil {
		log.Printf("Error on uploading file to S3 %s", err.Error())
		return err
	}
	return nil
}
