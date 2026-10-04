package models

import "testing"

func TestRolTrabajador_IsValid(t *testing.T) {
	tests := []struct {
		name string
		rol  RolTrabajador
		want bool
	}{
		{"Administrador válido", RolAdministrador, true},
		{"Mesero válido", RolMesero, true},
		{"Cocinero válido", RolCocinero, true},
		{"Domiciliario válido", RolDomiciliario, true},
		{"Oficios varios válido", RolOficiosVarios, true},
		{"Rol inválido", RolTrabajador("INVALIDO"), false},
		{"Rol vacío", RolTrabajador(""), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.rol.IsValid(); got != tt.want {
				t.Errorf("RolTrabajador.IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEstadoControlNomina_IsValid(t *testing.T) {
	tests := []struct {
		name   string
		estado EstadoControlNomina
		want   bool
	}{
		{"No generada válido", EstadoControlNominaNoGenerada, true},
		{"Generada válido", EstadoControlNominaGenerada, true},
		{"Regenerada válido", EstadoControlNominaReGenerada, true},
		{"Estado inválido", EstadoControlNomina("INVALIDO"), false},
		{"Estado vacío", EstadoControlNomina(""), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.estado.IsValid(); got != tt.want {
				t.Errorf("EstadoControlNomina.IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPlataformaNotificacion_IsValid(t *testing.T) {
	tests := []struct {
		name       string
		plataforma PlataformaNotificacion
		want       bool
	}{
		{"Web válido", PlataformaWeb, true},
		{"Android válido", PlataformaAndroid, true},
		{"iOS válido", PlataformaIOS, true},
		{"Plataforma inválida", PlataformaNotificacion("INVALIDO"), false},
		{"Plataforma vacía", PlataformaNotificacion(""), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.plataforma.IsValid(); got != tt.want {
				t.Errorf("PlataformaNotificacion.IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProveedorPush_IsValid(t *testing.T) {
	tests := []struct {
		name      string
		proveedor ProveedorPush
		want      bool
	}{
		{"WebPush válido", ProveedorWebPush, true},
		{"FCM válido", ProveedorFCM, true},
		{"Proveedor inválido", ProveedorPush("INVALIDO"), false},
		{"Proveedor vacío", ProveedorPush(""), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.proveedor.IsValid(); got != tt.want {
				t.Errorf("ProveedorPush.IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTipoDescuento_IsValid(t *testing.T) {
	tests := []struct {
		name string
		tipo TipoDescuento
		want bool
	}{
		{"Porcentaje válido", TipoDescuentoPorcentaje, true},
		{"Monto válido", TipoDescuentoMonto, true},
		{"Tipo inválido", TipoDescuento("INVALIDO"), false},
		{"Tipo vacío", TipoDescuento(""), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.tipo.IsValid(); got != tt.want {
				t.Errorf("TipoDescuento.IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCuponScope_IsValid(t *testing.T) {
	tests := []struct {
		name  string
		scope CuponScope
		want  bool
	}{
		{"Global válido", CuponScopeGlobal, true},
		{"Producto válido", CuponScopeProducto, true},
		{"Categoría válido", CuponScopeCategoria, true},
		{"Cliente válido", CuponScopeCliente, true},
		{"Scope inválido", CuponScope("INVALIDO"), false},
		{"Scope vacío", CuponScope(""), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.scope.IsValid(); got != tt.want {
				t.Errorf("CuponScope.IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEnumConstants(t *testing.T) {

	if EstadoDomicilioPendiente != "PENDIENTE" {
		t.Errorf("EstadoDomicilioPendiente = %v, want PENDIENTE", EstadoDomicilioPendiente)
	}
	if EstadoDomicilioEnCamino != "EN_CAMINO" {
		t.Errorf("EstadoDomicilioEnCamino = %v, want EN_CAMINO", EstadoDomicilioEnCamino)
	}
	if EstadoDomicilioEntregado != "ENTREGADO" {
		t.Errorf("EstadoDomicilioEntregado = %v, want ENTREGADO", EstadoDomicilioEntregado)
	}

	if EstadoNominaPago != "PAGO" {
		t.Errorf("EstadoNominaPago = %v, want PAGO", EstadoNominaPago)
	}
	if EstadoNominaNoPago != "NO_PAGO" {
		t.Errorf("EstadoNominaNoPago = %v, want NO_PAGO", EstadoNominaNoPago)
	}

	if EstadoPagoPagado != "PAGADO" {
		t.Errorf("EstadoPagoPagado = %v, want PAGADO", EstadoPagoPagado)
	}
	if EstadoPagoPendiente != "PENDIENTE" {
		t.Errorf("EstadoPagoPendiente = %v, want PENDIENTE", EstadoPagoPendiente)
	}
	if EstadoPagoNoPago != "NO_PAGO" {
		t.Errorf("EstadoPagoNoPago = %v, want NO_PAGO", EstadoPagoNoPago)
	}

	if EstadoPedidoIniciado != "INICIADO" {
		t.Errorf("EstadoPedidoIniciado = %v, want INICIADO", EstadoPedidoIniciado)
	}
	if EstadoPedidoEnPreparacion != "EN_PREPARACION" {
		t.Errorf("EstadoPedidoEnPreparacion = %v, want EN_PREPARACION", EstadoPedidoEnPreparacion)
	}
	if EstadoPedidoListo != "LISTO" {
		t.Errorf("EstadoPedidoListo = %v, want LISTO", EstadoPedidoListo)
	}
	if EstadoPedidoTerminado != "TERMINADO" {
		t.Errorf("EstadoPedidoTerminado = %v, want TERMINADO", EstadoPedidoTerminado)
	}
	if EstadoPedidoCancelado != "CANCELADO" {
		t.Errorf("EstadoPedidoCancelado = %v, want CANCELADO", EstadoPedidoCancelado)
	}

	if EstadoProductoDisponible != "DISPONIBLE" {
		t.Errorf("EstadoProductoDisponible = %v, want DISPONIBLE", EstadoProductoDisponible)
	}
	if EstadoProductoNoDisponible != "NO_DISPONIBLE" {
		t.Errorf("EstadoProductoNoDisponible = %v, want NO_DISPONIBLE", EstadoProductoNoDisponible)
	}

	if EstadoReservaPendiente != "PENDIENTE" {
		t.Errorf("EstadoReservaPendiente = %v, want PENDIENTE", EstadoReservaPendiente)
	}
	if EstadoReservaConfirmada != "CONFIRMADA" {
		t.Errorf("EstadoReservaConfirmada = %v, want CONFIRMADA", EstadoReservaConfirmada)
	}
	if EstadoReservaCancelada != "CANCELADA" {
		t.Errorf("EstadoReservaCancelada = %v, want CANCELADA", EstadoReservaCancelada)
	}
	if EstadoReservaCumplida != "CUMPLIDA" {
		t.Errorf("EstadoReservaCumplida = %v, want CUMPLIDA", EstadoReservaCumplida)
	}

	if DiaLunes != "Lunes" {
		t.Errorf("DiaLunes = %v, want Lunes", DiaLunes)
	}
	if DiaMartes != "Martes" {
		t.Errorf("DiaMartes = %v, want Martes", DiaMartes)
	}
	if DiaMiercoles != "Miércoles" {
		t.Errorf("DiaMiercoles = %v, want Miércoles", DiaMiercoles)
	}
	if DiaJueves != "Jueves" {
		t.Errorf("DiaJueves = %v, want Jueves", DiaJueves)
	}
	if DiaViernes != "Viernes" {
		t.Errorf("DiaViernes = %v, want Viernes", DiaViernes)
	}
	if DiaSabado != "Sábado" {
		t.Errorf("DiaSabado = %v, want Sábado", DiaSabado)
	}
	if DiaDomingo != "Domingo" {
		t.Errorf("DiaDomingo = %v, want Domingo", DiaDomingo)
	}

	if RolAdministrador != "Administrador" {
		t.Errorf("RolAdministrador = %v, want Administrador", RolAdministrador)
	}
	if RolMesero != "Mesero" {
		t.Errorf("RolMesero = %v, want Mesero", RolMesero)
	}
	if RolCocinero != "Cocinero" {
		t.Errorf("RolCocinero = %v, want Cocinero", RolCocinero)
	}
	if RolDomiciliario != "Domiciliario" {
		t.Errorf("RolDomiciliario = %v, want Domiciliario", RolDomiciliario)
	}
	if RolOficiosVarios != "Oficios_varios" {
		t.Errorf("RolOficiosVarios = %v, want Oficios_varios", RolOficiosVarios)
	}
}
