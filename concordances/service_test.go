package concordances

import (
	db "github.com/Financial-Times/concordances-rw-dynamodb/dynamodb"
	"github.com/Financial-Times/concordances-rw-dynamodb/sns"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
	"errors"
)

func createService(ddbClient db.Clienter, snsClient sns.Clienter) ConcordancesRwService {
	return ConcordancesRwService{
		DynamoDbTable: "TestTable",
		AwsRegion:     "TestRegion",
		ddb:           ddbClient,
		sns:           snsClient,
	}
}

func TestServiceRead_NoError(t *testing.T) {
	ddbClient := MockDynamoDBClient{Happy: true}
	snsClient := MockSNSClient{}
	srv := createService(&ddbClient, &snsClient)

	m, err := srv.Read(EXPECTED_UUID, "testing_tid_1234")
	assert.NoError(t, err, "Failed on service error.")
	assert.True(t, reflect.DeepEqual(oldModel, m), "Model did not match.")
	assert.False(t, snsClient.Invoked, "Should not send sns notifications on read")
}

func TestServiceRead_DynamoBbError(t *testing.T) {
	ddbClient := MockDynamoDBClient{Happy: false}
	snsClient := MockSNSClient{}
	srv := createService(&ddbClient, &snsClient)
	_, err := srv.Read(EXPECTED_UUID, "testing_tid_1234")
	assert.Equal(t, DDB_ERROR, err.Error(), "Failed to return service error.")
	assert.False(t, snsClient.Invoked, "Should not send sns notifications on read")
}

func TestServiceCreate_NoError(t *testing.T) {
	ddbClient := MockDynamoDBClient{Happy: true, model: db.ConcordancesModel{}}
	snsClient := MockSNSClient{Happy: true}
	srv := createService(&ddbClient, &snsClient)
	updateModel := db.ConcordancesModel{UUID: EXPECTED_UUID, ConcordedIds: []string{"A"}}

	status, err := srv.Write(updateModel, "testing_tid_1234")

	assert.NoError(t, err, "Failed on service error.")
	assert.Equal(t, db.CONCORDANCE_CREATED, status, "Did not detect that new record was created.")
	assert.True(t, snsClient.Invoked, "Did not envoke sns Client")
}

func TestServiceUpdate_NoError(t *testing.T) {
	ddbClient := MockDynamoDBClient{Happy: true, model: oldModel}
	snsClient := MockSNSClient{Happy: true}
	srv := createService(&ddbClient, &snsClient)
	updateModel := db.ConcordancesModel{UUID: EXPECTED_UUID, ConcordedIds: []string{"A"}}

	status, err := srv.Write(updateModel, "testing_tid_1234")

	assert.NoError(t, err, "Failed on service error.")
	assert.Equal(t, db.CONCORDANCE_UPDATED, status, "Did not detect that record was updated.")
	assert.True(t, snsClient.Invoked, "Did not envoke sns Client")
}

func TestServiceWrite_DynamoDbError(t *testing.T) {
	ddbClient := MockDynamoDBClient{Happy: false, model: oldModel}
	snsClient := MockSNSClient{Happy: true}
	srv := createService(&ddbClient, &snsClient)
	updateModel := db.ConcordancesModel{UUID: EXPECTED_UUID, ConcordedIds: []string{"A"}}
	status, err := srv.Write(updateModel, "testing_tid_1234")
	assert.Equal(t, DDB_ERROR, err.Error(), "Failed to return service error.")
	assert.Equal(t, db.CONCORDANCE_ERROR, status, "Did not detect existing record was updated.")
	assert.False(t, snsClient.Invoked, "Should not have invoked sns Client when error from DynamoDB")
}

func TestServiceWrite_SnsError(t *testing.T) {
	ddbClient := MockDynamoDBClient{Happy: true, model: db.ConcordancesModel{}}
	snsClient := MockSNSClient{Happy: false}
	srv := createService(&ddbClient, &snsClient)

	_, err := srv.Write(updateModel, "testing_tid_1234")

	assert.True(t, snsClient.Invoked, "Should have invoked sns Client when no error from DynamoDB")
	assert.Equal(t, SNS_ERROR, err.Error(), "Did not return SNS error.")
}

func TestServiceDelete_Deleted(t *testing.T) {
	ddbClient := MockDynamoDBClient{Happy: true, model: oldModel}
	snsClient := MockSNSClient{Happy: true}
	srv := createService(&ddbClient, &snsClient)
	status, err := srv.Delete(EXPECTED_UUID, "testing_tid_1234")
	assert.NoError(t, err, "Successful deletion should not have returned error")
	assert.Equal(t, db.CONCORDANCE_DELETED, status, "Successul deletion should have returned True.")
}

func TestServiceDelete_NotFound(t *testing.T) {
	ddbClient := MockDynamoDBClient{Happy: true, model: db.ConcordancesModel{}}
	snsClient := MockSNSClient{Happy: true}
	srv := createService(&ddbClient, &snsClient)
	status, err := srv.Delete(EXPECTED_UUID, "testing_tid_1234")
	assert.NoError(t, err, "Successful deletion should not have returned error")
	assert.Equal(t, db.CONCORDANCE_NOT_FOUND, status, "When no record to delete should have returned False.")
}

func TestServiceDelete_DynamoDbError(t *testing.T) {
	ddbClient := MockDynamoDBClient{Happy: false, model: oldModel}
	snsClient := MockSNSClient{Happy: true}
	srv := createService(&ddbClient, &snsClient)

	_, err := srv.Delete(EXPECTED_UUID, "testing_tid_1234")

	assert.Equal(t, DDB_ERROR, err.Error(), "Failed to return service error.")
	assert.False(t, snsClient.Invoked, "Should not have invoked sns Client when error from DynamoDB")
}

func TestServiceDelete_SnsError(t *testing.T) {
	ddbClient := MockDynamoDBClient{Happy: true, model: oldModel}
	snsClient := MockSNSClient{Happy: false}
	srv := createService(&ddbClient, &snsClient)
	updateModel := db.ConcordancesModel{UUID: EXPECTED_UUID, ConcordedIds: []string{"A"}}
	_, err := srv.Write(updateModel, "testing_tid_1234")

	assert.True(t, snsClient.Invoked, "Should have invoked SNS Client when no error from DynamoDB")
	assert.Equal(t, SNS_ERROR, err.Error(), "Did not return SNS error.")
}

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
		return oldModel, nil
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

func (ddb *MockDynamoDBClient) Healthcheck() (error) {
	return nil
}

type MockService struct {
	model   db.ConcordancesModel
	status db.Status
	count   int64
	err     error
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