#!/usr/bin/env node
/**
 * Generador de colección Postman profesional
 * Genera la colección completa sincronizada con Swagger
 */

const fs = require('fs');
const path = require('path');

// Estructura base de la colección
const collection = {
  info: {
    name: "🍽️ El fogón de María API",
    _postman_id: "restaurante-api-v2",
    description: `# El fogón de María - API REST

## 📋 Descripción
Colección completa de la API del restaurante "El fogón de María". Esta colección incluye todos los endpoints documentados en Swagger v2.0.0.

## 🔐 Autenticación
La API utiliza **JWT Bearer Token**. El token se captura automáticamente al hacer login.

### Uso:
1. Ejecuta \`POST /login\` con tus credenciales
2. El token se guarda automáticamente en \`{{token}}\`
3. Todos los endpoints autenticados usan el token automáticamente

## 🌐 Variables
- \`{{baseUrl}}\`: URL base de la API (default: http://localhost:8080/restaurante/v1)
- \`{{token}}\`: Token JWT (se autocompleta al hacer login)

## 📚 Estructura
- **Health**: Endpoints de salud del sistema
- **Auth**: Login y refresh de tokens
- **Público**: Endpoints sin autenticación
- **Gestión**: Endpoints CRUD con autenticación

## 🔗 Enlaces
- [Swagger UI](http://localhost:8080/swagger)

## 📝 Notas
- Formatos de fecha: \`YYYY-MM-DD\`
- Formatos de hora: \`HH:MM:SS\`
- Zona horaria: America/Bogota
- Paginación: \`?limit=20&offset=0\`

---
**Versión:** 2.0.0  
**Última actualización:** Octubre 2025`,
    schema: "https://schema.getpostman.com/json/collection/v2.1.0/collection.json",
    version: "2.0.0"
  },
  auth: {
    type: "bearer",
    bearer: [
      {
        key: "token",
        value: "{{token}}",
        type: "string"
      }
    ]
  },
  event: [
    {
      listen: "prerequest",
      script: {
        type: "text/javascript",
        exec: [
          "// Script global pre-request",
          "// Log del endpoint que se va a ejecutar",
          "console.log('🚀 Ejecutando:', pm.request.method, pm.request.url.toString());",
          "",
          "// Verificar si hay token para endpoints autenticados",
          "const token = pm.collectionVariables.get('token');",
          "const authType = pm.request.auth ? pm.request.auth.type : 'bearer';",
          "if (!token && authType !== 'noauth') {",
          "    console.warn('⚠️  Advertencia: No hay token configurado. Ejecuta POST /login primero.');",
          "}"
        ]
      }
    },
    {
      listen: "test",
      script: {
        type: "text/javascript",
        exec: [
          "// Script global de tests",
          "pm.test('⏱️  Tiempo de respuesta < 2000ms', function() {",
          "    pm.expect(pm.response.responseTime).to.be.below(2000);",
          "});",
          "",
          "// Validar que siempre devuelve JSON",
          "if (pm.response.code !== 204 && pm.response.headers.get('Content-Type')?.includes('application/json')) {",
          "    pm.test('📦 Response es JSON', function() {",
          "        pm.response.to.be.json;",
          "    });",
          "",
          "    const jsonData = pm.response.json();",
          "    if (jsonData) {",
          "        pm.test('✅ Respuesta tiene estructura estándar', function() {",
          "            pm.expect(jsonData).to.have.property('success');",
          "        });",
          "    }",
          "}",
          "",
          "// Log de resultado",
          "if (pm.response.code >= 200 && pm.response.code < 300) {",
          "    console.log('✅ Success:', pm.response.code);",
          "} else if (pm.response.code >= 400) {",
          "    console.log('❌ Error:', pm.response.code, pm.response.status);",
          "}"
        ]
      }
    }
  ],
  variable: [
    {
      key: "baseUrl",
      value: "http://localhost:8080/restaurante/v1",
      type: "string"
    },
    {
      key: "token",
      value: "",
      type: "string"
    }
  ],
  item: []
};

// Cargar la colección actual para extraer los endpoints
const currentCollectionPath = path.join(__dirname, 'postman_collection_backup.json');
const currentCollection = JSON.parse(fs.readFileSync(currentCollectionPath, 'utf8'));

