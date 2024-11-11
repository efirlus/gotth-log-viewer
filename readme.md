# GoTTH Log Viewer

![Sample Sreenshot.png](https://github.com/efirlus/gotth-log-viewer/blob/61f6bafe7cb06f340a22ecae1fc3a771d2bdc66d/Sample%20Sreenshot.png?raw=true)
A service-oriented log viewer built with Go, Templ, Tailwind, and HTMX (GoTTH stack). This project demonstrates the transformation of a React-based SPA into a server-side rendered application with real-time updates.

## Features

- 🔒 Secure authentication with scrypt-based password hashing
- 🔄 Real-time log monitoring with automatic updates
- 🔍 Advanced filtering capabilities (by program, log level, and search)
- 🎨 Clean, responsive UI with Catppuccin Mocha theme
- 📱 Mobile-friendly design
- 🚀 Fast server-side rendering with partial updates
- 💾 Efficient log parsing and caching
- 🌐 Almost Zero JavaScript framework dependencies
- 🔐 HTTPS support out of the box

## Architecture

### Service-Oriented Design

The application follows a clean, service-oriented architecture pattern:

```
internal/
├── auth/       # Authentication system
├── filters/    # Log filtering logic
├── handlers/   # HTTP request handlers
├── services/   # Core business logic
├── shared/     # Shared utilities
├── types/      # Type definitions
└── view/       # UI components (Templ)
```

Key architectural decisions:

1. **Security First**:
   - Secure authentication with scrypt hashing
   - HTTPS enforcement
   - Secure session management

2. **Separation of Concerns**:
   - `AuthService`: Handles authentication and session management
   - `LogService`: Handles log file reading, parsing, and caching
   - `FilterService`: Manages log filtering and sorting
   - `HandlerService`: Coordinates HTTP requests and responses

3. **Real-time Updates**:
   - Uses HTMX for seamless partial page updates
   - Polling mechanism with optimized cache layer
   - Server-side filtering to reduce network load

4. **Component-Based UI**:
   - Templ components for server-side rendering
   - Tailwind CSS for styling
   - HTMX for interactive updates

## Getting Started

### Prerequisites

- Go 1.21+
- TLS certificate and key for HTTPS
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
```

3. Create environment file:
```bash
cat > .env << EOL
LISTEN_ADDR=":5173" # Editable your own
LOG_PATH="/path/to/your/app.log"
CERT_FILE="/path/to/your/server.crt"
KEY_FILE="/path/to/your/server.key"
ALLOWED_ORIGINS="*" # Your own DDNS or IP
CRED_HASH="your_generated_hash_here"  # Generate using the provided tool
EOL
```

4. Generate credential hash:
```bash
go run cmd/hashgen/main.go username your_passphrase
```
Copy function from internal/auth/store.go, and make tempmain in other directory.
Copy the output hash to your .env file's CRED_HASH variable.

5. Build CSS:
```bash
make css-once
```

### Development

Start the development server with live reload:

```bash
make dev
```

This command runs:
- Templ template compiler with auto-reload
- Go server with hot reload (using Air)
- CSS compiler

### Production Setup

#### SystemD Service Setup

1. Create a service file:
```bash
sudo nano /etc/systemd/system/gotthlogviewer.service
```

2. Add the following content:
```ini
[Unit]
Description=GoTTH Log Viewer
After=network.target

[Service]
Type=simple
User=your_user
Group=your_group
WorkingDirectory=/path/to/gotthlogviewer
EnvironmentFile=/path/to/gotthlogviewer/.env
ExecStart=/path/to/gotthlogviewer/gotthlogviewer
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
```

3. Set appropriate permissions:
```bash
sudo chmod 644 /etc/systemd/system/gotthlogviewer.service
sudo chmod 600 /path/to/gotthlogviewer/.env
```

4. Start and enable the service:
```bash
sudo systemctl daemon-reload
sudo systemctl enable gotthlogviewer
sudo systemctl start gotthlogviewer
```

## Key Components

### Authentication System

Secure authentication implementation:

```go
type Credential struct {
    Username string
    Hash     []byte // scrypt hash of username_passphrase
}
```

Features:
- Scrypt-based password hashing
- Secure session management
- HTTPS enforcement
- Protection against timing attacks

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

## Security Notes

1. Always use HTTPS in production
2. Store the .env file securely with restricted permissions (chmod 600)
3. Keep your credential hash secure and never commit it to version control
4. Regularly update your TLS certificates
5. Consider implementing rate limiting for the login endpoint in production

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
