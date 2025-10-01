# El fogón de María - Backend (Beego v2 + PostgreSQL)

## Descripción
API REST en Go para gestionar operaciones de "El fogón de María": clientes, pedidos, pagos, productos, reservas, nómina y más. Basada en Beego v2 con documentación Swagger.

## Inicio rápido
- Levantar en desarrollo (con Swagger auto-generado):
  ```powershell
  bee run -downdoc=true -gendoc=true
  ```
- Ejecutar pruebas con cobertura:
  ```powershell
  powershell -ExecutionPolicy Bypass -File tools/cover.ps1 -Clean
  ```
  - Resultado esperado actual: ≈99.9% total. El directorio base está en 100%.

## Tecnologías
- Go 1.25
- Beego v2.3.8
- PostgreSQL (`github.com/lib/pq`) v1.10.9
- JWT (`github.com/golang-jwt/jwt/v5`) v5
- Swagger UI (`github.com/swaggo/http-swagger`) v1.3.4 / Generador (`github.com/swaggo/swag`) v1.16.6
- Goconvey v1.8.1 para pruebas

## Estructura del proyecto
```
.
├── conf/             # Configuraciones de la app (app.conf, app.test.conf)
├── controllers/      # Controladores HTTP y tests
├── database/         # Inicialización y utilidades de base de datos
├── docs/             # Definiciones Swagger (generadas)
├── models/           # Modelos de dominio
├── routers/          # Rutas y namespaces
├── static/           # Activos estáticos
├── tests/            # Pruebas y utilidades (unitarias/integración)
├── views/            # Plantillas HTML
└── main.go           # Punto de entrada
```

## Requisitos
- Go 1.25+
- PostgreSQL disponible (local o remoto)
- Opcional: Bee CLI para hot-reload/documentación (`bee`)

## Configuración
- Los archivos de configuración están en `conf/`.
- Por defecto se usa `conf/app.conf`. Para pruebas se usa `conf/app.test.conf`.
- Puedes sobrescribir la ruta del archivo con la variable de entorno `BEEGO_APP_CONFIG_FILE`.
- Variables relevantes:
  - `SKIP_CRON=1`: desactiva el cron.
  - `CRON_ONE_SHOT=1`: ejecuta una sola iteración del cron.
  - `SKIP_WEB_RUN=1`: no levanta el servidor web (tests).

- Push/FCM (Android/iOS):
  - `FIREBASE_PROJECT_ID`: ID del proyecto Firebase para FCM HTTP v1.
  - `FCM_BEARER_TOKEN`: token Bearer OAuth2 para `https://fcm.googleapis.com/` (opcional). Si no se define, el backend intentará usar ADC:
    1) Metadata server (GCE/GKE) `service-accounts/default/token`.
    2) Si falla, se requiere inyectar `FCM_BEARER_TOKEN` manualmente.
  - El payload FCM se arma desde `models.ContenidoNotificacion` (`Titulo`→`notification.title`, `Mensaje`→`notification.body`) y `Datos`→`data`.

- Web Push (navegadores web):
  - `VAPID_PUBLIC_KEY`: Clave pública VAPID para autenticación Web Push (base64url sin padding).
  - `VAPID_PRIVATE_KEY`: Clave privada VAPID para autenticación Web Push (base64url sin padding).
  - `VAPID_SUBJECT`: Email o URL del contacto responsable (formato: `mailto:admin@example.com` o `https://example.com`).
  - El backend detecta automáticamente la plataforma (`WEB`, `ANDROID`, `IOS`) y usa el proveedor correcto:
    - Dispositivos `WEB` → Web Push Protocol (RFC 8030 + VAPID RFC 8292)
    - Dispositivos `ANDROID`/`IOS` → FCM HTTP v1
  - Manejo automático de errores:
    - **410 Gone** / **404 Not Found**: Desactiva automáticamente el dispositivo en BD (`enabled = false`)
    - **401 Unauthorized**: Error de autenticación VAPID (verificar claves)
    - **429 Too Many Requests**: Rate limit del servidor push
  - Timeout configurado: 10 segundos por notificación
  - Ver documentación completa: [`API_NOTIFICACIONES_CUPONES_OFERTAS.md`](API_NOTIFICACIONES_CUPONES_OFERTAS.md)

