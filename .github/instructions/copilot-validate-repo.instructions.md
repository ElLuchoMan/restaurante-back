# Instrucciones para que Copilot valide el repositorio

Breve plan: crearé una lista de verificación y pasos concretos (comandos y criterios) que Copilot debe seguir para validar compilación, tests, linter, modelos de ORM y convenciones del proyecto.

## Checklist (requisitos explícitos)
- [ ] Compilar sin errores (`go build ./...`).
- [ ] Ejecutar tests y mostrar resultados (`go test ./...`).
- [ ] Ejecutar verificadores estáticos: `go vet`, `golangci-lint` (si está disponible).
- [ ] Ejecutar registro/arranque del ORM (Beego) en modo test o build y verificar que no produce panic por registros de modelos.
- [ ] Verificar que los modelos de ORM usan tipos y tags correctos (relaciones `rel(fk)`, `type(...)`, `null` cuando corresponde).
- [ ] Validar serialización JSON personalizada (MarshalJSON/UnmarshalJSON) contra los tipos definidos.
- [ ] Revisar generación de documentación Swagger (`swag`) si existe; comprobar que `docs` está consistente.
- [ ] Ejecutar `go mod tidy` y comprobar que `go.sum` y `go.mod` están sanos (opcional, previo a PR).
- [ ] Reportar cualquier cambio recomendado como parches aplicables con el menor alcance posible.

## Criterios de éxito
- Build OK (salida 0) y tests con estado `ok` o fallos documentados con pasos para reproducir.
- No panic en bootstrap de Beego ORM.
- Linter: cero issues nuevos de alta severidad; enumerar y justificar los restantes si existen.

## Pasos concretos (PowerShell)
Usa estos comandos en el workspace raíz del repositorio (Windows PowerShell):

```powershell
# compilar
go build ./...
# ejecutar tests (mostrará paquetes y fallos)
go test ./... | Tee-Object -Variable testOut
# ver salida resumida
$LASTEXITCODE
# vet estático
go vet ./...
# opcional: golangci-lint si está instalado
if (Get-Command golangci-lint -ErrorAction SilentlyContinue) { golangci-lint run } else { Write-Output "golangci-lint no instalado, se omite" }
# ordenar módulos y verificar
go mod tidy
# probar build de la app (arranque rápido, detener inmediatamente si inicia servidor)
# exportar variable para evitar efectos secundarios si el código depende de env vars
$env:SKIP_ORM_REGISTRATION = "1"; go build ./...; Remove-Item Env:\SKIP_ORM_REGISTRATION
```

Notas:
- Para validar el bootstrap de Beego ORM se recomienda ejecutar en un entorno controlado (tests o con `SKIP_ORM_REGISTRATION` desactivado según convenga) y observar la pila de panic.
- Si el proyecto arranca un servidor en `init()` (como ocurre en este repo), lanzar `go build` y revisar `main.init` y `database.InitDB()` para no ejecutar semillas peligrosas en un entorno de CI.

## Comprobaciones específicas a realizar en código
- Buscar `orm.RegisterModel` y revisar orden de registro; si una relación declara `rel(fk)` hacia un tipo, ese tipo debe estar registrado o la propiedad debe ser del tipo estructural (p. ej. `*Subcategoria`) con tag `rel(fk)`.
- Revisar campos de imagen/bytes: para PostgreSQL usar `[]byte` con tag `type(bytea)` y en JSON manejar base64 correctamente.
- Para relaciones simples: preferir `*TipoRelacionado` con `orm:"column(...);rel(fk)"` en vez de usar `int64` a menos que el proyecto registre y trate explícitamente los PK como scalars; en ese caso validar orden de `init()`.
- Comprobar que los métodos `MarshalJSON`/`UnmarshalJSON` mantienen coherencia entre los tipos JSON y los campos del struct (p. ej. evitar asignar `string` a `[]byte` sin conversión/base64).

## Resultado esperado del informe
Al terminar la validación, producir un informe corto (máx 1 página) que incluya:
- Resumen (Build OK / Tests OK / Lint issues N).
- Lista de fallos bloqueantes con archivos y líneas relevantes.
- Parches sugeridos (diffs) para cada fallo bloqueante o para mejoras de alta prioridad.
- Comandos exactos para reproducir localmente los problemas.

## Política para aplicar cambios automáticos
- Aplicar solo cambios no ambigüos: corrección de tags `orm`, cambio de tipo de campo para relaciones a `*Tipo`, y ajustes triviales en Marshal/Unmarshal que preserven contract JSON.
- Para cambios de mayor alcance (reorganizar init() o cambiar diseño de modelos), crear un PR con descripción y no aplicar automáticamente.

## Checks rápidos que Copilot debe ejecutar primero
1. `go build ./...`
2. `go test ./...`
3. `go vet ./...`
4. Buscar `panic` o stack traces en arranque (especialmente relacionados con `orm.RegisterModel` o `bootstrap`).

---
Archivo creado para uso de Copilot: `.github/instructions/copilot-validate-repo.instructions.md`

Si quieres, aplico automáticamente los primeros parches seguros (por ejemplo: cambiar `PK_ID_SUBCATEGORIA` de `int64` a `*Subcategoria`, o `IMAGEN` a `[]byte` con `type(bytea)`) y ejecuto build/tests para verificar; dime si proceda con parches automáticos o solo reporte.
