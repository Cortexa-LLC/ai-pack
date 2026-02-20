# Documentation Structure and Naming Conventions

This document defines the standard structure and naming conventions for the `docs/` directory.

---

## Directory Structure

```
docs/
├── market/                               # Market research and business strategy
├── product/                              # Product requirements and specifications
├── design/                               # UX/UI design and wireframes
├── architecture/                         # Technical architecture and design
├── adr/                                  # Architecture Decision Records
├── security/                             # Security vulnerabilities and audits
├── incidents/                            # Production incident reports
├── investigations/                       # Bug retrospectives and RCA
└── archaeology/                          # Historical system investigations
```

---

## Naming Conventions

### Date-Prefixed Directories

**Format:** `YYYY-MM-DD-descriptive-name`

**Used For:**
- Market requirements (`docs/market/`)
- Product requirements (`docs/product/`)
- Design specifications (`docs/design/`)
- Architecture documents (`docs/architecture/`)

**Benefits:**
- ✅ Chronologically sorted by default (`ls` shows oldest to newest)
- ✅ Date provides immediate context
- ✅ Complements task packet naming (`<beads-id>-<YYYYMMDDHHMMSS>-<short-desc>`)
- ✅ No need to track "next number"

**Examples:**
```
docs/market/2024-01-10-mobile-app/
docs/product/2024-01-15-user-authentication/
docs/architecture/2024-01-20-user-authentication/
```

### Sequential ID Files

**Format:** `PREFIX-YYYY-NNN-descriptive-name.md`

**Used For:**
- Architecture decisions (`ADR` prefix)
- Security vulnerabilities (`SEC` prefix)
- Production incidents (`INC` prefix)
- Bug investigations (`BUG` prefix)

**Benefits:**
- ✅ Trackable IDs for cross-referencing
- ✅ Sequential numbering shows volume/patterns
- ✅ Year provides temporal context
- ✅ Unique IDs prevent confusion

**Examples:**
```
docs/adr/001-use-graphql-federation.md
docs/security/SEC-2024-001-sql-injection.md
docs/incidents/INC-2024-001-database-outage.md
docs/investigations/BUG-2024-001-null-pointer.md
```

---

## docs/market/ - Market Requirements

**Purpose:** Market research, competitive analysis, and business cases

**Directory Naming:** `YYYY-MM-DD-product-name/`
- Date: When MRD was created
- Name: Product or market opportunity name

**Contents:**
- `mrd.md` - Market Requirements Document
- `competitive-analysis.md` - Competitive landscape
- `business-case.md` - ROI and business justification
- `market-research.md` - User research and market data

**Example:**
```
docs/market/2024-01-10-mobile-app/
├── mrd.md
├── competitive-analysis.md
├── business-case.md
└── market-research.md
```

**Created By:** Strategist role

---

## docs/product/ - Product Requirements

**Purpose:** Product requirements, features, epics, and user stories

**Directory Naming:** `YYYY-MM-DD-feature-name/`
- Date: When PRD was created
- Name: Feature or product area name

**Contents:**
- `prd.md` - Product Requirements Document
- `epics.md` - Epic breakdown
- `user-stories.md` - User stories with acceptance criteria

**Example:**
```
docs/product/2024-01-15-user-authentication/
├── prd.md
├── epics.md
└── user-stories.md
```

**Created By:** Product Manager role

---

## docs/design/ - UX/UI Design

**Purpose:** User experience design, wireframes, and interaction flows

**Directory Naming:** `YYYY-MM-DD-feature-name/`
- Date: When design work started
- Name: Feature being designed

**Contents:**
- `design-specs.md` - Design specifications
- `user-flows.md` - User flow diagrams
- `user-research.md` - User research findings
- `wireframes/` - Directory of HTML wireframes
  - `web-*.html` - Web wireframes
  - `ios-*.html` - iOS wireframes
  - `android-*.html` - Android wireframes

**Example:**
```
docs/design/2024-01-18-user-authentication/
├── design-specs.md
├── user-flows.md
├── user-research.md
└── wireframes/
    ├── web-login.html
    ├── web-signup.html
    ├── ios-login.html
    └── android-login.html
```

