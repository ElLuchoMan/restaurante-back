# API de Notificaciones Push, Cupones y Ofertas

## Resumen de Implementación

Se ha implementado un sistema completo de notificaciones push, cupones y ofertas para el backend del restaurante con las siguientes características:

### 🗄️ Modelos Implementados

1. **PushDispositivo** - Gestión de dispositivos para notificaciones push
2. **PushEnvio** - Registro de envíos de notificaciones
3. **Cupon** - Sistema de cupones de descuento
4. **CuponRedencion** - Registro de redenciones de cupones
5. **Oferta** - Sistema de ofertas por tiempo/día
6. **OfertaProducto** - Relación entre ofertas y productos
7. **PedidoDescuentoAplicado** - Registro de descuentos aplicados a pedidos

### 🔧 Servicios Implementados

- **CuponService** - Validación y redención de cupones
- **OfertaService** - Gestión de ofertas activas
- **PushService** - Gestión de dispositivos push
- **DescuentoService** - Aplicación de descuentos a pedidos

### 🌐 Endpoints REST (API v1)

Todos los endpoints están bajo el prefijo `/restaurante/v1` y están documentados con Swagger.

#### Push Notifications

- `POST /restaurante/v1/push/dispositivos` - Registrar dispositivo
- `PATCH /restaurante/v1/push/dispositivos/{id}/visto` - Actualizar última vista
- `PATCH /restaurante/v1/push/dispositivos/{id}/estado` - Habilitar/deshabilitar
- `PATCH /restaurante/v1/push/dispositivos/{id}/topics` - Actualizar topics
- `GET /restaurante/v1/push/dispositivos` - Listar dispositivos
- `POST /restaurante/v1/push/envios` - Registrar envío
- `GET /restaurante/v1/push/envios` - Listar envíos
- `POST /restaurante/v1/push/enviar` - **Enviar notificación push** ⭐ **NUEVO**

#### Cupones

- `GET /restaurante/v1/cupones` - Listar cupones
- `POST /restaurante/v1/cupones` - Crear cupón
- `GET /restaurante/v1/cupones/{id}` - Obtener cupón
- `PUT /restaurante/v1/cupones/{id}` - Actualizar cupón
- `POST /restaurante/v1/cupones/validar` - Validar cupón
- `POST /restaurante/v1/cupones/{codigo}/redimir` - Redimir cupón
- `GET /restaurante/v1/cupones/redenciones` - Listar redenciones

#### Ofertas

- `GET /restaurante/v1/ofertas` - Listar ofertas
- `POST /restaurante/v1/ofertas` - Crear oferta
- `GET /restaurante/v1/ofertas/activas` - Obtener ofertas activas (público)
- `GET /restaurante/v1/ofertas/{id}` - Obtener oferta
- `PUT /restaurante/v1/ofertas/{id}` - Actualizar oferta
- `POST /restaurante/v1/ofertas/{id}/productos` - Asociar producto
- `DELETE /restaurante/v1/ofertas/{id}/productos/{producto_id}` - Desasociar producto

#### Descuentos en Pedidos

- `GET /restaurante/v1/pedidos/{pedido_id}/descuentos` - Listar descuentos
- `POST /restaurante/v1/pedidos/{pedido_id}/descuentos` - Aplicar descuento

### 📋 Validaciones de Negocio

#### Cupones
- Porcentajes entre 1-100%
- Fechas de vigencia válidas
- Coherencia entre scope y targets (producto/categoría/cliente)
- Límites de uso global y por cliente
- Monto mínimo de pedido

#### Ofertas
- Porcentajes entre 1-100%
- Fechas de vigencia válidas
- Horarios opcionales (inicio y fin requeridos si se especifica)
- Días de la semana válidos
- Asociación con productos específicos

#### Dispositivos Push
- Coherencia entre plataforma y campos requeridos:
  - WEB: endpoint, p256dh, auth (no fcmToken)
  - ANDROID/IOS: fcmToken (no campos WEB)
- Exactamente uno de cliente o trabajador

### 🔒 Seguridad y Validaciones

- Autenticación JWT requerida para endpoints protegidos
- Validación de entrada con mensajes de error detallados
- Respuestas HTTP apropiadas (400, 404, 409, 422)
- Manejo de conflictos (límites de uso, duplicados)
- Exclusividad de descuentos por pedido

### 📊 Características Técnicas

- **Paginación** con limit/offset en endpoints de listado
- **Filtros** por múltiples criterios
- **Ordenación** configurable
- **Tipos de datos avanzados**:
  - `json.RawMessage` para campos JSONB
  - `pq.StringArray` para arrays de texto
  - Enums tipados con validación
- **Documentación Swagger** completa y actualizada

## Ejemplos de Uso

### Registrar Dispositivo Web Push

```bash
curl -X POST http://localhost:8080/restaurante/v1/push/dispositivos \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "plataforma": "WEB",
    "endpoint": "https://fcm.googleapis.com/fcm/send/...",
    "p256dh": "BNbN...",
    "auth": "k8J...",
    "documentoCliente": 12345678
  }'
```

### Crear Cupón de Descuento

```bash
curl -X POST http://localhost:8080/restaurante/v1/cupones \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "codigo": "DESCUENTO20",
    "scope": "GLOBAL",
    "tipoDescuento": "PORCENTAJE",
    "valorDescuento": 20,
    "fechaInicio": "2024-01-01",
    "fechaFin": "2024-12-31",
    "montoMinimo": 50000
  }'
```

### Validar Cupón

```bash
curl -X POST http://localhost:8080/restaurante/v1/cupones/validar \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "codigo": "DESCUENTO20",
    "clienteId": 12345678,
    "items": [
      {
        "productoId": 1,
        "cantidad": 2,
        "precio": 30000
      }
    ]
  }'
```

### Crear Oferta por Días

```bash
curl -X POST http://localhost:8080/restaurante/v1/ofertas \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "titulo": "Martes de Pizza",
    "tipoDescuento": "PORCENTAJE",
    "valorDescuento": 15,
    "fechaInicio": "2024-01-01",
    "fechaFin": "2024-12-31",
    "diasSemana": ["Martes"],
    "horaInicio": "18:00",
    "horaFin": "22:00",
    "restauranteId": 1
  }'
```

### Obtener Ofertas Activas (Público)

```bash
curl -X GET "http://localhost:8080/restaurante/v1/ofertas/activas?restaurante_id=1&fecha=2024-03-19&hora=20:00"
```

### Aplicar Descuento a Pedido

```bash
curl -X POST http://localhost:8080/restaurante/v1/pedidos/123/descuentos \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "cuponId": 1,
    "montoDescuento": 10000,
    "detalle": {"tipo": "cupon", "codigo": "DESCUENTO20"}
  }'
```

## Documentación Swagger

La documentación completa de la API está disponible en:
- **Swagger UI**: `http://localhost:8080/swagger/`
- **JSON**: `http://localhost:8080/swagger/doc.json`
- **YAML**: `http://localhost:8080/swagger/doc.yaml`

## Tests

Se han implementado tests básicos para las validaciones de negocio principales:
- Validación de cupones (porcentajes, fechas, scope)
- Validación de ofertas (horarios, días de semana)
- Validación de dispositivos push (coherencia de plataforma)

