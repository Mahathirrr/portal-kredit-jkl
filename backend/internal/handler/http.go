package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/Mahathirrr/kredit-jkl/internal/domain"
	"github.com/Mahathirrr/kredit-jkl/internal/middleware"
	"github.com/Mahathirrr/kredit-jkl/internal/repository"
	"github.com/Mahathirrr/kredit-jkl/internal/service"
)

type API struct {
	Auth      service.AuthService
	Pengajuan service.PengajuanService
	Users     repository.UserRepository
	Repo      repository.PengajuanRepository
	UploadDir string
}

func (a API) Routes(mux *http.ServeMux, auth middleware.AuthMiddleware) {
	mux.HandleFunc("OPTIONS /", a.cors)
	mux.HandleFunc("POST /api/auth/login", a.login)
	mux.Handle("GET /api/auth/me", auth.Require(http.HandlerFunc(a.me)))
	mux.Handle("GET /api/pengajuan", auth.Require(http.HandlerFunc(a.listPengajuan)))
	mux.Handle("POST /api/pengajuan", auth.Require(http.HandlerFunc(a.createPengajuan)))
	mux.Handle("GET /api/pengajuan/{id}", auth.Require(http.HandlerFunc(a.getPengajuan)))
	mux.Handle("PUT /api/pengajuan/{id}/konsumen", auth.Require(http.HandlerFunc(a.updateKonsumen)))
	mux.Handle("PUT /api/pengajuan/{id}/kendaraan", auth.Require(http.HandlerFunc(a.updateKendaraan)))
	mux.Handle("PUT /api/pengajuan/{id}/pinjaman", auth.Require(http.HandlerFunc(a.updatePinjaman)))
	mux.Handle("POST /api/pengajuan/{id}/dokumen", auth.Require(http.HandlerFunc(a.uploadDokumen)))
	mux.Handle("GET /api/pengajuan/{id}/dokumen/{docId}", auth.Require(http.HandlerFunc(a.downloadDokumen)))
	mux.Handle("POST /api/pengajuan/{id}/submit-dealer", auth.Require(http.HandlerFunc(a.submitDealer)))
	mux.Handle("POST /api/pengajuan/{id}/submit-marketing", auth.Require(http.HandlerFunc(a.submitMarketing)))
	mux.Handle("POST /api/pengajuan/{id}/approve", auth.Require(http.HandlerFunc(a.approve)))
	mux.Handle("POST /api/pengajuan/{id}/reject", auth.Require(http.HandlerFunc(a.reject)))
	mux.Handle("POST /api/pengajuan/{id}/mark-signed", auth.Require(http.HandlerFunc(a.markSigned)))
	mux.Handle("POST /api/pengajuan/{id}/disburse", auth.Require(http.HandlerFunc(a.disburse)))
}

func (a API) cors(w http.ResponseWriter, r *http.Request) {
	a.setCORS(w)
	w.WriteHeader(http.StatusNoContent)
}

func (a API) setCORS(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, OPTIONS")
}

func (a API) login(w http.ResponseWriter, r *http.Request) {
	a.setCORS(w)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	var body struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "body tidak valid")
		return
	}
	token, user, err := a.Auth.Login(r.Context(), body.Username, body.Password)
	if err != nil {
		writeError(w, http.StatusUnauthorized, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"token": token, "user": user})
}

func (a API) me(w http.ResponseWriter, r *http.Request) {
	a.setCORS(w)
	user, _ := middleware.CurrentUser(r.Context())
	writeJSON(w, http.StatusOK, user)
}

func (a API) listPengajuan(w http.ResponseWriter, r *http.Request) {
	a.setCORS(w)
	user, _ := middleware.CurrentUser(r.Context())
	items, err := a.Pengajuan.ListForUser(r.Context(), user)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, items)
}

func (a API) createPengajuan(w http.ResponseWriter, r *http.Request) {
	a.setCORS(w)
	user, _ := middleware.CurrentUser(r.Context())
	if user.Role != domain.RoleDealer {
		writeError(w, http.StatusForbidden, "hanya dealer yang boleh membuat draft")
		return
	}
	p, err := a.Pengajuan.CreateDraft(r.Context(), user)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, p)
}

