# 📊 Guía Completa para Implementar Frontend de Telemetría

## 🎯 Prompt para Cursor

```
Necesito implementar un servicio completo de telemetría en TypeScript/React que consuma los siguientes endpoints de mi API backend.

Crea:
1. **Servicio de Telemetría** (`telemetryService.ts`) con todos los métodos
2. **Modelos TypeScript** (`telemetryTypes.ts`) con todas las interfaces tipadas
3. **Mocks de datos** (`telemetryMocks.ts`) para desarrollo y testing
4. **Hook personalizado** (`useTelemetry.ts`) para React con manejo de estado
5. **Componentes de ejemplo** para mostrar las métricas

Usa Axios para las peticiones HTTP, incluye manejo de errores, loading states, y cache básico.
```

---

## 📡 Endpoints Disponibles

### Base URL: `http://localhost:8080/restaurante/v1/telemetria`
### Autenticación: `Authorization: Bearer {token}`

---

## 🌟 **NUEVO: Endpoint Público de Productos Populares**

### **Endpoint:** `GET /productos-populares` ⚡ **SIN AUTENTICACIÓN**

```bash
curl -X 'GET' \
  'http://localhost:8080/restaurante/v1/productos-populares?limit=4&periodo=ultimo_mes' \
  -H 'accept: application/json'
```

**Parámetros:**
- `limit` (opcional): Número de productos a retornar (default: 4)
- `periodo` (opcional): `hoy` | `ultima_semana` | `ultimo_mes` | `ultimos_3_meses` | `ultimos_6_meses` | `ultimo_año` | `historico`

**Respuesta esperada:**
```json
{
  "code": 200,
  "message": "Productos populares (ultimo_mes) obtenidos exitosamente",
  "data": {
    "productosPopulares": [
      {
        "productoId": 1,
        "nombreProducto": "Bandeja Paisa",
        "cantidadVendida": 145,
        "ingresoTotal": 5220000,
        "precio": 36000,
        "imagen": "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mP8/5+hHgAHggJ/PchI7wAAAABJRU5ErkJggg=="
      },
      {
        "productoId": 2,
        "nombreProducto": "Sancocho de Gallina",
        "cantidadVendida": 128,
        "ingresoTotal": 4480000,
        "precio": 35000,
        "imagen": "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mP8/5+hHgAHggJ/PchI7wAAAABJRU5ErkJggg=="
      },
      {
        "productoId": 3,
        "nombreProducto": "Ajiaco Santafereño",
        "cantidadVendida": 98,
        "ingresoTotal": 3430000,
        "precio": 35000,
        "imagen": "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mP8/5+hHgAHggJ/PchI7wAAAABJRU5ErkJggg=="
      },
      {
        "productoId": 4,
        "nombreProducto": "Arroz con Pollo",
        "cantidadVendida": 87,
        "ingresoTotal": 2610000,
        "precio": 30000,
        "imagen": "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mP8/5+hHgAHggJ/PchI7wAAAABJRU5ErkJggg=="
      }
    ]
  }
}
```

**🎯 Casos de uso:**
- ✅ **Frontend Home**: Mostrar "Platos Estrella" sin necesidad de autenticación
- ✅ **Landing Page**: Destacar productos más populares
- ✅ **Menú público**: Resaltar platos favoritos de los clientes
- ✅ **Marketing**: Promocionar productos con mejor rendimiento

---

## 1. 📈 Dashboard General

### **Endpoint:** `GET /dashboard`

```bash
curl -X 'GET' \
  'http://localhost:8080/restaurante/v1/telemetria/dashboard?periodo=ultimo_mes' \
  -H 'accept: application/json' \
  -H 'Authorization: Bearer {token}'
```

**Parámetros:**
- `periodo` (opcional): `hoy` | `ultima_semana` | `ultimo_mes` | `ultimos_3_meses` | `ultimos_6_meses` | `ultimo_año` | `historico`

**Respuesta esperada:**
```json
{
  "code": 200,
  "message": "Dashboard obtenido exitosamente",
  "data": {
    "totalPedidos": 1250,
    "ingresosTotales": 45750000.50,
    "usuariosRegistrados": 320,
    "valorPromedioOrden": 36600.40
  }
}
```

---

## 2. 💰 Análisis de Ventas

### **Endpoint:** `GET /sales`

```bash
curl -X 'GET' \
  'http://localhost:8080/restaurante/v1/telemetria/sales?periodo=ultimos_3_meses' \
  -H 'accept: application/json' \
  -H 'Authorization: Bearer {token}'
```

**Parámetros:**
- `periodo` (opcional): Filtros temporales disponibles

**Respuesta esperada:**
```json
{
  "code": 200,
  "message": "Análisis de ventas obtenido exitosamente",
  "data": {
    "pedidosPorMetodoPago": [
      {"metodoPago": "EFECTIVO", "cantidad": 450, "porcentaje": 36.0},
      {"metodoPago": "TARJETA", "cantidad": 520, "porcentaje": 41.6},
      {"metodoPago": "TRANSFERENCIA", "cantidad": 280, "porcentaje": 22.4}
    ],
    "ingresosPorMetodoPago": [
      {"metodoPago": "EFECTIVO", "ingresos": 16470000.00, "porcentaje": 36.0},
      {"metodoPago": "TARJETA", "ingresos": 19032000.00, "porcentaje": 41.6},
      {"metodoPago": "TRANSFERENCIA", "ingresos": 10248000.00, "porcentaje": 22.4}
    ],
    "tendenciasVentas": [
      {"fecha": "2024-01-01", "ventas": 1250000.50, "pedidos": 45},
      {"fecha": "2024-01-02", "ventas": 980000.25, "pedidos": 38},
      {"fecha": "2024-01-03", "ventas": 1450000.75, "pedidos": 52}
    ]
  }
}
```

