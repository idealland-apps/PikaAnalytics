# PikaAnalytics 📊 — Self-Hosted Website Analytics

A lightweight, self-hosted website analytics tool built with Go and React. Configure
your sites in the admin console, embed a one-line tracking script, and inspect traffic,
referrers, devices, and locations from a clean Material-UI dashboard.

## ✨ Features

- **Self-hosted, lightweight** — single Go binary + SQLite, no external services required
- **Site management** — create / edit / delete tracked sites with unique site keys
- **One-line embed** — copy a `<script>` tag from the admin and paste into your pages
- **Privacy-friendly tracking script** — uses `navigator.sendBeacon` when available, no cookies
- **Server-side enrichment** — User-Agent parsing for browser/OS/device, optional IP geolocation
- **Dashboard & analytics** — page views, unique visitors, visits trend, top pages,
  referrers, browsers, OS, and countries (powered by Chart.js)
- **Access logs** — drill into recent page views with site/date/search filters

## 🚀 Quick Start

```bash
docker run -d \
  --name pikaanalytics \
  -p 8080:8080 \
  -v pikaanalytics_data:/app/data \
  -e DATA_PATH=/app/data \
  --restart unless-stopped \
  bytetopia/pikaanalytics:latest
```

Open <http://localhost:8080/admin/> and log in with `admin` / `admin123` (change the
password right away).

## 🔌 Tracking a Site

1. Log into the admin console and go to **Sites → Add Site**.
2. Pick a unique **Site Key** (letters, digits, `-`, `_`).
3. Click the code icon next to the new site, copy the `<script>` tag, and paste it
   into the `<head>` of every page you want tracked:

   ```html
   <script async src="https://your-pikaanalytics-host/pulse.js?site=YOUR_SITE_KEY"></script>
   ```

4. Visit the page once to verify the event lands under **Visits**.

## ⚙️ Configuration

| Variable          | Required | Default | Description                                                |
|-------------------|----------|---------|------------------------------------------------------------|
| `DATA_PATH`       | No       | `.`     | Directory for the SQLite database                          |
| `GEO_IP_DB_PATH`  | No       | —       | Directory containing GeoLite2 `.mmdb` files for geolocation |
| `CORS_ORIGINS`    | No       | `*`     | Comma-separated list of allowed origins (defaults to all)  |

### Optional: GeoLite2 for country/city enrichment

1. Sign up at [maxmind.com/en/geolite2/signup](https://www.maxmind.com/en/geolite2/signup)
2. Download `GeoLite2-City.mmdb` (and optionally `GeoLite2-ASN.mmdb`)
3. Mount the directory into the container and set `GEO_IP_DB_PATH`:

   ```bash
   docker run -d --name pikaanalytics \
     -p 8080:8080 \
     -v pikaanalytics_data:/app/data \
     -v $(pwd)/geolite2:/app/geolite:ro \
     -e DATA_PATH=/app/data \
     -e GEO_IP_DB_PATH=/app/geolite \
     bytetopia/pikaanalytics:latest
   ```

## 🐳 Docker Compose

```bash
docker-compose up -d
```

See [docker-compose.yml](./docker-compose.yml) for a ready-to-edit template.

## 🔧 Development

### Backend (Go)

```bash
cd backend
go mod tidy
go run main.go         # serves on :8080
```

### Frontend (React)

```bash
cd frontend
npm install
npm start              # dev server on http://localhost:3000/admin/
```

The dev frontend talks to `http://localhost:8080/api` by default; override with
`REACT_APP_API_URL` if needed.

### Production build

```bash
./build.sh             # or build.ps1 on Windows
```

Outputs a `dist/` directory containing the binary and the built frontend.

## 🗄️ Data Model

- `users` — admin accounts
- `sites` — tracked sites (name, unique site_key, optional domain/description)
- `site_configs` — per-site key/value settings
- `page_views` — collected events (path, referrer, UA-derived browser/OS/device, IP, country, …)

All tables live in a single SQLite database at `${DATA_PATH}/pikaanalytics.db`.

## 📡 API Overview

Public:

- `POST /api/login`
- `GET  /pulse.js?site=SITE_KEY`
- `POST /api/pulse`

Authenticated (Bearer JWT):

- `GET/POST/PUT/DELETE /api/sites[/:id]`
- `GET /api/analytics/overview`
- `GET /api/analytics/pages`
- `GET /api/analytics/referrers`
- `GET /api/analytics/devices`
- `GET /api/analytics/locations`
- `GET /api/analytics/visits`
- `GET /api/analytics/recent`
- `POST /api/change-password`
- `GET /api/version`

All `/api/analytics/*` endpoints accept optional `site_key` (or `site_id`) and `range`
(in days) query parameters.

## 📄 License

MIT
