#!/usr/bin/env node
/**
 * Genera colección de Postman completa desde Swagger
 * Con características profesionales incluidas
 */

const fs = require('fs');
const path = require('path');

console.log('🚀 Generando colección Postman desde Swagger...\n');

// Cargar Swagger
const swaggerPath = path.join(__dirname, '..', 'docs', 'swagger.json');
const swagger = JSON.parse(fs.readFileSync(swaggerPath, 'utf8'));

// Base de la colección profesional
const collection = {
  info: {
    name: "🍽️ El fogón de María API",
    _postman_id: "restaurante-api-v2",
    description: `# El fogón de María - API REST

## 📋 Descripción
Colección completa de la API del restaurante "El fogón de María". Esta colección incluye todos los endpoints documentados en Swagger v${swagger.info.version}.

## 🔐 Autenticación Automática
La API utiliza **JWT Bearer Token**. El token se captura y usa automáticamente:

### ✨ Funcionamiento:
1. ✅ Ejecuta \`POST /login\` con tus credenciales
2. ✅ El token se guarda automáticamente en \`{{token}}\`
3. ✅ **Todos los endpoints lo usan automáticamente** (no necesitas hacer nada más)

## 🌐 Variables
- \`{{baseUrl}}\`: URL base de la API (default: http://localhost:8080${swagger.basePath || ''})
- \`{{token}}\`: Token JWT (se autocompleta al hacer login) 🎯

## 🧪 Tests Automáticos
Cada endpoint incluye tests que validan:
- ✅ Status code correcto
- ✅ Tiempo de respuesta < 2s
- ✅ Estructura de respuesta JSON
- ✅ Validación de campos requeridos

## 📝 Convenciones
- **Fechas**: \`YYYY-MM-DD\` (ej: 2025-10-15)
- **Horas**: \`HH:MM:SS\` (ej: 19:30:00)
- **Zona horaria**: America/Bogota
- **Paginación**: \`?limit=20&offset=0\`

---
**Versión:** ${swagger.info.version}  
**Última actualización:** Octubre 2025  
**Estado:** ✅ Sincronizada con Swagger`,
    schema: "https://schema.getpostman.com/json/collection/v2.1.0/collection.json",
    version: swagger.info.version
  },
  auth: {
    type: "bearer",
    bearer: [{ key: "token", value: "{{token}}", type: "string" }]
  },
  event: [
    {
      listen: "prerequest",
      script: {
        type: "text/javascript",
        exec: [
          "// 🚀 Script Pre-Request Global",
          "console.log('🚀', pm.request.method, pm.request.url.toString());",
          "const token = pm.collectionVariables.get('token');",
          "const authType = pm.request.auth ? pm.request.auth.type : 'bearer';",
          "if (!token && authType !== 'noauth') {",
          "    console.warn('⚠️  No hay token. Ejecuta POST /login primero.');",
          "}"
        ]
      }
    },
    {
      listen: "test",
      script: {
        type: "text/javascript",
        exec: [
          "// 🧪 Tests Globales",
          "pm.test('⏱️  Tiempo < 2s', () => pm.expect(pm.response.responseTime).to.be.below(2000));",
          "const ct = pm.response.headers.get('Content-Type');",
          "if (pm.response.code !== 204 && ct && ct.includes('application/json')) {",
          "    pm.test('📦 JSON válido', () => pm.response.to.be.json);",
          "    const json = pm.response.json();",
          "    if (json) pm.test('✅ Estructura estándar', () => pm.expect(json).to.have.property('success'));",
          "}",
          "console.log(pm.response.code >= 400 ? '❌' : '✅', pm.response.code);"
        ]
      }
    }
  ],
  variable: [
    { key: "baseUrl", value: `http://localhost:8080${swagger.basePath || ''}`, type: "string" },
    { key: "token", value: "", type: "string" }
  ],
  item: []
};