---

## 3. 🍽️ Análisis de Productos

### **Endpoint:** `GET /products`

```bash
curl -X 'GET' \
  'http://localhost:8080/restaurante/v1/telemetria/products?limit=10&periodo=ultimo_mes' \
  -H 'accept: application/json' \
  -H 'Authorization: Bearer {token}'
```

**Parámetros:**
- `limit` (opcional): Número de productos a retornar (default: 10)
- `periodo` (opcional): Filtros temporales disponibles

**Respuesta esperada:**
```json
{
  "code": 200,
  "message": "Análisis de productos obtenido exitosamente",
  "data": {
    "productosMasVendidos": [
      {"nombreProducto": "Bandeja Paisa", "cantidadVendida": 145, "ingresosTotales": 5220000.00},
      {"nombreProducto": "Sancocho de Gallina", "cantidadVendida": 128, "ingresosTotales": 4480000.00},
      {"nombreProducto": "Ajiaco Santafereño", "cantidadVendida": 98, "ingresosTotales": 3430000.00}
    ],
    "productosMenosVendidos": [
      {"nombreProducto": "Cazuela de Mariscos", "cantidadVendida": 12, "ingresosTotales": 600000.00},
      {"nombreProducto": "Churrasco Premium", "cantidadVendida": 18, "ingresosTotales": 1080000.00},
      {"nombreProducto": "Paella Valenciana", "cantidadVendida": 25, "ingresosTotales": 1375000.00}
    ],
    "frecuenciaProductos": [
      {"nombreProducto": "Bandeja Paisa", "frecuencia": 11.6},
      {"nombreProducto": "Sancocho de Gallina", "frecuencia": 10.2},
      {"nombreProducto": "Ajiaco Santafereño", "frecuencia": 7.8}
    ]
  }
}
```

---

## 4. 👥 Análisis de Usuarios

### **Endpoint:** `GET /users`

```bash
curl -X 'GET' \
  'http://localhost:8080/restaurante/v1/telemetria/users?limit=15&periodo=ultimos_6_meses' \
  -H 'accept: application/json' \
  -H 'Authorization: Bearer {token}'
```

**Parámetros:**
- `limit` (opcional): Número de usuarios a retornar (default: 10)
- `periodo` (opcional): Filtros temporales disponibles

**Respuesta esperada:**
```json
{
  "code": 200,
  "message": "Análisis de usuarios obtenido exitosamente",
  "data": {
    "usuariosFrecuentes": [
      {"nombreUsuario": "María González", "documento": 1015466494, "cantidadPedidos": 28, "totalGastado": 1260000.00},
      {"nombreUsuario": "Carlos Rodríguez", "documento": 1023456789, "cantidadPedidos": 24, "totalGastado": 1080000.00},
      {"nombreUsuario": "Ana Martínez", "documento": 1034567890, "cantidadPedidos": 21, "totalGastado": 945000.00}
    ],
    "usuariosInactivos": [
      {"nombreUsuario": "Pedro Sánchez", "documento": 1045678901, "cantidadPedidos": 1, "totalGastado": 45000.00},
      {"nombreUsuario": "Laura López", "documento": 1056789012, "cantidadPedidos": 2, "totalGastado": 78000.00}
    ],
    "totalGastadoPorUsuario": [
      {"nombreUsuario": "María González", "documento": 1015466494, "totalGastado": 1260000.00},
      {"nombreUsuario": "Carlos Rodríguez", "documento": 1023456789, "totalGastado": 1080000.00}
    ],
    "fechaUltimoPedido": [
      {"nombreUsuario": "María González", "documento": 1015466494, "fechaUltimoPedido": "2024-01-15T14:30:00Z"},
      {"nombreUsuario": "Carlos Rodríguez", "documento": 1023456789, "fechaUltimoPedido": "2024-01-14T19:45:00Z"}
    ]
  }
}
```

---

## 5. ⏰ Análisis Temporal

### **Endpoint:** `GET /time-analysis`

```bash
curl -X 'GET' \
  'http://localhost:8080/restaurante/v1/telemetria/time-analysis?periodo=ultimo_año' \
  -H 'accept: application/json' \
  -H 'Authorization: Bearer {token}'
```

**Parámetros:**
- `periodo` (opcional): Filtros temporales disponibles

**Respuesta esperada:**
```json
{
  "code": 200,
  "message": "Análisis temporal obtenido exitosamente",
  "data": {
    "pedidosPorHora": [
      {"hora": 12, "cantidadPedidos": 145, "porcentaje": 11.6},
      {"hora": 13, "cantidadPedidos": 168, "porcentaje": 13.4},
      {"hora": 19, "cantidadPedidos": 192, "porcentaje": 15.4}
    ],
    "pedidosPorDiaSemana": [
      {"diaSemana": "Viernes", "cantidadPedidos": 245, "porcentaje": 19.6},
      {"diaSemana": "Sábado", "cantidadPedidos": 268, "porcentaje": 21.4},
      {"diaSemana": "Domingo", "cantidadPedidos": 198, "porcentaje": 15.8}
    ],
    "ventasPorMes": [
      {"mes": "Enero", "ventas": 12450000.50, "pedidos": 456},
      {"mes": "Febrero", "ventas": 11280000.25, "pedidos": 412},
      {"mes": "Marzo", "ventas": 13670000.75, "pedidos": 498}
    ]
  }
}
```

