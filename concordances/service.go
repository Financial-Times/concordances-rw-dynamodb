package concordances

import (
	db "github.com/Financial-Times/concordances-rw-dynamodb/dynamodb"
	"github.com/Financial-Times/concordances-rw-dynamodb/sns"
	log "github.com/sirupsen/logrus"
)

type AppConfig struct {
	AWSRegion         string
	DynamoDbTableName string
	SNSTopic          string
	AppSystemCode     string
	AppDescription    string
	AppName           string
	Port              string
}

type Service interface {
	Read(uuid string, transactionId string) (db.ConcordancesModel, error)
	Write(m db.ConcordancesModel, transactionId string) (db.Status, error)
	Delete(uuid string, transactionId string) (db.Status, error)
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
	return &ConcordancesRwService{DynamoDbTable: conf.DynamoDbTableName, AwsRegion: conf.AWSRegion, ddb: db.NewDynamoDBClient(conf.DynamoDbTableName, conf.AWSRegion), sns: sns.NewSNSClient(conf.SNSTopic, conf.AWSRegion)}
}

func (s *ConcordancesRwService) Read(uuid string, transactionId string) (db.ConcordancesModel, error) {
	model, err := s.ddb.Read(uuid, transactionId)
	return model, err
}

func (s *ConcordancesRwService) Write(m db.ConcordancesModel, transactionId string) (status db.Status, err error) {
	status, err = s.ddb.Write(m, transactionId)
	if err != nil {
		return status, err
	}
	err = s.sns.SendMessage(m.UUID, transactionId)

	if err != nil {
		return db.CONCORDANCE_ERROR, err
	}

	return status, err
}

func (s *ConcordancesRwService) Delete(uuid string, transactionId string) (db.Status, error) {
	status, err := s.ddb.Delete(uuid, transactionId)

	if err != nil {
		return status, err
	}

	err = s.sns.SendMessage(uuid, transactionId)

	if err != nil {
		log.WithError(err).WithFields(log.Fields{"UUID": uuid, "transaction_id": transactionId}).Error("Error sending Concordance to SNS")
		return db.CONCORDANCE_ERROR, err
	}

	return status, nil
}

func (s *ConcordancesRwService) getDBClient() db.Clienter {
	return s.ddb
}
func (s *ConcordancesRwService) getSNSClient() sns.Clienter {
	return s.sns
}
