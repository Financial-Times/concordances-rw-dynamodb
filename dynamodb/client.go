package dynamodbclient

import (
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/aws/session"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
)
const (
	TableHashKeyColumn = "conceptId"
	TableDataColumn = "concordedIds"
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
	awsRegion string
	ddb dynamodbiface.DynamoDBAPI
}

func NewDynamoDbClient(dynamoDbTable string, awsRegion string) DynamoDbClient {
	sess := session.Must(session.NewSession(&aws.Config{Region: aws.String(awsRegion)}))
	ddb := dynamodb.New(sess)
	//fmt.Printf("%s", srv.Endpoint)
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
	fmt.Printf(">>>>>>>> %v", m)
	return m, err
}

func (s *DynamoDbClientImpl) Write(m Model) (model Model, err error) {

	k, err := dynamodbattribute.Marshal(m.UUID)
	l, err := dynamodbattribute.Marshal(m.ConcordedIds)
	//fmt.Printf("\n 1: %v\n", key)
	//fmt.Printf("\n 2: %v\n", L)

	input := &dynamodb.UpdateItemInput{}
	input.SetKey(map[string]*dynamodb.AttributeValue{"conceptId": k})
	input.SetUpdateExpression("SET concordedIds = :concordedIds")
	input.SetReturnValues(dynamodb.ReturnValueAllOld)
	input.SetTableName(s.dynamoDbTable)
	input.SetExpressionAttributeValues(map[string]*dynamodb.AttributeValue {":concordedIds" :l})
	//fmt.Println(input.GoString())
	//input.Validate()

	model = Model{}
	output, err := s.ddb.UpdateItem(input)
	//fmt.Printf("\n len %v \n\n %s\n", len(output.GoString()), output.GoString())
	dynamodbattribute.ConvertFromMap(output.Attributes, &model)
	fmt.Printf(">>>>>>>> %v", model)
	return model, err
}

func (s *DynamoDbClientImpl) Delete(uuid string) (model Model, err error) {
	input := &dynamodb.DeleteItemInput{}
	input.SetReturnValues(dynamodb.ReturnValueAllOld)
	input.SetTableName(s.dynamoDbTable)
	k, err := dynamodbattribute.Marshal(uuid)
	input.SetKey(map[string]*dynamodb.AttributeValue{"conceptId": k})
	fmt.Println(input.GoString())

	output, err :=s.ddb.DeleteItem(input)

	model = Model{}
	dynamodbattribute.ConvertFromMap(output.Attributes, &model)
	fmt.Printf(">>>>>>>> %v", model)
	return model, err
}

func (s *DynamoDbClientImpl) Count() (int64, error) {

	//err := errors.New("Not Implemented")
	return 0, nil
}