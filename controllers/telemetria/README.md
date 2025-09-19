# Sistema de Telemetría - El Fogón de María

## Descripción

Sistema de telemetría implementado para la aplicación "El Fogón de María" que proporciona análisis y métricas basadas únicamente en los datos existentes en la base de datos, sin crear nuevas tablas ni modificar esquemas existentes.

## Características

### ✅ Restricciones Cumplidas
- **NO** se crearon nuevas tablas ni migraciones
- **NO** se modificaron esquemas existentes
- **SOLO** se usan datos que ya existen en la BD
- Se trabaja únicamente con modelos actuales

### 🔐 Seguridad
- Solo usuarios con rol "Administrador" pueden acceder
- Utiliza el sistema de autenticación JWT existente
- Validación de roles en cada endpoint
- Rate limiting heredado del sistema actual

### 📊 Endpoints Disponibles

#### 1. Dashboard General
```
GET /restaurante/v1/telemetria/dashboard
```
**Funcionalidades:**
- Contar pedidos totales de tabla existente
- Sumar ingresos de pedidos existentes
- Contar usuarios registrados
- Calcular valor promedio de pedidos
- Pedidos e ingresos del día actual

#### 2. Análisis de Ventas
```
GET /restaurante/v1/telemetria/sales
```
**Funcionalidades:**
- Agrupar pedidos por método de pago
- Calcular totales por método de pago
- Tendencias basadas en fechas de pedidos existentes (últimos 30 días)
- Estadísticas generales de ventas

#### 3. Productos Estrella
```
GET /restaurante/v1/telemetria/products?limit=10
```
**Funcionalidades:**
- Contar productos más pedidos en tabla de pedidos/items
- Agrupar por producto_id y contar frecuencia
- Top N productos más vendidos
- Productos menos vendidos
- Estadísticas generales de productos

#### 4. Análisis de Usuarios
```
GET /restaurante/v1/telemetria/users?limit=10
```
**Funcionalidades:**
- Usuarios con más pedidos (frecuentes)
- Usuarios con menos pedidos (inactivos)
- Valor total gastado por usuario
- Fecha del último pedido por usuario
- Estadísticas generales de usuarios

#### 5. Análisis Temporal
```
GET /restaurante/v1/telemetria/time-analysis
```
**Funcionalidades:**
- Agrupar pedidos por hora del día (usando created_at)
- Agrupar pedidos por día de la semana
- Ventas por mes usando fechas existentes (últimos 12 meses)

### 🚀 Métricas Avanzadas

#### 6. Análisis de Rentabilidad
```
GET /restaurante/v1/telemetria/rentabilidad?limit=10&periodo=ultimo_mes
```
**Funcionalidades:**
- Productos más y menos rentables
- Margen de ganancia por producto (simulado: 60-70%)
- Estadísticas de rentabilidad general
- Cálculo de ganancias totales
- Identificación de productos más/menos rentables

#### 7. Segmentación de Clientes
```
GET /restaurante/v1/telemetria/segmentacion?limit=10&periodo=ultimo_mes
```
**Funcionalidades:**
- **Clientes VIP**: >5 pedidos y >$50,000 gastados
- **Clientes Regulares**: 2-5 pedidos
- **Clientes Ocasionales**: 1 pedido
- **Clientes Nuevos**: Sin pedidos en el período
- Valor de vida del cliente (CLV estimado)
- Estadísticas por segmento y porcentajes

#### 8. Análisis de Eficiencia Operacional
```
GET /restaurante/v1/telemetria/eficiencia?limit=10&periodo=ultimo_mes
```
**Funcionalidades:**
- Tiempos de entrega por pedido (simulados: 30-90 min)
- Rendimiento de trabajadores por pedidos atendidos
- Análisis de eficiencia por hora del día
- Capacidad utilizada del restaurante
- Score de eficiencia de trabajadores (1-10)
- Identificación de horas más/menos eficientes

### 🕒 Filtros Temporales

Todos los endpoints soportan el parámetro `periodo` para filtrar datos por tiempo:

```
?periodo=<valor>
```

**Valores disponibles:**
- `hoy`: Solo datos del día actual
- `ultima_semana`: Últimos 7 días
- `ultimo_mes`: Últimos 30 días (por defecto)
- `ultimos_3_meses`: Últimos 90 días
- `ultimos_6_meses`: Últimos 180 días
- `ultimo_año`: Últimos 365 días
- `historico`: Todos los datos disponibles

