#!/usr/bin/env bash
set -euo pipefail
BASE="http://127.0.0.1:8080/api"

login() {
  curl -sS -X POST "$BASE/auth/login" -H 'Content-Type: application/json' -d "{\"username\":\"$1\",\"password\":\"password123\"}" | tee /tmp/kredit-login-$1.json
}

token() {
  python3 - <<'PY' "$1"
import json,sys
print(json.load(open(f"/tmp/kredit-login-{sys.argv[1]}.json"))["token"])
PY
}

login dealer1
DEALER_TOKEN=$(token dealer1)
HDR=(-H "Authorization: Bearer $DEALER_TOKEN" -H 'Content-Type: application/json')

PID=$(curl -sS -X POST "$BASE/pengajuan" -H "Authorization: Bearer $DEALER_TOKEN" | python3 -c 'import sys,json;print(json.load(sys.stdin)["id"])')

curl -sS -X PUT "$BASE/pengajuan/$PID/konsumen" "${HDR[@]}" -d '{"nama":"Budi Santoso","nik":"3201012345678901"}' >/dev/null
curl -sS -X PUT "$BASE/pengajuan/$PID/kendaraan" "${HDR[@]}" -d '{"dealer":"Dealer Motor A","merk":"Honda","model":"Beat","tipe":"Street","warna":"Hitam","harga":18000000}' >/dev/null

for j in ktp spk bukti_tanda_jadi; do
  echo "dummy $j" > "/tmp/$j.pdf"
  curl -sS -X POST "$BASE/pengajuan/$PID/dokumen" -H "Authorization: Bearer $DEALER_TOKEN" -F "jenis=$j" -F "file=@/tmp/$j.pdf" >/dev/null
done

curl -sS -X POST "$BASE/pengajuan/$PID/submit-dealer" -H "Authorization: Bearer $DEALER_TOKEN" >/dev/null

login marketing1
MKT=$(token marketing1)
MH=(-H "Authorization: Bearer $MKT" -H 'Content-Type: application/json')
curl -sS -X PUT "$BASE/pengajuan/$PID/konsumen" "${MH[@]}" -d '{"nama":"Budi Santoso","nik":"3201012345678901","tanggal_lahir":"1995-05-12T00:00:00Z","status_perkawinan":"belum_menikah"}' >/dev/null
curl -sS -X PUT "$BASE/pengajuan/$PID/pinjaman" "${MH[@]}" -d '{"asuransi":"All Risk","uang_muka":3000000,"tenor_bulan":24,"angsuran_bulanan":750000}' >/dev/null
echo kk > /tmp/kk.pdf
curl -sS -X POST "$BASE/pengajuan/$PID/dokumen" -H "Authorization: Bearer $MKT" -F "jenis=kartu_keluarga" -F "file=@/tmp/kk.pdf" >/dev/null
curl -sS -X POST "$BASE/pengajuan/$PID/submit-marketing" -H "Authorization: Bearer $MKT" >/dev/null

login supervisor1
SUP=$(token supervisor1)
curl -sS -X POST "$BASE/pengajuan/$PID/approve" -H "Authorization: Bearer $SUP" -H 'Content-Type: application/json' -d '{"catatan":"ok"}' >/dev/null

echo kontrak > /tmp/kontrak_ttd.pdf
echo po > /tmp/po_ttd.pdf
curl -sS -X POST "$BASE/pengajuan/$PID/dokumen" -H "Authorization: Bearer $MKT" -F "jenis=kontrak_ttd" -F "file=@/tmp/kontrak_ttd.pdf" >/dev/null
curl -sS -X POST "$BASE/pengajuan/$PID/dokumen" -H "Authorization: Bearer $MKT" -F "jenis=po_ttd" -F "file=@/tmp/po_ttd.pdf" >/dev/null
curl -sS -X POST "$BASE/pengajuan/$PID/mark-signed" -H "Authorization: Bearer $MKT" >/dev/null

login admin1
ADM=$(token admin1)
curl -sS -X POST "$BASE/pengajuan/$PID/disburse" -H "Authorization: Bearer $ADM" -H 'Content-Type: application/json' -d '{"nominal":15000000,"tanggal_cair":"2025-11-26","catatan":"cair"}' >/dev/null

STATUS=$(curl -sS "$BASE/pengajuan/$PID" -H "Authorization: Bearer $ADM" | python3 -c 'import sys,json;print(json.load(sys.stdin)["status"])')
if [ "$STATUS" != "dana_cair" ]; then
  echo "status akhir salah: $STATUS"
  exit 1
fi
echo "flow ok, status=$STATUS"