---

## 6. 💼 Análisis de Rentabilidad

### **Endpoint:** `GET /rentabilidad`

```bash
curl -X 'GET' \
  'http://localhost:8080/restaurante/v1/telemetria/rentabilidad?limit=10&periodo=ultimo_mes' \
  -H 'accept: application/json' \
  -H 'Authorization: Bearer {token}'
```

**Parámetros:**
- `limit` (opcional): Número de productos a retornar (default: 10)
- `periodo` (opcional): Filtros temporales disponibles

**Respuesta esperada:**
```json
{
  "code": 200,
  "message": "Análisis de rentabilidad obtenido exitosamente",
  "data": {
    "productosMasRentables": [
      {"nombreProducto": "Bandeja Paisa", "margenBruto": 65.5, "ingresosTotales": 5220000.00, "cantidadVendida": 145},
      {"nombreProducto": "Sancocho de Gallina", "margenBruto": 58.2, "ingresosTotales": 4480000.00, "cantidadVendida": 128}
    ],
    "productosMenosRentables": [
      {"nombreProducto": "Cazuela de Mariscos", "margenBruto": 25.8, "ingresosTotales": 600000.00, "cantidadVendida": 12},
      {"nombreProducto": "Churrasco Premium", "margenBruto": 32.1, "ingresosTotales": 1080000.00, "cantidadVendida": 18}
    ],
    "estadisticas": {
      "margenPromedioGeneral": 52.3,
      "ingresosTotales": 45750000.50,
      "totalProductosAnalizados": 85
    }
  }
}
```

---

## 7. 🎯 Segmentación de Clientes

### **Endpoint:** `GET /segmentacion`

```bash
curl -X 'GET' \
  'http://localhost:8080/restaurante/v1/telemetria/segmentacion?limit=20&periodo=ultimos_6_meses' \
  -H 'accept: application/json' \
  -H 'Authorization: Bearer {token}'
```

**Parámetros:**
- `limit` (opcional): Número de clientes a retornar (default: 10)
- `periodo` (opcional): Filtros temporales disponibles

**Respuesta esperada:**
```json
{
  "code": 200,
  "message": "Segmentación de clientes obtenida exitosamente",
  "data": {
    "clientesVIP": [
      {"nombreCliente": "María González", "documento": 1015466494, "totalGastado": 1260000.00, "cantidadPedidos": 28, "valorPromedio": 45000.00},
      {"nombreCliente": "Carlos Rodríguez", "documento": 1023456789, "totalGastado": 1080000.00, "cantidadPedidos": 24, "valorPromedio": 45000.00}
    ],
    "clientesRegulares": [
      {"nombreCliente": "Ana Martínez", "documento": 1034567890, "totalGastado": 945000.00, "cantidadPedidos": 21, "valorPromedio": 45000.00},
      {"nombreCliente": "Luis Hernández", "documento": 1045678901, "totalGastado": 720000.00, "cantidadPedidos": 16, "valorPromedio": 45000.00}
    ],
    "clientesNuevos": [
      {"nombreCliente": "Pedro Sánchez", "documento": 1056789012, "totalGastado": 135000.00, "cantidadPedidos": 3, "valorPromedio": 45000.00},
      {"nombreCliente": "Laura López", "documento": 1067890123, "totalGastado": 90000.00, "cantidadPedidos": 2, "valorPromedio": 45000.00}
    ],
    "estadisticas": {
      "totalClientesVIP": 15,
      "totalClientesRegulares": 85,
      "totalClientesNuevos": 220,
      "valorPromedioVIP": 1125000.00,
      "valorPromedioRegular": 485000.00,
      "valorPromedioNuevo": 112500.00
    }
  }
}
```

---

## 8. ⚡ Análisis de Eficiencia

### **Endpoint:** `GET /eficiencia`

```bash
curl -X 'GET' \
  'http://localhost:8080/restaurante/v1/telemetria/eficiencia?limit=10&periodo=ultima_semana' \
  -H 'accept: application/json' \
  -H 'Authorization: Bearer {token}'
```

**Parámetros:**
- `limit` (opcional): Número de registros a retornar (default: 10)
- `periodo` (opcional): Filtros temporales disponibles

**Respuesta esperada:**
```json
{
  "code": 200,
  "message": "Análisis de eficiencia obtenido exitosamente",
  "data": {
    "tiemposEntrega": [
      {"fecha": "2024-01-15", "tiempoPromedioMinutos": 35.5, "pedidosEntregados": 45},
      {"fecha": "2024-01-14", "tiempoPromedioMinutos": 42.3, "pedidosEntregados": 38},
      {"fecha": "2024-01-13", "tiempoPromedioMinutos": 28.7, "pedidosEntregados": 52}
    ],
    "rendimientoTrabajadores": [
      {"nombreTrabajador": "Juan Pérez", "documento": 1098765432, "pedidosEntregados": 156, "tiempoPromedioMinutos": 32.5, "eficiencia": 92.3},
      {"nombreTrabajador": "Ana García", "documento": 1087654321, "pedidosEntregados": 142, "tiempoPromedioMinutos": 35.8, "eficiencia": 88.7}
    ],
    "eficienciaPorHora": [
      {"hora": 12, "pedidosCompletados": 25, "pedidosPendientes": 3, "eficiencia": 89.3},
      {"hora": 13, "pedidosCompletados": 32, "pedidosPendientes": 5, "eficiencia": 86.5},
      {"hora": 19, "pedidosCompletados": 28, "pedidosPendientes": 4, "eficiencia": 87.5}
    ],
    "estadisticas": {
      "tiempoPromedioGeneral": 34.2,
      "eficienciaPromedio": 88.5,
      "totalPedidosAnalizados": 1250,
      "trabajadoresActivos": 8
    }
  }
}
```

