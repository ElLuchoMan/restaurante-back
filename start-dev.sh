#!/bin/bash
# Script para iniciar el servidor en desarrollo con JWT_SECRET configurado
# Uso:
#   ./start-dev.sh                           # Normal
#   ./start-dev.sh -downdoc=true -gendoc=true  # Con generación de docs

echo "🚀 Iniciando servidor de desarrollo..."

# Cargar .env si existe
if [ -f .env ]; then
    echo "📄 Cargando variables desde .env..."
    export $(cat .env | grep -v '^#' | xargs)
    echo "   ✓ Variables cargadas"
fi

# Verificar JWT_SECRET
if [ -z "$JWT_SECRET" ]; then
    echo "⚠️  JWT_SECRET no configurado. Cárgalo desde .env o variable de entorno del sistema"
else
    echo "✅ JWT_SECRET configurado correctamente"
fi

echo "📝 Nota: El secret es consistente, tu sesión NO se cerrará en reinicios"
echo ""

# Ejecutar bee run con todos los argumentos pasados
if [ $# -eq 0 ]; then
    echo "▶️  Ejecutando: bee run"
    bee run
else
    echo "▶️  Ejecutando: bee run $@"
    bee run "$@"
fi
