// Package dino provides an easy to use adapter to AWS's dynamodb. It aims to be flexible in both
// manipulation and accessing of dynamodb data as well as promoting efficient data access patterns.
package dino

import (
	"errors"
	"fmt"
	"log"
	"reflect"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

// TODO set saner defaults for a package of this type & addition of a flattening delimiter
var config aws.Config

// Dino is the primary workhorse of the package. Contains the reference to the started DynamoDB session.
// Each instance of Dino represents a single DynamoDB table. Having a single table should be the goal of
// any application utilizing DynamoDB and is considered best practice. Use Local and Global Secondary Indexes
// as well as flattening techniques to get the most out of the system.
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
func (d *Dino) Error() error {
	return d.LastAction.Error
}

func (d *Dino) snapshot() instanceState {
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

	// TODO we can use the reflect.Value.MapRange function over strongly typing
	inputMap := input.(map[string]interface{})

	// check for specified primaryKeys
	for _, key := range d.primaryKeys {
		_, ok := inputMap[key]

		if !ok {
			d.LastAction.Error = fmt.Errorf("required table key missing: %s", key)
			return
		}
	}

	// TODO flattening option to config
	flattenMap(inputMap)

	if inputMap == nil {
		d.LastAction.Error = errors.New("error flattening map")
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

// TODO Handle a more broad range of nested map types.
// TODO Need to figure out how best to have flatten map and flatten struct work together
// TODO Benchmark the hell out of this. It cannot be slow
func flattenMap(i interface{}) {
	if reflect.ValueOf(i).Kind() != reflect.Map {
		return
	}

	// TODO handle a more broad set of map types. Might have to build a new map?
	input := i.(map[string]interface{})

	for key, value := range input {
		kind := reflect.ValueOf(value).Kind()
		if kind != reflect.Map {
			continue
		}

		iter := reflect.ValueOf(value).MapRange()

		for iter.Next() {
			k := iter.Key()

			// OH DAMN Stumbled upon what I wanted accidentally. I need to catch any nested
			// maps - because the key I'm adding to the map hasn't yet been iterated over
			// nesting is handled automatically for me as every time I add a new key  it is
			// guaranteed it will be checked
			newKey := fmt.Sprintf("%s#%v", key, k.String())
			input[newKey] = iter.Value().Interface()
		}
	}

	return
}
