#!/bin/bash
set -e

echo "🚀 CemirCol-Go Kurulum Başlatılıyor..."

# 1. Rust Kontrolü
if ! command -v cargo &> /dev/null; then
    echo "📦 Rust bulunamadı, kuruluyor..."
    curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh -s -- -y
    source $HOME/.cargo/env
else
    echo "✅ Rust zaten kurulu."
fi

# 2. Go Kontrolü
if ! command -v go &> /dev/null; then
    echo "📦 Go bulunamadı. Lütfen Go (1.21+) kurun veya PATH ayarlarınızı kontrol edin."
    echo "Öneri: /usr/local/go/bin dizinini PATH'e ekleyin."
else
    echo "✅ Go zaten kurulu: $(go version)"
fi

# 3. Rust Çekirdeğini Derle
echo "🛠️  Rust kütüphanesi derleniyor (release)..."
cargo build --release

# 4. Go Modüllerini Düzenle
echo "🧹 Go modülleri düzenleniyor..."
go mod tidy

echo "✨ Kurulum tamamlandı! Örnek uygulamayı çalıştırmak için: go run main.go"
