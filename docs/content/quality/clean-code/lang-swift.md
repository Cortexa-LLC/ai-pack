---
sidebar_position: 27
title: "Swift & SwiftUI Language-Specific Standards"
---

# Swift & SwiftUI Language-Specific Standards

> Based on Swift API Design Guidelines, SwiftUI Best Practices, and Apple Human Interface Guidelines

## Overview

This document provides comprehensive Swift and SwiftUI coding standards including:
- **Swift API Design Guidelines** - Apple's official conventions
- **SwiftUI Best Practices** - Modern declarative UI patterns
- **Effective Swift** - Idiomatic Swift patterns
- **Concurrency** - async/await and structured concurrency
- **Combine** - Reactive programming (when needed)

---

## Formatting Standards

### Indentation
**Standard:** **4 spaces** (no tabs)

Follows Apple's Swift conventions and Xcode defaults.

### Brace Style
**K&R style** - Opening brace on same line:
```swift
func calculateTotal(items: [Item]) -> Decimal {
    return items.reduce(0) { $0 + $1.price }
}

if condition {
    doSomething()
} else {
    doSomethingElse()
}
```text

### Line Length
- **Soft limit:** 100 characters
- **Hard limit:** 120 characters
- Break long lines at natural boundaries

### Trailing Commas
Use trailing commas in multiline arrays, dictionaries, and parameter lists:
```swift
let colors = [
    .red,
    .blue,
    .green,  // ✅ Trailing comma
]

func configure(
    name: String,
    age: Int,
    address: String  // ✅ Trailing comma
) {
    // ...
}
```text

---

## Naming Conventions

### Types and Protocols
**UpperCamelCase (PascalCase)**
```swift
class UserViewModel { }
struct UserProfile { }
enum LoadingState { }
protocol Identifiable { }
```text

### Functions and Properties
**lowerCamelCase**
```swift
func fetchUserData() { }
var userName: String
let itemCount: Int
```text

### Constants
**lowerCamelCase** (not SCREAMING_SNAKE_CASE)
```swift
// ✅ Swift style
let maximumLoginAttempts = 3
let defaultTimeout: TimeInterval = 30

// ❌ Don't use
let MAXIMUM_LOGIN_ATTEMPTS = 3
```text

### Boolean Properties and Functions
Use clear, affirmative names:
```swift
// ✅ Good
var isLoading: Bool
var hasUnsavedChanges: Bool
func canSubmit() -> Bool

// ❌ Avoid negatives
var isNotReady: Bool  // Prefer: isReady
```text

### Protocol Naming
- **Capability:** `-able`, `-ible` suffix
- **Description:** noun
```swift
protocol Drawable { }          // Capability
protocol DataSource { }        // Description
protocol ViewModelProtocol { } // ❌ Avoid redundant "Protocol"
```text

---

## Swift Best Practices

### 1. Type Inference
Let the compiler infer types when obvious:
```swift
// ✅ Good
let name = "John"
let count = items.count
let colors = [.red, .blue, .green]

// ❌ Redundant
let name: String = "John"
let count: Int = items.count
```text

### 2. Optionals
**Use guard for early returns:**
```swift
// ✅ Good
guard let user = currentUser else {
    return
}
// Use user safely here

// ❌ Avoid deep nesting
if let user = currentUser {
    // Deep nesting...
}
```text

**Use nil coalescing and optional chaining:**
```swift
// ✅ Good
let userName = user?.name ?? "Guest"
let count = users?.count ?? 0

// ❌ Verbose
let userName: String
if let user = user, let name = user.name {
    userName = name
} else {
    userName = "Guest"
}
```text

### 3. Collections
**Use higher-order functions:**
```swift
// ✅ Good
let adults = users.filter { $0.age >= 18 }
let names = users.map { $0.name }
let total = prices.reduce(0, +)

// ❌ Imperative style
var adults: [User] = []
for user in users {
    if user.age >= 18 {
        adults.append(user)
    }
}
```text