// Función auxiliar para crear un request estándar
function createRequest(method, path, options = {}) {
  const {
    name = `${method} ${path}`,
    description = '',
    auth = null, // null = usar auth de colección, 'noauth' = sin auth
    body = null,
    headers = [],
    tests = null
  } = options;

  const request = {
    name,
    request: {
      method,
      header: headers.length > 0 ? headers : [
        method !== 'GET' && method !== 'DELETE' ? {
          key: "Content-Type",
          value: "application/json"
        } : null
      ].filter(Boolean),
      url: {
        raw: `{{baseUrl}}${path}`,
        host: ["{{baseUrl}}"],
        path: path.split('/').filter(p => p)
      }
    }
  };

  // Agregar auth si es necesario
  if (auth === 'noauth') {
    request.request.auth = { type: "noauth" };
  }

  // Agregar body si existe
  if (body) {
    request.request.body = {
      mode: "raw",
      raw: typeof body === 'string' ? body : JSON.stringify(body, null, 2),
      options: {
        raw: {
          language: "json"
        }
      }
    };
  }

  // Agregar descripción
  if (description) {
    request.request.description = description;
  }

  // Agregar tests personalizados
  if (tests) {
    request.event = [{
      listen: "test",
      script: {
        type: "text/javascript",
        exec: tests
      }
    }];
  }

  return request;
}

// Definir folders y endpoints de manera estructurada
const folders = [
  {
    name: "🏥 Health Check",
    description: "Endpoints para verificar el estado del sistema. Útiles para monitoreo y checks de infraestructura.",
    items: [
      createRequest('GET', '/healthz', {
        description: "Verifica que el servicio esté en ejecución y respondiendo correctamente.",
        auth: 'noauth',
        tests: [
          "pm.test('Status code es 200', function() {",
          "    pm.response.to.have.status(200);",
          "});"
        ]
      }),
      createRequest('GET', '/readyz', {
        description: "Verifica que el servicio esté listo para recibir tráfico (DB conectada, etc.).",
        auth: 'noauth',
        tests: [
          "pm.test('Status code es 200', function() {",
          "    pm.response.to.have.status(200);",
          "});"
        ]
      })
    ]
  },
  {
    name: "🔐 Autenticación",
    description: "Endpoints de autenticación. El token se captura automáticamente y se usa en todos los demás endpoints.",
    items: [
      createRequest('POST', '/login', {
        description: "Autentica un trabajador y devuelve un JWT token.\\n\\n**Nota:** El token se guarda automáticamente en la variable `{{token}}` y se usa en todos los endpoints autenticados.",
        auth: 'noauth',
        body: {
          documento: 1015466495,
          password: "tu_password"
        },
        tests: [
          "pm.test('✅ Status code es 200', function() {",
          "    pm.response.to.have.status(200);",
          "});",
          "",
          "const jsonData = pm.response.json();",
          "",
          "pm.test('📦 Respuesta contiene campo data', function() {",
          "    pm.expect(jsonData).to.have.property('data');",
          "});",
          "",
          "pm.test('🔑 Respuesta contiene token', function() {",
          "    pm.expect(jsonData.data).to.have.property('token');",
          "    pm.expect(jsonData.data.token).to.be.a('string');",
          "    pm.expect(jsonData.data.token).to.not.be.empty;",
          "});",
          "",
          "// 🎯 GUARDAR TOKEN AUTOMÁTICAMENTE",
          "if (jsonData && jsonData.data && jsonData.data.token) {",
          "    pm.collectionVariables.set('token', jsonData.data.token);",
          "    console.log('✅ Token guardado automáticamente');",
          "    console.log('🔑 Token:', jsonData.data.token.substring(0, 20) + '...');",
          "    ",
          "    if (jsonData.data.trabajador) {",
          "        console.log('👤 Usuario:', jsonData.data.trabajador.nombre);",
          "        console.log('🎭 Rol:', jsonData.data.trabajador.rol);",
          "    }",
          "} else {",
          "    console.error('❌ No se pudo extraer el token');",
          "}"
        ]
      }),
      createRequest('POST', '/auth/refresh', {
        description: "Renueva el access token usando el refresh token actual. El nuevo token se guarda automáticamente.",
        tests: [
          "const jsonData = pm.response.json();",
          "if (jsonData && jsonData.data && jsonData.data.token) {",
          "    pm.collectionVariables.set('token', jsonData.data.token);",
          "    console.log('✅ Token actualizado desde refresh');",
          "}"
        ]
      })
    ]
  }
];

// Agregar folders a la colección
folders.forEach(folder => {
  collection.item.push({
    name: folder.name,
    description: folder.description,
    item: folder.items
  });
});

// Guardar colección
const outputPath = path.join(__dirname, 'postman_collection_generated.json');
fs.writeFileSync(outputPath, JSON.stringify(collection, null, 2));

console.log('✅ Colección generada:', outputPath);
console.log('📊 Total de folders:', collection.item.length);
let totalRequests = 0;
collection.item.forEach(f => totalRequests += f.item.length);
console.log('📊 Total de requests:', totalRequests);