---

## 9. 📅 Análisis de Reservas

### **Endpoint:** `GET /reservas-analisis`

```bash
curl -X 'GET' \
  'http://localhost:8080/restaurante/v1/telemetria/reservas-analisis?periodo=ultimo_mes' \
  -H 'accept: application/json' \
  -H 'Authorization: Bearer {token}'
```

**Parámetros:**
- `periodo` (opcional): Filtros temporales disponibles

**Respuesta esperada:**
```json
{
  "code": 200,
  "message": "Análisis de reservas obtenido exitosamente",
  "data": {
    "reservasPorDia": [
      {"fecha": "2024-01-15", "reservasCompletadas": 28, "porcentaje": 8.5},
      {"fecha": "2024-01-14", "reservasCompletadas": 32, "porcentaje": 9.7},
      {"fecha": "2024-01-13", "reservasCompletadas": 25, "porcentaje": 7.6}
    ],
    "reservasPorHora": [
      {"hora": 12, "reservasCompletadas": 45, "porcentaje": 13.6},
      {"hora": 13, "reservasCompletadas": 52, "porcentaje": 15.8},
      {"hora": 19, "reservasCompletadas": 68, "porcentaje": 20.6}
    ],
    "reservasPorDiaSemana": [
      {"diaSemana": "Viernes", "reservasCompletadas": 58, "porcentaje": 17.6},
      {"diaSemana": "Sábado", "reservasCompletadas": 72, "porcentaje": 21.8},
      {"diaSemana": "Domingo", "reservasCompletadas": 45, "porcentaje": 13.6}
    ],
    "estadisticas": {
      "totalReservasCompletadas": 330,
      "promedioReservasDiarias": 11.0,
      "horaPico": 19,
      "diaPico": "Sábado"
    }
  }
}
```

---

## 10. 🛍️ Análisis de Pedidos

### **Endpoint:** `GET /pedidos-analisis`

```bash
curl -X 'GET' \
  'http://localhost:8080/restaurante/v1/telemetria/pedidos-analisis?periodo=ultimo_mes' \
  -H 'accept: application/json' \
  -H 'Authorization: Bearer {token}'
```

**Parámetros:**
- `periodo` (opcional): Filtros temporales disponibles

**Respuesta esperada:**
```json
{
  "code": 200,
  "message": "Análisis de pedidos obtenido exitosamente",
  "data": {
    "pedidosPorDia": [
      {"fecha": "2024-01-15", "pedidosCompletados": 45, "porcentaje": 3.6},
      {"fecha": "2024-01-14", "pedidosCompletados": 52, "porcentaje": 4.2},
      {"fecha": "2024-01-13", "pedidosCompletados": 38, "porcentaje": 3.0}
    ],
    "pedidosPorHora": [
      {"hora": 12, "pedidosCompletados": 145, "porcentaje": 11.6},
      {"hora": 13, "pedidosCompletados": 168, "porcentaje": 13.4},
      {"hora": 19, "pedidosCompletados": 192, "porcentaje": 15.4}
    ],
    "pedidosPorDiaSemana": [
      {"diaSemana": "Viernes", "pedidosCompletados": 245, "porcentaje": 19.6},
      {"diaSemana": "Sábado", "pedidosCompletados": 268, "porcentaje": 21.4},
      {"diaSemana": "Domingo", "pedidosCompletados": 198, "porcentaje": 15.8}
    ],
    "estadisticas": {
      "totalPedidosCompletados": 1250,
      "promedioPedidosDiarios": 41.7,
      "horaPico": 19,
      "diaPico": "Sábado"
    }
  }
}
```

---

## 🏗️ Modelos TypeScript

