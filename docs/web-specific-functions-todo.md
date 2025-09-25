# Web-Specific Functions TODO

This document tracks web-specific PHP functions that are not currently implemented in Hey-Codex. These functions account for the 7% gap in our 93.1% function compatibility metric (5 missing out of 72 common PHP functions).

## Missing Functions (5/72)

### 1. header()
- **Purpose**: Send HTTP header to client
- **Category**: HTTP Response Management
- **Priority**: Low (web server specific)
- **Implementation Notes**: Would require HTTP context and response handling

### 2. setcookie()
- **Purpose**: Set HTTP cookie values
- **Category**: Cookie Management
- **Priority**: Low (web server specific)
- **Implementation Notes**: Requires HTTP response header manipulation

### 3. session_start()
- **Purpose**: Initialize PHP session handling
- **Category**: Session Management
- **Priority**: Low (web server specific)
- **Implementation Notes**: Would need session storage and state management

### 4. mysqli_connect()
- **Purpose**: Connect to MySQL database
- **Category**: Database Connectivity
- **Priority**: Medium (could be useful for CLI scripts)
- **Implementation Notes**: Requires MySQL driver integration

### 5. curl_init()
- **Purpose**: Initialize HTTP client request
- **Category**: HTTP Client
- **Priority**: Medium (useful for API interactions)
- **Implementation Notes**: HTTP client library integration needed

## Implementation Strategy

These functions are intentionally excluded because Hey-Codex is designed as a general-purpose PHP interpreter focused on:
- Core language features
- String/array manipulation
- File I/O operations
- Mathematical functions
- Control flow and OOP

Web-specific functionality would require:
- HTTP server context
- Database drivers
- Session storage mechanisms
- Cookie handling infrastructure

## Current Status

**Implemented**: 67/72 common PHP functions (93.1% compatibility)
**Focus**: Core language interpreter rather than web server runtime
**Next Priority**: Complete remaining string functions and improve PHP 8.0+ feature support