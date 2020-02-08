package main

import (
	"errors"
	"fmt"
	"log"
	"net/smtp"
	"os"

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

func handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	host := os.Getenv("SMTP_HOST")
	port := os.Getenv("SMTP_PORT")

	sess, err := session.NewSession()
	if err != nil {
		logrus.Errorf("Error creating new aws session: %v\n", err)
		return events.APIGatewayProxyResponse{}, err
	}

	client := ssm.New(sess)

	be := "billing-email"
	f, err := client.GetParameter(&ssm.GetParameterInput{Name: &be})
	if err != nil {
		logrus.Errorf("Error getting parameter: %v\n", err)
		return events.APIGatewayProxyResponse{}, err
	}

	bep := "billing-email-password"
	p, err := client.GetParameter(&ssm.GetParameterInput{Name: &bep})
	if err != nil {
		logrus.Errorf("Error getting parameter: %v\n", err)
		return events.APIGatewayProxyResponse{}, err
	}

	from := *f.Parameter.Value
	pass := *p.Parameter.Value
	recipient := "winfieldlee01@gmail.com"

	msg := "From: " + from + "\n" +
		"To: " + recipient + "\n" +
		"Subject: Hello there!\n\n" +
		"new body"

	err = smtp.SendMail(fmt.Sprintf("%v:%v", host, port),
		smtp.PlainAuth("", from, pass, host),
		from, []string{recipient}, []byte(msg))

	if err != nil {
		log.Printf("smtp error: %s", err)
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