**Ejemplos:**
```bash
# Dashboard de hoy
GET /restaurante/v1/telemetria/dashboard?periodo=hoy

# Rentabilidad de los últimos 6 meses
GET /restaurante/v1/telemetria/rentabilidad?periodo=ultimos_6_meses&limit=15

# Segmentación histórica
GET /restaurante/v1/telemetria/segmentacion?periodo=historico&limit=20
```

## Estructura de Respuesta

Todos los endpoints siguen la estructura estándar de la API:

```json
{
  "code": 200,
  "message": "Operación exitosa",
  "data": {
    // Datos específicos del endpoint
  }
}
```

## Casos de Uso

### 1. Frontend Home
```javascript
// Obtener productos más vendidos para mostrar "Platos Estrella"
fetch('/restaurante/v1/telemetria/products?limit=5', {
  headers: { 'Authorization': 'Bearer ' + token }
})
```

### 2. Dashboard Administrativo
```javascript
// Obtener métricas generales
fetch('/restaurante/v1/telemetria/dashboard', {
  headers: { 'Authorization': 'Bearer ' + token }
})
```

### 3. Reportes de Ventas
```javascript
// Obtener análisis de ventas
fetch('/restaurante/v1/telemetria/sales', {
  headers: { 'Authorization': 'Bearer ' + token }
})
```

## Consideraciones Técnicas

### Performance
- Consultas SQL optimizadas con índices existentes
- Uso de agregaciones eficientes
- Límites configurables en endpoints que retornan listas
- Consultas con ventanas de tiempo limitadas (30 días, 12 meses)

### Datos Utilizados
- **Tabla `pedido`**: Fechas, horas, estados, clientes
- **Tabla `pago`**: Montos, métodos de pago, estados
- **Tabla `detalle_pedido`**: Productos, cantidades, precios
- **Tabla `producto`**: Nombres, precios actuales
- **Tabla `cliente`**: Información de usuarios
- **Tabla `metodo_pago`**: Tipos de métodos de pago

### Limitaciones
- Solo datos existentes en las tablas actuales
- No se crean logs adicionales
- Análisis limitado a los campos disponibles
- Dependiente de la calidad de datos existentes

## Autenticación

### Requerimientos
- Token JWT válido
- Rol de "Administrador" en el token
- Header: `Authorization: Bearer <token>`

### Respuestas de Error
- **401**: Token no proporcionado o inválido
- **403**: Usuario sin permisos de administrador
- **500**: Error interno del servidor

## Documentación Swagger

Los endpoints están completamente documentados en Swagger UI:
- Acceder a `/swagger/` en modo desarrollo
- Buscar la sección "telemetria"
- Probar endpoints directamente desde la interfaz

## Pruebas

### Ejecutar Pruebas
```bash
go test ./controllers/telemetria/...
```

### Cobertura
```bash
go test -cover ./controllers/telemetria/...
```

## Ejemplos de Respuesta

### Dashboard
```json
{
  "code": 200,
  "message": "Dashboard obtenido exitosamente por Juan Pérez",
  "data": {
    "totalPedidos": 1250,
    "totalIngresos": 45000000,
    "totalUsuarios": 320,
    "promedioVentaPedido": 36000,
    "pedidosHoy": 15,
    "ingresosHoy": 540000
  }
}
```

### Productos Más Vendidos
```json
{
  "code": 200,
  "message": "Análisis de productos obtenido exitosamente por Juan Pérez",
  "data": {
    "productosMasVendidos": [
      {
        "productoId": 1,
        "nombreProducto": "Bandeja Paisa",
        "cantidadVendida": 450,
        "ingresoTotal": 18000000,
        "precio": 40000
      }
    ],
    "estadisticasProductos": {
      "totalProductosActivos": 85,
      "productoConMasVentas": "Bandeja Paisa",
      "productoConMenosVentas": "Ensalada César"
    }
  }
}
```

## Mantenimiento

### Monitoreo
- Revisar logs de errores en consultas SQL
- Monitorear performance de consultas complejas
- Verificar uso de memoria en agregaciones grandes

### Optimización
- Considerar índices adicionales si las consultas se vuelven lentas
- Implementar caché en memoria para consultas frecuentes
- Paginación en resultados grandes

## Contacto

Para soporte técnico o mejoras, contactar al equipo de desarrollo.
