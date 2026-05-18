# QuantumClaw One-Click Deployment for BaoTa Panel

> Deploy QuantumClaw AI API Gateway on servers managed by BaoTa panel (Ubuntu/Debian/CentOS).

## Quick Start

### Prerequisites

- A Linux server (Ubuntu 20.04+, Debian 11+, CentOS 7+) with BaoTa panel installed
- Root or sudo access
- Ports 80/443 open (if using a domain)

### One Command Deploy

```bash
bash <(curl -fsSL https://raw.githubusercontent.com/quantumclaw/quantumclaw/main/deploy/install.sh)
```

### Custom Deployment

```bash
# Deploy to a custom directory
bash install.sh --dir /data/quantumclaw

# Deploy with a domain for reverse proxy
bash install.sh --domain your-api.example.com

# Deploy with MySQL (default is SQLite)
bash install.sh --db-type mysql --domain your-api.example.com
```

## Script Options

| Option | Default | Description |
|--------|---------|-------------|
| `--dir` | `/opt/quantumclaw` | Installation directory |
| `--domain` | (none) | Domain for Nginx reverse proxy |
| `--db-type` | `sqlite` | Database type: `sqlite` or `mysql` |

## First Login

1. Open your browser and visit `http://<server-ip>:3666` (or your domain)
2. Log in with:
   - **Username**: `root`
   - **Password**: `123456`
3. **Immediately change the default password** in user settings!

## Managing via BaoTa Panel

1. Open your BaoTa panel
2. Navigate to **Docker** section
3. You'll see the `quantumclaw` container and related services
4. Use BaoTa's built-in management tools:
   - View real-time logs
   - Restart / Stop / Start services
   - Monitor resource usage
   - Set up automatic backups

## Directory Structure

```
/opt/quantumclaw/
├── .env                    # Environment variables (edit to configure)
├── docker-compose.yml      # Docker Compose configuration
├── data/                   # Persisted data (SQLite, Redis)
└── code/                   # Cloned source code (for updates)
```

## Configuration

Edit `/opt/quantumclaw/.env` to configure:

### Required Changes
```bash
SESSION_SECRET=<generate-a-random-string>
INITIAL_ROOT_TOKEN=<change-this-token>
```

### Optional: Enable OAuth Providers
```bash
GITHUB_OAUTH_ENABLED=true
GITHUB_CLIENT_ID=your-client-id
GITHUB_CLIENT_SECRET=your-client-secret
```

### Apply Changes
After editing `.env`, restart the services:
```bash
cd /opt/quantumclaw && docker compose restart
```

## Updating

```bash
cd /opt/quantumclaw
docker compose pull
docker compose up -d
```

## Troubleshooting

### Check logs
```bash
cd /opt/quantumclaw
docker compose logs -f
```

### Service not starting
```bash
docker compose logs quantumclaw
netstat -tlnp | grep 3666
```

### Nginx not working
```bash
nginx -t
systemctl status nginx
```

## Security Notes

1. **Always change the default root password** immediately after first login
2. **Set `INITIAL_ROOT_TOKEN` and `SESSION_SECRET`** to secure random values
3. **Use HTTPS** (BaoTa's SSL feature or Let's Encrypt) for production
4. **Regularly update** the Docker images
5. **Restrict port 3666** to localhost only if using Nginx (already configured by the script)

## License

MIT
