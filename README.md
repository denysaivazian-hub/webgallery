# Web Gallery

A small project with a Go HTTP server and an empty React SPA.

## Project structure

```text
webgallery/
├── server/       Go API
├── web-client/   React SPA powered by Vite
└── docker-compose.yaml
```

## Start the Go server with Docker

Build and start the service:

```bash
docker compose up --build
```

The API is available at <http://localhost:8090>.

The development container uses Air. After the initial build, saving a `.go`
file automatically recompiles and restarts the server. Follow its output with:

```bash
docker compose logs -f server
```

Stop the service with:

```bash
docker compose down
```

## Run the Go server without Docker

Go 1.25 or newer is required.

```bash
cd server
go run .
```

The server reads these optional environment variables:

| Variable | Default | Purpose |
| --- | --- | --- |
| `PORT` | `8090` | HTTP listening port |
| `UPLOAD_DIR` | `files` | Directory used for uploaded PNG files |

## Run Go tests

From the repository root:

```bash
cd server
go test ./...
```

Run tests with the race detector:

```bash
go test -race ./...
```

Run static analysis:

```bash
go vet ./...
```

Run one test by name:

```bash
go test -run TestFileUpload -v
```

If the development container is already running, tests can also run inside it:

```bash
docker compose exec server go test ./...
```

Current tests cover valid uploads, missing files, malformed multipart requests,
oversized requests, CORS preflight behavior, file persistence, and unknown routes.

## API

### `GET /`

Returns a test response:

```json
{"message":"test"}
```

### `POST /file`

Accepts a `multipart/form-data` request whose file field is named `image`.
Only valid PNG images are accepted. The total request limit is 10 MiB.

```bash
curl -D - -F "image=@photo.png" http://localhost:8090/file --output response.png
```

The validated image is saved in `UPLOAD_DIR` with a unique name such as
`upload-123456789.png`. The same PNG is returned in the response body, and the
generated filename is included in the `X-Stored-Filename` response header.

Browser requests are allowed from `http://localhost:5173` by the CORS policy.

## Start the React client

Install dependencies and start Vite:

```bash
cd web-client
npm install
npm run dev
```

The SPA is available at <http://localhost:5173>.

Create a production build with:

```bash
npm run build
```
