package concordances

import (
	db "github.com/Financial-Times/concordances-rw-dynamodb/dynamodb"
	"github.com/Financial-Times/concordances-rw-dynamodb/sns"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

func createService(ddbClient db.Client, snsClient sns.Client) ConcordancesRwService {
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

	m, err := srv.Read(EXPECTED_UUID)
	assert.NoError(t, err, "Failed on service error.")
	assert.True(t, reflect.DeepEqual(oldModel, m), "Model did not match.")
	assert.False(t, snsClient.Invoked, "Should not send sns notifications on read")
}

func TestServiceRead_DynamoBbError(t *testing.T) {
	ddbClient := MockDynamoDBClient{Happy: false}
	snsClient := MockSNSClient{}
	srv := createService(&ddbClient, &snsClient)
	_, err := srv.Read(EXPECTED_UUID)
	assert.Equal(t, DDB_ERROR, err.Error(), "Failed to return service error.")
	assert.False(t, snsClient.Invoked, "Should not send sns notifications on read")
}

func TestServiceCreate_NoError(t *testing.T) {
	ddbClient := MockDynamoDBClient{Happy: true, model: db.ConcordancesModel{}}
	snsClient := MockSNSClient{Happy: true}
	srv := createService(&ddbClient, &snsClient)
	updateModel := db.ConcordancesModel{UUID: EXPECTED_UUID, ConcordedIds: []string{"A"}}
	created, err := srv.Write(updateModel)
	assert.NoError(t, err, "Failed on service error.")
	assert.True(t, created, "Did not detect that new record was created.")
	assert.True(t, snsClient.Invoked, "Did not envoke sns Client")
}

func TestServiceUpdate_NoError(t *testing.T) {
	ddbClient := MockDynamoDBClient{Happy: true, model: oldModel}
	snsClient := MockSNSClient{Happy: true}
	srv := createService(&ddbClient, &snsClient)
	updateModel := db.ConcordancesModel{UUID: EXPECTED_UUID, ConcordedIds: []string{"A"}}
	created, err := srv.Write(updateModel)
	assert.NoError(t, err, "Failed on service error.")
	assert.False(t, created, "Did not detect that record was updated.")
	assert.True(t, snsClient.Invoked, "Did not envoke sns Client")
}

func TestServiceWrite_DynamoDbError(t *testing.T) {
	ddbClient := MockDynamoDBClient{Happy: false, model: oldModel}
	snsClient := MockSNSClient{Happy: true}
	srv := createService(&ddbClient, &snsClient)
	updateModel := db.ConcordancesModel{UUID: EXPECTED_UUID, ConcordedIds: []string{"A"}}
	created, err := srv.Write(updateModel)
	assert.Equal(t, DDB_ERROR, err.Error(), "Failed to return service error.")
	assert.False(t, created, "Did not detect existing record was updated.")
	assert.False(t, snsClient.Invoked, "Should not have invoked sns Client when error from DynamoDB")
}

func TestServiceWrite_SnsError(t *testing.T) {
	ddbClient := MockDynamoDBClient{Happy: true, model: db.ConcordancesModel{}}
	snsClient := MockSNSClient{Happy: false}
	srv := createService(&ddbClient, &snsClient)

	_, err := srv.Write(updateModel)

	assert.True(t, snsClient.Invoked, "Should have invoked sns Client when no error from DynamoDB")
	assert.Equal(t, SNS_ERROR, err.Error(), "Did not return SNS error.")
}

func TestServiceDelete_Deleted(t *testing.T) {
	ddbClient := MockDynamoDBClient{Happy: true, model: oldModel}
	snsClient := MockSNSClient{Happy: true}
	srv := createService(&ddbClient, &snsClient)
	deleted, err := srv.Delete(EXPECTED_UUID)
	assert.NoError(t, err, "Successful deletion should not have returned error")
	assert.True(t, deleted, "Successul deletion should have returned True.")
}

func TestServiceDelete_NotFound(t *testing.T) {
	ddbClient := MockDynamoDBClient{Happy: true, model: db.ConcordancesModel{}}
	snsClient := MockSNSClient{Happy: true}
	srv := createService(&ddbClient, &snsClient)
	deleted, err := srv.Delete(EXPECTED_UUID)
	assert.NoError(t, err, "Successful deletion should not have returned error")
	assert.False(t, deleted, "When no record to delete should have returned False.")
}

func TestServiceDelete_DynamoDbError(t *testing.T) {
	ddbClient := MockDynamoDBClient{Happy: false, model: oldModel}
	snsClient := MockSNSClient{Happy: true}
	srv := createService(&ddbClient, &snsClient)

	_, err := srv.Delete(EXPECTED_UUID)

	assert.Equal(t, DDB_ERROR, err.Error(), "Failed to return service error.")
	assert.False(t, snsClient.Invoked, "Should not have invoked sns Client when error from DynamoDB")
}

func TestServiceDelete_SnsError(t *testing.T) {
	ddbClient := MockDynamoDBClient{Happy: true, model: oldModel}
	snsClient := MockSNSClient{Happy: false}
	srv := createService(&ddbClient, &snsClient)
	updateModel := db.ConcordancesModel{UUID: EXPECTED_UUID, ConcordedIds: []string{"A"}}
	_, err := srv.Write(updateModel)

	assert.True(t, snsClient.Invoked, "Should have invoked SNS Client when no error from DynamoDB")
	assert.Equal(t, SNS_ERROR, err.Error(), "Did not return SNS error.")
}
