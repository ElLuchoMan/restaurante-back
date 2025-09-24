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

## 📖 **Convenciones de Campos y Datos**

### **💰 Campos Monetarios**
- **Todos los valores monetarios están en pesos colombianos (COP)**
- **Formato**: Números enteros sin decimales (ej: `45750000` = $45,750,000 COP)
- **Campos típicos**: `totalIngresos`, `ingresoTotal`, `totalGastado`, `precio`, `gananciaTotal`

### **📅 Campos de Fecha y Hora**
- **Fechas**: Formato `YYYY-MM-DD` (ej: `"2024-01-15"`)
- **Fechas con hora**: Formato ISO 8601 con UTC (ej: `"2024-01-15T14:30:00Z"`)
- **Horas**: Formato 24 horas `HH:MM:SS` (ej: `"14:30:00"`)

### **🔢 Campos Numéricos**
- **Cantidades**: Números enteros (ej: `totalPedidos: 1250`)
- **Porcentajes**: Números decimales representando el valor real (ej: `65.5` = 65.5%)
- **Promedios**: Números decimales con precisión (ej: `36600.40`)

### **👤 Identificadores**
- **IDs de productos**: `productoId` (número entero)
- **Documentos de clientes**: `documentoCliente` (número entero)
- **Documentos de trabajadores**: `documentoTrabajador` (número entero)

### **📊 Estructura de Respuestas**
Todas las respuestas siguen el formato estándar:
```json
{
  "code": 200,           // Código de estado HTTP
  "message": "...",      // Mensaje descriptivo en español
  "data": { ... },       // Datos de la respuesta
  "cause": "..."         // Solo presente en errores
}
```

---

## 🕒 **NUEVO: Filtros Temporales Avanzados**

Todos los endpoints de telemetría ahora soportan **3 tipos de filtros temporales** para análisis flexible:

### **1. Filtros Predefinidos (periodo)**
```bash
?periodo=ultimo_mes
```
- `hoy`: Solo datos del día actual
- `ultima_semana`: Últimos 7 días
- `ultimo_mes`: Últimos 30 días (por defecto)
- `ultimos_3_meses`: Últimos 90 días
- `ultimos_6_meses`: Últimos 180 días
- `ultimo_año`: Últimos 365 días
- `historico`: Todos los datos disponibles

### **2. Filtros por Mes y Año**
```bash
?mes=1&año=2024          # Enero 2024
?mes=12&año=2023         # Diciembre 2023
?año=2024                # Todo el año 2024
?mes=6                   # Junio del año actual
```

### **3. Filtros por Rango de Fechas y Horas**
```bash
# Rango de fechas
?fecha_inicio=2024-01-01&fecha_fin=2024-01-31

# Rango de fechas con horarios específicos
?fecha_inicio=2024-01-01&fecha_fin=2024-01-31&hora_inicio=08:00:00&hora_fin=18:00:00

# Solo filtro por horas (cualquier día)
?hora_inicio=12:00:00&hora_fin=14:00:00
```

### **🎯 Ejemplos de Uso Combinado:**

```bash
# Dashboard de enero 2024
curl -X 'GET' \
  'http://localhost:8080/restaurante/v1/telemetria/dashboard?mes=1&año=2024' \
  -H 'Authorization: Bearer {token}'

# Ventas entre fechas específicas
curl -X 'GET' \
  'http://localhost:8080/restaurante/v1/telemetria/sales?fecha_inicio=2024-01-15&fecha_fin=2024-02-15' \
  -H 'Authorization: Bearer {token}'

# Productos más vendidos en horario de almuerzo
curl -X 'GET' \
  'http://localhost:8080/restaurante/v1/telemetria/products?hora_inicio=12:00:00&hora_fin=14:00:00&limit=5' \
  -H 'Authorization: Bearer {token}'

# Análisis de eficiencia en fin de semana de diciembre 2023
curl -X 'GET' \
  'http://localhost:8080/restaurante/v1/telemetria/eficiencia?mes=12&año=2023&fecha_inicio=2023-12-02&fecha_fin=2023-12-31' \
  -H 'Authorization: Bearer {token}'
```

### **⚠️ Notas Importantes:**
- Los filtros se aplican en orden de prioridad: **Mes/Año** > **Rango de Fechas** > **Periodo Predefinido**
- Los filtros de hora se pueden combinar con cualquier filtro de fecha
- El campo `totalUsuarios` en el dashboard ahora refleja usuarios activos en el período filtrado
- Todos los análisis temporales respetan los filtros aplicados

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
- `mes` (opcional): Mes específico (1-12) para filtrar por mes y año
- `año` (opcional): Año específico (ej: 2024) para filtrar por mes y año
- `fecha_inicio` (opcional): Fecha de inicio para rango personalizado (YYYY-MM-DD)
- `fecha_fin` (opcional): Fecha de fin para rango personalizado (YYYY-MM-DD)
- `hora_inicio` (opcional): Hora de inicio para filtro horario (HH:MM:SS)
- `hora_fin` (opcional): Hora de fin para filtro horario (HH:MM:SS)

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
- `mes` (opcional): Mes específico (1-12) para filtrar por mes y año
- `año` (opcional): Año específico (ej: 2024) para filtrar por mes y año
- `fecha_inicio` (opcional): Fecha de inicio para rango personalizado (YYYY-MM-DD)
- `fecha_fin` (opcional): Fecha de fin para rango personalizado (YYYY-MM-DD)
- `hora_inicio` (opcional): Hora de inicio para filtro horario (HH:MM:SS)
- `hora_fin` (opcional): Hora de fin para filtro horario (HH:MM:SS)

