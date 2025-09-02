# Restaurant Backend

[![AutoGen with AI](https://img.shields.io/badge/AutoGen%20with%20AI-blue)](#)

## Description
Go-based REST API for managing restaurant operations such as customers, orders, payments, and scheduling.

## Technologies
- Go 1.25
- Beego v2.3.8
- PostgreSQL driver (`github.com/lib/pq`) v1.10.9
- JWT (`github.com/dgrijalva/jwt-go`) v3.2.0
- Swaggo (`github.com/swaggo/http-swagger`) v1.3.4 / `github.com/swaggo/swag` v1.16.6
- goconvey v1.8.1 for testing

## Project Structure
```
.
├── conf/             # application configuration files
├── controllers/      # HTTP controllers and tests
├── database/         # database initialization and helpers
├── docs/             # Swagger specifications
├── models/           # data models
├── routers/          # route definitions
├── static/           # static assets
├── tests/            # integration tests and utilities
├── views/            # HTML templates
└── main.go           # application entry point
```

## Installation
1. Ensure Go 1.25 is installed.
2. Clone the repository and enter its directory.
3. Download dependencies:
   ```sh
   go mod download
   ```
4. Configuration files reside in `conf/` (`app.conf`, `app.test.conf`).

## Development
- Start the development server with live reload and Swagger generation:
  ```sh
  bee run -downdoc=true -gendoc=true
  ```
- Alternatively, run the app directly:
  ```sh
  go run main.go
  ```

## Enumerated Types and Payroll
- **States** are modeled with enums under `models/enums.go`:
  - `EstadoReserva`: `PENDIENTE`, `CONFIRMADA`, `CANCELADA`, `CUMPLIDA`.
  - `EstadoNomina`: `PAGO`, `NO_PAGO`.
  - Additional enums exist for domicilios, pagos, pedidos, productos y días de la semana.

- **Price history** for products is stored in the `precio_producto_hist` table. A new entry is
  created when a product is registered or its price changes, storing the effective date of the
  change.
- `POST /producto_pedido` and `PUT /producto_pedido` no longer accept `precio` in the payload; the database trigger assigns it automatically.
- **Payroll (nómina)** operations:
- `POST /nominas` accepts optional `generar_nomina_automatica` and
  `verificar_nomina` flags to auto-generate payments (passing the payroll
  date as `p_fecha`) and verify payroll records without requiring an ID.
- `PUT /nominas?id=...` transitions a payroll to `PAGO`.
- `DELETE /nominas?id=...` performs a logical deletion setting the state to `NO_PAGO`.

## Database Tables
- **nomina**: stores payroll records including amount, state, and audit timestamps.
- **reserva**: manages restaurant reservations with date, time, party size, and status.
- **precio_producto_hist**: tracks product price changes over time with an effective date for each entry.
- **cambios_horario**: logs schedule adjustments such as opening/closing hours and whether the restaurant opens.
- **domicilio**: la columna `entregado` es generada automáticamente por la base de datos y no debe incluirse en las solicitudes de creación o actualización.
- **pedido**: la base de datos valida que `delivery` sea `false` o que `pk_id_domicilio` tenga un valor.

## Testing
```sh
go test ./...
```

## Testing with coverage file
```sh
powershell -ExecutionPolicy Bypass -File tools/cover.ps1 -Clean
```

## Production
1. Build the binary:
   ```sh
   go build -o restaurante-back
   ```
2. Run the server:
   ```sh
   ./restaurante-back
   ```

## How to make a PR
1. Fork the repository and create a feature branch.
2. Commit your changes and push the branch.
3. Open a pull request via [GitHub](https://github.com/ElLuchoMan/restaurante-back/pulls) in **Draft** state.
4. Address feedback, then mark the PR as **Ready for Review**.

## API Documentation
- Open `docs/swagger.yaml` or serve the project and navigate to `/swagger/` for Swagger UI.

## Additional Documentation
See the [`docs/` directory](docs/) for extended documentation.

## License
© 2025 ElLuchoMan
