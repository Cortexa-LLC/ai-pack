# Java Language-Specific Rules

> Based on Oracle Java Code Conventions, Google Java Style Guide, and Spring Framework best practices

## Formatting Standards (Java Specific)

**Indentation:** **4 spaces** (no tabs)

This follows Oracle Java Code Conventions and is the overwhelming standard in enterprise Java development (Spring, Apache, most projects).

**Note:** Cortexa LLC uses **language-specific indentation standards**:
- **C++**: 2 spaces (see lang-cpp.md)
- **Python**: 4 spaces (see lang-python.md)
- **JavaScript/TypeScript**: 2 spaces (see lang-javascript.md)
- **Java**: 4 spaces (this document)
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
        // 4-space indentation throughout
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

This file will contain Java-specific best practices including:
- **Oracle Java Code Conventions** - Official Java standards
- **Google Java Style Guide** - Google's conventions
- **Effective Java** - Joshua Bloch's best practices
- **Spring Framework** - Enterprise Java patterns
- **Java 17+ Features** - Modern Java features

---

## Quick Standards Summary

### Formatting
- **Indentation:** 4 spaces (no tabs)
- **Brace Style:** K&R (opening brace on same line)
- **Line Length:** 100-120 characters
- **One Statement Per Line:** No multiple statements on one line
- **Block Indentation:** +4 spaces

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

**For now, always use 4-space indentation per Oracle Java conventions. Full guidelines coming soon.**
