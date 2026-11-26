package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Mahathirrr/kredit-jkl/internal/config"
	"github.com/Mahathirrr/kredit-jkl/internal/domain"
	"github.com/Mahathirrr/kredit-jkl/internal/handler"
	"github.com/Mahathirrr/kredit-jkl/internal/middleware"
	"github.com/Mahathirrr/kredit-jkl/internal/repository"
	"github.com/Mahathirrr/kredit-jkl/internal/service"
	_ "github.com/go-sql-driver/mysql"
)

func main() {
	cfg := config.Load()
	uploadDir, err := filepath.Abs(cfg.UploadDir)
	if err != nil {
		log.Fatal(err)
	}
	if err := os.MkdirAll(uploadDir, 0o755); err != nil {
		log.Fatal(err)
	}

	db, err := sql.Open("mysql", cfg.DatabaseDSN)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	db.SetConnMaxLifetime(5 * time.Minute)
	db.SetMaxOpenConns(10)
	if err := db.Ping(); err != nil {
		log.Fatal(err)
	}

	if err := runMigrations(db, migrationPath()); err != nil {
		log.Fatal(err)
	}

	store := repository.Store{DB: db}
	if err := seedUsers(context.Background(), store); err != nil {
		log.Fatal(err)
	}

	authSvc := service.AuthService{Users: store, JWTSecret: cfg.JWTSecret}
	pengajuanSvc := service.PengajuanService{Repo: store}
	authMiddleware := middleware.AuthMiddleware{Auth: authSvc, Users: store}
	api := handler.API{
		Auth:      authSvc,
		Pengajuan: pengajuanSvc,
		Users:     store,
		Repo:      store,
		UploadDir: uploadDir,
	}

	mux := http.NewServeMux()
	api.Routes(mux, authMiddleware)
	server := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           withCORS(mux),
		ReadHeaderTimeout: 10 * time.Second,
	}

	log.Printf("API berjalan di http://localhost%s", cfg.HTTPAddr)
	log.Fatal(server.ListenAndServe())
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, OPTIONS")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func migrationPath() string {
	if p := os.Getenv("MIGRATION_FILE"); p != "" {
		return p
	}
	cwd, _ := os.Getwd()
	if strings.HasSuffix(cwd, string(filepath.Separator)+"backend") {
		return filepath.Join(cwd, "..", "database", "migrations", "001_init.sql")
	}
	return filepath.Join(cwd, "database", "migrations", "001_init.sql")
}

func runMigrations(db *sql.DB, path string) error {
	sqlPath, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	raw, err := os.ReadFile(sqlPath)
	if err != nil {
		return err
	}
	parts := strings.Split(string(raw), ";")
	for _, part := range parts {
		stmt := strings.TrimSpace(part)
		if stmt == "" {
			continue
		}
		if _, err := db.Exec(stmt); err != nil {
			msg := err.Error()
			// Index sudah ada kalau server dijalankan ulang.
			if strings.Contains(msg, "Duplicate key name") || strings.Contains(msg, "already exists") {
				continue
			}
			return fmt.Errorf("migration gagal: %w", err)
		}
	}
	return nil
}

func seedUsers(ctx context.Context, store repository.Store) error {
	hash, err := service.HashPassword("password123")
	if err != nil {
		return err
	}
	return store.CreateDemoUsersIfEmpty(ctx, []domain.User{
		{Username: "dealer1", PasswordHash: hash, Role: domain.RoleDealer, Nama: "Sales Dealer A"},
		{Username: "marketing1", PasswordHash: hash, Role: domain.RoleMarketing, Nama: "Marketing BCA"},
		{Username: "supervisor1", PasswordHash: hash, Role: domain.RoleSupervisor, Nama: "Atasan Marketing"},
		{Username: "admin1", PasswordHash: hash, Role: domain.RoleAdmin, Nama: "Admin Backoffice"},
	})
}