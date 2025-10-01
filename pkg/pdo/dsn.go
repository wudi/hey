package pdo

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

// ParseDSN parses a PDO DSN string into structured components
// Supports formats:
//   mysql:host=localhost;port=3306;dbname=test
//   sqlite:/path/to/database.db
//   pgsql:host=localhost;port=5432;dbname=test
func ParseDSN(dsn string) (*DSN, error) {
	parts := strings.SplitN(dsn, ":", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid DSN format: %s", dsn)
	}

	dsnInfo := &DSN{
		Driver:  parts[0],
		Options: make(map[string]string),
	}

	// SQLite uses file path format
	if dsnInfo.Driver == "sqlite" {
		dsnInfo.Database = parts[1]
		return dsnInfo, nil
	}

	// Parse key=value pairs for MySQL/PostgreSQL
	params := parts[1]
	pairs := strings.Split(params, ";")

	for _, pair := range pairs {
		kv := strings.SplitN(pair, "=", 2)
		if len(kv) != 2 {
			continue
		}

		key := strings.TrimSpace(kv[0])
		value := strings.TrimSpace(kv[1])

		switch key {
		case "host", "hostname":
			dsnInfo.Host = value
		case "port":
			port, err := strconv.Atoi(value)
			if err != nil {
				return nil, fmt.Errorf("invalid port: %s", value)
			}
			dsnInfo.Port = port
		case "dbname", "database":
			dsnInfo.Database = value
		case "user", "username":
			dsnInfo.Username = value
		case "password", "pass":
			dsnInfo.Password = value
		default:
			dsnInfo.Options[key] = value
		}
	}

	// Set default ports if not specified
	if dsnInfo.Port == 0 {
		switch dsnInfo.Driver {
		case "mysql":
			dsnInfo.Port = 3306
		case "pgsql":
			dsnInfo.Port = 5432
		}
	}

	return dsnInfo, nil
}

// BuildMySQLDSN builds a MySQL DSN string for database/sql
func BuildMySQLDSN(dsn *DSN, username, password string) string {
	// Format: user:password@tcp(host:port)/database?options
	var dsnBuilder strings.Builder

	// Username and password
	if username != "" {
		dsnBuilder.WriteString(username)
		if password != "" {
			dsnBuilder.WriteString(":")
			dsnBuilder.WriteString(password)
		}
		dsnBuilder.WriteString("@")
	}

	// Protocol and host
	dsnBuilder.WriteString("tcp(")
	if dsn.Host != "" {
		dsnBuilder.WriteString(dsn.Host)
	} else {
		dsnBuilder.WriteString("localhost")
	}
	dsnBuilder.WriteString(":")
	dsnBuilder.WriteString(strconv.Itoa(dsn.Port))
	dsnBuilder.WriteString(")/")

	// Database name
	if dsn.Database != "" {
		dsnBuilder.WriteString(dsn.Database)
	}

	// Additional options
	if len(dsn.Options) > 0 {
		dsnBuilder.WriteString("?")
		first := true
		for key, value := range dsn.Options {
			if !first {
				dsnBuilder.WriteString("&")
			}
			first = false
			dsnBuilder.WriteString(url.QueryEscape(key))
			dsnBuilder.WriteString("=")
			dsnBuilder.WriteString(url.QueryEscape(value))
		}
	}

	return dsnBuilder.String()
}

// BuildPostgreSQLDSN builds a PostgreSQL DSN string for database/sql
func BuildPostgreSQLDSN(dsn *DSN, username, password string) string {
	// Format: host=localhost port=5432 user=postgres password=secret dbname=test
	var params []string

	if dsn.Host != "" {
		params = append(params, fmt.Sprintf("host=%s", dsn.Host))
	} else {
		params = append(params, "host=localhost")
	}

	params = append(params, fmt.Sprintf("port=%d", dsn.Port))

	if username != "" {
		params = append(params, fmt.Sprintf("user=%s", username))
	}

	if password != "" {
		params = append(params, fmt.Sprintf("password=%s", password))
	}

	if dsn.Database != "" {
		params = append(params, fmt.Sprintf("dbname=%s", dsn.Database))
	}

	// Additional options
	for key, value := range dsn.Options {
		params = append(params, fmt.Sprintf("%s=%s", key, value))
	}

	// Default to sslmode=disable if not specified
	sslModeSet := false
	for _, param := range params {
		if strings.HasPrefix(param, "sslmode=") {
			sslModeSet = true
			break
		}
	}
	if !sslModeSet {
		params = append(params, "sslmode=disable")
	}

	return strings.Join(params, " ")
}

// BuildSQLiteDSN builds a SQLite DSN string
func BuildSQLiteDSN(dsn *DSN) string {
	// SQLite uses file path directly
	if dsn.Database == "" || dsn.Database == ":memory:" {
		// Use shared cache mode for :memory: databases to allow connection pooling
		// Without this, each connection gets its own separate in-memory database
		return "file::memory:?mode=memory&cache=shared"
	}
	return dsn.Database
}