```typescript
// telemetryTypes.ts

export type TimePeriod =
  | 'hoy'
  | 'ultima_semana'
  | 'ultimo_mes'
  | 'ultimos_3_meses'
  | 'ultimos_6_meses'
  | 'ultimo_año'
  | 'historico';

export interface ApiResponse<T> {
  code: number;
  message: string;
  data: T;
  cause?: string;
}

// Dashboard Types
export interface DashboardData {
  totalPedidos: number;
  ingresosTotales: number;
  usuariosRegistrados: number;
  valorPromedioOrden: number;
}

// Sales Types
export interface MetodoPagoStats {
  metodoPago: string;
  cantidad?: number;
  ingresos?: number;
  porcentaje: number;
}

export interface TendenciaVenta {
  fecha: string;
  ventas: number;
  pedidos: number;
}

export interface SalesData {
  pedidosPorMetodoPago: MetodoPagoStats[];
  ingresosPorMetodoPago: MetodoPagoStats[];
  tendenciasVentas: TendenciaVenta[];
}

// Products Types
export interface ProductoVenta {
  nombreProducto: string;
  cantidadVendida: number;
  ingresosTotales: number;
}

export interface ProductoFrecuencia {
  nombreProducto: string;
  frecuencia: number;
}

export interface ProductsData {
  productosMasVendidos: ProductoVenta[];
  productosMenosVendidos: ProductoVenta[];
  frecuenciaProductos: ProductoFrecuencia[];
}

// Users Types
export interface UsuarioStats {
  nombreUsuario: string;
  documento: number;
  cantidadPedidos: number;
  totalGastado: number;
}

export interface UsuarioUltimoPedido {
  nombreUsuario: string;
  documento: number;
  fechaUltimoPedido: string;
}

export interface UsersData {
  usuariosFrecuentes: UsuarioStats[];
  usuariosInactivos: UsuarioStats[];
  totalGastadoPorUsuario: UsuarioStats[];
  fechaUltimoPedido: UsuarioUltimoPedido[];
}

// Time Analysis Types
export interface PedidosPorHora {
  hora: number;
  cantidadPedidos: number;
  porcentaje: number;
}

export interface PedidosPorDia {
  diaSemana: string;
  cantidadPedidos: number;
  porcentaje: number;
}

export interface VentasPorMes {
  mes: string;
  ventas: number;
  pedidos: number;
}

export interface TimeAnalysisData {
  pedidosPorHora: PedidosPorHora[];
  pedidosPorDiaSemana: PedidosPorDia[];
  ventasPorMes: VentasPorMes[];
}

// Rentabilidad Types
export interface ProductoRentabilidad {
  nombreProducto: string;
  margenBruto: number;
  ingresosTotales: number;
  cantidadVendida: number;
}

export interface EstadisticasRentabilidad {
  margenPromedioGeneral: number;
  ingresosTotales: number;
  totalProductosAnalizados: number;
}

export interface RentabilidadData {
  productosMasRentables: ProductoRentabilidad[];
  productosMenosRentables: ProductoRentabilidad[];
  estadisticas: EstadisticasRentabilidad;
}

// Segmentación Types
export interface ClienteSegmento {
  nombreCliente: string;
  documento: number;
  totalGastado: number;
  cantidadPedidos: number;
  valorPromedio: number;
}

export interface EstadisticasSegmentacion {
  totalClientesVIP: number;
  totalClientesRegulares: number;
  totalClientesNuevos: number;
  valorPromedioVIP: number;
  valorPromedioRegular: number;
  valorPromedioNuevo: number;
}

export interface SegmentacionData {
  clientesVIP: ClienteSegmento[];
  clientesRegulares: ClienteSegmento[];
  clientesNuevos: ClienteSegmento[];
  estadisticas: EstadisticasSegmentacion;
}

// Eficiencia Types
export interface TiempoEntrega {
  fecha: string;
  tiempoPromedioMinutos: number;
  pedidosEntregados: number;
}

export interface RendimientoTrabajador {
  nombreTrabajador: string;
  documento: number;
  pedidosEntregados: number;
  tiempoPromedioMinutos: number;
  eficiencia: number;
}

export interface EficienciaPorHora {
  hora: number;
  pedidosCompletados: number;
  pedidosPendientes: number;
  eficiencia: number;
}

export interface EstadisticasEficiencia {
  tiempoPromedioGeneral: number;
  eficienciaPromedio: number;
  totalPedidosAnalizados: number;
  trabajadoresActivos: number;
}

export interface EficienciaData {
  tiemposEntrega: TiempoEntrega[];
  rendimientoTrabajadores: RendimientoTrabajador[];
  eficienciaPorHora: EficienciaPorHora[];
  estadisticas: EstadisticasEficiencia;
}

// Reservas Analysis Types
export interface ReservaPorDia {
  fecha: string;
  reservasCompletadas: number;
  porcentaje: number;
}

export interface ReservaPorHora {
  hora: number;
  reservasCompletadas: number;
  porcentaje: number;
}

export interface ReservaPorDiaSemana {
  diaSemana: string;
  reservasCompletadas: number;
  porcentaje: number;
}

export interface EstadisticasReservas {
  totalReservasCompletadas: number;
  promedioReservasDiarias: number;
  horaPico: number;
  diaPico: string;
}

export interface ReservasAnalisisData {
  reservasPorDia: ReservaPorDia[];
  reservasPorHora: ReservaPorHora[];
  reservasPorDiaSemana: ReservaPorDiaSemana[];
  estadisticas: EstadisticasReservas;
}

// Pedidos Analysis Types
export interface PedidoPorDia {
  fecha: string;
  pedidosCompletados: number;
  porcentaje: number;
}

export interface PedidoPorHora {
  hora: number;
  pedidosCompletados: number;
  porcentaje: number;
}

export interface PedidoPorDiaSemana {
  diaSemana: string;
  pedidosCompletados: number;
  porcentaje: number;
}

export interface EstadisticasPedidos {
  totalPedidosCompletados: number;
  promedioPedidosDiarios: number;
  horaPico: number;
  diaPico: string;
}

export interface PedidosAnalisisData {
  pedidosPorDia: PedidoPorDia[];
  pedidosPorHora: PedidoPorHora[];
  pedidosPorDiaSemana: PedidoPorDiaSemana[];
  estadisticas: EstadisticasPedidos;
}

// Service Parameters
export interface TelemetryParams {
  periodo?: TimePeriod;
  limit?: number;
}

// Productos Populares Types (Endpoint Público)
export interface ProductoVendido {
  productoId: number;
  nombreProducto: string;
  cantidadVendida: number;
  ingresoTotal: number;
  precio: number;
  imagen: string; // Base64 encoded image
}

export interface ProductosPopularesData {
  productosPopulares: ProductoVendido[];
}
```

---

## 🎭 Mocks de Ejemplo

