# 📁 DNS Inventory Server - Project Structure

## Directory Layout

```
dns-inventory-server/
├── README.md                    # Comprehensive project documentation
├── QUICKSTART.md               # 5-minute setup guide
├── LICENSE                     # MIT license
├── .gitignore                 # Git ignore patterns
├── .env.example               # Configuration template
├── go.mod                     # Go module definition
├── main.go                    # Main server entry point
│
├── cmd/                       # Application entry points
│   └── server/                
│       └── enhanced_main.go   # Legacy entry point (use main.go)
│
├── internal/                  # Private application code
│   ├── api/                   # External API clients
│   │   ├── cloudflare.go     # Cloudflare API integration
│   │   └── godaddy.go        # GoDaddy API integration
│   │
│   ├── config/                # Configuration management
│   │   └── config.go         # Environment variable handling
│   │
│   ├── database/              # Data persistence layer
│   │   ├── filedb.go         # File-based database operations
│   │   └── models.go         # Data models and structures
│   │
│   ├── handlers/              # HTTP request handlers
│   │   ├── common.go         # Shared handler utilities
│   │   ├── dns.go            # DNS records endpoints
│   │   ├── domain.go         # Domain management endpoints
│   │   ├── enhanced_migration.go  # Migration wizard
│   │   ├── export.go         # Data export functionality
│   │   └── user.go           # User management endpoints
│   │
│   └── services/              # Business logic layer
│       ├── dns.go            # DNS operations service
│       ├── domain.go         # Domain operations service
│       ├── enhanced_migration.go  # Migration processing
│       ├── notification.go   # Email notification service
│       └── user.go           # User management service
│
├── web/                       # Frontend assets
│   ├── static/                # Static files (CSS, JS, images)
│   │   ├── css/
│   │   │   └── app.css       # Main stylesheet
│   │   └── js/
│   │       └── app.js        # Frontend JavaScript
│   │
│   └── templates/             # HTML templates
│       ├── layout.html       # Base layout template
│       ├── domains.html      # Domain management page
│       ├── dns.html          # DNS records page
│       ├── users.html        # User management page
│       └── enhanced_migration.html  # Migration wizard
│
├── data/                      # Database files (auto-created)
│   ├── domains.json          # Domain records
│   ├── dns_records.json      # DNS record data
│   ├── users.json            # User accounts
│   ├── domain_assignments.json   # Domain-to-user mappings
│   ├── dns_assignments.json  # DNS-to-user mappings
│   ├── migration_jobs.json   # Migration job tracking
│   ├── migration_templates.json  # Migration templates
│   └── snapshots.json        # Historical data snapshots
│
├── scripts/                   # Utility scripts
│   ├── build.bat             # Windows build script
│   ├── build.sh              # Linux/macOS build script
│   ├── run.bat               # Windows run script
│   ├── run.sh                # Linux/macOS run script
│   └── install.sh            # Production installation script
│
└── docs/                      # Documentation
    └── PROJECT_STRUCTURE.md  # This file
```

## Key Components

### 🎯 **Entry Points**
- **`main.go`**: Primary server entry point with improved architecture
- **`cmd/server/enhanced_main.go`**: Legacy entry point (maintained for compatibility)

### 🧩 **Core Modules**

#### **API Layer** (`internal/api/`)
- **GoDaddy Client**: Domain and DNS management via GoDaddy API
- **Cloudflare Client**: Zone and record management via Cloudflare API
- **Connection Testing**: Automatic API health checks
- **Rate Limiting**: Built-in throttling to respect API limits

#### **Configuration** (`internal/config/`)
- **Environment Variables**: Flexible configuration via `.env` file
- **Validation**: Configuration validation with helpful error messages
- **Defaults**: Sensible defaults for all optional settings

#### **Database Layer** (`internal/database/`)
- **File-Based Storage**: JSON files for zero-dependency operation
- **Concurrent Access**: Thread-safe operations with mutex protection
- **Data Models**: Comprehensive structures for all data types
- **Migration Support**: Schema evolution and data migration capabilities

#### **HTTP Handlers** (`internal/handlers/`)
- **RESTful API**: Clean HTTP endpoints for all operations
- **Web Interface**: Server-side rendering for admin interface
- **File Upload**: Multipart form handling for migration files
- **Error Handling**: Consistent error responses and logging

#### **Business Services** (`internal/services/`)
- **Domain Service**: Domain lifecycle management and provider integration
- **DNS Service**: DNS record operations and synchronization
- **User Service**: User management and assignment operations
- **Migration Service**: Advanced batch processing with resume capability
- **Notification Service**: AWS SES email integration with smart templating

