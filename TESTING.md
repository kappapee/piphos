# Testing Guide for Piphos

This document describes the testing strategy and conventions used in the Piphos project.

## Overview

Piphos has a comprehensive test suite that follows Go idioms and best practices. The test suite achieves **81.1% overall code coverage** with critical packages reaching even higher coverage rates.

## Coverage Summary

| Package | Coverage | Test File |
|---------|----------|-----------|
| `internal/validate` | 100.0% | `validate_test.go` |
| `internal/exec` | 96.8% | `command_test.go` |
| `internal/beacon` | 92.3% | `beacon_test.go`, `web_test.go` |
| `internal/tender` | 88.3% | `github_test.go`, `tender_test.go` |
| `internal/config` | N/A (constants only) | - |

## Running Tests

### Run All Tests
```bash
go test ./...
```

### Run Tests with Verbose Output
```bash
go test -v ./...
```

### Run Tests with Coverage
```bash
go test -cover ./...
```

### Generate Coverage Report
```bash
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out
```

### View Coverage in Browser
```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Run Tests for a Specific Package
```bash
go test ./internal/validate
go test ./internal/beacon
go test ./internal/tender
go test ./internal/exec
```

## Testing Patterns

### Table-Driven Tests

All tests follow Go's table-driven testing pattern for clarity and maintainability:

```go
func TestFunction(t *testing.T) {
    tests := []struct {
        name          string
        input         string
        expectedError bool
    }{
        {name: "valid case", input: "value", expectedError: false},
        {name: "invalid case", input: "", expectedError: true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := Function(tt.input)
            if tt.expectedError && err == nil {
                t.Error("expected error but got nil")
            }
            if !tt.expectedError && err != nil {
                t.Errorf("expected no error but got: %v", err)
            }
        })
    }
}
```

### HTTP Mocking

Tests use Go's built-in `httptest` package to mock HTTP servers, avoiding external dependencies:

```go
server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("203.0.113.1"))
}))
defer server.Close()
```

### Context Testing

Tests verify proper context handling for timeouts and cancellations:

```go
ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
defer cancel()

