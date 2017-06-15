package concordances

import (
	db "github.com/Financial-Times/concordances-rw-dynamodb/dynamodb"
	"github.com/Financial-Times/concordances-rw-dynamodb/sns"
)

type AppConfig struct {
	AWSRegion         string
	DynamoDbTableName string
	SNSTopic          string
	AppSystemCode     string
	AppName           string
	Port              string
}

type Service interface {
	Read(uuid string) (db.ConcordancesModel, error)
	Write(m db.ConcordancesModel) (created bool, err error)
	Delete(uuid string) (bool, error)
	getDBClient() db.DynamoDBClient
	getSNSClient() sns.SNSClient
}

type ConcordancesRwService struct {
	DynamoDbTable string
	AwsRegion     string
	ddb           db.DynamoDBClient
	sns           sns.SNSClient
}

func NewConcordancesRwService(conf AppConfig) Service {
	ddbClient := db.NewDynamoDBClient(conf.DynamoDbTableName, conf.AWSRegion)
	snsClient := sns.NewSNSClient(conf.SNSTopic, conf.AWSRegion)
	s := ConcordancesRwService{DynamoDbTable: conf.DynamoDbTableName, AwsRegion: conf.AWSRegion, ddb: ddbClient, sns: snsClient}
	return &s

}

func (s *ConcordancesRwService) Read(uuid string) (db.ConcordancesModel, error) {
	model, err := s.ddb.Read(uuid)
	return model, err
}

func (s *ConcordancesRwService) Write(m db.ConcordancesModel) (created bool, err error) {
	model, err := s.ddb.Write(m)
	if err != nil {
		return created, err
	}
	if model.UUID == "" {
		created = true
	}
	err = s.sns.SendMessage(m.UUID)
	return created, err
}

func (s *ConcordancesRwService) Delete(uuid string) (bool, error) {
	model, err := s.ddb.Delete(uuid)
	if err != nil || model.UUID == "" {
		return false, err
	}

	err = s.sns.SendMessage(model.UUID)
	return true, err
}

func (s *ConcordancesRwService) getDBClient() db.DynamoDBClient {
	return s.ddb
}
func (s *ConcordancesRwService) getSNSClient() sns.SNSClient {
	return s.sns
}
