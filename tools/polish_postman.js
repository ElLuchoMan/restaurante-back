#!/usr/bin/env node
/**
 * Pule y perfecciona la colección de Postman
 * Agrega tests, bodies de ejemplo y mejora descripciones
 */

const fs = require("fs");
const path = require("path");

console.log("✨ Puliendo colección de Postman...\n");

const collectionPath = path.join(__dirname, "postman_collection.json");
const collection = JSON.parse(fs.readFileSync(collectionPath, "utf8"));

// Encontrar y mejorar el login
const loginFolder = collection.item.find((f) => f.name.includes("Login"));
if (loginFolder) {
  const loginReq = loginFolder.item.find((r) => r.name === "POST /login");
  if (loginReq) {
    console.log("🔐 Mejorando endpoint de login...");

    // Agregar body
    loginReq.request.body = {
      mode: "raw",
      raw: JSON.stringify(
        {
          documento: 1015466495,
          password: "tu_password",
        },
        null,
        2
      ),
      options: {
        raw: {
          language: "json",
        },
      },
    };

    // Agregar tests especiales
    loginReq.event = [
      {
        listen: "test",
        script: {
          type: "text/javascript",
          exec: [
            "// 🧪 Tests específicos para Login",
            "pm.test('✅ Status code es 200', function() {",
            "    pm.response.to.have.status(200);",
            "});",
            "",
            "const jsonData = pm.response.json();",
            "",
            "pm.test('📦 Respuesta contiene data', function() {",
            "    pm.expect(jsonData).to.have.property('data');",
            "});",
            "",
            "pm.test('🔑 Respuesta contiene token', function() {",
            "    pm.expect(jsonData.data).to.have.property('token');",
            "    pm.expect(jsonData.data.token).to.be.a('string');",
            "    pm.expect(jsonData.data.token).to.not.be.empty;",
            "});",
            "",
            "pm.test('🔄 Respuesta contiene refreshToken', function() {",
            "    pm.expect(jsonData.data).to.have.property('refreshToken');",
            "});",
            "",
            "pm.test('👤 Respuesta contiene datos del trabajador', function() {",
            "    pm.expect(jsonData.data).to.have.property('trabajador');",
            "    pm.expect(jsonData.data.trabajador).to.have.property('documento');",
            "    pm.expect(jsonData.data.trabajador).to.have.property('rol');",
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
            "    ",
            "    pm.test('💾 Token guardado en variables', function() {",
            "        pm.expect(pm.collectionVariables.get('token')).to.not.be.empty;",
            "    });",
            "} else {",
            "    console.error('❌ No se pudo extraer el token');",
            "    pm.test('❌ Error: Token no encontrado', function() {",
            "        pm.expect.fail('El token no está presente en data.token');",
            "    });",
            "}",
          ],
        },
      },
    ];

    loginReq.request.description =
      '🔐 **Autenticación de trabajadores**\\n\\nInicia sesión y obtiene un JWT token.\\n\\n**✨ El token se guarda automáticamente** en la variable `{{token}}` y se usa en todos los endpoints autenticados.\\n\\n**Ejemplo de respuesta:**\\n```json\\n{\\n  \\"success\\": true,\\n  \\"data\\": {\\n    \\"token\\": \\"eyJhbGc...\\",\\n    \\"refreshToken\\": \\"refresh...\\",\\n    \\"trabajador\\": {\\n      \\"documento\\": 1015466495,\\n      \\"nombre\\": \\"Juan\\",\\n      \\"rol\\": \\"Admin\\"\\n    }\\n  }\\n}\\n```';
  }
}

