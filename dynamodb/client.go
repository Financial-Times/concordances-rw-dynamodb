package dynamodbclient

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
)

const (
	TableHashKey = "conceptId"
)

type Model struct {
	UUID         string   `json:"conceptId"`
	ConcordedIds []string `json:"concordedIds"`
}

type DynamoDbClient interface {
	Read(uuid string) (Model, error)
	Write(m Model) (Model, error)
	Delete(uuid string) (Model, error)
	Count() (int64, error)
}

type DynamoDbClientImpl struct {
	dynamoDbTable string
	awsRegion     string
	ddb           dynamodbiface.DynamoDBAPI
}

func NewDynamoDbClient(dynamoDbTable string, awsRegion string) DynamoDbClient {
	sess := session.Must(session.NewSession(&aws.Config{Region: aws.String(awsRegion)}))
	ddb := dynamodb.New(sess)
	c := DynamoDbClientImpl{dynamoDbTable: dynamoDbTable, awsRegion: awsRegion, ddb: ddb}
	return &c
}

func (s *DynamoDbClientImpl) Read(uuid string) (Model, error) {
	input := &dynamodb.GetItemInput{}
	input.SetTableName(s.dynamoDbTable)
	k, err := dynamodbattribute.Marshal(uuid)
	input.SetKey(map[string]*dynamodb.AttributeValue{"conceptId": k})
	output, err := s.ddb.GetItem(input)
	m := Model{}
	err = dynamodbattribute.UnmarshalMap(output.Item, &m)
	return m, err
}

func (s *DynamoDbClientImpl) Write(m Model) (model Model, err error) {
	input, err := s.getUpdateInput(m)
	model = Model{}
	output, err := s.ddb.UpdateItem(input)
	dynamodbattribute.ConvertFromMap(output.Attributes, &model)
	return model, err
}
func (s *DynamoDbClientImpl) getUpdateInput(m Model) (*dynamodb.UpdateItemInput, error) {
	input := &dynamodb.UpdateItemInput{}
	k, err := dynamodbattribute.Marshal(m.UUID)
	if err != nil {
		return input, err
	}
	l, err := dynamodbattribute.Marshal(m.ConcordedIds)
	if err != nil {
		return input, err
	}

	input.SetKey(map[string]*dynamodb.AttributeValue{TableHashKey: k})
	input.SetUpdateExpression("SET concordedIds = :concordedIds")
	input.SetReturnValues(dynamodb.ReturnValueAllOld)
	input.SetTableName(s.dynamoDbTable)
	input.SetExpressionAttributeValues(map[string]*dynamodb.AttributeValue{":concordedIds": l})
	return input, nil
}

func (s *DynamoDbClientImpl) Delete(uuid string) (model Model, err error) {
	input := &dynamodb.DeleteItemInput{}
	input.SetReturnValues(dynamodb.ReturnValueAllOld)
	input.SetTableName(s.dynamoDbTable)
	k, err := dynamodbattribute.Marshal(uuid)
	input.SetKey(map[string]*dynamodb.AttributeValue{TableHashKey: k})

	output, err := s.ddb.DeleteItem(input)

	model = Model{}
	dynamodbattribute.ConvertFromMap(output.Attributes, &model)
	return model, err
}

func (s *DynamoDbClientImpl) Count() (int64, error) {

	//err := errors.New("Not Implemented")
	return 0, nil
}
