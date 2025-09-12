//go:build !integration
// +build !integration

package test

import "testing"

func TestSeedDataStub(t *testing.T) {
	SeedTestData()
}
