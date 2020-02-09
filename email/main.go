package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/mail"
	"net/smtp"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/scorredoira/email"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/sirupsen/logrus"
)

var (
	// DefaultHTTPGetAddress Default Address
	DefaultHTTPGetAddress = "https://checkip.amazonaws.com"

	// ErrNoIP No IP found in response
	ErrNoIP = errors.New("No IP in HTTP response")

	// ErrNon200Response non 200 status code in response
	ErrNon200Response = errors.New("Non 200 Response found")
)

func getDocument(sess *session.Session, fileName string) error {
	bucket, err := getParameter(sess, "bill-bucket")
	if err != nil {
		return err
	}

	downloader := s3manager.NewDownloader(sess)
	file, err := os.Create(fmt.Sprintf("/tmp/%v", fileName))
	defer file.Close()
	if err != nil {
		logrus.Errorf("Error creating file: %v\n", err)
		return err
	}
	_, err = downloader.Download(file,
		&s3.GetObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(fmt.Sprintf("recipient/%v", fileName)),
		})

	if err != nil {
		logrus.Errorf("Error downloading document from s3: %v\n", err)
		return err
	}

	return nil
}

func getParameter(sess *session.Session, pName string) (string, error) {
	client := ssm.New(sess)
	f, err := client.GetParameter(&ssm.GetParameterInput{Name: &pName})
	if err != nil {
		logrus.Errorf("Error getting parameter: %v\n", err)
		return "", err
	}

	return *f.Parameter.Value, nil
}

func sendEmail(sess *session.Session, recipient, subject, fileName string) error {
	host := os.Getenv("SMTP_HOST")
	port := os.Getenv("SMTP_PORT")

	// get parameters
	from, err := getParameter(sess, "billing-email")
	if err != nil {
		return err
	}

	pass, err := getParameter(sess, "billing-email-password")
	if err != nil {
		return err
	}

	// create message
	m := email.NewMessage(subject, "")
	m.From = mail.Address{Name: "Claudia Winfield", Address: from}
	m.To = []string{recipient}

	// add attachments
	if err := m.Attach(fmt.Sprintf("/tmp/%v", fileName)); err != nil {
		logrus.Fatal(err)
	}

	// add headers
	m.AddHeader("X-CUSTOMER-id", "xxxxx")

	// send it
	auth := smtp.PlainAuth("", from, pass, host)
	if err := email.Send(fmt.Sprintf("%v:%v", host, port), auth, m); err != nil {
		logrus.Fatal(err)
		return err
	}

	return nil
}

type EmailEvent struct {
	FileName  string `json:"fileName"`
	Recipient string `json:"recipient"`
	Subject   string `json:"subject"`
}

func handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	body := request.Body
	var event EmailEvent
	json.Unmarshal([]byte(body), &event)

	fileName := event.FileName
	recipient := event.Recipient
	subject := event.Subject

	sess, err := session.NewSession()
	if err != nil {
		logrus.Errorf("Error creating new aws session: %v\n", err)
		return events.APIGatewayProxyResponse{}, err
	}

	err = getDocument(sess, fileName)
	if err != nil {
		logrus.Errorf("Error getting document: %v\n", err)
		return events.APIGatewayProxyResponse{}, err
	}

	err = sendEmail(sess, recipient, subject, fileName)
	if err != nil {
		logrus.Errorf("Error sending email: %v\n", err)
		return events.APIGatewayProxyResponse{}, err
	}

	return events.APIGatewayProxyResponse{
		Body:       fmt.Sprintf("Successfully sent email to %v", recipient),
		StatusCode: 200,
	}, nil
}

func main() {
	lambda.Start(handler)
}
