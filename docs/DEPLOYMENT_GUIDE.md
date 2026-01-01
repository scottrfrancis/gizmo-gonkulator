# Complete Deployment Guide: Auth0 + AWS LightSail + Claude Desktop

This guide walks you through deploying the MCP Calculator Server on AWS LightSail with Auth0 authentication, then connecting it to Claude Desktop using `mcp-remote`.

**Time Required:** ~60-90 minutes
**Cost:** ~$5/month (LightSail) + Free (Auth0 free tier)

---

## Table of Contents

1. [Prerequisites](#1-prerequisites)
2. [Part 1: Auth0 Account Setup](#2-part-1-auth0-account-setup-20-minutes)
3. [Part 2: Auth0 API & Application Configuration](#3-part-2-auth0-api--application-configuration-15-minutes)
4. [Part 3: LightSail Deployment](#4-part-3-lightsail-deployment-25-minutes)
5. [Part 4: Test with mcp-remote + Bearer Token](#5-part-4-test-with-mcp-remote--bearer-token-15-minutes)
6. [Part 5: Full OAuth Configuration (Optional)](#6-part-5-full-oauth-configuration-optional)
7. [Troubleshooting](#7-troubleshooting)

---

## 1. Prerequisites

Before starting, ensure you have:

- [ ] AWS account (for LightSail)
- [ ] Email address for Auth0 account
- [ ] Claude Desktop installed
- [ ] Node.js 18+ installed (for mcp-remote)
- [ ] Terminal/SSH client
- [ ] A domain name (optional, recommended for production)

---

## 2. Part 1: Auth0 Account Setup (20 minutes)

This section provides step-by-step instructions for creating and configuring a new Auth0 account from scratch.

### Step 1.1: Navigate to Auth0 Signup

1. Open your browser and go to: **https://auth0.com**
2. Click the **"Sign Up"** button in the top-right corner
3. You'll be redirected to: `https://auth0.com/signup`

### Step 1.2: Create Your Account

You have three signup options:

**Option A: Email/Password (Recommended for demos)**
1. Enter your email address
2. Create a password (min 8 chars, 1 lowercase, 1 uppercase, 1 number)
3. Click **"Continue"**
4. Check your email for a verification link
5. Click the verification link to confirm your account

**Option B: Continue with GitHub**
1. Click **"Continue with GitHub"**
2. Authorize Auth0 to access your GitHub account
3. You'll be automatically logged in

**Option C: Continue with Google**
1. Click **"Continue with Google"**
2. Select your Google account
3. You'll be automatically logged in

### Step 1.3: Choose Your Account Type

After verifying your email (or OAuth login), you'll see:

```
What best describes you?
○ I need to add login/signup to my app (Developer)
○ I need to secure APIs and services (Developer)
○ I'm exploring Auth0 (Developer)
○ I need to manage identities for employees (IT Admin)
```

**Select:** `I need to secure APIs and services`

Click **"Continue"**

### Step 1.4: Configure Your Tenant

A "tenant" is your Auth0 environment. It becomes part of your domain.

1. **Tenant Domain:** Enter a unique name
   - Example: `mcp-calculator-demo`
   - This creates: `mcp-calculator-demo.us.auth0.com`
   - **Tip:** Use lowercase letters, numbers, and hyphens only
   - **Tip:** Keep it short but descriptive

2. **Region:** Select your preferred region
   - `US` - United States (us.auth0.com)
   - `EU` - Europe (eu.auth0.com)
   - `AU` - Australia (au.auth0.com)
   - **Choose the region closest to your LightSail deployment**

3. Click **"Create Account"**

### Step 1.5: Skip the Onboarding Wizard

Auth0 will show an onboarding wizard. You can skip it:

1. If you see "Let's get started" with a sample app, click **"Skip"** or **"Dashboard"**
2. You want to go directly to the Auth0 Dashboard

### Step 1.6: Understand the Free Tier

Auth0's free tier includes:

| Feature | Free Tier Limit |
|---------|-----------------|
| Monthly Active Users | 7,500 |
| Machine-to-Machine Tokens | 1,000/month |
| Social Connections | Unlimited |
| Custom Domains | Not included |
| Support | Community only |

**For a demo with <5 users, you'll stay well within free limits.**

### Step 1.7: Verify Your Dashboard Access

You should now see the Auth0 Dashboard with:

- Left sidebar: Applications, APIs, Actions, etc.
- Top: Your tenant name (e.g., `mcp-calculator-demo`)
- Main area: Getting Started guide or Activity

**Record your tenant domain:**
```
My Tenant Domain: _________________________.us.auth0.com
                  (e.g., mcp-calculator-demo)
```

---

## 3. Part 2: Auth0 API & Application Configuration (15 minutes)

Now we'll configure Auth0 to secure your MCP Calculator Server.

### Step 2.1: Create an API (Resource Server)

An "API" in Auth0 represents the resource you're protecting—in this case, your MCP Calculator Server.

1. In the left sidebar, click **"Applications"** → **"APIs"**
2. Click the **"+ Create API"** button (top right)
3. Fill in the form:

   | Field | Value | Notes |
   |-------|-------|-------|
   | **Name** | `MCP Calculator API` | Display name only |
   | **Identifier** | `https://mcp-calculator.example.com` | This is your "audience" |
   | **Signing Algorithm** | `RS256` | Leave as default |

   > **Important:** The Identifier (audience) is a logical name, not a real URL. It uniquely identifies your API. Use your real domain if you have one (e.g., `https://mcp.yourdomain.com`), or use a placeholder like `https://mcp-calculator.local`.

4. Click **"Create"**

### Step 2.2: Record Your API Identifier

After creation, you'll be on the API's settings page:

```
API Identifier (Audience): _________________________________
                           (e.g., https://mcp-calculator.example.com)
```

### Step 2.3: Add Custom Scopes (Permissions)

Scopes define what actions are allowed. While our server currently validates the JWT signature and audience (not specific scopes), setting these up is good practice.

1. Click the **"Permissions"** tab
2. Add these permissions by filling in the fields and clicking **"Add"** for each:

   | Permission (Scope) | Description |
   |--------------------|-------------|
   | `mcp:calculate` | Execute arithmetic calculations |
   | `mcp:read` | List available tools |

3. After adding both, you should see them listed in the table

### Step 2.4: Create a Machine-to-Machine Application

This is what will request tokens to access your API. Claude Desktop (via mcp-remote) will use these credentials.

1. In the left sidebar, click **"Applications"** → **"Applications"**
2. Click **"+ Create Application"** (top right)
3. Fill in:

   | Field | Value |
   |-------|-------|
   | **Name** | `Claude Desktop MCP Client` |
   | **Application Type** | **Machine to Machine Applications** |

4. Click **"Create"**

### Step 2.5: Authorize the Application for Your API

Immediately after creation, you'll see "Authorize Machine to Machine Application":

1. In the **"Select an API"** dropdown, choose: `MCP Calculator API`
2. Check the permissions to grant:
   - [x] `mcp:calculate`
   - [x] `mcp:read`
3. Click **"Authorize"**

If you missed this step:
1. Go to **"Applications"** → **"Applications"**
2. Click on `Claude Desktop MCP Client`
3. Go to the **"APIs"** tab
4. Find `MCP Calculator API` and click **"Authorize"**
5. Check the permissions and save

### Step 2.6: Get Your Credentials

1. Go to **"Applications"** → **"Applications"**
2. Click on **"Claude Desktop MCP Client"**
3. You're on the **"Settings"** tab
4. Record these values:

```
┌─────────────────────────────────────────────────────────────────┐
│ Auth0 Credentials (KEEP THESE SECRET!)                         │
├─────────────────────────────────────────────────────────────────┤
│ Domain:        _________________________.us.auth0.com           │
│                                                                 │
│ Client ID:     ______________________________________________   │
│                (32 characters, looks like: aBcD1234...)        │
│                                                                 │
│ Client Secret: ______________________________________________   │
│                (64 characters, very long string)               │
│                                                                 │
│ API Audience:  ______________________________________________   │
│                (from Step 2.2, e.g., https://mcp-calculator..) │
└─────────────────────────────────────────────────────────────────┘
```

> **⚠️ Security Warning:** Never commit the Client Secret to git or share it publicly!

### Step 2.7: Test Token Generation (Verify Setup)

Let's verify your Auth0 setup works by generating a test token:

```bash
# Replace with YOUR values
curl --request POST \
  --url "https://YOUR_TENANT.us.auth0.com/oauth/token" \
  --header "Content-Type: application/json" \
  --data '{
    "client_id": "YOUR_CLIENT_ID",
    "client_secret": "YOUR_CLIENT_SECRET",
    "audience": "YOUR_API_IDENTIFIER",
    "grant_type": "client_credentials"
  }'
```

**Example with real values:**
```bash
curl --request POST \
  --url "https://mcp-calculator-demo.us.auth0.com/oauth/token" \
  --header "Content-Type: application/json" \
  --data '{
    "client_id": "aBcDeFgH1234567890abcdefghijklmn",
    "client_secret": "xYz789...(long secret)...abc",
    "audience": "https://mcp-calculator.example.com",
    "grant_type": "client_credentials"
  }'
```

**Expected Response:**
```json
{
  "access_token": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCIsImtpZCI6Ij...(very long)...",
  "token_type": "Bearer",
  "expires_in": 86400
}
```

If you get an error:
- `unauthorized_client`: Check Client ID and Secret
- `invalid_audience`: Check the audience matches your API Identifier exactly
- Connection error: Check the tenant domain URL

**Save a test token for later:**
```bash
# Run this to save a token for testing
export MCP_TOKEN=$(curl -s --request POST \
  --url "https://YOUR_TENANT.us.auth0.com/oauth/token" \
  --header "Content-Type: application/json" \
  --data '{
    "client_id": "YOUR_CLIENT_ID",
    "client_secret": "YOUR_CLIENT_SECRET",
    "audience": "YOUR_API_IDENTIFIER",
    "grant_type": "client_credentials"
  }' | jq -r '.access_token')

echo "Token starts with: ${MCP_TOKEN:0:50}..."
```

---

## 4. Part 3: LightSail Deployment (25 minutes)

### Step 3.1: Log into AWS LightSail

1. Go to: **https://lightsail.aws.amazon.com/**
2. Log in with your AWS credentials
3. If this is your first time, you may see a welcome page—click **"Let's get started"**

### Step 3.2: Create a LightSail Instance

1. Click **"Create instance"**
2. Configure the instance:

   | Setting | Value | Notes |
   |---------|-------|-------|
   | **Instance location** | Choose closest to your users | e.g., `Virginia (us-east-1)` |
   | **Platform** | `Linux/Unix` | |
   | **Blueprint** | `OS Only` → **`Ubuntu 22.04 LTS`** | Scroll down in OS Only section |
   | **Instance plan** | `$5 USD/month` | 1 GB RAM, 1 vCPU, 40 GB SSD |
   | **Instance name** | `mcp-calculator` | |

3. Click **"Create instance"**
4. Wait 2-3 minutes for status to show **"Running"**

### Step 3.3: Configure Firewall Rules

1. Click on your instance name: **"mcp-calculator"**
2. Click the **"Networking"** tab
3. Under **"IPv4 Firewall"**, you'll see SSH (22) already open
4. Click **"+ Add rule"** and add:

   | Application | Protocol | Port Range |
   |-------------|----------|------------|
   | HTTPS | TCP | 443 |
   | Custom | TCP | 8080 |

5. Click **"Create"** for each rule

### Step 3.4: Create a Static IP

A static IP ensures your server address doesn't change on reboot.

1. Still in the **"Networking"** tab
2. Under **"Public IP"**, click **"Create static IP"**
3. Give it a name: `mcp-calculator-ip`
4. Click **"Create and attach"**
5. **Record your static IP:**

```
My LightSail Static IP: ___.___.___.___
```

### Step 3.5: Connect via SSH

**Option A: Browser-based SSH (Easiest)**
1. Go back to the instance page
2. Click the **"Connect using SSH"** button
3. A terminal opens in your browser

**Option B: Local SSH (Download key)**
1. Go to **"Account"** → **"SSH keys"** in LightSail
2. Download the default key for your region
3. Connect from terminal:
   ```bash
   chmod 400 ~/Downloads/LightsailDefaultKey-*.pem
   ssh -i ~/Downloads/LightsailDefaultKey-*.pem ubuntu@YOUR_STATIC_IP
   ```

### Step 3.6: Install Docker

Run these commands on your LightSail instance:

```bash
# Update system packages
sudo apt update && sudo apt upgrade -y

# Install Docker using the convenience script
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh

# Add ubuntu user to docker group (avoids needing sudo)
sudo usermod -aG docker ubuntu

# Install Docker Compose plugin
sudo apt install -y docker-compose-plugin

# Verify installation
docker --version
docker compose version

# IMPORTANT: Log out and back in for group changes to take effect
exit
```

**Reconnect via SSH** (browser or terminal)

Verify docker works without sudo:
```bash
docker ps
# Should show empty container list, not "permission denied"
```

### Step 3.7: Clone the Repository

```bash
# Clone the MCP Calculator repository
git clone https://github.com/scottrfrancis/mcp-calculator.git
cd mcp-calculator
```

### Step 3.8: Configure Environment

Create your environment file with Auth0 settings:

```bash
cat > .env << 'EOF'
# Server Configuration
MCP_HOST=0.0.0.0
MCP_PORT=8080
MCP_LOG_LEVEL=info

# OAuth Configuration - EDIT THESE VALUES
MCP_OAUTH_ENABLED=true
MCP_OAUTH_ISSUER=https://YOUR_TENANT.us.auth0.com/
MCP_OAUTH_AUDIENCE=https://mcp-calculator.example.com
EOF

# Edit with your actual values
nano .env
```

**In nano:**
1. Replace `YOUR_TENANT` with your Auth0 tenant name
2. Replace the audience with your API Identifier from Step 2.2
3. Press `Ctrl+O` to save, `Enter` to confirm, `Ctrl+X` to exit

**Verify your .env file:**
```bash
cat .env
```

Should look like:
```
MCP_HOST=0.0.0.0
MCP_PORT=8080
MCP_LOG_LEVEL=info
MCP_OAUTH_ENABLED=true
MCP_OAUTH_ISSUER=https://mcp-calculator-demo.us.auth0.com/
MCP_OAUTH_AUDIENCE=https://mcp-calculator.example.com
```

### Step 3.9: Build and Start the Server

```bash
# Build and start in detached mode
docker compose up -d --build

# Watch the logs (Ctrl+C to stop watching)
docker compose logs -f
```

You should see:
```
mcp-calculator-1  | 2025/01/XX 12:00:00 Starting MCP Calculator Server v1.0.0
mcp-calculator-1  | 2025/01/XX 12:00:00 OAuth enabled: issuer=https://mcp-calculator-demo.us.auth0.com/
mcp-calculator-1  | 2025/01/XX 12:00:00 Server listening on 0.0.0.0:8080
```

### Step 3.10: Test the Deployment

**Test 1: Health check (no auth required)**
```bash
curl http://localhost:8080/health
```
Expected: `{"status":"healthy","uptime":...,"version":"1.0.0"}`

**Test 2: Unauthenticated request (should fail)**
```bash
curl http://localhost:8080/mcp \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","id":1,"method":"initialize"}'
```
Expected: `401 Unauthorized` or `{"error":"missing or invalid token"}`

**Test 3: From your local machine**
```bash
curl http://YOUR_STATIC_IP:8080/health
```
Expected: Same healthy response

---

## 5. Part 4: Test with mcp-remote + Bearer Token (15 minutes)

Now we'll connect Claude Desktop to your secure LightSail server using `mcp-remote` with a Bearer token.

### Step 4.1: Generate an Access Token

On your **local machine**, generate a token:

```bash
# Set your Auth0 credentials
AUTH0_DOMAIN="YOUR_TENANT.us.auth0.com"
AUTH0_CLIENT_ID="YOUR_CLIENT_ID"
AUTH0_CLIENT_SECRET="YOUR_CLIENT_SECRET"
AUTH0_AUDIENCE="YOUR_API_IDENTIFIER"

# Get a token
MCP_TOKEN=$(curl -s --request POST \
  --url "https://${AUTH0_DOMAIN}/oauth/token" \
  --header "Content-Type: application/json" \
  --data "{
    \"client_id\": \"${AUTH0_CLIENT_ID}\",
    \"client_secret\": \"${AUTH0_CLIENT_SECRET}\",
    \"audience\": \"${AUTH0_AUDIENCE}\",
    \"grant_type\": \"client_credentials\"
  }" | jq -r '.access_token')

echo "Your token (first 50 chars): ${MCP_TOKEN:0:50}..."
echo ""
echo "Full token (copy this for Claude Desktop config):"
echo "$MCP_TOKEN"
```

**Copy the full token—you'll need it in Step 4.3**

### Step 4.2: Verify Token Works Against Server

Test the token against your LightSail server:

```bash
# Replace YOUR_STATIC_IP with your LightSail IP
curl -X POST "http://YOUR_STATIC_IP:8080/mcp" \
  -H "Authorization: Bearer $MCP_TOKEN" \
  -H "Content-Type: application/json" \
  -H "Accept: application/json, text/event-stream" \
  -d '{
    "jsonrpc": "2.0",
    "id": 1,
    "method": "initialize",
    "params": {
      "protocolVersion": "2025-03-26",
      "clientInfo": {"name": "test", "version": "1.0"}
    }
  }'
```

**Expected:** JSON response with server info and an `Mcp-Session-Id` header.

If this works, your OAuth setup is correct!

### Step 4.3: Install mcp-remote

Ensure Node.js 18+ is installed:
```bash
node --version  # Should be v18.x.x or higher
```

Install mcp-remote globally (optional, but useful for testing):
```bash
npm install -g mcp-remote
```

### Step 4.4: Configure Claude Desktop

Locate your Claude Desktop configuration file:

| OS | Path |
|----|------|
| **macOS** | `~/Library/Application Support/Claude/claude_desktop_config.json` |
| **Windows** | `%APPDATA%\Claude\claude_desktop_config.json` |
| **Linux** | `~/.config/Claude/claude_desktop_config.json` |

**Edit the config file** (create it if it doesn't exist):

```json
{
  "mcpServers": {
    "mcp-calculator": {
      "command": "npx",
      "args": [
        "mcp-remote",
        "http://YOUR_STATIC_IP:8080/mcp",
        "--header",
        "Authorization:Bearer YOUR_TOKEN_HERE"
      ]
    }
  }
}
```

**Important notes:**
- Replace `YOUR_STATIC_IP` with your LightSail static IP
- Replace `YOUR_TOKEN_HERE` with the full token from Step 4.1
- **No space after the colon** in `Authorization:Bearer` (workaround for a Claude Desktop bug)
- The token is very long (1000+ characters)—that's normal

**Alternative: Use environment variable for the token:**

```json
{
  "mcpServers": {
    "mcp-calculator": {
      "command": "npx",
      "args": [
        "mcp-remote",
        "http://YOUR_STATIC_IP:8080/mcp",
        "--header",
        "Authorization:Bearer ${MCP_AUTH_TOKEN}"
      ],
      "env": {
        "MCP_AUTH_TOKEN": "eyJhbGciOiJSUzI1NiIs...(your full token)..."
      }
    }
  }
}
```

### Step 4.5: Restart Claude Desktop

1. **Quit Claude Desktop completely**
   - macOS: Cmd+Q or right-click dock icon → Quit
   - Windows: Right-click system tray icon → Exit
   - Linux: Close the window or use system tray

2. **Reopen Claude Desktop**

3. **Look for the hammer icon** 🔨 in the bottom-right of the input box
   - If you see it, the MCP server connected successfully!
   - Click it to see available tools

### Step 4.6: Test the Integration

In Claude Desktop, try these prompts:

**Test 1: Basic Calculation**
> "Use the calculator to add 0.1 and 0.2"

Expected: Claude uses the MCP tool and returns exactly `0.3`

**Test 2: Financial Calculation**
> "Calculate compound interest: $10,000 principal, 7% annual rate, 10 years"

Expected: Claude calls the calculator and returns ~$19,671.51

**Test 3: Multiple Operations**
> "Calculate the sum of 100, 200, and 300, then find what percentage 200 is of that total"

Expected: Uses calculator for precise results

---

## 6. Part 5: Full OAuth Configuration (Optional)

Once mcp-remote with Bearer token works, you can optionally set up full OAuth flow. This requires:

1. **HTTPS** - OAuth requires secure connections
2. **Domain name** - Recommended for proper OAuth
3. **Caddy or nginx** - For TLS termination

### Step 5.1: Set Up HTTPS with Caddy

On your LightSail instance:

```bash
cd ~/mcp-calculator

# Create Caddyfile for your domain
cat > Caddyfile << 'EOF'
mcp.yourdomain.com {
    reverse_proxy mcp-calculator:8080
}
EOF

# Or for IP-only with self-signed cert (testing only):
cat > Caddyfile << 'EOF'
:443 {
    tls internal
    reverse_proxy mcp-calculator:8080
}
:80 {
    redir https://{host}{uri} permanent
}
EOF
```

Create production docker-compose:

```bash
cat > docker-compose.prod.yaml << 'EOF'
version: "3.9"

services:
  mcp-calculator:
    build: .
    environment:
      - MCP_HOST=0.0.0.0
      - MCP_PORT=8080
      - MCP_OAUTH_ENABLED=${MCP_OAUTH_ENABLED:-false}
      - MCP_OAUTH_ISSUER=${MCP_OAUTH_ISSUER}
      - MCP_OAUTH_AUDIENCE=${MCP_OAUTH_AUDIENCE}
    restart: unless-stopped
    networks:
      - mcp-network

  caddy:
    image: caddy:2-alpine
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./Caddyfile:/etc/caddy/Caddyfile
      - caddy_data:/data
      - caddy_config:/config
    depends_on:
      - mcp-calculator
    restart: unless-stopped
    networks:
      - mcp-network

networks:
  mcp-network:

volumes:
  caddy_data:
  caddy_config:
EOF
```

Start with production config:
```bash
docker compose -f docker-compose.prod.yaml up -d
```

### Step 5.2: Update Claude Desktop for HTTPS

Update your Claude Desktop config to use HTTPS:

```json
{
  "mcpServers": {
    "mcp-calculator": {
      "command": "npx",
      "args": [
        "mcp-remote",
        "https://mcp.yourdomain.com/mcp",
        "--header",
        "Authorization:Bearer YOUR_TOKEN_HERE"
      ]
    }
  }
}
```

---

## 7. Troubleshooting

### Auth0 Issues

**"unauthorized_client" error when getting token**
- Verify Client ID and Client Secret are correct
- Check the application is authorized for the API
- Go to Applications → Your App → APIs tab and verify authorization

**"invalid_audience" error**
- The audience in your token request must exactly match the API Identifier
- Check for trailing slashes, https vs http, etc.

**Token works in curl but not in Claude Desktop**
- Tokens expire after 24 hours (86400 seconds)
- Generate a fresh token
- Check for copy/paste errors (token is very long)

### LightSail Issues

**Connection refused**
- Check firewall rules include port 8080
- Verify docker container is running: `docker compose ps`
- Check logs: `docker compose logs`

**502 Bad Gateway (with Caddy)**
- The mcp-calculator container might not be running
- Check: `docker compose -f docker-compose.prod.yaml logs`

### Claude Desktop Issues

**Hammer icon doesn't appear**
- Check Claude Desktop logs:
  - macOS: `~/Library/Logs/Claude/`
  - Windows: `%APPDATA%\Claude\logs\`
- Verify Node.js 18+ is installed
- Verify JSON config is valid (no trailing commas)

**"spawn npx ENOENT" error**
- Node.js not in PATH
- Try using full path: `/usr/local/bin/npx`

**Tool calls fail silently**
- Check the server logs: `docker compose logs -f`
- Verify token hasn't expired

### Useful Commands

```bash
# Check server status
docker compose ps
docker compose logs -f

# Restart server
docker compose restart

# Rebuild and restart
docker compose down && docker compose up -d --build

# Test health endpoint
curl http://localhost:8080/health

# Get new token (on local machine)
curl -s -X POST "https://YOUR_TENANT.us.auth0.com/oauth/token" \
  -H "Content-Type: application/json" \
  -d '{"client_id":"...","client_secret":"...","audience":"...","grant_type":"client_credentials"}' \
  | jq -r '.access_token'
```

---

## Quick Reference Card

```
┌─────────────────────────────────────────────────────────────────┐
│                    MY DEPLOYMENT DETAILS                        │
├─────────────────────────────────────────────────────────────────┤
│ Auth0 Tenant:     _________________________.us.auth0.com        │
│ Auth0 Client ID:  __________________________________________    │
│ Auth0 Secret:     __________________________________________    │
│ API Audience:     __________________________________________    │
│                                                                 │
│ LightSail IP:     ___.___.___.___                               │
│ Domain (if any):  __________________________________________    │
│                                                                 │
│ Server URL:       http://[IP]:8080/mcp                          │
│ Health Check:     http://[IP]:8080/health                       │
│                                                                 │
│ Token URL:        https://[tenant].us.auth0.com/oauth/token     │
│ Token Lifetime:   24 hours (86400 seconds)                      │
└─────────────────────────────────────────────────────────────────┘
```

---

## Next Steps

After successful deployment:

1. **Monitor Auth0** - Dashboard → Monitoring → Logs
2. **Set up token refresh** - Tokens expire after 24 hours
3. **Add custom domain** - For production use
4. **Enable rate limiting** - Protect against abuse
5. **Set up alerts** - CloudWatch for LightSail

For security hardening, see [SECURITY.md](./SECURITY.md).

---

## Sources

- [mcp-remote - npm](https://www.npmjs.com/package/mcp-remote)
- [mcp-remote GitHub](https://github.com/geelen/mcp-remote)
- [Auth0 Machine-to-Machine Documentation](https://auth0.com/docs/get-started/authentication-and-authorization-flow/client-credentials-flow)
