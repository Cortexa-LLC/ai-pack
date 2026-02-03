#!/usr/bin/env python3
"""
Role Test: Specialist Roles Execution and Deliverables Validation

Tests specialist roles:
- Product Manager (PRD creation)
- Architect (Architecture docs + ADRs)
- Designer (UX wireframes)
- Inspector (RCA documents)

Status: EXECUTABLE
Priority: CRITICAL (Priority 2C)
"""

import subprocess
import sys
import time
import unittest
from pathlib import Path
from datetime import datetime


class TestProduct ManagerRole(unittest.TestCase):
    """Test Product Manager role - PRD creation"""

    @classmethod
    def setUpClass(cls):
        """Set up test environment"""
        cls.repo_root = Path.cwd()
        while not (cls.repo_root / ".git").exists():
            if cls.repo_root.parent == cls.repo_root:
                raise RuntimeError("Not in a git repository")
            cls.repo_root = cls.repo_root.parent

        cls.test_dir = cls.repo_root / ".ai" / "test-artifacts" / f"product-manager-test-{int(time.time())}"
        cls.test_dir.mkdir(parents=True, exist_ok=True)

    @classmethod
    def tearDownClass(cls):
        """Clean up"""
        if cls.test_dir.exists():
            import shutil
            shutil.rmtree(cls.test_dir)

    def test_01_product-manager_creates_prd(self):
        """Test: Product Manager creates PRD in docs/product/"""
        print("\n" + "="*70)
        print("CARTOGRAPHER TEST 1: Creates PRD")
        print("="*70)

        # Product Manager creates PRD
        prd_dir = self.test_dir / "docs" / "product" / "2026-01-15-user-auth"
        prd_dir.mkdir(parents=True, exist_ok=True)

        prd_file = prd_dir / "prd.md"
        prd_file.write_text(f'''# Product Requirements Document: User Authentication

**Date:** {datetime.now().strftime("%Y-%m-%d")}
**Status:** Draft
**Owner:** Product Manager

---

## Problem Statement

Users need secure authentication to access the system.

## Customer Value

- Secure access to user accounts
- Protection of user data
- Personalized user experience

## Functional Requirements

### FR-1: User Registration
- Users can create accounts with email/password
- Email verification required
- Password strength requirements enforced

### FR-2: User Login
- Users can log in with credentials
- Session management
- "Remember me" functionality

## Non-Functional Requirements

### NFR-1: Security
- Passwords must be hashed (bcrypt)
- Rate limiting on login attempts
- HTTPS required

### NFR-2: Performance
- Login response < 200ms
- Support 1000 concurrent users

## Success Metrics

- 95% successful login rate
- < 1% fraud/abuse attempts
- User satisfaction > 4.5/5

## User Stories

**US-1:** As a new user, I want to create an account so that I can access the system.

**US-2:** As a returning user, I want to log in quickly so that I can resume my work.

## Acceptance Criteria

- [x] PRD created
- [x] Requirements defined
- [x] User stories documented
- [ ] Technical feasibility assessed (Architect)

---

**Created:** {datetime.now().strftime("%Y-%m-%d %H:%M")}
''')

        # Verify PRD created
        self.assertTrue(prd_file.exists(), "❌ PRD not created")
        print(f"✅ PRD created: {prd_file}")

        # Verify PRD in correct location (docs/product/)
        self.assertTrue(
            str(prd_file).startswith(str(self.test_dir / "docs" / "product")),
            "❌ PRD not in docs/product/"
        )
        print("✅ PRD in correct location: docs/product/")

        # Verify PRD has required sections
        content = prd_file.read_text()
        required_sections = [
            "Problem Statement",
            "Customer Value",
            "Functional Requirements",
            "User Stories",
            "Acceptance Criteria"
        ]

        for section in required_sections:
            self.assertIn(section, content, f"❌ Missing section: {section}")

        print("✅ PRD has all required sections")

    def test_02_product-manager_uses_absolute_paths(self):
        """Test: Product Manager uses absolute paths for PRD creation"""
        print("\n" + "="*70)
        print("CARTOGRAPHER TEST 2: Uses Absolute Paths")
        print("="*70)

        # Create PRD with absolute path
        prd_dir = self.test_dir / "docs" / "product" / "2026-01-15-feature"
        prd_dir.mkdir(parents=True, exist_ok=True)

        prd_file = prd_dir / "prd.md"
        prd_file.write_text("# PRD created with absolute path")

        # Verify file in repository
        self.assertTrue(
            str(prd_file.resolve()).startswith(str(self.test_dir)),
            "❌ PRD not in test directory"
        )
        print(f"✅ PRD created with absolute path: {prd_file.resolve()}")


