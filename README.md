# NMAP Dashboard - Red Team Command Center

A web-based dashboard for managing and visualizing nmap scan results across multiple teams in cyber defense competitions. Features real-time scan tracking, user management, and a clean red team themed interface.

## Table of Contents
- [Features](#features)
- [Prerequisites](#prerequisites)
- [Installation](#installation)
- [Configuration](#configuration)
- [Setting Up Teams](#setting-up-teams)
- [User Management](#user-management)
- [Connecting Scanners](#connecting-scanners)
- [API Documentation](#api-documentation)
- [Running in Production](#running-in-production)
- [Troubleshooting](#troubleshooting)

---

## Features

- **Multi-team support** - Track scans across multiple competition teams
- **Real-time dashboard** - View hosts, ports, and scan status
- **User management** - Role-based access (Admin, Scanner, Viewer)
- **Job tracking** - Monitor scan jobs and their status
- **Dangerous port highlighting** - Automatically flag risky services
- **Host tracking** - Track online/offline status over time
- **REST API** - Full API with Swagger documentation
- **Dark theme** - Red/green accent colors for red team aesthetic

---

## Prerequisites

### Operating System
- Linux (tested on Ubuntu 22.04/24.04, Debian 12)

### Required Software

1. **Go 1.21 or later**
```bash
# Check if installed
go version

# If not installed (Ubuntu/Debian):
wget https://go.dev/dl/go1.21.5.linux-amd64.tar.gz
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf go1.21.5.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc
```

2. **GCC (for SQLite)**
```bash
sudo apt update && sudo apt install -y build-essential
```

---

## Installation

### Step 1: Extract the Dashboard

```bash
# Create directory
cd /opt

# Extract (assuming zip is in current directory)
sudo unzip nmap-dashboard-improved.zip
sudo mv nmap-dashboard-improved RedBoard
cd /opt/RedBoard

# Set ownership (optional, for non-root operation)
sudo chown -R $USER:$USER /opt/RedBoard
```

### Step 2: Configure Environment

```bash
# Copy example config
cp env_example .env

# Edit configuration
nano .env
```

**Recommended `.env` configuration:**
```bash
# Server settings
GIN_MODE=release
PORT=8080

# Security - CHANGE THESE!
SESSION_SECRET=your-random-32-char-string-here
ADMIN_PASSWORD=YourSecureAdminPassword123

# Database (SQLite by default)
DB_PATH=./data/dashboard.db
```

**Generate a secure session secret:**
```bash
openssl rand -base64 32
```

### Step 3: Build the Dashboard

```bash
# Download dependencies
go mod tidy

# Build the binary
go build -o Redteam-Dashboard-go

# Verify it built (ignore SQLite warnings)
ls -la Redteam-Dashboard-go
```

### Step 4: Initialize the Database

The database is created automatically on first run. The default admin account is:
- **Username:** `admin`
- **Password:** Value of `ADMIN_PASSWORD` in `.env` (default: `changeme`)

### Step 5: Start the Dashboard

```bash
# Run directly
./Redteam-Dashboard-go

# Or run in background
nohup ./Redteam-Dashboard-go > dashboard.log 2>&1 &
```

### Step 6: Access the Dashboard

Open your browser and navigate to:
```
http://YOUR_SERVER_IP:8080/login.html
```

Login with:
- Username: `admin`
- Password: (your ADMIN_PASSWORD from .env)

---

## Configuration

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8080` | HTTP port to listen on |
| `GIN_MODE` | `debug` | Set to `release` for production |
| `SESSION_SECRET` | (random) | Secret for session cookies - SET THIS! |
| `ADMIN_PASSWORD` | `changeme` | Initial admin password - CHANGE THIS! |
| `DB_PATH` | `./data/dashboard.db` | SQLite database path |
| `API_BASE_URL` | `` | Base URL for API (usually leave empty) |

### Example Production `.env`

```bash
GIN_MODE=release
PORT=8080
SESSION_SECRET=aVeryLongRandomStringAtLeast32Characters
ADMIN_PASSWORD=Competition2024SecurePass!
DB_PATH=/opt/RedBoard/data/dashboard.db
```

---

## Setting Up Teams

Teams define the target networks to scan. Each team has an IP range that scanners will target.

### Via Web Interface

1. Login as admin
2. Go to **Teams** page
3. Click **+ Add Team**
4. Fill in:
   - **Name**: Team identifier (e.g., "Team 1", "Blue Team Alpha")
   - **IP Range**: CIDR notation (e.g., `10.1.1.0/24`)
   - **Color**: For dashboard display (optional)
5. Click **Create**

### Example Team Setup for 5-Team Competition

| Team Name | IP Range |
|-----------|----------|
| Team 1 | 10.1.1.0/24 |
| Team 2 | 10.1.2.0/24 |
| Team 3 | 10.1.3.0/24 |
| Team 4 | 10.1.4.0/24 |
| Team 5 | 10.1.5.0/24 |

---

## User Management

### User Roles

| Role | Capabilities |
|------|--------------|
| **Admin** | Full access - manage users, teams, jobs, view all data |
| **Scanner** | Can authenticate and submit scan results |
| **Viewer** | Read-only access to dashboard and scan results |

### Creating Users

1. Login as admin
2. Go to **Users** page
3. Click **+ Add User**
4. Fill in:
   - **Username**: Unique identifier
   - **Password**: Minimum 8 characters
   - **Roles**: Select appropriate roles
   - **Activate immediately**: Enable for immediate access
5. Click **Create**

### Creating a Scanner Account

For the NMAP Agent to connect, create a dedicated scanner user:

1. Go to **Users** → **+ Add User**
2. Username: `scanner`
3. Password: (secure password)
4. Roles: Enable **Scanner** only
5. Activate immediately: **Yes**
6. Click **Create**

Use these credentials in the agent's `.env` file.

### Resetting Passwords

1. Go to **Users** page
2. Find the user
3. Click **Password** button
4. Enter new password (minimum 8 characters)
5. Confirm and click **Change**

---

## Connecting Scanners

### Quick Setup

1. Create a scanner user (see above)
2. On the scanner machine, configure `.env`:
```bash
API_USER=scanner
API_PASS=your_scanner_password
API_URL_BASE=http://DASHBOARD_IP:8080
```
3. Run the scanner:
```bash
sudo ./nmap-agent-improved
```

### Verifying Connection

1. Check the **Jobs** page - you should see jobs being created
2. Watch for jobs transitioning from "running" to "complete"
3. Check the **Dashboard** for scan results appearing

---

## API Documentation

The dashboard includes built-in Swagger API documentation.

### Accessing Swagger UI

Navigate to:
```
http://YOUR_SERVER_IP:8080/swagger/index.html
```

### Key Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/auth/login` | Authenticate user |
| GET | `/auth/users` | List all users (admin) |
| POST | `/auth/admin/create-user` | Create new user (admin) |
| PUT | `/auth/admin/reset-password/:uid` | Reset password (admin) |
| GET | `/teams` | List all teams |
| POST | `/teams` | Create team (admin) |
| GET | `/jobs` | List all jobs |
| GET | `/jobs/nmap/next` | Get next scan job (scanner) |
| POST | `/jobs/nmap/:jid` | Upload scan results (scanner) |
| GET | `/dashboard/data` | Get dashboard summary |
| GET | `/hosts/by-team/:tid` | Get hosts for a team |

---

## Running in Production

### Using systemd (Recommended)

Create a service file:

```bash
sudo nano /etc/systemd/system/nmap-dashboard.service
```

```ini
[Unit]
Description=NMAP Dashboard
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=/opt/RedBoard
ExecStart=/opt/RedBoard/Redteam-Dashboard-go
Restart=always
RestartSec=5
Environment=GIN_MODE=release

[Install]
WantedBy=multi-user.target
```

Enable and start:

```bash
sudo systemctl daemon-reload
sudo systemctl enable nmap-dashboard
sudo systemctl start nmap-dashboard

# Check status
sudo systemctl status nmap-dashboard

# View logs
sudo journalctl -u nmap-dashboard -f
```

### Using Screen

```bash
screen -S dashboard
cd /opt/RedBoard
./Redteam-Dashboard-go
# Press Ctrl+A, then D to detach

# Reattach later
screen -r dashboard
```

### Firewall Configuration

```bash
# Allow dashboard port
sudo ufw allow 8080/tcp

# If using a different port
sudo ufw allow YOUR_PORT/tcp
```

---

## Troubleshooting

### Dashboard won't start

**Check for port conflicts:**
```bash
sudo lsof -i :8080
```

**Check logs:**
```bash
./Redteam-Dashboard-go 2>&1 | head -50
```

**Verify Go version:**
```bash
go version  # Should be 1.21+
```

### Can't login

**Default credentials:**
- Username: `admin`
- Password: Value of `ADMIN_PASSWORD` in `.env` (default: `changeme`)

**Reset admin password:**
Delete the database and restart:
```bash
rm ./data/dashboard.db
./Redteam-Dashboard-go
```

### CSS not loading / Old styles showing

**Hard refresh your browser:**
- Chrome/Firefox: `Ctrl + Shift + R`
- Safari: `Cmd + Shift + R`

**Clear browser cache completely**

### Scanner can't connect

**Check connectivity:**
```bash
curl http://DASHBOARD_IP:8080/health
```

**Verify scanner user:**
1. Login to dashboard as admin
2. Go to Users
3. Verify scanner user exists, has Scanner role, and is active

**Check scanner logs for specific error**

### No scan results appearing

1. Check **Jobs** page - are jobs completing?
2. If jobs show "complete" but no data:
   - Check scanner output for upload errors
   - Verify job has hosts_found > 0
3. If jobs stuck on "running":
   - Scanner may have crashed
   - Check scanner logs

### Database errors

**Reset database:**
```bash
cd /opt/RedBoard
rm ./dashboard.db
./Redteam-Dashboard-go
```

**Check database permissions:**
```bash
ls -la ./dashboard.db
# File should be writable
```

**Specify custom database path:**
```bash
# In .env file:
DB_PATH=/path/to/your/dashboard.db
```

---

## Directory Structure

```
/opt/RedBoard/
├── Redteam-Dashboard-go    # Main binary
├── .env                    # Configuration
├── dashboard.db            # SQLite database (default location)
├── static/
│   ├── common.css          # Styles
│   └── common.js           # JavaScript utilities
├── templates/
│   ├── login.html
│   ├── main.html
│   ├── teams.html
│   ├── jobs.html
│   └── users.html
├── controllers/            # API handlers
├── models/                 # Database models
├── middleware/             # Auth middleware
└── server/                 # Router setup
```

---

## Security Recommendations

1. **Change default passwords** - Update ADMIN_PASSWORD before first run
2. **Use strong session secret** - Generate with `openssl rand -base64 32`
3. **Run behind reverse proxy** - Use nginx/caddy with HTTPS in production
4. **Restrict network access** - Use firewall to limit who can reach the dashboard
5. **Create separate scanner accounts** - Don't use admin credentials for scanners
6. **Regular backups** - Back up `./data/dashboard.db` regularly

---

## License

BSD 2-Clause License