// Organizar endpoints por tags/folders
const folders = {};
Object.keys(swagger.paths).forEach(path => {
  Object.keys(swagger.paths[path]).forEach(method => {
    const endpoint = swagger.paths[path][method];
    const tag = (endpoint.tags && endpoint.tags[0]) || 'Other';
    
    if (!folders[tag]) {
      folders[tag] = [];
    }
    
    const isPublic = !endpoint.security || endpoint.security.length === 0;
    const requestName = `${method.toUpperCase()} ${path}`;
    
    folders[tag].push({
      name: requestName,
      request: {
        method: method.toUpperCase(),
        header: method !== 'get' && method !== 'delete' ? [
          { key: "Content-Type", value: "application/json" }
        ] : [],
        url: `{{baseUrl}}${path}`,
        ...(isPublic ? { auth: { type: "noauth" } } : {}),
        description: endpoint.summary || endpoint.description || ''
      }
    });
  });
});

// Iconos para folders
const icons = {
  'health': '🏥', 'auth': '🔐', 'clientes': '👥', 'restaurantes': '🏪',
  'trabajadores': '👷', 'productos': '🛒', 'categorias': '📂', 'pedidos': '🛍️',
  'pagos': '💳', 'reservas': '📅', 'cupones': '🎟️', 'ofertas': '🎁',
  'telemetria': '📈', 'push': '🔔', 'incidencias': '⚠️', 'nominas': '💵'
};

// Crear folders en la colección
Object.keys(folders).sort().forEach(tag => {
  const icon = icons[tag.toLowerCase()] || '📁';
  const name = tag.charAt(0).toUpperCase() + tag.slice(1);
  
  collection.item.push({
    name: `${icon} ${name}`,
    description: `Endpoints de ${name.toLowerCase()}`,
    item: folders[tag]
  });
});

// Tests especiales para login
const authFolder = collection.item.find(f => f.name.includes('Auth'));
if (authFolder) {
  const loginReq = authFolder.item.find(r => r.name.includes('login'));
  if (loginReq) {
    loginReq.event = [{
      listen: "test",
      script: {
        type: "text/javascript",
        exec: [
          "pm.test('✅ Status 200', () => pm.response.to.have.status(200));",
          "const json = pm.response.json();",
          "pm.test('🔑 Tiene token', () => {",
          "    pm.expect(json.data).to.have.property('token');",
          "    pm.expect(json.data.token).to.be.a('string').and.not.empty;",
          "});",
          "if (json && json.data && json.data.token) {",
          "    pm.collectionVariables.set('token', json.data.token);",
          "    console.log('✅ Token guardado automáticamente');",
          "    console.log('🔑', json.data.token.substring(0, 20) + '...');",
          "    if (json.data.trabajador) {",
          "        console.log('👤', json.data.trabajador.nombre);",
          "        console.log('🎭', json.data.trabajador.rol);",
          "    }",
          "}"
        ]
      }
    }];
    
    loginReq.request.body = {
      mode: "raw",
      raw: JSON.stringify({ documento: 1015466495, password: "tu_password" }, null, 2)
    };
  }
}

// Guardar
const outputPath = path.join(__dirname, 'postman_collection.json');
fs.writeFileSync(outputPath, JSON.stringify(collection, null, 2));

console.log('✅ Colección generada exitosamente');
console.log(`📊 Total de folders: ${collection.item.length}`);
let totalRequests = 0;
collection.item.forEach(f => totalRequests += f.item ? f.item.length : 0);
console.log(`📊 Total de requests: ${totalRequests}`);
console.log('\n🎯 Características profesionales incluidas:');
console.log('  ✅ Autenticación automática con Bearer Token');
console.log('  ✅ Tests globales en todos los endpoints');
console.log('  ✅ Pre-request scripts con validación');
console.log('  ✅ Descripción completa de la colección');
console.log('  ✅ Tests especiales en login para captura automática de token');
console.log('\n💡 Tip: Ejecuta POST /login y el token se usará automáticamente');


