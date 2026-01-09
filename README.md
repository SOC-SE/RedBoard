# NMAP Dashboard

A modern web dashboard for tracking open ports across multiple networks during cybersecurity competitions (CCDC, etc.).

![Dashboard Preview](docs/dashboard-preview.png)

## Features

- **Modern Dark Theme UI** - SOC-style dashboard optimized for monitoring
- **Real-time Updates** - Auto-refresh without page reloads
- **Team Management** - Full CRUD with IP range validation
- **Port Monitoring** - Visual indicators for dangerous ports
- **User Management** - Role-based access control (admin, scanner, viewer)
- **Job Queue** - Round-robin scan scheduling across teams
- **API Documentation** - Swagger UI for all endpoints

## Quick Start

### Prerequisites

- Go 1.21 or later
- Git

### Installation

```bash
# Clone the repository
git clone https://github.com/MNCCDC-RedTeam/nmap-dashboard-rt.git
cd nmap-dashboard-rt

# Copy and configure environment
cp env_example .env
# Edit .env with your settings

# Build and run
go build
./nmap-dashboard-rt
```

### First Run

On first startup, the application will:
1. Create a SQLite database (`dashboard.db`)
2. Generate an `admin` user with a random password
3. Display the admin credentials in the console

**Save the admin password!** You'll need it to log in and manage users.

### Configuration

Edit `.env` with your settings:

```env
# Required: Your server URL
API_BASE_URL=https://your-domain.com

# Optional: Server port (default: 8080)
PORT=8080

# Optional: Session secret for persistent sessions
# Generate with: openssl rand -base64 32
SESSION_SECRET=your-secret-here

# Optional: Set to 'release' for production
GIN_MODE=release
```

## Usage

### Creating Teams

1. Log in as admin
2. Navigate to **Teams** in the navigation
3. Click **Add Team**
4. Enter team name and IP range
5. IP range supports:
   - CIDR notation: `192.168.1.0/24`
   - Range notation: `192.168.1.1-254`
   - Single IPs: `192.168.1.1`
   - Comma-separated: `192.168.1.1, 192.168.1.2`

### Setting Up Scanners

1. Register a new user account
2. As admin, go to **Users** and enable:
   - **Active** - Allow login
   - **Scanner** - Allow scan uploads

Scanners use the API to:
1. Request jobs: `GET /jobs/nmap/next`
2. Upload results: `POST /jobs/nmap/{job_id}`

### API Documentation

Full API documentation is available at `/swagger/index.html` when the server is running.

## Architecture

```
├── controllers/     # HTTP request handlers
├── middleware/      # Authentication & authorization
├── models/          # Database models
├── server/          # Router & server setup
├── static/          # CSS & JavaScript
├── templates/       # HTML templates
└── docs/            # Swagger documentation
```

### Key Improvements Over Original

| Feature | Original | Improved |
|---------|----------|----------|
| Team Creation | JSON via Swagger only | Full UI with validation |
| Delete Team | Broken (wrong field) | Fixed and working |
| UI Design | Basic Bootstrap | Modern dark SOC theme |
| Page Updates | Full reload every 30s | Smart auto-refresh |
| Query Performance | N+1 queries | Eager loading |
| Session Security | Hardcoded secret | Configurable/random |
| Scan Data | Delete all on update | Merge existing data |

## Security Notes

- Change the default admin password immediately
- Set `SESSION_SECRET` in production for persistent sessions
- Use HTTPS in production (configure with reverse proxy)
- Set `GIN_MODE=release` in production

## Companion Scanner

This dashboard works with an nmap scanner agent that:
1. Requests jobs from the dashboard
2. Runs nmap scans
3. Uploads results

See the companion repository for the scanner implementation.

## License

BSD 2-Clause License - See [LICENSE](LICENSE) for details.

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Submit a pull request

## Acknowledgments

- Original project by [brian-l-johnson](https://github.com/brian-l-johnson)
- Built for MNCCDC Red Team operations