```bash
go test ./services -v
```

## Estructura de Archivos

```
├── models/
│   ├── PushDispositivo.go
│   ├── PushEnvio.go
│   ├── Cupon.go
│   ├── CuponRedencion.go
│   ├── Oferta.go
│   ├── OfertaProducto.go
│   ├── PedidoDescuentoAplicado.go
│   ├── NotificacionRequests.go
│   ├── NotificacionResponses.go
│   └── enums.go (actualizado)
├── services/
│   ├── CuponService.go
│   ├── OfertaService.go
│   ├── PushService.go
│   └── DescuentoService.go
├── controllers/
│   ├── push/PushController.go
│   ├── cupon/CuponController.go
│   ├── oferta/OfertaController.go
│   └── descuento/DescuentoController.go
└── routers/router.go (actualizado)
```

## 📊 Datos de Ejemplo para Frontend (Mocks)

### 📱 Dispositivos Push

#### Dispositivo WEB - Cliente 1015466495
```json
{
  "pushDispositivoId": 1,
  "plataforma": "WEB",
  "endpoint": "https://push.example.com/subscription/cliente_1015466495",
  "p256dh": "test_p256dh_1",
  "auth": "test_auth_1",
  "fcmToken": null,
  "enabled": true,
  "locale": "es-CO",
  "timeZone": "America/Bogota",
  "appVersion": "1.0.0",
  "userAgent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
  "subscribedTopics": ["promos", "novedades"],
  "documentoCliente": 1015466495,
  "documentoTrabajador": null,
  "createdAt": "2025-01-01T10:00:00Z",
  "lastSeenAt": "2025-01-15T14:30:00Z"
}
```

#### Dispositivo ANDROID - Trabajador 1000000000
```json
{
  "pushDispositivoId": 2,
  "plataforma": "ANDROID",
  "endpoint": null,
  "p256dh": null,
  "auth": null,
  "fcmToken": "fcm_token_trabajador_1000000000",
  "enabled": true,
  "locale": "es-CO",
  "timeZone": "America/Bogota",
  "appVersion": "2.1.0",
  "subscribedTopics": ["domicilios"],
  "documentoCliente": null,
  "documentoTrabajador": 1000000000,
  "createdAt": "2025-01-01T08:00:00Z",
  "lastSeenAt": "2025-01-15T16:45:00Z"
}
```

### 🎫 Cupones

#### Cupón BIENVENIDA10 (Global)
```json
{
  "cuponId": 1,
  "codigo": "BIENVENIDA10",
  "scope": "GLOBAL",
  "tipoDescuento": "PORCENTAJE",
  "valorDescuento": 10,
  "fechaInicio": "2025-01-01",
  "fechaFin": "2025-12-31",
  "montoMinimo": null,
  "maxUsos": null,
  "limitePorCliente": null,
  "activo": true,
  "productoId": null,
  "categoriaId": null,
  "clienteId": null
}
```

#### Cupón HAMB5K (Producto específico)
```json
{
  "cuponId": 2,
  "codigo": "HAMB5K",
  "scope": "PRODUCTO",
  "tipoDescuento": "MONTO",
  "valorDescuento": 5000,
  "fechaInicio": "2025-01-01",
  "fechaFin": "2025-12-31",
  "montoMinimo": 10000,
  "maxUsos": null,
  "limitePorCliente": null,
  "activo": true,
  "productoId": 3,
  "categoriaId": null,
  "clienteId": null
}
```

#### Cupón CLIENTEVIP20 (Cliente específico)
```json
{
  "cuponId": 3,
  "codigo": "CLIENTEVIP20",
  "scope": "CLIENTE",
  "tipoDescuento": "PORCENTAJE",
  "valorDescuento": 20,
  "fechaInicio": "2025-01-01",
  "fechaFin": "2025-12-31",
  "montoMinimo": null,
  "maxUsos": null,
  "limitePorCliente": null,
  "activo": true,
  "productoId": null,
  "categoriaId": null,
  "clienteId": 1015466495
}
```

### 🎯 Ofertas

#### Oferta "Martes de Gaseosas"
```json
{
  "ofertaId": 1,
  "titulo": "Martes de Gaseosas",
  "tipoDescuento": "PORCENTAJE",
  "valorDescuento": 30,
  "fechaInicio": "2025-01-01",
  "fechaFin": "2025-12-31",
  "diasSemana": ["Martes"],
  "horaInicio": null,
  "horaFin": null,
  "activo": true,
  "restauranteId": 1,
  "productosIds": [1, 2]
}
```

### 🧾 Redenciones y Descuentos

#### Redención de Cupón
```json
{
  "redencionId": 1,
  "cuponId": 1,
  "codigo": "BIENVENIDA10",
  "clienteId": 1015466495,
  "pedidoId": 1,
  "montoDescuento": 2000,
  "fechaRedencion": "2025-01-15T14:30:00Z"
}
```

#### Descuento Aplicado al Pedido
```json
{
  "pedidoDescuentoId": 1,
  "pedidoId": 1,
  "cuponId": 1,
  "ofertaId": null,
  "montoDescuento": 2000,
  "detalle": {
    "fuente": "CUPON",
    "codigo": "BIENVENIDA10",
    "tipo": "PORCENTAJE",
    "valor": 10
  },
  "createdAt": "2025-01-15T14:30:00Z"
}
```

---

## 🔧 Ejemplos cURL Completos

### 📱 Push Notifications

#### Registrar Dispositivo Web
```bash
curl -X POST http://localhost:8080/restaurante/v1/push/dispositivos \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "plataforma": "WEB",
    "endpoint": "https://push.example.com/subscription/cliente_1015466495",
    "p256dh": "test_p256dh_1",
    "auth": "test_auth_1",
    "locale": "es-CO",
    "timeZone": "America/Bogota",
    "appVersion": "1.0.0",
    "userAgent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
    "subscribedTopics": ["promos", "novedades"],
    "documentoCliente": 1015466495
  }'
```

#### Registrar Dispositivo Android (Trabajador)
```bash
curl -X POST http://localhost:8080/restaurante/v1/push/dispositivos \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "plataforma": "ANDROID",
    "fcmToken": "fcm_token_trabajador_1000000000",
    "locale": "es-CO",
    "timeZone": "America/Bogota",
    "appVersion": "2.1.0",
    "subscribedTopics": ["domicilios"],
    "documentoTrabajador": 1000000000
  }'
```

#### Actualizar Topics de Dispositivo
```bash
curl -X PATCH http://localhost:8080/restaurante/v1/push/dispositivos/1/topics \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "subscribedTopics": ["promos", "novedades", "ofertas"]
  }'
```

#### Registrar Envío Push
```bash
curl -X POST http://localhost:8080/restaurante/v1/push/envios \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "pushDispositivoId": 1,
    "proveedor": "WEB_PUSH",
    "data": {
      "title": "¡Nueva oferta disponible!",
      "body": "Martes de Gaseosas - 30% de descuento",
      "icon": "/icon-192x192.png",
      "badge": "/badge-72x72.png",
      "url": "/ofertas/martes-gaseosas"
    },
    "exito": true,
    "statusCode": 200
  }'
```

#### ⭐ Enviar Notificación Push (NUEVO)

