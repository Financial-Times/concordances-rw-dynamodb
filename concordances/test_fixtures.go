package concordances

//This file contains common fixtures shared by all test files in this package
import (
	db "github.com/Financial-Times/concordances-rw-dynamodb/dynamodb"

	"errors"
	"github.com/Financial-Times/concordances-rw-dynamodb/sns"
)

const (
	DDB_ERROR     = "DynamoDB error"
	SNS_ERROR     = "SNS error"
	EXPECTED_UUID = "uuid_123"
)

var oldModel = db.ConcordancesModel{UUID: EXPECTED_UUID, ConcordedIds: []string{"A", "B"}}
var updateModel = db.ConcordancesModel{UUID: EXPECTED_UUID, ConcordedIds: []string{"A"}}

type MockSNSClient struct {
	Happy   bool
	Invoked bool
}

func (c *MockSNSClient) SendMessage(uuid string) error {
	c.Invoked = true
	if c.Happy {
		return nil
	}
	return errors.New(SNS_ERROR)
}

func (c *MockSNSClient) Healthcheck() (bool, error) {
	c.Invoked = true
	if c.Happy {
		return true, nil
	}
	return false, errors.New(SNS_ERROR)
}

type MockDynamoDBClient struct {
	Happy bool
	model db.ConcordancesModel
}

func (ddb *MockDynamoDBClient) Read(uuid string) (db.ConcordancesModel, error) {
	if ddb.Happy {
		return oldModel, nil
	}
	return db.ConcordancesModel{}, errors.New(DDB_ERROR)
}

func (ddb *MockDynamoDBClient) Write(m db.ConcordancesModel) (db.ConcordancesModel, error) {
	if !ddb.Happy {
		return db.ConcordancesModel{}, errors.New(DDB_ERROR)
	}

	if ddb.model.UUID == "" {
		ddb.model = m
		return db.ConcordancesModel{}, nil

	}
	ddb.model = m
	return oldModel, nil
}

func (ddb *MockDynamoDBClient) Delete(uuid string) (db.ConcordancesModel, error) {
	if !ddb.Happy {
		return db.ConcordancesModel{}, errors.New(DDB_ERROR)
	}
	if ddb.model.UUID == "" {
		return db.ConcordancesModel{}, nil
	}
	return oldModel, nil
}
func (ddb *MockDynamoDBClient) Count() (int64, error) {

	return 0, nil
}

type MockService struct {
	model   db.ConcordancesModel
	created bool
	deleted bool
	count   int64
	err     error
}

func (mock *MockService) Read(uuid string) (db.ConcordancesModel, error) {
	return mock.model, mock.err
}

func (mock *MockService) Write(m db.ConcordancesModel) (bool, error) {
	return mock.created, mock.err
}
func (mock *MockService) Delete(uuid string) (bool, error) {
	return mock.deleted, mock.err
}

func (mock *MockService) getDBClient() db.Client {
	if mock.err != nil {
		return &MockDynamoDBClient{Happy: false}
	}
	return &MockDynamoDBClient{Happy: true}
}

func (mock *MockService) getSNSClient() sns.Client {
	if mock.err != nil {
		return &MockSNSClient{Happy: false}
	}
	return &MockSNSClient{Happy: true}
}
