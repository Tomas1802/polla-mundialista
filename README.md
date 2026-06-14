# Polla Mundialista 2026

Aplicación web para administrar una polla de la Copa Mundial de la FIFA 2026.
Diseñada para ser **simple y obvia** — pensada para que cualquier persona,
incluidos adultos mayores, la use sin instrucciones.

Las reglas y el sistema de puntos están en
[`reglamento_polla_mundial_2026.md`](./reglamento_polla_mundial_2026.md) y se
muestran dentro de la app en la sección **Reglas**.

## Arquitectura

| Capa | Tecnología | Despliegue |
|---|---|---|
| Frontend | React + Vite | GitHub Pages (estático) |
| Backend | Go (API REST) | Google Cloud Run |
| Base de datos | PostgreSQL | Google Cloud SQL |
| Autenticación | Firebase Phone Auth (OTP por SMS) | — |
| Datos de partidos | football-data.org API v4 (`WC`) | cacheados en la BD |

```
POLLA/
├─ reglamento_polla_mundial_2026.md
├─ backend/                  API en Go
│  ├─ cmd/server/            arranque HTTP
│  ├─ internal/
│  │  ├─ scoring/            motor de puntuación (puro, con tests)
│  │  ├─ auth/               verificación Firebase + sesión JWT
│  │  ├─ football/           cliente football-data + sync diario
│  │  ├─ db/                 conexión + repositorios (pgx)
│  │  ├─ ranking/            ranking + desempate
│  │  └─ api/                handlers + router
│  ├─ migrations/            esquema PostgreSQL
│  └─ Dockerfile
└─ frontend/                 React + Vite
   ├─ src/pages/             Marcadores · Ranking
   └─ src/auth/              login celular + OTP
```

## Cómo funciona

### Autenticación
1. El usuario escribe su **número de celular**. Firebase envía un **PIN dinámico
   (OTP) por SMS**.
2. El frontend obtiene un *ID token* de Firebase y lo envía al backend.
3. El backend lo verifica, crea/actualiza el usuario y emite una **sesión propia**
   (JWT en cookie `httpOnly`) que se mantiene activa indefinidamente.
4. Al **cerrar sesión** se invalida la sesión (se incrementa `session_epoch`),
   de modo que para volver a entrar se requiere un nuevo OTP.

### Partidos y marcadores
- El backend consulta football-data.org **una vez al día** para conocer los
  partidos, y los **cachea** en la BD. Solo vuelve a consultar el resultado de un
  partido **después de que pasa su hora de inicio**, respetando los límites del
  plan gratuito.
- Las horas se guardan en UTC y se muestran en la **zona horaria del navegador**
  de cada usuario.
- Un marcador es **editable hasta que el partido empieza**; luego queda de solo
  lectura.

### Modelo de pronóstico
El usuario **solo ingresa marcadores**. A partir de esos marcadores la app
**deriva automáticamente** las tablas de posición de cada grupo (criterios de
desempate de la FIFA) y la clasificación. Así no hay formularios extra que
llenar.

### Pestañas
- **Marcadores:** muestra tu puesto global, una sección **Reglas** colapsible, y
  los marcadores en **orden cronológico**. Al entrar, la página hace scroll
  automático al **partido activo** (puedes subir para ver los anteriores y bajar
  para ver los siguientes).
- **Tablas:** compara, grupo por grupo, tu tabla de posiciones **derivada de tus
  marcadores** contra la tabla **real**. De aquí salen los puntos de la sección 2
  del reglamento.
- **Ranking:** posiciones de todos los jugadores, resaltando tu puesto.

### Puntuación
Implementada en [`backend/internal/scoring`](./backend/internal/scoring) como un
paquete puro y con tests, fiel al reglamento. Cubre tres pronósticos
independientes: **marcadores**, **posiciones de grupo** y **bracket**.

## Requisitos previos

- **Go** 1.23+ (backend)
- **Node** 20+ y npm (frontend)
- Una cuenta de **Google Cloud** con Firebase habilitado
- Una **API key** gratuita de [football-data.org](https://www.football-data.org/)

## Desarrollo local

> Instrucciones detalladas de configuración y despliegue (Firebase, GCP,
> GitHub Pages, variables de entorno) en **[DEPLOY.md](./DEPLOY.md)**.

```bash
# Backend
cd backend
cp .env.example .env        # completar credenciales
go test ./...               # ejecutar pruebas (incluye el motor de puntuación)
go run ./cmd/server

# Frontend
cd frontend
npm install
npm run dev
```

## Estado

Backend y frontend **completos y compilando**:

- ✅ Motor de puntuación (secciones 1-4) + motor de tablas FIFA, con tests
- ✅ Base de datos PostgreSQL con migraciones automáticas
- ✅ Auth Firebase + sesión JWT (persistente, invalidable al cerrar sesión)
- ✅ Cliente football-data + sincronización 1×día / post-inicio
- ✅ API REST + ranking con desempate
- ✅ Frontend React (Marcadores · Tablas · Ranking), accesible
- ✅ Dockerfile, GitHub Action de Pages y guía de despliegue

Pendiente de confirmación con el organizador (aislado y fácil de ajustar):
los **puntos de bracket** ("acertar los equipos que se enfrentan") y la
interpretación fina de los puntos de **posiciones de grupo**.
