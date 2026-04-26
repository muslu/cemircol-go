#!/bin/bash
set -e

# Versiyon bilgisi (isteğe bağlı)
VERSION=$(grep '^version' Cargo.toml | head -n1 | cut -d'"' -f2)

echo "📤 CemirCol-Go Yayınlanıyor (v$VERSION)..."

# 1. Temizlik
echo "🧹 Gereksiz dosyalar temizleniyor..."
git rm -r --cached target/ main 2>/dev/null || true

# 2. Git İşlemleri
echo "📝 Dosyalar ekleniyor ve commit ediliyor..."
git add .
read -p "Commit mesajı girin (varsayılan: Update v$VERSION): " msg
msg=${msg:-"Update v$VERSION"}
git commit -m "$msg"

# 3. Push
echo "🚀 GitHub'a pushlanıyor..."
git push origin main

# 4. Tag (Opsiyonel)
read -p "Versiyon etiketi (tag) oluşturulsun mu? (y/n): " tag_choice
if [ "$tag_choice" = "y" ]; then
    git tag "v$VERSION"
    git push origin "v$VERSION"
    echo "✅ Etiket v$VERSION oluşturuldu ve pushlandı."
fi

echo "✨ Yayınlama işlemi başarıyla tamamlandı!"
