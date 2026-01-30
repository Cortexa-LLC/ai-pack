# Upgrading to SonarQube Developer Edition

## Overview

SonarQube Developer Edition includes C/C++ analysis and additional features not available in Community Edition.

## Pricing

**Developer Edition:**
- **Price:** $150/year per 100K lines of code
- **Minimum:** $150/year (covers up to 100K LOC)
- **Your project:** xasm++ has ~1K-2K lines of C++ code = $150/year

**What's Included:**
- ✅ C/C++ language analysis (the main reason to upgrade)
- ✅ Pull request decoration (GitHub/GitLab/Bitbucket integration)
- ✅ Branch analysis (compare quality across branches)
- ✅ Additional security rules
- ✅ OWASP/CWE security standards
- ✅ 12 months of support
- ✅ All Community Edition features

**Free Trial:**
- 14-day free trial available
- Full features during trial
- No credit card required

## Purchase Process

### Step 1: Request Trial or Purchase

**Option A: Start Free Trial (Recommended)**

1. Visit: https://www.sonarsource.com/products/sonarqube/downloads/
2. Click "Try Developer Edition"
3. Fill out form with:
   - Name
   - Email
   - Company
   - Use case: "C++ static analysis for open source project"
4. Receive trial license key via email

**Option B: Purchase Directly**

1. Visit: https://www.sonarsource.com/plans-and-pricing/sonarqube/
2. Click "Buy Now" under Developer Edition
3. Select:
   - Lines of code: 100K (minimum)
   - Billing: Annual
4. Complete checkout
5. Receive license key via email

### Step 2: Install Developer Edition

**If using Docker (recommended for local development):**

```bash
# Stop Community Edition container
docker stop sonarqube
docker rm sonarqube

# Pull Developer Edition image
docker pull sonarqube:10-developer

# Start Developer Edition
docker run -d --name sonarqube \
  -p 9000:9000 \
  -v sonarqube_data:/opt/sonarqube/data \
  -v sonarqube_logs:/opt/sonarqube/logs \
  -v sonarqube_extensions:/opt/sonarqube/extensions \
  sonarqube:10-developer
```

**If using ai-pack setup-sonarqube.py:**

The ai-pack script may need modification to use Developer Edition. You can:
1. Use Docker commands above instead
2. Or modify `.ai-pack/docker-compose.sonarqube.yml` to use `sonarqube:10-developer` image

### Step 3: Apply License Key

1. **Access SonarQube:** http://localhost:9000
2. **Login as admin** (admin/admin)
3. **Navigate to:** Administration → Configuration → License Manager
4. **Enter license key** received via email
5. **Click "Save"**
6. **Restart SonarQube** if prompted

### Step 4: Verify C++ Plugin

1. Go to: **Administration → Marketplace**
2. Check "Installed" tab
3. Verify **"C/C++ (CFamily)"** plugin is listed
4. If not installed, search for "C/C++" in "All" tab and click "Install"
5. Restart SonarQube

### Step 5: Re-run Analysis

```bash
# Your C++ code will now be analyzed
source sonar.env
sonar-scanner

# Check results
open http://localhost:9000/dashboard?id=xasm-plus-plus
```

## Docker Compose Setup (Recommended)

**Create docker-compose.yml:**

```yaml
version: "3.8"

services:
  sonarqube:
    image: sonarqube:10-developer
    container_name: sonarqube-dev
    ports:
      - "9000:9000"
    environment:
      - SONAR_JDBC_URL=jdbc:postgresql://db:5432/sonar
      - SONAR_JDBC_USERNAME=sonar
      - SONAR_JDBC_PASSWORD=sonar
    volumes:
      - sonarqube_data:/opt/sonarqube/data
      - sonarqube_logs:/opt/sonarqube/logs
      - sonarqube_extensions:/opt/sonarqube/extensions
    depends_on:
      - db

  db:
    image: postgres:15-alpine
    container_name: sonarqube-db
    environment:
      - POSTGRES_USER=sonar
      - POSTGRES_PASSWORD=sonar
      - POSTGRES_DB=sonar
    volumes:
      - postgresql_data:/var/lib/postgresql/data

volumes:
  sonarqube_data:
  sonarqube_logs:
  sonarqube_extensions:
  postgresql_data:
```

