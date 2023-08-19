package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"manki/pkg/handler"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	_ "github.com/go-sql-driver/mysql"
)

const (
	DbDriverName = "sqlite3"
)

type secrets struct {
	DSN string `json:"DB_DSN"`
}

var mysecret secrets

func main() {
	loadSecret()

	ctx, stop := context.WithCancel(context.Background())
	defer stop()

	fmt.Printf("dsn: %s\n", mysecret.DSN)
	pool, err := sql.Open("mysql", mysecret.DSN)
	if err != nil {
		log.Fatalf("error opening the database: %s", err)
	}
	defer pool.Close()

	server := http.Server{
		Addr:    ":" + "3000",
		Handler: handler.New(ctx, pool),
	}

	go func() {
		log.Println("Running server...")
		if err := server.ListenAndServe(); err != nil {
			log.Fatalf("error running server: %s", err)
		}
	}()

	<-ctx.Done()
}

func loadSecret() {
	secretName := "prod/db_dsn"
	region := "us-east-1"

	config, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
	if err != nil {
		log.Fatal(err)
	}

	// Create Secrets Manager client
	svc := secretsmanager.NewFromConfig(config)

	input := &secretsmanager.GetSecretValueInput{
		SecretId:     aws.String(secretName),
		VersionStage: aws.String("AWSCURRENT"), // VersionStage defaults to AWSCURRENT if unspecified
	}

	result, err := svc.GetSecretValue(context.TODO(), input)
	if err != nil {
		// For a list of exceptions thrown, see
		// https://docs.aws.amazon.com/secretsmanager/latest/apireference/API_GetSecretValue.html
		log.Fatal(err.Error())
	}

	// Decrypts secret using the associated KMS key.
	json.Unmarshal([]byte(*result.SecretString), &mysecret)
}
