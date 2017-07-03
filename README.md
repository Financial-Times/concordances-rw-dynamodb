# concordances-rw-dynamodb
[![Circle CI](https://circleci.com/gh/Financial-Times/concordances-rw-dynamodb.svg?style=shield)](https://circleci.com/gh/Financial-Times/concordances-rw-dynamodb)[![Go Report Card](https://goreportcard.com/badge/github.com/Financial-Times/concordances-rw-dynamodb)](https://goreportcard.com/report/github.com/Financial-Times/concordances-rw-dynamodb) [![Coverage Status](https://coveralls.io/repos/github/Financial-Times/concordances-rw-dynamodb/badge.svg)](https://coveralls.io/github/Financial-Times/concordances-rw-dynamodb)

## Introduction

The concordance-rw-dynamodb service is responsible for taking a normalised concordance object and storing it into DynamoDB.
A concordance is linking one primary concept identifier to another concept identifier.

## Installation

Download the source code, dependencies and test dependencies:

        go get -u github.com/kardianos/govendor
        go get -u github.com/Financial-Times/concordances-rw-dynamodb
        cd $GOPATH/src/github.com/Financial-Times/concordances-rw-dynamodb
        govendor sync
        go build .

## Running locally

1. Run the tests and install the binary:

        govendor sync
        govendor test -v -race
        go install

2. Run the binary (using the `help` flag to see the available optional arguments):

        $GOPATH/bin/concordances-rw-dynamodb [--help]  

Options:

        --app-system-code="concordances-rw-dynamodb"            System Code of the application ($APP_SYSTEM_CODE)
        --app-name="Concordances RW DynamoDB"                   Application name ($APP_NAME)
        --port="8080"                                           Port to listen on ($APP_PORT)
        --awsRegion="eu-west-1"                                 AWS region of DynamoDB
        --dynamoDbTableName="upp-concordance-store-[env]"       Name of DynamoDB Table
        --snsTopicArn="arn:aws:sns:eu-west-1:..."               SNS Topic to notify about concordances events
        --logLeve="info"                                        Level of logging to be shown
       
Note that at this time DynamoDB and SNS topic are in the same AWS Region.  

### Test locally
Tests in dynamodb package rely on running instance of DynamoDB installed locally.  
Install Local DynamoDB following [instructions here](http://docs.aws.amazon.com/amazondynamodb/latest/developerguide/DynamoDBLocal.html)  
Start DynamoDB  
`java -Djava.library.path=./DynamoDBLocal_lib -jar DynamoDBLocal.jar -sharedDb -inMemory`
```
export AWS_SECRET_ACCESS_KEY=any_secret_key
export AWS_ACCESS_KEY_ID=any_access_id
```
`go test ./dynamodb/`

## Build and deployment

* Built by Docker Hub on merge to master: [coco/concordances-rw-dynamodb](https://hub.docker.com/r/coco/concordances-rw-dynamodb/)
* CI provided by CircleCI: [concordances-rw-dynamodb](https://circleci.com/gh/Financial-Times/concordances-rw-dynamodb)
* Code Coverage provided by coveralls.io [concordances-rw-dynamodb](https://coveralls.io/github/Financial-Times/concordances-rw-dynamodb)

## API 
* Based on the following [google doc](https://docs.google.com/document/d/1SFm7NwULX0nGqzfoX5JQGWZcd918YBwEGuO10kULovQ/edit?ts=591d86df#)   
* See the /api/api.yml for the swagger definitions of the endpoints below.  

## Utility endpoints

### GET
_summary:_ `Retrieves concordances record for a given UUID of a concept.`  
_description:_ `Given UUID of a concept as path parameter responds with concordances record for that concept in json format`  
_request:_
  
     curl -X GET "https://user:pass@pub-prod-up.ft.com/__concordances-rw-dynamodb/concordances/4f50b156-6c50-4693-b835-02f70d3f3bc0" -H  "accept: application/json; charset=UTF-8"
   
_response:_  
 
    HTTP/1.1 200 OK
    Content-Type: application/json
    X-Request-Id: transaction ID, e.g. tid_etmIWTJVeA

    {
      "uuid": "4f50b156-6c50-4693-b835-02f70d3f3bc0",
      "concordedIds": ["7c4b3931-361f-4ea4-b694-75d1630d7746", "1e5c86f8-3f38-4b6b-97ce-f75489ac3113", "0e5033fe-d079-485c-a6a1-8158ad4f37ce"]
    }
 
### PUT
_summary:_ `Stores the concordances record for a given UUID of a concept.`  
_description:_ `Expects body in json format. Expects uuid path parameter and uuid json property in the body to match. The UUID in the URL should be the primary object, if the distinction exists (eg. where the two objects are of the same type).`  
      
_request:_
```
curl -X PUT "https://user:pass@pub-prod-up.ft.com/__concordances-rw-dynamodb/concordances/4f50b156-6c50-4693-b835-02f70d3f3bc0" -H  "accept: application/json" -H  
"content-type: application/json; charset=utf-8" -d 
"{  
    "uuid": "4f50b156-6c50-4693-b835-02f70d3f3bc0",  
    "concordedIds": [
       "7c4b3931-361f-4ea4-b694-75d1630d7746 ",
       "1e5c86f8-3f38-4b6b-97ce-f75489ac3113",
       "0e5033fe-d079-485c-a6a1-8158ad4f37ce"
         ]
 }"
```
### DELETE
_summary:_ `Deletes the concordances record for a given UUID of a concept.`    
_description:_ `Given UUID of a concept as path parameter deletes the concordances record for that concept.`   

_request:_
  
    curl -X DELETE "https://user:pass@pub-prod-up.ft.com/__concordances-rw-dynamodb/concordances/4f50b156-6c50-4693-b835-02f70d3f3bc0" -H  "accept: application/json"



## Admin endpoints

`/__gtg`

`/__health`

`/__build-info`

There are several checks performed:  
 
* Checks that DynamoDB table is accessible, using parameters supplied on service startup. 
 * Checks that SNS topic is accessible, using parameters supplied on service startup. 

See the api/api.yml for the swagger definitions of these endpoints  

### Logging

* The application uses [logrus](https://github.com/Sirupsen/logrus); the log file is initialised in [main.go](main.go).
* Logging requires an `env` app parameter, for all environments other than `local` logs are written to file.
* When running locally, logs are written to console. If you want to log locally to file, you need to pass in an env parameter that is != `local`.
* NOTE: `/__build-info` and `/__gtg` endpoints are not logged as they are called every second from varnish/vulcand and this information is not needed in logs/splunk.

