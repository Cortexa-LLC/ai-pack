# Android Developer
<!-- skills/android_developer.skill.md -->

**Version:** 1.0
**InjectAt:** role_context
**Slot:** 50
**Tools:** Bash, Read, Write, Edit, Grep, Glob
**Gates:** material-design-compliance, accessibility-wcag-aa
**MaxExtraTokens:** 5000
**Optional:** true

---

## Android Platform Development

Platform-specific capabilities for Android development with Kotlin, Jetpack Compose, and Gradle tooling.

**Use when:** Building Android apps, working with Kotlin code, or Android-specific patterns.

---

## Platform Tooling

### Build and Test
```bash
# Build debug APK
./gradlew assembleDebug

# Build release APK
./gradlew assembleRelease

# Run unit tests
./gradlew test

# Run instrumented tests (requires emulator/device)
./gradlew connectedAndroidTest

# Run specific test
./gradlew test --tests com.example.FeatureTest.testSpecificCase
```

### Code Quality
```bash
# Run Android Lint
./gradlew lint

# Run detekt (Kotlin static analysis)
./gradlew detekt

# Run ktlint (Kotlin formatting)
./gradlew ktlintCheck

# Auto-fix ktlint issues
./gradlew ktlintFormat
```

### Emulator Management
```bash
# List available AVDs
emulator -list-avds

# Start emulator
emulator -avd Pixel_6_API_34 &

# List running devices
adb devices

# Install APK
adb install -r app/build/outputs/apk/debug/app-debug.apk
```

---

## Android Patterns and Best Practices

### Jetpack Compose Patterns

**State Management:**
```kotlin
// ✅ Use remember for view-local state
@Composable
fun Counter() {
    var count by remember { mutableStateOf(0) }
    Button(onClick = { count++ }) {
        Text("Count: $count")
    }
}

// ✅ Use rememberSaveable for state that survives config changes
var text by rememberSaveable { mutableStateOf("") }

// ✅ Use ViewModel for business logic
@Composable
fun FeatureScreen(viewModel: FeatureViewModel = viewModel()) {
    val state by viewModel.uiState.collectAsState()
    FeatureContent(state)
}

// ❌ Avoid: Creating ViewModel in composable
val viewModel = FeatureViewModel()  // WRONG: recreated on recomposition
```

**Side Effects:**
```kotlin
// ✅ Use LaunchedEffect for coroutines
LaunchedEffect(userId) {
    viewModel.loadUser(userId)
}

// ✅ Use DisposableEffect for cleanup
DisposableEffect(Unit) {
    val listener = createListener()
    onDispose {
        listener.dispose()
    }
}

// ✅ Use rememberCoroutineScope for user actions
val scope = rememberCoroutineScope()
Button(onClick = { scope.launch { doSomething() } }) {
    Text("Action")
}
```

### Lifecycle Management

**Activity/Fragment Lifecycle:**
```kotlin
// ✅ Use lifecycle-aware components
class FeatureViewModel : ViewModel() {
    override fun onCleared() {
        // Clean up resources
        subscription.cancel()
    }
}

// ✅ Observe lifecycle in composable
val lifecycleOwner = LocalLifecycleOwner.current
DisposableEffect(lifecycleOwner) {
    val observer = LifecycleEventObserver { _, event ->
        if (event == Lifecycle.Event.ON_PAUSE) {
            viewModel.pause()
        }
    }
    lifecycleOwner.lifecycle.addObserver(observer)
    onDispose {
        lifecycleOwner.lifecycle.removeObserver(observer)
    }
}

// ❌ Avoid: Holding Activity/Fragment references in background tasks
GlobalScope.launch {
    activity.updateUI()  // WRONG: context leak
}
```

### Memory Management

**Avoid Context Leaks:**
```kotlin
// ✅ Use Application context for long-lived operations
class Repository(private val appContext: Context) {
    fun loadData() {
        // Safe: Application context doesn't leak
    }
}

// ❌ Avoid: Activity context in singleton
object DataManager {
    lateinit var context: Context  // WRONG: leaks Activity
}

// ✅ Use WeakReference if you must hold Activity reference
class CustomView {
    private var activityRef: WeakReference<Activity>? = null
}
```

**Coroutine Scopes:**
```kotlin
// ✅ Use viewModelScope for ViewModel coroutines
viewModelScope.launch {
    repository.fetchData()
}

// ✅ Use lifecycleScope for Activity/Fragment
lifecycleScope.launch {
    repeatOnLifecycle(Lifecycle.State.STARTED) {
        viewModel.uiState.collect { /* update UI */ }
    }
}

// ❌ Avoid: GlobalScope (never gets cancelled)
GlobalScope.launch {
    doWork()  // WRONG: runs forever
}
```

### Navigation

**Compose Navigation:**
```kotlin
// ✅ Define navigation routes
sealed class Screen(val route: String) {
    object Home : Screen("home")
    data class Detail(val id: String) : Screen("detail/{id}")
}

// ✅ Set up NavHost
NavHost(navController, startDestination = Screen.Home.route) {
    composable(Screen.Home.route) { HomeScreen(navController) }
    composable(Screen.Detail.route) { backStackEntry ->
        val id = backStackEntry.arguments?.getString("id")
        DetailScreen(id)
    }
}

// ✅ Navigate with type safety
navController.navigate(Screen.Detail("123"))
```

### Accessibility

**Content Descriptions:**
```kotlin
// ✅ Provide content descriptions
Icon(
    imageVector = Icons.Default.Settings,
    contentDescription = "Settings"
)

// ✅ Group related elements
Row(modifier = Modifier.semantics(mergeDescendants = true) {}) {
    Text("Title")
    Text("Subtitle")
}

// ✅ Add custom accessibility actions
Button(
    onClick = { },
    modifier = Modifier.semantics {
        contentDescription = "Submit form"
        onClick(label = "Submit") {
            submitForm()
            true
        }
    }
) { Text("Submit") }
```

