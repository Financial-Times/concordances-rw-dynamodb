package concordances

import (
	fthealth "github.com/Financial-Times/go-fthealth/v1_1"
	"github.com/Financial-Times/service-status-go/gtg"
)

const healthPath = "/__health"

type healthService struct {
	config *healthConfig
	checks []fthealth.Check
}

type healthConfig struct {
	appSystemCode string
	appName       string
	port          string
	srv           Service
}

func newHealthService(config *healthConfig) *healthService {

	service := &healthService{config: config}
	service.checks = []fthealth.Check{
		service.dynamoDbCheck(), service.snsCheck(),
	}
	return service
}

func (service *healthService) gtg() gtg.Status {
	dynamoDbCheck := func() gtg.Status {
		return gtgCheck(service.dynamoDbChecker)
	}

	snsQueueCheck := func() gtg.Status {
		return gtgCheck(service.snsChecker)
	}

	return gtg.FailFastParallelCheck([]gtg.StatusChecker{
		dynamoDbCheck,
		snsQueueCheck,
	})()
}

func gtgCheck(handler func() (string, error)) gtg.Status {
	if _, err := handler(); err != nil {
		return gtg.Status{GoodToGo: false, Message: err.Error()}
	}
	return gtg.Status{GoodToGo: true}
}

func (service *healthService) dynamoDbChecker() (string, error) {
	dbClient := service.config.srv.getDBClient()
	err := dbClient.Healthcheck()
	if err != nil {
		return "Cannot connect to DynamoDB Table", err
	}
	return "DynamoDB connection is healthy", nil

}

func (service *healthService) dynamoDbCheck() fthealth.Check {
	return fthealth.Check{
		BusinessImpact: `DynamoDB healthcheck failure will cause service not to be able to store concordances in cache
		and notify downstream services of created, updated or deleted concordances records.`,
		Name:       service.config.appName,
		PanicGuide: "https://dewey.ft.com/concordances-rw-dynamodb.html",
		Severity:   1,
		TechnicalSummary: "DynamoDB healthcheck checks if the service can connect to DynamoDB, and access the table. " +
			"The failure of this healthcheck may be due to " +
			"1) incorrect name or region of DynamoDB; " +
			"2) incorrect AWS security credentials; " +
			"3) missing permissions to the DynamoDB table; " +
			"4) the table may not exist;",
		Checker: service.dynamoDbChecker,
	}
}

func (service *healthService) snsChecker() (string, error) {
	snsClient := service.config.srv.getSNSClient()
	_,err := snsClient.Healthcheck()

	if err != nil {
		return "Cannot send notifications to SNS topic", err
	}
	return "SNS Client is healthy", nil
}

func (service *healthService) snsCheck() fthealth.Check {
	return fthealth.Check{
		BusinessImpact: `SNS healthcheck failure will cause service not to be able to notify downstream services of created, updated or deleted concordances records.`,
		Name:           "SNS healthcheck",
		PanicGuide:     "https://dewey.ft.com/concordances-rw-dynamodb.html",
		Severity:       1,
		TechnicalSummary: "SNS healthcheck checks if the service can send concordances notifications to an SNS topic." +
			" The failure of this healthcheck may be due to" +
			" 1) incorrect region of SNS Topic;" +
			" 2) incorrect AWS security credentials;" +
			" 3) missing permissions For SNS Topic;" +
			" 4) Topic does not exist;",
		Checker: service.snsChecker,
	}
}
