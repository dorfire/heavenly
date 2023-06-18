package goparse_test

import (
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/dorfire/heavenly/pkg/goparse"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPackageImports(t *testing.T) {
	regImports, err := goparse.PackageImports("testdata/pkga", false)
	assert.NoError(t, err)
	assertSetsEqual(t, mapset.NewThreadUnsafeSet("log", "github.com/samber/lo"), regImports)

	testImports, err := goparse.PackageImports("testdata/pkga", true)
	assert.NoError(t, err)
	assertSetsEqual(t, mapset.NewThreadUnsafeSet("testing", "github.com/stretchr/testify/assert", "example.com/pkga"), testImports)

	mainImports, err := goparse.PackageImports("testdata/pkga/cmd", false)
	assert.NoError(t, err)
	assertSetsEqual(t, mapset.NewThreadUnsafeSet("log", "net/http", "example.com/pkga"), mainImports)
}

func assertSetsEqual(t *testing.T, want, got mapset.Set[string]) {
	t.Helper()
	if !assert.True(t, got.Equal(want)) {
		t.Logf("SymmetricDifference: %v", want.SymmetricDifference(got).ToSlice())
		t.Logf("Difference: %v", got.Difference(want).ToSlice())
		t.Logf("Difference: %v", want.Difference(got).ToSlice())
	}
}
