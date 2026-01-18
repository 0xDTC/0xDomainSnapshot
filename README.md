# 0xDomainSnapshot

A comprehensive cloud asset inventory dashboard for tracking resources across multiple cloud providers including AWS, Azure, GCP, Cloudflare, DigitalOcean, Oracle Cloud, and Alibaba Cloud.

## Features

- **Multi-Cloud Support**: Track assets across AWS, Azure, GCP, Cloudflare, DigitalOcean, Oracle, and Alibaba Cloud
- **100+ Services**: Comprehensive coverage of compute, storage, database, networking, and security services
- **Domain Management**: Monitor domains and DNS records
- **Removed Assets Tracking**: Keep history of deleted/inactive resources with date filtering
- **Global Search**: Search across all providers and services at once
- **Account Filter**: Filter assets by cloud account
- **Dark/Light Theme**: Toggle between dark and light themes with preference persistence
- **Column Visibility**: Show/hide columns as needed
- **Sortable & Filterable Tables**: Find assets quickly with column filters
- **CSV Export**: Export any asset list to CSV
- **URL State Persistence**: Filters, sorting, and pagination state saved in URL
- **Responsive Design**: Works on desktop and mobile devices
- **No Backend Required**: Pure static HTML/JS/CSS (backend collectors coming soon)

## Directory Structure

```
0xDomainSnapshot/
├── index.html              # Dashboard homepage
├── search.html             # Global search page
├── server.go               # Go HTTP server
│
├── dns/                    # Domain & DNS pages
│   ├── domains.html
│   └── subdomains.html
│
├── aws/                    # AWS services (51 services)
│   ├── ec2.html
│   ├── eks.html
│   ├── lambda.html
│   ├── rds.html
│   ├── s3.html
│   └── ...
│
├── azure/                  # Azure services (25 services)
│   ├── vms.html
│   ├── aks.html
│   ├── functions.html
│   └── ...
│
├── gcp/                    # GCP services (22 services)
│   ├── compute.html
│   ├── gke.html
│   ├── functions.html
│   └── ...
│
├── cloudflare/             # Cloudflare services (11 services)
│   ├── workers.html
│   ├── pages.html
│   └── ...
│
├── digitalocean/           # DigitalOcean services (5 services)
│   ├── droplets.html
│   └── ...
│
├── oracle/                 # Oracle Cloud services (4 services)
│   ├── compute.html
│   └── ...
│
├── alibaba/                # Alibaba Cloud services (4 services)
│   ├── ecs.html
│   └── ...
│
├── removed/                # Removed assets pages (per provider)
│   ├── aws-removed.html
│   ├── azure-removed.html
│   └── ...
│
├── static/
│   ├── css/
│   │   └── style.css       # Theme styles (dark/light)
│   └── js/
│       ├── datatable.js    # DataTable component
│       └── nav.js          # Navigation component
│
├── data/                   # JSON data files
│   ├── domains.json
│   ├── subdomains.json
│   ├── metadata.json       # Service update timestamps
│   ├── aws/
│   ├── azure/
│   ├── gcp/
│   ├── cloudflare/
│   ├── digitalocean/
│   ├── oracle/
│   └── alibaba/
│
└── OLD/                    # Archived original scripts
```

## Quick Start

### Option 1: Go Server (Recommended)

```bash
# Build and run
go run server.go

# Or with custom port
go run server.go -port 3000

# Then open: http://localhost:8080
```

### Option 2: Python Server

```bash
python -m http.server 8080
# Then open: http://localhost:8080
```

### Option 3: VS Code Live Server

Install the "Live Server" extension and click "Go Live".

## Data Format

All data files are JSON arrays. Each asset should have:

```json
{
  "account": "prod-account",
  "name": "resource-name",
  "status": "active",
  "discovery_date": "2025-01-15",
  "last_seen": "2025-01-16"
}
```

### Required Fields

| Field | Description |
|-------|-------------|
| `account` | Cloud account/subscription name |
| `status` | `active` or `removed` |

### Status Values

- `active` - Currently active/running
- `removed` - Deleted or no longer present

## Key Features

### Account Filter
Each table has an account dropdown to filter by specific cloud accounts. Useful when managing multiple accounts.

### Global Search
Access via the magnifying glass icon in the navbar. Searches across all providers and services simultaneously.

### Theme Toggle
Click the moon/sun icon in the navbar to switch between dark and light themes. Preference is saved in localStorage.

### Column Visibility
Click the "Columns" button to show/hide specific columns. State is persisted in the URL.

### URL State
All filter, sort, and pagination state is saved in the URL, making it easy to share specific views.

## Adding Your Data

1. Replace the sample JSON files in `data/` with your actual asset data
2. Ensure each record has an `account` field and `status` field
3. Match the JSON structure to the column definitions in each HTML page
4. Refresh the browser to see your data

## Metadata (Optional)

Create `data/metadata.json` to show service update status on the dashboard:

```json
{
  "services": {
    "aws": {
      "name": "AWS",
      "schedule": "hourly",
      "last_updated": "2025-01-18T10:30:00Z"
    }
  }
}
```

## Browser Support

- Chrome (recommended)
- Firefox
- Safari
- Edge

## License

MIT
