# Java Language-Specific Rules

> Based on Google Java Style Guide, Effective Java, Spring Framework best practices, and SonarQube code quality rules

## Formatting Standards (Java Specific)

**Indentation:** **2 spaces** (no tabs)

This is a **Cortexa LLC override** of the Google Java Style Guide (which specifies 4 spaces). We use 2-space indentation for consistency with our C++ and JavaScript/TypeScript codebases.

**Note:** Cortexa LLC uses **language-specific indentation standards**:
- **C++**: 2 spaces (see lang-cpp.md)
- **Python**: 4 spaces (see lang-python.md)
- **JavaScript/TypeScript**: 2 spaces (see lang-javascript.md)
- **Java**: 2 spaces (this document)
- **Kotlin**: 4 spaces (see lang-kotlin.md)

**Example:**
```java
package com.cortexa.services;

import java.util.List;
import java.util.Optional;

/**
 * Service for managing user data.
 */
public class UserService {
  private final UserRepository repository;
  private final EmailService emailService;

  public UserService(UserRepository repository, EmailService emailService) {
    this.repository = repository;
    this.emailService = emailService;
  }

  /**
   * Retrieves a user by ID.
   *
   * @param id the user ID
   * @return an Optional containing the user if found
   */
  public Optional<User> getUserById(Long id) {
    // 2-space indentation throughout
    if (id == null || id <= 0) {
      throw new IllegalArgumentException("ID must be positive");
    }

    return repository.findById(id)
      .map(user -> {
        user.setLastAccessed(LocalDateTime.now());
        return user;
      });
  }

  /**
   * Creates a new user.
   *
   * @param userData the user data
   * @return the created user
   */
  public User createUser(UserData userData) {
    var user = User.builder()
      .name(userData.getName())
      .email(userData.getEmail())
      .createdAt(LocalDateTime.now())
      .build();

    user = repository.save(user);
    emailService.sendWelcomeEmail(user);

    return user;
  }
}
```

---

## Overview

This file contains Java-specific best practices including:
- **Google Java Style Guide** - Industry-standard conventions
- **Effective Java** - Joshua Bloch's best practices
- **Spring Framework** - Enterprise Java patterns
- **Java 17+ Features** - Modern Java features

---

## Quick Standards Summary

### Formatting
- **Indentation:** 2 spaces (no tabs) - **Cortexa LLC override**
- **Brace Style:** K&R (opening brace on same line)
- **Line Length:** 100 characters (Google Java Style Guide)
- **One Statement Per Line:** No multiple statements on one line
- **Block Indentation:** +2 spaces

### Naming
- `packagename` - all lowercase
- `ClassName` - UpperCamelCase (PascalCase)
- `methodName` - lowerCamelCase
- `variableName` - lowerCamelCase
- `CONSTANT_NAME` - UPPER_SNAKE_CASE
- `_privateField` - (avoid, use private keyword)

### Class Structure (Order)
```java
public class Example {
  // 1. Static fields
  private static final String CONSTANT = "value";

  // 2. Instance fields
  private final Dependency dependency;
  private String mutableField;

  // 3. Constructors
  public Example(Dependency dependency) {
    this.dependency = dependency;
  }

  // 4. Public methods
  public void publicMethod() {
    // ...
  }

  // 5. Package-private methods
  void packagePrivateMethod() {
    // ...
  }

  // 6. Protected methods
  protected void protectedMethod() {
    // ...
  }

  // 7. Private methods
  private void privateMethod() {
    // ...
  }

  // 8. Static nested classes
  public static class Builder {
    // ...
  }

  // 9. Non-static nested classes
  private class Inner {
    // ...
  }
}
```

### Modern Java Features (Java 17+)
```java
// Records (immutable data carriers)
public record User(Long id, String name, String email) {
  // Compact constructor with validation
  public User {
    if (name == null || name.isBlank()) {
      throw new IllegalArgumentException("Name cannot be blank");
    }
  }
}

// Sealed classes (restrict inheritance)
public sealed interface Shape
  permits Circle, Rectangle, Triangle {
  double area();
}

// Pattern matching for instanceof
public String format(Object obj) {
  return switch (obj) {
    case Integer i -> String.format("int %d", i);
    case String s -> String.format("String %s", s);
    case User(var id, var name, var email) -> // Record pattern
      String.format("User[id=%d, name=%s]", id, name);
    default -> obj.toString();
  };
}

// Text blocks
String query = """
  SELECT u.id, u.name, u.email
  FROM users u
  WHERE u.active = true
  ORDER BY u.created_at DESC
  """;

// var for local variables (Java 10+)
var users = userRepository.findAll();  // Type inferred
var result = process(data);

// Stream API
List<String> activeUserNames = users.stream()
  .filter(User::isActive)
  .map(User::getName)
  .sorted()
  .collect(Collectors.toList());
```