- `appname` en `conf/*.conf`: `el_fogon_de_maria`.

### CORS
- Define `CORS_ALLOWED_ORIGINS` con la lista de orígenes permitidos (separados por coma). Ejemplos:
  - Producción:
    ```
    CORS_ALLOWED_ORIGINS="https://lacocinademaria.netlify.app,https://elfogondemaria.netlify.app"
    ```
  - Staging/Desarrollo (incluyendo local):
    ```
    CORS_ALLOWED_ORIGINS="https://lacocinademaria.netlify.app,https://elfogondemaria.netlify.app,http://localhost:4200"
    ```

## Variables de entorno (DB y ejecución)
- DB (equivalentes a `conf/app.conf`): `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASS`, `DB_NAME`, `DB_SSLMODE`.
- App/Test: `BEEGO_APP_CONFIG_FILE`, `INTEGRATION`, `SKIP_DB_SEED`, `SKIP_WEB_RUN`, `SKIP_CRON`, `CRON_ONE_SHOT`, `CORS_ALLOWED_ORIGINS`.
- Auth: `JWT_SECRET` (obligatorio en prod; en dev/test se genera efímero si está vacío).
- Push Notifications:
  - FCM: `FIREBASE_PROJECT_ID`, `FCM_BEARER_TOKEN` (opcional, usa ADC si no está definido).
  - Web Push: `VAPID_PUBLIC_KEY`, `VAPID_PRIVATE_KEY`, `VAPID_SUBJECT` (obligatorios para notificaciones web).

## Enlaces rápidos
- Script de cobertura: [`tools/cover.ps1`](tools/cover.ps1)
- Definición de rutas: [`routers/router.go`](routers/router.go)

## Ejecución local
- Con Bee (opcional):
  ```powershell
  bee run -downdoc=true -gendoc=true
  ```
- Directo con Go:
  ```powershell
  go run main.go
  ```
- Swagger UI: `/swagger/` (base path `/restaurante/v1`).

## Pruebas
- Unitarias (no tocan DB real por defecto):
  ```powershell
  go test ./...
  ```
- Integración (requiere DB y `conf/app.test.conf`):
  ```powershell
  $env:INTEGRATION = "1"; go test ./...; Remove-Item Env:\INTEGRATION
  ```
- Cobertura recomendada:
  ```powershell
  powershell -ExecutionPolicy Bypass -File tools/cover.ps1 -Clean
  ```
  - Variables útiles antes de correr (si falla algún init en tests):
    ```powershell
    $env:JWT_SECRET = "testsecret"; $env:QUIET_TESTS = "1"
    ```
  - El reporte HTML queda en `coverage.html`.

### Notas para Windows (race/CGO y variables)
- `-race` requiere CGO habilitado. En Windows, si deseas correr `go test -race` localmente:
  ```powershell
  $env:CGO_ENABLED = "1"; go test -race ./...
  ```
  Si no necesitas el detector de carreras local, usa el script `tools/cover.ps1` (no activa `-race`). En CI sí se ejecuta con `-race`.
- Define un `JWT_SECRET` temporal para evitar fallos en tests que cargan el secreto:
  ```powershell
  $env:JWT_SECRET = "testsecret123"
  ```
- Para evitar levantar servidor o cron durante las pruebas unitarias:
  ```powershell
  $env:SKIP_WEB_RUN = "1"; $env:SKIP_CRON = "1"
  ```

### Notas de pruebas unitarias (inyecciones y expectativas)
- Nómina (`controllers/NominaController.go`): se expuso el punto de inyección `findExistingNominaFn` para simular la validación de existencia de nómina mensual en tests. Esto permite alcanzar el flujo de inserción sin depender de la consulta real.
  - Ejemplo en tests:
    ```go
    orig := findExistingNominaFn
    findExistingNominaFn = func(o orm.Ormer, fecha time.Time) (*models.Nomina, error) { return nil, orm.ErrNoRows }
    t.Cleanup(func(){ findExistingNominaFn = orig })
    ```
