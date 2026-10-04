# Actualización de Colección Postman

**Fecha:** 15 de octubre de 2025  
**Versión:** 2.0.0  
**Estado:** ✅ Sincronizada con Swagger  
**Tipo:** 🎯 Profesional y completa con autenticación automática

## Resumen

La colección de Postman ha sido completamente actualizada para estar acorde con la documentación de Swagger (v2.0.0).

### Estadísticas

- **Total de carpetas:** 26
- **Total de requests:** 143
- **Endpoints en Swagger:** 71

## Cambios Principales

### 1. Configuración General

- **baseUrl actualizado:** Ahora es `http://localhost:8080/restaurante/v1` (incluye el prefijo)
- **Nombre actualizado:** "El fogón de María API (Beego v2)"
- **Versión:** 2.0.0
- **Descripción:** Incluye referencia a sincronización con Swagger

### 2. Nuevos Endpoints Agregados

#### Incidencias (5 endpoints)
- `GET /incidencias`
- `GET /incidencias/search?id`
- `POST /incidencias`
- `PUT /incidencias?id`
- `DELETE /incidencias?id`

#### Métodos de Pago (5 endpoints)
- `GET /metodos_pago`
- `GET /metodos_pago/search?id`
- `POST /metodos_pago`
- `PUT /metodos_pago?id`
- `DELETE /metodos_pago?id`

#### Telemetría (10 endpoints)
- `GET /telemetria/dashboard`
- `GET /telemetria/sales`
- `GET /telemetria/products`
- `GET /telemetria/users`
- `GET /telemetria/time-analysis`
- `GET /telemetria/pedidos-analisis`
- `GET /telemetria/reservas-analisis`
- `GET /telemetria/rentabilidad`
- `GET /telemetria/eficiencia`
- `GET /telemetria/segmentacion`

#### Endpoints Auxiliares
- `GET /estados-pedidos` - Consultar estados disponibles de pedidos
- `GET /productos-disponibles` - Productos disponibles (público)
- `GET /productos-populares` - Productos más populares (público)
- `GET /reservas/cliente?cliente_id` - Reservas por cliente
- `GET /reservas/documento?documento` - Reservas por documento
- `GET /restaurante_dia/search?id` - Búsqueda de horarios específicos

### 3. Endpoints Corregidos

#### Cupones
- **Antes:** `POST /cupones/redimir?codigo=CODIGO`
- **Ahora:** `POST /cupones/{codigo}/redimir` (path parameter)

#### Auth Refresh
- **Antes:** Usaba body con refreshToken
- **Ahora:** Usa header `Authorization: Bearer {token}`

### 4. Endpoints Eliminados

Estos endpoints no existen en Swagger y fueron eliminados de Postman:

- `DELETE /pedidos?id` - Swagger solo tiene GET y POST para pedidos
- `DELETE /producto_pedido?pedido_id` - Swagger solo tiene GET, POST y PUT
- `PUT /pedidos?id` - No existe en Swagger

### 5. Duplicados Eliminados

- `GET /ofertas/activas` - Movido a sección "Público", eliminado duplicado de "Ofertas"

## Estructura de la Colección

```
📁 El fogón de María API (Beego v2)
├── 📂 Health (2)
│   ├── GET /healthz
│   └── GET /readyz
├── 📂 Auth (2)
│   ├── POST /login (con auto-captura de token)
│   └── POST /auth/refresh
├── 📂 Público (11) - Sin autenticación
│   ├── POST /clientes (registro)
│   ├── GET /productos (con filtros)
│   ├── GET /productos/search?id
│   ├── GET /reservas
│   ├── GET /reservas/search?id
│   ├── GET /reservas/parameter
│   ├── GET /reservas/cliente
│   ├── GET /reservas/documento
│   ├── POST /reservas
│   ├── GET /cambios_horario/actual
│   └── GET /ofertas/activas
├── 📂 Clientes (4)
├── 📂 Restaurantes (5)
├── 📂 Trabajadores (5)
├── 📂 Horario Trabajador (4)
├── 📂 Productos (escritura) (3)
├── 📂 Categorías y Subcategorías (10)
├── 📂 Precio Producto Hist (2)
├── 📂 Pedidos (6)
├── 📂 Producto Pedido (3)
├── 📂 Domicilios (6)
├── 📂 Pagos (5)
├── 📂 Cambios de horario (gestión) (4)
├── 📂 Reservas (gestión) (2)
├── 📂 Nómina (4)
├── 📂 Nómina Trabajador (6)
├── 📂 Lecturas auxiliares (9)
├── 📂 Incidencias (5) ⭐ NUEVO
├── 📂 Métodos de Pago (5) ⭐ NUEVO
├── 📂 Push Notifications (15)
├── 📂 Cupones (10)
├── 📂 Ofertas (8)
├── 📂 Descuentos (4)
└── 📂 Telemetría (10) ⭐ NUEVO
```

## Uso de Variables

La colección utiliza dos variables:

- **`{{baseUrl}}`** - Base URL de la API (default: `http://localhost:8080/restaurante/v1`)
- **`{{token}}`** - Token JWT (se captura automáticamente al hacer login)

## Autenticación

- La mayoría de los endpoints requieren autenticación Bearer Token
- Los endpoints en la carpeta "Público" no requieren autenticación
- El endpoint de login captura automáticamente el token en la variable `{{token}}`
- El endpoint de refresh token actualiza automáticamente el token

## Notas Importantes

1. **Formato de fechas:** Usar `YYYY-MM-DD` para fechas y `HH:MM:SS` para horas
2. **Query strings:** Beego usa query strings (`?id=1`), no path params (`:id`)
3. **Paginación:** Endpoints con listas soportan `?limit=20&offset=0`
4. **Filtros:** Muchos endpoints GET soportan filtros adicionales (ver ejemplos)

## Pruebas

Para verificar que la colección funciona correctamente:

1. Importar `tools/postman_collection.json` en Postman
2. Configurar la variable `baseUrl` si es necesario
3. Ejecutar el request `POST /login` para obtener el token
4. El token se guardará automáticamente en `{{token}}`
5. Probar cualquier endpoint autenticado

## Próximos Pasos

- [ ] Agregar tests automáticos a más endpoints
- [ ] Crear colecciones de prueba específicas por módulo
- [ ] Agregar ejemplos de respuestas esperadas
- [ ] Documentar casos de error comunes

---

**Mantenimiento:** Esta colección debe actualizarse cada vez que se modifique `docs/swagger.json`


