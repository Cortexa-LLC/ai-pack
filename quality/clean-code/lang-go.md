# Go Language-Specific Rules

> Based on Effective Go, Go Code Review Comments, Uber Go Style Guide, and Go community best practices

## Formatting Standards (Go Specific)

**Indentation:** **Tabs** (converted to appropriate spaces by tooling)

Go is the **only language** in our supported stack that uses tabs instead of spaces. This is mandated by `gofmt` and is non-negotiable in the Go community.

**Note:** Each language follows its community's indentation standards:
- **C++**: 2 spaces (see lang-cpp.md)
- **Python**: 4 spaces (see lang-python.md)
- **JavaScript/TypeScript**: 2 spaces (see lang-javascript.md)
- **Java**: 4 spaces (see lang-java.md)
- **Kotlin**: 4 spaces (see lang-kotlin.md)
- **Go**: Tabs (this document) - **gofmt standard**

**Automatic Formatting:**
```bash
# Format all Go files (required before commit)
gofmt -w .

# Or use goimports (preferred - also manages imports)
goimports -w .
```

**Example:**
```go
package service

import (
	"context"
	"fmt"
	"time"

	"github.com/yourorg/yourproject/internal/domain"
)

// UserService manages user operations.
type UserService struct {
	repo  domain.UserRepository
	cache domain.CacheService
}

// NewUserService creates a new user service.
func NewUserService(repo domain.UserRepository, cache domain.CacheService) *UserService {
	return &UserService{
		repo:  repo,
		cache: cache,
	}
}

// GetUser retrieves a user by ID.
func (s *UserService) GetUser(ctx context.Context, id string) (*domain.User, error) {
	// Check cache first
	if user, err := s.cache.Get(ctx, id); err == nil {
		return user, nil
	}

	// Fetch from repository
	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to find user %s: %w", id, err)
	}

	// Update cache
	_ = s.cache.Set(ctx, id, user, 5*time.Minute)

	return user, nil
}
```

---

## Overview

This file contains Go-specific best practices including:
- **Effective Go** - Official Go documentation
- **Go Code Review Comments** - Common mistakes in Go code reviews
- **Uber Go Style Guide** - Enterprise Go patterns
- **Go Proverbs** - Rob Pike's Go wisdom

---

## Quick Standards Summary

### Formatting
- **Indentation:** Tabs (gofmt enforced)
- **Line Length:** No hard limit, but keep reasonable (~100 chars)
- **Brace Style:** K&R (required by gofmt) - opening brace on same line
- **Use gofmt/goimports:** Always format code before commit

### Naming
- `packagename` - short, lowercase, no underscores
- `TypeName` - MixedCaps (PascalCase)
- `MethodName` - MixedCaps (PascalCase)
- `functionName` - mixedCaps (camelCase) for unexported
- `variableName` - mixedCaps or short (i, err, ctx)
- `ConstantName` - MixedCaps (not UPPER_CASE like other languages)

### Package Structure
```go
package domain

import (
	// Standard library first
	"context"
	"fmt"
	"time"

	// Third-party packages next
	"github.com/google/uuid"

	// Your own packages last
	"github.com/yourorg/project/internal/errors"
)

// Exported type
type User struct {
	ID        string
	Name      string
	Email     string
	CreatedAt time.Time
}

// unexported type
type userCache struct {
	// ...
}
```

---

## Naming Conventions

### Package Names
**DO:**
```go
package http      // short, lowercase
package user      // singular
package template  // not "templates"
```

**DON'T:**
```go
package httpServer  // no mixed caps
package users       // avoid plural
package utils       // too generic
package common      // meaningless
```

### Type and Interface Names
**DO:**
```go
type User struct { }          // MixedCaps
type UserRepository interface { }
type HTTPClient struct { }    // Acronyms all caps
type APIKey string           // Not ApiKey
```

**DON'T:**
```go
type userType struct { }     // unexported when should be exported
type IUserRepository interface { }  // no "I" prefix
type UserRepositoryInterface interface { }  // redundant "Interface"
```

