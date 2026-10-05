# Syrva

**Syrva — terminal-first server management.**

Syrva adalah server manager open-source berbasis terminal yang ringan untuk mengelola SSH host, remote command, transfer file, dan SSH tunnel.

## Kenapa Syrva dibuat

Workflow SSH sehari-hari tidak perlu dashboard, akun, atau service tambahan. Syrva menawarkan alternatif terminal yang kecil, cepat, keyboard-first, dan developer-friendly untuk workflow dasar aplikasi seperti Termius. SSH tetap dikerjakan oleh OpenSSH, bukan implementasi protokol baru.

## Fitur v0.1.0

- Server selector Bubble Tea: navigasi keyboard, fuzzy search, nama, host, user, dan tags.
- Quick connect dengan session SSH interaktif.
- CRUD host, list rapi, dan info konfigurasi lokal.
- Preset remote command dengan working directory.
- Upload/download file memakai executable `scp`.
- Local SSH port forwarding yang berjalan di foreground.
- Import dasar `~/.ssh/config` dengan preview dan pemilihan host.
- YAML lokal; mendukung client Windows, Linux, dan macOS. Prioritas pengujian: Windows/Linux.

Stack: Go, Cobra, Bubble Tea, Bubbles, Lip Gloss, YAML, dan native OpenSSH.

## Installation

**Prasyarat:** `ssh` dan `scp` tersedia di `PATH`. Gunakan OpenSSH **9+**; transfer memakai mode SFTP dari executable `scp` (minimal OpenSSH 8.7), sehingga server perlu mengaktifkan subsystem SFTP. Tidak ada fallback ke protokol SCP legacy/remote-shell.

Instal **Go 1.24.2+**, lalu:

```bash
go install github.com/RefalFalah/syrva@latest
```

Atau bangun dari source:

```bash
git clone https://github.com/RefalFalah/syrva.git
cd syrva
go mod download
go test ./...
go install .
```

Tambahkan direktori `GOBIN` (atau `$(go env GOPATH)/bin` jika `GOBIN` kosong) ke `PATH`. Di PowerShell, periksa dengan `go env GOPATH` lalu tambahkan direktori `bin` pada PATH user.

Atau buat executable lokal:

```powershell
# Windows
go build -trimpath -o syrva.exe .
.\syrva.exe --version
```

```bash
# Linux/macOS
go build -trimpath -o syrva .
./syrva --version
```

Arsip distribusi dapat dibuat dengan `go run ./scripts/release.go`; hasilnya ada di `dist/`, termasuk `THIRD_PARTY_NOTICES.txt` berisi lisensi dependency. Binary hasil build tidak membutuhkan instalasi Go. Proyek ini belum menerbitkan release publik; lihat [RELEASING.md](RELEASING.md).

## Quick Start

```bash
syrva add
syrva list
syrva websku-dev info
syrva websku-dev
syrva
```

`add` meminta alias, nama, hostname/IP, username, port, path SSH key, tags, dan working directory. Alias, hostname, dan username wajib; default port adalah `22`. Key boleh kosong untuk memakai agent/konfigurasi OpenSSH.

**Keyboard TUI:** ↑/↓ atau `j`/`k` untuk navigasi, `/` untuk fuzzy search, Enter untuk connect, A untuk add, Q/Esc untuk keluar. Saat mencari, huruf menjadi input pencarian (termasuk Q/A/J/K); ↑/↓ tetap memilih server. Tab memindahkan fokus antara search dan navigasi; Esc selalu keluar. Ukuran terminal kecil mendapat layout ringkas.

## Command

| Command | Fungsi |
| --- | --- |
| `syrva` | Buka TUI server selector |
| `syrva list` | Daftar host lokal |
| `syrva add` | Tambah host interaktif |
| `syrva edit <host>` | Edit host; Enter mempertahankan value lama |
| `syrva remove <host>` | Hapus setelah konfirmasi `[y/N]` |
| `syrva import` | Preview/pilih host dari `~/.ssh/config`, lalu konfirmasi `[Y/n]` |
| `syrva <host>` | Langsung connect, tanpa konfirmasi Syrva |
| `syrva <host> info` | Baca info lokal, tanpa SSH |
| `syrva <host> run` | Daftar preset, tanpa SSH |
| `syrva <host> run <preset>` | Jalankan preset remote |
| `syrva <host> upload <local> <remote>` | Upload satu file |
| `syrva <host> download <remote> <local>` | Download file |
| `syrva <host> tunnel <name>` | Jalankan tunnel sampai Ctrl+C |
| `syrva --help` / `syrva --version` | Bantuan / versi |

```bash
syrva websku-dev run logs
syrva websku-dev upload ./backup.sql /tmp/
syrva websku-dev download /var/log/nginx/error.log .
syrva websku-dev tunnel mysql
syrva import --file ./ssh-config
syrva --config-dir ./local-config list
```

Quote path berspasi sesuai shell Anda, misalnya `syrva websku-dev upload "./My Files/backup.sql" "/tmp/My Files/"`. Path lokal Windows (`C:\...`, UNC) dan `~/` didukung. Gunakan `--` sebelum argumen jika nama file dimulai dengan `-`. MVP tidak menyediakan transfer direktori rekursif.

Saat edit, ketik `-` untuk mengosongkan field opsional. Rename alias tidak menimpa host lain. Preset/tunnel dipertahankan ketika mengedit host; definisinya diedit langsung di YAML.

## Config Example

Lokasi menggunakan `os.UserConfigDir()`:

