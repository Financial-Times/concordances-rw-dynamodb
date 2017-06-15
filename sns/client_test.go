package sns

import (
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/aws/aws-sdk-go/service/sns/snsiface"
	"github.com/stretchr/testify/assert"
	"testing"
	"errors"
)

const (
	AWS_REGION      = "eu-west-1"
	TOPIC           = "arn:aws:sns:eu-west-1:027104099916:upp-concordance-semantic-SNSTopic-SCOTT1234"
	UUID            = "9b40e89c-e87b-3d4f-b72c-2cf7511d2146"
	ExpectedMessage = `{"Records":[{"s3":{"object":{"key":"9b40e89c/e87b/3d4f/b72c/2cf7511d2146"}}}]}`
)

type AssertPublishInput struct {
	snsiface.SNSAPI
	tT *testing.T
	happy bool
}

func (c AssertPublishInput) Publish(in *sns.PublishInput) (*sns.PublishOutput, error) {
	assert.Equal(c.tT, *in.Message, ExpectedMessage, "Did not pass message body to PublishInput to sent to SNS")
	assert.Equal(c.tT, *in.TopicArn, TOPIC, "Did not pass topic name to PublishInput to sent to SNS")
	return nil, nil
}

func (c AssertPublishInput) GetTopicAttributes(input *sns.GetTopicAttributesInput) (*sns.GetTopicAttributesOutput, error) {

	output := sns.GetTopicAttributesOutput{}

	if c.happy {
		var topicARN string =  "topicARN"
		attributes := map[string]*string{"TopicArn":&topicARN}
		output.SetAttributes(attributes)
		return &output, nil
	}
	return  &output, errors.New("SNS is unhappy.")
}

func TestMessageFormattedCorrectly(t *testing.T) {
	mockSnsService := AssertPublishInput{}
	client := SNSClient{client: &mockSnsService, topicArn: TOPIC, awsRegion: AWS_REGION}
	actualMessage := client.message(UUID)
	assert.Equal(t, ExpectedMessage, *actualMessage, "Expected and Actual messages did not match.")
}

func TestPublishInputHasData(t *testing.T) {
	mockSnsService := AssertPublishInput{tT: t}
	client := SNSClient{client: &mockSnsService, topicArn: TOPIC, awsRegion: AWS_REGION}
	err := client.SendMessage(UUID)
	assert.NoError(t, err, "Received error")
}

func TestSNSClient_HealthcheckHappy(t *testing.T) {
	mockSnsService := AssertPublishInput{tT: t, happy: true}
	client := SNSClient{client: &mockSnsService, topicArn: TOPIC, awsRegion: AWS_REGION}
	happy, err := client.Healthcheck()
	assert.NoError(t, err, "Received error.")
	assert.True(t, happy, "SNS is not happy.")
}

func TestSNSClient_HealthcheckUnhappy(t *testing.T) {
	mockSnsService := AssertPublishInput{tT: t, happy: false}
	client := SNSClient{client: &mockSnsService, topicArn: TOPIC, awsRegion: AWS_REGION}
	happy, err := client.Healthcheck()
	assert.NotEmpty(t, err.Error())
	assert.False(t, happy)
}