##### Enviar a todos los dispositivos
```bash
curl -X POST http://localhost:8080/restaurante/v1/push/enviar \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "remitente": {
      "tipo": "TRABAJADOR",
      "documentoTrabajador": 1000000000,
      "nombre": "Juan Pérez"
    },
    "destinatarios": {
      "tipo": "TODOS"
    },
    "notificacion": {
      "titulo": "¡Nueva oferta disponible!",
      "mensaje": "Martes de Gaseosas - 30% de descuento en todas las bebidas",
      "datos": {
        "tipo": "OFERTA",
        "ofertaId": 1,
        "url": "/ofertas/martes-gaseosas"
      }
    }
  }'
```

##### Enviar a un cliente específico
```bash
curl -X POST http://localhost:8080/restaurante/v1/push/enviar \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "remitente": {
      "tipo": "SISTEMA"
    },
    "destinatarios": {
      "tipo": "CLIENTE",
      "documentoCliente": 1015466495
    },
    "notificacion": {
      "titulo": "¡Tu pedido está listo!",
      "mensaje": "Tu pedido #123 está listo para recoger",
      "datos": {
        "tipo": "PEDIDO",
        "pedidoId": 123,
        "url": "/pedidos/123"
      }
    }
  }'
```

##### Enviar por topic (tema)
```bash
curl -X POST http://localhost:8080/restaurante/v1/push/enviar \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "remitente": {
      "tipo": "TRABAJADOR",
      "documentoTrabajador": 1000000000,
      "nombre": "María García"
    },
    "destinatarios": {
      "tipo": "TOPIC",
      "topic": "domicilios"
    },
    "notificacion": {
      "titulo": "Nuevo domicilio asignado",
      "mensaje": "Se te ha asignado un nuevo domicilio en la zona norte",
      "datos": {
        "tipo": "DOMICILIO",
        "domicilioId": 456,
        "url": "/domicilios/456"
      }
    }
  }'
```

### 🎫 Cupones

#### Crear Cupón Global
```bash
curl -X POST http://localhost:8080/restaurante/v1/cupones \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "codigo": "BIENVENIDA10",
    "scope": "GLOBAL",
    "tipoDescuento": "PORCENTAJE",
    "valorDescuento": 10,
    "fechaInicio": "2025-01-01",
    "fechaFin": "2025-12-31"
  }'
```

#### Crear Cupón para Producto Específico
```bash
curl -X POST http://localhost:8080/restaurante/v1/cupones \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "codigo": "HAMB5K",
    "scope": "PRODUCTO",
    "tipoDescuento": "MONTO",
    "valorDescuento": 5000,
    "fechaInicio": "2025-01-01",
    "fechaFin": "2025-12-31",
    "montoMinimo": 10000,
    "productoId": 3
  }'
```

#### Crear Cupón para Cliente VIP
```bash
curl -X POST http://localhost:8080/restaurante/v1/cupones \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "codigo": "CLIENTEVIP20",
    "scope": "CLIENTE",
    "tipoDescuento": "PORCENTAJE",
    "valorDescuento": 20,
    "fechaInicio": "2025-01-01",
    "fechaFin": "2025-12-31",
    "clienteId": 1015466495
  }'
```

#### Validar Cupón con Items
```bash
curl -X POST http://localhost:8080/restaurante/v1/cupones/validar \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "codigo": "BIENVENIDA10",
    "clienteId": 1015466495,
    "items": [
      {
        "productoId": 1,
        "cantidad": 2,
        "precio": 15000
      },
      {
        "productoId": 2,
        "cantidad": 1,
        "precio": 8000
      }
    ]
  }'
```

#### Redimir Cupón
```bash
curl -X POST http://localhost:8080/restaurante/v1/cupones/BIENVENIDA10/redimir \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "clienteId": 1015466495,
    "pedidoId": 1
  }'
```

### 🎯 Ofertas

#### Crear Oferta por Días de la Semana
```bash
curl -X POST http://localhost:8080/restaurante/v1/ofertas \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "titulo": "Martes de Gaseosas",
    "tipoDescuento": "PORCENTAJE",
    "valorDescuento": 30,
    "fechaInicio": "2025-01-01",
    "fechaFin": "2025-12-31",
    "diasSemana": ["Martes"],
    "restauranteId": 1
  }'
```

#### Asociar Productos a Oferta
```bash
curl -X POST http://localhost:8080/restaurante/v1/ofertas/1/productos \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "productoId": 1
  }'

curl -X POST http://localhost:8080/restaurante/v1/ofertas/1/productos \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "productoId": 2
  }'
```

#### Obtener Ofertas Activas (Público)
```bash
# Ofertas activas para un restaurante específico
curl -X GET "http://localhost:8080/restaurante/v1/ofertas/activas?restaurante_id=1"

# Ofertas activas para un día específico
curl -X GET "http://localhost:8080/restaurante/v1/ofertas/activas?restaurante_id=1&fecha=2025-03-18"

# Ofertas activas para un producto específico
curl -X GET "http://localhost:8080/restaurante/v1/ofertas/activas?restaurante_id=1&producto_id=1"
```

### 💰 Descuentos en Pedidos

#### Aplicar Descuento por Cupón
```bash
curl -X POST http://localhost:8080/restaurante/v1/pedidos/1/descuentos \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "cuponId": 1,
    "montoDescuento": 2000,
    "detalle": {
      "fuente": "CUPON",
      "codigo": "BIENVENIDA10",
      "tipo": "PORCENTAJE",
      "valor": 10
    }
  }'
```

#### Aplicar Descuento por Oferta
```bash
curl -X POST http://localhost:8080/restaurante/v1/pedidos/2/descuentos \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "ofertaId": 1,
    "montoDescuento": 4500,
    "detalle": {
      "fuente": "OFERTA",
      "titulo": "Martes de Gaseosas",
      "tipo": "PORCENTAJE",
      "valor": 30
    }
  }'
```

#### Listar Descuentos de un Pedido
```bash
curl -X GET http://localhost:8080/restaurante/v1/pedidos/1/descuentos \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

---

## 🎭 Respuestas de Ejemplo

### Validación de Cupón Exitosa
```json
{
  "code": 200,
  "message": "Cupón validado exitosamente",
  "data": {
    "aplicable": true,
    "montoDescuento": 2000,
    "motivo": null
  }
}
```

### Validación de Cupón Fallida
```json
{
  "code": 200,
  "message": "Cupón validado",
  "data": {
    "aplicable": false,
    "montoDescuento": 0,
    "motivo": "El monto mínimo requerido es 10000"
  }
}
```

### Ofertas Activas
```json
{
  "code": 200,
  "message": "Ofertas activas obtenidas exitosamente",
  "data": [
    {
      "ofertaId": 1,
      "titulo": "Martes de Gaseosas",
      "tipoDescuento": "PORCENTAJE",
      "valorDescuento": 30,
      "productosIds": [1, 2]
    }
  ]
}
```

### ⭐ Respuesta de Envío de Notificación (NUEVO)
```json
{
  "totalDispositivos": 15,
  "enviosExitosos": 13,
  "enviosFallidos": 2,
  "detalleEnvios": [
    {
      "pushDispositivoId": 1,
      "plataforma": "WEB",
      "exito": true,
      "statusCode": 200,
      "documentoCliente": 1015466495
    },
    {
      "pushDispositivoId": 2,
      "plataforma": "ANDROID",
      "exito": true,
      "statusCode": 200,
      "documentoTrabajador": 1000000000
    },
    {
      "pushDispositivoId": 3,
      "plataforma": "IOS",
      "exito": false,
      "statusCode": 410,
      "errorCode": "DEVICE_UNREGISTERED",
      "documentoCliente": 1023456789
    }
  ],
  "resumenDestinatarios": {
    "tipoDestinatario": "TODOS",
    "clientesNotificados": [1015466495, 1023456789, 1034567890],
    "trabajadoresNotificados": [1000000000, 1087654321],
    "topicsNotificados": []
  }
}
```

### Error de Validación (422)
```json
{
  "code": 422,
  "message": "Error de validación",
  "data": {
    "errors": {
      "valorDescuento": "el porcentaje de descuento debe estar entre 1 y 100"
    }
  }
}
```

### Error de Conflicto (409)
```json
{
  "code": 409,
  "message": "Conflicto de recursos",
  "data": {
    "error": "el pedido ya tiene un descuento aplicado",
    "code": "DESCUENTO_DUPLICADO"
  }
}
```

---

## 🏗️ Modelos TypeScript para Frontend

```typescript
// notificationTypes.ts

