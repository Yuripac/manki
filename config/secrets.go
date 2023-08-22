package config

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
)

const region = "us-east-1"

type SecretClient struct {
	client *secretsmanager.Client
}

type Secret map[string]string

func NewSecretClient() (*SecretClient, error) {
	config, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
	if err != nil {
		return nil, err
	}
	client := secretsmanager.NewFromConfig(config)

	return &SecretClient{client: client}, nil
}

func (s SecretClient) GetSecret(name string) (Secret, error) {
	input := &secretsmanager.GetSecretValueInput{
		SecretId:     aws.String(name),
		VersionStage: aws.String("AWSCURRENT"), // VersionStage defaults to AWSCURRENT if unspecified
	}

	result, err := s.client.GetSecretValue(context.TODO(), input)
	if err != nil {
		fmt.Printf("error get secret: %s", err)
		return nil, err
	}

	var awsecret Secret
	if err = json.Unmarshal([]byte(*result.SecretString), &awsecret); err != nil {
		return nil, err
	}

	return awsecret, nil
}
