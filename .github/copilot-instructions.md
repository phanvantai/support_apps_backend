# GitHub Copilot Instructions - Support App Backend

## Project Overview
Go-based RESTful API for support tickets using **Test-Driven Development (TDD)** and **Clean Architecture**.

**Tech Stack**: Go 1.24+, Gin, PostgreSQL, GORM, JWT, Testify

## TDD Development Rules

### 1. Always Follow Red-Green-Refactor Cycle
1. **Red**: Write failing test first
2. **Green**: Write minimal code to pass
3. **Refactor**: Improve while keeping tests green

### 2. Testing Requirements
- **Test Coverage**: >80%
- **Test Naming**: `TestComponent_Method_Scenario`
- **Test Types**: Unit, Integration, Repository, Handler, Middleware
- **Test Files**: `*_test.go` alongside source files

### 3. Clean Architecture Layers
```
cmd/           → Entry point
internal/
├── handlers/  → HTTP endpoints
├── services/  → Business logic
├── repositories/ → Data access
├── models/    → Data structures
└── middleware/ → Auth, rate limiting
```

### 4. Development Workflow for New Features
1. Write handler test with scenarios (success, validation errors, edge cases)
2. Create request/response structs with validation tags
3. Implement handler with proper error handling
4. Add service layer test for business logic
5. Implement service method
6. Add repository test if database needed
7. Implement repository method
8. Register route and update docs

### 5. Code Standards
- Use Gin binding tags for validation
- Return structured errors with HTTP status codes
- Use GORM with transactions for multi-step operations
- Add godoc comments for public functions
- Never log sensitive data

### 6. Testing Commands
```bash
make test              # Run all tests
make test-coverage     # Run with coverage
go test -v ./internal/handlers -run TestSpecific
```

**Remember**: Write tests first, keep them focused, test both happy path and error scenarios.