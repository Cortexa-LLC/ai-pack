# iOS Developer
<!-- skills/ios_developer.skill.md -->

**Version:** 1.0
**InjectAt:** role_context
**Slot:** 50
**Tools:** Bash, Read, Write, Edit, Grep, Glob
**Gates:** ios-hig-compliance, accessibility-wcag-aa
**MaxExtraTokens:** 5000
**Optional:** true

---

## iOS Platform Development

Platform-specific capabilities for iOS development with Swift, SwiftUI, UIKit, and Xcode tooling.

**Use when:** Building iOS apps, working with Swift code, or iOS-specific patterns.

---

## Platform Tooling

### Build and Test
```bash
# Build iOS target
xcodebuild -workspace App.xcworkspace \
  -scheme AppScheme \
  -destination 'platform=iOS Simulator,name=iPhone 15' \
  build

# Run tests
xcodebuild test \
  -workspace App.xcworkspace \
  -scheme AppScheme \
  -destination 'platform=iOS Simulator,name=iPhone 15'

# Run specific test
xcodebuild test \
  -workspace App.xcworkspace \
  -scheme AppScheme \
  -only-testing:AppTests/FeatureTests/testSpecificCase
```

### Code Quality
```bash
# Run SwiftLint
swiftlint lint --config .swiftlint.yml

# Auto-fix SwiftLint issues
swiftlint lint --fix --config .swiftlint.yml

# Run SwiftFormat
swiftformat . --config .swiftformat
```

### UI Testing
```bash
# Launch simulator
xcrun simctl boot "iPhone 15"

# Run UI tests
xcodebuild test \
  -workspace App.xcworkspace \
  -scheme AppScheme \
  -destination 'platform=iOS Simulator,name=iPhone 15' \
  -only-testing:AppUITests
```

---

## iOS Patterns and Best Practices

### SwiftUI Patterns

**State Management:**
```swift
// ✅ Use @State for view-local state
@State private var isExpanded = false

// ✅ Use @StateObject for owned ObservableObjects
@StateObject private var viewModel = FeatureViewModel()

// ✅ Use @ObservedObject for passed-in ObservableObjects
@ObservedObject var sharedModel: SharedModel

// ✅ Use @EnvironmentObject for injected dependencies
@EnvironmentObject var appState: AppState

// ❌ Avoid: Creating ObservableObject in body
var body: some View {
    let viewModel = ViewModel() // WRONG: recreated on every render
}
```

**View Composition:**
```swift
// ✅ Extract reusable views
struct FeatureView: View {
    var body: some View {
        VStack {
            HeaderView()
            ContentView()
            FooterView()
        }
    }
}

// ✅ Use ViewBuilder for conditional views
@ViewBuilder
func content() -> some View {
    if condition {
        ViewA()
    } else {
        ViewB()
    }
}

// ❌ Avoid: Complex logic in body
// Extract to computed properties or functions
```

### Memory Management

**Avoid Retain Cycles:**
```swift
// ✅ Use [weak self] in escaping closures
viewModel.fetchData { [weak self] result in
    guard let self else { return }
    self.handleResult(result)
}

// ✅ Use [unowned self] when self lifetime guaranteed
Timer.scheduledTimer(withTimeInterval: 1.0, repeats: true) { [unowned self] _ in
    self.update()
}

// ❌ Avoid: Strong capture in closures
someLongRunningTask {
    self.property = value  // WRONG: retain cycle
}
```

**Lifecycle Management:**
```swift
// ✅ Clean up in deinit
deinit {
    NotificationCenter.default.removeObserver(self)
    timer?.invalidate()
    cancellables.forEach { $0.cancel() }
}

// ✅ Use onDisappear for SwiftUI cleanup
.onDisappear {
    viewModel.cleanup()
}
```

### Concurrency

**Modern Async/Await:**
```swift
// ✅ Use async/await for async work
func fetchData() async throws -> Data {
    let (data, _) = try await URLSession.shared.data(from: url)
    return data
}

// ✅ Use Task for calling async from sync
func loadData() {
    Task {
        do {
            let data = try await fetchData()
            await MainActor.run {
                self.updateUI(with: data)
            }
        } catch {
            handleError(error)
        }
    }
}

// ✅ Use @MainActor for UI updates
@MainActor
func updateUI(with data: Data) {
    self.items = parse(data)
}

// ❌ Avoid: Completion handler pyramids
// Use async/await instead
```

### Accessibility

