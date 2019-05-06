package dino

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDinoSave(t *testing.T) {
	var tableName = "dino"
	d := NewDino(&tableName, "id")

	err := d.Save(map[string]interface{}{"id": "vance", "bob": []string{"bob"}}).Error()

	err = d.Save(map[string]interface{}{"id": "vance", "last": d.LastAction.RawAction}).Error()
	assert.Nil(t, err)
}
