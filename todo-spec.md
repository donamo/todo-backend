# AI-alapú Projekt- és Feladatkezelő Rendszer

## Funkcionális specifikáció

### Verzió

v1.0

### Cél

Egy személyes használatra optimalizált projekt- és feladatkezelő rendszer létrehozása, amely ötvözi a klasszikus todo alkalmazások egyszerűségét és a projektmenedzsment rendszerek strukturáltságát.

A rendszer fő célja:

* személyes feladatok kezelése
* otthoni projektek kezelése
* fejlesztési projektek kezelése
* hosszabb távú célok kezelése
* AI segítségével projektek és feladatok generálása
* AI segítségével projekttervezés támogatása

---

# Alapelvek

* Gyors adatbevitel
* Egyszerű használat
* Projektközpontú működés
* AI támogatott tervezés
* Dátum alapú tervezés
* Nincs óra-perc kezelés
* Nincs klasszikus naptár funkció

---

# Hierarchia

```text
Epic
 └─ Project
     └─ Stage
         └─ Todo
```

A Todo közvetlenül Projecthez is kapcsolódhat.

```text
Project
 ├─ Stage
 │   └─ Todo
 │
 └─ Todo
```

---

# Objektumok

## Epic

Legfelső szintű gyűjtő.

### Példák

* Munka
* Otthon
* Egészség
* Fejlesztés

### Mezők

* id
* név
* leírás
* szín
* sorrend
* létrehozás dátuma
* módosítás dátuma

---

## Project

Konkrét cél vagy kezdeményezés.

### Példák

* RSS Reader
* Nappali festése
* Akvárium indítása

### Mezők

* id
* epic
* név
* leírás
* státusz
* kezdési dátum
* cél dátum
* sorrend
* létrehozás dátuma
* módosítás dátuma

### Státuszok

* aktív
* szüneteltetett
* lezárt
* archivált

---

## Stage

Projekten belüli végrehajtási vagy szállítási szakasz.

### Példák

* Előkészítés
* Bontás
* Fejlesztés
* Tesztelés
* Festés
* Takarítás
* Deploy

### Mezők

* id
* project
* név
* leírás
* státusz
* kezdési dátum
* cél dátum
* sorrend
* létrehozás dátuma
* módosítás dátuma

### Státuszok

* tervezett
* folyamatban
* kész

---

## Todo

Konkrét végrehajtandó feladat.

### Példák

* Festék vásárlása
* PostgreSQL telepítése
* Fal glettelése
* Backend API elkészítése

### Mezők

* id
* project
* stage
* cím
* leírás
* prioritás
* státusz
* label-ek
* kezdési dátum
* határidő dátum
* becsült ráfordítás
* sorrend
* next_action
* milestone
* elkészült dátuma
* létrehozás dátuma
* módosítás dátuma

### Prioritások

* alacsony
* normál
* magas
* kritikus

### Státuszok

* nyitott
* folyamatban
* kész
* blokkolt

---

## Label

Szabadon definiálható címke.

### Példák

* backend
* frontend
* AI
* vásárlás
* telefon
* fontos

### Mezők

* id
* név
* szín

---

# Dátumkezelés

A rendszer kizárólag dátumokat kezel.

### Támogatott

* 2026-06-10
* 2026-06-15

### Nem támogatott

* 2026-06-10 14:30
* 2026-06-10 16:00

A rendszer nem naptár alkalmazás.

---

# Nézetek

## Inbox

Gyorsan rögzített elemek.

Még nem kerültek projektbe vagy stage-be.

---

## Dashboard

Megjeleníti:

* nyitott feladatok
* lejárt feladatok
* következő teendők
* aktív projektek
* közelgő határidők

---

## Következő teendők

Minden projektből a kijelölt következő feladat.

---

## Minden nyitott

Minden nyitott vagy folyamatban lévő feladat.

---

## Lejárt

Határidőn túli feladatok.

---

## Projektek

