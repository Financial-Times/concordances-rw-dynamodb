package dynamodb

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	log "github.com/Sirupsen/logrus"
	"strings"
)

const (
	TableHashKey = "conceptId"
)

type Status int

const (
	CONCORDANCE_CREATED   Status = iota
	CONCORDANCE_DELETED
	CONCORDANCE_NOT_FOUND
	CONCORDANCE_UPDATED
	CONCORDANCE_ERROR
)

type ConcordancesModel struct {
	UUID         string   `json:"uuid"`
	ConcordedIds []string `json:"concordedIds"`
}

type DynamoConcordancesModel struct {
	UUID         string   `json:"conceptId"`
	ConcordedIds []string `json:"concordedIds"`
}

type Clienter interface {
	Read(uuid string, transactionId string) (ConcordancesModel, error)
	Write(m ConcordancesModel, transactionId string) (Status, error)
	Delete(uuid string, transactionId string) (Status, error)
	Healthcheck() (error)
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

func (s *Client) Read(uuid string, transactionId string) (ConcordancesModel, error) {
	m := DynamoConcordancesModel{}
	input := &dynamodb.GetItemInput{}
	input.SetTableName(s.dynamoDbTable)
	k, err := dynamodbattribute.Marshal(uuid)

	if err != nil {
		log.WithError(err).WithFields(log.Fields{"UUID": uuid, "transaction_id": transactionId}).Error("Error marshalling UUID to get the key for the concordance")
		return ConcordancesModel{}, err
	}

	input.SetKey(map[string]*dynamodb.AttributeValue{"conceptId": k})
	output, err := s.ddb.GetItem(input)

	if err != nil {
		log.WithError(err).WithFields(log.Fields{"UUID": uuid, "transaction_id": transactionId}).Error("Error Getting Concordance Record")
		return ConcordancesModel{}, err
	}

	if output.Item == nil {
		log.WithFields(log.Fields{"UUID": uuid, "transaction_id": transactionId}).Info("No concordance record was found")
	}

	err = dynamodbattribute.UnmarshalMap(output.Item, &m)

	if err != nil {
		log.WithError(err).WithFields(log.Fields{"UUID": uuid, "transaction_id": transactionId}).Error("Error unmarshalling the response to reading a concordance record")
		return ConcordancesModel{}, err
	}

	return ConcordancesModel{m.UUID, m.ConcordedIds}, err
}

func (s *Client) Write(m ConcordancesModel, transactionId string) (updateStatus Status, err error) {
	input, err := s.getUpdateInput(m)
	model := DynamoConcordancesModel{}
	output, err := s.ddb.UpdateItem(input)
	if err != nil {
		log.WithError(err).WithFields(log.Fields{"UUID": m.UUID, "ConcordedIds": strings.Join(m.ConcordedIds, ", "), "transaction_id": transactionId}).Error("Error Getting Concordance Record")
		return CONCORDANCE_ERROR, err
	}

	err = dynamodbattribute.UnmarshalMap(output.Attributes, &model)
	if err != nil {
		log.WithError(err).WithFields(log.Fields{"UUID": m.UUID, "ConcordedIds": strings.Join(m.ConcordedIds, ", "), "transaction_id": transactionId}).Error("Error unmarshalling the response to writing Concordance Record")
		return CONCORDANCE_ERROR, err
	}

	if model.UUID != "" {
		log.WithFields(log.Fields{"UUID": m.UUID, "ConcordedIds": strings.Join(m.ConcordedIds, ", "), "transaction_id": transactionId}).Info("Concordance updated")
		return CONCORDANCE_UPDATED, nil
	} else {
		log.WithFields(log.Fields{"UUID": m.UUID, "ConcordedIds": strings.Join(m.ConcordedIds, ", "), "transaction_id": transactionId}).Info("Concordance created")
		return CONCORDANCE_CREATED, nil
	}
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

func (s *Client) Delete(uuid string, transactionId string) (status Status, err error) {
	model := DynamoConcordancesModel{}

	input := &dynamodb.DeleteItemInput{}
	input.SetReturnValues(dynamodb.ReturnValueAllOld)
	input.SetTableName(s.dynamoDbTable)

	k, err := dynamodbattribute.Marshal(uuid)
	if err != nil {
		log.WithError(err).WithFields(log.Fields{"UUID": uuid, "transaction_id": transactionId}).Error("Error marshalling UUID to Dynamo Key for Deletion of a concordance")
		return CONCORDANCE_ERROR, err
	}

	input.SetKey(map[string]*dynamodb.AttributeValue{TableHashKey: k})
	output, err := s.ddb.DeleteItem(input)
	if err != nil {
		log.WithError(err).WithFields(log.Fields{"UUID": uuid, "transaction_id": transactionId}).Error("Error Deleting Concordance")
		return CONCORDANCE_ERROR, err
	}

	err = dynamodbattribute.UnmarshalMap(output.Attributes, &model)
	if err != nil {
		log.WithError(err).WithFields(log.Fields{"UUID": uuid, "transaction_id": transactionId}).Error("Error Unmarshalling response from deleting concordance - Unable to ascertain whether the delete was a delete/not found")
		return CONCORDANCE_ERROR, err
	}

	if model.UUID != "" {
		return CONCORDANCE_DELETED, nil
	} else {
		return CONCORDANCE_NOT_FOUND, nil
	}
}

func (s *Client) Healthcheck() (error) {
	_, err := s.ddb.DescribeTable(&dynamodb.DescribeTableInput{TableName: &s.dynamoDbTable})
	return err
}