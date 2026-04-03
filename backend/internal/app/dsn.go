package app

import "strings"

func formatDSNForLog(dbType, dsn string) string {
	switch strings.ToLower(dbType) {
	case "postgres", "pg", "postgresql", "pgx":
		return maskKeyValue(dsn, "password")
	case "mysql", "mariadb":
		return maskMySQLPassword(dsn)
	default:
		return dsn
	}
}

func maskKeyValue(dsn, key string) string {
	keyLower := strings.ToLower(key)
	parts := strings.Fields(dsn)
	for i, part := range parts {
		kv := strings.SplitN(part, "=", 2)
		if len(kv) != 2 {
			continue
		}
		if strings.EqualFold(kv[0], keyLower) {
			parts[i] = kv[0] + "=****"
		}
	}
	return strings.Join(parts, " ")
}

func maskMySQLPassword(dsn string) string {
	at := strings.Index(dsn, "@")
	if at <= 0 {
		return dsn
	}

	creds := dsn[:at]
	colon := strings.Index(creds, ":")
	if colon <= 0 {
		return dsn
	}

	return creds[:colon+1] + "****" + dsn[at:]
}

func localURL(addr, path string) string {
	host := strings.TrimSpace(addr)
	if host == "" {
		host = "localhost:8080"
	}
	if strings.HasPrefix(host, ":") {
		host = "localhost" + host
	} else if strings.HasPrefix(host, "0.0.0.0:") {
		host = "localhost" + host[len("0.0.0.0"):]
	}
	if !strings.HasPrefix(host, "http://") && !strings.HasPrefix(host, "https://") {
		host = "http://" + host
	}
	return host + path
}
