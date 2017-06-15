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
		Value:  "upp-concordance-store-test",
		Desc:   "Name of DynamoDB Table",
		EnvVar: "DYNAMODB_TABLE_NAME",
	})
	snsTopicArn := app.String(cli.StringOpt{
		Name:   "snsTopicArn",
		Value:  "arn:aws:sns:eu-west-1:027104099916:upp-concordance-semantic-SNSTopic-SCOTT1234",
		Desc:   "SNS Topic to notify about concordances events",
		EnvVar: "SNS_TOPIC_NAME",
	})
	logLevel := app.String(cli.StringOpt{
		Name:   "logLevel",
		Value:  "info",
		Desc:   "Level of logging to be shown",
		EnvVar: "LOG_LEVEL",
	})

	lvl, err := log.ParseLevel(*logLevel)
	if err != nil {
		log.Warnf("Log level %s could not be parsed, defaulting to info")
		lvl = log.InfoLevel
	}
	log.SetLevel(lvl)
	log.SetFormatter(&log.JSONFormatter{})

	log.Infof("[Startup] %s is starting", *appSystemCode)

	app.Action = func() {
		log.WithFields(log.Fields{
			"System code": 	*appSystemCode,
			"App Name": *appName,
			"Port": *port,
			"DynamoDb Table": *dynamoDbTableName,
			"AWS Region": *awsRegion,
			"SNS Topic": *snsTopicArn,

		}).Infof("Logging set to %s level", *logLevel)

		conf := concordances.AppConfig{
			AWSRegion:         *awsRegion,
			DynamoDbTableName: *dynamoDbTableName,
			SNSTopic:          *snsTopicArn,
			AppSystemCode:     *appSystemCode,
			AppName:           *appName,
			Port:              *port,
		}

		router := mux.NewRouter()
		srv := concordances.NewConcordancesRwService(conf)
		concordances.NewHandler(router, conf, srv)

		log.Infof("Listening on %v", *port)
		if err := http.ListenAndServe(":"+*port, router); err != nil {
			log.Fatalf("Unable to start server: %v", err)
		}

	}
	err = app.Run(os.Args)
	if err != nil {
		log.Errorf("App could not start, error=[%s]\n", err)
		return
	}
}
