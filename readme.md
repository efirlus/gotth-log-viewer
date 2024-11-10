# GoTTH Log Viewer

A service-oriented log viewer built with Go, Templ, Tailwind, and HTMX (GoTTH stack). This project demonstrates the transformation of a React-based SPA into a server-side rendered application with real-time updates.

## Features

- 🔄 Real-time log monitoring with automatic updates
- 🔍 Advanced filtering capabilities (by program, log level, and search)
- 🎨 Clean, responsive UI with Catppuccin Mocha theme
- 📱 Mobile-friendly design
- 🚀 Fast server-side rendering with partial updates
- 💾 Efficient log parsing and caching
- 🌐 Almost Zero JavaScript framework dependencies

## Architecture

### Service-Oriented Design

The application follows a clean, service-oriented architecture pattern:

```
internal/
├── filters/     # Log filtering logic
├── handlers/    # HTTP request handlers
├── services/    # Core business logic
├── shared/      # Shared utilities
├── types/       # Type definitions
└── view/        # UI components (Templ)
```

Key architectural decisions:

1. **Separation of Concerns**:
   - `LogService`: Handles log file reading, parsing, and caching
   - `FilterService`: Manages log filtering and sorting
   - `HandlerService`: Coordinates HTTP requests and responses

2. **Real-time Updates**:
   - Uses HTMX for seamless partial page updates
   - Polling mechanism with optimized cache layer
   - Server-side filtering to reduce network load

3. **Component-Based UI**:
   - Templ components for server-side rendering
   - Tailwind CSS for styling
   - HTMX for interactive updates

## Getting Started

### Prerequisites

- Go 1.21+
- Node.js (for Tailwind CSS)
- `templ` CLI tool

### Installation

1. Clone the repository:
```bash
git clone https://github.com/yourusername/gotth-logviewer.git
cd gotth-logviewer
```

2. Install dependencies:
```bash
go mod download
npm install
```

3. Build CSS:
```bash
make css-once
```

### Development

Start the development server with live reload:

```bash
make dev
```

This command runs:
- Tailwind CSS watcher
- Templ template compiler
- Go server with hot reload (using Air)

### Configuration

Environment variables (`.env`):
```
LISTEN_ADDR=":5174"  # Server port
```

### Production Build

1. Build CSS:
```bash
make css-once
```

2. Build Go binary:
```bash
go build -o logviewer ./cmd/server
```

## Key Components

### LogService

The core service managing log operations:

```go
type LogService struct {
    filepath    string
    cache       []types.LogEntry
    lastRead    time.Time
    lastModTime time.Time
    mu          sync.RWMutex
    onChange    func([]types.LogEntry)
}
```

Features:
- Efficient log parsing and caching
- File change detection
- Thread-safe operations
- Change notification system

### Log Filtering

Implements flexible filtering logic:

```go
type LogFilters struct {
    Program string
    Level   string
    Search  string
}
```

Supports:
- Program-based filtering
- Log level filtering
- Full-text search
- Combined filters

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/AmazingFeature`)
3. Commit your changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- [Templ](https://github.com/a-h/templ) for the templating system
- [HTMX](https://htmx.org/) for the interactive capabilities
- [Tailwind CSS](https://tailwindcss.com/) for styling
- [Catppuccin](https://github.com/catppuccin/catppuccin) for the color scheme