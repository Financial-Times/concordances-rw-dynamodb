package concordances

import (
	"fmt"
	db "github.com/Financial-Times/concordances-rw-dynamodb/dynamodb"
)

type AppConfig struct {
	AWSRegion         string
	DynamoDbTableName string
	appSystemCode     string
	appName           string
	port              string
}



type Service interface {
	Read(uuid string) (db.Model, error)
	Write(m db.Model) (created bool, err error)
	Delete(uuid string) (bool, error)
	Count() (int64, error)
}

type ConcordancesRwService struct {
	DynamoDbTable string
	AwsRegion     string
	ddb           db.DynamoDbClient
}

func NewConcordancesRwService(conf AppConfig) Service {
	ddbClient := db.NewDynamoDbClient(conf.DynamoDbTableName, conf.AWSRegion)
	s := ConcordancesRwService{DynamoDbTable: conf.DynamoDbTableName, AwsRegion: conf.AWSRegion, ddb: ddbClient}
	return &s

}

func (s *ConcordancesRwService) Read(uuid string) (db.Model, error) {
	model, err :=s.ddb.Read(uuid)
	fmt.Printf(">>>>>>>> %v", model)
	return model, err
}

func (s *ConcordancesRwService) Write(m db.Model) (created bool, err error) {
	model, err := s.ddb.Write(m)
	fmt.Printf(">>>>>>>> %v", model)
	if err != nil {
		return created, err
	}
	if model.UUID == "" {
		created = true
	}
	//TODO send message to SNS

	return created, err
}

func (s *ConcordancesRwService) Delete(uuid string) (bool, error) {
	model, err := s.ddb.Delete(uuid)
	if err != nil || model.UUID == "" {
		return false, err
	}

	//TODO send message to SNS
	return true, err
}

func (s *ConcordancesRwService) Count() (int64, error) {

	//err := errors.New("Not Implemented")
	return 0, nil
}