```typescript
// telemetryMocks.ts

import {
  DashboardData,
  SalesData,
  ProductsData,
  UsersData,
  TimeAnalysisData,
  RentabilidadData,
  SegmentacionData,
  EficienciaData,
  ReservasAnalisisData,
  PedidosAnalisisData,
  ApiResponse
} from './telemetryTypes';

export const mockDashboardData: ApiResponse<DashboardData> = {
  code: 200,
  message: "Dashboard obtenido exitosamente",
  data: {
    totalPedidos: 1250,
    ingresosTotales: 45750000.50,
    usuariosRegistrados: 320,
    valorPromedioOrden: 36600.40
  }
};

export const mockSalesData: ApiResponse<SalesData> = {
  code: 200,
  message: "Análisis de ventas obtenido exitosamente",
  data: {
    pedidosPorMetodoPago: [
      { metodoPago: "EFECTIVO", cantidad: 450, porcentaje: 36.0 },
      { metodoPago: "TARJETA", cantidad: 520, porcentaje: 41.6 },
      { metodoPago: "TRANSFERENCIA", cantidad: 280, porcentaje: 22.4 }
    ],
    ingresosPorMetodoPago: [
      { metodoPago: "EFECTIVO", ingresos: 16470000.00, porcentaje: 36.0 },
      { metodoPago: "TARJETA", ingresos: 19032000.00, porcentaje: 41.6 },
      { metodoPago: "TRANSFERENCIA", ingresos: 10248000.00, porcentaje: 22.4 }
    ],
    tendenciasVentas: [
      { fecha: "2024-01-01", ventas: 1250000.50, pedidos: 45 },
      { fecha: "2024-01-02", ventas: 980000.25, pedidos: 38 },
      { fecha: "2024-01-03", ventas: 1450000.75, pedidos: 52 }
    ]
  }
};

export const mockProductsData: ApiResponse<ProductsData> = {
  code: 200,
  message: "Análisis de productos obtenido exitosamente",
  data: {
    productosMasVendidos: [
      { nombreProducto: "Bandeja Paisa", cantidadVendida: 145, ingresosTotales: 5220000.00 },
      { nombreProducto: "Sancocho de Gallina", cantidadVendida: 128, ingresosTotales: 4480000.00 },
      { nombreProducto: "Ajiaco Santafereño", cantidadVendida: 98, ingresosTotales: 3430000.00 }
    ],
    productosMenosVendidos: [
      { nombreProducto: "Cazuela de Mariscos", cantidadVendida: 12, ingresosTotales: 600000.00 },
      { nombreProducto: "Churrasco Premium", cantidadVendida: 18, ingresosTotales: 1080000.00 },
      { nombreProducto: "Paella Valenciana", cantidadVendida: 25, ingresosTotales: 1375000.00 }
    ],
    frecuenciaProductos: [
      { nombreProducto: "Bandeja Paisa", frecuencia: 11.6 },
      { nombreProducto: "Sancocho de Gallina", frecuencia: 10.2 },
      { nombreProducto: "Ajiaco Santafereño", frecuencia: 7.8 }
    ]
  }
};

export const mockUsersData: ApiResponse<UsersData> = {
  code: 200,
  message: "Análisis de usuarios obtenido exitosamente",
  data: {
    usuariosFrecuentes: [
      { nombreUsuario: "María González", documento: 1015466494, cantidadPedidos: 28, totalGastado: 1260000.00 },
      { nombreUsuario: "Carlos Rodríguez", documento: 1023456789, cantidadPedidos: 24, totalGastado: 1080000.00 },
      { nombreUsuario: "Ana Martínez", documento: 1034567890, cantidadPedidos: 21, totalGastado: 945000.00 }
    ],
    usuariosInactivos: [
      { nombreUsuario: "Pedro Sánchez", documento: 1045678901, cantidadPedidos: 1, totalGastado: 45000.00 },
      { nombreUsuario: "Laura López", documento: 1056789012, cantidadPedidos: 2, totalGastado: 78000.00 }
    ],
    totalGastadoPorUsuario: [
      { nombreUsuario: "María González", documento: 1015466494, totalGastado: 1260000.00 },
      { nombreUsuario: "Carlos Rodríguez", documento: 1023456789, totalGastado: 1080000.00 }
    ],
    fechaUltimoPedido: [
      { nombreUsuario: "María González", documento: 1015466494, fechaUltimoPedido: "2024-01-15T14:30:00Z" },
      { nombreUsuario: "Carlos Rodríguez", documento: 1023456789, fechaUltimoPedido: "2024-01-14T19:45:00Z" }
    ]
  }
};

export const mockTimeAnalysisData: ApiResponse<TimeAnalysisData> = {
  code: 200,
  message: "Análisis temporal obtenido exitosamente",
  data: {
    pedidosPorHora: [
      { hora: 12, cantidadPedidos: 145, porcentaje: 11.6 },
      { hora: 13, cantidadPedidos: 168, porcentaje: 13.4 },
      { hora: 19, cantidadPedidos: 192, porcentaje: 15.4 }
    ],
    pedidosPorDiaSemana: [
      { diaSemana: "Viernes", cantidadPedidos: 245, porcentaje: 19.6 },
      { diaSemana: "Sábado", cantidadPedidos: 268, porcentaje: 21.4 },
      { diaSemana: "Domingo", cantidadPedidos: 198, porcentaje: 15.8 }
    ],
    ventasPorMes: [
      { mes: "Enero", ventas: 12450000.50, pedidos: 456 },
      { mes: "Febrero", ventas: 11280000.25, pedidos: 412 },
      { mes: "Marzo", ventas: 13670000.75, pedidos: 498 }
    ]
  }
};

export const mockRentabilidadData: ApiResponse<RentabilidadData> = {
  code: 200,
  message: "Análisis de rentabilidad obtenido exitosamente",
  data: {
    productosMasRentables: [
      { nombreProducto: "Bandeja Paisa", margenBruto: 65.5, ingresosTotales: 5220000.00, cantidadVendida: 145 },
      { nombreProducto: "Sancocho de Gallina", margenBruto: 58.2, ingresosTotales: 4480000.00, cantidadVendida: 128 }
    ],
    productosMenosRentables: [
      { nombreProducto: "Cazuela de Mariscos", margenBruto: 25.8, ingresosTotales: 600000.00, cantidadVendida: 12 },
      { nombreProducto: "Churrasco Premium", margenBruto: 32.1, ingresosTotales: 1080000.00, cantidadVendida: 18 }
    ],
    estadisticas: {
      margenPromedioGeneral: 52.3,
      ingresosTotales: 45750000.50,
      totalProductosAnalizados: 85
    }
  }
};

export const mockSegmentacionData: ApiResponse<SegmentacionData> = {
  code: 200,
  message: "Segmentación de clientes obtenida exitosamente",
  data: {
    clientesVIP: [
      { nombreCliente: "María González", documento: 1015466494, totalGastado: 1260000.00, cantidadPedidos: 28, valorPromedio: 45000.00 },
      { nombreCliente: "Carlos Rodríguez", documento: 1023456789, totalGastado: 1080000.00, cantidadPedidos: 24, valorPromedio: 45000.00 }
    ],
    clientesRegulares: [
      { nombreCliente: "Ana Martínez", documento: 1034567890, totalGastado: 945000.00, cantidadPedidos: 21, valorPromedio: 45000.00 },
      { nombreCliente: "Luis Hernández", documento: 1045678901, totalGastado: 720000.00, cantidadPedidos: 16, valorPromedio: 45000.00 }
    ],
    clientesNuevos: [
      { nombreCliente: "Pedro Sánchez", documento: 1056789012, totalGastado: 135000.00, cantidadPedidos: 3, valorPromedio: 45000.00 },
      { nombreCliente: "Laura López", documento: 1067890123, totalGastado: 90000.00, cantidadPedidos: 2, valorPromedio: 45000.00 }
    ],
    estadisticas: {
      totalClientesVIP: 15,
      totalClientesRegulares: 85,
      totalClientesNuevos: 220,
      valorPromedioVIP: 1125000.00,
      valorPromedioRegular: 485000.00,
      valorPromedioNuevo: 112500.00
    }
  }
};

export const mockEficienciaData: ApiResponse<EficienciaData> = {
  code: 200,
  message: "Análisis de eficiencia obtenido exitosamente",
  data: {
    tiemposEntrega: [
      { fecha: "2024-01-15", tiempoPromedioMinutos: 35.5, pedidosEntregados: 45 },
      { fecha: "2024-01-14", tiempoPromedioMinutos: 42.3, pedidosEntregados: 38 },
      { fecha: "2024-01-13", tiempoPromedioMinutos: 28.7, pedidosEntregados: 52 }
    ],
    rendimientoTrabajadores: [
      { nombreTrabajador: "Juan Pérez", documento: 1098765432, pedidosEntregados: 156, tiempoPromedioMinutos: 32.5, eficiencia: 92.3 },
      { nombreTrabajador: "Ana García", documento: 1087654321, pedidosEntregados: 142, tiempoPromedioMinutos: 35.8, eficiencia: 88.7 }
    ],
    eficienciaPorHora: [
      { hora: 12, pedidosCompletados: 25, pedidosPendientes: 3, eficiencia: 89.3 },
      { hora: 13, pedidosCompletados: 32, pedidosPendientes: 5, eficiencia: 86.5 },
      { hora: 19, pedidosCompletados: 28, pedidosPendientes: 4, eficiencia: 87.5 }
    ],
    estadisticas: {
      tiempoPromedioGeneral: 34.2,
      eficienciaPromedio: 88.5,
      totalPedidosAnalizados: 1250,
      trabajadoresActivos: 8
    }
  }
};

export const mockReservasAnalisisData: ApiResponse<ReservasAnalisisData> = {
  code: 200,
  message: "Análisis de reservas obtenido exitosamente",
  data: {
    reservasPorDia: [
      { fecha: "2024-01-15", reservasCompletadas: 28, porcentaje: 8.5 },
      { fecha: "2024-01-14", reservasCompletadas: 32, porcentaje: 9.7 },
      { fecha: "2024-01-13", reservasCompletadas: 25, porcentaje: 7.6 }
    ],
    reservasPorHora: [
      { hora: 12, reservasCompletadas: 45, porcentaje: 13.6 },
      { hora: 13, reservasCompletadas: 52, porcentaje: 15.8 },
      { hora: 19, reservasCompletadas: 68, porcentaje: 20.6 }
    ],
    reservasPorDiaSemana: [
      { diaSemana: "Viernes", reservasCompletadas: 58, porcentaje: 17.6 },
      { diaSemana: "Sábado", reservasCompletadas: 72, porcentaje: 21.8 },
      { diaSemana: "Domingo", reservasCompletadas: 45, porcentaje: 13.6 }
    ],
    estadisticas: {
      totalReservasCompletadas: 330,
      promedioReservasDiarias: 11.0,
      horaPico: 19,
      diaPico: "Sábado"
    }
  }
};

export const mockPedidosAnalisisData: ApiResponse<PedidosAnalisisData> = {
  code: 200,
  message: "Análisis de pedidos obtenido exitosamente",
  data: {
    pedidosPorDia: [
      { fecha: "2024-01-15", pedidosCompletados: 45, porcentaje: 3.6 },
      { fecha: "2024-01-14", pedidosCompletados: 52, porcentaje: 4.2 },
      { fecha: "2024-01-13", pedidosCompletados: 38, porcentaje: 3.0 }
    ],
    pedidosPorHora: [
      { hora: 12, pedidosCompletados: 145, porcentaje: 11.6 },
      { hora: 13, pedidosCompletados: 168, porcentaje: 13.4 },
      { hora: 19, pedidosCompletados: 192, porcentaje: 15.4 }
    ],
    pedidosPorDiaSemana: [
      { diaSemana: "Viernes", pedidosCompletados: 245, porcentaje: 19.6 },
      { diaSemana: "Sábado", pedidosCompletados: 268, porcentaje: 21.4 },
      { diaSemana: "Domingo", pedidosCompletados: 198, porcentaje: 15.8 }
    ],
    estadisticas: {
      totalPedidosCompletados: 1250,
      promedioPedidosDiarios: 41.7,
      horaPico: 19,
      diaPico: "Sábado"
    }
  }
};

// Función helper para simular delay de red
export const mockDelay = (ms: number = 1000) =>
  new Promise(resolve => setTimeout(resolve, ms));

// Mock de error para testing
export const mockErrorResponse = {
  code: 500,
  message: "Error interno del servidor",
  cause: "Error de conexión a la base de datos"
};

// Mock para productos populares (endpoint público)
export const mockProductosPopularesData: ApiResponse<ProductosPopularesData> = {
  code: 200,
  message: "Productos populares (ultimo_mes) obtenidos exitosamente",
  data: {
    productosPopulares: [
      {
        productoId: 1,
        nombreProducto: "Bandeja Paisa",
        cantidadVendida: 145,
        ingresoTotal: 5220000,
        precio: 36000,
        imagen: "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mP8/5+hHgAHggJ/PchI7wAAAABJRU5ErkJggg=="
      },
      {
        productoId: 2,
        nombreProducto: "Sancocho de Gallina",
        cantidadVendida: 128,
        ingresoTotal: 4480000,
        precio: 35000,
        imagen: "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mP8/5+hHgAHggJ/PchI7wAAAABJRU5ErkJggg=="
      },
      {
        productoId: 3,
        nombreProducto: "Ajiaco Santafereño",
        cantidadVendida: 98,
        ingresoTotal: 3430000,
        precio: 35000,
        imagen: "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mP8/5+hHgAHggJ/PchI7wAAAABJRU5ErkJggg=="
      },
      {
        productoId: 4,
        nombreProducto: "Arroz con Pollo",
        cantidadVendida: 87,
        ingresoTotal: 2610000,
        precio: 30000,
        imagen: "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mP8/5+hHgAHggJ/PchI7wAAAABJRU5ErkJggg=="
      }
    ]
  }
};
```