**Created By:** Designer role

---

## docs/architecture/ - Technical Architecture

**Purpose:** System architecture, API specifications, and data models

**Directory Naming:** `YYYY-MM-DD-feature-name/`
- Date: When architecture design started
- Name: Feature or system being designed

**Contents:**
- `architecture.md` - System architecture overview
- `api-spec.md` - API specifications
- `data-models.md` - Database schema and data structures
- `sequence-diagrams.md` - Interaction diagrams
- `feasibility-assessment.md` - Technical feasibility analysis (optional)

**Example:**
```
docs/architecture/2024-01-20-user-authentication/
├── architecture.md
├── api-spec.md
├── data-models.md
└── sequence-diagrams.md
```

**Created By:** Architect role

---

## docs/adr/ - Architecture Decision Records

**Purpose:** Document significant architecture and design decisions

**File Naming:** `NNN-decision-title.md`
- **NNN**: Sequential number (001, 002, ...)
- **Numbering:** Sequential across **entire project** (not per-feature)
- **Format:** Lowercase with hyphens

**Index:** `docs/adr/README.md` - Index of all ADRs by status and category

**Example:**
```
docs/adr/
├── 001-use-graphql-federation.md
├── 002-postgresql-for-transactions.md
├── 003-event-sourcing-for-billing.md
└── README.md
```

**Template:** `.ai-pack/templates/adr/adr-template.md`
**Created By:** Architect role

**See:** `.ai-pack/quality/clean-code/11-documentation-standards.md` for ADR format

---

## docs/security/ - Security Vulnerabilities

**Purpose:** Track security vulnerabilities, audits, and remediation

**File Naming:** `SEC-YYYY-NNN-short-description.md`
- **SEC**: Security prefix
- **YYYY**: Year discovered
- **NNN**: Sequential number within year (001-999)
- **Numbering:** Sequential across all vulnerabilities (not per-severity)

**Index:** `docs/security/README.md` - Index by severity, status, and OWASP category

**Example:**
```
docs/security/
├── SEC-2024-001-sql-injection.md
├── SEC-2024-002-xss-profile.md
├── SEC-2024-003-csrf-api.md
└── README.md
```

**Template:** `.ai-pack/templates/security/vulnerability-template.md`
**Created By:** Security team, code reviews, penetration testing

**CVE Reference:** Internal SEC IDs can reference CVE IDs if publicly disclosed

---

## docs/incidents/ - Production Incidents

**Purpose:** Document production incidents, RCA, and post-mortems

**File Naming:** `INC-YYYY-NNN-short-description.md`
- **INC**: Incident prefix
- **YYYY**: Year of incident
- **NNN**: Sequential number within year (001-999)

**Index:** `docs/incidents/README.md` - Index by severity and status

**Example:**
```
docs/incidents/
├── INC-2024-001-database-outage.md
├── INC-2024-002-api-timeout.md
├── INC-2024-003-deployment-rollback.md
└── README.md
```

**Template:** `.ai-pack/templates/incidents/incident-template.md`
**Created By:** Spelunker role (for runtime investigation)

**Severity Levels:** SEV-1 (Critical) through SEV-4 (Low)

---

## docs/investigations/ - Bug Retrospectives

**Purpose:** Root cause analysis and retrospectives for bugs

**File Naming:** `BUG-YYYY-NNN-short-description.md`
- **BUG**: Bug investigation prefix
- **YYYY**: Year bug was reported
- **NNN**: Sequential number within year (001-999)

**Index:** `docs/investigations/README.md` - Index by pattern/root cause category

**Example:**
```
docs/investigations/
├── BUG-2024-001-null-pointer-payment.md
├── BUG-2024-002-race-condition-orders.md
├── BUG-2024-003-validation-bypass-auth.md
├── README.md
└── patterns/
    ├── null-reference-bugs.md
    ├── race-conditions.md
    └── validation-issues.md
```

**Template:** `.ai-pack/templates/investigations/retrospective-template.md`
**Created By:** Inspector role (for code investigation)

**Pattern Docs:** Group similar bugs for systemic improvement

