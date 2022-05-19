# Talent Chooser Building Block (Archived)

Go project to provide rest service for rokwire building block talent chooser results.

The service is based on clear hexagonal architecture.

## Set Up

### Prerequisites

MongoDB v4.2.2+

Go v1.15+

### Environment variables
The following Environment variables are supported. The service will not start unless those marked as Required are supplied.

Name|Value|Required|Description
---|---|---|---
ROKWIRE_API_KEYS | <value1,value2,value3> | yes | Comma separated list of rokwire api keys
TCH_MONGO_AUTH | <mongodb://USER:PASSWORD@HOST:PORT/DATABASE NAME> | yes | MongoDB authentication string. The user must have read/write privileges.
TCH_MONGO_DATABASE | < value > | yes | MongoDB database name
TCH_MONGO_TIMEOUT | < value > | no | MongoDB timeout in milliseconds. Set default value(500 milliseconds) if omitted
TCH_JWT_KEY | < value > | yes | JWT key
TCH_HOST | < value > | yes | Host
TCH_OIDC_PROVIDER | < value > | yes | OIDC provider
TCH_OIDC_CLIENT_ID | < value > | yes | OIDC client id
TCH_OIDC_CLIENT_SECRET | < value > | yes | OIDC client secret
TCH_OIDC_REDIRECT_URL | < value > | yes | OIDC redirect url

### Run Application

#### Run locally without Docker

1. Clone the repo (outside GOPATH)

2. Open the terminal and go to the root folder
  
3. Make the project  
```
$ make
...
▶ building executable(s)… 1.2.0 2020-08-13T10:00:00+0300
```

4. Run the executable
```
$ ./bin/talent-chooser
```

#### Run locally as Docker container

1. Clone the repo (outside GOPATH)

2. Open the terminal and go to the root folder
  
3. Create Docker image  
```
docker build -t talent-chooser .
```
4. Run as Docker container
```
docker run -e ROKWIRE_API_KEYS -e TCH_MONGO_AUTH -e TCH_MONGO_DATABASE -e TCH_MONGO_TIMEOUT -e TCH_JWT_KEY -e TCH_HOST -e TCH_OIDC_PROVIDER -e TCH_OIDC_CLIENT_ID -e TCH_OIDC_CLIENT_SECRET -e TCH_OIDC_REDIRECT_URL -p 80:80 talent-chooser
```

#### Tools

##### Run tests
```
$ make tests
```

##### Run code coverage tests
```
$ make cover
```

##### Run golint
```
$ make lint
```

##### Run gofmt to check formatting on all source files
```
$ make checkfmt
```

##### Run gofmt to fix formatting on all source files
```
$ make fixfmt
```

##### Cleanup everything
```
$ make clean
```

##### Run help
```
$ make help
```

##### Generate Swagger docs
```
$ make swagger
```

### Test Application APIs

Verify the service is running as calling the get version API.

#### Call get version API

curl -X GET -i http://localhost/talent-chooser/api/version

Response
```
1.2.0
```

## Documentation

The documentation is placed here - https://api-dev.rokwire.illinois.edu/docs/?urls.primaryName=Talent%20Chooser%20Building%20Block

Alternativelly the documentation is served by the service on the following url - https://api-dev.rokwire.illinois.edu/talent-chooser/doc/ui/

