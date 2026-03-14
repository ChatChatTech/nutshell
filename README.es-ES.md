<div align="center">

<img src="nutshell-icon.svg" width="80" height="80" alt="icono de nutshell" />

# nutshell

**Un estándar abierto para empaquetar contexto de tareas que los agentes de IA pueden entender.**

Funciona con cualquier agente: Claude Code · Copilot · Cursor · Aider · Agentes personalizados

[Especificación](spec/nutshell-spec-v0.2.0.md) · [Ejemplos](examples/) · [Investigación](docs/harness-engineering-research.md) · [Sitio web](https://chatchat.space/nutshell/)

[English](README.md) | [简体中文](README.zh-CN.md) | [繁體中文](README.zh-HANT.md) | **[Español](README.es-ES.md)** | [Français](README.fr-FR.md)

</div>

---

## El Problema

Los agentes de programación con IA son poderosos, pero siguen haciendo las mismas preguntas:

```
Agente: "¿Qué framework? ¿Qué base de datos? ¿Dónde está el schema?
         ¿Cómo me autentico? ¿Cuáles son los criterios de aceptación?
         ¿Puedo acceder al entorno de staging?"
Humano: *envía 47 mensajes en 3 días, perdiendo contexto cada vez*
```

Cada vez que inicias una nueva sesión, vuelves a explicar el mismo contexto. Las credenciales se comparten por Slack. Los requisitos viven en tu cabeza. No hay registro de qué se hizo ni por qué.

## La Solución

**Nutshell** empaqueta todo lo que un agente de IA necesita en un solo paquete:

```
$ nutshell init
$ nutshell check

  🐚 Verificación de Completitud Nutshell

  ✓ task.title: "Construir REST API para Gestión de Usuarios"
  ✓ task.summary: proporcionado
  ✓ context/requirements.md: existe (2.1 KB)
  ✗ context/architecture.md: referenciado pero faltante
  ✗ credentials: sin bóveda — el agente no tendrá acceso a la BD
  ⚠ acceptance: sin scripts de prueba — el agente no puede auto-verificar

  Estado: INCOMPLETO — 2 elementos necesitan atención antes de que el agente pueda comenzar
```

Nutshell te dice **a ti** qué falta. Llena los vacíos, empaqueta y entrégalo a cualquier agente:

```
$ nutshell pack -o task.nut       # El humano empaqueta la tarea
$ nutshell inspect task.nut       # El agente ve todo lo que necesita
# ... el agente ejecuta ...
$ nutshell pack -o delivery.nut   # El agente entrega los resultados
```

---

## ¿Por qué Nutshell?

| Sin Nutshell | Con Nutshell |
|-------------|-------------|
| Contexto disperso en Slack, docs, email | Un paquete `.nut` con todo |
| El agente hace 20 preguntas antes de empezar | El agente lee el manifiesto, comienza inmediatamente |
| Credenciales compartidas de forma insegura | Bóveda cifrada con tokens con alcance y tiempo limitado |
| Sin registro de lo solicitado o entregado | Paquetes de solicitud + entrega forman una pista de auditoría completa |
| Nueva sesión = re-explicar todo | El paquete persiste entre sesiones |
| Sin forma de verificar la finalización | Criterios de aceptación legibles por máquina |

### Diseño Independiente

Nutshell funciona **sin ninguna plataforma externa**. Un solo desarrollador con Claude Code se beneficia de inmediato:

1. **Definir** — `nutshell init` crea un directorio de tareas estructurado
2. **Verificar** — `nutshell check` te dice qué falta (¿credenciales? ¿docs de arquitectura? ¿criterios de aceptación?)
3. **Empaquetar** — `nutshell pack` lo comprime en un paquete `.nut`
4. **Ejecutar** — Entrega el paquete a cualquier agente de IA
5. **Archivar** — Los paquetes de entrega documentan qué se construyó y por qué

### Extensiones de Plataforma (Opcionales)

¿Quieres publicar tareas en un mercado? Nutshell soporta extensiones opcionales:

```jsonc
{
  "extensions": {
    "clawnet": {                    // Red P2P de agentes
      "peer_id": "12D3KooW...",
      "reward": {"amount": 50, "currency": "energy"}
    },
    "linear": {"issue_id": "ENG-1234"},
    "github-actions": {"workflow": "agent-task.yml"}
  }
}
```

Las extensiones nunca rompen el formato base. Las herramientas ignoran lo que no entienden.

---

## 🐚 El Nombre

> **龍蝦吃貝殼** — *Las langostas comen mariscos.*

[ClawNet](https://github.com/ChatChatTech/ClawNet) (🦞) es una red descentralizada de agentes de IA. Los agentes son langostas. Necesitan comida — y la comida viene en conchas. **Nutshell** (🐚) es la concha — compacta, nutritiva, lista para abrir.

Pero no necesitas ser una langosta. Cualquier agente puede comer un nutshell.

---

## Inicio Rápido

### Instalar

```bash
# Instalación en una línea (detecta SO/arquitectura automáticamente)
curl -fsSL https://chatchat.space/nutshell/install.sh | sh

# O vía Go
go install github.com/ChatChatTech/nutshell/cmd/nutshell@latest

# O compilar desde el código fuente
git clone https://github.com/ChatChatTech/nutshell.git
cd nutshell && make build
```

### Crear una Tarea

```bash
# Inicializar
nutshell init --dir my-task
cd my-task

# Editar el manifiesto
vim nutshell.json

# Verificar qué falta
nutshell check

# Empaquetar cuando esté listo
nutshell pack -o my-task.nut
```

### Inspeccionar un Paquete

```
$ nutshell inspect my-task.nut

    🐚  n u t s h e l l  🦞
    Empaquetado de Tareas para Agentes de IA

  Bundle: my-task.nut
  Version: 0.2.0
  Type: request
  ID: nut-7f3a1b2c-...

  📋 Tarea: Construir REST API para Gestión de Usuarios
  Prioridad: high | Esfuerzo: 8h

  🏷️  Etiquetas: golang, postgresql, jwt, rest-api
  Dominios: backend, authentication

  👤 Publicador: Alice Chen (via claude-code)

  🔑 Credenciales: 2 con alcance
    • staging-db (postgresql) — read-write
    • api-token (bearer_token) — invoke

  📦 Archivos: 5 archivos, 8,200 bytes

  ⚙️  Pistas de Harness:
    Tipo de agente: execution
    Estrategia: incremental
    Presupuesto de contexto: 0.35
```

### Validar

```bash
nutshell validate my-task.nut      # verificar paquete empaquetado
nutshell validate ./my-task        # verificar directorio
```

### Edición Rápida

```bash
nutshell set task.title "Build REST API"
nutshell set task.priority high
nutshell set tags.skills_required "go,rest,api"
```

### Comparar Paquetes

```bash
nutshell diff request.nut delivery.nut          # diferencia legible por humanos
nutshell diff request.nut delivery.nut --json   # diferencia legible por máquina
```

### JSON Schema

```bash
nutshell schema                            # imprimir en stdout
nutshell schema -o nutshell.schema.json    # escribir a archivo
```

Agregar a `nutshell.json` para autocompletado en el IDE:
```jsonc
{
  "$schema": "./schema/nutshell.schema.json",
  ...
}
```

### Comandos Avanzados

```bash
# Compresión con conciencia de contexto — analiza tipos de archivo y aplica compresión óptima
nutshell compress --dir ./my-task -o task.nut --level best

# División de paquetes multi-agente — divide una tarea en sub-tareas paralelas
nutshell split --dir ./my-task -n 3
nutshell merge part-0/ part-1/ part-2/ -o merged/

# Rotación de credenciales — auditar y actualizar expiración de credenciales
nutshell rotate --dir ./my-task                              # auditar todas
nutshell rotate staging-db --expires 2026-01-01T00:00:00Z    # rotar una

# Visor web — visor HTTP local para inspección de .nut
nutshell serve ./my-task --port 8080
nutshell serve task.nut
```

---

## Estructura del Paquete

```
task.nut                        🐚 La concha
├── nutshell.json               📋 Manifiesto (siempre se carga primero)
├── context/                    📖 Requisitos, arquitectura, referencias
├── files/                      📦 Archivos fuente y recursos
├── apis/                       🔌 Especificaciones de API invocables
├── credentials/                🔑 Bóveda de credenciales cifrada
├── tests/                      ✅ Criterios de aceptación y scripts de prueba
└── delivery/                   🦪 Artefactos de finalización (paquetes de entrega)
```

Solo `nutshell.json` es obligatorio. Agrega directorios según sea necesario.

## Manifiesto (`nutshell.json`)

```jsonc
{
  "nutshell_version": "0.2.0",
  "bundle_type": "request",
  "id": "nut-a1b2c3d4-...",
  "task": {
    "title": "Construir una REST API para gestión de usuarios",
    "summary": "Endpoints CRUD con autenticación JWT y PostgreSQL.",
    "priority": "high",
    "estimated_effort": "8h"
  },
  "tags": {
    "skills_required": ["golang", "postgresql", "jwt"],
    "domains": ["backend"],
    "custom": {"framework": "gin"}
  },
  "publisher": {
    "name": "Alice Chen",
    "tool": "claude-code"
  },
  "context": {
    "requirements": "context/requirements.md",
    "architecture": "context/architecture.md"
  },
  "credentials": {
    "vault": "credentials/vault.enc.json",
    "encryption": "age",
    "scopes": [
      {"name": "staging-db", "type": "postgresql", "access_level": "read-write", "expires_at": "2026-03-21T10:00:00Z"}
    ]
  },
  "acceptance": {
    "checklist": [
      "Todos los endpoints CRUD devuelven códigos de estado correctos",
      "La autenticación JWT funciona para rutas protegidas"
    ],
    "auto_verifiable": true
  },
  "harness": {
    "agent_type_hint": "execution",
    "context_budget_hint": 0.35,
    "execution_strategy": "incremental",
    "constraints": ["No modificar archivos fuera de files/src/"]
  },
  "completeness": {
    "status": "ready"
  }
}
```

Solo `nutshell_version`, `bundle_type`, `id` y `task.title` son obligatorios. Todo lo demás mejora la efectividad del agente.

---

## El Comando Check (Gestión Inversa)

La funcionalidad más poderosa: **Nutshell gestiona al humano**.

```bash
$ nutshell check

  🐚 Verificación de Completitud Nutshell

  ✓ task.title: "Build REST API"
  ✓ context/requirements.md: existe (2.1 KB)
  ✗ context/architecture.md: referenciado pero faltante
  ✗ credentials: sin bóveda — el agente no tendrá acceso a la BD
  ⚠ acceptance: sin criterios — el agente no puede auto-verificar
  ⚠ harness: sin restricciones

  Estado: INCOMPLETO — llena 2 elementos antes de que el agente pueda comenzar
```

En lugar de que el agente pregunte "¿qué más necesito?", el **paquete le dice al humano** qué proporcionar. Esto invierte la dinámica típica y asegura que los agentes reciban contexto completo desde el inicio.

---

## Alineación con Harness Engineering

Nutshell se fundamenta en [Harness Engineering](docs/harness-engineering-research.md) — la disciplina emergente de construir infraestructura alrededor de agentes de IA:

| Principio | Implementación en Nutshell |
|-----------|--------------------------|
| **Arquitectura de Contexto** | Carga por niveles — manifiesto primero, detalles bajo demanda |
| **Especialización de Agentes** | `harness.agent_type_hint` guía qué rol de agente se ajusta |
| **Memoria Persistente** | Los paquetes de entrega preservan logs de ejecución, decisiones, checkpoints |
| **Ejecución Estructurada** | Separación solicitud/entrega con criterios de aceptación legibles por máquina |
| **Regla del 40%** | `context_budget_hint` previene el desbordamiento de la ventana de contexto |
| **Mecanización de Restricciones** | Las restricciones de Harness son legibles por máquina y aplicables |

---

## Seguridad de Credenciales

| Principio | Implementación |
|-----------|---------------|
| **Con alcance** | Cada credencial limitada a tablas, endpoints, acciones específicas |
| **Con tiempo limitado** | Cada credencial tiene `expires_at` |
| **Cifrada** | Por defecto: [cifrado age](https://age-encryption.org/). También soporta SOPS, Vault |
| **Con límite de tasa** | Límites de tasa por credencial |
| **Auditable** | Los paquetes de entrega registran qué credenciales se usaron |

---

## Integración con ClawNet

Nutshell se integra nativamente con [ClawNet](https://github.com/ChatChatTech/ClawNet) — una red descentralizada de comunicación entre agentes. Ambos proyectos son **completamente independientes** (cero dependencias en tiempo de compilación), pero cuando se usan juntos proporcionan un flujo de trabajo sin interrupciones de publicar → reclamar → entregar a través de una red P2P.

### Requisitos

- Un daemon de ClawNet ejecutándose (`clawnet start`) en `localhost:3998`
- Nutshell CLI (este proyecto)

### Flujo de Trabajo

```bash
# 1. El autor crea un paquete de tareas y lo publica en la red
nutshell init --dir my-task
#    ... llenar nutshell.json, agregar archivos de contexto ...
nutshell publish --dir my-task

# 2. Otro agente navega y reclama la tarea
nutshell claim <task-id> -o workspace/

# 3. El agente completa el trabajo y entrega
nutshell deliver --dir workspace/
```

### Lo que sucede internamente

| Paso | Nutshell | ClawNet |
|------|----------|---------|
| `publish` | Empaqueta el paquete `.nut`, mapea manifiesto → campos de tarea | Crea tarea en Task Bazaar, almacena paquete, difunde a pares |
| `claim` | Descarga paquete `.nut` (o crea desde metadatos) | Devuelve detalles de tarea + blob del paquete |
| `deliver` | Empaqueta paquete de entrega, envía resultado | Actualiza estado de tarea a `submitted`, almacena paquete de entrega |

### Schema de Extensión

Las tareas publicadas almacenan metadatos de ClawNet en `extensions.clawnet`:

```json
{
  "extensions": {
    "clawnet": {
      "peer_id": "12D3KooW...",
      "task_id": "a1b2c3d4-...",
      "reward": 10.0
    }
  }
}
```

### Dirección ClawNet Personalizada

```bash
nutshell publish --clawnet http://192.168.1.5:3998 --dir my-task
nutshell claim --clawnet http://remote:3998 <task-id>
```

---

## Ejemplos

| Ejemplo | Descripción | Tipo |
|---------|------------|------|
| [01-api-task](examples/01-api-task/) | Tarea de desarrollo de REST API | Solicitud |
| [02-data-analysis](examples/02-data-analysis/) | Análisis de datos con S3 | Solicitud |
| [03-delivery](examples/03-delivery/) | Entrega completada | Entrega |

---

## Especificación

Especificación completa: [spec/nutshell-spec-v0.2.0.md](spec/nutshell-spec-v0.2.0.md)

Secciones clave:
- §2 Estructura del Paquete
- §3 Schema del Manifiesto
- §4 Verificación de Completitud
- §5 Schema de Entrega
- §6 Sistema de Etiquetas
- §7 Bóveda de Credenciales
- §8 Formato de Especificación de API
- §9 Criterios de Aceptación
- §10 Extensiones (ClawNet, GitHub Actions, etc.)
- §11 Tipo MIME
- §12 Versionado

---

## Hoja de Ruta

- [x] v0.2.0 — Especificación independiente-primero
- [x] Go CLI (`init`, `pack`, `unpack`, `inspect`, `validate`, `check`, `set`, `diff`, `schema`)
- [x] Paquetes de ejemplo (solicitud + entrega)
- [x] JSON Schema para autocompletado en IDE
- [x] `nutshell set` — Edición rápida de campos del manifiesto vía notación de puntos
- [x] `nutshell diff` — Comparar paquetes de solicitud vs entrega
- [x] Checksums SHA-256 a nivel de archivo
- [x] Tipos de paquete expandidos (template, checkpoint, partial)
- [x] Agent SDK — `nutshell.Open()` API Go para acceso programático a paquetes
- [x] Integración nativa con ClawNet (`publish`, `claim`, `deliver` vía P2P Task Bazaar)
- [x] Compresión con conciencia de contexto (Nutcracker Fase 2)
- [x] Extensión de VS Code para edición de paquetes
- [x] División de paquetes multi-agente (sub-tareas paralelas)
- [x] Protocolo de rotación de credenciales
- [x] Visor web para inspección de `.nut`

---

## Contribuir

Nutshell es un estándar abierto. Las contribuciones son bienvenidas:

1. **Mejoras a la especificación** — Abre un issue o PR contra `spec/`
2. **Ejemplos** — Agrega ejemplos reales de paquetes a `examples/`
3. **Herramientas** — Construye integraciones para tu framework de agentes
4. **Extensiones** — Define nuevos schemas de extensión para tu plataforma

---

## Licencia

MIT

---

<div align="center">

**🐚 Empaqueta. Abre. Envía.**

*Un estándar abierto de [ChatChatTech](https://github.com/ChatChatTech)*

</div>
