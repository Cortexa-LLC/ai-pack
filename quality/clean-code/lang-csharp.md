# C# Language-Specific Rules

> Based on Microsoft C# Coding Conventions, StyleCop Analyzers, and .NET Best Practices

## Formatting Standards (C# Specific)

**Indentation:** **4 spaces** (no tabs)

This follows the Microsoft C# Coding Conventions and StyleCop default rules, which are the industry standard for .NET development.

**Note:** Cortexa LLC uses **language-specific indentation standards**:
- **C++**: 2 spaces (Google C++ Style Guide)
- **C#**: 4 spaces (Microsoft standard)
- **Java**: 2 spaces (Cortexa LLC override)
- **JavaScript/TypeScript**: 2 spaces (ecosystem standard)
- **Kotlin**: 4 spaces (JetBrains standard)
- **Python**: 4 spaces (PEP 8 mandatory)

**Example:**
```csharp
namespace Cortexa.Services;

using System;
using System.Collections.Generic;
using System.Linq;
using System.Threading.Tasks;

/// <summary>
/// Service for managing user data.
/// </summary>
public class UserService
{
    private readonly IUserRepository _repository;
    private readonly IEmailService _emailService;

    public UserService(IUserRepository repository, IEmailService emailService)
    {
        _repository = repository ?? throw new ArgumentNullException(nameof(repository));
        _emailService = emailService ?? throw new ArgumentNullException(nameof(emailService));
    }

    /// <summary>
    /// Retrieves a user by ID.
    /// </summary>
    /// <param name="id">The user ID.</param>
    /// <returns>The user if found, null otherwise.</returns>
    public async Task<User?> GetUserByIdAsync(long id)
    {
        // 4-space indentation throughout
        if (id <= 0)
        {
            throw new ArgumentException("ID must be positive", nameof(id));
        }

        var user = await _repository.FindByIdAsync(id);
        if (user != null)
        {
            user.LastAccessed = DateTime.UtcNow;
            await _repository.UpdateAsync(user);
        }

        return user;
    }

    /// <summary>
    /// Creates a new user.
    /// </summary>
    /// <param name="userData">The user data.</param>
    /// <returns>The created user.</returns>
    public async Task<User> CreateUserAsync(UserData userData)
    {
        ArgumentNullException.ThrowIfNull(userData);

        var user = new User
        {
            Name = userData.Name,
            Email = userData.Email,
            CreatedAt = DateTime.UtcNow
        };

        await _repository.SaveAsync(user);
        await _emailService.SendWelcomeEmailAsync(user);

        return user;
    }
}
```

---

## Overview

This file contains C#-specific best practices including:
- **Microsoft C# Coding Conventions** - Official .NET standards
- **StyleCop Analyzers** - Industry-standard static analysis
- **.NET Best Practices** - Modern .NET patterns
- **C# 12 Features** - Latest language features
- **Async/Await Patterns** - Asynchronous programming best practices

---

## Quick Standards Summary

