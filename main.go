package main

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"os"

	"github.com/I1Asyl/berliner_backend/pkg/secrets"
	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

func main() {
	fmt.Println("before config")

	err := setupConfigs()
	if err != nil {
		log.Fatalf(err.Error())
	}

	fmt.Println(os.Getenv("dsn"))

	// Create config for Wire
	config := Config{
		DSN: os.Getenv("dsn"),
	}

	// Initialize the app using Wire
	router, err := InitializeApp(config)
	fmt.Println(router)
	if err != nil {
		log.Fatalf("Failed to initialize app: %v", err)
	}

	router.Run()
}

func setupConfigs() error {
	viper.SetConfigName("config")
	viper.AddConfigPath("configs/")
	if err := viper.ReadInConfig(); err != nil {
		return err
	}

	var dbPassword, jwtSecret string

	if viper.GetBool("aws.enabled") {
		// Load secrets from AWS Secrets Manager
		awsRegion := viper.GetString("aws.region")
		dbPasswordSecretName := viper.GetString("aws.secrets.db_password")
		jwtSecretName := viper.GetString("aws.secrets.jwt_secret")

		secretsClient, err := secrets.NewClient(awsRegion)
		if err != nil {
			return fmt.Errorf("failed to create secrets client: %w", err)
		}

		ctx := context.Background()

		dbPassword, err = secretsClient.GetSecret(ctx, dbPasswordSecretName)
		if err != nil {
			return fmt.Errorf("failed to get DB_PASSWORD: %w", err)
		}

		jwtSecret, err = secretsClient.GetSecret(ctx, jwtSecretName)
		if err != nil {
			return fmt.Errorf("failed to get JWT_SECRET: %w", err)
		}
	} else {
		// Load secrets from local .env file
		if err := godotenv.Load("configs/.env"); err != nil {
			return fmt.Errorf("failed to load .env file: %w", err)
		}

		dbPassword = os.Getenv("DB_PASSWORD")
		if dbPassword == "" {
			return fmt.Errorf("DB_PASSWORD is not set in .env file")
		}

		jwtSecret = os.Getenv("JWT_SECRET")
		if jwtSecret == "" {
			return fmt.Errorf("JWT_SECRET is not set in .env file")
		}
	}

	// Set environment variables for use by other packages
	os.Setenv("DB_PASSWORD", dbPassword)
	os.Setenv("JWT_SECRET", jwtSecret)

	username := viper.GetString("db.user")
	address := viper.GetString("db.address")
	dbname := viper.GetString("db.name")
	sslmode := viper.GetString("db.sslmode")

	u := &url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(username, dbPassword),
		Host:   address,
		Path:   dbname,
	}

	q := u.Query()
	q.Set("sslmode", sslmode)
	q.Set("connect_timeout", "10") // Fail after 10 seconds instead of hanging
	u.RawQuery = q.Encode()

	dsn := u.String()

	os.Setenv("dsn", dsn)

	return nil
}
