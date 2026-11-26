CREATE TABLE IF NOT EXISTS users (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    username VARCHAR(50) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    role ENUM('dealer', 'marketing', 'supervisor', 'admin') NOT NULL,
    nama VARCHAR(100) NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS pengajuan (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    nomor VARCHAR(30) NOT NULL UNIQUE,
    dealer_user_id BIGINT NOT NULL,
    marketing_user_id BIGINT NULL,
    status VARCHAR(40) NOT NULL,
    alasan_penolakan TEXT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    CONSTRAINT fk_pengajuan_dealer FOREIGN KEY (dealer_user_id) REFERENCES users (id),
    CONSTRAINT fk_pengajuan_marketing FOREIGN KEY (marketing_user_id) REFERENCES users (id)
);

CREATE TABLE IF NOT EXISTS data_konsumen (
    pengajuan_id BIGINT PRIMARY KEY,
    nama VARCHAR(100) NOT NULL DEFAULT '',
    nik VARCHAR(16) NOT NULL DEFAULT '',
    tanggal_lahir DATE NULL,
    status_perkawinan VARCHAR(20) NOT NULL DEFAULT '',
    nama_pasangan VARCHAR(100) NOT NULL DEFAULT '',
    nik_pasangan VARCHAR(16) NOT NULL DEFAULT '',
    CONSTRAINT fk_konsumen_pengajuan FOREIGN KEY (pengajuan_id) REFERENCES pengajuan (id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS data_kendaraan (
    pengajuan_id BIGINT PRIMARY KEY,
    dealer VARCHAR(100) NOT NULL DEFAULT '',
    merk VARCHAR(50) NOT NULL DEFAULT '',
    model VARCHAR(50) NOT NULL DEFAULT '',
    tipe VARCHAR(50) NOT NULL DEFAULT '',
    warna VARCHAR(30) NOT NULL DEFAULT '',
    harga DECIMAL(18, 2) NOT NULL DEFAULT 0,
    CONSTRAINT fk_kendaraan_pengajuan FOREIGN KEY (pengajuan_id) REFERENCES pengajuan (id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS data_pinjaman (
    pengajuan_id BIGINT PRIMARY KEY,
    asuransi VARCHAR(50) NOT NULL DEFAULT '',
    uang_muka DECIMAL(18, 2) NOT NULL DEFAULT 0,
    tenor_bulan INT NOT NULL DEFAULT 0,
    angsuran_bulanan DECIMAL(18, 2) NOT NULL DEFAULT 0,
    CONSTRAINT fk_pinjaman_pengajuan FOREIGN KEY (pengajuan_id) REFERENCES pengajuan (id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS dokumen (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    pengajuan_id BIGINT NOT NULL,
    jenis VARCHAR(40) NOT NULL,
    nama_file VARCHAR(255) NOT NULL,
    path_file VARCHAR(500) NOT NULL,
    uploaded_by BIGINT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_dokumen_pengajuan FOREIGN KEY (pengajuan_id) REFERENCES pengajuan (id) ON DELETE CASCADE,
    CONSTRAINT fk_dokumen_user FOREIGN KEY (uploaded_by) REFERENCES users (id)
);

CREATE TABLE IF NOT EXISTS riwayat_status (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    pengajuan_id BIGINT NOT NULL,
    status_lama VARCHAR(40) NOT NULL,
    status_baru VARCHAR(40) NOT NULL,
    catatan TEXT NULL,
    user_id BIGINT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_riwayat_pengajuan FOREIGN KEY (pengajuan_id) REFERENCES pengajuan (id) ON DELETE CASCADE,
    CONSTRAINT fk_riwayat_user FOREIGN KEY (user_id) REFERENCES users (id)
);

CREATE TABLE IF NOT EXISTS approval_pengajuan (
    pengajuan_id BIGINT PRIMARY KEY,
    disetujui TINYINT(1) NOT NULL,
    catatan TEXT NULL,
    approved_by BIGINT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_approval_pengajuan FOREIGN KEY (pengajuan_id) REFERENCES pengajuan (id) ON DELETE CASCADE,
    CONSTRAINT fk_approval_user FOREIGN KEY (approved_by) REFERENCES users (id)
);

CREATE TABLE IF NOT EXISTS pencairan_dana (
    pengajuan_id BIGINT PRIMARY KEY,
    nominal DECIMAL(18, 2) NOT NULL,
    tanggal_cair DATE NOT NULL,
    catatan TEXT NULL,
    admin_user_id BIGINT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_pencairan_pengajuan FOREIGN KEY (pengajuan_id) REFERENCES pengajuan (id) ON DELETE CASCADE,
    CONSTRAINT fk_pencairan_user FOREIGN KEY (admin_user_id) REFERENCES users (id)
);

CREATE INDEX idx_pengajuan_status ON pengajuan (status);
CREATE INDEX idx_pengajuan_dealer ON pengajuan (dealer_user_id);