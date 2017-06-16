package dynamodb

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

const (
	TableHashKey = "conceptId"
)

type ConcordancesModel struct {
	UUID         string   `json:"conceptId"`
	ConcordedIds []string `json:"concordedIds"`
}

type Clienter interface {
	Read(uuid string) (ConcordancesModel, error)
	Write(m ConcordancesModel) (ConcordancesModel, error)
	Delete(uuid string) (ConcordancesModel, error)
}

type Client struct {
	dynamoDbTable string
	awsRegion     string
	ddb           *dynamodb.DynamoDB
}

func NewDynamoDBClient(dynamoDbTable string, awsRegion string) Clienter {
	sess := session.Must(session.NewSession(&aws.Config{Region: aws.String(awsRegion)}))
	ddb := dynamodb.New(sess)
	c := Client{dynamoDbTable: dynamoDbTable, awsRegion: awsRegion, ddb: ddb}
	return &c
}

func (s *Client) Read(uuid string) (ConcordancesModel, error) {
	m := ConcordancesModel{}
	input := &dynamodb.GetItemInput{}
	input.SetTableName(s.dynamoDbTable)
	k, err := dynamodbattribute.Marshal(uuid)

	if err != nil {
		return m, err
	}
	input.SetKey(map[string]*dynamodb.AttributeValue{"conceptId": k})
	output, err := s.ddb.GetItem(input)
	if err != nil {
		return m, err
	}
	if output.Item != nil {
		err = dynamodbattribute.UnmarshalMap(output.Item, &m)
	}
	return m, err
}

func (s *Client) Write(m ConcordancesModel) (model ConcordancesModel, err error) {
	input, err := s.getUpdateInput(m)
	model = ConcordancesModel{}
	output, err := s.ddb.UpdateItem(input)
	if err != nil {
		return model, err
	}
	dynamodbattribute.UnmarshalMap(output.Attributes, &model)
	return model, err
}
func (s *Client) getUpdateInput(m ConcordancesModel) (*dynamodb.UpdateItemInput, error) {
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

func (s *Client) Delete(uuid string) (model ConcordancesModel, err error) {
	model = ConcordancesModel{}
	input := &dynamodb.DeleteItemInput{}
	input.SetReturnValues(dynamodb.ReturnValueAllOld)
	input.SetTableName(s.dynamoDbTable)
	k, err := dynamodbattribute.Marshal(uuid)
	if err != nil {
		return model, err
	}
	input.SetKey(map[string]*dynamodb.AttributeValue{TableHashKey: k})
	output, err := s.ddb.DeleteItem(input)
	dynamodbattribute.UnmarshalMap(output.Attributes, &model)
	return model, err
}