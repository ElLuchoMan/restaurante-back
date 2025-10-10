# Gestión de Fechas y Zonas Horarias

## Problema Identificado

El sistema presentaba una discrepancia de **5 horas** en las fechas y horas mostradas en el frontend respecto a las registradas en el backend. Por ejemplo, un pedido creado a las 10:12 AM se mostraba con hora 05:15:59.

## Causa Raíz

El problema se originaba por una **doble conversión de zona horaria** en el backend:

1. **Al crear registros**: Se usaba `time.Now().In(loc)` con `loc = "America/Bogota"`, guardando la hora correcta de Bogotá en PostgreSQL
2. **Configuración de conexión**: La cadena de conexión a PostgreSQL incluye `TimeZone=America/Bogota` (ver `database/database.go:107`)
3. **Campos de tipo date/time**: PostgreSQL almacena estos tipos **sin información de zona horaria** (a diferencia de `timestamptz`)
4. **Error al leer**: El código en `GetDetails` y los métodos `MarshalJSON` asumían incorrectamente que los datos venían en UTC y aplicaban conversión a Bogotá usando `.In(loc)`, causando el desajuste

## Solución Implementada

### Backend (Go)

#### 1. Conexión a PostgreSQL
La conexión ya está correctamente configurada con zona horaria de Bogotá:

```go
// database/database.go:107
connStr := fmt.Sprintf("user=%s password=%s host=%s port=%s dbname=%s sslmode=%s TimeZone=America/Bogota",
    dbUser, dbPass, dbHost, dbPort, dbName, sslMode)
```

#### 2. Creación de Registros
Al crear pedidos (y otros registros con fecha/hora), se obtiene la hora actual en zona horaria de Bogotá:

```go
// controllers/pedido/PedidoController.go:163-171
loc, errLoc := loadLocationPedido("America/Bogota")
if errLoc != nil {
    // manejo de error
}
now := time.Now().In(loc)

pedido.FECHA = now
pedido.HORA = now
```

#### 3. Serialización JSON (MarshalJSON)
**ANTES (incorrecto)**:
```go
func (d Pedido) MarshalJSON() ([]byte, error) {
    loc, _ := time.LoadLocation("America/Bogota")
    fechaBogota := d.FECHA.In(loc)  // ❌ Conversión innecesaria
    horaBogota := d.HORA.In(loc)    // ❌ Conversión innecesaria
    return json.Marshal(&struct {
        FECHA string `json:"fechaPedido"`
        HORA  string `json:"horaPedido"`
        Alias
    }{
        FECHA: fechaBogota.Format("02-01-2006"),
        HORA:  horaBogota.Format("15:04:05"),
        Alias: (Alias)(d),
    })
}
```

**DESPUÉS (correcto)**:
```go
func (d Pedido) MarshalJSON() ([]byte, error) {
    // Los datos ya están en zona horaria de Bogotá
    // Solo formateamos directamente sin conversión adicional
    return json.Marshal(&struct {
        FECHA string `json:"fechaPedido"`
        HORA  string `json:"horaPedido"`
        Alias
    }{
        FECHA: d.FECHA.Format("02-01-2006"),      // ✅ Sin conversión
        HORA:  d.HORA.Format("15:04:05"),         // ✅ Sin conversión
        Alias: (Alias)(d),
    })
}
```

#### 4. Endpoint GetDetails
**ANTES (incorrecto)**:
```go
// Asumía que los datos venían en UTC y los convertía a Bogotá
timestampUTC := time.Date(..., time.UTC)
timestampBogota := timestampUTC.In(loc)  // ❌ Conversión incorrecta
```

**DESPUÉS (correcto)**:
```go
// Los datos en BD ya están en zona horaria de Bogotá
// No es necesario hacer conversión, solo formatear si es necesario
```

### Modelos Actualizados

Los siguientes modelos fueron corregidos para eliminar conversiones innecesarias:

- ✅ `Pedido` - fechas y horas de pedidos
- ✅ `Pago` - fechas y horas de pagos
- ✅ `Reserva` - fechas y horas de reservas
- ✅ `CambiosHorario` - fechas de cambios de horario
- ✅ `Nomina` - fechas de nóminas
- ✅ `Incidencia` - fechas de incidencias
- ✅ `Trabajador` - fechas de nacimiento, ingreso y retiro

### Frontend (Angular)

El frontend ya estaba correctamente configurado con el pipe `FormatDatePipe`:

```typescript
// src/app/shared/pipes/format-date.pipe.ts

@Pipe({ name: 'formatDate' })
export class FormatDatePipe implements PipeTransform {
  transform(value: Date | string | null, mode: Mode = 'date'): string {
    const TZ = 'America/Bogota';  // ✅ Zona horaria correcta

    // Maneja formatos:
    // - DD-MM-YYYY (formato español)
    // - HH:mm:ss (formato de hora)
    // - Aplica timezone cuando es necesario
  }
}
```

## Estrategia de Manejo de Fechas

### Principios

1. **Una sola fuente de verdad**: PostgreSQL con `TimeZone=America/Bogota`
2. **Sin conversiones múltiples**: Los datos se guardan y leen en la misma zona horaria
3. **Formateo simple**: Solo se formatea al serializar, sin conversiones de timezone
4. **Tipos de datos**:
   - `date`: Para fechas sin hora
   - `time`: Para horas sin fecha
   - `timestamptz`: Para fecha-hora con timezone (usado en campos como `UPDATED_AT`)

### Flujo de Datos

```
1. Usuario crea pedido (10:12 AM Bogotá)
   ↓
2. Backend: time.Now().In(loc) → 10:12 AM Bogotá
   ↓
3. PostgreSQL guarda (TimeZone=America/Bogota) → 10:12 AM
   ↓
4. PostgreSQL lee → 10:12 AM
   ↓
5. Backend: MarshalJSON (sin conversión) → "10:12:00"
   ↓
6. Frontend: FormatDatePipe → "10:12:00"
   ↓
7. Usuario ve: 10:12 AM ✅
```

## Consideraciones Importantes

### Campos de Tipo `timestamptz`

Los campos que usan `timestamptz` (como `UPDATED_AT`) mantienen información de zona horaria. Go los maneja automáticamente:

```go
UPDATED_AT: d.UPDATED_AT.Format("02-01-2006 15:04:05")  // ✅ Funciona correctamente
```

### Despliegue en Diferentes Servidores

El sistema ahora es **independiente de la zona horaria del servidor** donde esté desplegado:

- La conexión PostgreSQL siempre usa `TimeZone=America/Bogota`
- Todos los `time.Now()` se convierten explícitamente a `America/Bogota`
- El frontend aplica `America/Bogota` al formatear

### Migración y Datos Existentes

Los datos existentes en la base de datos **no necesitan migración**. Ya están almacenados correctamente con la zona horaria de Bogotá.

## Testing

### Verificación Manual

1. Crear un pedido desde el frontend
2. Verificar que la hora mostrada coincida con la hora actual de Bogotá
3. Refrescar la página y verificar que la hora no cambie

### Tests Unitarios

Los tests de modelos ya contemplan la zona horaria correcta:

```go
// models/pedido_test.go
func TestPedidoMarshalJSON(t *testing.T) {
    loc, _ := time.LoadLocation("America/Bogota")
    fecha := time.Date(2024, time.May, 1, 0, 0, 0, 0, loc)
    p := Pedido{FECHA: fecha}
    // ... verificación de formato
}
```

## Mantenimiento Futuro

### Al Crear Nuevos Modelos con Fechas

1. **Definir el campo** con el tipo apropiado:
   ```go
   FECHA time.Time `orm:"column(fecha);type(date)" json:"fecha"`
   ```

2. **Implementar MarshalJSON** sin conversión de timezone:
   ```go
   func (m MiModelo) MarshalJSON() ([]byte, error) {
       return json.Marshal(&struct {
           FECHA string `json:"fecha"`
           Alias
       }{
           FECHA: m.FECHA.Format("02-01-2006"),  // ✅ Sin .In(loc)
           Alias: (Alias)(m),
       })
   }
   ```

3. **Al crear registros**, usar la zona horaria explícitamente:
   ```go
   loc, _ := time.LoadLocation("America/Bogota")
   now := time.Now().In(loc)
   modelo.FECHA = now
   ```

### Al Consultar/Filtrar por Fechas

No hacer conversiones en las consultas. Los datos ya están en Bogotá:

```go
// ✅ Correcto
query := "SELECT * FROM pedido WHERE fecha = ?"
```

## Referencias

- Configuración de conexión: `database/database.go:107`
- Modelos corregidos: `models/Pedido.go`, `models/Pago.go`, etc.
- Frontend pipe: `src/app/shared/pipes/format-date.pipe.ts`
- Controlador de pedidos: `controllers/pedido/PedidoController.go`

---

**Última actualización**: Octubre 2025