### Effective Java Patterns
```java
// Use Optional for return types
public Optional<User> findUser(Long id) {
  return Optional.ofNullable(users.get(id));
}

// Builder pattern for complex objects
public class Config {
  private final String apiUrl;
  private final int timeout;
  private final int retries;

  private Config(Builder builder) {
    this.apiUrl = builder.apiUrl;
    this.timeout = builder.timeout;
    this.retries = builder.retries;
  }

  public static class Builder {
    private String apiUrl;
    private int timeout = 5000;
    private int retries = 3;

    public Builder apiUrl(String apiUrl) {
      this.apiUrl = apiUrl;
      return this;
    }

    public Builder timeout(int timeout) {
      this.timeout = timeout;
      return this;
    }

    public Builder retries(int retries) {
      this.retries = retries;
      return this;
    }

    public Config build() {
      return new Config(this);
    }
  }
}

// Try-with-resources for AutoCloseable
try (var reader = new BufferedReader(new FileReader(path))) {
  return reader.lines().collect(Collectors.toList());
}

// Static factory methods
public class User {
  public static User of(String name, String email) {
    return new User(generateId(), name, email);
  }

  public static User guest() {
    return new User(0L, "Guest", "guest@example.com");
  }
}
```

---

## Code Quality: SonarQube Enforcement

**MANDATORY:** All Java code must pass SonarQube's default Java rules for code smell detection.

### SonarQube Integration

**Official Reference:** https://rules.sonarsource.com/java/

SonarQube provides automated static code analysis with **733 total rules** organized into:

| Category | Count | Focus |
|----------|-------|-------|
| **Code Smell** | 458 | Maintainability and design issues |
| **Bug** | 175 | Correctness and logic defects |
| **Vulnerability** | 60 | Security weaknesses in code |
| **Security Hotspot** | 40 | Security-sensitive operations requiring review |

### Critical Security Vulnerabilities (Top Priority)

**All 60 vulnerability rules are MANDATORY:**

- **SQL Injection:** Database queries must use prepared statements, never string concatenation
- **XML External Entity (XXE):** XML parsers must disable external entity processing
- **Deserialization Attacks:** Deserialization must validate input to prevent malicious object instantiation
- **Weak Cryptography:** Cipher algorithms must be robust (AES-256, RSA-2048+)
- **JWT Security:** JSON Web Tokens must use strong signing algorithms (RS256, ES256)
- **Path Traversal:** File paths must be validated to prevent directory traversal attacks
- **Cross-Site Scripting (XSS):** User input must be sanitized before display
- **LDAP Injection:** LDAP queries must use parameterized queries
- **Command Injection:** OS commands must never include unsanitized user input
- **Hard-coded Credentials:** Never embed passwords, API keys, or secrets in code

### Critical Bug Detection (175 Rules)

**All bug rules are MANDATORY:**

- **Null Pointer Dereference:** Check for null before dereferencing
- **Resource Leaks:** Always close streams, connections, and files (use try-with-resources)
- **Concurrent Modification:** Don't modify collections while iterating
- **ClassCastException:** Verify types before casting
- **Array Index Out of Bounds:** Validate array indices
- **Division by Zero:** Check divisor before division operations
- **Infinite Loops:** Ensure loop termination conditions
- **Thread Safety:** Synchronize access to shared mutable state
- **Incorrect Synchronization:** Synchronize on appropriate objects
- **Return Value Ignored:** Check return values of critical operations

**Default Java Rules Include:**

#### Critical Code Smells
- **Cognitive Complexity:** Methods should not exceed cognitive complexity threshold (15)
- **Cyclomatic Complexity:** Methods should not exceed cyclomatic complexity (10)
- **Method Length:** Methods should be short and focused (max 50 lines)
- **Class Coupling:** Classes should not have too many dependencies
- **Inheritance Depth:** Inheritance tree should not be too deep (max 5 levels)
- **Too Many Parameters:** Methods should not have too many parameters (max 7)