export type PlataformaNotificacion = 'WEB' | 'ANDROID' | 'IOS';
export type ProveedorPush = 'WEB_PUSH' | 'FCM';
export type TipoDescuento = 'PORCENTAJE' | 'MONTO';
export type CuponScope = 'GLOBAL' | 'PRODUCTO' | 'CATEGORIA' | 'CLIENTE';

export interface ApiResponse<T> {
  code: number;
  message: string;
  data: T;
  cause?: string;
}

// Push Notifications
export interface PushDispositivo {
  pushDispositivoId: number;
  plataforma: PlataformaNotificacion;
  endpoint?: string;
  p256dh?: string;
  auth?: string;
  fcmToken?: string;
  enabled: boolean;
  locale?: string;
  timeZone?: string;
  appVersion?: string;
  userAgent?: string;
  subscribedTopics: string[];
  documentoCliente?: number;
  documentoTrabajador?: number;
  createdAt: string;
  lastSeenAt?: string;
}

export interface RegistrarDispositivoRequest {
  plataforma: PlataformaNotificacion;
  endpoint?: string;
  p256dh?: string;
  auth?: string;
  fcmToken?: string;
  locale?: string;
  timeZone?: string;
  appVersion?: string;
  userAgent?: string;
  subscribedTopics?: string[];
  documentoCliente?: number;
  documentoTrabajador?: number;
}

export interface PushEnvio {
  pushEnvioId: number;
  pushDispositivoId: number;
  proveedor: ProveedorPush;
  data?: any;
  exito: boolean;
  statusCode?: number;
  errorCode?: string;
  sentAt: string;
}

// Cupones
export interface Cupon {
  cuponId: number;
  codigo: string;
  scope: CuponScope;
  tipoDescuento: TipoDescuento;
  valorDescuento: number;
  fechaInicio: string;
  fechaFin: string;
  montoMinimo?: number;
  maxUsos?: number;
  limitePorCliente?: number;
  activo: boolean;
  productoId?: number;
  categoriaId?: number;
  clienteId?: number;
}

export interface CrearCuponRequest {
  codigo: string;
  scope: CuponScope;
  tipoDescuento: TipoDescuento;
  valorDescuento: number;
  fechaInicio: string;
  fechaFin: string;
  montoMinimo?: number;
  maxUsos?: number;
  limitePorCliente?: number;
  productoId?: number;
  categoriaId?: number;
  clienteId?: number;
}

export interface ValidarCuponRequest {
  codigo: string;
  clienteId: number;
  items: ValidarCuponItemRequest[];
}

export interface ValidarCuponItemRequest {
  productoId: number;
  cantidad: number;
  precio: number;
}

export interface ValidarCuponResponse {
  aplicable: boolean;
  montoDescuento: number;
  motivo?: string;
}

export interface CuponRedencion {
  redencionId: number;
  cuponId: number;
  codigo: string;
  clienteId: number;
  pedidoId?: number;
  montoDescuento: number;
  fechaRedencion: string;
}

// Ofertas
export interface Oferta {
  ofertaId: number;
  titulo: string;
  tipoDescuento: TipoDescuento;
  valorDescuento: number;
  fechaInicio: string;
  fechaFin: string;
  diasSemana?: string[];
  horaInicio?: string;
  horaFin?: string;
  activo: boolean;
  restauranteId: number;
}

export interface OfertaActiva {
  ofertaId: number;
  titulo: string;
  tipoDescuento: TipoDescuento;
  valorDescuento: number;
  productosIds: number[];
}

export interface CrearOfertaRequest {
  titulo: string;
  tipoDescuento: TipoDescuento;
  valorDescuento: number;
  fechaInicio: string;
  fechaFin: string;
  diasSemana?: string[];
  horaInicio?: string;
  horaFin?: string;
  restauranteId: number;
}

// Descuentos
export interface PedidoDescuentoAplicado {
  pedidoDescuentoId: number;
  pedidoId: number;
  cuponId?: number;
  ofertaId?: number;
  montoDescuento: number;
  detalle?: any;
  createdAt: string;
}

export interface AplicarDescuentoRequest {
  cuponId?: number;
  ofertaId?: number;
  montoDescuento: number;
  detalle?: any;
}

// Parámetros de consulta
export interface NotificationParams {
  limit?: number;
  offset?: number;
  plataforma?: PlataformaNotificacion;
  enabled?: boolean;
  cliente_id?: number;
  trabajador_id?: number;
}

export interface CuponParams {
  limit?: number;
  offset?: number;
  activo?: boolean;
  scope?: CuponScope;
  tipo_descuento?: TipoDescuento;
}

export interface OfertaParams {
  limit?: number;
  offset?: number;
  activo?: boolean;
  restaurante_id?: number;
  fecha?: string;
  hora?: string;
  producto_id?: number;
}

// ⭐ NUEVO: Tipos para envío de notificaciones
export type TipoRemitente = 'TRABAJADOR' | 'SISTEMA';
export type TipoDestinatario = 'TODOS' | 'CLIENTE' | 'TRABAJADOR' | 'TOPIC';

export interface RemitenteNotificacion {
  tipo: TipoRemitente;
  documentoTrabajador?: number;
  nombre?: string;
}

export interface DestinatariosNotificacion {
  tipo: TipoDestinatario;
  documentoCliente?: number;
  documentoTrabajador?: number;
  topic?: string;
}

export interface ContenidoNotificacion {
  titulo: string;
  mensaje: string;
  datos?: any;
}

export interface EnviarNotificacionRequest {
  remitente: RemitenteNotificacion;
  destinatarios: DestinatariosNotificacion;
  notificacion: ContenidoNotificacion;
}

export interface DetalleEnvioNotificacion {
  pushDispositivoId: number;
  plataforma: string;
  exito: boolean;
  statusCode?: number;
  errorCode?: string;
  documentoCliente?: number;
  documentoTrabajador?: number;
}

