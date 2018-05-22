package engine

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func init() {
	addToRegistry("initRegistry1", "initObj1", "initVal1")
	addToRegistry("initRegistry2", "initObj2", "initVal2")
}

func TestRegistryUseInInit(t *testing.T) {
	assert := assert.New(t)
	assert.Equal("initVal1", findInRegistry("initRegistry1", "initObj1"))
	assert.Equal("initVal2", findInRegistry("initRegistry2", "initObj2"))

	assert.Nil(findInRegistry("initRegistry1", "obj2"))
	assert.Nil(findInRegistry("initRegistry2", "obj1"))
}

func TestRegistry(t *testing.T) {
	addToRegistry("testRegistry1", "obj1", "val1")
	addToRegistry("testRegistry2", "obj2", "val2")

	assert := assert.New(t)
	assert.Equal("val1", findInRegistry("testRegistry1", "obj1"))
	assert.Equal("val2", findInRegistry("testRegistry2", "obj2"))

	assert.Nil(findInRegistry("testRegistry1", "obj2"))
	assert.Nil(findInRegistry("testRegistry2", "obj1"))

	assert.Nil(findInRegistry("testRegistry3", "obj1"))
	assert.Nil(findInRegistry("testRegistry3", "obj2"))
}