### Interface Naming
**Single-method interfaces:** Use method name + "er"
```go
type Reader interface {
	Read(p []byte) (n int, err error)
}

type Writer interface {
	Write(p []byte) (n int, err error)
}

type Closer interface {
	Close() error
}
```

**Multi-method interfaces:** Use descriptive name without "er"
```go
type UserRepository interface {
	FindByID(ctx context.Context, id string) (*User, error)
	Save(ctx context.Context, user *User) error
	Delete(ctx context.Context, id string) error
}
```

### Variable Names
**Short names for short scopes:**
```go
// Good
for i := 0; i < n; i++ {
	// i is fine here
}

// Good
if err := doSomething(); err != nil {
	return err
}

// Good
ctx := context.Background()
```

**Longer names for broader scopes:**
```go
// Package-level variable
var DefaultTimeout = 30 * time.Second

// Struct field
type Config struct {
	DatabaseConnectionString string
	MaxRetryAttempts        int
}
```

### Avoid Stutter
**DON'T repeat package name in type names:**
```go
// DON'T
user.UserService   // package user, type UserService

// DO
user.Service       // package user, type Service
```

---

## Error Handling

### Always Check Errors
**DO:**
```go
result, err := doSomething()
if err != nil {
	return fmt.Errorf("failed to do something: %w", err)
}
```

**DON'T:**
```go
result, _ := doSomething()  // ignoring errors
```

### Error Wrapping
Use `fmt.Errorf` with `%w` to wrap errors:
```go
func (s *Service) Process(id string) error {
	user, err := s.repo.Find(id)
	if err != nil {
		return fmt.Errorf("failed to find user %s: %w", id, err)
	}
	// ...
}
```

### Custom Error Types
```go
// Define custom error type
type ValidationError struct {
	Field string
	Err   error
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation failed for %s: %v", e.Field, e.Err)
}

func (e *ValidationError) Unwrap() error {
	return e.Err
}

// Usage
func Validate(user *User) error {
	if user.Email == "" {
		return &ValidationError{
			Field: "email",
			Err:   errors.New("required"),
		}
	}
	return nil
}
```

### Sentinel Errors
```go
// Define package-level errors for common cases
var (
	ErrNotFound      = errors.New("not found")
	ErrUnauthorized  = errors.New("unauthorized")
	ErrInvalidInput  = errors.New("invalid input")
)

// Check with errors.Is
if errors.Is(err, ErrNotFound) {
	// handle not found
}
```

---

## Context Usage

### Always Pass Context as First Parameter
```go
// DO
func (s *Service) GetUser(ctx context.Context, id string) (*User, error)

// DON'T
func (s *Service) GetUser(id string, ctx context.Context) (*User, error)
```

### Don't Store Context in Structs
**DON'T:**
```go
type Service struct {
	ctx context.Context  // NO!
}
```

**DO:**
```go
type Service struct {
	// Store dependencies, not context
	repo Repository
}

func (s *Service) DoSomething(ctx context.Context) error {
	// Pass context as parameter
}
```

### Use Context for Cancellation and Timeouts
```go
func (s *Service) CallExternalAPI(ctx context.Context) error {
	// Create timeout context
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return err
	}

	resp, err := s.client.Do(req)
	// ...
}
```

---

## Concurrency

### Use Channels to Communicate
**DO:**
```go
func worker(jobs <-chan Job, results chan<- Result) {
	for job := range jobs {
		results <- processJob(job)
	}
}

func main() {
	jobs := make(chan Job, 100)
	results := make(chan Result, 100)

	// Start workers
	for i := 0; i < 3; i++ {
		go worker(jobs, results)
	}

	// Send jobs
	for _, job := range allJobs {
		jobs <- job
	}
	close(jobs)

	// Collect results
	for i := 0; i < len(allJobs); i++ {
		result := <-results
		// handle result
	}
}
```

### Use sync.WaitGroup for Coordination
```go
func processItems(items []Item) error {
	var wg sync.WaitGroup
	errChan := make(chan error, len(items))

	for _, item := range items {
		wg.Add(1)
		go func(it Item) {
			defer wg.Done()
			if err := process(it); err != nil {
				errChan <- err
			}
		}(item)
	}

	wg.Wait()
	close(errChan)

	// Check for errors
	for err := range errChan {
		if err != nil {
			return err
		}
	}
	return nil
}
```