### 4. Error Handling
**Use proper error handling:**
```swift
// ✅ Good
enum NetworkError: Error {
    case invalidURL
    case requestFailed(Error)
    case invalidResponse
}

func fetchData() async throws -> Data {
    guard let url = URL(string: endpoint) else {
        throw NetworkError.invalidURL
    }

    let (data, response) = try await URLSession.shared.data(from: url)

    guard let httpResponse = response as? HTTPURLResponse,
          (200...299).contains(httpResponse.statusCode) else {
        throw NetworkError.invalidResponse
    }

    return data
}

// Usage
do {
    let data = try await fetchData()
    // Process data
} catch NetworkError.invalidURL {
    // Handle specific error
} catch {
    // Handle general error
}
```text

### 5. Computed Properties vs Functions
**Use computed properties for lightweight, side-effect-free values:**
```swift
// ✅ Computed property (no side effects, cheap)
var fullName: String {
    "\(firstName) \(lastName)"
}

// ✅ Function (has side effects or expensive)
func fetchUserData() async throws -> User {
    // Network call
}
```text

### 6. Access Control
**Use explicit access control:**
```swift
// ✅ Good
public class ViewModel { }
private let cache = Cache()
fileprivate func helper() { }

// Default is internal, but be explicit when it matters
```text

### 7. Extensions for Organization
**Group related functionality:**
```swift
// MARK: - Initialization
extension UserViewModel {
    convenience init(userId: String) {
        self.init()
        self.userId = userId
    }
}

// MARK: - Public Methods
extension UserViewModel {
    func loadUser() async {
        // ...
    }
}

// MARK: - Private Helpers
private extension UserViewModel {
    func validateEmail(_ email: String) -> Bool {
        // ...
    }
}
```text

---

## SwiftUI Best Practices

### 1. View Composition
**Break complex views into smaller components:**
```swift
// ✅ Good - Composed
struct UserProfileView: View {
    let user: User

    var body: some View {
        VStack(spacing: 16) {
            ProfileHeaderView(user: user)
            ProfileDetailsView(user: user)
            ProfileActionsView(user: user)
        }
    }
}

// ❌ Bad - Monolithic
struct UserProfileView: View {
    let user: User

    var body: some View {
        VStack {
            // 200 lines of UI code...
        }
    }
}
```text

### 2. View Modifiers
**Extract custom modifiers for reusability:**
```swift
// Define custom modifier
struct CardStyle: ViewModifier {
    func body(content: Content) -> some View {
        content
            .padding()
            .background(Color.white)
            .cornerRadius(12)
            .shadow(radius: 4)
    }
}

extension View {
    func cardStyle() -> some View {
        modifier(CardStyle())
    }
}

// Usage
Text("Hello")
    .cardStyle()
```text

### 3. State Management
**Follow single source of truth:**
```swift
// ✅ Good
@StateObject private var viewModel = UserViewModel()

struct ContentView: View {
    @StateObject private var viewModel = UserViewModel()

    var body: some View {
        UserListView(users: viewModel.users)
            .task {
                await viewModel.loadUsers()
            }
    }
}

// Child view
struct UserListView: View {
    let users: [User]  // ✅ Pass data down

    var body: some View {
        List(users) { user in
            Text(user.name)
        }
    }
}
```text

**Use appropriate property wrappers:**
```swift
@State         // View-local state
@StateObject   // Own the object lifetime
@ObservedObject // Observe externally-owned object
@EnvironmentObject // Shared across view hierarchy
@Binding       // Two-way binding to parent state
@AppStorage    // UserDefaults-backed storage
@Published     // Observable property in ObservableObject
```text

### 4. Async Operations
**Use .task modifier for async work:**
```swift
// ✅ Good
struct ContentView: View {
    @StateObject private var viewModel = ViewModel()

    var body: some View {
        List(viewModel.items) { item in
            Text(item.name)
        }
        .task {
            await viewModel.loadItems()
        }
        .refreshable {
            await viewModel.refresh()
        }
    }
}

// ❌ Avoid .onAppear with Task
.onAppear {
    Task {  // Manual task management
        await viewModel.loadItems()
    }
}
```text

