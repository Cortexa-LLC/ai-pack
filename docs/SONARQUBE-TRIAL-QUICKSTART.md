# SonarQube Developer Edition Setup

## Status: ✅ Trial Requested

You've signed up for the Developer Edition trial. You should receive a license key via email within a few minutes.

## Quick Setup (3 steps)

### Step 1: Wait for License Key Email

Check your email for:
- **From:** SonarSource
- **Subject:** "Your SonarQube Developer Edition trial"
- **Contains:** License key (long string)

### Step 2: Upgrade to Developer Edition

Once you have the license key, run:

```bash
./upgrade-to-sonarqube-developer.sh
```

This script will:
- ✅ Stop Community Edition
- ✅ Start Developer Edition (same Docker volumes - no data loss)
- ✅ Preserve all projects, settings, and history

**All your data is safe** - Docker volumes are reused.

### Step 3: Apply License Key

1. **Open SonarQube:** http://localhost:9000
2. **Login:** admin/admin
3. **Navigate to:** Administration → Configuration → License Manager
4. **Paste license key** from email
5. **Click "Save"**
6. **Restart if prompted:**
   ```bash
   docker-compose \
     -f .ai-pack/docker-compose.sonarqube.yml \
     -f docker-compose.sonarqube-dev.yml \
     restart sonarqube
   ```

### Step 4: Verify C++ Plugin

1. Go to: **Administration → Marketplace**
2. Click **"Installed"** tab
3. Look for: **"C/C++ (CFamily)"**

If not installed:
- Search for "C/C++" in "All" tab
- Click "Install"
- Restart SonarQube

### Step 5: Run C++ Analysis

```bash
# Analyze your C++ code
source sonar.env
sonar-scanner

# Check results
open http://localhost:9000/dashboard?id=xasm-plus-plus
```

You should now see **C++ code analysis results**! 🎉

## What Changed

**Before (Community Edition):**
- ✅ Python analysis
- ❌ C++ analysis (not supported)
- Result: 30 files indexed, 0 analyzed

**After (Developer Edition):**
- ✅ Python analysis
- ✅ **C++ analysis** (now supported!)
- Result: 30 files indexed, ~35 C++ files analyzed

## Files Created

- **`docker-compose.sonarqube-dev.yml`** - Developer Edition override
- **`upgrade-to-sonarqube-developer.sh`** - Automated upgrade script
- **`DEVELOPER_EDITION_SETUP.md`** - This file

## Manual Commands (if not using script)

**Stop Community Edition:**
```bash
cd .ai-pack
docker-compose -f docker-compose.sonarqube.yml down
cd ..
```

**Start Developer Edition:**
```bash
docker-compose \
  -f .ai-pack/docker-compose.sonarqube.yml \
  -f docker-compose.sonarqube-dev.yml \
  up -d
```

**View logs:**
```bash
docker-compose \
  -f .ai-pack/docker-compose.sonarqube.yml \
  -f docker-compose.sonarqube-dev.yml \
  logs -f sonarqube
```

**Stop Developer Edition:**
```bash
docker-compose \
  -f .ai-pack/docker-compose.sonarqube.yml \
  -f docker-compose.sonarqube-dev.yml \
  down
```

## Troubleshooting

### Container won't start
```bash
# Check logs
docker logs sonarqube

# Common issue: License not applied yet
# Solution: Apply license via web UI
```

### C++ plugin missing
```bash
# Install via Marketplace
# Administration → Marketplace → All → Search "C/C++" → Install → Restart
```

### Analysis still shows 0 files
```bash
# Verify license is active
# Administration → Configuration → License Manager

# Verify C++ plugin installed
# Administration → Marketplace → Installed

# Check compile_commands.json exists
ls -lh build/compile_commands.json

# Re-run analysis
source sonar.env
sonar-scanner
```

## Trial Details

- **Duration:** 14 days
- **Features:** Full Developer Edition (C++, PR decoration, branch analysis, etc.)
- **After trial:** Either purchase ($150/year) or revert to Community Edition

## Next Steps After Trial

**Option A: Purchase Developer Edition**
- Provides continued C++ analysis
- Cost: $150/year for up to 100K LOC

**Option B: Use Free Alternatives**
- Revert to Community Edition: `cd .ai-pack && docker-compose -f docker-compose.sonarqube.yml up -d`
- Use clang-tidy for C++ static analysis (free)
- Use compiler warnings (already have strict flags)

## Support

- **Documentation:** https://docs.sonarsource.com/sonarqube/latest/
- **C++ Analysis Guide:** https://docs.sonarsource.com/sonarqube/latest/analyzing-source-code/languages/c-family/
- **Community:** https://community.sonarsource.com/

---

**Status:** Waiting for license key email
**Next:** Run `./upgrade-to-sonarqube-developer.sh` when you receive the key
