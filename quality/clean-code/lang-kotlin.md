# Kotlin Language-Specific Rules

> Based on Kotlin Coding Conventions (JetBrains official), Android Kotlin Style Guide, and SonarQube code quality rules

## Formatting Standards (Kotlin Specific)

**Indentation:** **4 spaces** (no tabs)

This follows the official Kotlin Coding Conventions from JetBrains and is the standard across the Kotlin ecosystem including Android development.

**Note:** Cortexa LLC uses **language-specific indentation standards**:
- **C++**: 2 spaces (Google C++ Style Guide)
- **C#**: 4 spaces (Microsoft standard)
- **Java**: 2 spaces (Cortexa LLC override)
- **JavaScript/TypeScript**: 2 spaces (ecosystem standard)
- **Kotlin**: 4 spaces (JetBrains standard)
- **Python**: 4 spaces (PEP 8 mandatory)

**Example:**
```kotlin
package com.cortexa.services

import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.map

/**
 * Service for managing user data.
 */
class UserService(
    private val repository: UserRepository,
    private val emailService: EmailService
) {
    /**
     * Retrieves a user by ID.
     *
     * @param id the user ID
     * @return the user if found, null otherwise
     */
    suspend fun getUserById(id: Long): User? {
        // 4-space indentation throughout
        require(id > 0) { "ID must be positive" }

        return repository.findById(id)?.also { user ->
            user.lastAccessed = Clock.System.now()
        }
    }

    /**
     * Creates a new user.
     *
     * @param userData the user data
     * @return the created user
     */
    suspend fun createUser(userData: UserData): User {
        val user = User(
            name = userData.name,
            email = userData.email,
            createdAt = Clock.System.now()
        )

        return repository.save(user).also { savedUser ->
            emailService.sendWelcomeEmail(savedUser)
        }
    }

    /**
     * Gets all active users as a Flow.
     */
    fun getActiveUsers(): Flow<List<User>> = repository.findAll()
        .map { users -> users.filter { it.isActive } }
}
```

---

## Overview

This file will contain Kotlin-specific best practices including:
- **Kotlin Coding Conventions** - Official JetBrains standards
- **Android Kotlin Style Guide** - Android-specific conventions
- **Effective Kotlin** - Best practices and idioms
- **Coroutines** - Asynchronous programming patterns
- **Flow** - Reactive streams in Kotlin

---

## Quick Standards Summary

### Formatting
- **Indentation:** 4 spaces (no tabs)
- **Brace Style:** K&R (opening brace on same line)
- **Line Length:** 100 characters (soft limit), 120 (hard limit)
- **Continuation Indent:** 4 spaces (not 8)
- **Trailing Commas:** Use them in multiline collections

### Naming
- `packagename` - all lowercase
- `ClassName` - UpperCamelCase (PascalCase)
- `functionName` - lowerCamelCase
- `propertyName` - lowerCamelCase
- `CONSTANT_NAME` - UPPER_SNAKE_CASE
- `_backing` - underscore prefix for backing properties

### Class Structure (Order)
```kotlin
class Example(
    private val dependency: Dependency
) {
    // 1. Companion object
    companion object {
        const val CONSTANT = "value"
        fun create(): Example = Example(createDependency())
    }

    // 2. Properties
    private val privateProperty: String = "value"
    var publicProperty: String = "value"

    // 3. Init blocks
    init {
        println("Initializing")
    }

    // 4. Secondary constructors
    constructor() : this(Dependency())

    // 5. Public functions
    fun publicFunction() {
        // ...
    }

    // 6. Internal functions
    internal fun internalFunction() {
        // ...
    }

    // 7. Protected functions
    protected fun protectedFunction() {
        // ...
    }

    // 8. Private functions
    private fun privateFunction() {
        // ...
    }

    // 9. Nested classes
    inner class Inner {
        // ...
    }

    // 10. Object declarations
    object Nested {
        // ...
    }
}
```

### Kotlin Idioms
```kotlin
// Data classes for DTOs
data class User(
    val id: Long,
    val name: String,
    val email: String
)

// Sealed classes for type-safe state
sealed class Result<out T> {
    data class Success<T>(val data: T) : Result<T>()
    data class Error(val exception: Throwable) : Result<Nothing>()
    object Loading : Result<Nothing>()
}

// Extension functions
fun String.isValidEmail(): Boolean =
    this.matches(Regex("^[A-Za-z0-9+_.-]+@[A-Za-z0-9.-]+$"))

// Scope functions
fun processUser(userId: Long): User? {
    return repository.findById(userId)?.apply {
        lastAccessed = Clock.System.now()
    }?.also {
        logAccess(it)
    }
}

// Null safety
fun safeLengthOf(str: String?): Int = str?.length ?: 0

// Elvis operator for null coalescing
val name = user?.name ?: "Unknown"

// Safe calls with let
user?.let { u ->
    sendEmail(u.email)
}

// Destructuring
val (id, name, email) = user

// Operator overloading
data class Vector(val x: Int, val y: Int) {
    operator fun plus(other: Vector) = Vector(x + other.x, y + other.y)
}

// Delegation
class Repository(database: Database) : Database by database

// Property delegation
class Config {
    val apiUrl: String by lazy {
        loadFromEnvironment("API_URL")
    }

    var cachedData: String? by observable(null) { prop, old, new ->
        println("$old -> $new")
    }
}
```