func (a API) getPengajuan(w http.ResponseWriter, r *http.Request) {
	a.setCORS(w)
	user, _ := middleware.CurrentUser(r.Context())
	id, err := idFromPath(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "id tidak valid")
		return
	}
	p, err := a.Pengajuan.Get(r.Context(), user, id)
	if err != nil {
		writeError(w, http.StatusForbidden, err.Error())
		return
	}
	if p == nil {
		writeError(w, http.StatusNotFound, "pengajuan tidak ditemukan")
		return
	}
	writeJSON(w, http.StatusOK, p)
}

func (a API) updateKonsumen(w http.ResponseWriter, r *http.Request) {
	a.setCORS(w)
	user, _ := middleware.CurrentUser(r.Context())
	id, err := idFromPath(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "id tidak valid")
		return
	}
	var body domain.Konsumen
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "body tidak valid")
		return
	}
	body.PengajuanID = id
	if body.TanggalLahir != nil && body.TanggalLahir.IsZero() {
		body.TanggalLahir = nil
	}
	if err := a.Pengajuan.UpdateKonsumen(r.Context(), user, body); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "data konsumen tersimpan"})
}

func (a API) updateKendaraan(w http.ResponseWriter, r *http.Request) {
	a.setCORS(w)
	user, _ := middleware.CurrentUser(r.Context())
	id, err := idFromPath(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "id tidak valid")
		return
	}
	var body domain.Kendaraan
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "body tidak valid")
		return
	}
	body.PengajuanID = id
	if err := a.Pengajuan.UpdateKendaraan(r.Context(), user, body); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "data kendaraan tersimpan"})
}

func (a API) updatePinjaman(w http.ResponseWriter, r *http.Request) {
	a.setCORS(w)
	user, _ := middleware.CurrentUser(r.Context())
	id, err := idFromPath(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "id tidak valid")
		return
	}
	var body domain.Pinjaman
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "body tidak valid")
		return
	}
	body.PengajuanID = id
	if err := a.Pengajuan.UpdatePinjaman(r.Context(), user, body); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "data pinjaman tersimpan"})
}

func (a API) uploadDokumen(w http.ResponseWriter, r *http.Request) {
	a.setCORS(w)
	user, _ := middleware.CurrentUser(r.Context())
	id, err := idFromPath(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "id tidak valid")
		return
	}
	if err := r.ParseMultipartForm(12 << 20); err != nil {
		writeError(w, http.StatusBadRequest, "form upload tidak valid")
		return
	}
	jenis := strings.TrimSpace(r.FormValue("jenis"))
	file, header, err := r.FormFile("file")
	if err != nil {
		writeError(w, http.StatusBadRequest, "file wajib diunggah")
		return
	}
	defer file.Close()
	if jenis == "" {
		writeError(w, http.StatusBadRequest, "jenis dokumen wajib diisi")
		return
	}
	if header.Size > 5<<20 {
		writeError(w, http.StatusBadRequest, "ukuran file maksimal 5 MB")
		return
	}
	ext := strings.ToLower(filepath.Ext(header.Filename))
	switch ext {
	case ".pdf", ".jpg", ".jpeg", ".png":
	default:
		writeError(w, http.StatusBadRequest, "file harus pdf, jpg, atau png")
		return
	}
	if err := os.MkdirAll(a.UploadDir, 0o755); err != nil {
		writeError(w, http.StatusInternalServerError, "gagal menyiapkan folder upload")
		return
	}
	safeName := fmt.Sprintf("%d_%s_%d%s", id, jenis, time.Now().UnixNano(), filepath.Ext(header.Filename))
	target := filepath.Join(a.UploadDir, safeName)
	out, err := os.Create(target)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "gagal menyimpan file")
		return
	}
	defer out.Close()
	if _, err := io.Copy(out, file); err != nil {
		writeError(w, http.StatusInternalServerError, "gagal menulis file")
		return
	}
	doc := domain.Dokumen{
		PengajuanID: id,
		Jenis:       jenis,
		NamaFile:    header.Filename,
		PathFile:    target,
		UploadedBy:  user.ID,
	}
	if err := a.Pengajuan.AddDokumen(r.Context(), user, doc); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, map[string]string{"message": "dokumen berhasil diunggah"})
}