**Respuesta esperada:**
```json
{
  "code": 200,
  "message": "Dashboard obtenido exitosamente",
  "data": {
    "totalPedidos": 1250,        // Total de pedidos en el período filtrado
    "totalIngresos": 45750000,   // Ingresos totales en pesos colombianos
    "totalUsuarios": 320,        // Usuarios activos en el período filtrado
    "promedioVentaPedido": 36600.40,  // Valor promedio por pedido (ticket promedio)
    "pedidosHoy": 45,           // Pedidos realizados hoy específicamente
    "ingresosHoy": 1650000      // Ingresos generados hoy específicamente
  }
}
```

**📋 Descripción de campos:**
- **`totalPedidos`**: Cantidad total de pedidos completados en el período seleccionado
- **`totalIngresos`**: Suma de todos los ingresos en pesos colombianos (COP)
- **`totalUsuarios`**: Número de usuarios únicos que realizaron pedidos en el período
- **`promedioVentaPedido`**: Valor promedio por pedido (totalIngresos ÷ totalPedidos)
- **`pedidosHoy`**: Pedidos del día actual (independiente del filtro de período)
- **`ingresosHoy`**: Ingresos del día actual (independiente del filtro de período)

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
    "ventasPorMetodoPago": [
      {
        "metodoPago": "EFECTIVO",     // Método de pago utilizado
        "total": 16470000,           // Ingresos totales por este método (COP)
        "cantidad": 450              // Número de pedidos pagados con este método
      },
      {"metodoPago": "TARJETA", "total": 19032000, "cantidad": 520},
      {"metodoPago": "TRANSFERENCIA", "total": 10248000, "cantidad": 280}
    ],
    "tendenciaVentas": [
      {
        "fecha": "2024-01-01",       // Fecha del análisis (YYYY-MM-DD)
        "total": 1250000,           // Ingresos totales del día (COP)
        "cantidad": 45              // Número de pedidos completados
      },
      {"fecha": "2024-01-02", "total": 980000, "cantidad": 38},
      {"fecha": "2024-01-03", "total": 1450000, "cantidad": 52}
    ],
    "estadisticasGenerales": {
      "ventaPromedioDiaria": 1226666.67,   // Promedio de ingresos por día (COP)
      "pedidoPromedioDiario": 44.33,       // Promedio de pedidos por día
      "ticketPromedio": 27666.67           // Valor promedio por pedido (COP)
    }
  }
}
```

**📋 Descripción de campos:**

**`ventasPorMetodoPago`**: Desglose de ventas por método de pago
- **`metodoPago`**: Tipo de pago (EFECTIVO, TARJETA, TRANSFERENCIA, etc.)
- **`total`**: Ingresos totales generados por este método de pago
- **`cantidad`**: Número de transacciones realizadas con este método

**`tendenciaVentas`**: Evolución de ventas día a día
- **`fecha`**: Fecha específica en formato YYYY-MM-DD
- **`total`**: Ingresos totales del día
- **`cantidad`**: Número de pedidos completados en el día

**`estadisticasGenerales`**: Métricas calculadas del período
- **`ventaPromedioDiaria`**: Ingresos promedio por día del período
- **`pedidoPromedioDiario`**: Número promedio de pedidos por día
- **`ticketPromedio`**: Valor promedio por pedido (total ingresos ÷ total pedidos)

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
- `periodo` (opcional): `hoy` | `ultima_semana` | `ultimo_mes` | `ultimos_3_meses` | `ultimos_6_meses` | `ultimo_año` | `historico`
- `mes` (opcional): Mes específico (1-12) para filtrar por mes y año
- `año` (opcional): Año específico (ej: 2024) para filtrar por mes y año
- `fecha_inicio` (opcional): Fecha de inicio para rango personalizado (YYYY-MM-DD)
- `fecha_fin` (opcional): Fecha de fin para rango personalizado (YYYY-MM-DD)
- `hora_inicio` (opcional): Hora de inicio para filtro horario (HH:MM:SS)
- `hora_fin` (opcional): Hora de fin para filtro horario (HH:MM:SS)

**Respuesta esperada:**
```json
{
  "code": 200,
  "message": "Análisis de productos obtenido exitosamente",
  "data": {
    "productosMasVendidos": [
      {
        "productoId": 1,                    // ID único del producto
        "nombreProducto": "Bandeja Paisa",  // Nombre del producto
        "cantidadVendida": 145,            // Unidades vendidas en el período
        "ingresoTotal": 5220000,           // Ingresos totales generados (COP)
        "precio": 36000,                   // Precio unitario actual (COP)
        "imagen": "base64_string"          // Imagen del producto en base64
      },
      {"productoId": 2, "nombreProducto": "Sancocho de Gallina", "cantidadVendida": 128, "ingresoTotal": 4480000, "precio": 35000, "imagen": "base64_string"},
      {"productoId": 3, "nombreProducto": "Ajiaco Santafereño", "cantidadVendida": 98, "ingresoTotal": 3430000, "precio": 35000, "imagen": "base64_string"}
    ],
    "productosMenosVendidos": [
      {"productoId": 4, "nombreProducto": "Cazuela de Mariscos", "cantidadVendida": 12, "ingresoTotal": 600000, "precio": 50000, "imagen": "base64_string"},
      {"productoId": 5, "nombreProducto": "Churrasco Premium", "cantidadVendida": 18, "ingresoTotal": 1080000, "precio": 60000, "imagen": "base64_string"},
      {"productoId": 6, "nombreProducto": "Paella Valenciana", "cantidadVendida": 25, "ingresoTotal": 1375000, "precio": 55000, "imagen": "base64_string"}
    ],
    "estadisticasProductos": {
      "totalProductosActivos": 85,                    // Total de productos disponibles
      "productoConMasVentas": "Bandeja Paisa",       // Producto más vendido del período
      "productoConMenosVentas": "Cazuela de Mariscos" // Producto menos vendido del período
    }
  }
}
```

**📋 Descripción de campos:**

**`productosMasVendidos` / `productosMenosVendidos`**: Lista de productos ordenados por ventas
- **`productoId`**: Identificador único del producto en la base de datos
- **`nombreProducto`**: Nombre comercial del producto
- **`cantidadVendida`**: Número de unidades vendidas en el período filtrado
- **`ingresoTotal`**: Ingresos totales generados por este producto (cantidadVendida × precio)
- **`precio`**: Precio unitario actual del producto en pesos colombianos
- **`imagen`**: Imagen del producto codificada en base64 (puede ser null)

**`estadisticasProductos`**: Resumen estadístico de productos
- **`totalProductosActivos`**: Número total de productos disponibles en el menú
- **`productoConMasVentas`**: Nombre del producto con mayor cantidad de ventas
- **`productoConMenosVentas`**: Nombre del producto con menor cantidad de ventas

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
- `periodo` (opcional): `hoy` | `ultima_semana` | `ultimo_mes` | `ultimos_3_meses` | `ultimos_6_meses` | `ultimo_año` | `historico`
- `mes` (opcional): Mes específico (1-12) para filtrar por mes y año
- `año` (opcional): Año específico (ej: 2024) para filtrar por mes y año
- `fecha_inicio` (opcional): Fecha de inicio para rango personalizado (YYYY-MM-DD)
- `fecha_fin` (opcional): Fecha de fin para rango personalizado (YYYY-MM-DD)
- `hora_inicio` (opcional): Hora de inicio para filtro horario (HH:MM:SS)
- `hora_fin` (opcional): Hora de fin para filtro horario (HH:MM:SS)

**Respuesta esperada:**
```json
{
  "code": 200,
  "message": "Análisis de usuarios obtenido exitosamente",
  "data": {
    "usuariosFrecuentes": [
      {
        "documentoCliente": 1015466494,           // Número de documento del cliente
        "nombreCompleto": "María González",       // Nombre completo del cliente
        "totalPedidos": 28,                      // Total de pedidos realizados
        "totalGastado": 1260000,                 // Total gastado en el período (COP)
        "ultimoPedido": "2024-01-15T14:30:00Z"   // Fecha y hora del último pedido (ISO 8601)
      },
      {"documentoCliente": 1023456789, "nombreCompleto": "Carlos Rodríguez", "totalPedidos": 24, "totalGastado": 1080000, "ultimoPedido": "2024-01-14T19:45:00Z"},
      {"documentoCliente": 1034567890, "nombreCompleto": "Ana Martínez", "totalPedidos": 21, "totalGastado": 945000, "ultimoPedido": "2024-01-13T12:15:00Z"}
    ],
    "usuariosInactivos": [
      {
        "documentoCliente": 1045678901,           // Documento del cliente inactivo
        "nombreCompleto": "Pedro Sánchez",        // Nombre del cliente
        "totalPedidos": 1,                       // Pocos pedidos realizados
        "ultimoPedido": "2023-12-01T10:30:00Z"   // Último pedido hace mucho tiempo
      },
      {"documentoCliente": 1056789012, "nombreCompleto": "Laura López", "totalPedidos": 2, "ultimoPedido": "2023-11-15T16:45:00Z"}
    ],
    "estadisticasUsuarios": {
      "totalClientes": 320,                      // Total de clientes registrados
      "clientesActivos": 285,                    // Clientes con pedidos en el período
      "clientesInactivos": 35,                   // Clientes sin pedidos recientes
      "promedioGastoPorCliente": 156250.00       // Gasto promedio por cliente (COP)
    }
  }
}
```

**📋 Descripción de campos:**

**`usuariosFrecuentes`**: Clientes con mayor actividad en el período
- **`documentoCliente`**: Número de identificación del cliente
- **`nombreCompleto`**: Nombre y apellido del cliente
- **`totalPedidos`**: Cantidad de pedidos realizados en el período filtrado
- **`totalGastado`**: Suma total gastada por el cliente en pesos colombianos
- **`ultimoPedido`**: Fecha y hora del pedido más reciente en formato ISO 8601

**`usuariosInactivos`**: Clientes con poca actividad reciente
- **`documentoCliente`**: Número de identificación del cliente
- **`nombreCompleto`**: Nombre y apellido del cliente
- **`totalPedidos`**: Cantidad limitada de pedidos realizados
- **`ultimoPedido`**: Fecha del último pedido (generalmente antigua)

**`estadisticasUsuarios`**: Métricas generales de la base de clientes
- **`totalClientes`**: Número total de clientes registrados en el sistema
- **`clientesActivos`**: Clientes que realizaron al menos un pedido en el período
- **`clientesInactivos`**: Clientes sin actividad en el período
- **`promedioGastoPorCliente`**: Gasto promedio por cliente activo

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
    "ventasPorHora": [
      {"hora": 12, "total": 5220000, "cantidad": 145},
      {"hora": 13, "total": 6048000, "cantidad": 168},
      {"hora": 19, "total": 6912000, "cantidad": 192}
    ],
    "ventasPorDiaSemana": [
      {"diaSemana": "Viernes", "total": 8820000, "cantidad": 245},
      {"diaSemana": "Sábado", "total": 9648000, "cantidad": 268},
      {"diaSemana": "Domingo", "total": 7128000, "cantidad": 198}
    ],
    "ventasPorMes": [
      {"mes": "Enero", "total": 12450000, "cantidad": 456},
      {"mes": "Febrero", "total": 11280000, "cantidad": 412},
      {"mes": "Marzo", "total": 13670000, "cantidad": 498}
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
- `periodo` (opcional): `hoy` | `ultima_semana` | `ultimo_mes` | `ultimos_3_meses` | `ultimos_6_meses` | `ultimo_año` | `historico`
- `mes` (opcional): Mes específico (1-12) para filtrar por mes y año
- `año` (opcional): Año específico (ej: 2024) para filtrar por mes y año
- `fecha_inicio` (opcional): Fecha de inicio para rango personalizado (YYYY-MM-DD)
- `fecha_fin` (opcional): Fecha de fin para rango personalizado (YYYY-MM-DD)
- `hora_inicio` (opcional): Hora de inicio para filtro horario (HH:MM:SS)
- `hora_fin` (opcional): Hora de fin para filtro horario (HH:MM:SS)

**Respuesta esperada:**
```json
{
  "code": 200,
  "message": "Análisis de rentabilidad obtenido exitosamente",
  "data": {
    "productosRentables": [
      {
        "productoId": 1,                    // ID único del producto
        "nombreProducto": "Bandeja Paisa",  // Nombre del producto
        "precioVenta": 36000,              // Precio de venta al público (COP)
        "cantidadVendida": 145,            // Unidades vendidas en el período
        "ingresoTotal": 5220000,           // Ingresos brutos totales (COP)
        "margenGanancia": 65.5,            // Porcentaje de margen de ganancia (%)
        "gananciaTotal": 3419000           // Ganancia neta total (COP)
      },
      {"productoId": 2, "nombreProducto": "Sancocho de Gallina", "precioVenta": 35000, "cantidadVendida": 128, "ingresoTotal": 4480000, "margenGanancia": 58.2, "gananciaTotal": 2607360}
    ],
    "productosMenosRentables": [
      {
        "productoId": 4,                         // ID del producto menos rentable
        "nombreProducto": "Cazuela de Mariscos", // Producto con bajo margen
        "precioVenta": 50000,                   // Precio alto pero pocos compradores
        "cantidadVendida": 12,                  // Pocas unidades vendidas
        "ingresoTotal": 600000,                 // Ingresos limitados
        "margenGanancia": 25.8,                 // Margen de ganancia bajo (%)
        "gananciaTotal": 154800                 // Ganancia neta baja
      },
      {"productoId": 5, "nombreProducto": "Churrasco Premium", "precioVenta": 60000, "cantidadVendida": 18, "ingresoTotal": 1080000, "margenGanancia": 32.1, "gananciaTotal": 346680}
    ],
    "estadisticasRentabilidad": {
      "margenPromedioGeneral": 52.3,           // Margen promedio de todos los productos (%)
      "productoMasRentable": "Bandeja Paisa",  // Producto con mayor margen de ganancia
      "productoMenosRentable": "Cazuela de Mariscos", // Producto con menor margen
      "totalGanancias": 23887500,              // Suma de todas las ganancias netas (COP)
      "totalIngresos": 45750000                // Suma de todos los ingresos brutos (COP)
    }
  }
}
```

**📋 Descripción de campos:**

**`productosRentables` / `productosMenosRentables`**: Productos ordenados por rentabilidad
- **`productoId`**: Identificador único del producto
- **`nombreProducto`**: Nombre comercial del producto
- **`precioVenta`**: Precio de venta al cliente final en pesos colombianos
- **`cantidadVendida`**: Número de unidades vendidas en el período
- **`ingresoTotal`**: Ingresos brutos totales (precioVenta × cantidadVendida)
- **`margenGanancia`**: Porcentaje de margen de ganancia sobre el precio de venta
- **`gananciaTotal`**: Ganancia neta total después de costos (ingresoTotal × margenGanancia/100)

**`estadisticasRentabilidad`**: Métricas generales de rentabilidad
- **`margenPromedioGeneral`**: Margen de ganancia promedio de todos los productos
- **`productoMasRentable`**: Nombre del producto con mayor margen de ganancia
- **`productoMenosRentable`**: Nombre del producto con menor margen de ganancia
- **`totalGanancias`**: Suma de todas las ganancias netas del período
- **`totalIngresos`**: Suma de todos los ingresos brutos del período

> **💡 Nota**: El margen de ganancia se calcula como: `(precioVenta - costoProducto) / precioVenta × 100`

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
- `periodo` (opcional): `hoy` | `ultima_semana` | `ultimo_mes` | `ultimos_3_meses` | `ultimos_6_meses` | `ultimo_año` | `historico`
- `mes` (opcional): Mes específico (1-12) para filtrar por mes y año
- `año` (opcional): Año específico (ej: 2024) para filtrar por mes y año
- `fecha_inicio` (opcional): Fecha de inicio para rango personalizado (YYYY-MM-DD)
- `fecha_fin` (opcional): Fecha de fin para rango personalizado (YYYY-MM-DD)
- `hora_inicio` (opcional): Hora de inicio para filtro horario (HH:MM:SS)
- `hora_fin` (opcional): Hora de fin para filtro horario (HH:MM:SS)

**Respuesta esperada:**
```json
{
  "code": 200,
  "message": "Segmentación de clientes obtenida exitosamente",
  "data": {
    "clientesVIP": [
      {"documentoCliente": 1015466494, "nombreCompleto": "María González", "totalPedidos": 28, "totalGastado": 1260000, "promedioGasto": 45000.00, "ultimoPedido": "2024-01-15T14:30:00Z", "diasSinPedir": 2, "segmento": "VIP", "valorVida": 2520000},
      {"documentoCliente": 1023456789, "nombreCompleto": "Carlos Rodríguez", "totalPedidos": 24, "totalGastado": 1080000, "promedioGasto": 45000.00, "ultimoPedido": "2024-01-14T19:45:00Z", "diasSinPedir": 3, "segmento": "VIP", "valorVida": 2160000}
    ],
    "clientesRegulares": [
      {"documentoCliente": 1034567890, "nombreCompleto": "Ana Martínez", "totalPedidos": 21, "totalGastado": 945000, "promedioGasto": 45000.00, "ultimoPedido": "2024-01-13T12:15:00Z", "diasSinPedir": 4, "segmento": "Regular", "valorVida": 1890000},
      {"documentoCliente": 1045678901, "nombreCompleto": "Luis Hernández", "totalPedidos": 16, "totalGastado": 720000, "promedioGasto": 45000.00, "ultimoPedido": "2024-01-12T18:20:00Z", "diasSinPedir": 5, "segmento": "Regular", "valorVida": 1440000}
    ],
    "clientesOcasionales": [
      {"documentoCliente": 1056789012, "nombreCompleto": "Pedro Sánchez", "totalPedidos": 8, "totalGastado": 360000, "promedioGasto": 45000.00, "ultimoPedido": "2024-01-10T15:30:00Z", "diasSinPedir": 7, "segmento": "Ocasional", "valorVida": 720000}
    ],
    "clientesNuevos": [
      {"documentoCliente": 1067890123, "nombreCompleto": "Laura López", "totalPedidos": 2, "totalGastado": 90000, "promedioGasto": 45000.00, "ultimoPedido": "2024-01-16T11:45:00Z", "diasSinPedir": 1, "segmento": "Nuevo", "valorVida": 180000}
    ],
    "estadisticasSegmentacion": {
      "totalClientesVIP": 15,
      "totalClientesRegulares": 85,
      "totalClientesOcasionales": 120,
      "totalClientesNuevos": 100,
      "promedioGastoVIP": 1125000.00,
      "promedioGastoRegular": 485000.00,
      "porcentajeVIP": 4.69
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
- `periodo` (opcional): `hoy` | `ultima_semana` | `ultimo_mes` | `ultimos_3_meses` | `ultimos_6_meses` | `ultimo_año` | `historico`
- `mes` (opcional): Mes específico (1-12) para filtrar por mes y año
- `año` (opcional): Año específico (ej: 2024) para filtrar por mes y año
- `fecha_inicio` (opcional): Fecha de inicio para rango personalizado (YYYY-MM-DD)
- `fecha_fin` (opcional): Fecha de fin para rango personalizado (YYYY-MM-DD)
- `hora_inicio` (opcional): Hora de inicio para filtro horario (HH:MM:SS)
- `hora_fin` (opcional): Hora de fin para filtro horario (HH:MM:SS)

**Respuesta esperada:**
```json
{
  "code": 200,
  "message": "Análisis de eficiencia obtenido exitosamente",
  "data": {
    "tiemposEntrega": [
      {"pedidoId": 1001, "cliente": "María González", "fechaPedido": "2024-01-15", "horaPedido": "14:30:00", "tiempoPreparacion": 35, "estadoPedido": "ENTREGADO", "trabajadorAsignado": "Juan Pérez"},
      {"pedidoId": 1002, "cliente": "Carlos Rodríguez", "fechaPedido": "2024-01-14", "horaPedido": "19:45:00", "tiempoPreparacion": 42, "estadoPedido": "ENTREGADO", "trabajadorAsignado": "Ana García"},
      {"pedidoId": 1003, "cliente": "Ana Martínez", "fechaPedido": "2024-01-13", "horaPedido": "12:15:00", "tiempoPreparacion": 28, "estadoPedido": "ENTREGADO", "trabajadorAsignado": "Juan Pérez"}
    ],
    "rendimientoTrabajadores": [
      {"documentoTrabajador": 1098765432, "nombreTrabajador": "Juan Pérez", "pedidosAtendidos": 156, "tiempoPromedioAtencion": 32.5, "eficienciaScore": 9.2, "horasTrabajadas": 160.0},
      {"documentoTrabajador": 1087654321, "nombreTrabajador": "Ana García", "pedidosAtendidos": 142, "tiempoPromedioAtencion": 35.8, "eficienciaScore": 8.8, "horasTrabajadas": 155.0}
    ],
    "analisisPorHora": [
      {"hora": "12:00", "pedidosRecibidos": 25, "tiempoPromedioPrep": 30.5, "capacidadUtilizada": 89.3, "nivelEficiencia": "Alto"},
      {"hora": "13:00", "pedidosRecibidos": 32, "tiempoPromedioPrep": 33.2, "capacidadUtilizada": 86.5, "nivelEficiencia": "Alto"},
      {"hora": "19:00", "pedidosRecibidos": 28, "tiempoPromedioPrep": 35.1, "capacidadUtilizada": 87.5, "nivelEficiencia": "Alto"}
    ],
    "estadisticasEficiencia": {
      "tiempoPromedioGeneral": 34.2,
      "horaMasEficiente": "12:00",
      "horaMenosEficiente": "19:00",
      "trabajadorMasEficiente": "Juan Pérez",
      "capacidadPromedioUso": 87.8,
      "pedidosPendientes": 12
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
      {"fecha": "2024-01-15", "totalReservas": 35, "reservasCompletadas": 28, "totalPersonas": 112, "porcentajeCompletado": 80.0},
      {"fecha": "2024-01-14", "totalReservas": 40, "reservasCompletadas": 32, "totalPersonas": 128, "porcentajeCompletado": 80.0},
      {"fecha": "2024-01-13", "totalReservas": 30, "reservasCompletadas": 25, "totalPersonas": 100, "porcentajeCompletado": 83.3}
    ],
    "reservasPorHora": [
      {"hora": "12:00", "totalReservas": 56, "reservasCompletadas": 45, "totalPersonas": 180, "porcentajeCompletado": 80.4},
      {"hora": "13:00", "totalReservas": 65, "reservasCompletadas": 52, "totalPersonas": 208, "porcentajeCompletado": 80.0},
      {"hora": "19:00", "totalReservas": 85, "reservasCompletadas": 68, "totalPersonas": 272, "porcentajeCompletado": 80.0}
    ],
    "reservasPorDiaSemana": [
      {"diaSemana": "Viernes", "totalReservas": 72, "reservasCompletadas": 58, "totalPersonas": 232, "porcentajeCompletado": 80.6},
      {"diaSemana": "Sábado", "totalReservas": 90, "reservasCompletadas": 72, "totalPersonas": 288, "porcentajeCompletado": 80.0},
      {"diaSemana": "Domingo", "totalReservas": 56, "reservasCompletadas": 45, "totalPersonas": 180, "porcentajeCompletado": 80.4}
    ],
    "estadisticasReservas": {
      "totalReservasCompletadas": 330,
      "diaMasReservas": "Sábado",
      "horaMasReservas": "19:00",
      "promedioPersonasPorReserva": 4.0,
      "tasaCompletamiento": 80.5
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
      {"fecha": "2024-01-15", "totalPedidos": 50, "pedidosTerminados": 45, "ingresoTotal": 1620000, "tasaCompletamiento": 90.0},
      {"fecha": "2024-01-14", "totalPedidos": 58, "pedidosTerminados": 52, "ingresoTotal": 1872000, "tasaCompletamiento": 89.7},
      {"fecha": "2024-01-13", "totalPedidos": 42, "pedidosTerminados": 38, "ingresoTotal": 1368000, "tasaCompletamiento": 90.5}
    ],
    "pedidosPorHora": [
      {"hora": "12:00", "totalPedidos": 160, "pedidosTerminados": 145, "ingresoTotal": 5220000, "tasaCompletamiento": 90.6},
      {"hora": "13:00", "totalPedidos": 186, "pedidosTerminados": 168, "ingresoTotal": 6048000, "tasaCompletamiento": 90.3},
      {"hora": "19:00", "totalPedidos": 213, "pedidosTerminados": 192, "ingresoTotal": 6912000, "tasaCompletamiento": 90.1}
    ],
    "pedidosPorDiaSemana": [
      {"diaSemana": "Viernes", "totalPedidos": 272, "pedidosTerminados": 245, "ingresoTotal": 8820000, "tasaCompletamiento": 90.1},
      {"diaSemana": "Sábado", "totalPedidos": 298, "pedidosTerminados": 268, "ingresoTotal": 9648000, "tasaCompletamiento": 89.9},
      {"diaSemana": "Domingo", "totalPedidos": 220, "pedidosTerminados": 198, "ingresoTotal": 7128000, "tasaCompletamiento": 90.0}
    ],
    "estadisticasPedidos": {
      "totalPedidosTerminados": 1250,
      "diaMasPedidos": "Sábado",
      "horaMasPedidos": "19:00",
      "ingresoPromedioHora": 191600.0,
      "tasaCompletamientoGeneral": 90.1
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
  totalIngresos: number;
  totalUsuarios: number;
  promedioVentaPedido: number;
  pedidosHoy: number;
  ingresosHoy: number;
}

// Sales Types
export interface VentaPorMetodo {
  metodoPago: string;
  total: number;
  cantidad: number;
}

export interface VentaPorFecha {
  fecha: string;
  total: number;
  cantidad: number;
}

export interface EstadisticasVentas {
  ventaPromedioDiaria: number;
  pedidoPromedioDiario: number;
  ticketPromedio: number;
}

export interface SalesData {
  ventasPorMetodoPago: VentaPorMetodo[];
  tendenciaVentas: VentaPorFecha[];
  estadisticasGenerales: EstadisticasVentas;
}

// Products Types
export interface ProductoVendido {
  productoId: number;
  nombreProducto: string;
  cantidadVendida: number;
  ingresoTotal: number;
  precio: number;
  imagen: string;
}

export interface EstadisticasProductos {
  totalProductosActivos: number;
  productoConMasVentas: string;
  productoConMenosVentas: string;
}

export interface ProductsData {
  productosMasVendidos: ProductoVendido[];
  productosMenosVendidos: ProductoVendido[];
  estadisticasProductos: EstadisticasProductos;
}

// Users Types
export interface UsuarioFrecuente {
  documentoCliente: number;
  nombreCompleto: string;
  totalPedidos: number;
  totalGastado: number;
  ultimoPedido: string;
}

export interface UsuarioInactivo {
  documentoCliente: number;
  nombreCompleto: string;
  totalPedidos: number;
  ultimoPedido: string;
}

export interface EstadisticasUsuarios {
  totalClientes: number;
  clientesActivos: number;
  clientesInactivos: number;
  promedioGastoPorCliente: number;
}

export interface UsersData {
  usuariosFrecuentes: UsuarioFrecuente[];
  usuariosInactivos: UsuarioInactivo[];
  estadisticasUsuarios: EstadisticasUsuarios;
}

// Time Analysis Types
export interface VentaPorHora {
  hora: number;
  total: number;
  cantidad: number;
}

export interface VentaPorDiaSemana {
  diaSemana: string;
  total: number;
  cantidad: number;
}

export interface VentaPorMes {
  mes: string;
  total: number;
  cantidad: number;
}

export interface TimeAnalysisData {
  ventasPorHora: VentaPorHora[];
  ventasPorDiaSemana: VentaPorDiaSemana[];
  ventasPorMes: VentaPorMes[];
}

// Rentabilidad Types
export interface ProductoRentabilidad {
  productoId: number;
  nombreProducto: string;
  precioVenta: number;
  cantidadVendida: number;
  ingresoTotal: number;
  margenGanancia: number;
  gananciaTotal: number;
}

export interface EstadisticasRentabilidad {
  margenPromedioGeneral: number;
  productoMasRentable: string;
  productoMenosRentable: string;
  totalGanancias: number;
  totalIngresos: number;
}

export interface RentabilidadData {
  productosRentables: ProductoRentabilidad[];
  productosMenosRentables: ProductoRentabilidad[];
  estadisticasRentabilidad: EstadisticasRentabilidad;
}

// Segmentación Types
export interface ClienteSegmento {
  documentoCliente: number;
  nombreCompleto: string;
  totalPedidos: number;
  totalGastado: number;
  promedioGasto: number;
  ultimoPedido: string;
  diasSinPedir: number;
  segmento: string;
  valorVida: number;
}

export interface EstadisticasSegmentacion {
  totalClientesVIP: number;
  totalClientesRegulares: number;
  totalClientesOcasionales: number;
  totalClientesNuevos: number;
  promedioGastoVIP: number;
  promedioGastoRegular: number;
  porcentajeVIP: number;
}

export interface SegmentacionData {
  clientesVIP: ClienteSegmento[];
  clientesRegulares: ClienteSegmento[];
  clientesOcasionales: ClienteSegmento[];
  clientesNuevos: ClienteSegmento[];
  estadisticasSegmentacion: EstadisticasSegmentacion;
}

// Eficiencia Types
export interface TiempoEntrega {
  pedidoId: number;
  cliente: string;
  fechaPedido: string;
  horaPedido: string;
  tiempoPreparacion: number;
  estadoPedido: string;
  trabajadorAsignado: string;
}

export interface RendimientoTrabajador {
  documentoTrabajador: number;
  nombreTrabajador: string;
  pedidosAtendidos: number;
  tiempoPromedioAtencion: number;
  eficienciaScore: number;
  horasTrabajadas: number;
}

export interface EficienciaPorHora {
  hora: string;
  pedidosRecibidos: number;
  tiempoPromedioPrep: number;
  capacidadUtilizada: number;
  nivelEficiencia: string;
}

export interface EstadisticasEficiencia {
  tiempoPromedioGeneral: number;
  horaMasEficiente: string;
  horaMenosEficiente: string;
  trabajadorMasEficiente: string;
  capacidadPromedioUso: number;
  pedidosPendientes: number;
}

export interface EficienciaData {
  tiemposEntrega: TiempoEntrega[];
  rendimientoTrabajadores: RendimientoTrabajador[];
  analisisPorHora: EficienciaPorHora[];
  estadisticasEficiencia: EstadisticasEficiencia;
}

// Reservas Analysis Types
export interface ReservaPorDia {
  fecha: string;
  totalReservas: number;
  reservasCompletadas: number;
  totalPersonas: number;
  porcentajeCompletado: number;
}

export interface ReservaPorHora {
  hora: string;
  totalReservas: number;
  reservasCompletadas: number;
  totalPersonas: number;
  porcentajeCompletado: number;
}

export interface ReservaPorDiaSemana {
  diaSemana: string;
  totalReservas: number;
  reservasCompletadas: number;
  totalPersonas: number;
  porcentajeCompletado: number;
}

export interface EstadisticasReservas {
  totalReservasCompletadas: number;
  diaMasReservas: string;
  horaMasReservas: string;
  promedioPersonasPorReserva: number;
  tasaCompletamiento: number;
}

export interface ReservasAnalisisData {
  reservasPorDia: ReservaPorDia[];
  reservasPorHora: ReservaPorHora[];
  reservasPorDiaSemana: ReservaPorDiaSemana[];
  estadisticasReservas: EstadisticasReservas;
}

// Pedidos Analysis Types
export interface PedidoPorDia {
  fecha: string;
  totalPedidos: number;
  pedidosTerminados: number;
  ingresoTotal: number;
  tasaCompletamiento: number;
}

export interface PedidoPorHora {
  hora: string;
  totalPedidos: number;
  pedidosTerminados: number;
  ingresoTotal: number;
  tasaCompletamiento: number;
}

export interface PedidoPorDiaSemana {
  diaSemana: string;
  totalPedidos: number;
  pedidosTerminados: number;
  ingresoTotal: number;
  tasaCompletamiento: number;
}

export interface EstadisticasPedidos {
  totalPedidosTerminados: number;
  diaMasPedidos: string;
  horaMasPedidos: string;
  ingresoPromedioHora: number;
  tasaCompletamientoGeneral: number;
}

export interface PedidosAnalisisData {
  pedidosPorDia: PedidoPorDia[];
  pedidosPorHora: PedidoPorHora[];
  pedidosPorDiaSemana: PedidoPorDiaSemana[];
  estadisticasPedidos: EstadisticasPedidos;
}

// Service Parameters
export interface TelemetryParams {
  periodo?: TimePeriod;
  limit?: number;
  mes?: number;           // 1-12
  año?: number;           // ej: 2024
  fecha_inicio?: string;  // YYYY-MM-DD
  fecha_fin?: string;     // YYYY-MM-DD
  hora_inicio?: string;   // HH:MM:SS
  hora_fin?: string;      // HH:MM:SS
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
    totalIngresos: 45750000,
    totalUsuarios: 320,
    promedioVentaPedido: 36600.40,
    pedidosHoy: 45,
    ingresosHoy: 1650000
  }
};

export const mockSalesData: ApiResponse<SalesData> = {
  code: 200,
  message: "Análisis de ventas obtenido exitosamente",
  data: {
    ventasPorMetodoPago: [
      { metodoPago: "EFECTIVO", total: 16470000, cantidad: 450 },
      { metodoPago: "TARJETA", total: 19032000, cantidad: 520 },
      { metodoPago: "TRANSFERENCIA", total: 10248000, cantidad: 280 }
    ],
    tendenciaVentas: [
      { fecha: "2024-01-01", total: 1250000, cantidad: 45 },
      { fecha: "2024-01-02", total: 980000, cantidad: 38 },
      { fecha: "2024-01-03", total: 1450000, cantidad: 52 }
    ],
    estadisticasGenerales: {
      ventaPromedioDiaria: 1226666.67,
      pedidoPromedioDiario: 44.33,
      ticketPromedio: 27666.67
    }
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
// Resultados esperados en el componente de telemetría

- TelemetryDashboard: Vista principal con KPIs
- SalesChart: Gráficos de ventas y métodos de pago
- ProductsRanking: Top productos más/menos vendidos
- UsersSegmentation: Análisis de clientes VIP/regulares/nuevos
- TimeAnalysisCharts: Gráficos por hora/día/mes
- EfficiencyMetrics: Métricas de rendimiento y tiempos
- ReservationsAnalysis: Análisis de reservas por tiempo
- OrdersAnalysis: Análisis de pedidos completados
- ProfitabilityChart: Análisis de rentabilidad por producto

// 🆕 Componentes de Filtros Avanzados
- AdvancedPeriodFilter: Selector con 3 tipos de filtros temporales
  - PredefinedPeriodSelector: Filtros predefinidos (hoy, semana, mes, etc.)
  - MonthYearPicker: Selector de mes y año específicos
  - CustomDateRangePicker: Rango de fechas personalizado con horas
- FilterPresets: Botones de filtros rápidos comunes
- DateTimeRangePicker: Componente combinado fecha + hora
- FilterSummary: Resumen visual de filtros aplicados
- PopularProductsPublic: Componente público para productos populares (sin auth)
```

---

## 📝 Notas Importantes

1. **Autenticación**: Todos los endpoints requieren token JWT válido con rol "Administrador" (excepto `/productos-populares`)
2. **Filtros temporales avanzados**: Todos los endpoints soportan 3 tipos de filtros:
   - **Predefinidos**: `periodo` (hoy, ultima_semana, ultimo_mes, etc.)
   - **Mes/Año**: `mes` (1-12) y `año` (ej: 2024)
   - **Rango personalizado**: `fecha_inicio`, `fecha_fin`, `hora_inicio`, `hora_fin`
3. **Prioridad de filtros**: Mes/Año > Rango de fechas > Periodo predefinido
4. **Paginación**: Algunos endpoints soportan el parámetro `limit`
5. **Formato de fechas**: Las fechas se devuelven en formato ISO 8601
6. **Formato de horas**: Las horas se especifican en formato HH:MM:SS (24 horas)
7. **Moneda**: Los valores monetarios están en pesos colombianos (COP)
8. **Porcentajes**: Se devuelven como números decimales (ej: 36.5 = 36.5%)
9. **TotalUsuarios**: Ahora refleja usuarios activos en el período filtrado, no el total global
10. **Endpoint público**: `/productos-populares` no requiere autenticación y soporta todos los filtros

---

## 🚀 ¡Listo para implementar!

Con esta guía tienes todo lo necesario para crear un frontend completo de telemetría que consuma todos los endpoints del sistema. ¡Manos a la obra! 🎯