### Coroutines and Flow
```kotlin
// Suspend functions for async operations
suspend fun fetchUser(id: Long): User {
    return withContext(Dispatchers.IO) {
        api.getUser(id)
    }
}

// Flow for reactive streams
fun observeUsers(): Flow<List<User>> = flow {
    while (true) {
        emit(repository.findAll())
        delay(1000)
    }
}.flowOn(Dispatchers.IO)

// Structured concurrency
suspend fun loadUserProfile(userId: Long): UserProfile = coroutineScope {
    val user = async { fetchUser(userId) }
    val posts = async { fetchPosts(userId) }
    val friends = async { fetchFriends(userId) }

    UserProfile(
        user = user.await(),
        posts = posts.await(),
        friends = friends.await()
    )
}

// StateFlow for state management
class ViewModel {
    private val _state = MutableStateFlow<State>(State.Loading)
    val state: StateFlow<State> = _state.asStateFlow()

    fun loadData() {
        viewModelScope.launch {
            _state.value = State.Loading
            try {
                val data = repository.loadData()
                _state.value = State.Success(data)
            } catch (e: Exception) {
                _state.value = State.Error(e)
            }
        }
    }
}
```

### Modern Kotlin Features
```kotlin
// Inline classes for type safety
@JvmInline
value class UserId(val value: Long)

// Context receivers (Kotlin 1.6.20+)
context(Logger)
fun processData(data: List<String>) {
    log("Processing ${data.size} items")
    // ...
}

// Contracts for smart casts
fun String?.isNullOrEmpty(): Boolean {
    contract {
        returns(false) implies (this@isNullOrEmpty != null)
    }
    return this == null || this.isEmpty()
}

// Builders with type-safe DSL
fun html(init: Html.() -> Unit): Html = Html().apply(init)

class Html {
    private val children = mutableListOf<Element>()

    fun body(init: Body.() -> Unit) {
        children.add(Body().apply(init))
    }
}

// Usage
val page = html {
    body {
        h1 { +"Title" }
        p { +"Content" }
    }
}
```

### Android Specific
```kotlin
// Composable functions (Jetpack Compose)
@Composable
fun UserCard(user: User, modifier: Modifier = Modifier) {
    Card(modifier = modifier) {
        Column(
            modifier = Modifier.padding(16.dp)
        ) {
            Text(
                text = user.name,
                style = MaterialTheme.typography.h6
            )
            Text(
                text = user.email,
                style = MaterialTheme.typography.body2
            )
        }
    }
}

// ViewModel with lifecycle
class UserViewModel(
    private val repository: UserRepository
) : ViewModel() {
    private val _users = MutableStateFlow<List<User>>(emptyList())
    val users: StateFlow<List<User>> = _users.asStateFlow()

    init {
        viewModelScope.launch {
            repository.observeUsers()
                .collect { users ->
                    _users.value = users
                }
        }
    }
}
```

---

## Code Quality: SonarQube Enforcement

**MANDATORY:** All Kotlin code must pass SonarQube's default Kotlin rules for code smell detection.

### SonarQube Integration

**Official Reference:** https://rules.sonarsource.com/kotlin/

SonarQube provides automated static code analysis with **174 total rules** organized into:

| Category | Count | Focus |
|----------|-------|-------|
| **Code Smell** | 86 | Quality and maintainability issues |
| **Vulnerability** | 42 | Security exploits and weaknesses |
| **Bug** | 27 | Logic errors and runtime failures |
| **Security Hotspot** | 19 | Sensitive operations requiring review |

### Critical Security Vulnerabilities (Top Priority)

**All 42 vulnerability rules are MANDATORY:**

- **Mobile Database Encryption:** Encryption keys must not be disclosed in mobile applications
- **Hard-coded Credentials:** Never embed passwords, API keys, tokens, or secrets in code
- **SSL/TLS Verification:** Server hostnames must be verified during SSL/TLS connections
- **Weak Cryptography:** Use strong cipher algorithms (AES-256, RSA-2048+)
- **Android Permissions:** Request minimal necessary permissions, document security-sensitive ones
- **Biometric Authentication:** Implement proper biometric authentication with crypto objects
- **WebView Security:** Disable JavaScript and file access unless required, validate content
- **SQL Injection:** Use parameterized queries, never string concatenation
- **Path Traversal:** Validate file paths to prevent directory traversal
- **Insecure Random:** Use SecureRandom for cryptographic purposes

