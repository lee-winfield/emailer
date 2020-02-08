package main

import (
	"errors"
	"fmt"
	"log"
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

func getDocument(sess *session.Session, fileName string, key string) error {
	fmt.Println("Getting Documents!!!!!!!!!!")
	bucket, err := getParameter(sess, "bill-bucket")
	if err != nil {
		return err
	}

	downloader := s3manager.NewDownloader(sess)
	file, err := os.Create(fileName)
	defer file.Close()
	if err != nil {
		logrus.Errorf("Error creating file: %v\n", err)
		return err
	}
	numBytes, err := downloader.Download(file,
		&s3.GetObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(key),
		})

	if err != nil {
		logrus.Errorf("Error downloading document from s3: %v\n", err)
		return err
	}

	fmt.Println("Downloaded", file.Name(), numBytes, "bytes")

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

func handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	fileName := "/tmp/test.pdf"
	key := "recipient/1001.pdf"
	recipient := "winfieldlee01@gmail.com"

	host := os.Getenv("SMTP_HOST")
	port := os.Getenv("SMTP_PORT")

	sess, err := session.NewSession()
	if err != nil {
		logrus.Errorf("Error creating new aws session: %v\n", err)
		return events.APIGatewayProxyResponse{}, err
	}

	err = getDocument(sess, fileName, key)
	if err != nil {
		logrus.Errorf("Error getting document: %v\n", err)
		return events.APIGatewayProxyResponse{}, err
	}

	from, err := getParameter(sess, "billing-email")
	if err != nil {
		return events.APIGatewayProxyResponse{}, err
	}

	pass, err := getParameter(sess, "billing-email-password")
	if err != nil {
		return events.APIGatewayProxyResponse{}, err
	}

	m := email.NewMessage("Billing", "HI")
	m.From = mail.Address{Name: "From", Address: from}
	m.To = []string{recipient}

	// add attachments
	if err := m.Attach(fileName); err != nil {
		log.Fatal(err)
	}

	// add headers
	m.AddHeader("X-CUSTOMER-id", "xxxxx")

	// send it
	auth := smtp.PlainAuth("", from, pass, host)
	if err := email.Send(fmt.Sprintf("%v:%v", host, port), auth, m); err != nil {
		log.Fatal(err)
	}

	return events.APIGatewayProxyResponse{
		Body:       fmt.Sprintf("Successfully sent email to %v", recipient),
		StatusCode: 200,
	}, nil
}

func main() {
	lambda.Start(handler)
}
