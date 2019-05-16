package dino

import (
	"time"

	"github.com/google/uuid"
)

type TestOuterPerson struct {
	ID           string `dino:"primarykey"`
	Name         string
	Age          int
	Date         time.Time
	IgnoredField string `dino:"-"`
	Flotsom      interface{}
}

type TestOuterPersonMap map[string]interface{}

func (p TestOuterPerson) AfterSave() error {
	return nil
}

func (p TestOuterPersonMap) AfterSave() error {
	return nil
}

func structAllTypes() TestOuterPerson {
	return TestOuterPerson{uuid.New().String(), "John", 12, time.Now(), "Ignored", ""}
}