// Mejorar auth/refresh
const authFolder = collection.item.find((f) => f.name.includes("Auth"));
if (authFolder) {
  const refreshReq = authFolder.item.find((r) => r.name.includes("refresh"));
  if (refreshReq) {
    console.log("🔄 Mejorando endpoint de refresh...");

    // Cambiar de noauth a usar el token actual
    delete refreshReq.request.auth;

    // Agregar tests
    refreshReq.event = [
      {
        listen: "test",
        script: {
          type: "text/javascript",
          exec: [
            "const jsonData = pm.response.json();",
            "",
            "pm.test('Status code es 200', function() {",
            "    pm.response.to.have.status(200);",
            "});",
            "",
            "// Actualizar token automáticamente",
            "if (jsonData && jsonData.data && jsonData.data.token) {",
            "    pm.collectionVariables.set('token', jsonData.data.token);",
            "    console.log('✅ Token actualizado desde refresh');",
            "    console.log('🔑 Nuevo token:', jsonData.data.token.substring(0, 20) + '...');",
            "    ",
            "    pm.test('🔄 Token refresh exitoso', function() {",
            "        pm.expect(jsonData.data.token).to.be.a('string');",
            "    });",
            "} else {",
            "    console.error('❌ No se pudo extraer el token del refresh');",
            "}",
          ],
        },
      },
    ];

    refreshReq.request.description =
      "🔄 **Renovar token de acceso**\\n\\nRenueva el access token usando el refresh token actual (en el header Authorization).\\n\\n**✨ El nuevo token se guarda automáticamente.**";
  }
}

// Mejorar iconos de folders
const folderIcons = {
  Cambios_horario: "⏰",
  Control_nomina: "📋",
  Domicilios: "🚚",
  Horario_trabajador: "📅",
  Metodos_pago: "💳",
  Nomina_trabajador: "💼",
  Precio_producto_hist: "💰",
  Producto_pedido: "📦",
  Restaurante_dia: "🗓️",
  Reserva_contacto: "📞",
  Subcategorias: "📂",
  Push: "🔔",
};

console.log("🎨 Mejorando iconos de folders...");
collection.item.forEach((folder) => {
  // Si tiene icono 📁 genérico, intentar mejorarlo
  if (folder.name.startsWith("📁 ")) {
    const baseName = folder.name.replace("📁 ", "");
    const icon = folderIcons[baseName];
    if (icon) {
      folder.name = `${icon} ${baseName}`;
    }
  }
});

// Agregar ejemplos de body para algunos POST comunes
console.log("📝 Agregando ejemplos de body...");

const bodyExamples = {
  "POST /clientes": {
    documentoCliente: 1001,
    nombre: "Juan",
    apellido: "Pérez",
    telefono: "3000000000",
    password: "123456",
  },
  "POST /reservas": {
    clienteId: 1015466495,
    restauranteId: 1,
    fecha: "2025-10-20",
    hora: "19:00:00",
    personas: 4,
  },
  "POST /pedidos": {
    clienteId: 1015466495,
    restauranteId: 1,
    delivery: false,
  },
  "POST /cupones": {
    codigo: "BIENVENIDA10",
    scope: "GLOBAL",
    tipoDescuento: "PORCENTAJE",
    valorDescuento: 10,
    fechaInicio: "2025-01-01",
    fechaFin: "2025-12-31",
  },
};

collection.item.forEach((folder) => {
  folder.item?.forEach((request) => {
    const bodyExample = bodyExamples[request.name];
    if (
      bodyExample &&
      request.request.method === "POST" &&
      !request.request.body
    ) {
      request.request.body = {
        mode: "raw",
        raw: JSON.stringify(bodyExample, null, 2),
        options: {
          raw: {
            language: "json",
          },
        },
      };
    }
  });
});

// Guardar
fs.writeFileSync(collectionPath, JSON.stringify(collection, null, 2));

console.log("\n✅ Colección pulida exitosamente");
console.log("📊 Total de folders:", collection.item.length);
let totalRequests = 0;
collection.item.forEach((f) => (totalRequests += f.item ? f.item.length : 0));
console.log("📊 Total de requests:", totalRequests);

console.log("\n🎯 Mejoras aplicadas:");
console.log("  ✅ Login con tests de captura automática de token");
console.log("  ✅ Refresh con actualización automática de token");
console.log("  ✅ Iconos mejorados en folders");
console.log("  ✅ Ejemplos de body en endpoints comunes");
console.log("  ✅ Descripciones detalladas");

console.log("\n💡 Cómo usar:");
console.log("  1. Importa la colección en Postman");
console.log("  2. Ejecuta POST /login");
console.log("  3. ✨ El token se guarda automáticamente");
console.log(
  "  4. 🚀 Usa cualquier endpoint - el token se aplica automáticamente"
);
