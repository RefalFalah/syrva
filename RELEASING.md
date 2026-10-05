# Syrva v0.1.0 — Release

Release preparation tidak otomatis mempublikasikan repository, tag, atau GitHub Release.

1. Repository source adalah `https://github.com/RefalFalah/syrva`, dengan module path `github.com/RefalFalah/syrva`. Jika namespace berubah, ubah `go.mod` dan import internal sebelum publikasi.
2. Perbarui `cmd.Version`, README, dan catatan release jika versi berubah.
3. Jalankan `go fmt ./...`, `go vet ./...`, `go build ./...`, dan `go test ./...`. Pastikan CI tiga OS lulus; jalankan race detector pada lingkungan yang mendukung.
4. Smoke test pada terminal Windows/Linux: config kosong, add/edit/remove, fuzzy search, resize, dan keluar dari TUI. Dengan server uji milik sendiri, periksa SSH interaktif, preset, upload/download, tunnel, dan Ctrl+C.
5. Dari root repository, jalankan:

   ```bash
   go run ./scripts/release.go
   ```

   Script membuat binary tanpa CGO untuk Windows/Linux/macOS, masing-masing amd64 dan arm64. `dist/` berisi ZIP Windows, TAR.GZ Unix, README, LICENSE, `THIRD_PARTY_NOTICES.txt` (lisensi asli dependency dan Go runtime), contoh YAML, dan `checksums.txt`. Tidak ada config/key user yang dikemas.

6. Periksa SHA-256, isi arsip, dan `syrva --version`. Cantumkan prasyarat OpenSSH 9+/SFTP serta batasan import/preset remote POSIX pada release notes.
7. Setelah seluruh pemeriksaan lulus, maintainer dapat membuat tag `v0.1.0` dan menerbitkan arsip/checksums pada repository yang benar.
