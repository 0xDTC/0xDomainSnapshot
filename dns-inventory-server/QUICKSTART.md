# 🚀 Quick Start Guide - DNS Inventory Server

Get DNS Inventory running in under 5 minutes!

## Option 1: Zero Configuration Start

```bash
# Windows
scripts\build.bat
dns-inventory-server.exe

# Linux/macOS
./scripts/build.sh
./dns-inventory-server
```

🌐 **Open**: http://localhost:8080

That's it! You now have:
- ✅ Working web interface
- ✅ Migration functionality  
- ✅ User management
- ✅ Export capabilities

## Option 2: With Data Collection

1. **Create configuration file:**
   ```bash
   cp .env.example .env
   ```

2. **Add API credentials to `.env`:**
   ```env
   GODADDY_API_KEY=your_key_here
   GODADDY_API_SECRET=your_secret_here
   CLOUDFLARE_API_TOKEN=your_token_here
   ```

3. **Start server:**
   ```bash
   # Windows: scripts\run.bat
   # Linux/macOS: ./scripts/run.sh
   ```

## Option 3: Full Enterprise Setup

Add to `.env`:
```env
# API Credentials
GODADDY_API_KEY=your_key
GODADDY_API_SECRET=your_secret
CLOUDFLARE_API_TOKEN=your_token

# Email Notifications
AWS_SES_REGION=us-east-1
AWS_SES_ACCESS_KEY=your_access_key
AWS_SES_SECRET_KEY=your_secret_key
NOTIFICATION_FROM_EMAIL=noreply@yourcompany.com
NOTIFICATION_TO_EMAIL=admin@yourcompany.com
```

## 🎯 What You Can Do Now

### 📊 **Dashboard** - http://localhost:8080
- System overview and statistics
- Real-time API connection status
- Quick access to all features

### 🌐 **Domain Management** - http://localhost:8080/domains
- View domains from all providers
- Assign domains to team members
- Export to CSV for analysis

### 📡 **DNS Records** - http://localhost:8080/dns  
- Complete DNS record inventory
- Filter by domain, type, provider
- Bulk operations and exports

### 👥 **User Management** - http://localhost:8080/users
- Create and manage team members
- Organize into groups (admin, security, ops)
- View assignments and workload

### 🔄 **Migration Wizard** - http://localhost:8080/migration
- Upload JSON files (domains, DNS records)
- 4-step guided process
- Real-time progress tracking
- Resume failed migrations

## 📁 Common Migration Tasks

### Import Domain List
1. Go to `/migration`
2. Upload your `domains.json` file
3. Select "Domain Import" template
4. Configure batch size (50-100 recommended)
5. Choose duplicate strategy (usually "merge")
6. Execute and monitor progress

### Import DNS Records
1. Upload `dns_records.json`
2. Select "DNS Records" template  
3. Enable domain validation if needed
4. Assign to users during import
5. Monitor real-time progress

## 🔧 Configuration Tips

### **Performance Tuning**
```env
# For large datasets
MIGRATION_BATCH_SIZE=50
DATA_COLLECTION_INTERVAL=120

# For high traffic
MAX_CONCURRENT_REQUESTS=200
```

### **Security Settings**
```env
API_RATE_LIMIT=1000
SESSION_TIMEOUT=7200
```

## 🆘 Troubleshooting

### **"Port already in use"**
```env
SERVER_PORT=8081
```

### **"Template error"**
- Check file permissions in `web/templates/`
- Verify working directory is correct

### **"API connection failed"**
- Verify credentials in `.env`
- Check network connectivity
- Review API rate limits

### **Migration stuck**
```bash
# Check server logs for errors
# Reduce batch size if memory issues
# Ensure sufficient disk space
```

## 📚 Next Steps

- **Production Deployment**: Use `scripts/install.sh` (Linux)
- **Add SSL/TLS**: Configure reverse proxy (Nginx)
- **Monitoring**: Set up logging and alerts
- **Backup**: Regular export of data directory
- **API Integration**: Use REST API for automation

## 🔗 Useful Links

- **Health Check**: http://localhost:8080/health
- **API Documentation**: http://localhost:8080/api/ (shows available endpoints)
- **Export Functions**: 
  - Domains CSV: http://localhost:8080/export/domains.csv
  - DNS CSV: http://localhost:8080/export/dns.csv

---

🎉 **You're all set!** DNS Inventory is now running and ready to manage your DNS assets.