# ECLAT

## Project Overview

**Eclat** is a desktop asset manager designed for creative professionals. It helps organize, view, and manage various digital assets such as 3D models, textures, and images.

Built with **Go** (backend) and **React** (frontend) using the **Wails** framework, it combines high-performance system integration with a modern, reactive user interface.

### Key Features
*   **Asset Scanning:** Recursively scans configured directories for supported file types (Images, 3D Models, Textures).
*   **Smart Recognition:** Detects duplicates via file hashing and groups related files (e.g., texture sets) using heuristic name matching.
*   **Metadata Extraction:** Automatically extracts metadata like dimensions, bit depth, and calculates dominant colors.
*   **Thumbnail Generation:** Generates thumbnails for a wide range of formats including `.blend`, `.psd`, `.fbx`, `.exr`, etc.
*   **Real-time Watching:** Monitors file system changes to keep the asset library in sync.

## Architecture

### Backend (Go)
*   **Framework:** [Wails v2](https://wails.io/)
*   **Database:** SQLite (using `modernc.org/sqlite` pure Go driver).
*   **Data Access:** Type-safe Go code generated from SQL queries using **sqlc**.
*   **Core Services:**
    *   `Scanner`: Handles file walking, hashing, duplicate detection, and batch DB updates.
    *   `Watcher`: Listens for file system events (create, modify, delete).
    *   `ThumbnailGenerator`: Generates previews for assets.
    *   `SettingsService`: Manages application configuration.

### Frontend (React/TypeScript)
*   **Build Tool:** Vite
*   **Language:** TypeScript
*   **UI Framework:** [HeroUI](https://www.heroui.com/) (formerly NextUI) + **Tailwind CSS v4**.
*   **State Management:** Zustand
*   **Data Fetching:** TanStack Query (React Query)
*   **Routing:** React Router
*   **Package Manager:** Bun

## Directory Structure

*   `main.go`: Application entry point.
*   `wails.json`: Wails project configuration.
*   `frontend/`: React source code.
    *   `src/`: Components, hooks, stores, and services.
    *   `wailsjs/`: Auto-generated bindings for Go methods.
*   `internal/`: Private Go code.
    *   `app/`: Main application logic and lifecycle management.
    *   `scanner/`: File scanning and asset processing logic.
    *   `database/`: Generated database code.
    *   `config/`: Configuration structs and defaults.
*   `sql/`: Database schema and queries.
    *   `schema/`: SQL migration files.
    *   `queries/`: SQL queries for `sqlc`.

## Development

### Prerequisites
*   **Go** 1.21+
*   **Node.js** & **Bun** (for frontend package management)
*   **Wails CLI**: `go install github.com/wailsapp/wails/v2/cmd/wails@latest`

### Common Commands

| Task | Command | Description |
| :--- | :--- | :--- |
| **Start Dev Server** | `wails dev` | Runs the app with hot-reload for both Go and React. |
| **Build Production** | `wails build` | Compiles the binary for the current platform. |
| **Frontend Install** | `cd frontend && bun install` | Install frontend dependencies. |
| **Generate DB Code** | `sqlc generate` | Regenerate Go code from SQL queries (requires `sqlc` installed). |
| **Lint Frontend** | `cd frontend && bun run lint` | Lint/Fix frontend code using Biome. |

### Development Guidelines
*   **Database:** Modifying `sql/queries/*.sql` or `sql/schema/*.sql` requires running `sqlc generate` to update the Go code.
*   **Wails Bindings:** Adding public methods to the `App` struct (or other bound structs) requires a rebuild to update `frontend/wailsjs` so after every change that might affect bindings run shell Command `wails generate module`.
*   **Styling:** Follow the existing Tailwind CSS v4 patterns. Use HeroUI components where possible for consistency.
*   **Idiomatic GO:** Follow the Idiomatic approach to writing a backend GOLANG code. Try minimalise the use of external package.
*   **Testing:** After each major change instead of rebulding project try run test wit the command go run test './...'.