**TalkBack Support:**
```kotlin
// ✅ Set traversal order
Column(
    modifier = Modifier.semantics {
        traversalIndex = 1f
    }
) { }

// ✅ Mark decorative elements
Image(
    painter = painterResource(R.drawable.decoration),
    contentDescription = null  // Decorative, skip for TalkBack
)
```

---

## Material Design Compliance

### Components
- Use Material3 components from `androidx.compose.material3`
- Follow Material Design color system (primary, secondary, tertiary)
- Use Material theming for consistent appearance
- Implement adaptive layouts for tablets and foldables

### Typography
```kotlin
// ✅ Use Material Typography
Text(
    text = "Headline",
    style = MaterialTheme.typography.headlineMedium
)

// ❌ Avoid: Custom text styles without theming
Text(
    text = "Headline",
    fontSize = 24.sp,
    fontWeight = FontWeight.Bold
)
```

### Colors
```kotlin
// ✅ Use theme colors
Card(colors = CardDefaults.cardColors(
    containerColor = MaterialTheme.colorScheme.surface
)) { }

// ❌ Avoid: Hardcoded colors
Card(colors = CardDefaults.cardColors(
    containerColor = Color(0xFF6200EE)  // WRONG: breaks dark mode
)) { }
```

---

## Common Android Anti-Patterns to Avoid

| Anti-Pattern | Risk | Fix |
|--------------|------|-----|
| **Context leak in singleton** | Memory leak, crash | Use Application context or weak reference |
| **GlobalScope.launch** | Never cancelled, memory leak | Use `viewModelScope` or `lifecycleScope` |
| **Blocking main thread** | ANR (App Not Responding) | Use coroutines or WorkManager |
| **findViewById in loop** | Performance degradation | Use ViewBinding or Compose |
| **Hardcoded strings** | Can't localize | Use string resources |
| **No null checks** | NullPointerException | Use Kotlin null safety (`?.`, `?:`) |
| **Creating new instance in recomposition** | Performance issues | Use `remember` or `rememberSaveable` |
| **Missing content descriptions** | TalkBack unusable | Add `contentDescription` to all UI elements |

---

## Testing Guidelines

### Unit Tests
```kotlin
// ✅ Test ViewModel logic
@Test
fun `fetch data updates state`() = runTest {
    val viewModel = FeatureViewModel(FakeRepository())
    viewModel.fetchData()

    val state = viewModel.uiState.value
    assertEquals(3, state.items.size)
}

// ✅ Test error handling
@Test
fun `network error shows error state`() = runTest {
    val viewModel = FeatureViewModel(FailingRepository())
    viewModel.fetchData()

    assertTrue(viewModel.uiState.value.hasError)
}
```

### Instrumented Tests
```kotlin
// ✅ Test UI with Compose Testing
@get:Rule
val composeTestRule = createComposeRule()

@Test
fun loginFlow() {
    composeTestRule.setContent {
        LoginScreen()
    }

    composeTestRule.onNodeWithTag("email")
        .performTextInput("user@example.com")

    composeTestRule.onNodeWithTag("password")
        .performTextInput("password")

    composeTestRule.onNodeWithText("Login")
        .performClick()

    composeTestRule.onNodeWithText("Welcome")
        .assertIsDisplayed()
}
```

---

## Build Configuration

### Gradle Build Settings
```kotlin
// build.gradle.kts (app)
android {
    compileSdk = 34

    defaultConfig {
        minSdk = 24
        targetSdk = 34
    }

    buildTypes {
        release {
            isMinifyEnabled = true
            proguardFiles(
                getDefaultProguardFile("proguard-android-optimize.txt"),
                "proguard-rules.pro"
            )
        }
    }

    composeOptions {
        kotlinCompilerExtensionVersion = "1.5.8"
    }
}
```

### ProGuard Rules
```
# Keep Compose runtime
-keep class androidx.compose.** { *; }

# Keep ViewModel
-keep class * extends androidx.lifecycle.ViewModel { *; }

# Keep serialization
-keepattributes *Annotation*, InnerClasses
-dontnote kotlinx.serialization.**
```

---

## Play Store Requirements

### Manifest Permissions
- Request only necessary permissions
- Use runtime permissions for dangerous permissions
- Provide clear rationale in UI before requesting

### Privacy and Security
- Use HTTPS for all network requests
- Encrypt sensitive data with EncryptedSharedPreferences
- Implement certificate pinning for critical APIs
- Handle deep links securely

### App Bundle
```bash
# Build App Bundle for Play Store
./gradlew bundleRelease

# Output: app/build/outputs/bundle/release/app-release.aab
```

---

## Integration with Other Skills

- **Uses:** general (core capabilities)
- **Complements:** code_review (Android-specific patterns)
- **Used by roles:** engineer, designer
- **Gates enforced:** material-design-compliance, accessibility-wcag-aa

---

## Example Usage

```
User: "Build and test the Android app"

Agent (with android_developer skill):
1. Runs ktlint for code formatting check
2. Runs Android Lint for code quality
3. Builds debug APK
4. Runs unit tests
5. Runs instrumented tests on emulator
6. Reports results with Material Design violations
```

---

## Notes

- **Prefer Jetpack Compose** for new UI code (Android 5.0+)
- **Use coroutines** over RxJava for async work
- **Support Android 7.0+** (minSdk 24) unless specified
- **Test on multiple screen sizes** (phone, tablet, foldable)
- **Profile with Android Profiler** for performance issues
- **Follow Material Design 3** guidelines
