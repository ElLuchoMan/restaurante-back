# Despliegue de staging: backend (Render) + base de datos (Neon)

Guía probada contra el código del repo: el esquema se cargó desde cero en Postgres 16 y el backend arrancó
en modo `prod` respondiendo `/restaurante/v1/productos` con esas variables de entorno.

## 0. Qué se despliega y dónde

| Pieza | Servicio | Repo / carpeta |
|---|---|---|
| Base de datos Postgres | Neon (plan gratis) | `restaurante-db` |
| Backend Go + Beego | Render (web service, Docker) | `restaurante-back` (`Dockerfile`, `render.yaml`) |
| Frontend Angular | Netlify o Cloudflare Pages | `restaurante-frontend` |

## 1. Base de datos en Neon

1. Crea una cuenta en <https://neon.com> y un proyecto nuevo. Elige la región más cercana a ti.
2. En el panel del proyecto abre **Connect**. Desactiva **Pooled connection** y copia la cadena de
   conexión directa. El backend usa prepared statements; el pooler en modo transacción puede dar problemas.
   Tiene esta forma:
   `postgresql://USUARIO:CLAVE@ep-xxxx.region.aws.neon.tech/neondb?sslmode=require`
3. Desde la carpeta `restaurante-db` ejecuta los scripts **en este orden** (necesitas `psql`):

   ```bash
   psql "CADENA_DE_CONEXION" -f Scripts/Restaurante.sql
   psql "CADENA_DE_CONEXION" -f Scripts/Funciones.sql
   psql "CADENA_DE_CONEXION" -f Scripts/Datos.sql      # solo staging: datos de prueba
   ```

   Equivale a `psql "CADENA_DE_CONEXION" -f init.sql` ejecutado desde la raíz del repo.
   Sin `psql` puedes pegar cada archivo en el **SQL Editor** de Neon, en el mismo orden.
4. Verifica: `psql "CADENA_DE_CONEXION" -c "\dt"` debe listar 33 tablas.

Notas:
- `Datos.sql` termina en `COMMIT` (el README antiguo decía `ROLLBACK`). No lo cargues en producción.
- Los archivos de `Scripts/migrations/` son solo para bases creadas con un esquema anterior. `Restaurante.sql`
  ya incluye todo; ejecutarlos encima de una base nueva da error de llaves primarias duplicadas.
- Plan gratis de Neon: se suspende tras 5 minutos sin uso; la primera petición después tarda un poco más.

## 2. Backend en Render

1. Sube la rama que contiene `Dockerfile` y `render.yaml` a GitHub.
2. En <https://render.com>: **New → Blueprint**, conecta el repo `restaurante-back` y elige esa rama.
   Render lee `render.yaml` y crea el servicio `restaurante-back` (plan Free, runtime Docker).
3. Render pide los valores marcados como secretos. Sácalos de la cadena de Neon:

   | Variable | Valor |
   |---|---|
   | `DB_HOST` | `ep-xxxx.region.aws.neon.tech` |
   | `DB_PORT` | `5432` |
   | `DB_USER` | usuario de Neon |
   | `DB_PASS` | clave de Neon |
   | `DB_NAME` | `neondb` (o el nombre de tu base) |
   | `DB_SSLMODE` | `require` (ya viene en `render.yaml`) |
   | `JWT_SECRET` | Render lo genera solo; cámbialo si ya tienes tokens emitidos |
   | `CORS_ALLOWED_ORIGINS` | URL del frontend, separadas por coma. Ej.: `https://tu-sitio.netlify.app` |
   | `PORT` | `8080` (ya viene en `render.yaml`; Beego escucha en 8080) |
   | `VAPID_PUBLIC_KEY`, `VAPID_PRIVATE_KEY`, `VAPID_SUBJECT` | solo si usas notificaciones push |

   En modo `prod` el CORS solo acepta los orígenes de `CORS_ALLOWED_ORIGINS` (más los que trae por defecto).
4. Espera el primer deploy. Prueba:

   ```bash
   curl https://TU-SERVICIO.onrender.com/restaurante/v1/productos-disponibles
   ```

   Debe responder `{"code":200,...}`. Plan gratis: el servicio se duerme tras 15 minutos sin tráfico y tarda
   cerca de 1 minuto en despertar.
5. `render.yaml` no define `healthCheckPath` a propósito: cualquier ruta que consulte la base mantendría
   despierto a Neon y gastaría sus horas gratis.

## 3. Frontend contra el backend desplegado

- GitHub Actions (`deploy.yml`) genera `public/app-config.json` con el secret `APP_API_BASE`. Créalo en
  **Settings → Secrets and variables → Actions** con el valor
  `https://TU-SERVICIO.onrender.com/restaurante/v1`.
- Si usas el proxy del `netlify.toml` (`/restaurante/v1/*`), cambia `http://localhost:8080` por la URL de Render.
- Angular 22 necesita Node 22.22.3 o superior; en CI y Netlify está fijado en Node 24.

## 4. Problemas frecuentes

| Síntoma | Causa probable |
|---|---|
| `no pg_hba.conf entry` o `SSL is not enabled` | falta `DB_SSLMODE=require` |
| Errores raros de prepared statement | usaste la cadena **pooled** de Neon; usa la directa |
| El navegador bloquea las llamadas (CORS) | falta la URL exacta del frontend en `CORS_ALLOWED_ORIGINS` |
| Primera petición tarda 1 minuto | Render gratis dormido; es normal |
| `relation ... does not exist` | no se cargó `Restaurante.sql` o se cargó incompleto |

## 5. Correr todo en local

1. `psql -U postgres -c "create database restaurante_db"` y carga los tres scripts como en el paso 1.
2. Copia `.env.example` a `.env` y completa `DB_*` y `JWT_SECRET`.
3. `go run .` (requiere Go 1.26 si usas la rama de dependencias nuevas, 1.25 en `develop`).
