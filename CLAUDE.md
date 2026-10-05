# Reglas del proyecto (restaurante-back)

## Tests y cobertura
- **Todo cambio debe incluir sus tests** en el mismo PR (código nuevo o modificado → tests nuevos o actualizados).
- La cobertura de statements es **100 %**: el CI (`.github/workflows/ci.yml`) falla si el total o cualquier función queda por debajo; Codecov exige 100 % en proyecto y parche. Si 100 % deja de ser sostenible, el mínimo aceptado es 99 %, pero siempre se documentan y entregan tests.
- CI corre sin build tags: `go test -race -count=1 -covermode=atomic -coverpkg=./... ./...` con `SKIP_WEB_RUN=1 SKIP_CRON=1 SKIP_DB_SEED=1`.
- Código inalcanzable: eliminarlo; no usar exclusiones de cobertura.
- ORM sin base real: driver SQL falso registrado en `TestMain` (patrón en `controllers/cupon/orm_mock_driver_test.go`).
- Antes de subir: `gofmt -l .` vacío, `go vet ./...`, `golangci-lint run` y los tests anteriores.

## Flujo
- Trabajar en `develop`; `master` solo vía PR (push directo rechazado). Tras cada merge a master, el workflow `sync-develop` hace backport a develop.