### 5. Preview Provider
**Provide comprehensive previews:**
```swift
struct UserView_Previews: PreviewProvider {
    static var previews: some View {
        Group {
            // Default state
            UserView(user: .sample)
                .previewDisplayName("Default")

            // Empty state
            UserView(user: .empty)
                .previewDisplayName("Empty")

            // Dark mode
            UserView(user: .sample)
                .preferredColorScheme(.dark)
                .previewDisplayName("Dark Mode")

            // Different sizes
            UserView(user: .sample)
                .previewLayout(.sizeThatFits)
                .previewDisplayName("Size That Fits")
        }
    }
}

// Sample data extension
extension User {
    static let sample = User(
        name: "John Doe",
        email: "john@example.com"
    )

    static let empty = User(name: "", email: "")
}
```text

### 6. Accessibility
**Always include accessibility support:**
```swift
struct UserRowView: View {
    let user: User

    var body: some View {
        HStack {
            Image(systemName: "person.circle")
                .accessibilityHidden(true)  // Decorative

            VStack(alignment: .leading) {
                Text(user.name)
                    .font(.headline)
                Text(user.email)
                    .font(.caption)
                    .foregroundColor(.secondary)
            }
        }
        .accessibilityElement(children: .combine)
        .accessibilityLabel("\(user.name), \(user.email)")
    }
}
```text

### 7. Performance
**Use @ViewBuilder for conditional views:**
```swift
// ✅ Good
@ViewBuilder
var statusView: some View {
    switch status {
    case .loading:
        ProgressView()
    case .loaded(let data):
        DataView(data: data)
    case .error(let message):
        ErrorView(message: message)
    }
}

// Use .id() to force view recreation when needed
List(items, id: \.id) { item in
    ItemRow(item: item)
}
.id(refreshID)  // Change ID to recreate view
```text

---

## Modern Concurrency (async/await)

### Async Functions
```swift
// ✅ Good
func fetchUser(id: String) async throws -> User {
    let url = URL(string: "https://api.example.com/users/\(id)")!
    let (data, _) = try await URLSession.shared.data(from: url)
    return try JSONDecoder().decode(User.self, from: data)
}

// Multiple concurrent operations
func loadData() async {
    async let users = fetchUsers()
    async let posts = fetchPosts()
    async let comments = fetchComments()

    let (userList, postList, commentList) = await (users, posts, comments)
    // Process all data
}
```text

### Task Groups
```swift
func fetchAllUserData(userIDs: [String]) async throws -> [User] {
    try await withThrowingTaskGroup(of: User.self) { group in
        for id in userIDs {
            group.addTask {
                try await fetchUser(id: id)
            }
        }

        var users: [User] = []
        for try await user in group {
            users.append(user)
        }
        return users
    }
}
```text

### Actors
```swift
// ✅ Use actors for thread-safe mutable state
actor UserCache {
    private var cache: [String: User] = [:]

    func get(_ id: String) -> User? {
        cache[id]
    }

    func set(_ user: User, for id: String) {
        cache[id] = user
    }
}

// Usage
let cache = UserCache()
await cache.set(user, for: user.id)
let cachedUser = await cache.get(user.id)
```text

---

## Testing Best Practices

### Unit Tests
```swift
import XCTest
@testable import MyApp

final class UserViewModelTests: XCTestCase {
    var sut: UserViewModel!
    var mockRepository: MockUserRepository!

    override func setUp() {
        super.setUp()
        mockRepository = MockUserRepository()
        sut = UserViewModel(repository: mockRepository)
    }

    override func tearDown() {
        sut = nil
        mockRepository = nil
        super.tearDown()
    }

    func testLoadUsers_whenSuccessful_updatesUsers() async {
        // Given
        let expectedUsers = [User.sample]
        mockRepository.users = expectedUsers

        // When
        await sut.loadUsers()

        // Then
        XCTAssertEqual(sut.users, expectedUsers)
        XCTAssertFalse(sut.isLoading)
    }

    func testLoadUsers_whenFails_setsError() async {
        // Given
        mockRepository.shouldFail = true

        // When
        await sut.loadUsers()

        // Then
        XCTAssertNotNil(sut.error)
        XCTAssertTrue(sut.users.isEmpty)
    }
}
```text