### Avoid Goroutine Leaks
**Always ensure goroutines can exit:**
```go
func search(ctx context.Context, query string) (Result, error) {
	results := make(chan Result, 1)

	go func() {
		// Ensure we can exit if context is cancelled
		select {
		case results <- expensiveSearch(query):
		case <-ctx.Done():
			return
		}
	}()

	select {
	case result := <-results:
		return result, nil
	case <-ctx.Done():
		return Result{}, ctx.Err()
	}
}
```

---

## Struct Design

### Use Pointers for Large Structs or When Mutation Needed
```go
// Small struct - value receiver is fine
type Point struct {
	X, Y int
}

func (p Point) String() string {
	return fmt.Sprintf("(%d, %d)", p.X, p.Y)
}

// Large struct or needs mutation - pointer receiver
type User struct {
	ID        string
	Name      string
	Email     string
	Metadata  map[string]string
}

func (u *User) Update(name, email string) {
	u.Name = name
	u.Email = email
}
```

### Use Functional Options for Complex Constructors
```go
type Server struct {
	addr    string
	timeout time.Duration
	logger  Logger
}

type Option func(*Server)

func WithTimeout(d time.Duration) Option {
	return func(s *Server) {
		s.timeout = d
	}
}

func WithLogger(logger Logger) Option {
	return func(s *Server) {
		s.logger = logger
	}
}

func NewServer(addr string, opts ...Option) *Server {
	s := &Server{
		addr:    addr,
		timeout: 30 * time.Second, // default
		logger:  defaultLogger,     // default
	}

	for _, opt := range opts {
		opt(s)
	}

	return s
}

// Usage
server := NewServer(
	":8080",
	WithTimeout(60*time.Second),
	WithLogger(customLogger),
)
```

### Zero Values Are Useful
Design structs so zero values are valid:
```go
// Good - zero value is valid
type Config struct {
	Timeout time.Duration // 0 is valid
	Retries int          // 0 is valid
}

// Better - with defaults
func NewConfig() *Config {
	return &Config{
		Timeout: 30 * time.Second,
		Retries: 3,
	}
}
```

---

## Interface Design

### Accept Interfaces, Return Concrete Types
```go
// DON'T accept concrete type
func ProcessUser(u *User) error

// DO accept interface
type UserLike interface {
	GetID() string
	GetEmail() string
}

func ProcessUser(u UserLike) error
```

```go
// DON'T return interface unnecessarily
func NewUserService() UserRepository

// DO return concrete type
func NewUserService() *UserService
```

### Keep Interfaces Small
```go
// DON'T - too many methods
type Repository interface {
	Create(user *User) error
	Update(user *User) error
	Delete(id string) error
	FindByID(id string) (*User, error)
	FindByEmail(email string) (*User, error)
	List(page, size int) ([]*User, error)
}

// DO - split into focused interfaces
type UserCreator interface {
	Create(ctx context.Context, user *User) error
}

type UserUpdater interface {
	Update(ctx context.Context, user *User) error
}

type UserFinder interface {
	FindByID(ctx context.Context, id string) (*User, error)
}
```

### Define Interfaces Where They're Used
```go
// DON'T define in package with implementation
package storage

type UserRepository interface {  // defined in storage package
	// ...
}

type postgresRepo struct { }  // implementation
```

```go
// DO define in package that uses it
package service

type UserRepository interface {  // defined where it's used
	// ...
}

type UserService struct {
	repo UserRepository  // depends on interface
}
```

---

## Testing

### Table-Driven Tests
```go
func TestValidateEmail(t *testing.T) {
	tests := []struct {
		name    string
		email   string
		wantErr bool
	}{
		{
			name:    "valid email",
			email:   "user@example.com",
			wantErr: false,
		},
		{
			name:    "missing @",
			email:   "userexample.com",
			wantErr: true,
		},
		{
			name:    "empty email",
			email:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateEmail(tt.email)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateEmail() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
```

