package dynamodbclient

import (
	"log"
	"testing"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/stretchr/testify/assert"
	"fmt"
)
const (
	DDB_TABLE = "upp-concordance-store-semantic"
)
func setupDynamoDBLocal(t *testing.T) *dynamodb.DynamoDB {
	assert := assert.New(t)
	sess, err := session.NewSession(&aws.Config{
		Region:   aws.String("us-west-2"),
		Endpoint: aws.String("http://localhost:8000")})
	assert.NoError(err, "Should be able to create a session talking to local DynamoDB. Make sure this is running")
	db := dynamodb.New(sess)
	createTableIfNotExists(t)
	return db
}

func createTableIfNotExists(t *testing.T) {
	assert := assert.New(t)
	 db := setupDynamoDBLocal(t)
		params := &dynamodb.DescribeTableInput{
			TableName: aws.String(DDB_TABLE),
		}

	_, err := db.DescribeTable(params)
	if err != nil {

		if awsErr, ok := err.(awserr.Error); ok {
			log.Println("Error found:", awsErr.Code(), awsErr.Message())
			//TODO what if not this code?
			if awsErr.Code() == dynamodb.ErrCodeResourceNotFoundException {
				params := &dynamodb.CreateTableInput{
					AttributeDefinitions: []*dynamodb.AttributeDefinition{// Required
						{// Required
							AttributeName: aws.String(TableHashKeyColumn), // Required
							AttributeType: aws.String(dynamodb.ScalarAttributeTypeS), // Required
						},
						// More values...
					},
					KeySchema: []*dynamodb.KeySchemaElement{// Required
						{// Required
							AttributeName: aws.String(TableHashKeyColumn), // Required
							KeyType:       aws.String(dynamodb.KeyTypeHash), // Required
						},
						// More values...
					},
					ProvisionedThroughput: &dynamodb.ProvisionedThroughput{// Required
						ReadCapacityUnits:  aws.Int64(5), // Required
						WriteCapacityUnits: aws.Int64(5), // Required
					},
					TableName: aws.String(DDB_TABLE), // Required
				}
				_, err := db.CreateTable(params)
				assert.NoError(t, err)
				//if err != nil {
				//	fmt.Println(err.Error())
				//} else {
				//fmt.Println(out.GoString())
				//}

			}

		} else {
			assert.Fail(fmt.Sprintf("Failed to connect to local DynamoDB. Error: %s", err.Error()))
			return
		}

	}
	//out, err = db.DescribeTable(params)
	//if err != nil {
	//	fmt.Println(err.Error())
	//} else {
	//	fmt.Println(out.GoString())
	//}
}