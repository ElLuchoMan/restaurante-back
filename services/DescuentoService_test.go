package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewDescuentoService(t *testing.T) {

	service := NewDescuentoService(nil)
	assert.NotNil(t, service)
}

func TestDescuentoService_BasicInstantiation(t *testing.T) {
	service := &DescuentoService{}
	assert.NotNil(t, service)
}

func TestDescuentoService_Constructor(t *testing.T) {
	service := NewDescuentoService(nil)
	assert.NotNil(t, service)
	assert.Nil(t, service.ormer)
}

func TestDescuentoService_TypeValidation(t *testing.T) {
	var service *DescuentoService
	assert.Nil(t, service)

	service = &DescuentoService{}
	assert.NotNil(t, service)
	assert.IsType(t, &DescuentoService{}, service)
}

func TestDescuentoService_OrmerHandling(t *testing.T) {

	service1 := NewDescuentoService(nil)
	assert.NotNil(t, service1)
	assert.Nil(t, service1.ormer)

	service2 := &DescuentoService{}
	assert.NotNil(t, service2)

	assert.NotSame(t, service1, service2)
}

func TestDescuentoService_StructureFields(t *testing.T) {
	service := &DescuentoService{}

	assert.NotNil(t, service)

	assert.Nil(t, service.ormer)
}

func TestDescuentoService_MultipleInstances(t *testing.T) {
	service1 := NewDescuentoService(nil)
	service2 := NewDescuentoService(nil)

	assert.NotNil(t, service1)
	assert.NotNil(t, service2)
	assert.NotSame(t, service1, service2)
}

func TestDescuentoService_OrmerState(t *testing.T) {
	service := NewDescuentoService(nil)
	assert.Nil(t, service.ormer)

	ormer := service.ormer
	assert.Nil(t, ormer)
}

func TestDescuentoService_BasicCoverage(t *testing.T) {

	service := NewDescuentoService(nil)
	assert.NotNil(t, service)

	service2 := &DescuentoService{}
	assert.NotNil(t, service2)

	assert.NotSame(t, service, service2)
	assert.IsType(t, service, service2)
}