### 🎨 **Frontend** (`web/`)

#### **Templates**
- **Modular Design**: Template inheritance with shared layout
- **Responsive UI**: Mobile-friendly interface with Tailwind CSS styling
- **Real-time Updates**: JavaScript components for dynamic content
- **Progressive Enhancement**: Works with and without JavaScript

#### **Static Assets**
- **Optimized CSS**: Single compiled stylesheet with utility classes
- **Modern JavaScript**: ES6+ features with browser compatibility
- **Asset Caching**: Proper cache headers for static resources

### 📊 **Data Storage** (`data/`)
- **JSON Format**: Human-readable data files for easy debugging
- **Atomic Operations**: Safe concurrent access with file locking
- **Backup Ready**: Simple file-based backup and restore
- **Performance**: Optimized for datasets up to 100K+ records

### 🔧 **Build System** (`scripts/`)
- **Cross-Platform**: Scripts for Windows, Linux, and macOS
- **Production Ready**: Optimized builds with proper flags
- **Zero Dependencies**: No build tools or package managers required
- **Easy Deployment**: One-command installation for production servers

## Architecture Principles

### 🏗️ **Clean Architecture**
- **Separation of Concerns**: Clear boundaries between layers
- **Dependency Injection**: Services are injected into handlers
- **Interface Boundaries**: APIs defined by interfaces, not implementations
- **Testability**: Each layer can be tested independently

### 📦 **Zero Dependencies**
- **Standard Library Only**: No external Go packages required
- **Self-Contained**: Everything needed is included in the repository
- **Portable**: Runs on any system with Go 1.21+
- **Secure**: No third-party security vulnerabilities

### ⚡ **Performance**
- **Concurrent Operations**: Multi-threaded processing where beneficial
- **Memory Efficient**: Streaming operations for large datasets
- **Caching**: In-memory caching for frequently accessed data
- **Resource Management**: Proper cleanup and garbage collection

### 🔒 **Security**
- **Input Validation**: All user input is validated and sanitized
- **File Safety**: Secure file upload and processing
- **Error Handling**: No sensitive information leaked in errors
- **Access Control**: Foundation for role-based access control

## Development Workflow

### 🚀 **Getting Started**
```bash
# 1. Clone/download the project
# 2. Build and run
./scripts/build.sh && ./dns-inventory-server

# 3. Access the interface
open http://localhost:8080
```

### 🔧 **Development**
```bash
# Run directly (for development)
go run main.go

# Build optimized binary
go build -ldflags="-s -w" -o dns-inventory-server main.go

# Test all functionality
go test ./...
```

### 📦 **Deployment**
```bash
# Production installation (Linux)
sudo ./scripts/install.sh

# Manual deployment
cp dns-inventory-server /opt/dns-inventory/
cp -r web /opt/dns-inventory/
cp .env.example /opt/dns-inventory/.env
```

## File Responsibilities

### **Configuration**
- **`.env.example`**: Template with all configuration options
- **`internal/config/config.go`**: Configuration loading and validation

### **Data Models**
- **`internal/database/models.go`**: All data structures and types
- **`data/*.json`**: Actual data storage files

### **Business Logic**
- **`internal/services/*.go`**: Core business operations
- **`internal/api/*.go`**: External API communication

### **Web Interface**
- **`internal/handlers/*.go`**: HTTP request/response handling
- **`web/templates/*.html`**: HTML page templates
- **`web/static/`**: CSS, JavaScript, and other assets

### **Utilities**
- **`scripts/`**: Build, run, and deployment automation
- **`docs/`**: Comprehensive documentation

## Best Practices

### 🎯 **Code Organization**
- **Package by Feature**: Related functionality grouped together
- **Clear Naming**: Self-documenting function and variable names
- **Consistent Style**: Following Go best practices throughout
- **Minimal Interfaces**: Small, focused interfaces

### 📝 **Documentation**
- **README First**: Comprehensive project documentation
- **Code Comments**: Explain why, not what
- **API Documentation**: Clear endpoint documentation
- **Example Usage**: Working examples in all docs

### 🧪 **Quality Assurance**
- **Error Handling**: Comprehensive error handling throughout
- **Input Validation**: All inputs validated at entry points
- **Resource Management**: Proper cleanup of resources
- **Performance Monitoring**: Built-in metrics and logging

---

This structure provides a solid foundation for DNS asset management while maintaining simplicity and avoiding external dependencies.