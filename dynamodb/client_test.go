package dynamodb

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
	"log"
	"fmt"
)

const (
	UUID         = "4f50b156-6c50-4693-b835-02f70d3f3bc0"
	DDB_TABLE    = "upp-concordance-store-test"
	AWS_REGION   = "eu-west-1"
	DDB_ENDPOINT = "http://localhost:8000"
)

var goodModel = ConcordancesModel{
	UUID:         UUID,
	ConcordedIds: []string{"7c4b3931-361f-4ea4-b694-75d1630d7746", "1e5c86f8-3f38-4b6b-97ce-f75489ac3113"},
}

var db *dynamodb.DynamoDB
var c Client

var DescribeTableParams = &dynamodb.DescribeTableInput{TableName: aws.String(DDB_TABLE)}

func init() {
	log.Println("Create DynamoDb")
	db = setupDynamoDBLocal()
	c = Client{dynamoDbTable: DDB_TABLE, awsRegion: AWS_REGION, ddb: db}
}

func setupTestCase(t *testing.T) func(t *testing.T) {
	t.Log("Create table \n")
	c = Client{dynamoDbTable: DDB_TABLE, awsRegion: AWS_REGION, ddb: db}
	err := createTableIfNotExists(t)

	assert.NoError(t, err, "Unexpected error creating table")
	return func(t *testing.T) {
		errs := deleteTableIfExists(t)
		assert.NoError(t, errs, "Unexpected error creating table")
		t.Log("Destroy Table \n")
	}
}

func deleteTableIfExists(t *testing.T) error {
	input := &dynamodb.DeleteTableInput{TableName: aws.String(DDB_TABLE)}
	_, err := db.DeleteTable(input)
	if err == nil {
		return nil
	} else if err.(awserr.Error).Code() != dynamodb.ErrCodeResourceNotFoundException {
		assert.Fail(t, "Failed to delete table. ", err.Error())
	} else {
		t.Log("Table doesn't exist")
	}
	return err
}

func setupDynamoDBLocal() *dynamodb.DynamoDB {
	t := &testing.T{}
	assert := assert.New(t)
	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(AWS_REGION),
		Endpoint:    aws.String(DDB_ENDPOINT),
		Credentials: credentials.NewEnvCredentials(),
	})
	assert.NoError(err, "Should be able to create a session talking to local DynamoDB. Make sure this is running")
	ddb := dynamodb.New(sess)
	return ddb
}

func createTableIfNotExists(t *testing.T) error {
	_, err := db.DescribeTable(DescribeTableParams)
	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			if awsErr.Code() == dynamodb.ErrCodeResourceNotFoundException {
				params := &dynamodb.CreateTableInput{
					AttributeDefinitions: []*dynamodb.AttributeDefinition{ // Required
						{ // Required
							AttributeName: aws.String(TableHashKey),                  // Required
							AttributeType: aws.String(dynamodb.ScalarAttributeTypeS), // Required
						},
						// More values...
					},
					KeySchema: []*dynamodb.KeySchemaElement{ // Required
						{ // Required
							AttributeName: aws.String(TableHashKey),         // Required
							KeyType:       aws.String(dynamodb.KeyTypeHash), // Required
						},
						// More values...
					},
					ProvisionedThroughput: &dynamodb.ProvisionedThroughput{ // Required
						ReadCapacityUnits:  aws.Int64(5), // Required
						WriteCapacityUnits: aws.Int64(5), // Required
					},
					TableName: aws.String(DDB_TABLE), // Required
				}
				_, err := db.CreateTable(params)
				assert.NoError(t, err, "Unable to create the table")
			}
		} else {
			assert.Fail(t, fmt.Sprintf("Failed to connect to local DynamoDB. Error: %s", err.Error()))
			return err
		}

	}
	return nil
}

func TestUpdateInputIsValid(t *testing.T) {
	tearDownTestCase := setupTestCase(t)
	defer tearDownTestCase(t)

	input, err := c.getUpdateInput(goodModel)

	assert.NoError(t, err, "Received error")
	assert.NoError(t, input.Validate(), "Update Input is valid.")
}

func TestCreateConcordance(t *testing.T) {
	tearDownTestCase := setupTestCase(t)
	defer tearDownTestCase(t)

	status, err := c.Write(goodModel)

	assert.NoError(t, err, "Failed to write concordance.")
	assert.Equal(t, status, CONCORDANCE_CREATED)
	newModel, err := c.Read(UUID)
	assert.True(t, reflect.DeepEqual(goodModel, newModel), "Failed to create concordance record")
}

func TestUpdateConcordance(t *testing.T) {
	tearDownTestCase := setupTestCase(t)
	defer tearDownTestCase(t)

	_, err := c.Write(goodModel)
	assert.NoError(t, err, "Failed to write concordance.")
	newModel := ConcordancesModel{
		UUID:         "4f50b156-6c50-4693-b835-02f70d3f3bc0",
		ConcordedIds: []string{"7c4b3931-361f-4ea4-b694-75d1630d7746"},
	}
	status, err := c.Write(newModel)

	updatedModel, err := c.Read(UUID)

	assert.Equal(t, status, CONCORDANCE_UPDATED)
	assert.True(t, reflect.DeepEqual(newModel, updatedModel), "Failed to update concordance record")
}

func TestDeleteExistingConcordance(t *testing.T) {
	tearDownTestCase := setupTestCase(t)
	defer tearDownTestCase(t)

	_, err := c.Write(goodModel)
	assert.NoError(t, err, "Failed to set up concordance to be deleted")

	status, err := c.Delete(UUID)

	assert.NoError(t, err, "Deletion operation resulted in error.")
	assert.Equal(t, status, CONCORDANCE_DELETED,  "Unexpected status on deleting existing concordance")
}

func TestDeleteNonExistingConcordance(t *testing.T) {
	tearDownTestCase := setupTestCase(t)
	defer tearDownTestCase(t)

	status, err := c.Delete(UUID)

	assert.NoError(t, err, "Deletion operation resulted in error.")
	assert.Equal(t, status, CONCORDANCE_NOT_FOUND, "Unexpected status, expected to not find a concordance")
}

func TestReadExistingConcordance(t *testing.T) {
	tearDownTestCase := setupTestCase(t)
	defer tearDownTestCase(t)

	_, err := c.Write(goodModel)
	assert.NoError(t, err, "failed to set up concordance to be read.")

	model, err := c.Read(UUID)

	assert.NoError(t, err, "Retrieving concordance resulted in error.")
	assert.True(t, reflect.DeepEqual(goodModel, model), "Failed to retrive old concordance record")
}

func TestReadNonExistingConcordance(t *testing.T) {
	tearDownTestCase := setupTestCase(t)
	defer tearDownTestCase(t)

	model, err := c.Read(UUID)

	assert.NoError(t, err, "Retrieving concordance resulted in error.")
	assert.Empty(t, model.UUID, "Failed to retrive old concordance record upon deletion")
	assert.Empty(t, model.ConcordedIds, "Failed to retrive old concordance record upon deletion")
}

func TestClient_Healthcheck(t *testing.T) {
	tearDownTestCase := setupTestCase(t)
	defer tearDownTestCase(t)

	err := c.Healthcheck()

	assert.NoError(t, err, "Unexpected error occurred in healthcheck")
}