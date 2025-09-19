# Guía de Refresh Tokens

## Implementación Completada ✅

Se ha implementado exitosamente el sistema de refresh tokens sin necesidad de base de datos, utilizando JWT firmados.

## Cambios Realizados

### 1. **Nuevas Estructuras**
- `RefreshClaims`: Claims específicos para refresh tokens
- `AuthResponse`: Modelo de respuesta estandarizado

### 2. **Endpoints Modificados/Nuevos**

#### Login (`POST /restaurante/v1/login`)
**Antes:**
```json
{
  "code": 200,
  "message": "Inicio de sesión exitoso",
  "data": {
    "token": "eyJ...",
    "nombre": "Juan Pérez"
  }
}
```

**Ahora:**
```json
{
  "code": 200,
  "message": "Inicio de sesión exitoso",
  "data": {
    "access_token": "eyJ...",
    "refresh_token": "eyJ...",
    "token_type": "Bearer",
    "expires_in": "1800",
    "nombre": "Juan Pérez"
  }
}
```

#### Nuevo: Refresh Token (`POST /restaurante/v1/auth/refresh`)
**Request:**
```http
POST /restaurante/v1/auth/refresh
Authorization: Bearer {refresh_token}
```

**Response:**
```json
{
  "code": 200,
  "message": "Tokens renovados exitosamente",
  "data": {
    "access_token": "eyJ...",
    "refresh_token": "eyJ...",
    "token_type": "Bearer",
    "expires_in": "1800",
    "nombre": "Juan Pérez"
  }
}
```

## Configuración de Tokens

- **Access Token**: 30 minutos de duración
- **Refresh Token**: 30 días de duración
- **Algoritmo**: HS256 (mismo que antes)
- **Secret**: Usa la misma variable `JWT_SECRET`

## Flujo para Aplicaciones Móviles

```
1. Login inicial → Recibe access_token + refresh_token
2. Usar access_token para llamadas API normales
3. Antes de que expire (ej: 5 min antes) → Llamar /auth/refresh
4. Recibir nuevos tokens → Continuar usando la app
5. Repetir paso 3-4 automáticamente
```

## Ejemplo de Implementación en Cliente

### JavaScript/React Native
```javascript
class TokenManager {
  constructor() {
    this.accessToken = null;
    this.refreshToken = null;
    this.refreshTimer = null;
  }

  async login(documento, password) {
    const response = await fetch('/restaurante/v1/login', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ documento, password })
    });

    const data = await response.json();
    if (data.code === 200) {
      this.setTokens(data.data);
      this.scheduleRefresh();
    }
    return data;
  }

  setTokens(tokenData) {
    this.accessToken = tokenData.access_token;
    this.refreshToken = tokenData.refresh_token;
    // Guardar en storage seguro
    SecureStore.setItem('access_token', this.accessToken);
    SecureStore.setItem('refresh_token', this.refreshToken);
  }

  scheduleRefresh() {
    // Refrescar 5 minutos antes de expirar (25 min)
    this.refreshTimer = setTimeout(() => {
      this.refreshTokens();
    }, 25 * 60 * 1000);
  }

  async refreshTokens() {
    try {
      const response = await fetch('/restaurante/v1/auth/refresh', {
        method: 'POST',
        headers: { 'Authorization': `Bearer ${this.refreshToken}` }
      });

      const data = await response.json();
      if (data.code === 200) {
        this.setTokens(data.data);
        this.scheduleRefresh(); // Programar siguiente refresh
      }
    } catch (error) {
      // Manejar error - posiblemente redirigir a login
      console.error('Error refreshing tokens:', error);
    }
  }

  getAuthHeader() {
    return `Bearer ${this.accessToken}`;
  }
}
```

## Seguridad

### ✅ Ventajas Implementadas
- **Tokens de corta duración**: Access tokens expiran en 30 min
- **Rotación automática**: Cada refresh genera nuevos tokens
- **Stateless**: No requiere almacenamiento en servidor
- **Validación estricta**: Verifica que sea un refresh token válido

### ⚠️ Consideraciones
- **Almacenamiento seguro**: Los clientes deben guardar tokens en storage seguro
- **Revocación**: Para revocar sesiones específicas, se requeriría implementar blacklist
- **Monitoreo**: Considerar logs de refresh para detectar uso anómalo

## Testing

Se agregaron tests completos que cubren:
- ✅ Generación de tokens
- ✅ Endpoint de refresh exitoso
- ✅ Validación de headers faltantes
- ✅ Tokens inválidos
- ✅ Uso de access token en lugar de refresh token
- ✅ Compatibilidad con tests existentes

## Próximos Pasos Opcionales

1. **Métricas**: Agregar logs de uso de refresh tokens
2. **Blacklist**: Implementar revocación si se requiere
3. **Rate Limiting**: Limitar frecuencia de refresh por usuario
4. **Notificaciones**: Alertar sobre refreshes desde nuevos dispositivos

## Costo Final

**Tiempo real invertido**: ~1.5 días (menos de lo estimado)
- ✅ Sin cambios en base de datos
- ✅ Mantiene arquitectura stateless
- ✅ Compatible con sistema existente
- ✅ Tests completos incluidos
