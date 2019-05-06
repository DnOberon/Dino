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

type Dino struct {
	session     *dynamodb.DynamoDB
	primaryKeys []string
	tableName   *string
	LastAction  Action
}

type Action struct {
	RawAction  interface{}
	Input      interface{}
	ExecutedAt time.Time
	Error      error
}

func NewDino(table *string, primaryKeys ...string) *Dino {
	sess := session.Must(session.NewSession())

	return &Dino{session: dynamodb.New(sess), primaryKeys: primaryKeys, tableName: table}
}

func (d *Dino) Save(input interface{}) Dino {
	// first find out what type we're dealing with - supports map[string]interface, structs, []byte
	if input == nil {
		d.LastAction.Error = errors.New("provided input nil")
		return *d
	}

	d.LastAction.Input = input
	d.LastAction.ExecutedAt = time.Now().UTC()

	inputValue := reflect.ValueOf(input)

	switch inputValue.Kind() {
	case reflect.Map:
		d.saveMap()

	case reflect.Struct:

	default:
		d.LastAction.Error = fmt.Errorf("unacceptable input type : %s", inputValue.Kind())
	}

	return *d
}

func (d Dino) Error() error {
	return d.LastAction.Error
}

func (d *Dino) saveMap() {
	d.LastAction.RawAction = d.saveMap
	inputMap := d.LastAction.Input.(map[string]interface{})

	// check for specified primaryKeys
	for _, key := range d.primaryKeys {
		_, ok := inputMap[key]

		if !ok {
			d.LastAction.Error = fmt.Errorf("required table key missing: %s", key)
			return
		}
	}

	// TODO: Here we'll add the ability to flatten out slices and slices of struct
	// that's a hard stop for it

	// marshal and send
	result, err := dynamodbattribute.MarshalMap(inputMap)
	if err != nil {
		log.Printf("%s", err.Error())
		return
	}

	request := dynamodb.PutItemInput{
		Item:      result,
		TableName: d.tableName,
	}

	_, err = d.session.PutItem(&request)

	if err != nil {
		d.LastAction.Error = err
	}

	return
}

func (d *Dino) saveStruct() {
	d.LastAction.RawAction = d.saveStruct
}