- ProductoPedido: en entorno unitario sin DB, los controladores devuelven HTTP 200 con mensajes de error en el cuerpo cuando falta parámetro, JSON es inválido, no hay stock o hay errores de consulta. Los tests validan esos mensajes (p.ej., "Inventario insuficiente...", "Error al buscar los detalles del pedido").
- Cron (generación de nómina): los tests validan eventos de slog (`cron.nomina.*`) con un handler en memoria, no por stdout.
- Confirmación: no se modificaron scripts SQL ni el esquema de la base de datos; sólo se ajustaron tests y puntos de inyección para pruebas.

## Logging
- Se usa `log/slog` con un helper `logging.LogControllerError` para registrar errores con contexto de negocio y sanitización.
- Endpoints críticos registran el body recibido ante errores para facilitar diagnóstico:
  - Domicilios (`POST`, `PUT`)
  - Pagos (`POST`, `PUT`)
  - Pedido (`POST`)
  - ProductoPedido (`POST`, `PUT`)
  - Reservas (`POST`, `PUT`)
- Sanitización automática: campos sensibles como `password`, `token` y textos muy largos son filtrados/truncados por el helper antes de emitirse.

## Formato y Lint
- Formatea con gofmt:
  ```powershell
  gofmt -s -w .
  ```
- Lint (si usas golangci-lint):
  ```powershell
  golangci-lint run
  ```
- Recomendado: pre-commit con gofmt + golangci-lint.

### Pre-commit (formato y linter automáticos)