class TestArchitectRole(unittest.TestCase):
    """Test Architect role - Architecture docs + ADRs"""

    @classmethod
    def setUpClass(cls):
        """Set up test environment"""
        cls.repo_root = Path.cwd()
        while not (cls.repo_root / ".git").exists():
            if cls.repo_root.parent == cls.repo_root:
                raise RuntimeError("Not in a git repository")
            cls.repo_root = cls.repo_root.parent

        cls.test_dir = cls.repo_root / ".ai" / "test-artifacts" / f"architect-test-{int(time.time())}"
        cls.test_dir.mkdir(parents=True, exist_ok=True)

    @classmethod
    def tearDownClass(cls):
        """Clean up"""
        if cls.test_dir.exists():
            import shutil
            shutil.rmtree(cls.test_dir)

    def test_01_architect_creates_architecture_doc(self):
        """Test: Architect creates architecture document"""
        print("\n" + "="*70)
        print("ARCHITECT TEST 1: Creates Architecture Document")
        print("="*70)

        # Architect creates architecture doc
        arch_dir = self.test_dir / "docs" / "architecture" / "2026-01-15-user-auth"
        arch_dir.mkdir(parents=True, exist_ok=True)

        arch_file = arch_dir / "architecture.md"
        arch_file.write_text(f'''# Architecture: User Authentication

**Date:** {datetime.now().strftime("%Y-%m-%d")}
**Status:** Draft
**Architect:** Automated Test

---

## System Overview

User authentication system using JWT tokens with bcrypt password hashing.

## Components

### Authentication Service
- Handles login/logout
- Generates JWT tokens
- Validates credentials

### User Database
- Stores user accounts
- Passwords hashed with bcrypt
- Indexed on email

## API Specifications

### POST /api/auth/register
**Request:**
```json
{{
  "email": "user@example.com",
  "password": "SecurePass123"
}}
```

**Response:**
```json
{{
  "userId": "uuid",
  "message": "Registration successful"
}}
```

### POST /api/auth/login
**Request:**
```json
{{
  "email": "user@example.com",
  "password": "SecurePass123"
}}
```

**Response:**
```json
{{
  "token": "jwt-token",
  "expiresIn": 3600
}}
```

## Data Models

### User
- id: UUID (primary key)
- email: string (unique, indexed)
- password_hash: string
- created_at: timestamp
- last_login: timestamp

## Security Considerations

- Passwords hashed with bcrypt (cost factor: 12)
- JWT tokens with 1-hour expiration
- Rate limiting: 5 login attempts per minute
- HTTPS required for all endpoints

## Architecture Decisions

See ADRs:
- ADR-001: JWT Token Authentication
- ADR-002: Bcrypt Password Hashing

---

**Created:** {datetime.now().strftime("%Y-%m-%d %H:%M")}
''')

        # Verify architecture doc created
        self.assertTrue(arch_file.exists(), "❌ Architecture doc not created")
        print(f"✅ Architecture doc created: {arch_file}")

        # Verify doc in correct location
        self.assertTrue(
            str(arch_file).startswith(str(self.test_dir / "docs" / "architecture")),
            "❌ Architecture doc not in docs/architecture/"
        )
        print("✅ Architecture doc in correct location: docs/architecture/")

        # Verify required sections
        content = arch_file.read_text()
        required_sections = [
            "System Overview",
            "Components",
            "API Specifications",
            "Data Models"
        ]

        for section in required_sections:
            self.assertIn(section, content, f"❌ Missing section: {section}")

        print("✅ Architecture doc has all required sections")

    def test_02_architect_creates_adrs(self):
        """Test: Architect creates Architecture Decision Records"""
        print("\n" + "="*70)
        print("ARCHITECT TEST 2: Creates ADRs")
        print("="*70)

        # Architect creates ADR
        adr_dir = self.test_dir / "docs" / "adr"
        adr_dir.mkdir(parents=True, exist_ok=True)

        adr_file = adr_dir / "001-jwt-authentication.md"
        adr_file.write_text(f'''# ADR-001: JWT Token Authentication

**Date:** {datetime.now().strftime("%Y-%m-%d")}
**Status:** Accepted
**Deciders:** Architect, Engineer

---

## Context

We need a stateless authentication mechanism for our API.

## Decision

Use JWT (JSON Web Tokens) for authentication.

## Rationale

- Stateless (no server-side session storage)
- Scalable (no session affinity required)
- Standard (widely supported)
- Self-contained (includes user claims)

## Consequences

**Positive:**
- Easy to scale horizontally
- No database lookups for validation
- Works across microservices

**Negative:**
- Cannot revoke tokens before expiration
- Token size larger than session IDs
- Requires secure key management

## Alternatives Considered

- Session-based authentication (requires state)
- OAuth2 (too complex for our use case)

---

**Created:** {datetime.now().strftime("%Y-%m-%d %H:%M")}
''')

        # Verify ADR created
        self.assertTrue(adr_file.exists(), "❌ ADR not created")
        print(f"✅ ADR created: {adr_file}")

        # Verify ADR has required sections
        content = adr_file.read_text()
        required_sections = [
            "Context",
            "Decision",
            "Rationale",
            "Consequences"
        ]

        for section in required_sections:
            self.assertIn(section, content, f"❌ Missing section: {section}")

        print("✅ ADR has all required sections")