---

## 🔧 Configuración del Servicio

```typescript
// telemetryService.ts - Estructura base esperada

class TelemetryService {
  private baseURL = 'http://localhost:8080/restaurante/v1';

  // Métodos esperados:

  // 🌟 NUEVO: Método público (sin autenticación)
  async getProductosPopulares(params?: TelemetryParams): Promise<ApiResponse<ProductosPopularesData>>

  // Métodos protegidos (requieren autenticación admin)
  async getDashboard(params?: TelemetryParams): Promise<ApiResponse<DashboardData>>
  async getSales(params?: TelemetryParams): Promise<ApiResponse<SalesData>>
  async getProducts(params?: TelemetryParams): Promise<ApiResponse<ProductsData>>
  async getUsers(params?: TelemetryParams): Promise<ApiResponse<UsersData>>
  async getTimeAnalysis(params?: TelemetryParams): Promise<ApiResponse<TimeAnalysisData>>
  async getRentabilidad(params?: TelemetryParams): Promise<ApiResponse<RentabilidadData>>
  async getSegmentacion(params?: TelemetryParams): Promise<ApiResponse<SegmentacionData>>
  async getEficiencia(params?: TelemetryParams): Promise<ApiResponse<EficienciaData>>
  async getReservasAnalisis(params?: TelemetryParams): Promise<ApiResponse<ReservasAnalisisData>>
  async getPedidosAnalisis(params?: TelemetryParams): Promise<ApiResponse<PedidosAnalisisData>>
}
```

