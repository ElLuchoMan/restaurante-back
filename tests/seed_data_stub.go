//go:build !integration
// +build !integration

package test

// SeedTestData es un stub para builds sin la etiqueta 'integration'.
// Se incluye un return para aportar una instrucción que pueda ser
// contabilizada por la cobertura.
func SeedTestData() {
	// no-op en modo unit
	return
}
