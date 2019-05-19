package dino

import (
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStructQuery(t *testing.T) {
	table := "dino"
	d := NewDino(nil, &table, "id")

	test := structAllTypes()

	testInner := structAllTypes()
	testInner.Flotsom = structAllTypes()

	test.Flotsom = testInner

	d.Save(test)
	assert.Nil(t, d.LastAction.Error)

	query := QueryRequest{}
	query.PrimaryKey = "id"
	query.PrimaryKeyValue = test.ID
	query.PrimaryOperator = "="

	d.Query(nil, query)

	assert.Nil(t, d.LastAction.Error)

	log.Println(d.LastAction.Output)
}