**VoiceOver Support:**
```swift
// ✅ Provide accessibility labels
Image("icon")
    .accessibilityLabel("Settings")

// ✅ Group related elements
VStack {
    Text("Title")
    Text("Subtitle")
}
.accessibilityElement(children: .combine)

// ✅ Add accessibility hints
Button("Submit") { }
    .accessibilityHint("Submits the form")

// ✅ Set accessibility traits
Text("Error message")
    .accessibilityAddTraits(.isStaticText)
```

**Dynamic Type:**
```swift
// ✅ Use scalable fonts
Text("Content")
    .font(.body)

// ✅ Support dynamic type scaling
Text("Headline")
    .font(.custom("CustomFont", size: 17, relativeTo: .body))

// ❌ Avoid: Fixed font sizes
Text("Content")
    .font(.system(size: 14))  // WRONG: doesn't scale
```

---

## Human Interface Guidelines (HIG) Compliance

### Navigation Patterns
- Use `NavigationStack` for hierarchical navigation (iOS 16+)
- Use `TabView` for peer-level navigation (max 5 tabs)
- Use modal presentations sparingly, prefer push navigation

### Visual Design
- Respect system colors and appearances (light/dark mode)
- Use SF Symbols for icons
- Follow standard spacing and padding conventions
- Support all device orientations when appropriate

### User Input
- Validate input in real-time
- Show clear error messages near the input
- Use native keyboard types (`.numberPad`, `.emailAddress`)
- Dismiss keyboard on scroll or tap outside

---

## Common iOS Anti-Patterns to Avoid

| Anti-Pattern | Risk | Fix |
|--------------|------|-----|
| **Creating ObservableObject in body** | Recreated every render, lost state | Use `@StateObject` or pass as parameter |
| **Strong self in closure** | Retain cycles, memory leaks | Use `[weak self]` or `[unowned self]` |
| **Force unwrapping (!!)** | Runtime crashes | Use optional binding or nil coalescing |
| **DispatchQueue.main.async in async context** | Unnecessary nesting | Use `await MainActor.run {}` |
| **Massive View Controller** | Unmaintainable code | Extract view models, use composition |
| **UIKit in SwiftUI** | Performance issues | Use native SwiftUI when possible |
| **No accessibility labels** | VoiceOver unusable | Add `.accessibilityLabel()` |

---

## Testing Guidelines

### Unit Tests
```swift
// ✅ Test business logic in ViewModels
func testFetchData() async throws {
    let viewModel = FeatureViewModel(service: MockService())
    try await viewModel.fetchData()
    XCTAssertEqual(viewModel.items.count, 3)
}

// ✅ Test error handling
func testErrorHandling() async {
    let viewModel = FeatureViewModel(service: FailingService())
    await viewModel.fetchData()
    XCTAssertTrue(viewModel.hasError)
}
```

### UI Tests
```swift
// ✅ Test critical user flows
func testLoginFlow() throws {
    let app = XCUIApplication()
    app.launch()

    app.textFields["Email"].tap()
    app.textFields["Email"].typeText("user@example.com")

    app.secureTextFields["Password"].tap()
    app.secureTextFields["Password"].typeText("password")

    app.buttons["Login"].tap()

    XCTAssertTrue(app.staticTexts["Welcome"].exists)
}
```

---

## Build Configuration

### Xcode Build Settings
```bash
# Get build settings
xcodebuild -workspace App.xcworkspace \
  -scheme AppScheme \
  -showBuildSettings

# Build for release
xcodebuild -workspace App.xcworkspace \
  -scheme AppScheme \
  -configuration Release \
  -destination 'generic/platform=iOS' \
  archive -archivePath ./build/App.xcarchive
```

### Swift Compiler Flags
- Enable strict concurrency checking: `-strict-concurrency=complete`
- Enable upcoming features: `-enable-upcoming-feature ConciseMagicFile`
- Optimize for size: `-O` with `-Osize`

---

## Integration with Other Skills

- **Uses:** general (core capabilities)
- **Complements:** code_review (iOS-specific patterns)
- **Used by roles:** engineer, designer
- **Gates enforced:** ios-hig-compliance, accessibility-wcag-aa

---

## Example Usage

```
User: "Build and test the iOS app"

Agent (with ios_developer skill):
1. Runs SwiftLint for code quality
2. Builds iOS scheme for simulator
3. Runs unit tests
4. Runs UI tests
5. Reports results with any HIG violations
```

---

## Notes

- **Prefer SwiftUI** for new UI code (iOS 15+)
- **Use async/await** over completion handlers (iOS 15+)
- **Support iOS 15+** unless specified otherwise
- **Test on real devices** for final validation
- **Profile with Instruments** for performance issues