export interface ResumenDestinatarios {
  tipoDestinatario: string;
  clientesNotificados?: number[];
  trabajadoresNotificados?: number[];
  topicsNotificados?: string[];
}

export interface EnviarNotificacionResponse {
  totalDispositivos: number;
  enviosExitosos: number;
  enviosFallidos: number;
  detalleEnvios: DetalleEnvioNotificacion[];
  resumenDestinatarios: ResumenDestinatarios;
}
```

---

## 🎯 Casos de Uso Completos

### 1. Flujo de Notificación Push
1. **Registrar dispositivo** → `POST /push/dispositivos`
2. **Actualizar topics** → `PATCH /push/dispositivos/{id}/topics`
3. **Enviar notificación** → `POST /push/envios`
4. **Actualizar última vista** → `PATCH /push/dispositivos/{id}/visto`

### 2. Flujo de Cupón
1. **Crear cupón** → `POST /cupones`
2. **Validar cupón** → `POST /cupones/validar`
3. **Redimir cupón** → `POST /cupones/{codigo}/redimir`
4. **Aplicar descuento** → `POST /pedidos/{id}/descuentos`

### 3. Flujo de Oferta
1. **Crear oferta** → `POST /ofertas`
2. **Asociar productos** → `POST /ofertas/{id}/productos`
3. **Consultar ofertas activas** → `GET /ofertas/activas`
4. **Aplicar descuento** → `POST /pedidos/{id}/descuentos`

---

---

# 📱 Guía Completa para Implementar Frontend de Notificaciones, Cupones y Ofertas

## 🎯 Prompt para Cursor

```
Necesito implementar un servicio completo de notificaciones push, cupones y ofertas en TypeScript/React que consuma los siguientes endpoints de mi API backend.

Crea:
1. **Servicios especializados** (`pushService.ts`, `cuponService.ts`, `ofertaService.ts`, `descuentoService.ts`) con todos los métodos
2. **Modelos TypeScript** (`notificationTypes.ts`) con todas las interfaces tipadas
3. **Mocks de datos** (`notificationMocks.ts`) para desarrollo y testing
4. **Hooks personalizados** (`usePush.ts`, `useCupones.ts`, `useOfertas.ts`) para React con manejo de estado
5. **Componentes de ejemplo** para gestión de notificaciones, cupones y ofertas
6. **Utilidades** para validación de cupones y manejo de notificaciones push

Usa Axios para las peticiones HTTP, incluye manejo de errores, loading states, y cache básico.
```

---

## 📡 Endpoints Disponibles

### Base URL: `http://localhost:8080/restaurante/v1`
### Autenticación: `Authorization: Bearer {token}` (excepto ofertas activas)

---

## 📖 **Convenciones de Campos y Datos**

### **💰 Campos Monetarios**
- **Todos los valores monetarios están en pesos colombianos (COP)**
- **Formato**: Números enteros sin decimales (ej: `5000` = $5,000 COP)
- **Campos típicos**: `valorDescuento`, `montoMinimo`, `montoDescuento`

### **📅 Campos de Fecha y Hora**
- **Fechas**: Formato `YYYY-MM-DD` (ej: `"2025-01-01"`)
- **Fechas con hora**: Formato ISO 8601 con UTC (ej: `"2025-01-15T14:30:00Z"`)
- **Horas**: Formato 24 horas `HH:MM:SS` (ej: `"14:30:00"`)

### **🔢 Campos Numéricos**
- **Porcentajes**: Números enteros (ej: `20` = 20%)
- **IDs**: Números enteros únicos
- **Cantidades**: Números enteros positivos

### **👤 Identificadores**
- **Documentos de clientes**: `documentoCliente` (número entero)
- **Documentos de trabajadores**: `documentoTrabajador` (número entero)
- **IDs de productos**: `productoId` (número entero)

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

## 📱 **1. Push Notifications**

### **Endpoint Base:** `/restaurante/v1/push`

#### **Registrar Dispositivo**
```bash
curl -X POST "http://localhost:8080/restaurante/v1/push/dispositivos" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer {token}" \
  -d '{
    "plataforma": "WEB",
    "endpoint": "https://push.example.com/subscription/cliente_1015466495",
    "p256dh": "test_p256dh_1",
    "auth": "test_auth_1",
    "locale": "es-CO",
    "timeZone": "America/Bogota",
    "appVersion": "1.0.0",
    "subscribedTopics": ["promos", "novedades"],
    "documentoCliente": 1015466495
  }'
```

**Respuesta esperada:**
```json
{
  "code": 200,
  "message": "Dispositivo registrado exitosamente",
  "data": {
    "pushDispositivoId": 1,
    "plataforma": "WEB",
    "enabled": true,
    "createdAt": "2025-01-15T14:30:00Z"
  }
}
```

#### **Listar Dispositivos**
```bash
curl -X GET "http://localhost:8080/restaurante/v1/push/dispositivos?limit=10&offset=0" \
  -H "Authorization: Bearer {token}"
```

#### **Enviar Notificación Push** ⭐ **NUEVO**
```bash
curl -X POST "http://localhost:8080/restaurante/v1/push/enviar" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer {token}" \
  -d '{
    "remitente": {
      "tipo": "TRABAJADOR",
      "documentoTrabajador": 1000000000,
      "nombre": "Juan Pérez"
    },
    "destinatarios": {
      "tipo": "TODOS"
    },
    "notificacion": {
      "titulo": "¡Nueva oferta disponible!",
      "mensaje": "Martes de Gaseosas - 30% de descuento",
      "datos": {
        "tipo": "OFERTA",
        "ofertaId": 1,
        "url": "/ofertas/martes-gaseosas"
      }
    }
  }'
```

---

## 🎫 **2. Cupones**

### **Endpoint Base:** `/restaurante/v1/cupones`

#### **Crear Cupón**
```bash
curl -X POST "http://localhost:8080/restaurante/v1/cupones" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer {token}" \
  -d '{
    "codigo": "BIENVENIDA10",
    "scope": "GLOBAL",
    "tipoDescuento": "PORCENTAJE",
    "valorDescuento": 10,
    "fechaInicio": "2025-01-01",
    "fechaFin": "2025-12-31"
  }'
```

#### **Validar Cupón**
```bash
curl -X POST "http://localhost:8080/restaurante/v1/cupones/validar" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer {token}" \
  -d '{
    "codigo": "BIENVENIDA10",
    "clienteId": 1015466495,
    "items": [
      {
        "productoId": 1,
        "cantidad": 2,
        "precio": 15000
      }
    ]
  }'
```

**Respuesta esperada:**
```json
{
  "code": 200,
  "message": "Cupón validado exitosamente",
  "data": {
    "aplicable": true,
    "montoDescuento": 2000,
    "motivo": null
  }
}
```

#### **Redimir Cupón**
```bash
curl -X POST "http://localhost:8080/restaurante/v1/cupones/BIENVENIDA10/redimir" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer {token}" \
  -d '{
    "clienteId": 1015466495,
    "pedidoId": 1
  }'
```

---

## 🎯 **3. Ofertas**

### **Endpoint Base:** `/restaurante/v1/ofertas`

