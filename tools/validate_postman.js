#!/usr/bin/env node
/**
 * Script de validación de colección Postman vs Swagger
 * Verifica que los endpoints de Postman estén sincronizados con Swagger
 */

const fs = require("fs");
const path = require("path");

// Colores para terminal
const colors = {
  reset: "\x1b[0m",
  green: "\x1b[32m",
  yellow: "\x1b[33m",
  red: "\x1b[31m",
  cyan: "\x1b[36m",
  bold: "\x1b[1m",
};

function log(message, color = "reset") {
  console.log(`${colors[color]}${message}${colors.reset}`);
}

// Cargar archivos
const swaggerPath = path.join(__dirname, "..", "docs", "swagger.json");
const postmanPath = path.join(__dirname, "postman_collection.json");

try {
  log("\n=== VALIDACIÓN DE COLECCIÓN POSTMAN ===\n", "bold");

  // Validar que los archivos existen
  if (!fs.existsSync(swaggerPath)) {
    log("❌ Error: No se encuentra docs/swagger.json", "red");
    process.exit(1);
  }

  if (!fs.existsSync(postmanPath)) {
    log("❌ Error: No se encuentra tools/postman_collection.json", "red");
    process.exit(1);
  }

  // Cargar y parsear archivos
  const swagger = JSON.parse(fs.readFileSync(swaggerPath, "utf8"));
  const postman = JSON.parse(fs.readFileSync(postmanPath, "utf8"));

  log("✅ JSON válido: swagger.json", "green");
  log("✅ JSON válido: postman_collection.json", "green");

  // Extraer endpoints de Swagger
  const swaggerEndpoints = new Set();
  const swaggerMethods = {};

  Object.keys(swagger.paths).forEach((path) => {
    const methods = Object.keys(swagger.paths[path]);
    methods.forEach((method) => {
      const endpoint = `${method.toUpperCase()} ${path}`;
      swaggerEndpoints.add(endpoint);
      if (!swaggerMethods[path]) {
        swaggerMethods[path] = [];
      }
      swaggerMethods[path].push(method.toUpperCase());
    });
  });

  // Extraer requests de Postman
  const postmanRequests = [];
  const extractRequests = (items, folder = "") => {
    items.forEach((item) => {
      if (item.item) {
        // Es una carpeta
        extractRequests(item.item, item.name);
      } else if (item.request) {
        // Es un request
        const method = item.request.method;
        const url = item.request.url;
        let path = "";

        if (typeof url === "string") {
          // Extraer path de la URL
          path = url.replace("{{baseUrl}}", "").split("?")[0];
        } else if (url.raw) {
          path = url.raw.replace("{{baseUrl}}", "").split("?")[0];
        }

        // Convertir path params de Postman a Swagger format
        // Ej: /cupones/BIENVENIDA10/redimir -> /cupones/{codigo}/redimir
        const normalizedPath = path.replace(
          /\/[A-Z0-9_]+\/redimir$/,
          "/{codigo}/redimir"
        );

        postmanRequests.push({
          name: item.name,
          method,
          path: normalizedPath,
          folder,
          original: `${method} ${path}`,
        });
      }
    });
  };

  extractRequests(postman.item);

  // Estadísticas
  log("\n=== ESTADÍSTICAS ===\n", "cyan");
  log(`Carpetas en Postman: ${postman.item.length}`);
  log(`Requests en Postman: ${postmanRequests.length}`);
  log(`Endpoints en Swagger: ${swaggerEndpoints.size}`);
  log(`Paths en Swagger: ${Object.keys(swagger.paths).length}`);

  // Verificar configuración
  log("\n=== CONFIGURACIÓN ===\n", "cyan");
  const baseUrlVar = postman.variable.find((v) => v.key === "baseUrl");
  if (baseUrlVar) {
    log(`✅ baseUrl: ${baseUrlVar.value}`, "green");
  } else {
    log("❌ No se encontró variable baseUrl", "red");
  }

  if (postman.info.version) {
    log(`✅ Versión: ${postman.info.version}`, "green");
  }

  // Buscar requests que no están en Swagger
  log("\n=== ANÁLISIS DE SINCRONIZACIÓN ===\n", "cyan");

  const notInSwagger = [];
  const inSwagger = [];

  postmanRequests.forEach((req) => {
    const endpoint = `${req.method} ${req.path}`;

    // Ignorar algunos endpoints especiales
    const isSpecial =
      req.path === "/healthz" ||
      req.path === "/readyz" ||
      req.path.includes("(") || // Requests con anotaciones
      req.name.includes("filtros") ||
      req.name.includes("con filtros");

    if (!swaggerEndpoints.has(endpoint) && !isSpecial) {
      notInSwagger.push(req);
    } else {
      inSwagger.push(req);
    }
  });

  log(`✅ Requests sincronizados: ${inSwagger.length}`, "green");

  if (notInSwagger.length > 0) {
    log(
      `⚠️  Requests no encontrados en Swagger: ${notInSwagger.length}`,
      "yellow"
    );
    notInSwagger.forEach((req) => {
      log(`   - ${req.original} (${req.folder})`, "yellow");
    });
  } else {
    log("✅ Todos los requests están en Swagger", "green");
  }

  // Buscar endpoints de Swagger que no están en Postman
  log("\n=== ENDPOINTS FALTANTES EN POSTMAN ===\n", "cyan");

  const postmanEndpoints = new Set(
    postmanRequests.map((r) => `${r.method} ${r.path}`)
  );

  const missingInPostman = [];
  swaggerEndpoints.forEach((endpoint) => {
    if (!postmanEndpoints.has(endpoint)) {
      // Ignorar algunos métodos especiales que Postman combina en requests con filtros
      const [method, path] = endpoint.split(" ");
      if (method === "GET" && postmanEndpoints.has(`GET ${path}?`)) {
        // Está cubierto por request con query params
        return;
      }
      missingInPostman.push(endpoint);
    }
  });

  if (missingInPostman.length > 0) {
    log(
      `⚠️  Endpoints de Swagger no en Postman: ${missingInPostman.length}`,
      "yellow"
    );
    missingInPostman.forEach((endpoint) => {
      log(`   - ${endpoint}`, "yellow");
    });
  } else {
    log("✅ Todos los endpoints de Swagger están en Postman", "green");
  }

  // Resumen final
  log("\n=== RESUMEN ===\n", "bold");

  const totalIssues = notInSwagger.length + missingInPostman.length;

  if (totalIssues === 0) {
    log(
      "🎉 La colección de Postman está completamente sincronizada con Swagger!",
      "green"
    );
    process.exit(0);
  } else {
    log(
      `⚠️  Se encontraron ${totalIssues} diferencias entre Postman y Swagger`,
      "yellow"
    );
    log("   Revisa los detalles arriba para más información.", "yellow");
    process.exit(0); // No fallar, solo advertir
  }
} catch (error) {
  log(`\n❌ Error durante la validación: ${error.message}`, "red");
  console.error(error);
  process.exit(1);
}
