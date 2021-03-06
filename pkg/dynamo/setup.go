package dynamo

import (
	"context"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/awslabs/goformation"
)

func localDynamoDB(t *testing.T) *dynamodb.DynamoDB {
	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String("local"),
		Credentials: credentials.NewStaticCredentials("local", "local", "local"),
	})
	if err != nil {
		t.Fatal("could not setup db connection")
	}
	db := dynamodb.New(sess, &aws.Config{Endpoint: aws.String("http://localhost:8000")})

	ctx, cancel := context.WithTimeout(context.Background(), 500 * time.Millisecond)
	defer cancel()
	_, err = db.ListTablesWithContext(ctx, nil)
	if err != nil {
		t.Fatal("make sure DynamoDB local runs on port :8000")
	}
	return db
}

// SetupTable creates table defined in the CloudFormation template file under `path`.
// It returns connection to the DynamoDB and cleanup function, that needs to be run after tests.
func SetupTable(t *testing.T, ctx context.Context, tableName, path string) (*dynamodb.DynamoDB, func()) {
	db := localDynamoDB(t)
	tmpl, err := goformation.Open(path)
	if err != nil {
		t.Fatal(err)
	}

	table, err := tmpl.GetAWSDynamoDBTableWithName(tableName)
	if err != nil {
		t.Fatal(err)
	}
	input := FromCloudFormationToCreateInput(*table)
	_, err = db.CreateTableWithContext(ctx, &input)
	if err != nil {
		t.Fatal(err)
	}
	return db, func() {
		db.DeleteTableWithContext(ctx, &dynamodb.DeleteTableInput{TableName: aws.String(tableName)})
	}
}