#### **Crear Oferta**
```bash
curl -X POST "http://localhost:8080/restaurante/v1/ofertas" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer {token}" \
  -d '{
    "titulo": "Martes de Gaseosas",
    "tipoDescuento": "PORCENTAJE",
    "valorDescuento": 30,
    "fechaInicio": "2025-01-01",
    "fechaFin": "2025-12-31",
    "diasSemana": ["Martes"],
    "restauranteId": 1
  }'
```

#### **Obtener Ofertas Activas** ⚡ **SIN AUTENTICACIÓN**
```bash
curl -X GET "http://localhost:8080/restaurante/v1/ofertas/activas?restaurante_id=1&fecha=2025-03-18"
```

**Respuesta esperada:**
```json
{
  "code": 200,
  "message": "Ofertas activas obtenidas exitosamente",
  "data": [
    {
      "ofertaId": 1,
      "titulo": "Martes de Gaseosas",
      "tipoDescuento": "PORCENTAJE",
      "valorDescuento": 30,
      "productosIds": [1, 2]
    }
  ]
}
```

#### **Asociar Producto a Oferta**
```bash
curl -X POST "http://localhost:8080/restaurante/v1/ofertas/1/productos" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer {token}" \
  -d '{
    "productoId": 1
  }'
```

---

## 💰 **4. Descuentos**

### **Endpoint Base:** `/restaurante/v1/descuentos`

#### **Aplicar Descuento a Pedido**
```bash
curl -X POST "http://localhost:8080/restaurante/v1/descuentos/pedidos" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer {token}" \
  -d '{
    "pedidoId": 1,
    "cuponId": 1,
    "montoDescuento": 2000,
    "detalle": {
      "fuente": "CUPON",
      "codigo": "BIENVENIDA10",
      "tipo": "PORCENTAJE",
      "valor": 10
    }
  }'
```

#### **Listar Descuentos de Pedidos**
```bash
curl -X GET "http://localhost:8080/restaurante/v1/descuentos/pedidos?limit=10&offset=0" \
  -H "Authorization: Bearer {token}"
```

---

## 🏗️ **Modelos TypeScript Completos**

```typescript
// notificationTypes.ts

export type PlataformaNotificacion = 'WEB' | 'ANDROID' | 'IOS';
export type ProveedorPush = 'WEB_PUSH' | 'FCM';
export type TipoDescuento = 'PORCENTAJE' | 'MONTO';
export type CuponScope = 'GLOBAL' | 'PRODUCTO' | 'CATEGORIA' | 'CLIENTE';
export type TipoRemitente = 'TRABAJADOR' | 'SISTEMA';
export type TipoDestinatario = 'TODOS' | 'CLIENTE' | 'TRABAJADOR' | 'TOPIC';

export interface ApiResponse<T> {
  code: number;
  message: string;
  data: T;
  cause?: string;
}

// Push Notifications
export interface PushDispositivo {
  pushDispositivoId: number;
  plataforma: PlataformaNotificacion;
  endpoint?: string;
  p256dh?: string;
  auth?: string;
  fcmToken?: string;
  enabled: boolean;
  locale?: string;
  timeZone?: string;
  appVersion?: string;
  userAgent?: string;
  subscribedTopics: string[];
  documentoCliente?: number;
  documentoTrabajador?: number;
  createdAt: string;
  lastSeenAt?: string;
}

export interface RegistrarDispositivoRequest {
  plataforma: PlataformaNotificacion;
  endpoint?: string;
  p256dh?: string;
  auth?: string;
  fcmToken?: string;
  locale?: string;
  timeZone?: string;
  appVersion?: string;
  userAgent?: string;
  subscribedTopics?: string[];
  documentoCliente?: number;
  documentoTrabajador?: number;
}

export interface EnviarNotificacionRequest {
  remitente: {
    tipo: TipoRemitente;
    documentoTrabajador?: number;
    nombre?: string;
  };
  destinatarios: {
    tipo: TipoDestinatario;
    documentoCliente?: number;
    documentoTrabajador?: number;
    topic?: string;
  };
  notificacion: {
    titulo: string;
    mensaje: string;
    datos?: any;
  };
}

export interface EnviarNotificacionResponse {
  totalDispositivos: number;
  enviosExitosos: number;
  enviosFallidos: number;
  detalleEnvios: Array<{
    pushDispositivoId: number;
    plataforma: string;
    exito: boolean;
    statusCode?: number;
    errorCode?: string;
    documentoCliente?: number;
    documentoTrabajador?: number;
  }>;
  resumenDestinatarios: {
    tipoDestinatario: string;
    clientesNotificados?: number[];
    trabajadoresNotificados?: number[];
    topicsNotificados?: string[];
  };
}

// Cupones
export interface Cupon {
  cuponId: number;
  codigo: string;
  scope: CuponScope;
  tipoDescuento: TipoDescuento;
  valorDescuento: number;
  fechaInicio: string;
  fechaFin: string;
  montoMinimo?: number;
  maxUsos?: number;
  limitePorCliente?: number;
  activo: boolean;
  productoId?: number;
  categoriaId?: number;
  clienteId?: number;
}

export interface CrearCuponRequest {
  codigo: string;
  scope: CuponScope;
  tipoDescuento: TipoDescuento;
  valorDescuento: number;
  fechaInicio: string;
  fechaFin: string;
  montoMinimo?: number;
  maxUsos?: number;
  limitePorCliente?: number;
  productoId?: number;
  categoriaId?: number;
  clienteId?: number;
}

export interface ValidarCuponRequest {
  codigo: string;
  clienteId: number;
  items: Array<{
    productoId: number;
    cantidad: number;
    precio: number;
  }>;
}

export interface ValidarCuponResponse {
  aplicable: boolean;
  montoDescuento: number;
  motivo?: string;
}

export interface RedimirCuponRequest {
  clienteId: number;
  pedidoId?: number;
}

export interface CuponRedencion {
  redencionId: number;
  cuponId: number;
  codigo: string;
  clienteId: number;
  pedidoId?: number;
  montoDescuento: number;
  fechaRedencion: string;
}

// Ofertas
export interface Oferta {
  ofertaId: number;
  titulo: string;
  tipoDescuento: TipoDescuento;
  valorDescuento: number;
  fechaInicio: string;
  fechaFin: string;
  diasSemana?: string[];
  horaInicio?: string;
  horaFin?: string;
  activo: boolean;
  restauranteId: number;
}

export interface CrearOfertaRequest {
  titulo: string;
  tipoDescuento: TipoDescuento;
  valorDescuento: number;
  fechaInicio: string;
  fechaFin: string;
  diasSemana?: string[];
  horaInicio?: string;
  horaFin?: string;
  restauranteId: number;
}

export interface OfertaActiva {
  ofertaId: number;
  titulo: string;
  tipoDescuento: TipoDescuento;
  valorDescuento: number;
  productosIds: number[];
}

export interface AsociarProductoRequest {
  productoId: number;
}

// Descuentos
export interface PedidoDescuentoAplicado {
  pedidoDescuentoId: number;
  pedidoId: number;
  cuponId?: number;
  ofertaId?: number;
  montoDescuento: number;
  detalle?: any;
  createdAt: string;
}

export interface AplicarDescuentoRequest {
  pedidoId: number;
  cuponId?: number;
  ofertaId?: number;
  montoDescuento: number;
  detalle?: any;
}

// Parámetros de consulta
export interface PushParams {
  limit?: number;
  offset?: number;
  plataforma?: PlataformaNotificacion;
  enabled?: boolean;
  documentoCliente?: number;
  documentoTrabajador?: number;
}

export interface CuponParams {
  limit?: number;
  offset?: number;
  activo?: boolean;
  scope?: CuponScope;
  tipoDescuento?: TipoDescuento;
}

export interface OfertaParams {
  limit?: number;
  offset?: number;
  activo?: boolean;
  restauranteId?: number;
  fecha?: string;
  hora?: string;
  productoId?: number;
}

export interface DescuentoParams {
  limit?: number;
  offset?: number;
  pedidoId?: number;
  cuponId?: number;
  ofertaId?: number;
}
```

