#!/bin/bash

# Ustawienia (dopasowane do Twojego kodu w main.go)
CACHE_DIR="$HOME/.cache/eclat"
DB_FILE="$CACHE_DIR/db/assets.db"
THUMBS_DIR="$CACHE_DIR/thumbnails"

echo "☢️  ROZPOCZYNAM CZYSZCZENIE ECLAT ☢️"

# 1. Usuń bazę danych
if [ -f "$DB_FILE" ]; then
    rm "$DB_FILE"
    echo "✅ Baza danych usunięta: $DB_FILE"
else
    echo "⚠️  Baza danych nie istnieje (nic do zrobienia)."
fi
if [ -d "$THUMBS_DIR" ]; then
    rm -rf "$THUMBS_DIR"/*
    echo "✅ Miniatury wyczyszczone."
fi

echo "✨ Gotowe! Uruchom 'wails dev' aby stworzyć czystą bazę."