#### Common Anti-Patterns
- **God Class:** Classes that do too much
- **Feature Envy:** Methods that access data from other classes excessively
- **Shotgun Surgery:** Changes requiring modifications across many classes
- **Dead Code:** Unused private methods, fields, or local variables
- **Duplicated Code:** Identical or very similar code blocks
- **Magic Numbers:** Unexplained numeric literals (use constants)
- **Long Parameter List:** Methods with many parameters (use parameter objects)

#### Naming Conventions
- **Class Names:** UpperCamelCase (PascalCase)
- **Method Names:** lowerCamelCase, verb-based
- **Constant Names:** UPPER_SNAKE_CASE
- **Package Names:** all lowercase, reverse domain notation
- **Boolean Names:** Should ask a question (isActive, hasPermission, canEdit)

#### Resource Management
- **Close Resources:** Always use try-with-resources for AutoCloseable
- **Avoid Resource Leaks:** Ensure streams, connections, files are properly closed
- **Null Checks:** Prefer Optional over null returns
- **Exception Handling:** Don't catch generic Exception, be specific

#### Security Rules
- **SQL Injection:** Use prepared statements, never concatenate SQL
- **Path Traversal:** Validate file paths
- **XSS Prevention:** Sanitize user input
- **Secure Random:** Use SecureRandom for cryptographic purposes
- **Credentials:** Never hardcode passwords or API keys

#### Best Practices
- **Immutability:** Prefer immutable objects where possible
- **Thread Safety:** Document thread-safety guarantees
- **Equals and HashCode:** Override both or neither
- **ToString:** Provide meaningful toString implementations
- **Serialization:** Be careful with Serializable, consider alternatives

### Configuration

Add SonarQube to your project:

```xml
<!-- Maven pom.xml -->
<properties>
  <sonar.java.source>17</sonar.java.source>
  <sonar.coverage.jacoco.xmlReportPaths>
    target/site/jacoco/jacoco.xml
  </sonar.coverage.jacoco.xmlReportPaths>
</properties>

<profiles>
  <profile>
    <id>sonar</id>
    <activation>
      <activeByDefault>false</activeByDefault>
    </activation>
    <properties>
      <sonar.host.url>https://sonarqube.cortexa.com</sonar.host.url>
    </properties>
  </profile>
</profiles>
```

```gradle
// Gradle build.gradle
plugins {
  id "org.sonarqube" version "4.0.0.2929"
}

sonar {
  properties {
    property "sonar.projectKey", "your-project-key"
    property "sonar.organization", "cortexa"
    property "sonar.host.url", "https://sonarqube.cortexa.com"
    property "sonar.java.source", "17"
  }
}
```

### Running SonarQube Analysis

```bash
# Maven
mvn clean verify sonar:sonar

# Gradle
./gradlew sonarqube

# With authentication token
mvn sonar:sonar -Dsonar.login=your-token
```

### Quality Gate Requirements

All code must meet these thresholds:
- **New Code Coverage:** ≥ 80%
- **Duplicated Lines:** < 3%
- **Maintainability Rating:** A
- **Reliability Rating:** A
- **Security Rating:** A
- **Security Hotspots Reviewed:** 100%
- **Blocker/Critical Issues:** 0

### IDE Integration

**IntelliJ IDEA:**
```
Settings → Plugins → Install "SonarLint"
Settings → Tools → SonarLint → Bind to SonarQube
```

**VS Code:**
```
Extensions → Install "SonarLint"
Settings → SonarLint: Connected Mode
```

### Pre-commit Hook (Optional)

Add to `.git/hooks/pre-commit`:
```bash
#!/bin/bash
# Run SonarQube analysis on staged files
mvn sonar:sonar -Dsonar.analysis.mode=preview
if [ $? -ne 0 ]; then
  echo "SonarQube analysis failed. Commit rejected."
  exit 1
fi
```

---

## TODO: Full Java Guidelines

This file will be expanded to include:
- [ ] Complete Effective Java coverage
- [ ] Spring Framework best practices
- [ ] JPA/Hibernate patterns
- [ ] Testing with JUnit 5 and Mockito
- [ ] Exception handling strategies
- [ ] Concurrency and threading
- [ ] Collections and Streams
- [ ] Dependency Injection
- [ ] Common anti-patterns
- [ ] Performance optimization

---

**For now, always use 2-space indentation (Cortexa LLC override). All code must pass SonarQube's default Java rules. Full guidelines coming soon.**
