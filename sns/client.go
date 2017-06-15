package sns

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/aws/aws-sdk-go/service/sns/snsiface"
	"strings"
)

const (
	//Message format expected by consumers
	SNS_MSG = (`{"Records":[{"s3":{"object":{"key":"%s"}}}]}`)
)

type Client interface {
	SendMessage(uuid string) error
	Healthcheck() (bool, error)
}

type SNSClient struct {
	client    snsiface.SNSAPI
	topicArn  string
	awsRegion string
}

func NewSNSClient(topic string, region string) *SNSClient {
	sess := session.Must(session.NewSession(&aws.Config{Region: aws.String(region)}))
	svc := sns.New(sess)
	snsClient := SNSClient{client: svc, topicArn: topic, awsRegion: region}
	return &snsClient
}

func (c *SNSClient) message(uuid string) *string {
	n := strings.Replace(uuid, "-", "/", -1)
	m := fmt.Sprintf(SNS_MSG, n)
	return aws.String(m)
}

func (c *SNSClient) SendMessage(uuid string) (err error) {

	params := &sns.PublishInput{
		Message:  c.message(uuid),
		TopicArn: aws.String(c.topicArn),
	}
	resp, err := c.client.Publish(params)

	if resp != nil {
		log.Infof("Concordance Notification for concept uuid [%s] was posted to topic[%s]. SNS response: %s", uuid, c.topicArn, resp.String())
	}
	return err
}

func (c *SNSClient) Healthcheck() (bool, error) {
	params := &sns.GetTopicAttributesInput {TopicArn: aws.String(c.topicArn)}
	output, err := c.client.GetTopicAttributes(params)
	var attributes map[string]*string = output.Attributes
	if len(attributes) > 0 {
		return true, nil
	}
	return false, err
}