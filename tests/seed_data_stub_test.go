//go:build !integration
// +build !integration

package test

import "testing"

func TestSeedDataStub(t *testing.T) {
	// Calling SeedTestData ensures the stub is covered
	SeedTestData()
}
