package concordances

import (
	"errors"
	"reflect"
	"testing"

	"fmt"

	db "github.com/Financial-Times/concordances-rw-dynamodb/dynamodb"
	"github.com/Financial-Times/concordances-rw-dynamodb/sns"
	"github.com/stretchr/testify/assert"
)

func createService(ddbClient db.Clienter, snsClient sns.Clienter) ConcordancesRwService {
	return ConcordancesRwService{
		DynamoDbTable: "TestTable",
		AwsRegion:     "TestRegion",
		ddb:           ddbClient,
		sns:           snsClient,
	}
}

func TestServiceRead(t *testing.T) {
	tests := []struct {
		testName           string
		mockDynamoDBClient MockDynamoDBClient
		errorString        string
	}{
		{"Successful Read", MockDynamoDBClient{Happy: true}, ""},
		{"Unsuccessful Read due to DynamoDB", MockDynamoDBClient{Happy: false}, "DynamoDB error"},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			mockSNSClient := MockSNSClient{}
			srv := createService(&test.mockDynamoDBClient, &mockSNSClient)
			m, err := srv.Read(EXPECTED_UUID, "testing_tid_1234")

			if test.errorString != "" {
				assert.Error(t, err, errors.New(test.errorString))
			} else {
				assert.NoError(t, err, "Failed on service error.")
				assert.True(t, reflect.DeepEqual(db.ConcordancesModel{UUID: EXPECTED_UUID, ConcordedIds: []string{"A", "B"}}, m), "Model did not match.")
			}
			assert.False(t, mockSNSClient.Invoked, "Should not send SNS notifications on read")
		})
	}
}

func TestServiceWrite(t *testing.T) {
	tests := []struct {
		testName         string
		mockDynamoClient MockDynamoDBClient
		mockSNSClient    MockSNSClient
		model            db.ConcordancesModel
		status           db.Status
		errorString      string
	}{
		{"Successful Create", MockDynamoDBClient{Happy: true, model: db.ConcordancesModel{}},
			MockSNSClient{Happy: true}, db.ConcordancesModel{UUID: EXPECTED_UUID, ConcordedIds: []string{"A"}},
			db.CONCORDANCE_CREATED, ""},
		{"Successful Update", MockDynamoDBClient{Happy: true, model: db.ConcordancesModel{}},
			MockSNSClient{Happy: true}, db.ConcordancesModel{UUID: EXPECTED_UUID, ConcordedIds: []string{"A"}},
			db.CONCORDANCE_UPDATED, ""},
		{"Un-successful Write Due to DynamoDB", MockDynamoDBClient{Happy: false, model: db.ConcordancesModel{UUID: EXPECTED_UUID, ConcordedIds: []string{"A", "B"}}},
			MockSNSClient{Happy: true}, db.ConcordancesModel{UUID: EXPECTED_UUID, ConcordedIds: []string{"A", "B"}}, db.CONCORDANCE_ERROR, "DynamoDB error"},
		{"Un-successful Write Due to SNS", MockDynamoDBClient{Happy: true, model: db.ConcordancesModel{}},
			MockSNSClient{Happy: false}, db.ConcordancesModel{}, db.CONCORDANCE_ERROR, "SNS error"},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			srv := createService(&test.mockDynamoClient, &test.mockSNSClient)
			status, err := srv.Write(test.model, "testing_tid_1234")

			if test.status == db.CONCORDANCE_UPDATED {
				status, err = srv.Write(test.model, "testing_tid_1234")
			}
			if test.errorString != "" {
				assert.Error(t, err, errors.New(test.errorString))
				if test.mockSNSClient.Happy != false {
					assert.False(t, test.mockSNSClient.Invoked, "Should not send SNS notifications on error")
				}
			} else {
				assert.NoError(t, err, "Failed on service error.")
				assert.True(t, test.mockSNSClient.Invoked, "Did not envoke SNS Client")
			}
			assert.Equal(t, test.status, status, fmt.Sprintf("Status did not match. Expected status: %v, Actual status: %v", test.status, status))

		})
	}
}

