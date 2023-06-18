package pkga_test

import (
	"example.com/pkga"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDoSomething(t *testing.T) {
	assert.NoError(t, pkga.DoSomething())
}
