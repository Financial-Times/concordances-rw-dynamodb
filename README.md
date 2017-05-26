# concordances-rw-dynamodb
[![Circle CI](https://circleci.com/gh/Financial-Times/concordances-rw-dynamodb/tree/master.png?style=shield)](https://circleci.com/gh/Financial-Times/concordances-rw-dynamodb/tree/master)[![Go Report Card](https://goreportcard.com/badge/github.com/Financial-Times/concordances-rw-dynamodb)](https://goreportcard.com/report/github.com/Financial-Times/concordances-rw-dynamodb) [![Coverage Status](https://coveralls.io/repos/github/Financial-Times/concordances-rw-dynamodb/badge.svg)](https://coveralls.io/github/Financial-Times/concordances-rw-dynamodb)

## Introduction

This service reads and writes concorded concepts to DynamoDB

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
        

## Build and deployment
_How can I build and deploy it (lots of this will be links out as the steps will be common)_

* Built by Docker Hub on merge to master: [coco/concordances-rw-dynamodb](https://hub.docker.com/r/coco/concordances-rw-dynamodb/)
* CI provided by CircleCI: [concordances-rw-dynamodb](https://circleci.com/gh/Financial-Times/concordances-rw-dynamodb)

## Service endpoints
See the api/api.yml for the swagger definitions of these endpoints

### GET

Using curl:  
  _request:_
  
     curl -X GET "https://user:pass@prod-up.ft.com/__concordances-rw-dynamodb/concordances/4f50b156-6c50-4693-b835-02f70d3f3bc0" -H  "accept: application/json; charset=UTF-8"
   
 _response:_  
 
    HTTP/1.1 200 OK
    Content-Type: application/json
    X-Request-Id: transaction ID, e.g. tid_etmIWTJVeA

    {
      "uuid": "4f50b156-6c50-4693-b835-02f70d3f3bc0",
      "concordedIds": ["7c4b3931-361f-4ea4-b694-75d1630d7746", "1e5c86f8-3f38-4b6b-97ce-f75489ac3113", "0e5033fe-d079-485c-a6a1-8158ad4f37ce"]
    }
 
  _request:_
  
    curl -X GET "https://user:pass@prod-up.ft.com/__concordances-rw-dynamodb/concordances/__count" -H  "accept: text/plain"
    HTTP/1.1 200 OK
    Content-Type: text/plain
    100

### PUT
Using curl:  
  _request:_
```
curl -X PUT "https://user:pass@prod-up.ft.com/__concordances-rw-dynamodb/concordances/4f50b156-6c50-4693-b835-02f70d3f3bc0" -H  "accept: application/json" -H  
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
Using curl:  
  _request:_
  
    curl -X DELETE "https://api.ft.com/__concordances-rw-dynamodb/concordances/4f50b156-6c50-4693-b835-02f70d3f3bc0" -H  "accept: application/json"

Based on the following [google doc](https://docs.google.com/document/d/1SFm7NwULX0nGqzfoX5JQGWZcd918YBwEGuO10kULovQ/edit?ts=591d86df#)

## Utility endpoints
N/A  

## Healthchecks
Admin endpoints are:

`/__gtg`

`/__health`

`/__build-info`

There are several checks performed:  
 
* Checks that DynamoDB table is accessible, using parameters supplied on service startup.  

See the api/api.yml for the swagger definitions of these endpoints  

## Other information
TODO  
_Anything else you want to add._

_e.g. (NB: this example may be something we want to extract as it's probably common to a lot of services)_

### Logging

* The application uses [logrus](https://github.com/Sirupsen/logrus); the log file is initialised in [main.go](main.go).
* Logging requires an `env` app parameter, for all environments other than `local` logs are written to file.
* When running locally, logs are written to console. If you want to log locally to file, you need to pass in an env parameter that is != `local`.
* NOTE: `/__build-info` and `/__gtg` endpoints are not logged as they are called every second from varnish/vulcand and this information is not needed in logs/splunk.