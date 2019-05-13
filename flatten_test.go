package dino

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TODO Clean the test cases up, build generators for these kind of nested
// cases. Or write them in json (much easier)
// TODO YOU CANNOT RELIABELY DEEP EQUAL TEST MAPS! Redo the tests to reflecct that
func TestFlattenMap(t *testing.T) {
	control := map[string]interface{}{}
	control["test"] = map[string]interface{}{"inner": "inner2"}

	test := map[string]interface{}{}

	test["test"] = map[string]interface{}{"inner": "inner2"}
	test["test#inner"] = "inner2"

	check := flattenMap(control)
	assert.Equal(t, test, check, "simple string")

	control = map[string]interface{}{}
	control["test"] = map[string]interface{}{"inner": 2,
		"inner'er": []string{"test"}}

	test = map[string]interface{}{}
	test["test"] = map[string]interface{}{"inner": 2,
		"inner'er": []string{"test"}}

	test["test#inner"] = 2
	test["test#inner'er"] = []string{"test"}

	check = flattenMap(control)
	assert.Equal(t, test, check, "integer and slice")

	// Struct management (non nested)
	type sampleStruct struct {
		Name string
	}

	control = map[string]interface{}{}
	control["test"] = map[string]interface{}{"inner": 2,
		"inner'er":      []string{"test"},
		"sample struct": sampleStruct{Name: "John"}}

	test = map[string]interface{}{}
	test["test"] = map[string]interface{}{"inner": 2,
		"inner'er":      []string{"test"},
		"sample struct": sampleStruct{Name: "John"}}

	test["test#inner"] = 2
	test["test#inner'er"] = []string{"test"}
	test["test#sample struct"] = sampleStruct{Name: "John"}

	check = flattenMap(control)
	assert.Equal(t, test, check, "struct")
}

func TestFlattenMapNested(t *testing.T) {
	control := map[string]interface{}{}
	control["upper"] = map[string]interface{}{"middle": map[string]interface{}{"lower": map[string]interface{}{"lowest": "lowest value"}}}

	test := map[string]interface{}{}
	test["upper"] = map[string]interface{}{"middle": map[string]interface{}{"lower": map[string]interface{}{"lowest": "lowest value"}}}
	test["upper#middle"] = map[string]interface{}{"lower": map[string]interface{}{"lowest": "lowest value"}}

	test["upper#middle#lower"] = map[string]interface{}{"lowest": "lowest value"}
	test["upper#middle#lower#lowest"] = "lowest value"

	check := flattenMap(control)
	assert.Equal(t, test, check, "nested")
}