func (a API) downloadDokumen(w http.ResponseWriter, r *http.Request) {
	a.setCORS(w)
	user, _ := middleware.CurrentUser(r.Context())
	pengajuanID, err := idFromPath(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "id tidak valid")
		return
	}
	docID, err := strconv.ParseInt(r.PathValue("docId"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "id dokumen tidak valid")
		return
	}
	if _, err := a.Pengajuan.Get(r.Context(), user, pengajuanID); err != nil {
		writeError(w, http.StatusForbidden, err.Error())
		return
	}
	doc, err := a.Repo.FindDokumen(r.Context(), docID)
	if err != nil || doc == nil || doc.PengajuanID != pengajuanID {
		writeError(w, http.StatusNotFound, "dokumen tidak ditemukan")
		return
	}
	safe := strings.ReplaceAll(doc.NamaFile, `"`, "")
	w.Header().Set("Content-Disposition", "attachment; filename=\""+safe+"\"")
	http.ServeFile(w, r, doc.PathFile)
}

func (a API) submitDealer(w http.ResponseWriter, r *http.Request) {
	a.actionTransition(w, r, func(user domain.User, id int64) error {
		return a.Pengajuan.SubmitDealer(r.Context(), user, id)
	})
}

func (a API) submitMarketing(w http.ResponseWriter, r *http.Request) {
	a.actionTransition(w, r, func(user domain.User, id int64) error {
		return a.Pengajuan.SubmitMarketing(r.Context(), user, id)
	})
}

func (a API) approve(w http.ResponseWriter, r *http.Request) {
	a.actionWithNote(w, r, func(user domain.User, id int64, note string) error {
		return a.Pengajuan.Approve(r.Context(), user, id, note)
	})
}

func (a API) reject(w http.ResponseWriter, r *http.Request) {
	a.actionWithNote(w, r, func(user domain.User, id int64, note string) error {
		return a.Pengajuan.Reject(r.Context(), user, id, note)
	})
}

func (a API) markSigned(w http.ResponseWriter, r *http.Request) {
	a.actionTransition(w, r, func(user domain.User, id int64) error {
		return a.Pengajuan.MarkSignedUploaded(r.Context(), user, id)
	})
}

func (a API) disburse(w http.ResponseWriter, r *http.Request) {
	a.setCORS(w)
	user, _ := middleware.CurrentUser(r.Context())
	id, err := idFromPath(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "id tidak valid")
		return
	}
	var body struct {
		Nominal   float64 `json:"nominal"`
		Tanggal   string  `json:"tanggal_cair"`
		Catatan   string  `json:"catatan"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "body tidak valid")
		return
	}
	tgl, err := time.Parse("2006-01-02", body.Tanggal)
	if err != nil {
		writeError(w, http.StatusBadRequest, "tanggal cair tidak valid")
		return
	}
	if err := a.Pengajuan.Disburse(r.Context(), user, id, body.Nominal, tgl, body.Catatan); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "pencairan dicatat"})
}

func (a API) actionTransition(w http.ResponseWriter, r *http.Request, fn func(domain.User, int64) error) {
	a.setCORS(w)
	user, _ := middleware.CurrentUser(r.Context())
	id, err := idFromPath(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "id tidak valid")
		return
	}
	if err := fn(user, id); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "status berhasil diperbarui"})
}

func (a API) actionWithNote(w http.ResponseWriter, r *http.Request, fn func(domain.User, int64, string) error) {
	a.setCORS(w)
	user, _ := middleware.CurrentUser(r.Context())
	id, err := idFromPath(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "id tidak valid")
		return
	}
	var body struct {
		Catatan string `json:"catatan"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	if err := fn(user, id, body.Catatan); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "keputusan tersimpan"})
}

func idFromPath(r *http.Request) (int64, error) {
	return strconv.ParseInt(r.PathValue("id"), 10, 64)
}