package concordances

import (
	db "github.com/Financial-Times/concordances-rw-dynamodb/dynamodb"
	"github.com/Financial-Times/concordances-rw-dynamodb/sns"
)

type AppConfig struct {
	AWSRegion         string
	DynamoDbTableName string
	SnsTopic          string
	AppSystemCode     string
	AppName           string
	Port              string
}

type Service interface {
	Read(uuid string) (db.Model, error)
	Write(m db.Model) (created bool, err error)
	Delete(uuid string) (bool, error)
	Count() (int64, error)
	getDbClient() (db.DynamoDbClient)
	getSnsClient() (sns.SnsClient)
}

type ConcordancesRwService struct {
	DynamoDbTable string
	AwsRegion     string
	ddb           db.DynamoDbClient
	sns           sns.SnsClient
}

func NewConcordancesRwService(conf AppConfig) Service {
	ddbClient := db.NewDynamoDbClient(conf.DynamoDbTableName, conf.AWSRegion)
	snsClient := sns.NewSnsClient(conf.SnsTopic, conf.AWSRegion)
	s := ConcordancesRwService{DynamoDbTable: conf.DynamoDbTableName, AwsRegion: conf.AWSRegion, ddb: ddbClient, sns: snsClient}
	return &s

}

func (s *ConcordancesRwService) Read(uuid string) (db.Model, error) {
	model, err :=s.ddb.Read(uuid)
	return model, err
}

func (s *ConcordancesRwService) Write(m db.Model) (created bool, err error) {
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

func (s *ConcordancesRwService) Count() (int64, error) {
	//not implemented
	return 0, nil
}

func (s *ConcordancesRwService) getDbClient() (db.DynamoDbClient) {
	return s.ddb
}
func (s *ConcordancesRwService) getSnsClient() (sns.SnsClient) {
	return s.sns
}