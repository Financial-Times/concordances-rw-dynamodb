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

type SNSClient interface {
	SendMessage(uuid string) error
}

type SNSClientImpl struct {
	client    snsiface.SNSAPI
	topicArn  string
	awsRegion string
}

func NewSNSClient(topic string, region string) *SNSClientImpl {
	sess := session.Must(session.NewSession(&aws.Config{Region: aws.String(region)}))
	svc := sns.New(sess)
	snsClient := SNSClientImpl{client: svc, topicArn: topic, awsRegion: region}
	return &snsClient
}

func (c *SNSClientImpl) message(uuid string) *string {
	n := strings.Replace(uuid, "-", "/", -1)
	m := fmt.Sprintf(SNS_MSG, n)
	return aws.String(m)
}

func (c *SNSClientImpl) SendMessage(uuid string) (err error) {

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
