package config

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type Config struct {
	HTTPAddr    string
	DatabaseDSN string
	JWTSecret   string
	UploadDir   string
}

func Load() Config {
	port := getenv("HTTP_PORT", "8080")
	upload := getenv("UPLOAD_DIR", "")
	if upload == "" {
		upload = defaultUploadDir()
	}
	return Config{
		HTTPAddr:    ":" + port,
		DatabaseDSN: getenv("DATABASE_DSN", "kredit_jkl:kredit_jkl_dev@tcp(127.0.0.1:3306)/kredit_jkl?parseTime=true&charset=utf8mb4"),
		JWTSecret:   getenv("JWT_SECRET", "ganti-di-production"),
		UploadDir:   upload,
	}
}

func defaultUploadDir() string {
	cwd, err := os.Getwd()
	if err != nil {
		return "uploads"
	}
	if strings.HasSuffix(cwd, string(os.PathSeparator)+"backend") {
		return filepath.Join(cwd, "..", "uploads")
	}
	return filepath.Join(cwd, "uploads")
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getenvInt(key string, fallback int) int {
	raw := os.Getenv(key)
	if raw == "" {
		return fallback
	}
	n, err := strconv.Atoi(raw)
	if err != nil {
		return fallback
	}
	return n
}