_, err := b.Ping(ctx)
if err == nil {
    t.Error("expected timeout error but got nil")
}
```

## Test Coverage by Package

### `internal/validate` (100% Coverage)

Tests all validation functions:
- `TestCommand`: Validates command-line argument counts
- `TestIP`: Validates IPv4 and IPv6 addresses with edge cases
- `TestToken`: Validates authentication tokens

**Coverage**: All validation logic including error paths

### `internal/beacon` (92.3% Coverage)

Tests beacon provider factory and HTTP-based IP detection:

**beacon_test.go** - Tests for beacon.go:
- `TestNew`: Validates beacon provider creation (haz, aws, unknown)
- `TestNewBeaconURLs`: Verifies correct URLs for each beacon provider

**web_test.go** - Tests for web.go:
- `TestWebPing`: Tests IP detection with various responses (IPv4, IPv6, errors)
- `TestWebPingTimeout`: Validates timeout handling
- `TestWebPingCancellation`: Tests context cancellation
- `TestWebPingLargeResponse`: Validates response size handling
- `TestWebPingReadBodyError`: Tests read error handling
- `TestNewWeb`: Verifies web beacon construction

**Coverage**: Factory pattern, HTTP requests, error handling, context management, edge cases

### `internal/tender` (88.3% Coverage)

Tests tender provider factory and GitHub Gist storage:
- `TestNew`: Validates tender provider creation with environment variables
- `TestGithubPull_*`: Tests retrieving data from GitHub Gists
  - No existing gist
  - Valid gist with data
  - Truncated gist (error case)
  - Missing file (error case)
  - Invalid JSON content
- `TestGithubPush_*`: Tests storing data to GitHub Gists
  - Creating new gist
  - Updating existing gist
  - Skipping unchanged IP (optimization)
  - Error responses
- `TestGithubRequestTimeout`: Validates timeout handling
- `TestGithubHeaders`: Verifies correct HTTP headers
- `TestGithubInvalidJSON`: Tests JSON parsing errors

**Coverage**: Factory pattern, GitHub API interactions, error handling, edge cases

### `internal/exec` (96.8% Coverage)

Tests command execution logic:
- `TestPing`: Tests IP detection command with various beacon providers
- `TestPingExtraArguments`: Validates argument parsing
- `TestPull`: Tests retrieving hostname→IP mappings
- `TestPullMissingToken`: Validates environment variable requirements
- `TestPush`: Tests updating hostname→IP mappings
- `TestPushMissingToken`: Validates authentication
- `TestHelp`: Verifies help output

**Coverage**: Command coordination, flag parsing, provider initialization

## Testing Best Practices Used

### 1. **No External Dependencies**
- All tests are self-contained
- Use `httptest` for mocking HTTP servers
- No third-party mocking frameworks needed

### 2. **Table-Driven Tests**
- Clear test case descriptions
- Easy to add new test cases
- Consistent structure across all tests

### 3. **Subtests with t.Run**
- Isolated test execution
- Clear failure reporting
- Ability to run specific test cases

### 4. **Comprehensive Error Testing**
- Valid input paths
- Invalid input paths
- Edge cases (empty strings, malformed data)
- Network errors (timeouts, bad status codes)
- JSON parsing errors

### 5. **Context-Aware Testing**
- Timeout behavior
- Cancellation behavior
- Proper context propagation

### 6. **Resource Cleanup**
- Use `defer` for server cleanup
- Use `defer` for environment variable cleanup
- Proper HTTP response body closing

### 7. **Realistic Test Data**
- Use RFC 5737 TEST-NET addresses (203.0.113.x)
- Use realistic IPv6 addresses
- Test both valid and invalid IP formats

## What's Not Tested

### `cmd/piphos/main.go` (0% Coverage)
The main entry point is not tested as it's a thin wrapper around the `exec` package. Integration testing of the CLI is better suited for end-to-end tests rather than unit tests.

### `internal/config/config.go`
This file contains only constants and requires no testing.

## Future Testing Improvements

If you want to extend the test suite, consider:

1. **Integration Tests**: Test the full CLI workflow end-to-end
2. **Benchmark Tests**: Add performance benchmarks for critical paths
3. **Fuzz Testing**: Use Go's built-in fuzzing for input validation
4. **Race Detection**: Run tests with `-race` flag in CI
5. **Example Tests**: Add testable examples in documentation

## Continuous Integration

The test suite runs automatically on GitHub Actions:
- `.github/workflows/check.yaml` runs tests on every push
- Tests run on Go 1.24 and 1.25
- Coverage reports are generated

## Writing New Tests

When adding new features, follow these guidelines:

1. **Write tests first** (TDD approach when appropriate)
2. **Use table-driven tests** for multiple scenarios
3. **Test error paths** as thoroughly as success paths
4. **Mock external dependencies** using `httptest`
5. **Use meaningful test names** that describe the scenario
6. **Keep tests focused** - one behavior per test
7. **Clean up resources** with `defer`
8. **Aim for 80%+ coverage** on new code

## Example: Adding a New Test

```go
func TestNewFeature(t *testing.T) {
    tests := []struct {
        name          string
        input         string
        expected      string
        expectedError bool
    }{
        {
            name:          "valid input",
            input:         "test",
            expected:      "result",
            expectedError: false,
        },
        {
            name:          "invalid input",
            input:         "",
            expectedError: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result, err := NewFeature(tt.input)
            
            if tt.expectedError {
                if err == nil {
                    t.Error("expected error but got nil")
                }
                return
            }
            
            if err != nil {
                t.Errorf("expected no error but got: %v", err)
            }
            
            if result != tt.expected {
                t.Errorf("expected %s but got %s", tt.expected, result)
            }
        })
    }
}
```

## Troubleshooting

### Tests Hang
- Check for missing context timeouts
- Ensure HTTP servers are properly closed with `defer`

### Flaky Tests
- Avoid hardcoded timeouts that are too short
- Ensure proper cleanup between test cases
- Check for environment variable pollution

### Coverage Gaps
- Run `go tool cover -html=coverage.out` to visualize
- Focus on untested error paths
- Add edge case scenarios

## References

- [Go Testing Package](https://pkg.go.dev/testing)
- [Table Driven Tests](https://github.com/golang/go/wiki/TableDrivenTests)
- [Go Testing Best Practices](https://go.dev/doc/effective_go#testing)
- [httptest Package](https://pkg.go.dev/net/http/httptest)
