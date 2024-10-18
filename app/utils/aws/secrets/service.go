package secrets

import (
	"encoding/json"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	log "github.com/sirupsen/logrus"
)

type Service interface {
	GetAndUnmarshal(secretName string, v interface{}) error
	GetAndUnmarshalStrict(secretName string, v interface{}) error
	MustGetAndUnmarshal(secretName string, v interface{})
}

type service struct {
	manager *secretsmanager.SecretsManager
}

func NewService() Service {
	return &service{manager: manager()}
}

// GetAndUnmarshal retrieves the unmarshalled value behind the secret.
func (s *service) GetAndUnmarshal(secretName string, v interface{}) error {
	secretValue, err := s.manager.GetSecretValue(&secretsmanager.GetSecretValueInput{
		SecretId: &secretName,
	})
	if err != nil {
		return err
	}
	if err := json.Unmarshal([]byte(*secretValue.SecretString), v); err != nil {
		return err
	}
	return nil
}

// GetAndUnmarshalStrict retrieves the unmarshalled value behind the secret
// without allowing unknown fields
func (s *service) GetAndUnmarshalStrict(secretName string, v interface{}) error {
	secretValue, err := s.manager.GetSecretValue(&secretsmanager.GetSecretValueInput{
		SecretId: &secretName,
	})
	if err != nil {
		return err
	}
	decoder := json.NewDecoder(strings.NewReader(*secretValue.SecretString))
	decoder.DisallowUnknownFields()
	return decoder.Decode(v)
}

// MustGetAndUnmarshal retrieves the unmarshalled value behind the secret.
func (s *service) MustGetAndUnmarshal(secretName string, v interface{}) {
	secretValue, err := s.manager.GetSecretValue(&secretsmanager.GetSecretValueInput{
		SecretId: &secretName,
	})
	if err != nil {
		log.WithField("secretName", secretName).WithError(err).Error("configuration issue: problem getting secret.")
		panic(err)
	}
	if err := json.Unmarshal([]byte(*secretValue.SecretString), v); err != nil {
		log.WithField("secretName", secretName).WithError(err).Error("configuration issue: problem unmarshalling secret.")
		panic(err)
	}
}

func manager() *secretsmanager.SecretsManager {
	var sess *session.Session
	if _, ok := os.LookupEnv("AWS_REGION"); ok {
		sess = session.Must(session.NewSession())
	} else {
		sess = session.Must(session.NewSession(&aws.Config{Region: aws.String("eu-west-1")}))
	}
	return secretsmanager.New(sess)
}