---

## 🎨 Componentes de UI Sugeridos

```typescript
// Resultados esperados en el componente de teletemetría

- TelemetryDashboard: Vista principal con KPIs
- SalesChart: Gráficos de ventas y métodos de pago
- ProductsRanking: Top productos más/menos vendidos
- UsersSegmentation: Análisis de clientes VIP/regulares/nuevos
- TimeAnalysisCharts: Gráficos por hora/día/mes
- EfficiencyMetrics: Métricas de rendimiento y tiempos
- ReservationsAnalysis: Análisis de reservas por tiempo
- OrdersAnalysis: Análisis de pedidos completados
- ProfitabilityChart: Análisis de rentabilidad por producto
- PeriodFilter: Selector de filtros temporales
```

---

## 📝 Notas Importantes

1. **Autenticación**: Todos los endpoints requieren token JWT válido con rol "Administrador"
2. **Filtros temporales**: Todos los endpoints soportan el parámetro `periodo`
3. **Paginación**: Algunos endpoints soportan el parámetro `limit`
4. **Formato de fechas**: Las fechas se devuelven en formato ISO 8601
5. **Moneda**: Los valores monetarios están en pesos colombianos (COP)
6. **Porcentajes**: Se devuelven como números decimales (ej: 36.5 = 36.5%)

---

## 🚀 ¡Listo para implementar!

Con esta guía tienes todo lo necesario para crear un frontend completo de telemetría que consuma todos los endpoints del sistema. ¡Manos a la obra! 🎯
