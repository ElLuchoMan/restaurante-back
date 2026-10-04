package models

import (
	"encoding/json"
	"testing"
)

func TestDashboardDataMarshalJSON(t *testing.T) {
	data := DashboardData{
		TotalPedidos:        100,
		TotalIngresos:       5000000,
		TotalUsuarios:       50,
		PromedioVentaPedido: 50000.0,
		PedidosHoy:          10,
		IngresosHoy:         500000,
	}

	b, err := json.Marshal(data)
	if err != nil {
		t.Fatalf("json.Marshal returned error: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(b, &result); err != nil {
		t.Fatalf("json.Unmarshal returned error: %v", err)
	}

	if result["totalPedidos"].(float64) != 100 {
		t.Errorf("expected totalPedidos 100, got %v", result["totalPedidos"])
	}
	if result["totalIngresos"].(float64) != 5000000 {
		t.Errorf("expected totalIngresos 5000000, got %v", result["totalIngresos"])
	}
}

func TestSalesDataStructure(t *testing.T) {
	salesData := SalesData{
		VentasPorMetodoPago: []VentaPorMetodo{
			{MetodoPago: "EFECTIVO", Total: 100000, Cantidad: 5},
			{MetodoPago: "TARJETA", Total: 200000, Cantidad: 8},
		},
		TendenciaVentas: []VentaPorFecha{
			{Fecha: "2024-01-01", Total: 150000, Cantidad: 3},
		},
		EstadisticasGenerales: EstadisticasVentas{
			VentaPromedioDiaria:  50000.0,
			PedidoPromedioDiario: 10.0,
			TicketPromedio:       5000.0,
		},
	}

	b, err := json.Marshal(salesData)
	if err != nil {
		t.Fatalf("json.Marshal returned error: %v", err)
	}

	var result SalesData
	if err := json.Unmarshal(b, &result); err != nil {
		t.Fatalf("json.Unmarshal returned error: %v", err)
	}

	if len(result.VentasPorMetodoPago) != 2 {
		t.Errorf("expected 2 ventas por método de pago, got %d", len(result.VentasPorMetodoPago))
	}
	if result.VentasPorMetodoPago[0].MetodoPago != "EFECTIVO" {
		t.Errorf("expected EFECTIVO, got %s", result.VentasPorMetodoPago[0].MetodoPago)
	}
}

func TestProductsDataStructure(t *testing.T) {
	productsData := ProductsData{
		ProductosMasVendidos: []ProductoVendido{
			{
				ProductoID:      1,
				NombreProducto:  "Bandeja Paisa",
				CantidadVendida: 50,
				IngresoTotal:    1250000,
				Precio:          25000,
				Imagen:          "base64image",
			},
		},
		EstadisticasProductos: EstadisticasProductos{
			TotalProductosActivos:  25,
			ProductoConMasVentas:   "Bandeja Paisa",
			ProductoConMenosVentas: "Ensalada",
		},
	}

	b, err := json.Marshal(productsData)
	if err != nil {
		t.Fatalf("json.Marshal returned error: %v", err)
	}

	var result ProductsData
	if err := json.Unmarshal(b, &result); err != nil {
		t.Fatalf("json.Unmarshal returned error: %v", err)
	}

	if len(result.ProductosMasVendidos) != 1 {
		t.Errorf("expected 1 producto más vendido, got %d", len(result.ProductosMasVendidos))
	}
	if result.ProductosMasVendidos[0].NombreProducto != "Bandeja Paisa" {
		t.Errorf("expected Bandeja Paisa, got %s", result.ProductosMasVendidos[0].NombreProducto)
	}
}

func TestUsersDataStructure(t *testing.T) {
	usersData := UsersData{
		UsuariosFrecuentes: []UsuarioFrecuente{
			{
				DocumentoCliente: 12345678,
				NombreCompleto:   "Juan Pérez",
				TotalPedidos:     15,
				TotalGastado:     750000,
				UltimoPedido:     "2024-01-15",
			},
		},
		EstadisticasUsuarios: EstadisticasUsuarios{
			TotalClientes:           100,
			ClientesActivos:         75,
			ClientesInactivos:       25,
			PromedioGastoPorCliente: 50000.0,
		},
	}

	b, err := json.Marshal(usersData)
	if err != nil {
		t.Fatalf("json.Marshal returned error: %v", err)
	}

	var result UsersData
	if err := json.Unmarshal(b, &result); err != nil {
		t.Fatalf("json.Unmarshal returned error: %v", err)
	}

	if len(result.UsuariosFrecuentes) != 1 {
		t.Errorf("expected 1 usuario frecuente, got %d", len(result.UsuariosFrecuentes))
	}
	if result.UsuariosFrecuentes[0].NombreCompleto != "Juan Pérez" {
		t.Errorf("expected Juan Pérez, got %s", result.UsuariosFrecuentes[0].NombreCompleto)
	}
}