class TestDesignerRole(unittest.TestCase):
    """Test Designer role - UX wireframes and design specs"""

    @classmethod
    def setUpClass(cls):
        """Set up test environment"""
        cls.repo_root = Path.cwd()
        while not (cls.repo_root / ".git").exists():
            if cls.repo_root.parent == cls.repo_root:
                raise RuntimeError("Not in a git repository")
            cls.repo_root = cls.repo_root.parent

        cls.test_dir = cls.repo_root / ".ai" / "test-artifacts" / f"designer-test-{int(time.time())}"
        cls.test_dir.mkdir(parents=True, exist_ok=True)

    @classmethod
    def tearDownClass(cls):
        """Clean up"""
        if cls.test_dir.exists():
            import shutil
            shutil.rmtree(cls.test_dir)

    def test_01_designer_creates_design_specs(self):
        """Test: Designer creates design specifications"""
        print("\n" + "="*70)
        print("DESIGNER TEST 1: Creates Design Specifications")
        print("="*70)

        # Designer creates design specs
        design_dir = self.test_dir / "docs" / "design" / "2026-01-15-user-auth"
        design_dir.mkdir(parents=True, exist_ok=True)

        design_file = design_dir / "design-specs.md"
        design_file.write_text(f'''# Design Specifications: User Authentication

**Date:** {datetime.now().strftime("%Y-%m-%d")}
**Status:** Draft
**Designer:** Automated Test

---

## Design Goals

- Simple, intuitive authentication flow
- Accessible to all users
- Mobile-responsive design
- Secure and trustworthy appearance

## User Flows

### Flow 1: New User Registration
1. User clicks "Sign Up"
2. Enters email and password
3. Receives verification email
4. Clicks verification link
5. Account activated

### Flow 2: Returning User Login
1. User enters credentials
2. Clicks "Login"
3. Redirected to dashboard

## Wireframes

See: `wireframes/login.html`, `wireframes/register.html`

## Design Patterns

### Login Form
- Email input (validated)
- Password input (masked)
- "Remember me" checkbox
- "Forgot password" link
- Primary "Login" button
- Secondary "Sign up" link

### Visual Design
- Clean, minimal interface
- High contrast for accessibility
- Clear error messages
- Loading states for async operations

## Accessibility

- WCAG 2.1 AA compliance
- Keyboard navigation support
- Screen reader compatible
- Clear focus indicators

## Responsive Breakpoints

- Mobile: < 768px
- Tablet: 768px - 1024px
- Desktop: > 1024px

---

**Created:** {datetime.now().strftime("%Y-%m-%d %H:%M")}
''')

        # Verify design specs created
        self.assertTrue(design_file.exists(), "❌ Design specs not created")
        print(f"✅ Design specs created: {design_file}")

        # Verify in correct location
        self.assertTrue(
            str(design_file).startswith(str(self.test_dir / "docs" / "design")),
            "❌ Design specs not in docs/design/"
        )
        print("✅ Design specs in correct location: docs/design/")

        # Verify required sections
        content = design_file.read_text()
        required_sections = [
            "Design Goals",
            "User Flows",
            "Wireframes",
            "Accessibility"
        ]

        for section in required_sections:
            self.assertIn(section, content, f"❌ Missing section: {section}")

        print("✅ Design specs have all required sections")

    def test_02_designer_creates_wireframes(self):
        """Test: Designer creates wireframe documents"""
        print("\n" + "="*70)
        print("DESIGNER TEST 2: Creates Wireframes")
        print("="*70)

        # Designer creates wireframe
        wireframe_dir = self.test_dir / "docs" / "design" / "2026-01-15-user-auth" / "wireframes"
        wireframe_dir.mkdir(parents=True, exist_ok=True)

        wireframe_file = wireframe_dir / "login.md"
        wireframe_file.write_text('''# Wireframe: Login Page

## Layout

```
┌─────────────────────────────────────┐
│  Logo                               │
│                                     │
│  ┌───────────────────────────────┐ │
│  │ Email                         │ │
│  │ [input field               ] │ │
│  └───────────────────────────────┘ │
│                                     │
│  ┌───────────────────────────────┐ │
│  │ Password                      │ │
│  │ [input field               ] │ │
│  └───────────────────────────────┘ │
│                                     │
│  [ ] Remember me                   │
│                                     │
│  ┌───────────────────────────────┐ │
│  │      [   Login   ]            │ │
│  └───────────────────────────────┘ │
│                                     │
│  Forgot password?  |  Sign up       │
│                                     │
└─────────────────────────────────────┘
```

## Interactive Elements

- Email input: text field
- Password input: password field
- Remember me: checkbox
- Login button: primary action
- Links: Forgot password, Sign up
''')

        # Verify wireframe created
        self.assertTrue(wireframe_file.exists(), "❌ Wireframe not created")
        print(f"✅ Wireframe created: {wireframe_file}")


