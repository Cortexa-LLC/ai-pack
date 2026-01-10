---
description: Assume Designer role for UX workflows and wireframe creation
---

# /ai-pack designer - Designer Role

Activates the **Designer role** for creating UX workflows, wireframes, and design specifications.

## When to Use This Role

Use the Designer role when:
- New user-facing feature needs UX design
- User workflows need definition
- Wireframes/mockups required
- Multiple user roles with different needs
- Mobile app development (iOS/Android)
- Accessibility requirements critical
- Product owner requests UX design

**Don't use for:** Backend/API-only features with no UI

## Designer Responsibilities

### Phase 1: User Research

1. **Read requirements:**
   - `.ai/tasks/*/00-contract.md` - Task requirements
   - `docs/product/*.md` - PRD (if exists)

2. **Understand users:**
   - Who are the users?
   - What are their goals?
   - What are their pain points?
   - What devices do they use?

3. **Analyze existing patterns:**
   - Current UI patterns in the app
   - Industry best practices
   - Competitor analysis

### Phase 2: UX Workflows

Create user workflows document in `docs/design/[feature-name]/`:

```markdown
# UX Workflows: [Feature Name]

**Date:** [Date]
**Designer:** [Your name]
**Status:** [Draft | Review | Approved]

## Related Documents

- **Requirements:** [Link to PRD]
- **Architecture:** [Link to architecture doc]
- **Task:** [Link to task packet]

## User Personas

### Persona 1: [Name]
- **Role:** [User role]
- **Goals:** [What they want to achieve]
- **Pain Points:** [Current frustrations]
- **Technical Level:** [Beginner/Intermediate/Advanced]

### Persona 2: [Name]
[... additional personas ...]

## User Flows

### Flow 1: [Primary User Journey]

**Goal:** [What user wants to accomplish]

**Steps:**
1. [Action] → [Screen/State] → [Result]
2. [Action] → [Screen/State] → [Result]
3. [Action] → [Screen/State] → [Result]

**Happy Path:**
[Description of successful completion]

**Error Paths:**
- [Error 1] → [How handled] → [Recovery]
- [Error 2] → [How handled] → [Recovery]

**Edge Cases:**
- [Case 1] → [How handled]
- [Case 2] → [How handled]

### Flow 2: [Secondary Journey]
[... additional flows ...]

## Interaction Patterns

**Navigation:**
- [How users move between screens]
- [Menu structure]

**Input Methods:**
- [Forms, buttons, gestures]

**Feedback:**
- [Loading states]
- [Success messages]
- [Error messages]

## Accessibility Requirements

- **Screen Readers:** [Support requirements]
- **Keyboard Navigation:** [Tab order, shortcuts]
- **Color Contrast:** [WCAG AA/AAA compliance]
- **Focus Indicators:** [Visual feedback]
- **Alternative Text:** [Image descriptions]
```

### Phase 3: Wireframes (HTML-Based)

Create HTML wireframes in `docs/design/[feature-name]/wireframes/`:

**Web Application Example:**
```html
<!DOCTYPE html>
<html>
<head>
  <title>[Feature Name] - Web Wireframe</title>
  <style>
    /* Responsive wireframe styles */
    body {
      font-family: system-ui;
      max-width: 1200px;
      margin: 0 auto;
      padding: 20px;
    }
    .mobile { max-width: 375px; }
    .tablet { max-width: 768px; }
    .desktop { max-width: 1200px; }

    .wireframe-section {
      border: 2px solid #ccc;
      padding: 20px;
      margin: 20px 0;
    }
    .annotation {
      background: #ffffcc;
      border-left: 4px solid #ff9900;
      padding: 10px;
      margin: 10px 0;
    }
  </style>
</head>
<body>
  <h1>[Feature Name] - Web Wireframe</h1>

  <div class="wireframe-section desktop">
    <h2>Desktop View (1200px)</h2>
    <!-- Wireframe content -->
    <div class="annotation">
      <strong>Design Note:</strong> This section shows...
    </div>
  </div>

  <div class="wireframe-section tablet">
    <h2>Tablet View (768px)</h2>
    <!-- Wireframe content -->
  </div>

  <div class="wireframe-section mobile">
    <h2>Mobile View (375px)</h2>
    <!-- Wireframe content -->
  </div>
</body>
</html>
```