### Critical Bug Detection (27 Rules)

**All bug rules are MANDATORY:**

- **Return Value Validation:** Check return values from file operations and synchronization primitives
- **Array Equality in Data Classes:** Override equals in data classes containing array fields
- **Regex Validity:** Ensure regular expressions are syntactically correct
- **Null Safety:** Use safe calls (?.) and null checks properly
- **Type Casting:** Verify types before casting using smart casts or safe casts (as?)
- **Coroutine Context:** Use appropriate dispatcher (IO, Main, Default) for operations
- **StateFlow Updates:** Update StateFlow.value atomically to prevent race conditions
- **Resource Management:** Close resources properly (use 'use' function for AutoCloseable)
- **Infinite Loops:** Ensure loop termination conditions are reachable
- **Concurrent Modification:** Don't modify collections while iterating

### Code Quality Standards (86 Code Smell Rules)

**Key code smell rules for maintainability:**

#### Complexity Management
- **Cognitive Complexity:** Functions should maintain reasonable complexity thresholds (max 15)
- **Function Length:** Decompose overly long functions for readability (max 50 lines)
- **Parameter Count:** Limit function parameters (max 5, use data classes for more)

#### Kotlin Idioms
- **Immutability:** Prefer `val` over `var` for immutable properties
- **Data Classes:** Use data classes for DTOs and simple value objects
- **Sealed Classes:** Use sealed classes for restricted type hierarchies
- **Extension Functions:** Prefer extension functions over utility classes
- **Scope Functions:** Use appropriate scope functions (let, run, with, apply, also)

#### Coroutines Best Practices
- **Structured Concurrency:** Use coroutineScope for structured concurrency
- **Exception Handling:** Handle exceptions in coroutines appropriately
- **Flow Collection:** Collect flows in appropriate lifecycle scopes
- **Cancellation:** Support cancellation in long-running operations

#### Android-Specific
- **Lifecycle Awareness:** Use lifecycle-aware components (ViewModel, LiveData, Flow)
- **Memory Leaks:** Avoid memory leaks (clear listeners, cancel coroutines)
- **Main Thread:** Don't perform heavy operations on the main thread
- **Context References:** Avoid holding references to Activity/Fragment contexts

### Configuration

Add SonarQube to your Kotlin project:

**Gradle (Kotlin DSL):**
```kotlin
plugins {
    id("org.sonarqube") version "4.0.0.2929"
}

sonar {
    properties {
        property("sonar.projectKey", "your-project-key")
        property("sonar.organization", "cortexa")
        property("sonar.host.url", "https://sonarqube.cortexa.com")
        property("sonar.kotlin.source.version", "1.9")
    }
}
```

**Gradle (Groovy):**
```groovy
plugins {
    id "org.sonarqube" version "4.0.0.2929"
}

sonar {
    properties {
        property "sonar.projectKey", "your-project-key"
        property "sonar.organization", "cortexa"
        property "sonar.host.url", "https://sonarqube.cortexa.com"
        property "sonar.kotlin.source.version", "1.9"
    }
}
```

### Running SonarQube Analysis

```bash
# Gradle
./gradlew sonarqube

# With authentication token
./gradlew sonarqube -Dsonar.login=your-token

# Android project
./gradlew sonarqube -Dsonar.androidLint.reportPaths=build/reports/lint-results.xml
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

**Android Studio / IntelliJ IDEA:**
```
Settings → Plugins → Install "SonarLint"
Settings → Tools → SonarLint → Bind to SonarQube/SonarCloud
Configure connection and bind to your project
```

**SonarLint Real-time Analysis:**
- Provides immediate feedback while coding
- Highlights issues with quick fixes
- Syncs with SonarQube server rules
- Shows issue descriptions and remediation guidance

---

## TODO: Full Kotlin Guidelines

This file will be expanded to include:
- [ ] Complete Kotlin Coding Conventions
- [ ] Effective Kotlin best practices
- [ ] Coroutines and Flow patterns
- [ ] Android development best practices
- [ ] Jetpack Compose guidelines
- [ ] Testing with JUnit and Kotest
- [ ] Error handling patterns
- [ ] Multiplatform considerations
- [ ] Common anti-patterns
- [ ] Performance optimization

---

**For now, always use 4-space indentation per Kotlin Coding Conventions. All code must pass SonarQube's default Kotlin rules. Full guidelines coming soon.**