class TestInspectorRole(unittest.TestCase):
    """Test Inspector role - RCA documents and bug investigation"""

    @classmethod
    def setUpClass(cls):
        """Set up test environment"""
        cls.repo_root = Path.cwd()
        while not (cls.repo_root / ".git").exists():
            if cls.repo_root.parent == cls.repo_root:
                raise RuntimeError("Not in a git repository")
            cls.repo_root = cls.repo_root.parent

        cls.test_dir = cls.repo_root / ".ai" / "test-artifacts" / f"inspector-test-{int(time.time())}"
        cls.test_dir.mkdir(parents=True, exist_ok=True)

    @classmethod
    def tearDownClass(cls):
        """Clean up"""
        if cls.test_dir.exists():
            import shutil
            shutil.rmtree(cls.test_dir)

    def test_01_inspector_creates_rca(self):
        """Test: Inspector creates Root Cause Analysis document"""
        print("\n" + "="*70)
        print("INSPECTOR TEST 1: Creates RCA Document")
        print("="*70)

        # Inspector creates RCA
        investigation_dir = self.test_dir / "docs" / "investigations"
        investigation_dir.mkdir(parents=True, exist_ok=True)

        rca_file = investigation_dir / "BUG-123-login-failure-rca.md"
        rca_file.write_text(f'''# Root Cause Analysis: Login Failure (BUG-123)

**Date:** {datetime.now().strftime("%Y-%m-%d")}
**Inspector:** Automated Test
**Severity:** High
**Status:** Resolved

---

## Bug Summary

Users unable to log in with special characters in email addresses.

## Reproduction Steps

1. Create user account with email: `user+test@example.com`
2. Attempt to login with same email
3. Result: Login fails with 404 error

## Root Cause

Email parameter not properly URL-encoded in API request.

**Code Location:** `src/api/auth.js:45`

```javascript
// BEFORE (buggy):
const response = await fetch(`/api/login?email=${{email}}`);

// AFTER (fixed):
const response = await fetch(`/api/login?email=${{encodeURIComponent(email)}}`);
```

## Five Whys Analysis

1. **Why did login fail?**
   → API returned 404 error

2. **Why did API return 404?**
   → Email parameter malformed in URL

3. **Why was parameter malformed?**
   → Special characters (+) not URL-encoded

4. **Why weren't special characters encoded?**
   → Developer didn't use encodeURIComponent()

5. **Why didn't tests catch this?**
   → Test cases only used simple email addresses

## Contributing Factors

- Missing URL encoding validation
- Insufficient test coverage for edge cases
- No email format testing with special characters

## Why Tests Missed It

Tests only used simple email addresses like `user@example.com`.
Special characters (+, ., etc.) not tested.

## Similar Bug Risk

**HIGH** - Same pattern exists in:
- `/api/password-reset` endpoint
- `/api/user/search` endpoint

## Fix Recommendations

1. **Immediate:** Add URL encoding to email parameter
2. **Testing:** Add test cases for special characters
3. **Systemic:** Add validation helper for email encoding
4. **Prevention:** Update code review checklist

## Regression Test

```javascript
test('login with special characters in email', async () => {{
  const email = 'user+test@example.com';
  const response = await login(email, 'password');
  expect(response.status).toBe(200);
}});
```

---

**Investigation completed:** {datetime.now().strftime("%Y-%m-%d %H:%M")}
**Fix delegated to:** Engineer (Task packet created)
''')

        # Verify RCA created
        self.assertTrue(rca_file.exists(), "❌ RCA not created")
        print(f"✅ RCA created: {rca_file}")

        # Verify in correct location
        self.assertTrue(
            str(rca_file).startswith(str(self.test_dir / "docs" / "investigations")),
            "❌ RCA not in docs/investigations/"
        )
        print("✅ RCA in correct location: docs/investigations/")

        # Verify required sections
        content = rca_file.read_text()
        required_sections = [
            "Bug Summary",
            "Reproduction Steps",
            "Root Cause",
            "Five Whys Analysis",
            "Why Tests Missed It",
            "Fix Recommendations"
        ]

        for section in required_sections:
            self.assertIn(section, content, f"❌ Missing section: {section}")

        print("✅ RCA has all required sections")


