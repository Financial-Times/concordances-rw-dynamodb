package sns

import (
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sns/snsiface"
	"github.com/aws/aws-sdk-go/aws"
	"fmt"
	"strings"
	log "github.com/Sirupsen/logrus"
)

const (
	//Message format expected by consumers
	SNS_MSG = (`{"Records":[{"s3":{"object":{"key":"%s"}}}]}`)
)

type SnsClient interface {
	SendMessage(uuid string) error
}

type SnsClientImpl struct {
	client snsiface.SNSAPI
	topicArn string
	awsRegion string
}

func NewSnsClient(topic string, region string) *SnsClientImpl {
	sess := session.Must(session.NewSession(&aws.Config{Region: aws.String(region)}))
	svc := sns.New(sess)
	snsClient := SnsClientImpl{client: svc, topicArn: topic, awsRegion: region}
	return &snsClient
}

func (c *SnsClientImpl) message(uuid string) *string {
	n := strings.Replace(uuid, "-", "/", -1)
	m := fmt.Sprintf(SNS_MSG, n)
	return aws.String(m)
}

func (c *SnsClientImpl) SendMessage(uuid string) (err error) {

	params := &sns.PublishInput{
		Message: c.message(uuid), // This is the message itself (can be XML / JSON / Text - anything you want)
		TopicArn: aws.String(c.topicArn),
	}
	resp, err := c.client.Publish(params)   //Call to publish the message

	if resp != nil {
		log.Infof("Concordance Notification for concept uuid [%s] was posted to topic[%s]. SNS response: %s", uuid, c.topicArn, resp.String())
	}
	return err
}

