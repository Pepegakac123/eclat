# 🚀 Frontend Features - Backlog

## 1. Saved Searches (Smart Collections) 🧠
**Cel:** Pozwolenie użytkownikowi na zapisywanie aktualnych filtrów jako "Kolekcji".
**Priorytet:** High (Killer Feature)

- [ ] **UI:** Dodać przycisk "Save Search" (ikonka dyskietki/bookmark) w nagłówku sekcji Filtrów (obok "Filters").
- [ ] **Modal:** Po kliknięciu modal z inputem na nazwę wyszukiwania (np. "Czerwone Modele 3D").
- [ ] **API Integration:**
    - `POST /api/saved-searches` -> Wysyła obecny obiekt `filters` (JSON).
    - `GET /api/saved-searches` -> Pobiera listę do wyświetlenia w Sidebarze.
- [ ] **Sidebar:** Nowa sekcja "Saved Searches" nad lub pod "Collections". Kliknięcie ładuje filtry do Store.

## 2. Top Toolbar & Chips Sync 🔍
**Cel:** Synchronizacja paska wyszukiwania i filtrów, lepszy feedback wizualny.
**Priorytet:** Medium (UX Polish)

- [ ] **Chips (Tagi) na górze:**
    - Wyświetlanie aktywnych filtrów (np. "Rating: 4+", "Color: #F00") jako usuwalnych "Chipsów" pod Top Toolbarem.
    - Kliknięcie 'X' na chipsie usuwa konkretny filtr ze Store.
- [ ] **Search Bar behavior:**
    - Wpisanie tekstu w SearchBar powinno albo resetować inne filtry, albo działać addytywnie (decyzja UX).
- [ ] **Clear All:** Przycisk "Clear All" widoczny, gdy cokolwiek jest pofiltrowane.

## 3. Inspector Panel Implementation (The Right Sidebar) 🕵️‍♂️
**Cel:** Stworzenie panelu "Inspector" z podziałem na tryb Single i Multi-Select.
**Lokalizacja:** `src/features/inspector/components/InspectorPanel.tsx`

- [X] **Krok 1: Layout & Single Mode Skeleton**
    - [] Struktura Flexbox: Sticky Header (góra), Scrollable Content (środek), Sticky Footer (dół).
    - [X] Mockowanie danych na podstawie `AssetDetailsDto` (żeby widzieć UI bez API).
- [ ] **Krok 2: Header Logic (Thumbnail & Title)**
    - [X] Wyświetlanie nazwy pliku (Input editable).
    - [X] Ścieżka pliku pod nazwą + przycisk "Copy Path" (do schowka).
    - [ ] **Thumbnail Hover UX:**
        - [ ] Dla obrazków: Ikonka Lupy (otwiera modal podglądu).
        - [ ] Dla 3D/Innych: Ikonka Ołówka (upload custom thumbnail).
- [ ] **Krok 3: Core Editor (Scrollable Area)**
    - [X] **Rating:** Interaktywne gwiazdki (1-5).
    - [X] **Description:** Textarea z `auto-save on blur`.
    - [X] **Tags Area:** Input typu "Chips" + lista tagów.
- [ ] **Krok 4: Metadata Tabs**
    - [X] Implementacja Tabs: "Details", "Versions", "Collections".
    - [X] **Details Tab:** Grid wyświetlający techniczne dane (FileSize, Dimensions, BitDepth, Alpha, FileHash).
    - [X] **Versions Tab:** Placeholder na listę wersji (Faza 8).
- [ ] **Krok 5: Sticky Footer Actions**
    - [ ] Przyciski: Open File, Open Explorer, Favorite (Heart).
    - [ ] Kolekcje: "Add to Collection" (zawsze) i "Remove from Collection" (tylko gdy jesteśmy w widoku kolekcji).
- [ ] **Krok 6: Multi-Select (Batch Mode) 📦**
    - [ ] Wykrywanie zaznaczenia > 1 elementu.
    - [ ] UI dla Batch Actions:
        - [ ] "Add to Collection" (dla wszystkich).
        - [ ] "Tagging" (dodaj tag do wszystkich).
        - [ ] "Rating" (ustaw ocenę dla wszystkich).