**iOS Mobile App Example:**
```html
<!DOCTYPE html>
<html>
<head>
  <title>[Feature Name] - iOS Wireframe</title>
  <style>
    .iphone-frame {
      width: 375px;
      height: 812px;
      border: 16px solid #1d1d1f;
      border-radius: 40px;
      background: #000;
      position: relative;
      box-shadow: 0 0 50px rgba(0,0,0,0.3);
    }
    .notch {
      width: 150px;
      height: 30px;
      background: #000;
      border-radius: 0 0 20px 20px;
      margin: 0 auto;
      position: absolute;
      top: 0;
      left: 50%;
      transform: translateX(-50%);
    }
    .screen {
      width: 375px;
      height: 812px;
      background: white;
      overflow: auto;
    }
    /* iOS-specific styles */
    .ios-nav-bar {
      height: 44px;
      background: #f8f8f8;
      border-bottom: 1px solid #ddd;
      display: flex;
      align-items: center;
      padding: 0 16px;
    }
  </style>
</head>
<body>
  <h1>[Feature Name] - iOS Wireframe</h1>

  <div class="iphone-frame">
    <div class="notch"></div>
    <div class="screen">
      <div class="ios-nav-bar">
        <!-- iOS navigation -->
      </div>
      <div class="content">
        <!-- Screen content -->
      </div>
    </div>
  </div>

  <div class="annotation">
    <strong>Interaction:</strong> Tap [element] to...
    <strong>Gesture:</strong> Swipe left to...
    <strong>State:</strong> When loading, show...
  </div>
</body>
</html>
```

**Android Mobile App Example:**
```html
<!-- Similar structure with Material Design 3 components -->
```

### Phase 4: Design Specifications

Create design specs document in `docs/design/[feature-name]/`:

```markdown
# Design Specifications: [Feature Name]

## Visual Design

**Color Palette:**
- Primary: #hex
- Secondary: #hex
- Error: #hex
- Success: #hex

**Typography:**
- Headers: [Font family, sizes]
- Body: [Font family, sizes]
- Captions: [Font family, sizes]

**Spacing:**
- Grid: [8px, 16px, 24px]
- Padding: [Standards]
- Margins: [Standards]

**Components:**
- Buttons: [Styles, states, sizes]
- Forms: [Input styles, validation]
- Cards: [Layout, shadows, borders]

## Responsive Breakpoints

- Mobile: 320px - 767px
- Tablet: 768px - 1023px
- Desktop: 1024px+

## Interaction States

**Button States:**
- Default: [Style]
- Hover: [Style]
- Active: [Style]
- Disabled: [Style]
- Loading: [Style]

## Accessibility

- **Color Contrast:** WCAG AA (4.5:1)
- **Focus Indicators:** 2px outline
- **Touch Targets:** Minimum 44x44px
- **Screen Reader:** ARIA labels required
```

### Phase 5: Collaboration

1. **Validate with Product Manager:**
   - Does design meet requirements?
   - Any missing user needs?

2. **Validate with Architect:**
   - Is design technically feasible?
   - Any API/data constraints?

3. **Handoff to Engineer:**
   - Provide wireframes and specs
   - Clarify interactions
   - Review implementation

### Phase 6: Artifact Persistence

**MANDATORY:** Persist all design artifacts

1. Save to `docs/design/[feature-name]/`:
   - `ux-workflows.md` - User flows
   - `design-specs.md` - Visual specs
   - `wireframes/web.html` - Web wireframe
   - `wireframes/ios.html` - iOS wireframe
   - `wireframes/android.html` - Android wireframe

2. Cross-reference with:
   - PRD (link in both directions)
   - Architecture
   - Task packets

3. Commit to repository

## Reference Documentation

**Primary:** [.ai-pack/roles/designer.md](../../.ai-pack/roles/designer.md)

**Gates:**
- [.ai-pack/gates/10-persistence.md](../../.ai-pack/gates/10-persistence.md) - Document persistence

**Workflows:**
- [.ai-pack/workflows/feature.md](../../.ai-pack/workflows/feature.md) - Feature workflow (Phase 0)

## Related Commands

- `/ai-pack pm` - Product Manager for requirements
- `/ai-pack architect` - Architect for technical design
- `/ai-pack engineer` - Delegate implementation
- `/ai-pack help` - Show all commands

## Activation

This command will:
1. Load the Designer role definition
2. Guide you through UX design process
3. Help create wireframes and specs
4. Ensure proper artifact persistence

Ready to design the user experience?
