#!/bin/bash

# Sprawdzenie zależności
if ! command -v jq &> /dev/null;
then
    echo "Błąd: Program 'jq' nie jest zainstalowany. Zainstaluj go (sudo dnf install jq)."
    exit 1
fi

if ! command -v gh &> /dev/null;
then
    echo "Błąd: GitHub CLI ('gh') nie jest zainstalowany. Zainstaluj go i zaloguj się (gh auth login)."
    exit 1
fi

# 1. Odczyt wersji z wails.json
VERSION=$(jq -r '.info.productVersion' wails.json)
if [ -z "$VERSION" ] || [ "$VERSION" == "null" ];
then
    echo "Błąd: Nie można odczytać wersji z wails.json"
    exit 1
fi

TAG="v$VERSION"
NOTES_FILE="release_notes.tmp"

# 2. Ekstrakcja tekstu z CHANGELOG.md
# Logika: Szukamy nagłówka pasującego do wersji (z 'v' lub bez) i kopiujemy tekst 
# aż do napotkania kolejnego nagłówka zaczynającego się od '## ['
awk "/^## \[(v)?${VERSION}\]/{p=1;next} /^## \[/{p=0} p" CHANGELOG.md > "$NOTES_FILE"

# Usuwamy ewentualne puste linie z początku i końca
sed -i '././,$!d' "$NOTES_FILE"

# Sprawdzenie czy notatki nie są puste
if [ ! -s "$NOTES_FILE" ];
then
    echo "Błąd: Nie znaleziono notatek dla wersji $VERSION w pliku CHANGELOG.md"
    echo "Upewnij się, że nagłówek ma format: ## [v$VERSION] lub ## [$VERSION]"
    rm -f "$NOTES_FILE"
    exit 1
fi

# 3. Podgląd dla użytkownika
echo "--------------------------------------------------"
echo "ZIDENTYFIKOWANA WERSJA: $TAG"
echo "--------------------------------------------------"
echo "NOTATKI WYDANIA:"
cat "$NOTES_FILE"
echo "--------------------------------------------------"

# 4. Potwierdzenie od użytkownika
read -p "Czy chcesz wypuścić release $TAG i przesłać pliki z build/bin/? [y/N]: " CONFIRM
if [[ ! "$CONFIRM" =~ ^[yY]$ ]];
then
    echo "Anulowano."
    rm -f "$NOTES_FILE"
    exit 0
fi

# 5. Tworzenie Release na GitHub
echo "Tworzenie wydania $TAG na GitHubie..."

# Sprawdzamy czy folder build/bin istnieje i nie jest pusty
if [ ! -d "build/bin" ] || [ -z "$(ls -A build/bin)" ];
then
    echo "Ostrzeżenie: Folder build/bin/ jest pusty lub nie istnieje. Release zostanie utworzony bez załączników."
    gh release create "$TAG" \
        --title "$TAG" \
        --notes-file "$NOTES_FILE"
else
    gh release create "$TAG" \
        --title "$TAG" \
        --notes-file "$NOTES_FILE" \
        build/bin/*
fi

# 6. Porządki
rm -f "$NOTES_FILE"

echo "--------------------------------------------------"
echo "SUKCES: Wydanie $TAG zostało opublikowane!"
