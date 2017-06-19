package concordances

import (
	db "github.com/Financial-Times/concordances-rw-dynamodb/dynamodb"
	"github.com/Financial-Times/concordances-rw-dynamodb/sns"
	log "github.com/Sirupsen/logrus"
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
	Write(m db.ConcordancesModel) (db.Status, error)
	Delete(uuid string) (db.Status, error)
	getDBClient() db.Clienter
	getSNSClient() sns.Clienter
}

type ConcordancesRwService struct {
	DynamoDbTable string
	AwsRegion     string
	ddb           db.Clienter
	sns           sns.Clienter
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

func (s *ConcordancesRwService) Write(m db.ConcordancesModel) (status db.Status, err error) {
	status, err = s.ddb.Write(m)
	if err != nil {
		return status, err
	}
	err = s.sns.SendMessage(m.UUID)
	return status, err
}

func (s *ConcordancesRwService) Delete(uuid string) (db.Status, error) {
	status, err := s.ddb.Delete(uuid)
	if err != nil {
		return status, err
	}

	err = s.sns.SendMessage(uuid)

	if err != nil {
		log.WithError(err).WithFields(log.Fields{"UUID": uuid}).Error("Error sending Concordance to SNS")
	}

	return status, nil
}

func (s *ConcordancesRwService) getDBClient() db.Clienter {
	return s.ddb
}
func (s *ConcordancesRwService) getSNSClient() sns.Clienter {
	return s.sns
}
