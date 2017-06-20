package concordances

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func getConfig(srv Service) *healthConfig {
	return &healthConfig{
		appSystemCode: "appSystemCode",
		appName:       "appName",
		port:          "port",
		srv:           srv,
	}
}
func TestCheck_DynamoDBIsHealthy(t *testing.T) {
	happyService := &MockService{}
	config := getConfig(happyService)
	healthService := newHealthService(config)
	_, err := healthService.dynamoDbChecker()
	assert.NoError(t, err, "DynamoBD healthcheck failed to detect healthy state")
}

// TODO Fix this
//func TestCheck_DynamoDBIsNotHealthy(t *testing.T) {
//	unhappyService := &MockService{err: errors.New("")}
//	config := getConfig(unhappyService)
//	healthService := newHealthService(config)
//	_, err := healthService.dynamoDbChecker()
//	assert.Error(t, err)
//	assert.Equal(t, DDB_ERROR, err.Error(), "DynamoDB healthcheck failed to detect unhealthy state")
//}

func TestCheck_SnsIsHealthy(t *testing.T) {
	happyService := &MockService{}
	config := getConfig(happyService)
	healthService := newHealthService(config)
	_, err := healthService.snsChecker()
	assert.NoError(t, err, "SNS healthcheck failed to detect healthy state")
}

func TestCheck_SnsIsNotHealthy(t *testing.T) {
	unhappyService := &MockService{err: errors.New("")}
	config := getConfig(unhappyService)
	healthService := newHealthService(config)
	_, err := healthService.snsChecker()
	assert.Equal(t, SNS_ERROR, err.Error(), "SNS healthcheck failed to detect unhealthy state")

}