---

## 🎭 **Mocks de Datos Completos**

```typescript
// notificationMocks.ts

import {
  ApiResponse,
  PushDispositivo,
  Cupon,
  Oferta,
  CuponRedencion,
  PedidoDescuentoAplicado,
  ValidarCuponResponse,
  EnviarNotificacionResponse,
  OfertaActiva
} from './notificationTypes';

// Push Dispositivos
export const mockPushDispositivoWeb: ApiResponse<PushDispositivo> = {
  code: 200,
  message: "Dispositivo obtenido exitosamente",
  data: {
    pushDispositivoId: 1,
    plataforma: "WEB",
    endpoint: "https://push.example.com/subscription/cliente_1015466495",
    p256dh: "test_p256dh_1",
    auth: "test_auth_1",
    fcmToken: null,
    enabled: true,
    locale: "es-CO",
    timeZone: "America/Bogota",
    appVersion: "1.0.0",
    userAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
    subscribedTopics: ["promos", "novedades"],
    documentoCliente: 1015466495,
    documentoTrabajador: null,
    createdAt: "2025-01-01T10:00:00Z",
    lastSeenAt: "2025-01-15T14:30:00Z"
  }
};

export const mockPushDispositivoAndroid: ApiResponse<PushDispositivo> = {
  code: 200,
  message: "Dispositivo obtenido exitosamente",
  data: {
    pushDispositivoId: 2,
    plataforma: "ANDROID",
    endpoint: null,
    p256dh: null,
    auth: null,
    fcmToken: "fcm_token_trabajador_1000000000",
    enabled: true,
    locale: "es-CO",
    timeZone: "America/Bogota",
    appVersion: "2.1.0",
    userAgent: null,
    subscribedTopics: ["domicilios"],
    documentoCliente: null,
    documentoTrabajador: 1000000000,
    createdAt: "2025-01-01T08:00:00Z",
    lastSeenAt: "2025-01-15T16:45:00Z"
  }
};

export const mockListaDispositivos: ApiResponse<PushDispositivo[]> = {
  code: 200,
  message: "Dispositivos obtenidos exitosamente",
  data: [
    mockPushDispositivoWeb.data,
    mockPushDispositivoAndroid.data
  ]
};

// Cupones
export const mockCuponBienvenida: ApiResponse<Cupon> = {
  code: 200,
  message: "Cupón obtenido exitosamente",
  data: {
    cuponId: 1,
    codigo: "BIENVENIDA10",
    scope: "GLOBAL",
    tipoDescuento: "PORCENTAJE",
    valorDescuento: 10,
    fechaInicio: "2025-01-01",
    fechaFin: "2025-12-31",
    montoMinimo: null,
    maxUsos: null,
    limitePorCliente: null,
    activo: true,
    productoId: null,
    categoriaId: null,
    clienteId: null
  }
};

export const mockCuponHamburguesa: ApiResponse<Cupon> = {
  code: 200,
  message: "Cupón obtenido exitosamente",
  data: {
    cuponId: 2,
    codigo: "HAMB5K",
    scope: "PRODUCTO",
    tipoDescuento: "MONTO",
    valorDescuento: 5000,
    fechaInicio: "2025-01-01",
    fechaFin: "2025-12-31",
    montoMinimo: 10000,
    maxUsos: null,
    limitePorCliente: null,
    activo: true,
    productoId: 3,
    categoriaId: null,
    clienteId: null
  }
};

export const mockCuponClienteVIP: ApiResponse<Cupon> = {
  code: 200,
  message: "Cupón obtenido exitosamente",
  data: {
    cuponId: 3,
    codigo: "CLIENTEVIP20",
    scope: "CLIENTE",
    tipoDescuento: "PORCENTAJE",
    valorDescuento: 20,
    fechaInicio: "2025-01-01",
    fechaFin: "2025-12-31",
    montoMinimo: null,
    maxUsos: null,
    limitePorCliente: null,
    activo: true,
    productoId: null,
    categoriaId: null,
    clienteId: 1015466495
  }
};

export const mockListaCupones: ApiResponse<Cupon[]> = {
  code: 200,
  message: "Cupones obtenidos exitosamente",
  data: [
    mockCuponBienvenida.data,
    mockCuponHamburguesa.data,
    mockCuponClienteVIP.data
  ]
};

export const mockValidarCuponExitoso: ApiResponse<ValidarCuponResponse> = {
  code: 200,
  message: "Cupón validado exitosamente",
  data: {
    aplicable: true,
    montoDescuento: 2000,
    motivo: null
  }
};

export const mockValidarCuponFallido: ApiResponse<ValidarCuponResponse> = {
  code: 200,
  message: "Cupón validado",
  data: {
    aplicable: false,
    montoDescuento: 0,
    motivo: "El monto mínimo requerido es 10000"
  }
};

// Redenciones
export const mockRedencionCupon: ApiResponse<CuponRedencion> = {
  code: 200,
  message: "Cupón redimido exitosamente",
  data: {
    redencionId: 1,
    cuponId: 1,
    codigo: "BIENVENIDA10",
    clienteId: 1015466495,
    pedidoId: 1,
    montoDescuento: 2000,
    fechaRedencion: "2025-01-15T14:30:00Z"
  }
};

// Ofertas
export const mockOfertaMartes: ApiResponse<Oferta> = {
  code: 200,
  message: "Oferta obtenida exitosamente",
  data: {
    ofertaId: 1,
    titulo: "Martes de Gaseosas",
    tipoDescuento: "PORCENTAJE",
    valorDescuento: 30,
    fechaInicio: "2025-01-01",
    fechaFin: "2025-12-31",
    diasSemana: ["Martes"],
    horaInicio: null,
    horaFin: null,
    activo: true,
    restauranteId: 1
  }
};

export const mockOfertasActivas: ApiResponse<OfertaActiva[]> = {
  code: 200,
  message: "Ofertas activas obtenidas exitosamente",
  data: [
    {
      ofertaId: 1,
      titulo: "Martes de Gaseosas",
      tipoDescuento: "PORCENTAJE",
      valorDescuento: 30,
      productosIds: [1, 2]
    }
  ]
};

// Descuentos
export const mockDescuentoAplicado: ApiResponse<PedidoDescuentoAplicado> = {
  code: 200,
  message: "Descuento aplicado exitosamente",
  data: {
    pedidoDescuentoId: 1,
    pedidoId: 1,
    cuponId: 1,
    ofertaId: null,
    montoDescuento: 2000,
    detalle: {
      "fuente": "CUPON",
      "codigo": "BIENVENIDA10",
      "tipo": "PORCENTAJE",
      "valor": 10
    },
    createdAt: "2025-01-15T14:30:00Z"
  }
};

// Notificaciones
export const mockEnvioNotificacionExitoso: ApiResponse<EnviarNotificacionResponse> = {
  code: 200,
  message: "Notificación enviada exitosamente",
  data: {
    totalDispositivos: 15,
    enviosExitosos: 13,
    enviosFallidos: 2,
    detalleEnvios: [
      {
        pushDispositivoId: 1,
        plataforma: "WEB",
        exito: true,
        statusCode: 200,
        documentoCliente: 1015466495
      },
      {
        pushDispositivoId: 2,
        plataforma: "ANDROID",
        exito: true,
        statusCode: 200,
        documentoTrabajador: 1000000000
      },
      {
        pushDispositivoId: 3,
        plataforma: "IOS",
        exito: false,
        statusCode: 410,
        errorCode: "DEVICE_UNREGISTERED",
        documentoCliente: 1023456789
      }
    ],
    resumenDestinatarios: {
      tipoDestinatario: "TODOS",
      clientesNotificados: [1015466495, 1023456789, 1034567890],
      trabajadoresNotificados: [1000000000, 1087654321],
      topicsNotificados: []
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

// Mock de error de validación
export const mockValidationError = {
  code: 422,
  message: "Error de validación",
  data: {
    errors: {
      valorDescuento: "el porcentaje de descuento debe estar entre 1 y 100"
    }
  }
};

// Mock de error de conflicto
export const mockConflictError = {
  code: 409,
  message: "Conflicto de recursos",
  data: {
    error: "el pedido ya tiene un descuento aplicado",
    code: "DESCUENTO_DUPLICADO"
  }
};
```

