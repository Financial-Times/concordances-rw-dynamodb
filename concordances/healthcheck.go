package concordances

import (
	health "github.com/Financial-Times/go-fthealth/v1_1"
	"github.com/Financial-Times/service-status-go/gtg"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/Financial-Times/concordances-rw-dynamodb/sns"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
)

const healthPath = "/__health"

type healthService struct {
	config *healthConfig
	checks []health.Check

}

type healthConfig struct {
	appSystemCode string
	appName       string
	port          string
	DynamoDbTable string
	AwsRegion     string
	SnsTopic      string
	ddb    dynamodbiface.DynamoDBAPI
	sns    sns.SnsClient
}

func newHealthService(config *healthConfig) *healthService {
	sess := session.Must(session.NewSession(&aws.Config{Region: aws.String(config.AwsRegion)}))
	config.ddb = dynamodb.New(sess)
	config.sns = sns.NewSnsClient(config.SnsTopic, config.AwsRegion)

	service := &healthService{config: config}
	service.checks = []health.Check{
		service.dynamoDbCheck(), service.snsCheck(),
	}
	return service
}


func (service *healthService) gtgCheck() gtg.Status {
	for _, check := range service.checks {
		if _, err := check.Checker(); err != nil {
			return gtg.Status{GoodToGo: false, Message: err.Error()}
		}
	}
	return gtg.Status{GoodToGo: true}
}

func (service *healthService) dynamoDbChecker() (string, error) {
	params := &dynamodb.DescribeTableInput{	TableName: aws.String(service.config.DynamoDbTable)}
	_, err := service.config.ddb.DescribeTable(params)
	if err != nil {
		return "Cannot connect to DynamoDB Table", err
	}
	return "DynamoDB connection is healthy", nil

}

func (service *healthService) dynamoDbCheck() health.Check {

	return health.Check{
		BusinessImpact:   `DynamoDB healthcheck failure will cause service not to be able to store concordances in cache
		and notify downstream services of created, updated or deleted concordances records.`,
		Name:             "DynamoDB healthcheck",
		PanicGuide:       "https://dewey.ft.com/concordances-rw-dynamodb.html",
		Severity:         1,
		TechnicalSummary: `DynamoDB healthcheck checks if the service can connect to DynamoDB, and access the table.
		The failure of this healthcheck may be due to
		1) incorrect name or region of DynamoDB;
		2) incorrect aws security credentials;
		3) missing permissions to the DynamoDB table;
		4) the table may not exists;
		`,
		Checker:          service.dynamoDbChecker,
	}
}

func (service *healthService) snsChecker() (string, error) {
	err := service.config.sns.SendMessage("healthcheck")
	if err != nil {
		return "Cannot send notifications to SNS topic", err
	}
	return "SNS Client is healthy", nil
}

func (service *healthService) snsCheck() health.Check {

	return health.Check{
		BusinessImpact:   `SNS healthcheck failure will cause service not to be able to notify downstream services of created, updated or deleted concordances records.`,
		Name:             "SNS healthcheck",
		PanicGuide:       "https://dewey.ft.com/concordances-rw-dynamodb.html",
		Severity:         1,
		TechnicalSummary: `SNS healthcheck checks if the service can send concordances notifications to an SNS topic.
		The failure of this healthcheck may be due to
		1) incorrect region of SNS Topic;
		2) incorrect aws security credentials;
		3) missing permissions For SNS Topic;
		4) Topic does not exist;
		`,
		Checker:          service.snsChecker,
	}
}

