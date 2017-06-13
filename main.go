package main

import (
	"github.com/Financial-Times/concordances-rw-dynamodb/concordances"
	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	"github.com/jawher/mow.cli"
	"net/http"
	"os"
)

const appDescription = "Reads / Writes concorded concepts to DynamoDB"

func main() {
	app := cli.App("concordances-rw-dynamodb", appDescription)

	appSystemCode := app.String(cli.StringOpt{
		Name:   "app-system-code",
		Value:  "concordances-rw-dynamodb",
		Desc:   "System Code of the application",
		EnvVar: "APP_SYSTEM_CODE",
	})

	appName := app.String(cli.StringOpt{
		Name:   "app-name",
		Value:  "Concordances RW DynamoDB",
		Desc:   "Application name",
		EnvVar: "APP_NAME",
	})

	port := app.String(cli.StringOpt{
		Name:   "port",
		Value:  "8080",
		Desc:   "Port to listen on",
		EnvVar: "APP_PORT",
	})
	awsRegion := app.String(cli.StringOpt{
		Name:   "awsRegion",
		Value:  "eu-west-1",
		Desc:   "AWS region of DynamoDB",
		EnvVar: "AWS_REGION",
	})
	dynamoDbTableName := app.String(cli.StringOpt{
		Name:   "dynamoDbTableName",
		Value:  "upp-concordance-store-semantic",
		Desc:   "Name of DynamoDB Table",
		EnvVar: "DYNAMODB_TABLE_NAME",
	})
	snsTopicArn := app.String(cli.StringOpt{
		Name:   "snsTopicArn",
		Value:  "arn:aws:sns:eu-west-1:027104099916:upp-concordance-semantic-SNSTopic-SCOTT1234",
		Desc:   "SNS Topic to notify about concordances events",
		EnvVar: "SNS_TOPIC_NAME",
	})

	log.SetLevel(log.InfoLevel)
	log.Infof("[Startup] concordances-rw-dynamodb is starting ")

	app.Action = func() {
		log.Infof("System code: %s, App Name: %s, Port: %s, DynamoDb Table: %s, AWS Region: %s, SNS Topic: %s",
			*appSystemCode, *appName, *port, *dynamoDbTableName, *awsRegion, *snsTopicArn)

		conf := concordances.AppConfig{
			AWSRegion:         *awsRegion,
			DynamoDbTableName: *dynamoDbTableName,
			SnsTopic:          *snsTopicArn,
			AppSystemCode:     *appSystemCode,
			AppName:           *appName,
			Port:              *port,
		}

		router := mux.NewRouter()
		srv := concordances.NewConcordancesRwService(conf)
		concordances.NewConcordanceRwHandler(router, conf, srv)

		log.Infof("Listening on %v", *port)
		if err := http.ListenAndServe(":"+*port, nil); err != nil {
			log.Fatalf("Unable to start server: %v", err)
		}

	}
	err := app.Run(os.Args)
	if err != nil {
		log.Errorf("App could not start, error=[%s]\n", err)
		return
	}
}
