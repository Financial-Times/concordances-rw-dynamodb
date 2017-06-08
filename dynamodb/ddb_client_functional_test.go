package dynamodbclient
import (
	"testing"

	"reflect"
	"fmt"

)

const (
	UUID = "4f50b156-6c50-4693-b835-02f70d3f3bc0"
	DynamoDbTableName = "upp-concordance-store-semantic"
	AWSRegion = "eu-west-1"
)

func TestWriteToDynamoDb(t *testing.T) {
	c := NewDynamoDbClient(DynamoDbTableName, AWSRegion)
	newModel := Model{
		UUID: "4f50b156-6c50-4693-b835-02f70d3f3bc0",
		ConcordedIds: []string{"7c4b3931-361f-4ea4-b694-75d1630d7746", "1e5c86f8-3f38-4b6b-97ce-f75489ac3113"},
	}
	oldModel, _ := c.Write(newModel)

	eq := reflect.DeepEqual(newModel, oldModel)
	if eq {
		fmt.Println("They're equal.")
	} else {
		fmt.Println("They're unequal.")
	}
}

func TestDeleteFromDynamoDb(t *testing.T) {
	c := NewDynamoDbClient(DynamoDbTableName, AWSRegion)
	c.Delete("4f50b156-6c50-4693-b835-02f70d3f3bc0")
}

func TestReadFromDynamoDb(t *testing.T) {
	c := NewDynamoDbClient(DynamoDbTableName, AWSRegion)
	c.Read("4f50b156-6c50-4693-b835-02f70d3f3bc0")
}