| Platform | Direktori default |
| --- | --- |
| Windows | `%AppData%\syrva\` |
| Linux | `$XDG_CONFIG_HOME/syrva/`, atau `~/.config/syrva/` |
| macOS | `~/Library/Application Support/syrva/` |

Direktori, `config.yaml`, dan `hosts.yaml` dibuat otomatis saat konfigurasi dibutuhkan. Flag global `--config-dir <directory>` mengganti lokasi default. `config.yaml` hanya berisi versi **schema**:

```yaml
version: 1
```

Jika sebelumnya memakai build dengan nama lama, salin `config.yaml` dan `hosts.yaml` dari direktori config lama ke direktori `syrva`, atau gunakan `--config-dir` untuk membacanya langsung. Config lama tidak diubah atau dihapus otomatis.

`hosts.yaml`:

```yaml
hosts:
  websku-dev:
    name: Websku Development
    host: dev.example.com
    user: root
    port: 22
    identity_file: ~/.ssh/id_ed25519
    tags: [development, websku]
    working_directory: /var/www/websku
    commands:
      logs:
        description: Laravel Logs
        command: tail -f storage/logs/laravel.log
      nginx:
        description: Status Nginx
        command: systemctl status nginx
    tunnels:
      mysql:
        local_port: 3307
        remote_host: 127.0.0.1
        remote_port: 3306
```

Contoh siap salin: [examples/hosts.example.yaml](examples/hosts.example.yaml). Config kosong (`hosts: {}`) valid. Port yang dihilangkan menjadi `22`; nilai eksplisit harus `1–65535`. Field tidak dikenal ditolak agar typo atau field credential tidak diam-diam tersimpan.

Alias: huruf/angka pertama, lalu huruf, angka, titik, `-`, atau `_`; maksimal 128 karakter. Nama command utama (`add`, `edit`, `remove`, `list`, `import`, `help`, `completion`, `version`) tidak dapat dipakai sebagai alias. Config disimpan atomik dan diurutkan; komentar/format YAML akan dinormalisasi ketika CRUD/import menyimpan. Jangan menjalankan beberapa penulis config bersamaan.

Working directory berlaku **hanya untuk preset**, bukan login SSH. Gunakan path remote absolut; `cd -- '<directory>' && <command>` mengasumsikan shell remote POSIX. Client Windows tetap dapat mengelola server Linux; working directory preset untuk shell remote Windows belum didukung.

### Import OpenSSH

Parser membaca `Host`, `HostName`, `User`, `Port`, dan satu `IdentityFile`; mendukung beberapa alias, quote, komentar, CRLF, dan nilai pertama yang cocok. Wildcards/negasi dipakai untuk defaults, tetapi **tidak dibuat menjadi server**. Tanpa `HostName`, alias menjadi hostname; tanpa `User`, username OS lokal dipakai; tanpa `Port`, default `22`.

Pilih alias dengan koma (Enter = semua host baru, `-` = batal), lalu konfirmasi. Host yang sudah ada tidak ditimpa. EOF/pilihan tidak valid tidak menyimpan perubahan.

Import ini bukan salinan seluruh perilaku OpenSSH: `Include`, `Match`, proxy, dan opsi lain tidak disalin. Syrva menghubungi **hostname** yang disimpan, sehingga opsi native yang hanya cocok dengan alias asli belum tentu berlaku. Periksa konfigurasi hasil import sebelum connect, khususnya host yang memakai jump host atau banyak identity file. Parser tidak menjalankan `Match exec` atau command lain.

## Security

- Tidak menyimpan password, isi private key, akun Syrva, atau credential vault. Hanya path key yang disimpan.
- Authentication/password/passphrase prompt tetap milik OpenSSH. SSH Agent diwariskan tanpa mengubah environment.
- `known_hosts` dan host-key checking tetap berlaku; Syrva tidak menambahkan `StrictHostKeyChecking=no` atau bypass lain.
- Executable dipanggil langsung dengan argv, bukan `cmd.exe`, PowerShell, atau shell lokal. Konfigurasi native OpenSSH milik Anda tetap dipercaya dan dapat menjalankan directive lokalnya sendiri.
- Port, alias, hostname, username, dan karakter kontrol divalidasi. Path working directory di-quote; isi preset memang merupakan command remote yang dipercaya.
- Tunnel secara eksplisit bind ke `127.0.0.1`, memakai `ExitOnForwardFailure=yes`, dan tidak berjalan sebagai daemon.
- File config baru memakai mode `0600`, direktori `0700` pada Unix. Windows memakai ACL user yang diwariskan; `chmod` bukan pengganti ACL Windows.
- Tidak ada logging credential/private key. Error YAML tidak menampilkan raw value. Exit code proses diteruskan; pembatalan Ctrl+C memakai `130`.

Simpan YAML di lokasi pribadi, gunakan key terenkripsi/agent, dan jangan memasukkan rahasia ke preset. Syrva menganggap konfigurasi lokal dan executable di `PATH` sebagai input tepercaya.

## Roadmap

v0.1.0 berfokus pada workflow dasar yang sudah tersedia di atas. Iterasi berikutnya: polish UX keyboard, pengujian terminal lebih luas, packaging/release, dan penyempurnaan import. Tidak ada dashboard, database, cloud sync, monitoring, atau implementasi SSH sendiri dalam MVP.

## Contributing

```bash
go fmt ./...
go vet ./...
go build ./...
go test ./...
# Opsional, pada lingkungan dengan compiler C:
go test -race ./...
go test ./internal/ssh -run '^$' -fuzz '^FuzzParseConfig$' -fuzztime=10s
```

Test memakai temp directory dan executable helper lokal; tidak terhubung ke server publik. CI disiapkan untuk Windows, Linux, dan macOS. Pertahankan scope kecil dan tambahkan test parser/argument builder untuk perubahan perilaku. Untuk pengujian SSH sungguhan, gunakan server uji milik Anda sendiri.

## License

MIT — lihat [LICENSE](LICENSE).
