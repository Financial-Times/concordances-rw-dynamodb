package concordances

import (
	"testing"
	"github.com/stretchr/testify/assert"
	db "github.com/Financial-Times/concordances-rw-dynamodb/dynamodb"
	"github.com/Financial-Times/concordances-rw-dynamodb/sns"
	"reflect"
)

func creatService(ddbClient db.DynamoDbClient, snsClient sns.SnsClient) ConcordancesRwService {
	return ConcordancesRwService{
		DynamoDbTable: "TestTable",
		AwsRegion: "TestRegion",
		ddb: ddbClient,
		sns: snsClient,
	}
}

func TestServiceRead_NoError(t *testing.T) {
	ddbClient := MockDynamoDbClient{Happy: true}
	snsClient := MockSnsClient{}
	srv := creatService(&ddbClient, &snsClient)

	m, err := srv.Read(EXPECTED_UUID)
	assert.NoError(t, err, "Failed on service error.")
	assert.True(t, reflect.DeepEqual(oldModel, m), "Model did not match.")
	assert.False(t, snsClient.Invoked, "Should not send sns notifications on read")
}

func TestServiceRead_DynamoBbError(t *testing.T) {
	ddbClient := MockDynamoDbClient{Happy: false}
	snsClient := MockSnsClient{}
	srv := creatService(&ddbClient, &snsClient)
	_, err := srv.Read(EXPECTED_UUID)
	assert.Equal(t, DDB_ERROR, err.Error(), "Failed to return service error.")
	assert.False(t, snsClient.Invoked, "Should not send sns notifications on read")
}

func TestServiceCreate_NoError(t *testing.T) {
	ddbClient := MockDynamoDbClient{Happy: true, model: db.Model{}}
	snsClient := MockSnsClient{Happy: true}
	srv := creatService(&ddbClient, &snsClient)
	updateModel := db.Model{UUID: EXPECTED_UUID, ConcordedIds: []string{"A"}}
	created, err := srv.Write(updateModel)
	assert.NoError(t, err, "Failed on service error.")
	assert.True(t, created, "Did not detect that new record was created.")
	assert.True(t, snsClient.Invoked, "Did not envoke sns Client")
}

func TestServiceUpdate_NoError(t *testing.T) {
	ddbClient := MockDynamoDbClient{Happy: true, model: oldModel}
	snsClient := MockSnsClient{Happy: true}
	srv := creatService(&ddbClient, &snsClient)
	updateModel := db.Model{UUID: EXPECTED_UUID, ConcordedIds: []string{"A"}}
	created, err := srv.Write(updateModel)
	assert.NoError(t, err, "Failed on service error.")
	assert.False(t, created, "Did not detect that record was updated.")
	assert.True(t, snsClient.Invoked, "Did not envoke sns Client")
}

func TestServiceWrite_DynamoDbError(t *testing.T) {
	ddbClient := MockDynamoDbClient{Happy: false, model: oldModel}
	snsClient := MockSnsClient{Happy: true}
	srv := creatService(&ddbClient, &snsClient)
	updateModel := db.Model{UUID: EXPECTED_UUID, ConcordedIds: []string{"A"}}
	created, err := srv.Write(updateModel)
	assert.Equal(t, DDB_ERROR, err.Error(), "Failed to return service error.")
	assert.False(t, created, "Did not detect existing record was updated.")
	assert.False(t, snsClient.Invoked, "Should not have invoked sns Client when error from DynamoDB")
}

func TestServiceWrite_SnsError(t *testing.T) {
	ddbClient := MockDynamoDbClient{Happy: true, model: db.Model{}}
	snsClient := MockSnsClient{Happy: false}
	srv := creatService(&ddbClient, &snsClient)

	_, err := srv.Write(updateModel)

	assert.True(t, snsClient.Invoked, "Should have invoked sns Client when no error from DynamoDB")
	assert.Equal(t, SNS_ERROR, err.Error(), "Did not return SNS error.")
}

func TestServiceDelete_Deleted(t *testing.T) {
	ddbClient := MockDynamoDbClient{Happy: true, model: oldModel}
	snsClient := MockSnsClient{Happy: true}
	srv := creatService(&ddbClient, &snsClient)
	deleted, err := srv.Delete(EXPECTED_UUID)
	assert.NoError(t, err, "Successful deletion should not have returned error")
	assert.True(t, deleted, "Successul deletion should have returned True.")
}

func TestServiceDelete_NotFound(t *testing.T) {
	ddbClient := MockDynamoDbClient{Happy: true, model: db.Model{}}
	snsClient := MockSnsClient{Happy: true}
	srv := creatService(&ddbClient, &snsClient)
	deleted, err := srv.Delete(EXPECTED_UUID)
	assert.NoError(t, err, "Successful deletion should not have returned error")
	assert.False(t, deleted, "When no record to delete should have returned False.")
}

func TestServiceDelete_DynamoDbError(t *testing.T) {
	ddbClient := MockDynamoDbClient{Happy: false, model: oldModel}
	snsClient := MockSnsClient{Happy: true}
	srv := creatService(&ddbClient, &snsClient)

	_, err := srv.Delete(EXPECTED_UUID)

	assert.Equal(t, DDB_ERROR, err.Error(), "Failed to return service error.")
	assert.False(t, snsClient.Invoked, "Should not have invoked sns Client when error from DynamoDB")
}

func TestServiceDelete_SnsError(t *testing.T) {
	ddbClient := MockDynamoDbClient{Happy: true, model: oldModel}
	snsClient := MockSnsClient{Happy: false}
	srv := creatService(&ddbClient, &snsClient)
	updateModel := db.Model{UUID: EXPECTED_UUID, ConcordedIds: []string{"A"}}
	_, err := srv.Write(updateModel)

	assert.True(t, snsClient.Invoked, "Should have invoked SNS Client when no error from DynamoDB")
	assert.Equal(t, SNS_ERROR, err.Error(), "Did not return SNS error.")
}

func TestServiceCount(t *testing.T) {
	ddbClient := MockDynamoDbClient{}
	snsClient := MockSnsClient{}
	srv := creatService(&ddbClient, &snsClient)
	cnt, err := srv.Count()
	var zero int64 = 0
	assert.NoError(t, err, "Count not implemented and always returns 0")
	assert.Equal(t, zero, cnt, "Count not implemented and always returns 0")
}