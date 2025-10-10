# Script para iniciar el servidor en desarrollo con JWT_SECRET configurado
# Uso:
#   .\start-dev.ps1                                    # Normal
#   .\start-dev.ps1 -downdoc=true -gendoc=true        # Con generacion de docs

Write-Host "Iniciando servidor de desarrollo..." -ForegroundColor Green

# Cargar .env si existe
if (Test-Path .env) {
    Write-Host "Cargando variables desde .env..." -ForegroundColor Cyan
    Get-Content .env | ForEach-Object {
        $line = $_.Trim()
        if ($line -and !$line.StartsWith("#")) {
            $parts = $line.Split("=", 2)
            if ($parts.Count -eq 2) {
                $key = $parts[0].Trim()
                $value = $parts[1].Trim()
                Set-Item -Path "env:$key" -Value $value
                Write-Host "   Configurado: $key" -ForegroundColor Gray
            }
        }
    }
}

# Verificar JWT_SECRET
if ([string]::IsNullOrEmpty($env:JWT_SECRET)) {
    Write-Host "JWT_SECRET no configurado. Cárgalo desde .env o variable de entorno del sistema." -ForegroundColor Yellow
} else {
    Write-Host "JWT_SECRET configurado correctamente" -ForegroundColor Green
}

Write-Host "Nota: El secret es consistente, tu sesion NO se cerrara en reinicios" -ForegroundColor Cyan
Write-Host ""

# Construir comando bee run con argumentos adicionales
$beeArgs = $args -join " "
if ([string]::IsNullOrEmpty($beeArgs)) {
    Write-Host "Ejecutando: bee run" -ForegroundColor Magenta
    bee run
} else {
    Write-Host "Ejecutando: bee run $beeArgs" -ForegroundColor Magenta
    Invoke-Expression "bee run $beeArgs"
}
