package dino

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TODO Clean the test cases up, build generators for these kind of nested
// cases. Or write them in json (much easier)
func TestFlattenMap(t *testing.T) {
	control := map[string]interface{}{}
	control["test"] = map[string]interface{}{"inner": "inner2"}

	test := map[string]interface{}{}

	test["test"] = map[string]interface{}{"inner": "inner2"}
	test["test#inner"] = "inner2"

	flattenMap(control)
	assert.Equal(t, test, control, "simple string")

	control = map[string]interface{}{}
	control["test"] = map[string]interface{}{"inner": 2,
		"inner'er": []string{"test"}}

	test = map[string]interface{}{}
	test["test"] = map[string]interface{}{"inner": 2,
		"inner'er": []string{"test"}}

	test["test#inner"] = 2
	test["test#inner'er"] = []string{"test"}

	flattenMap(control)
	assert.Equal(t, test, control, "integer and slice")

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

	flattenMap(control)
	assert.Equal(t, test, control, "struct")
}

func TestFlattenMapNested(t *testing.T) {
	control := map[string]interface{}{}
	control["upper"] = map[string]interface{}{"middle": map[string]interface{}{"lower": map[string]interface{}{"lowest": "lowest value"}}}

	test := map[string]interface{}{}
	test["upper"] = map[string]interface{}{"middle": map[string]interface{}{"lower": map[string]interface{}{"lowest": "lowest value"}}}
	test["upper#middle"] = map[string]interface{}{"lower": map[string]interface{}{"lowest": "lowest value"}}

	test["upper#middle#lower"] = map[string]interface{}{"lowest": "lowest value"}
	test["upper#middle#lower#lowest"] = "lowest value"

	flattenMap(control)
	assert.Equal(t, test, control, "nested")
}
