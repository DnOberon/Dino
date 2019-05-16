package dino

import (
	"fmt"
	"reflect"
	"strings"
	"unicode"
)

// TODO Need to figure out how best to have flatten map and flatten struct work together
// TODO Benchmark the hell out of this. It cannot be slow - if we can change values in place
// on the parent map let's do so
func flattenMap(i interface{}) interface{} {
	inValue := reflect.ValueOf(i)

	if inValue.Kind() != reflect.Map {
		return i
	}

	input := map[string]interface{}{}
	safeLoop := inValue.MapRange()

	for safeLoop.Next() {
		key := safeLoop.Key()
		value := safeLoop.Value()

		input[key.String()] = value.Interface()
	}

	for key, value := range input {
		kind := reflect.ValueOf(value).Kind()

		switch kind {
		case reflect.Map:
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

		case reflect.Struct:
			flatStruct := flattenStruct(value, 1)

			for key, value := range flatStruct {
				input[key] = value
			}

		default:
			continue
		}

	}

	return input
}

// TODO the goal here is to get a struct paired down to a map[string]interface
// though tempted, don't do anything super crazy right now. Basic struct tags for naming
// conventions, primary keys, omit,
func flattenStruct(in interface{}, level int) map[string]interface{} {
	output := map[string]interface{}{}
	inType := reflect.TypeOf(in)

	if reflect.ValueOf(in).Kind() != reflect.Struct {
		return output
	}

	totalFields := inType.NumField()
	structName := inType.Name()
	fieldList := []reflect.StructField{}

	for i := 0; i < totalFields; i++ {
		fieldList = append(fieldList, inType.Field(i))
	}

	for i, field := range fieldList {
		tags, ok := field.Tag.Lookup("dino")

		if ok {
			if tags == "-" {
				continue
			}
		}

		// ignore private fields
		if !unicode.IsUpper(rune(field.Name[0])) {
			continue
		}

		formattedKey := fmt.Sprintf("%s#%s", structName, strings.ToLower(field.Name))

		// if considered root level, don't append the struct name
		if level == 0 {
			formattedKey = strings.ToLower(field.Name)
		}

		// TODO uncaught edgecases for handling all int types
		if reflect.ValueOf(in).Field(i).Kind() == reflect.Int {
			output[formattedKey] = reflect.ValueOf(in).Field(i).Int()
			continue
		}

		output[formattedKey] = reflect.ValueOf(in).Field(i).Interface()
	}

	return output
}