### UI Tests
```swift
import XCTest

final class UserFlowUITests: XCTestCase {
    let app = XCUIApplication()

    override func setUpWithError() throws {
        continueAfterFailure = false
        app.launch()
    }

    func testUserLoginFlow() throws {
        // Given
        let emailField = app.textFields["email"]
        let passwordField = app.secureTextFields["password"]
        let loginButton = app.buttons["login"]

        // When
        emailField.tap()
        emailField.typeText("test@example.com")

        passwordField.tap()
        passwordField.typeText("password123")

        loginButton.tap()

        // Then
        let welcomeText = app.staticTexts["Welcome"]
        XCTAssertTrue(welcomeText.waitForExistence(timeout: 5))
    }
}
```text

---

## Common Anti-Patterns to Avoid

### ❌ Force Unwrapping
```swift
// ❌ Bad
let user = users.first!  // Crashes if empty

// ✅ Good
guard let user = users.first else { return }
```text

### ❌ Pyramid of Doom
```swift
// ❌ Bad
if let a = optionalA {
    if let b = optionalB {
        if let c = optionalC {
            // Deep nesting
        }
    }
}

// ✅ Good
guard let a = optionalA,
      let b = optionalB,
      let c = optionalC else {
    return
}
// Flat code
```text

### ❌ Massive View Controllers/ViewModels
```swift
// ❌ Bad - 1000+ lines
class UserViewModel: ObservableObject {
    // Everything user-related
}

// ✅ Good - Split responsibilities
class UserProfileViewModel: ObservableObject { }
class UserSettingsViewModel: ObservableObject { }
class UserPostsViewModel: ObservableObject { }
```text

### ❌ Implicit Self in Closures
```swift
// ❌ Bad - potential retain cycles
class ViewModel: ObservableObject {
    func loadData() {
        service.fetch { data in
            self.data = data  // Strong reference
        }
    }
}

// ✅ Good - explicit capture
class ViewModel: ObservableObject {
    func loadData() {
        service.fetch { [weak self] data in
            self?.data = data
        }
    }
}
```text

---

## Code Review Checklist

### General
- [ ] Follows Swift API Design Guidelines
- [ ] Proper access control (private, fileprivate, internal, public)
- [ ] No force unwrapping (!)
- [ ] Proper error handling
- [ ] No retain cycles in closures

### SwiftUI Specific
- [ ] Views are composable and reusable
- [ ] Proper property wrappers (@State, @Binding, etc.)
- [ ] Accessibility labels and hints provided
- [ ] Preview providers included
- [ ] .task used for async operations

### Testing
- [ ] Unit tests for business logic
- [ ] UI tests for critical flows
- [ ] Tests follow AAA pattern (Arrange, Act, Assert)
- [ ] Mock/stub dependencies properly

### Performance
- [ ] No unnecessary view redraws
- [ ] Lazy loading for expensive operations
- [ ] Proper use of .id() for list performance
- [ ] Async operations properly structured

---

## Tools and Linters

### SwiftLint
Use SwiftLint for automated style checking:
```yaml
# .swiftlint.yml
disabled_rules:
  - trailing_whitespace
opt_in_rules:
  - empty_count
  - explicit_init
line_length: 120
```text

### SwiftFormat
Use SwiftFormat for consistent formatting:
```text
--indent 4
--maxwidth 120
--wraparguments before-first
--wrapcollections before-first
```text

---

## Resources

- [Swift API Design Guidelines](https://swift.org/documentation/api-design-guidelines/)
- [SwiftUI Documentation](https://developer.apple.com/documentation/swiftui/)
- [Swift.org Documentation](https://swift.org/documentation/)
- [Human Interface Guidelines](https://developer.apple.com/design/human-interface-guidelines/)
- [Effective Swift](https://www.swift.org/documentation/)

---

**Follow these standards for clean, maintainable, and idiomatic Swift/SwiftUI code.**
