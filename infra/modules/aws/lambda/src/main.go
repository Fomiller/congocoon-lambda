package main

import (
	"bytes"
	"context"
	"fmt"
	"hash/fnv"
	"io/ioutil"
	"log"
	"net/http"
	"net/smtp"
	"os"
	"strings"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
)

type Event struct {
	Name string `json:"name"`
}

type Mail struct {
	Sender  string
	To      []string
	Subject string
	Body    string
}

func exitErrorf(msg string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, msg+"\n", args...)
	os.Exit(1)
}

func main() {
	lambda.Start(HandleRequest)
}

func HandleRequest(ctx context.Context, event Event) {
	log.Printf("EVENT: %s", event)

	s3_session := s3.New(session.New())
	remoteHash, err := getHTMLHash("http://congocoon.com/kittensforsale.html")
	if err != nil {
		panic(err)
	}

	// list bucket objects
	res, err := s3_session.ListObjectsV2(&s3.ListObjectsV2Input{Bucket: aws.String("fomiller-dev")})
	if err != nil {
		exitErrorf("Unable to list items in bucket %q, %v", "fomiller-dev", err)
	}

	// print out bucket objects
	for _, item := range res.Contents {
		log.Println("Name: ", *item.Key)
	}

	// check if file in bucket objects if not create the object
	if containsFile(res.Contents, "congocoon.hash") {
		log.Println("Found file in s3")
	} else {
		log.Printf("object with key %s does not exist in bucket %s", "congocoon.hash", "fomiller-dev")
		err := putFile(s3_session, remoteHash)
		if err != nil {
			panic(err)
		}
		log.Printf("created object with key %s in bucket %s", "congocoon.hash", "fomiller-dev")
	}

	// get the currently stored version in s3
	fileObject, err := getFile("congocoon.hash", s3_session)
	if err != nil {
		// For information on other S3 API error codes see:
		// http://docs.aws.amazon.com/AmazonS3/latest/API/ErrorResponses.html
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case s3.ErrCodeNoSuchBucket:
				exitErrorf("bucket %s does not exist", "fomiller-dev")
			case s3.ErrCodeNoSuchKey:
				exitErrorf("object with key %s does not exist in bucket %s", "congocoon.hash", "fomiller-dev")
			}
		}
	}

	// read file from s3
	buf := new(bytes.Buffer)
	buf.ReadFrom(fileObject.Body)
	storedHash := buf.String()

	if storedHash != fmt.Sprint(remoteHash) {
		log.Println("THERE HAVE BEEN CHANGES TO THE PAGE")
		log.Printf("REMOTE HASH: %v", remoteHash)
		log.Printf("STORED HASH: %v", storedHash)
		err := putFile(s3_session, remoteHash)
		if err != nil {
			exitErrorf("ERROR updating file in s3 %v", err)
		}
		sendEmail()
	} else {
		log.Println("NO CHANGES HAVE OCCURED")
		log.Printf("REMOTE HASH: %v", remoteHash)
		log.Printf("STORED HASH: %v", storedHash)
	}
}

func putFile(s3_client *s3.S3, remoteHash uint32) error {
	params := &s3.PutObjectInput{
		Bucket:        aws.String("fomiller-dev"),
		Key:           aws.String("congocoon.hash"),
		ACL:           aws.String("bucket-owner-full-control"),
		Body:          bytes.NewReader([]byte(fmt.Sprint(remoteHash))),
		ContentLength: aws.Int64(int64(len(fmt.Sprint(remoteHash)))),
	}
	_, err := s3_client.PutObject(params)
	return err
}

func getHTMLHash(endpoint string) (uint32, error) {
	res, err := http.Get(endpoint)
	if err != nil {
		return 0, err
	}

	defer res.Body.Close()
	html, err := ioutil.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}
	remoteHash := hash(string(html))
	return remoteHash, nil
}

func hash(s string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(s))
	return h.Sum32()
}

func getFile(fileName string, s3_client *s3.S3) (*s3.GetObjectOutput, error) {
	fileObject, err := s3_client.GetObject(
		&s3.GetObjectInput{
			Bucket: aws.String("fomiller-dev"),
			Key:    aws.String("congocoon.hash"),
		},
	)
	return fileObject, err
}

func containsFile(s []*s3.Object, str string) bool {
	for _, v := range s {
		if *v.Key == str {
			return true
		}
	}
	return false
}

func sendEmail() {
	// from is senders email address

	secret := aws.String("fomiller-gmail-api-key")
	sess := session.Must(session.NewSession())
	sm := secretsmanager.New(sess, aws.NewConfig().WithRegion("us-east-1"))
	// we used environment variables to load the
	// email address and the password from the shell
	// you can also directly assign the email address
	// and the password
	from := "forrestmillerj@gmail.com"
	password, err := sm.GetSecretValue(&secretsmanager.GetSecretValueInput{SecretId: secret})
	if err != nil {
		log.Fatal(err.Error())
	}

	// toList is list of email address that email is to be sent.
	toList := []string{"forrestmillerj@gmail.com", "juliettermiller@gmail.com", "millergrayson0@gmail.com"}

	// host is address of server that the
	// sender's email address belongs,
	// in this case its gmail.
	// For e.g if your are using yahoo
	// mail change the address as smtp.mail.yahoo.com
	host := "smtp.gmail.com"

	// Its the default port of smtp server
	port := "587"

	subject := "MainCoon Kittens Available"
	body := `<p>Kittens are Available</p>

  <a href="http://www.congocoon.com/kittensforsale.html">click here</a>`

	request := Mail{
		Sender:  from,
		To:      toList,
		Subject: subject,
		Body:    body,
	}
	// subject
	// This is the message to send in the mail
	// We can't send strings directly in mail,
	// strings need to be converted into slice bytes
	msg := BuildMessage(request)

	// PlainAuth uses the given username and password to
	// authenticate to host and act as identity.
	// Usually identity should be the empty string,
	// to act as username.
	auth := smtp.PlainAuth("", from, *password.SecretString, host)

	// SendMail uses TLS connection to send the mail
	// The email is sent to all address in the toList,
	// the body should be of type bytes, not strings
	// This returns error if any occured.
	err = smtp.SendMail(host+":"+port, auth, from, toList, []byte(msg))

	// handling the errors
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println("Successfully sent mail to all user in toList")
}

func BuildMessage(mail Mail) string {
	msg := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\r\n"
	msg += fmt.Sprintf("From: %s\r\n", mail.Sender)
	msg += fmt.Sprintf("To: %s\r\n", strings.Join(mail.To, ";"))
	msg += fmt.Sprintf("Subject: %s\r\n", mail.Subject)
	msg += fmt.Sprintf("\r\n%s\r\n", mail.Body)

	return msg
}
