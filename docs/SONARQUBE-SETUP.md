# SonarQube Local Setup

This document explains how to configure SonarQube credentials for local development without committing sensitive tokens to git.

## Quick Start

### Option 1: Environment Variables (Recommended)

1. **Create local environment file:**
   ```bash
   cp sonar.env.example sonar.env
   # Edit sonar.env and add your token
   ```

2. **Source the file before running analysis:**
   ```bash
   source sonar.env
   sonar-scanner
   ```

3. **Or add to your shell profile** (~/.zshrc or ~/.bashrc):
   ```bash
   export SONAR_TOKEN=your_token_here
   export SONAR_HOST_URL=http://localhost:9000
   ```

### Option 2: Local Properties File

1. **Create local properties file:**
   ```bash
   cp sonar-project.local.properties.example sonar-project.local.properties
   # Edit and add your token
   ```

2. **Pass to sonar-scanner:**
   ```bash
   sonar-scanner -Dproject.settings=sonar-project.local.properties
   ```

## Getting Your Token

1. Open http://localhost:9000 in your browser
2. Login (default: admin/admin for first-time setup)
3. Go to: **Account → Security → Generate Token**
4. Copy the generated token
5. Add to `sonar.env` or `sonar-project.local.properties`

## Files Excluded from Git

The following files are in `.gitignore` and safe for storing tokens:

- `sonar.env` - Environment variables
- `sonar-project.local.properties` - Local SonarQube configuration
- `.env` - Generic environment file

**Never commit tokens to git!**

## Using AI-Pack Scripts

The ai-pack framework provides convenience scripts:

```bash
# With environment variable
export SONAR_TOKEN=your_token_here
.ai-pack/scripts/validate-with-sonarqube.py src --project xasm-plus-plus --language cpp

# Or pass token directly (not recommended for repeated use)
.ai-pack/scripts/validate-with-sonarqube.py src --project xasm-plus-plus --language cpp --token YOUR_TOKEN
```

## Verification

Test your setup:

```bash
# Source environment
source sonar.env

# Run analysis
sonar-scanner

# Should see: "EXECUTION SUCCESS"
# Results at: http://localhost:9000/dashboard?id=xasm-plus-plus
```

## Troubleshooting

### Error: "HTTP 401 Unauthorized"
- Token not set or expired
- Regenerate token in SonarQube web UI
- Update `sonar.env` or `sonar-project.local.properties`

### Error: "No API token provided"
- Using ai-pack script without token
- Pass `--token` argument or set `SONAR_TOKEN` environment variable

### Server Not Running
```bash
# Check SonarQube status
curl http://localhost:9000/api/system/status

# Start SonarQube (if using Docker)
cd .ai-pack
python3 scripts/setup-sonarqube.py
```