Usamos hooks de [`pre-commit`](https://pre-commit.com) con [`pre-commit-hooks`](https://github.com/pre-commit/pre-commit-hooks) y `golangci-lint`.

Instalación rápida:

```bash
pip install pre-commit  # o brew/choco
pre-commit install      # instala los hooks declarados en .pre-commit-config.yaml
pre-commit run --all-files  # opcional, validar todo el repo
```

Qué hace:
- `gofmt -s -w .` aplica formato.
- `golangci-lint run --fix` intenta corregir y luego falla si quedan issues.
- Hooks utilitarios: `trailing-whitespace`, `end-of-file-fixer`, `check-merge-conflict`.

#### Detección de secretos con GitGuardian (ggshield)
- Añadido hook `ggshield-secret` para prevenir filtraciones de secretos.
- Instalación local:
  ```bash
  pip install ggshield  # o pipx/poetry
  ggshield auth login   # sigue las instrucciones para enlazar tu cuenta/ciudador
  ```
- Uso:
  - Automático en `pre-commit`.
  - Manual: `ggshield secret scan repo .` o `ggshield secret scan pre-commit`

## Swagger
- UI en `/swagger/`.
- Regenerar docs:
  ```powershell
  go install github.com/swaggo/swag/cmd/swag@latest
  swag init -g main.go -o docs
  ```
- Bypass de token en dev: las solicitudes iniciadas desde Swagger UI están permitidas sin token (solo modo `dev`).
 - En producción: `swagger = false` en `conf/app.prod.conf`. La ruta `/swagger/*` existe, pero se recomienda restringir acceso desde el reverse proxy si no deseas exponerla públicamente.

## Probes de salud
- `GET /healthz`: indica disponibilidad básica.
- `GET /readyz`: verifica la conexión a base de datos antes de reportar listo.

## Autenticación y rutas
- Autenticación: Bearer Token en cabecera `Authorization`.
- Base path: `/restaurante/v1`.
- Rutas principales (ver `routers/router.go` para el detalle):
  - Público: `POST /clientes` (registro), `GET /productos`, `GET /restaurantes`, `GET/POST /reservas`, `GET /cambios_horario/actual`, `POST /login`.
  - Protegido: clientes (GET/PUT/DELETE), productos (POST/PUT/DELETE), reservas (PUT/DELETE), pedidos, domicilios, trabajadores, horario_trabajador, métodos de pago, pagos, nóminas, incidencias, nomina_trabajador, producto_pedido, categorías, subcategorías.
  - Lectura auxiliar (protegido): `precio_producto_hist`, `control_nomina`, `restaurante_dia`, `reserva_contacto`.

## Comportamiento y reglas de negocio (clave)
- Cuerpos mínimos en Swagger: se usan DTOs en `models/Requests.go` para mostrar solo campos necesarios y sugerir formatos (fechas `YYYY-MM-DD`, horas `HH:MM:SS`).
- Tolerancia de casing en inputs: se prioriza camelCase pero en algunos endpoints se aceptan claves UPPERCASE (p. ej. `PUT /pagos`).
- Respuestas vacías: cuando no hay resultados, se retorna 200 con `data: []` y mensaje claro.
- Productos (imagen en POST/PUT):
  - JSON: campo `imagen` como Base64.
  - `multipart/form-data`: adjuntar archivo en campo `imagen` (recomendado desde Swagger).
- Historial de precios (`precio_producto_hist`):
  - `POST /productos` y cambio de `precio` en `PUT /productos` insertan una vigencia (secuencias alineadas automáticamente).
  - Endpoints devuelven sólo: `nombre`, `estadoProducto`, `precio`, `fechaVigencia`.
- Pedido y detalles (`producto_pedido`):
  - BD fija `precio` por trigger según vigencia.
  - Inventario: al crear/actualizar detalles, se descuenta/devolver stock en transacción.
  - Sin stock suficiente: 400 con detalle por `productoId` (`requerido`/`disponible`).
- Domicilios: `delivery=false` exige `pk_id_domicilio = NULL`; asignar domicilio marca `delivery=true` automáticamente.
- Filtros corregidos: `domicilios` (estado, updated_by, trabajador), `pedidos` (año sin mes), `trabajadores` (fecha ingreso exacta), etc.
- Nómina (`nominas`):
  - Validaciones en `POST`:
    - No crear antes del día 20 del mes.
    - Si ya existe una nómina en ese mes: no se crea otra; se marca `control_nomina` como `REGENERADA` y se devuelve 200 con la nómina existente.
  - `PUT /nominas?id=...` cambia estado a `PAGO`.
  - `DELETE /nominas?id=...` marca `NO_PAGO` (lógico).
  - `nomina_trabajador` (POST) idempotente: si existe relación para la última nómina, devuelve 200 con existente. La descripción siempre proviene de DB.
- `restaurante_dia`: los endpoints de lectura devuelven `restauranteId`, `nombreRestaurante`, `horaApertura` (`HH:MM:SS`) y `dia`.

## Ejemplos útiles
- Subir imagen de producto con archivo (PUT):
  ```bash
  curl -X PUT 'http://localhost:8080/restaurante/v1/productos?id=1' \
    -H 'Content-Type: multipart/form-data' \
    -F 'nombre=Pizza' -F 'precio=18000' -F 'estadoProducto=DISPONIBLE' \
    -F 'cantidad=10' -F 'subcategoriaId=3' -F 'imagen=@C:/ruta/imagen.jpg'
  ```
- Historial de precios por producto:
  ```bash
  curl 'http://localhost:8080/restaurante/v1/precio_producto_hist?producto_id=2'
  # -> nombre, estadoProducto, precio, fechaVigencia
  ```
- Días de servicio y hora de apertura:
  ```bash
  curl 'http://localhost:8080/restaurante/v1/restaurante_dia?restaurante_id=1&dia=Lunes'
  # -> restauranteId, nombreRestaurante, horaApertura (HH:MM:SS), dia
  ```

## Despliegue
Compilación y ejecución:
```powershell
# Build optimizado
go build -trimpath -ldflags "-s -w" -o restaurante-back.exe .
./restaurante-back.exe
```
En Linux/macOS, cambia el nombre del binario si lo deseas.

## Cómo contribuir (PRs)
1. Rama desde `develop`.
2. Commits pequeños y descriptivos en español (modo imperativo).
3. PR como Draft y referencia issues con `#id`.
4. Pasa gofmt, lint y pruebas. Luego marca como Ready.

## Documentación adicional
Consulta el directorio `docs/` y la UI en `/swagger/`.

## Licencia
© 2025 ElLuchoMan