Projekt alapú nézet.

---

## Epic-ek

Epic alapú nézet.

---

## Stage-ek

Stage alapú nézet.

---

## Kész feladatok

Archivált feladatok.

---

# Rendezés

A következő rendezések támogatottak:

* kézi sorrend
* prioritás szerint
* határidő szerint
* létrehozás szerint
* módosítás szerint

A kézi sorrend drag & drop segítségével módosítható.

---

# Szűrés

Szűrhető:

* Epic
* Project
* Stage
* Label
* Prioritás
* Státusz
* Dátum

---

# Projekt előrehaladás

Automatikusan számított érték.

## Project progress

Példa:

```text
RSS Reader

80%
████████░░
```

## Stage progress

Példa:

```text
Backend

65%
██████░░░░
```

---

# Következő teendő (Next Action)

Egy Todo kijelölhető aktuális következő feladatként.

Minden projektben egyszerre legfeljebb egy Next Action lehet.

A Dashboard külön listában mutatja.

---

# Mérföldkövek (Milestone)

Todo jelölhető mérföldkőként.

Példák:

* Első működő verzió
* Festés kész
* Költözés kész

---

# Projekt jegyzetek

Minden projekthez tartozhat jegyzetgyűjtemény.

Tartalmazhat:

* szabad szöveget
* URL-eket
* ötleteket
* technikai dokumentációt
* bevásárló listákat

---

# Projekt sablonok

## Lakásfestés

Stage-ek:

* Előkészítés
* Bontás
* Glettelés
* Festés
* Takarítás

## Költözés

Stage-ek:

* Tervezés
* Csomagolás
* Szállítás
* Kipakolás

## Backend fejlesztés

Stage-ek:

* Tervezés
* Repository
* Fejlesztés
* Tesztelés
* Deploy

---

# Ismétlődő feladatok

Támogatott ismétlődések:

* napi
* heti
* havi
* éves

---

# Függőségek

Todo függhet más Todo-tól.

Példa:

```text
Deploy

függ:
- Tesztelés
- Release elkészítése
```

A rendszer automatikusan blokkoltnak jelölheti.

---

# AI Funkciók

## AI Inbox

Szabad szöveges bevitel.

Példa:

"Ki kell festeni a nappalit."

---

## AI Projekt Generálás

Input:

"Go backend PostgreSQL-lel"

AI javasol:

* Epic
* Project
* Stage-ek
* Todo-k

---

## AI Projekt Bontás

Meglévő projektből:

* Stage-ek
* Todo-k

generálása.

---

## AI Inbox Rendezés

Inbox elemek automatikus besorolása.

---

## AI Státusz Jelentés

Projekt állapotának automatikus összefoglalása.

Példa:

Kész:

* Backend API
* PostgreSQL

Folyamatban:

* Frontend

Hátra van:

* AI integráció

---

## AI Következő Lépések

AI javaslatok:

* következő feladatok
* hiányzó feladatok
* potenciális kockázatok

---

# Nem cél funkciók

A rendszer első verziójában nem támogatott:

* Óra-perc alapú időpontok
* Naptár alkalmazás
* Meeting kezelés
* Chat rendszer
* Többfelhasználós együttműködés
* Erőforrás tervezés
* Gantt diagram
* Kanban board
* Time tracking

---

# Termékvízió

A rendszer egy AI által támogatott személyes projekt- és feladatkezelő alkalmazás, amely a klasszikus todo listák egyszerűségét kombinálja a strukturált projektmenedzsmenttel.

A központi modell:

Epic → Project → Stage → Todo

A felhasználó képes:

* gyorsan ötleteket rögzíteni
* projekteket strukturálni
* végrehajtási terveket készíteni
* AI segítségével projektet generálni
* AI segítségével feladatokat létrehozni
* AI segítségével állapotjelentéseket készíteni

miközben a rendszer egyszerű, áttekinthető és személyes használatra optimalizált marad.
