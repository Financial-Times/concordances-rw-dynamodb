package concordances
//This file contains common fixtures shared by all test files in this package
import (
	db "github.com/Financial-Times/concordances-rw-dynamodb/dynamodb"

	"errors"
	"github.com/Financial-Times/concordances-rw-dynamodb/sns"
)

const (
	DDB_ERROR = "DynamoDB error"
	SNS_ERROR = "SNS error"
	EXPECTED_UUID = "uuid_123"
)

var oldModel = db.Model{UUID: EXPECTED_UUID, ConcordedIds: []string{"A", "B"}}
var updateModel = db.Model{UUID: EXPECTED_UUID, ConcordedIds: []string{"A"}}

type MockSnsClient struct {
	Happy bool
	Invoked bool
}

func (c *MockSnsClient) SendMessage(uuid string) (error) {
	c.Invoked = true
	if c.Happy {
		return nil
	}
	return errors.New(SNS_ERROR)
}

type MockDynamoDbClient struct {
	Happy bool
	model db.Model
}

func (ddb *MockDynamoDbClient) Read(uuid string) (db.Model, error) {
	if ddb.Happy {
		return oldModel, nil
	}
	return db.Model{}, errors.New(DDB_ERROR)
}

func (ddb *MockDynamoDbClient)  Write(m db.Model) (db.Model, error) {
	if !ddb.Happy {
		return db.Model{}, errors.New(DDB_ERROR)
	}

	if ddb.model.UUID == "" {
		ddb.model = m
		return db.Model{}, nil

	}
	ddb.model = m
	return oldModel, nil
}

func (ddb *MockDynamoDbClient) Delete(uuid string) (db.Model, error) {
	if !ddb.Happy {
		return db.Model{}, errors.New(DDB_ERROR)
	}
	if ddb.model.UUID == "" {
		return db.Model{}, nil
	}
	return oldModel, nil
}
func (ddb *MockDynamoDbClient) Count() (int64, error) {

	return 0, nil
}

type MockService struct {
	model   db.Model
	created bool
	deleted bool
	count   int64
	err     error
}

func (mock *MockService) Read(uuid string) (db.Model, error) {
	return mock.model, mock.err
}

func (mock *MockService) Write(m db.Model) (bool, error) {
	return mock.created, mock.err
}
func (mock *MockService) Delete(uuid string) (bool, error) {
	return mock.deleted, mock.err
}
func (mock *MockService) Count() (int64, error) {
	return mock.count, mock.err
}

func (mock *MockService) getDbClient() (db.DynamoDbClient) {
	if mock.err != nil {
		return &MockDynamoDbClient{Happy: false}
	}
	return &MockDynamoDbClient{Happy: true}
}

func (mock *MockService) getSnsClient() (sns.SnsClient) {
	if mock.err != nil {
		return &MockSnsClient{Happy: false}
	}
	return &MockSnsClient{Happy: true}
}