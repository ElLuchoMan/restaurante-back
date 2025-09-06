# Restaurante - Backend (Beego v2 + PostgreSQL)

## Descripción
API REST en Go para gestionar operaciones de un restaurante: clientes, pedidos, pagos, productos, reservas, nómina y más. Basada en Beego v2 con documentación Swagger.

## Inicio rápido
- Levantar en desarrollo (con Swagger auto-generado):
  ```powershell
  bee run -downdoc=true -gendoc=true
  ```
- Ejecutar pruebas con cobertura (genera coverage.out y coverage.html):
  ```powershell
  powershell -ExecutionPolicy Bypass -File tools/cover.ps1 -Clean
  ```

## Tecnologías
- Go 1.25
- Beego v2.3.8
- PostgreSQL (`github.com/lib/pq`) v1.10.9
- JWT (`github.com/dgrijalva/jwt-go`) v3.2.0
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
- Por defecto se usa `conf/app.conf`. Para pruebas se usa `conf/app.test.conf` (ver `tests/setup_test.go`).
- Puedes sobrescribir la ruta del archivo con la variable de entorno `BEEGO_APP_CONFIG_FILE`.
- Variables relevantes en tiempo de ejecución:
  - `SKIP_CRON=1`: desactiva el cron de nómina automática.
  - `CRON_ONE_SHOT=1`: ejecuta una sola iteración del cron (para pruebas).
  - `SKIP_WEB_RUN=1`: evita levantar el servidor web (útil en tests unitarios).

## Variables de entorno (DB y ejecución)
- Base de datos (equivalentes a `conf/app.conf` por defecto):
  - `db_host` (default: `localhost`)
  - `db_port` (default: `5432`)
  - `db_user` (default: `postgres`)
  - `db_pass` (default: `12346`)
  - `db_name` (default: `restaurante_db`)
- App/Tests:
  - `BEEGO_APP_CONFIG_FILE`: ruta a `app.conf` alternativo.
  - `INTEGRATION=1`: habilita pruebas de integración y el seed de datos de prueba.
  - `SKIP_DB_SEED=1`: omite seed de datos (se fuerza a `1` en modo no integración).
  - `SKIP_WEB_RUN=1`, `SKIP_CRON=1`, `CRON_ONE_SHOT=1`: ver arriba.

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
- Swagger UI estará disponible en `/swagger/` con base path `/restaurante/v1`.

## Pruebas
- Unitarias (por defecto no tocan DB real):
  ```powershell
  go test ./...
  ```
- Integración (requiere DB y `conf/app.test.conf`):
  ```powershell
  $env:INTEGRATION = "1"; go test ./...; Remove-Item Env:\INTEGRATION
  ```

### Cobertura (recomendada)
Usa el script de `tools/cover.ps1` para ejecutar todas las pruebas con `-covermode=atomic`, combinar perfiles por paquete si hace falta y generar `coverage.out` + `coverage.html`:
```powershell
powershell -ExecutionPolicy Bypass -File tools/cover.ps1 -Clean
```

## Formato y Lint
- Formatea con gofmt:
  ```powershell
  gofmt -s -w .
  ```
- Lint (si usas golangci-lint):
  ```powershell
  golangci-lint run
  ```
- Recomendado: configurar pre-commit para ejecutar gofmt y golangci-lint antes de hacer commit.

## Swagger
- La documentación se sirve en `/swagger/` cuando la app está corriendo.
- Para regenerar los archivos en `docs/` (si cambias los comentarios Swagger):
  ```powershell
  go install github.com/swaggo/swag/cmd/swag@latest
  swag init -g main.go -o docs
  ```

## Autenticación y rutas
- Autenticación vía Bearer Token en cabecera `Authorization` (ver `@securityDefinitions.apikey BearerAuth`).
- Base path: `/restaurante/v1`.
- Rutas destacadas (ver `routers/router.go` para el detalle):
  - Público: `POST /clientes` (registro), `POST /login`, `GET/POST/PUT/DELETE /productos`, `GET/POST/PUT/DELETE /reservas`.
  - Protegido (requiere token): clientes (GET/PUT/DELETE), pedidos, domicilios, trabajadores, horario_trabajador, métodos de pago, pagos, nóminas, incidencias, nomina_trabajador, producto_pedido.

## Notas de dominio
- Enums en `models/enums.go` (reservas, nómina, domicilios, pagos, pedidos, productos, días de semana).
- Historial de precios: tabla `precio_producto_hist`. Se crea registro al crear producto o cambiar precio con fecha de vigencia.
- `POST/PUT /producto_pedido`: ya no se envía `precio`; la base de datos lo asigna vía trigger.
- Nómina:
  - `POST /nominas` acepta `generar_nomina_automatica` y `verificar_nomina` (usa `p_fecha`).
  - `PUT /nominas?id=...` cambia estado a `PAGO`.
  - `DELETE /nominas?id=...` realiza borrado lógico (`NO_PAGO`).
- Validaciones a nivel de BD:
  - `domicilio.entregado` es generado automáticamente (no incluir en payloads).
  - En `pedido`: `delivery=false` o `pk_id_domicilio` debe tener valor.
- Esquema de BD mantenido fuera del repo (no hay migraciones aquí).

## Despliegue
Compilación y ejecución:
```powershell
go build -o restaurante-back
./restaurante-back
```
En Windows el binario será `restaurante-back.exe`.

## Cómo contribuir (PRs)
1. Crea una rama a partir de `develop`.
2. Haz commits pequeños y descriptivos en español (modo imperativo).
3. Abre un Pull Request como Draft en GitHub y referencia issues con `#id`.
4. Asegura pasar gofmt, lint y pruebas. Luego marca como Ready for Review.

## Documentación adicional
Consulta el directorio `docs/` y la UI en `/swagger/`.

## Licencia
© 2025 ElLuchoMan
