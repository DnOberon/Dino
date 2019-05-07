// Package dino provides an easy to use adapter to AWS's dynamodb. It aims to be flexible in both
// manipulation and accessing of dynamodb data as well as promoting efficient data access patterns.
package dino

import (
	"errors"
	"fmt"
	"log"
	"reflect"
	"time"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

// TODO set saner defaults for a package of this type
var config aws.Config

// Dino is the primary workhorse of the package. Contains the reference to the started DynamoDB session.
type Dino struct {
	session     *dynamodb.DynamoDB
	primaryKeys []string
	tableName   *string
	LastAction  Action
}

// Action contains information on the last executed action. Will also allow for replay/reversal of certain actions.
type Action struct {
	RawAction  interface{}
	Input      interface{}
	ExecutedAt time.Time
	Snapshot   instanceState
	Error      error
}

type instanceState struct {
	session     dynamodb.DynamoDB
	primaryKeys []string
	tableName   string
	capturedAt  time.Time
}

// NewDino creates a new Dino instance with the provided table name and primary keys. This is the most basic
// initialization of the service - more complicated configuration, including underlying access to the AWS Config
// struct can be accomplished through the *Dino.Config() option after initialization.
func NewDino(sess *session.Session, table *string, primaryKeys ...string) *Dino {
	if sess == nil {
		sess = session.Must(session.NewSession())
	}

	return &Dino{session: dynamodb.New(sess), primaryKeys: primaryKeys, tableName: table}
}

// Save persists data to Dino.tableName. Accepts either a struct (with proper tags) or a map[string]interface{}
func (d *Dino) Save(input interface{}) Dino {
	// first find out what type we're dealing with - supports map[string]interface, structs, []byte
	if input == nil {
		d.LastAction.Error = errors.New("provided input nil")
		return *d
	}

	switch reflect.ValueOf(input).Kind() {
	case reflect.Map:
		d.saveMap(input)

	default:
		d.LastAction.Error = fmt.Errorf("unacceptable input type : %s", reflect.ValueOf(input).Kind())
	}

	return *d
}

// Error returns any errors from the last performed action
func (d Dino) Error() error {
	return d.LastAction.Error
}

func (d Dino) snapshot() instanceState {
	return instanceState{
		session:     *d.session,
		primaryKeys: d.primaryKeys,
		tableName:   *d.tableName,
		capturedAt:  time.Now().UTC(),
	}
}

func (d *Dino) saveMap(input interface{}) {
	d.LastAction.Input = input
	d.LastAction.RawAction = d.saveMap
	d.LastAction.ExecutedAt = time.Now().UTC()
	d.LastAction.Snapshot = d.snapshot()

	inputMap := input.(map[string]interface{})

	// check for specified primaryKeys
	for _, key := range d.primaryKeys {
		_, ok := inputMap[key]

		if !ok {
			d.LastAction.Error = fmt.Errorf("required table key missing: %s", key)
			return
		}
	}

	// marshal and send
	result, err := dynamodbattribute.MarshalMap(inputMap)
	if err != nil {
		// TODO issue #12 accept logger interface, default to stdout
		log.Printf("%s", err.Error())
		return
	}

	request := dynamodb.PutItemInput{
		// ConditionExpression: TODO issue #13 should we accept condtion expressions
		Item:      result,
		TableName: d.tableName,
	}

	// TODO use item output in issue #11 - self consumed capacity management
	_, err = d.session.PutItem(&request)
	if err != nil {
		d.LastAction.Error = err
	}

	return
}
