package dino

import (
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

// QueryRequest wraps needed information
type QueryRequest struct {
	PrimaryKey      string
	PrimaryOperator string
	PrimaryKeyValue interface{}

	SortKey      string
	SortOperator string
	SortKeyValue interface{}

	Limit int64 // Limit is the limit of items read, NOT the limit of items returned

	FilterExpression string
}

// Query wraps the basic query functionality of Dynamodb
// TODO more adept query expression handling
func (d *Dino) Query(target interface{}, query QueryRequest) {
	queryInput := dynamodb.QueryInput{}
	queryInput.TableName = d.tableName

	q := fmt.Sprintf("%s %s :%s",
		query.PrimaryKey,
		query.PrimaryOperator,
		strings.Split(query.PrimaryKey, " ")[0])

	if query.SortKey != "" {
		q = fmt.Sprintf("%s %s :%s",
			query.SortKey,
			query.SortOperator,
			strings.ToLower(strings.Split(query.SortKey, " ")[0]))
	}

	if query.Limit != 0 {
		queryInput.SetLimit(query.Limit)
	}

	queryInput.KeyConditionExpression = aws.String(q)

	// set attribute values
	av := map[string]*dynamodb.AttributeValue{}

	primaryVal, err := dynamodbattribute.Marshal(query.PrimaryKeyValue)
	if err != nil {
		d.LastAction.Error = err
		return
	}

	av[":"+strings.ToLower(strings.Split(query.PrimaryKey, " ")[0])] = primaryVal

	if query.SortKey != "" {
		sortVal, err := dynamodbattribute.Marshal(query.SortKeyValue)
		if err != nil {
			d.LastAction.Error = err
			return
		}

		av[strings.ToLower(strings.Split(query.SortKey, " ")[0])] = sortVal
	}

	if query.FilterExpression != "" {
		queryInput.SetFilterExpression(query.FilterExpression)
	}

	queryInput.SetExpressionAttributeValues(av)

	log.Println(queryInput.String())
	log.Println(queryInput.Validate())
	// figure out best way to unmarshal output into the supplied target. Might be best to make this a small wrapper
	queryOutput, err := d.session.Query(&queryInput)
	if err != nil {
		d.LastAction.Error = err
		return
	}

	rawReturn := []map[string]interface{}{}

	err = dynamodbattribute.UnmarshalListOfMaps(queryOutput.Items, &rawReturn)
	if err != nil {
		d.LastAction.Error = err
		return
	}

	d.LastAction.Output = rawReturn
}
