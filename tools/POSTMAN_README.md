# 🍽️ Colección Postman - El fogón de María API

## 🎯 Descripción

Colección **profesional y completa** de la API del restaurante "El fogón de María". Incluye **138 requests en 30 folders**, completamente sincronizada con Swagger v2.0.0.

## ✨ Características Profesionales

### 🔐 Autenticación Automática
**El token se captura y usa automáticamente en todos los endpoints:**

1. ✅ Ejecuta `POST /login` con tus credenciales
2. ✅ El token se guarda automáticamente en `{{token}}`
3. ✅ **Todos los endpoints lo usan automáticamente** - ¡No necesitas hacer nada más!

### 🧪 Tests Automáticos Globales
Cada request incluye tests que validan:
- ✅ Status code correcto
- ✅ Tiempo de respuesta < 2 segundos
- ✅ Estructura de respuesta JSON válida
- ✅ Campos requeridos presentes

### 📝 Ejemplos y Documentación
- ✅ Bodies de ejemplo en POST/PUT
- ✅ Descripciones detalladas en cada endpoint
- ✅ Iconos visuales para fácil navegación
- ✅ Pre-request scripts con validación de token

## 📦 Instalación

### Opción 1: Importar en Postman
1. Abre Postman
2. Click en **Import**
3. Selecciona `tools/postman_collection.json`
4. Click en **Import**

### Opción 2: Usar URL (si está en repo público)
1. Abre Postman
2. Click en **Import** → **Link**
3. Pega la URL del archivo
4. Click en **Continue** → **Import**

## 🚀 Uso Rápido

### 1. Configurar Variables (Opcional)

La colección viene con valores por defecto, pero puedes personalizarlos:

- **`baseUrl`**: `http://localhost:8080/restaurante/v1` (default)
- **`token`**: Se autocompleta al hacer login

### 2. Hacer Login

1. Abre la carpeta **🔐 Login**
2. Selecciona **POST /login**
3. Modifica el body con tus credenciales:
   ```json
   {
     "documento": 1015466495,
     "password": "tu_password"
   }
   ```
4. Click en **Send**
5. ✅ **¡El token se guarda automáticamente!**

Verás en la consola:
```
✅ Token guardado automáticamente
🔑 Token: eyJhbGc...
👤 Usuario: Juan
🎭 Rol: Admin
```

### 3. Usar Cualquier Endpoint

Ahora puedes ejecutar **cualquier endpoint autenticado** y el token se usará automáticamente. Por ejemplo:

- **GET /clientes** - Lista clientes
- **POST /pedidos** - Crear pedido
- **GET /telemetria/dashboard** - Ver dashboard

**¡No necesitas copiar/pegar el token manualmente!** 🎉

## 📚 Estructura de la Colección

```
🍽️ El fogón de María API (138 requests en 30 folders)
│
├── 🏥 Health (2)
│   ├── GET /healthz
│   └── GET /readyz
│
├── 🔐 Auth (1)
│   └── POST /auth/refresh
│
├── 🔐 Login (1)
│   └── POST /login ⭐ (Captura automática de token)
│
├── 👥 Clientes (5)
├── 🏪 Restaurantes (5)
├── 👷 Trabajadores (5)
├── 📅 Horario_trabajador (4)
├── 🛒 Productos (8)
├── 📂 Categorias (5)
├── 📂 Subcategorias (5)
├── 💰 Precio_producto_hist (2)
├── 🛍️ Pedidos (6)
├── 📦 Producto_pedido (3)
├── 🚚 Domicilios (6)
├── 💳 Pagos (5)
├── 💳 Metodos_pago (5)
├── ⏰ Cambios_horario (5)
├── 📅 Reservas (10)
├── 📞 Reserva_contacto (2)
├── 🗓️ Restaurante_dia (2)
├── 💵 Nominas (4)
├── 💼 Nomina_trabajador (3)
├── 📋 Control_nomina (2)
├── ⚠️ Incidencias (5)
├── 🔔 Push (17)
├── 🎟️ Cupones (10)
├── 🎁 Ofertas (7)
├── 💸 Descuentos (2)
└── 📈 Telemetria (10)
```

## 🔍 Endpoints Destacados

