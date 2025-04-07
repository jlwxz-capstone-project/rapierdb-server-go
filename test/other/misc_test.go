package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	pe "github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestError(t *testing.T) {
	err := pe.WithStack(errors.New("test error"))
	fmt.Printf("%+v\n", err)
}

type DebugPrintable interface {
	DebugPrint() string
}

type StructA struct {
	field1 int
	field2 int
	DebugPrintable
}

func (s *StructA) DebugPrint() string {
	return fmt.Sprintf("StructA{field1: %d, field2: %d}", s.field1, s.field2)
}

var _ DebugPrintable = &StructA{}

func TestNil(t *testing.T) {
	var v1 *int
	assert.True(t, v1 == nil)

	var v2 any
	assert.True(t, v2 == nil)

	v2 = v1
	assert.True(t, v2 == nil)
}

func TestJson(t *testing.T) {
	val := map[string]any{
		"field1": "value1",
		"field2": 2,
	}

	json, err := json.Marshal(val)
	assert.NoError(t, err)
	fmt.Println(string(json))
}

type Animal struct {
	data animalData
}

type animalData interface {
	AsObj() *Animal
}

type BaseAnimal struct {
	Animal
}

func (n *BaseAnimal) AsObj() *Animal {
	return &n.Animal
}

type Cat struct {
	BaseAnimal
}

func NewCat() *Cat {
	cat := &Cat{
		BaseAnimal: BaseAnimal{
			Animal: Animal{},
		},
	}
	cat.data = cat
	return cat
}

func (o *Animal) IsCat() bool { _, ok := o.data.(*Cat); return ok }
func (o *Animal) AsCat() *Cat { return o.data.(*Cat) }

func (cat *Cat) MakeSound() string {
	return "meow"
}

type Dog struct {
	BaseAnimal
}

func NewDog() *Dog {
	dog := &Dog{
		BaseAnimal: BaseAnimal{
			Animal: Animal{},
		},
	}
	dog.data = dog
	return dog
}

func (o *Animal) IsDog() bool { _, ok := o.data.(*Dog); return ok }
func (o *Animal) AsDog() *Dog { return o.data.(*Dog) }

func (dog *Dog) MakeSound() string {
	return "woof"
}

func TestNode(t *testing.T) {
	cat := NewCat()
	dog := NewDog()

	assert.True(t, cat.AsObj().IsCat())
	assert.True(t, dog.AsObj().IsDog())

	assert.Equal(t, cat, cat.AsObj().AsCat())
	assert.Equal(t, dog, dog.AsObj().AsDog())

	assert.Equal(t, "meow", cat.MakeSound())
	assert.Equal(t, "woof", dog.MakeSound())
}