---

## docs/archaeology/ - Historical Investigations

**Purpose:** Historical system analysis, evolution timelines, and pattern discovery

**File Naming:** `[system-name]-[aspect].md`
- System name in kebab-case
- Aspect: evolution, decisions, debt, patterns, onboarding

**Index:** `docs/archaeology/README.md` - Index of investigations by system

**Example:**
```
docs/archaeology/
├── authentication-system-evolution.md
├── authentication-system-decisions.md
├── authentication-system-debt.md
├── payment-system-evolution.md
├── payment-system-patterns.md
└── README.md
```

**Created By:** Archaeologist role

---

## Cross-Referencing

All documents MUST include a "Related Documents" section for traceability:

**MRD → PRD:**
```markdown
## Related Documents
- PRD: [User Auth](../product/2024-01-15-user-authentication/prd.md)
```

**PRD → Architecture:**
```markdown
## Related Documents
- MRD: [Mobile App](../market/2024-01-10-mobile-app/mrd.md)
- Architecture: [User Auth](../architecture/2024-01-20-user-authentication/architecture.md)
- ADRs: [ADR-005](../adr/005-jwt-tokens.md)
```

**Architecture → Implementation:**
```go
// Implements: docs/architecture/2024-01-20-user-authentication/api-spec.md
// Requirement: FR-10 from docs/product/2024-01-15-user-authentication/prd.md
func AuthenticateUser(username, password string) (*User, error) {
    // ...
}
```

---

## Traceability Chain

```
Market (MRD) → Product (PRD) → Design → Architecture (ADRs) → Implementation → Tests
```

Each document links to:
- **Upstream:** What drives this (MRD → PRD → Architecture)
- **Downstream:** What this informs (Architecture → Implementation)
- **Related:** ADRs, security issues, incidents

---

## README Templates

Each `docs/` subdirectory MUST have a `README.md` that serves as an index:

**Available Templates:**
- `.ai-pack/templates/adr/README-template.md` - ADR index
- `.ai-pack/templates/security/README-template.md` - Security vulnerability index
- `.ai-pack/templates/incidents/README-template.md` - Incident index
- `.ai-pack/templates/investigations/README-template.md` - Bug retrospective index

---

## Sorting and Organization

**Chronological (Date-Prefixed):**
```bash
ls docs/product/
# Shows oldest first by default
2024-01-15-user-authentication/
2024-02-20-payment-system/
2024-03-10-notification-service/
```

**Sequential (ID-Prefixed):**
```bash
ls docs/security/
# Shows by ID
SEC-2024-001-sql-injection.md
SEC-2024-002-xss-profile.md
SEC-2024-003-csrf-api.md
```

**By Status (via README):**
- README.md groups by status (Open, Remediated, etc.)
- README.md provides statistics and trends
- README.md enables pattern detection

---

## Summary Table

| Directory | Naming | Prefix | Example | Created By |
|-----------|--------|--------|---------|------------|
| `market/` | Date-prefixed | - | `2024-01-10-mobile-app/` | Strategist |
| `product/` | Date-prefixed | - | `2024-01-15-user-auth/` | Product Manager |
| `design/` | Date-prefixed | - | `2024-01-18-user-auth/` | Designer |
| `architecture/` | Date-prefixed | - | `2024-01-20-user-auth/` | Architect |
| `adr/` | Sequential | - | `001-decision.md` | Architect |
| `security/` | Year + Sequential | `SEC` | `SEC-2024-001-desc.md` | Security |
| `incidents/` | Year + Sequential | `INC` | `INC-2024-001-desc.md` | Spelunker |
| `investigations/` | Year + Sequential | `BUG` | `BUG-2024-001-desc.md` | Inspector |
| `archaeology/` | System-based | - | `system-evolution.md` | Archaeologist |

---

**Version:** 1.0.0
**Last Updated:** 2026-01-14
**See Also:**
- [Persistence Gates](gates/10-persistence.md)
- [Architect Role](roles/architect.md)
- [Product Manager Role](roles/product-manager.md)
- [Strategist Role](roles/strategist.md)