---

## 🔧 **Servicios TypeScript**

```typescript
// notificationService.ts - Estructura base esperada

class NotificationService {
  private baseURL = 'http://localhost:8080/restaurante/v1';

  // Push Notifications
  async registrarDispositivo(data: RegistrarDispositivoRequest): Promise<ApiResponse<PushDispositivo>>
  async listarDispositivos(params?: PushParams): Promise<ApiResponse<PushDispositivo[]>>
  async actualizarTopics(id: number, topics: string[]): Promise<ApiResponse<void>>
  async actualizarUltimaVista(id: number): Promise<ApiResponse<void>>
  async enviarNotificacion(data: EnviarNotificacionRequest): Promise<ApiResponse<EnviarNotificacionResponse>>

  // Cupones
  async crearCupon(data: CrearCuponRequest): Promise<ApiResponse<Cupon>>
  async listarCupones(params?: CuponParams): Promise<ApiResponse<Cupon[]>>
  async obtenerCupon(id: number): Promise<ApiResponse<Cupon>>
  async actualizarCupon(id: number, data: Partial<CrearCuponRequest>): Promise<ApiResponse<Cupon>>
  async validarCupon(data: ValidarCuponRequest): Promise<ApiResponse<ValidarCuponResponse>>
  async redimirCupon(codigo: string, data: RedimirCuponRequest): Promise<ApiResponse<CuponRedencion>>
  async listarRedenciones(params?: CuponParams): Promise<ApiResponse<CuponRedencion[]>>

  // Ofertas
  async crearOferta(data: CrearOfertaRequest): Promise<ApiResponse<Oferta>>
  async listarOfertas(params?: OfertaParams): Promise<ApiResponse<Oferta[]>>
  async obtenerOferta(id: number): Promise<ApiResponse<Oferta>>
  async actualizarOferta(id: number, data: Partial<CrearOfertaRequest>): Promise<ApiResponse<Oferta>>
  async obtenerOfertasActivas(params?: OfertaParams): Promise<ApiResponse<OfertaActiva[]>> // SIN AUTH
  async asociarProducto(ofertaId: number, data: AsociarProductoRequest): Promise<ApiResponse<void>>
  async desasociarProducto(ofertaId: number, productoId: number): Promise<ApiResponse<void>>

  // Descuentos
  async aplicarDescuento(data: AplicarDescuentoRequest): Promise<ApiResponse<PedidoDescuentoAplicado>>
  async listarDescuentos(params?: DescuentoParams): Promise<ApiResponse<PedidoDescuentoAplicado[]>>
}
```

---

## 🎨 **Componentes de UI Sugeridos**

```typescript
// Resultados esperados en los componentes

// Push Notifications
- PushDeviceManager: Gestión de dispositivos registrados
- NotificationSender: Interfaz para enviar notificaciones
- TopicSubscriptionManager: Gestión de suscripciones a topics
- NotificationHistory: Historial de notificaciones enviadas

// Cupones
- CuponCreator: Formulario para crear cupones
- CuponValidator: Validador de cupones en tiempo real
- CuponList: Lista de cupones con filtros
- CuponRedemptionHistory: Historial de redenciones
- CuponUsageStats: Estadísticas de uso de cupones

// Ofertas
- OfertaCreator: Formulario para crear ofertas
- OfertaScheduler: Programador de ofertas por días/horas
- OfertasActivasPublic: Componente público para mostrar ofertas activas
- ProductOfertaAssociator: Asociar productos a ofertas
- OfertaCalendar: Calendario de ofertas programadas

// Descuentos
- DescuentoApplicator: Aplicador de descuentos a pedidos
- DescuentoHistory: Historial de descuentos aplicados
- DescuentoValidator: Validador de descuentos antes de aplicar

// Utilidades
- CuponCodeGenerator: Generador de códigos de cupones
- DiscountCalculator: Calculadora de descuentos
- NotificationTemplateEditor: Editor de plantillas de notificaciones
- OfertaTimeValidator: Validador de horarios de ofertas
```

---

## 📝 **Notas Importantes**

1. **Autenticación**: Todos los endpoints requieren token JWT válido (excepto `/ofertas/activas`)
2. **Validaciones**: Los cupones y ofertas tienen validaciones estrictas de negocio
3. **Plataformas Push**: WEB requiere endpoint/p256dh/auth, ANDROID/IOS requiere fcmToken
4. **Exclusividad**: Solo un descuento por pedido permitido
5. **Formato de fechas**: Las fechas se manejan en formato ISO 8601
6. **Moneda**: Los valores monetarios están en pesos colombianos (COP)
7. **Scopes de cupones**: GLOBAL, PRODUCTO, CATEGORIA, CLIENTE
8. **Tipos de descuento**: PORCENTAJE (1-100) o MONTO (valor fijo)
9. **Topics de notificaciones**: Permiten segmentación de audiencias
10. **Ofertas por días**: Soportan días de la semana y rangos horarios opcionales

---

## 🚀 **¡Listo para implementar!**

Con esta guía tienes todo lo necesario para crear un frontend completo de notificaciones push, cupones y ofertas que consuma todos los endpoints del sistema. ¡Manos a la obra! 🎯

---

## Próximos Pasos

1. **Implementar logging** adecuado en los controladores
2. **Añadir tests de integración** con base de datos
3. **Implementar rate limiting** para endpoints públicos
4. **Añadir métricas** de uso de cupones y ofertas
5. **Implementar notificaciones** automáticas para ofertas activas