### Use t.Helper() for Test Helpers
```go
func assertNoError(t *testing.T, err error) {
	t.Helper()  // Marks this as helper - errors point to caller
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSomething(t *testing.T) {
	err := doSomething()
	assertNoError(t, err)  // Error points to this line, not inside assertNoError
}
```

### Use Subtests for Parallel Testing
```go
func TestConcurrent(t *testing.T) {
	tests := []struct {
		name string
		// ...
	}{
		// ...
	}

	for _, tt := range tests {
		tt := tt  // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()  // Run this subtest in parallel
			// test logic
		})
	}
}
```

---

## Common Mistakes to Avoid

### Don't Use `init()` for Side Effects
**DON'T:**
```go
func init() {
	db = connectToDatabase()  // side effects in init!
}
```

**DO:**
```go
func NewService() (*Service, error) {
	db, err := connectToDatabase()
	if err != nil {
		return nil, err
	}
	return &Service{db: db}, nil
}
```

### Don't Ignore Errors from defer
**DON'T:**
```go
defer file.Close()  // error ignored
```

**DO:**
```go
defer func() {
	if err := file.Close(); err != nil {
		log.Printf("failed to close file: %v", err)
	}
}()
```

### Don't Use Pointer to Interface
**DON'T:**
```go
func Process(r *io.Reader) error  // pointer to interface is wrong
```

**DO:**
```go
func Process(r io.Reader) error  // interfaces are already references
```

### Avoid Empty Interface
**DON'T:**
```go
func Process(data interface{}) error  // loses type safety
```

**DO:**
```go
func Process(data map[string]string) error  // use specific type
// Or use generics in Go 1.18+
func Process[T any](data T) error
```

---

## Code Organization

### Package Layout
```
project/
├── cmd/
│   └── server/
│       └── main.go          # Application entry points
├── internal/
│   ├── domain/              # Core business logic
│   │   ├── user.go
│   │   └── user_test.go
│   ├── repository/          # Data access
│   │   └── postgres/
│   │       └── user_repo.go
│   ├── service/             # Business services
│   │   └── user_service.go
│   └── http/                # HTTP handlers
│       └── user_handler.go
├── pkg/                     # Public libraries (if any)
├── go.mod
└── go.sum
```

### File Naming
- `user_service.go` - snake_case for files
- `user_service_test.go` - tests in same package
- `user_service_integration_test.go` - integration tests

---

## Go Proverbs (Rob Pike)

- Don't communicate by sharing memory, share memory by communicating
- Concurrency is not parallelism
- Channels orchestrate; mutexes serialize
- The bigger the interface, the weaker the abstraction
- Make the zero value useful
- interface{} says nothing
- Gofmt's style is no one's favorite, yet gofmt is everyone's favorite
- A little copying is better than a little dependency
- Syscall must always be guarded with build tags
- Cgo must always be guarded with build tags
- Cgo is not Go
- With the unsafe package there are no guarantees
- Clear is better than clever
- Reflection is never clear
- Errors are values
- Don't just check errors, handle them gracefully
- Design the architecture, name the components, document the details
- Documentation is for users
- Don't panic

---

## Tools and Linters

### Required Tools
```bash
# Format code
go fmt ./...
goimports -w .

# Vet code
go vet ./...

# Run tests
go test ./...
go test -race ./...  # with race detector

# golangci-lint (comprehensive linter)
golangci-lint run
```

### Recommended golangci-lint Configuration
```yaml
# .golangci.yml
linters:
  enable:
    - gofmt
    - goimports
    - govet
    - errcheck
    - staticcheck
    - unused
    - gosimple
    - structcheck
    - varcheck
    - ineffassign
    - deadcode
    - typecheck
    - gocritic
    - gocyclo
    - dupl
    - misspell
    - revive

linters-settings:
  gocyclo:
    min-complexity: 15
  dupl:
    threshold: 100
```

---

## References

- [Effective Go](https://golang.org/doc/effective_go)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- [Uber Go Style Guide](https://github.com/uber-go/guide/blob/master/style.md)
- [Go Proverbs](https://go-proverbs.github.io/)
- [Standard Package Layout](https://github.com/golang-standards/project-layout)
