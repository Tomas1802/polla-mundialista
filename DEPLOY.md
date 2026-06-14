# Guía de despliegue

Esta guía cubre todo lo que necesitas para poner la Polla en producción:
**Firebase** (login por celular), **football-data.org** (partidos), **Google
Cloud** (backend + base de datos) y **GitHub Pages** (frontend).

Hay tres cosas que solo tú puedes crear porque usan tus cuentas:
1. Un proyecto **Firebase / Google Cloud** (es el mismo proyecto).
2. Una **API key** de football-data.org.
3. El **repositorio en GitHub** con Pages activado.

---

## 1. Firebase (autenticación por celular)

1. Entra a <https://console.firebase.google.com/> y crea un proyecto (esto crea
   también el proyecto de Google Cloud con el mismo id).
2. **Authentication → Sign-in method → Phone**: actívalo.
3. **Project settings → General → Your apps → Web app (</>)**: registra una app
   web y copia el `firebaseConfig`. Necesitarás `apiKey`, `authDomain`,
   `projectId`, `appId` (van en las variables `VITE_FIREBASE_*`).
4. **Authentication → Settings → Authorized domains**: agrega
   `TU_USUARIO.github.io` (y `localhost` ya viene incluido) para que el login
   funcione desde GitHub Pages.
5. **Service account** (para que el backend verifique los tokens en local):
   *Project settings → Service accounts → Generate new private key*. Guarda el
   JSON y apunta `FIREBASE_CREDENTIALS_FILE` a él en tu `.env` local. En Cloud
   Run **no** necesitas el archivo (se usa la identidad del servicio).

> El plan gratuito de Firebase Phone Auth cubre un volumen bajo de SMS, ideal
> para una polla familiar. Como la sesión se mantiene hasta cerrarla, el SMS solo
> se envía al iniciar sesión.

---

## 2. football-data.org (partidos)

1. Regístrate gratis en <https://www.football-data.org/client/register>.
2. Copia tu token y ponlo en `FOOTBALL_DATA_TOKEN`.
3. El plan gratuito incluye el Mundial (`WC`) y permite ~10 llamadas/minuto. La
   app hace **una sincronización al día** más una tras el inicio de cada partido,
   así que va muy holgada.

---

## 3. Desarrollo local

**Base de datos** (con Docker):
```bash
docker run --name polla-db -e POSTGRES_USER=polla -e POSTGRES_PASSWORD=polla \
  -e POSTGRES_DB=polla -p 5432:5432 -d postgres:16
```

**Backend:**
```bash
cd backend
cp .env.example .env        # completa FIREBASE_PROJECT_ID, JWT_SECRET, etc.
go test ./...
go run ./cmd/server         # corre migraciones y arranca en :8080
```

**Frontend:**
```bash
cd frontend
cp .env.example .env.local  # completa VITE_API_BASE y VITE_FIREBASE_*
npm install
npm run dev                 # http://localhost:5173
```

---

## 4. Google Cloud (backend + base de datos)

Usa el mismo proyecto de Firebase. Habilita las APIs de Cloud Run y Cloud SQL.

**4.1 Cloud SQL (PostgreSQL):**
```bash
gcloud sql instances create polla-db --database-version=POSTGRES_16 \
  --tier=db-f1-micro --region=us-central1
gcloud sql databases create polla --instance=polla-db
gcloud sql users create polla --instance=polla-db --password=UNA_CLAVE_FUERTE
```
Anota el *connection name*: `PROJECT:REGION:INSTANCE`.

**4.2 Desplegar el backend a Cloud Run** (desde `backend/`):
```bash
gcloud run deploy polla-api \
  --source . \
  --region us-central1 \
  --allow-unauthenticated \
  --add-cloudsql-instances PROJECT:REGION:polla-db \
  --set-env-vars "FIREBASE_PROJECT_ID=TU_PROJECT_ID" \
  --set-env-vars "FOOTBALL_DATA_TOKEN=TU_TOKEN" \
  --set-env-vars "JWT_SECRET=UN_SECRETO_LARGO" \
  --set-env-vars "ALLOWED_ORIGIN=https://TU_USUARIO.github.io" \
  --set-env-vars "COOKIE_SECURE=true" \
  --set-env-vars "DATABASE_URL=postgres://polla:UNA_CLAVE_FUERTE@/polla?host=/cloudsql/PROJECT:REGION:polla-db"
```
- Cloud Run usa la identidad del servicio como credenciales de Firebase (ADC), así
  que **no** se necesita `FIREBASE_CREDENTIALS_FILE`. Dale al service account el rol
  *Firebase Authentication Viewer* si la verificación de tokens fallara.
- Guarda la URL que te devuelve (algo como `https://polla-api-xxxx.run.app`).

---

## 5. GitHub Pages (frontend)

1. Sube el repo a GitHub y ve a **Settings → Pages → Build and deployment →
   Source: GitHub Actions**.
2. En **Settings → Secrets and variables → Actions**:
   - **Variables**: `VITE_API_BASE` = la URL de Cloud Run,
     `VITE_FIREBASE_AUTH_DOMAIN`, `VITE_FIREBASE_PROJECT_ID`.
   - **Secrets**: `VITE_FIREBASE_API_KEY`, `VITE_FIREBASE_APP_ID`.
3. Haz push a `main`. El workflow `deploy-frontend.yml` construye y publica el
   sitio en `https://TU_USUARIO.github.io/REPO/`.
4. Verifica que ese dominio esté en los **Authorized domains** de Firebase (paso
   1.4) y que `ALLOWED_ORIGIN` del backend coincida con
   `https://TU_USUARIO.github.io`.

---

## Lista de verificación final

- [ ] Phone auth activado en Firebase y dominio de Pages autorizado.
- [ ] Backend en Cloud Run con `COOKIE_SECURE=true` y `ALLOWED_ORIGIN` correcto.
- [ ] `VITE_API_BASE` apunta a la URL de Cloud Run.
- [ ] `FOOTBALL_DATA_TOKEN` configurado (si no, no se cargan partidos).
- [ ] Cloud SQL conectado al servicio de Cloud Run.