**Usage:**

```bash
# Start
docker-compose up -d

# Stop
docker-compose down

# View logs
docker-compose logs -f sonarqube
```

## Cost-Benefit Analysis for xasm++

### Current Quality Measures (Free):
- ✅ Compiler warnings with `-Wall -Wextra -Wpedantic -Werror`
- ✅ 500 tests with 85%+ coverage
- ✅ Code review (grade A-)
- ✅ Doxygen documentation (100% API coverage)

### With Developer Edition ($150/year):
- ✅ Everything above, plus:
- ✅ Automated C++ static analysis (code smells, bugs, vulnerabilities)
- ✅ Security hotspot detection (OWASP Top 10, CWE)
- ✅ Cognitive complexity metrics
- ✅ Duplication detection
- ✅ Technical debt calculation
- ✅ Quality gate enforcement
- ✅ Historical trend tracking

### Alternative (Free):
- ✅ Everything in "Current Quality Measures"
- ✅ clang-tidy for static analysis (free, similar rules)
- ✅ cppcheck for additional checks (free)
- ✅ SonarLint in IDE (free, real-time feedback)

### Recommendation:

**For xasm++ (open source project):**

1. **Start with free trial** (14 days) to evaluate
2. **Compare with clang-tidy** results
3. **Decision factors:**
   - If working solo: clang-tidy is likely sufficient
   - If working with team: Developer Edition provides centralized dashboard
   - If need CI/CD integration: Developer Edition has better PR decoration
   - If budget-conscious: clang-tidy + compiler warnings is already excellent

## Free Trial Activation Steps

**Quick Start (5 minutes):**

```bash
# 1. Request trial at: https://www.sonarsource.com/products/sonarqube/downloads/
# Form fields:
#   Name: [Your Name]
#   Email: [Your Email]
#   Company: Cortexa LLC
#   Use case: "C++ static analysis for xasm++ cross-assembler"

# 2. Check email for license key

# 3. Stop Community Edition
docker stop sonarqube
docker rm sonarqube

# 4. Start Developer Edition
docker run -d --name sonarqube \
  -p 9000:9000 \
  -v sonarqube_data:/opt/sonarqube/data \
  -v sonarqube_logs:/opt/sonarqube/logs \
  -v sonarqube_extensions:/opt/sonarqube/extensions \
  sonarqube:10-developer

# 5. Apply license at http://localhost:9000
#    Administration → Configuration → License Manager

# 6. Analyze your code
source sonar.env
sonar-scanner
```

## Support and Documentation

- **Documentation:** https://docs.sonarsource.com/sonarqube/latest/
- **C++ Analysis:** https://docs.sonarsource.com/sonarqube/latest/analyzing-source-code/languages/c-family/
- **Support:** support@sonarsource.com (with paid license)
- **Community:** https://community.sonarsource.com/

## Important Notes

1. **License is per organization** - One license covers all projects
2. **Lines of code counted:** Active maintained code (not comments, blanks)
3. **Trial limitations:** 14 days, full features
4. **Renewal:** Annual subscription, auto-renews unless cancelled
5. **Open source discount:** May be available, contact sales

## Next Steps

**To proceed with trial:**

```bash
# 1. Request trial
open https://www.sonarsource.com/products/sonarqube/downloads/

# 2. Wait for email (usually within minutes)

# 3. I can help with installation once you have the license key
```

**To purchase directly:**

```bash
open https://www.sonarsource.com/plans-and-pricing/sonarqube/
```

Let me know when you have the license key and I'll help with installation!
