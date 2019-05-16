package dino

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestMapSave(t *testing.T) {
	table := "dino"
	d := NewDino(nil, &table, "id")

	test := map[string]interface{}{}

	test["id"] = uuid.New().String()
	test["name"] = "john"
	test["age"] = 2
	test["time"] = time.Now()

	inner := structAllTypes()
	inner.Flotsom = structAllTypes()

	test["Person"] = inner

	d.Save(test)
}

func TestStructSave(t *testing.T) {
	table := "dino"
	d := NewDino(nil, &table, "id")

	test := structAllTypes()

	testInner := structAllTypes()
	testInner.Flotsom = structAllTypes()

	test.Flotsom = testInner

	d.Save(test)
}