func TestServiceDelete(t *testing.T) {
	tests := []struct {
		testName         string
		mockDynamoClient MockDynamoDBClient
		mockSNSClient    MockSNSClient
		uuid             string
		status           db.Status
		errorString      string
	}{
		{"Successful Delete", MockDynamoDBClient{Happy: true, model: db.ConcordancesModel{}},
			MockSNSClient{Happy: true}, EXPECTED_UUID, db.CONCORDANCE_DELETED, ""},
		{"Unsuccessful Delete due to DynamoDB", MockDynamoDBClient{Happy: false, model: db.ConcordancesModel{}},
			MockSNSClient{Happy: true}, EXPECTED_UUID, db.CONCORDANCE_ERROR, "DynamoDB error"},
		{"Unsuccessful Delete due to SNS", MockDynamoDBClient{Happy: true, model: db.ConcordancesModel{}},
			MockSNSClient{Happy: false}, EXPECTED_UUID, db.CONCORDANCE_ERROR, "SNS error"},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			srv := createService(&test.mockDynamoClient, &test.mockSNSClient)

			status, err := srv.Write(db.ConcordancesModel{UUID: "123456789", ConcordedIds: []string{"A"}}, "testing_tid_1234")
			status, err = srv.Delete(test.uuid, "testing_tid_1234")

			if test.errorString != "" {
				assert.Contains(t, err.Error(), test.errorString, "Error incorrect")

				if test.mockSNSClient.Happy != false {
					assert.False(t, test.mockSNSClient.Invoked, "Should not send SNS notifications on error")
				}

			} else {
				assert.NoError(t, err, "Failed on service error.")
				assert.True(t, test.mockSNSClient.Invoked, "Did not envoke SNS Client")

			}
			assert.Equal(t, test.status, status, fmt.Sprintf("Status did not match. Expected status: %v, Actual status: %v", test.status, status))
		})
	}
}

const (
	DDB_ERROR     = "DynamoDB error"
	SNS_ERROR     = "SNS error"
	EXPECTED_UUID = "uuid_123"
)

type MockSNSClient struct {
	Happy   bool
	Invoked bool
}

func (c *MockSNSClient) SendMessage(uuid string, transaction_id string) error {
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

func (ddb *MockDynamoDBClient) Read(uuid string, transaction_id string) (db.ConcordancesModel, error) {
	if ddb.Happy {
		return db.ConcordancesModel{UUID: EXPECTED_UUID, ConcordedIds: []string{"A", "B"}}, nil
	}
	return db.ConcordancesModel{}, errors.New(DDB_ERROR)
}

func (ddb *MockDynamoDBClient) Write(m db.ConcordancesModel, transaction_id string) (db.Status, error) {
	if !ddb.Happy {
		return db.CONCORDANCE_ERROR, errors.New(DDB_ERROR)
	}

	if ddb.model.UUID == "" {
		ddb.model = m
		return db.CONCORDANCE_CREATED, nil
	}
	ddb.model = m
	return db.CONCORDANCE_UPDATED, nil
}

func (ddb *MockDynamoDBClient) Delete(uuid string, transaction_id string) (db.Status, error) {
	if !ddb.Happy {
		return db.CONCORDANCE_ERROR, errors.New(DDB_ERROR)
	}
	if ddb.model.UUID == "" {
		return db.CONCORDANCE_NOT_FOUND, nil
	}
	return db.CONCORDANCE_DELETED, nil
}

func (ddb *MockDynamoDBClient) Healthcheck() error {
	return nil
}

type MockService struct {
	model  db.ConcordancesModel
	status db.Status
	count  int64
	err    error
}

func (mock *MockService) Read(uuid string, transaction_id string) (db.ConcordancesModel, error) {
	return mock.model, mock.err
}

func (mock *MockService) Write(m db.ConcordancesModel, transaction_id string) (db.Status, error) {
	if mock.status == 0 {
		return db.CONCORDANCE_CREATED, mock.err
	}
	return mock.status, mock.err
}

func (mock *MockService) Delete(uuid string, transaction_id string) (db.Status, error) {
	if mock.status == 0 {
		return db.CONCORDANCE_DELETED, mock.err
	}
	return mock.status, mock.err
}

func (mock *MockService) getDBClient() db.Clienter {
	if mock.err != nil {
		return &MockDynamoDBClient{Happy: false}
	}
	return &MockDynamoDBClient{Happy: true}
}

func (mock *MockService) getSNSClient() sns.Clienter {
	if mock.err != nil {
		return &MockSNSClient{Happy: false}
	}
	return &MockSNSClient{Happy: true}
}
