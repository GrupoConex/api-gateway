# Flujo de Comunicación Inter-Servicios (Fibex Gateway)

Este documento describe el estándar de comunicación entre backends utilizando el API Gateway y Keycloak para autenticación Machine-to-Machine (M2M).

## 1. Identidad del Servicio
Cada microservicio debe tener sus propias credenciales en Keycloak (Client ID y Client Secret) configuradas en su archivo `.env`:
- `KC_CLIENT_ID`: Identificador del cliente.
- `KC_CLIENT_SECRET`: Secreto del cliente.
- `KEYCLOAK_URL`: URL del servidor de identidad.
- `GATEWAY_URL`: URL del API Gateway.

## 2. Proceso de Comunicación (Backend A -> Backend B)

### Paso A: Obtención del Token (M2M)
El Backend A no utiliza el token del usuario final. En su lugar, solicita un token propio a Keycloak:
- **Endpoint:** `${KEYCLOAK_URL}/realms/fibex/protocol/openid-connect/token`
- **Método:** `POST`
- **Body (form-urlencoded):**
  - `grant_type`: `client_credentials`
  - `client_id`: `${KC_CLIENT_ID}`
  - `client_secret`: `${KC_CLIENT_SECRET}`

### Paso B: Petición al Gateway
Con el token obtenido, el Backend A realiza la petición al Gateway especificando el servicio destino en la URL:
- **URL:** `${GATEWAY_URL}/api/{target_service}/{endpoint}`
- **Header:** `Authorization: Bearer {access_token}`

### Paso C: Validación y Proxy
1. El **Gateway** recibe la petición en `/api/*`.
2. El middleware del Gateway valida el token JWT contra Keycloak.
3. El Gateway identifica el `{target_service}` (ej: `intranet`) y busca su URL interna en su configuración (`PROXY_INTRANET`).
4. El Gateway redirige la petición al servicio destino.

## 3. Implementación en Viáticos (Ejemplo)

Se ha implementado el `GatewayIntegrationService` que encapsula este flujo. Para usarlo:

```typescript
// Inyectar el servicio
constructor(private readonly gateway: GatewayIntegrationService) {}

// Realizar una llamada a Intranet
async syncData() {
  const data = await this.gateway.callGateway('intranet', '/users/profile', 'GET');
  return data;
}
```
