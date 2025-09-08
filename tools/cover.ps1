# Ejecuta cobertura de todo el repo (Go), reintenta por paquete si es necesario,
# normaliza nombre del perfil y genera coverage.html + coverage-YYYYMMDD-HHMMSS.html (cache-busting).

param(
  [switch]$Clean
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

# Salida UTF-8 (acentos bien)
try { [Console]::OutputEncoding = [System.Text.UTF8Encoding]::new($false) } catch { }

# Ir a la raíz del repo si el script vive en tools/
try {
  $scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
  $repoRoot = Resolve-Path (Join-Path $scriptDir "..")
  Set-Location $repoRoot
}
catch { }

if ($Clean) {
  Remove-Item -Force coverage.out, coverage.html, tmp.out -ErrorAction SilentlyContinue
  Get-ChildItem -Filter 'coverage-*.html' | Remove-Item -Force -ErrorAction SilentlyContinue
}

# 1) Primer intento: perfil directo (todo el módulo)
& go test -count=1 -covermode=atomic -coverpkg=./... -coverprofile=coverage.out ./...

# 2) Normalizar nombre si vino como "coverage" sin extensión
if (!(Test-Path .\coverage.out) -and (Test-Path .\coverage)) {
  Rename-Item .\coverage coverage.out
}

# 3) Si no existe o está "vacío" (solo 'mode: atomic'), combinar por paquete
$needsCombine = $true
if (Test-Path .\coverage.out) {
  $len = (Get-Item .\coverage.out).Length
  if ($len -ge 50) { $needsCombine = $false }
}

if ($needsCombine) {
  Remove-Item -Force tmp.out, coverage.out -ErrorAction SilentlyContinue
  $pkgs = & go list ./...
  if (-not $pkgs) { Write-Error "No se encontraron paquetes"; exit 1 }

  $first = $true
  foreach ($p in $pkgs) {
    & go test -count=1 -covermode=atomic -coverpkg=./... -coverprofile=tmp.out $p | Out-Null
    if (Test-Path .\tmp.out) {
      if ($first) {
        Move-Item .\tmp.out .\coverage.out
        $first = $false
      }
      else {
        Get-Content .\tmp.out | Select-Object -Skip 1 | Add-Content .\coverage.out
        Remove-Item .\tmp.out
      }
    }
  }
}

if (!(Test-Path .\coverage.out)) {
  Write-Error "No se generó coverage.out"
  exit 1
}

# 4) Generar HTML único (cover.html) y sobrescribir si existe
$profile = (Resolve-Path .\coverage.out).Path
$outHtml = (Join-Path (Resolve-Path .).Path 'coverage.html')

Remove-Item -Force "$outHtml" -ErrorAction SilentlyContinue

# Evitar sintaxis con '=' que a veces falla en PowerShell
& (Get-Command go).Source tool cover -html "$profile" -o "$outHtml"

# 5) Resumen en consola
Write-Host "`nResumen por función:"
& (Get-Command go).Source tool cover -func "$profile"

# 6) Abrir en navegador el cover.html estable
$fileUrl = "file:///$outHtml"
Start-Process "$fileUrl"

Write-Host "`nOK -> $outHtml"