### Públicos (Sin autenticación)
- `GET /healthz` - Health check
- `GET /productos` - Consultar productos
- `GET /ofertas/activas` - Ver ofertas activas
- `POST /clientes` - Registro de clientes
- `POST /reservas` - Crear reserva

### Autenticados (Con token automático)
- `GET /clientes` - Listar clientes
- `POST /pedidos` - Crear pedido
- `GET /telemetria/dashboard` - Dashboard de analytics
- `POST /cupones` - Crear cupón de descuento
- `GET /trabajadores` - Listar trabajadores

## 🧪 Tests y Validaciones

### Tests Globales (Todos los endpoints)
```javascript
✅ Tiempo de respuesta < 2s
📦 Response es JSON válido
✅ Estructura de respuesta estándar (tiene campo 'success')
```

### Tests Especiales del Login
```javascript
✅ Status code es 200
📦 Respuesta contiene data
🔑 Respuesta contiene token
🔄 Respuesta contiene refreshToken
👤 Respuesta contiene datos del trabajador
💾 Token guardado en variables
```

## 🛠️ Scripts Automáticos

### Pre-Request (Antes de cada request)
```javascript
// Verifica que hay token para endpoints autenticados
// Log del endpoint que se va a ejecutar
```

### Test (Después de cada request)
```javascript
// Valida tiempo de respuesta
// Valida estructura JSON
// Log de resultado con emojis
```

## 📝 Convenciones

### Formatos de Fecha/Hora
- **Fechas**: `YYYY-MM-DD` (ej: `2025-10-20`)
- **Horas**: `HH:MM:SS` (ej: `19:30:00`)
- **Zona horaria**: America/Bogota

### Paginación
```
?limit=20&offset=0
```

### Query Strings
Beego usa query strings, NO path params:
- ✅ `/clientes?id=1015466495`
- ❌ `/clientes/1015466495`

## 🔧 Mantenimiento

### Regenerar Colección desde Swagger
```bash
node tools/swagger_to_postman.js
node tools/polish_postman.js
```

### Validar Sincronización con Swagger
```bash
node tools/validate_postman.js
```

Debe mostrar:
```
🎉 La colección de Postman está completamente sincronizada con Swagger!
```

## 💡 Tips y Trucos

### 1. Ver el Token Actual
```javascript
// En la consola de Postman:
pm.collectionVariables.get('token')
```

### 2. Limpiar Token
```javascript
// En Pre-request Script:
pm.collectionVariables.set('token', '')
```

### 3. Cambiar Ambiente
Puedes crear ambientes diferentes (dev, staging, prod) y solo cambiar:
- `baseUrl`: URL del servidor
- El token se sigue capturando automáticamente

### 4. Ejecutar Toda la Colección
1. Click en los 3 puntos (...) junto al nombre de la colección
2. **Run collection**
3. Verás todos los tests ejecutándose

## ❓ Troubleshooting

### "No hay token. Ejecuta POST /login primero"
**Solución**: Ejecuta el endpoint `POST /login` para obtener el token.

### "Token inválido o expirado"
**Solución**: 
1. Ejecuta `POST /login` de nuevo, o
2. Usa `POST /auth/refresh` para renovar el token

### "Endpoints públicos con error 401"
**Problema**: El endpoint no está marcado como público.
**Solución**: Verifica que tenga `"auth": { "type": "noauth" }` en Postman.

## 📊 Estadísticas

- **Total Folders**: 30
- **Total Requests**: 138
- **Endpoints Públicos**: ~15
- **Endpoints Autenticados**: ~123
- **Versión**: 2.0.0
- **Última actualización**: Octubre 2025

## 🔗 Enlaces Útiles

- [Swagger UI](http://localhost:8080/swagger) - Documentación interactiva
- [Repositorio GitHub](https://github.com/tu-repo) - Código fuente
- [Postman Learning Center](https://learning.postman.com/) - Tutoriales de Postman

## 📄 Licencia

Esta colección es parte del proyecto "El fogón de María" y sigue la misma licencia del proyecto principal.

---

**¿Preguntas o problemas?**  
Revisa la documentación de Swagger o contacta al equipo de desarrollo.

**¡Disfruta de la API! 🚀**