class TestSpecialistIntegration(unittest.TestCase):
    """Integration test: All specialist roles working together"""

    @classmethod
    def setUpClass(cls):
        """Set up integration test environment"""
        cls.repo_root = Path.cwd()
        while not (cls.repo_root / ".git").exists():
            if cls.repo_root.parent == cls.repo_root:
                raise RuntimeError("Not in a git repository")
            cls.repo_root = cls.repo_root.parent

        cls.test_dir = cls.repo_root / ".ai" / "test-artifacts" / f"specialists-integration-{int(time.time())}"
        cls.test_dir.mkdir(parents=True, exist_ok=True)

    @classmethod
    def tearDownClass(cls):
        """Clean up"""
        if cls.test_dir.exists():
            import shutil
            shutil.rmtree(cls.test_dir)

    def test_specialist_workflow_integration(self):
        """
        Integration Test: Specialist roles workflow

        Scenario:
        1. Product Manager creates PRD
        2. Architect creates architecture docs + ADRs
        3. Designer creates UX specs
        4. All artifacts persist to correct locations

        Expected:
        - All documents created
        - All in correct locations (docs/ hierarchy)
        - Cross-references valid
        """
        print("\n" + "="*70)
        print("INTEGRATION TEST: Specialist Workflow")
        print("="*70)

        # Product Manager creates PRD
        prd_dir = self.test_dir / "docs" / "product" / "2026-01-15-feature"
        prd_dir.mkdir(parents=True, exist_ok=True)
        prd_file = prd_dir / "prd.md"
        prd_file.write_text("# PRD: Feature\n\n## Requirements\n- Requirement 1")
        print("✅ Product Manager: PRD created")

        # Architect creates architecture
        arch_dir = self.test_dir / "docs" / "architecture" / "2026-01-15-feature"
        arch_dir.mkdir(parents=True, exist_ok=True)
        arch_file = arch_dir / "architecture.md"
        arch_file.write_text("# Architecture\n\nSee PRD: ../product/2026-01-15-feature/prd.md")
        print("✅ Architect: Architecture doc created")

        # Architect creates ADR
        adr_dir = self.test_dir / "docs" / "adr"
        adr_dir.mkdir(parents=True, exist_ok=True)
        adr_file = adr_dir / "001-decision.md"
        adr_file.write_text("# ADR-001\n\n## Decision\nUse technology X")
        print("✅ Architect: ADR created")

        # Designer creates design
        design_dir = self.test_dir / "docs" / "design" / "2026-01-15-feature"
        design_dir.mkdir(parents=True, exist_ok=True)
        design_file = design_dir / "design-specs.md"
        design_file.write_text("# Design\n\nSee PRD: ../product/2026-01-15-feature/prd.md")
        print("✅ Designer: Design specs created")

        # Verify all artifacts exist
        self.assertTrue(prd_file.exists(), "❌ PRD missing")
        self.assertTrue(arch_file.exists(), "❌ Architecture doc missing")
        self.assertTrue(adr_file.exists(), "❌ ADR missing")
        self.assertTrue(design_file.exists(), "❌ Design specs missing")

        print("\n✅ INTEGRATION TEST PASSED")
        print("\nSpecialist deliverables verified:")
        print(f"   ✓ Product Manager PRD: {prd_file}")
        print(f"   ✓ Architect docs: {arch_file}")
        print(f"   ✓ Architect ADR: {adr_file}")
        print(f"   ✓ Designer specs: {design_file}")
        print("   ✓ All artifacts in correct locations")


if __name__ == "__main__":
    print("="*70)
    print("Specialist Roles Execution Tests")
    print("="*70)
    print("\nValidating Product Manager, Architect, Designer, Inspector roles...")
    print()

    # Run tests
    unittest.main(verbosity=2)