### Formatting
- **Indentation:** 4 spaces (no tabs)
- **Brace Style:** Allman style (opening brace on new line)
- **Line Length:** 120 characters (soft limit)
- **One Statement Per Line:** No multiple statements on one line
- **Block Indentation:** +4 spaces
- **File-scoped Namespaces:** Use file-scoped namespace declarations (C# 10+)

### Naming Conventions

| Type | Convention | Example |
|------|------------|---------|
| Namespace | PascalCase | `Cortexa.Services` |
| Class | PascalCase | `UserService` |
| Interface | IPascalCase (I prefix) | `IUserRepository` |
| Method | PascalCase | `GetUserById` |
| Property | PascalCase | `FirstName` |
| Event | PascalCase | `DataReceived` |
| Constant | PascalCase | `MaxRetries` |
| Private Field | _camelCase (underscore prefix) | `_repository` |
| Parameter | camelCase | `userId` |
| Local Variable | camelCase | `user` |
| Type Parameter | TPascalCase (T prefix) | `TEntity` |

### Modifier Order (StyleCop SA1206)
```csharp
public class Example
{
    // Correct modifier order:
    // 1. Access modifiers (public, protected, internal, private)
    // 2. static
    // 3. abstract/virtual/override
    // 4. readonly/const
    // 5. extern
    // 6. new
    // 7. unsafe
    // 8. volatile
    // 9. async

    public static readonly string Constant = "value";
    private readonly IDependency _dependency;

    protected virtual async Task<int> ProcessAsync()
    {
        return await Task.FromResult(0);
    }
}
```

### Class Structure (StyleCop Ordering)
```csharp
public class Example
{
    // 1. Constants
    private const int MaxRetries = 3;

    // 2. Static fields
    private static readonly Logger _logger = LogManager.GetCurrentClassLogger();

    // 3. Instance fields
    private readonly IDependency _dependency;
    private string _cachedValue;

    // 4. Constructors
    public Example(IDependency dependency)
    {
        _dependency = dependency;
    }

    // 5. Finalizers (destructors)
    ~Example()
    {
        // Cleanup
    }

    // 6. Delegates
    public delegate void DataHandler(object sender, EventArgs e);

    // 7. Events
    public event DataHandler DataReceived;

    // 8. Properties
    public string Name { get; set; }

    public int Age { get; private set; }

    // 9. Indexers
    public string this[int index] => _items[index];

    // 10. Methods (public first, then protected, then private)
    public void PublicMethod()
    {
        // ...
    }

    protected void ProtectedMethod()
    {
        // ...
    }

    private void PrivateMethod()
    {
        // ...
    }

    // 11. Nested types
    public class NestedClass
    {
        // ...
    }
}
```

### Modern C# Features (C# 10+)

```csharp
// File-scoped namespace (C# 10+)
namespace Cortexa.Services;

// Global usings (defined in separate GlobalUsings.cs)
// global using System;
// global using System.Collections.Generic;

// Records (C# 9+)
public record User(long Id, string Name, string Email)
{
    // Init-only properties
    public DateTime CreatedAt { get; init; } = DateTime.UtcNow;

    // With expressions for immutable updates
    public User WithName(string newName) => this with { Name = newName };
}

// Record structs (C# 10+)
public readonly record struct Point(double X, double Y);

// Pattern matching
public string ClassifyUser(User user) => user switch
{
    { Age: < 18 } => "Minor",
    { Age: >= 18 and < 65 } => "Adult",
    { Age: >= 65 } => "Senior",
    _ => "Unknown"
};

// Target-typed new (C# 9+)
List<string> names = new();
User user = new() { Name = "John", Email = "john@example.com" };

// Null-coalescing assignment (C# 8+)
_cache ??= new Dictionary<string, object>();

// Nullable reference types (C# 8+)
#nullable enable
public string? FindUser(int id)  // May return null
{
    return _users.ContainsKey(id) ? _users[id] : null;
}

// Required properties (C# 11+)
public class Config
{
    public required string ApiUrl { get; init; }
    public required string ApiKey { get; init; }
}

// Raw string literals (C# 11+)
string json = """
    {
        "name": "John",
        "email": "john@example.com"
    }
    """;

// List patterns (C# 11+)
public bool IsValidSequence(int[] numbers) => numbers switch
{
    [1, 2, 3] => true,
    [var first, .., var last] when first == last => true,
    _ => false
};
```

### Async/Await Best Practices

```csharp
// Always use Async suffix for async methods
public async Task<User> GetUserAsync(int id)
{
    return await _repository.FindByIdAsync(id);
}

// ConfigureAwait(false) for library code
public async Task<Data> FetchDataAsync()
{
    return await _httpClient.GetAsync(url).ConfigureAwait(false);
}

// ValueTask for hot paths
public ValueTask<int> GetCachedValueAsync(string key)
{
    if (_cache.TryGetValue(key, out var value))
    {
        return new ValueTask<int>(value);
    }

    return new ValueTask<int>(FetchFromDatabaseAsync(key));
}

// Cancellation tokens
public async Task<User> GetUserAsync(int id, CancellationToken cancellationToken)
{
    return await _repository.FindByIdAsync(id, cancellationToken);
}

// Task.WhenAll for parallel operations
public async Task<IEnumerable<User>> GetUsersAsync(IEnumerable<int> ids)
{
    var tasks = ids.Select(id => GetUserAsync(id));
    return await Task.WhenAll(tasks);
}
```

### LINQ Best Practices

```csharp
// Method syntax for complex queries
var adults = users
    .Where(u => u.Age >= 18)
    .OrderBy(u => u.Name)
    .Select(u => new UserDto
    {
        Name = u.Name,
        Email = u.Email
    })
    .ToList();

// Query syntax for joins
var userOrders = from user in users
                 join order in orders on user.Id equals order.UserId
                 where order.Total > 100
                 select new { user.Name, order.Total };

// Use Any() instead of Count() > 0
if (users.Any(u => u.IsActive))
{
    // More efficient than users.Count(u => u.IsActive) > 0
}

// Use SingleOrDefault for single results
var user = users.SingleOrDefault(u => u.Id == targetId);
```

---

## StyleCop Analyzer Rules

**MANDATORY:** All C# code must pass StyleCop Analyzers checks.

### Installation

Add to your project file:

```xml
<ItemGroup>
  <PackageReference Include="StyleCop.Analyzers" Version="1.2.0-beta.435">
    <PrivateAssets>all</PrivateAssets>
    <IncludeAssets>runtime; build; native; contentfiles; analyzers; buildtransitive</IncludeAssets>
  </PackageReference>
</ItemGroup>
```

### StyleCop Configuration (stylecop.json)

```json
{
  "$schema": "https://raw.githubusercontent.com/DotNetAnalyzers/StyleCopAnalyzers/master/StyleCop.Analyzers/StyleCop.Analyzers/Settings/stylecop.schema.json",
  "settings": {
    "documentationRules": {
      "companyName": "Cortexa LLC",
      "copyrightText": "Copyright (c) {companyName}. All rights reserved.",
      "xmlHeader": true,
      "fileNamingConvention": "stylecop"
    },
    "orderingRules": {
      "usingDirectivesPlacement": "outsideNamespace",
      "systemUsingDirectivesFirst": true
    },
    "namingRules": {
      "allowCommonHungarianPrefixes": false,
      "allowedHungarianPrefixes": []
    }
  }
}
```

### Critical StyleCop Rules

#### Documentation Rules (SA1600-SA1648)
- **SA1600:** Elements must be documented
- **SA1633:** File must have header
- **SA1642:** Constructor summary documentation must begin with standard text

```csharp
/// <summary>
/// Represents a user in the system.
/// </summary>
public class User
{
    /// <summary>
    /// Initializes a new instance of the <see cref="User"/> class.
    /// </summary>
    /// <param name="id">The user ID.</param>
    /// <param name="name">The user name.</param>
    public User(int id, string name)
    {
        Id = id;
        Name = name;
    }

    /// <summary>
    /// Gets or sets the user ID.
    /// </summary>
    public int Id { get; set; }

    /// <summary>
    /// Gets or sets the user name.
    /// </summary>
    public string Name { get; set; }
}
```

#### Spacing Rules (SA1000-SA1028)
- **SA1000:** Keywords must be spaced correctly
- **SA1001:** Commas must be spaced correctly
- **SA1008:** Opening parenthesis must be spaced correctly
- **SA1009:** Closing parenthesis must be spaced correctly

```csharp
// Correct spacing
if (condition)
{
    DoSomething(param1, param2);
}

// Wrong spacing
if(condition)
{
    DoSomething(param1,param2);
}
```

#### Readability Rules (SA1100-SA1142)
- **SA1101:** Prefix local calls with this (disabled by default)
- **SA1116:** Split parameters should start on line after declaration
- **SA1117:** Parameters must be on same line or separate lines
- **SA1124:** Do not use regions
- **SA1133:** Do not combine attributes

```csharp
// Correct attribute placement
[Serializable]
[DataContract]
public class User
{
    [DataMember]
    public string Name { get; set; }
}

// Wrong attribute placement
[Serializable, DataContract]  // SA1133 violation
public class User
{
}
```

#### Ordering Rules (SA1200-SA1214)
- **SA1200:** Using directives must be placed correctly
- **SA1201:** Elements must appear in the correct order
- **SA1202:** Elements must be ordered by access
- **SA1204:** Static elements must appear before instance elements

```csharp
// Correct using placement (outside namespace)
using System;
using System.Collections.Generic;

namespace Cortexa.Services
{
    public class Example
    {
        // Constants first
        private const int MaxSize = 100;

        // Static fields before instance fields
        private static readonly Logger _logger = new();
        private readonly IService _service;

        // Public members before private
        public void PublicMethod() { }
        private void PrivateMethod() { }
    }
}
```

#### Naming Rules (SA1300-SA1313)
- **SA1300:** Element must begin with upper-case letter
- **SA1302:** Interface names must begin with I
- **SA1303:** Const field names must begin with upper-case letter
- **SA1306:** Field names must begin with lower-case letter
- **SA1307:** Accessible fields must begin with upper-case letter
- **SA1309:** Field names must not begin with underscore (disabled - we use underscore for private fields)

```csharp
// Correct naming
public interface IUserService { }
public class UserService { }
private readonly IRepository _repository;
public const int MaxRetries = 3;
```

#### Maintainability Rules (SA1400-SA1413)
- **SA1400:** Access modifier must be declared
- **SA1401:** Fields must be private
- **SA1402:** File may only contain a single type
- **SA1404:** Code analysis suppression must have justification

```csharp
// Each type in separate file
// File: UserService.cs
public class UserService  // SA1400: explicit 'public'
{
    private readonly IRepository _repository;  // SA1401: private field

    [SuppressMessage("StyleCop.CSharp.DocumentationRules", "SA1600",
        Justification = "Internal utility method")]  // SA1404: justification provided
    internal void UtilityMethod() { }
}
```

---

## .NET Best Practices

### Dependency Injection

```csharp
// Use constructor injection
public class UserService
{
    private readonly IUserRepository _repository;
    private readonly ILogger<UserService> _logger;

    public UserService(IUserRepository repository, ILogger<UserService> logger)
    {
        _repository = repository ?? throw new ArgumentNullException(nameof(repository));
        _logger = logger ?? throw new ArgumentNullException(nameof(logger));
    }
}

// Register in Program.cs
builder.Services.AddScoped<IUserRepository, UserRepository>();
builder.Services.AddScoped<UserService>();
```

### Exception Handling

```csharp
// Be specific with exceptions
public User GetUser(int id)
{
    if (id <= 0)
    {
        throw new ArgumentException("ID must be positive", nameof(id));
    }

    var user = _repository.Find(id);
    if (user == null)
    {
        throw new UserNotFoundException($"User with ID {id} not found");
    }

    return user;
}

// Use guard clauses
public void ProcessOrder(Order order)
{
    ArgumentNullException.ThrowIfNull(order);

    if (order.Items.Count == 0)
    {
        throw new InvalidOperationException("Order must contain items");
    }

    // Process order
}

// Don't catch generic exceptions
try
{
    ProcessData();
}
catch (IOException ex)  // Specific exception
{
    _logger.LogError(ex, "Failed to read data");
    throw;
}
catch (Exception)  // Avoid this
{
    // Too generic
}
```

### Resource Management

```csharp
// Use 'using' declarations (C# 8+)
public async Task<string> ReadFileAsync(string path)
{
    using var stream = File.OpenRead(path);
    using var reader = new StreamReader(stream);
    return await reader.ReadToEndAsync();
}

// IDisposable implementation
public class DatabaseConnection : IDisposable
{
    private bool _disposed;
    private SqlConnection _connection;

    public void Dispose()
    {
        Dispose(true);
        GC.SuppressFinalize(this);
    }

    protected virtual void Dispose(bool disposing)
    {
        if (!_disposed)
        {
            if (disposing)
            {
                // Dispose managed resources
                _connection?.Dispose();
            }

            // Dispose unmanaged resources
            _disposed = true;
        }
    }

    ~DatabaseConnection()
    {
        Dispose(false);
    }
}
```

### Collections

```csharp
// Use collection expressions (C# 12+)
List<int> numbers = [1, 2, 3, 4, 5];
int[] array = [1, 2, 3];

// Use IReadOnlyCollection/IReadOnlyList for immutable interfaces
public IReadOnlyList<User> GetUsers()
{
    return _users.AsReadOnly();
}

// Use immutable collections for thread-safety
private readonly ImmutableList<string> _allowedRoles =
    ImmutableList.Create("Admin", "User", "Guest");
```

### Performance

```csharp
// Use Span<T> for performance-critical code
public int Sum(ReadOnlySpan<int> numbers)
{
    int total = 0;
    foreach (var num in numbers)
    {
        total += num;
    }
    return total;
}

// Use StringBuilder for string concatenation
public string BuildReport(IEnumerable<User> users)
{
    var sb = new StringBuilder();
    foreach (var user in users)
    {
        sb.AppendLine($"{user.Name}: {user.Email}");
    }
    return sb.ToString();
}

// Use ArrayPool for temporary arrays
var pool = ArrayPool<byte>.Shared;
byte[] buffer = pool.Rent(1024);
try
{
    // Use buffer
}
finally
{
    pool.Return(buffer);
}
```

---

## EditorConfig for C#

Add to `.editorconfig`:

```ini
# C# files
[*.cs]

# Indentation
indent_style = space
indent_size = 4
tab_width = 4

# New line preferences
end_of_line = crlf
insert_final_newline = true
charset = utf-8-bom

# Organize usings
dotnet_sort_system_directives_first = true
dotnet_separate_import_directive_groups = false

# Language conventions
dotnet_style_qualification_for_field = false:warning
dotnet_style_qualification_for_property = false:warning
dotnet_style_qualification_for_method = false:warning
dotnet_style_qualification_for_event = false:warning

# Use language keywords instead of framework type names
dotnet_style_predefined_type_for_locals_parameters_members = true:warning
dotnet_style_predefined_type_for_member_access = true:warning

# Modifier preferences
dotnet_style_require_accessibility_modifiers = always:warning
csharp_preferred_modifier_order = public,private,protected,internal,static,extern,new,virtual,abstract,sealed,override,readonly,unsafe,volatile,async:warning

# Expression preferences
dotnet_style_prefer_auto_properties = true:warning
dotnet_style_prefer_conditional_expression_over_assignment = true:suggestion
dotnet_style_prefer_conditional_expression_over_return = false:suggestion
dotnet_style_prefer_inferred_tuple_names = true:suggestion
dotnet_style_prefer_inferred_anonymous_type_member_names = true:suggestion

# Pattern matching
csharp_style_pattern_matching_over_is_with_cast_check = true:warning
csharp_style_pattern_matching_over_as_with_null_check = true:warning

# Null checking
csharp_style_throw_expression = true:suggestion
csharp_style_conditional_delegate_call = true:warning

# Expression-bodied members
csharp_style_expression_bodied_methods = when_on_single_line:suggestion
csharp_style_expression_bodied_constructors = false:suggestion
csharp_style_expression_bodied_operators = when_on_single_line:suggestion
csharp_style_expression_bodied_properties = when_on_single_line:suggestion
csharp_style_expression_bodied_indexers = when_on_single_line:suggestion
csharp_style_expression_bodied_accessors = when_on_single_line:suggestion

# Var preferences
csharp_style_var_for_built_in_types = true:warning
csharp_style_var_when_type_is_apparent = true:warning
csharp_style_var_elsewhere = true:suggestion

# Code style
csharp_prefer_braces = true:warning
csharp_prefer_simple_using_statement = true:suggestion
csharp_style_namespace_declarations = file_scoped:warning
csharp_style_prefer_method_group_conversion = true:suggestion
csharp_style_prefer_top_level_statements = true:suggestion

# Formatting
csharp_new_line_before_open_brace = all
csharp_new_line_before_else = true
csharp_new_line_before_catch = true
csharp_new_line_before_finally = true
csharp_new_line_before_members_in_object_initializers = true
csharp_new_line_before_members_in_anonymous_types = true
csharp_new_line_between_query_expression_clauses = true

# Spacing
csharp_space_after_cast = false
csharp_space_after_keywords_in_control_flow_statements = true
csharp_space_between_method_call_parameter_list_parentheses = false
csharp_space_between_method_declaration_parameter_list_parentheses = false
csharp_space_between_parentheses = false
csharp_space_before_colon_in_inheritance_clause = true
csharp_space_after_colon_in_inheritance_clause = true

# Wrapping
csharp_preserve_single_line_statements = false
csharp_preserve_single_line_blocks = true
```

---

## TODO: Full C# Guidelines

This file will be expanded to include:
- [ ] ASP.NET Core best practices
- [ ] Entity Framework Core patterns
- [ ] Blazor component patterns
- [ ] Testing with xUnit and NUnit
- [ ] Minimal APIs patterns
- [ ] gRPC service patterns
- [ ] Logging and monitoring
- [ ] Security best practices
- [ ] Performance optimization

---

**For now, always use 4-space indentation per Microsoft conventions. All code must pass StyleCop Analyzers checks. Full guidelines coming soon.**
