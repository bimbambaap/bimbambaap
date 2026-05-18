# API Documentatie

Een overzicht van alle endpoints, hoe je de API aanroept vanuit je frontend, en hoe je dat veilig doet.

---

## Basis URL

```
https://jouw-project.railway.app/api/v1
```

---

## Authenticatie

De API gebruikt **JWT tokens**. Na het inloggen of registreren krijg je een token terug. Dit token stuur je mee bij elke beveiligde route via de `Authorization` header:

```
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6...
```

Tokens zijn **30 dagen geldig**. Daarna moet de gebruiker opnieuw inloggen.

---

## Endpoints

### Auth

#### Registreren
```
POST /auth/register
```

Body:
```json
{
  "username": "pietje",
  "email": "pietje@example.com",
  "password": "minachttekens"
}
```

Regels:
- `username` — minimaal 3, maximaal 50 tekens
- `email` — moet geldig email formaat zijn
- `password` — minimaal 8 tekens

Response `201`:
```json
{
  "token": "eyJ...",
  "user": {
    "id": 1,
    "username": "pietje",
    "email": "pietje@example.com",
    "created_at": "2024-01-01T00:00:00Z"
  }
}
```

---

#### Inloggen
```
POST /auth/login
```

Body:
```json
{
  "email": "pietje@example.com",
  "password": "minachttekens"
}
```

Response `200`:
```json
{
  "token": "eyJ...",
  "user": {
    "id": 1,
    "username": "pietje",
    "email": "pietje@example.com",
    "avatar_url": "",
    "bio": ""
  }
}
```

> Let op: bij een verkeerd wachtwoord of onbekend email geeft de API altijd dezelfde foutmelding terug. Dit is bewust zodat je niet kan raden of een email bestaat.

---

### Users

#### Eigen profiel ophalen
```
GET /users/me
```
Vereist token.

Response `200`:
```json
{
  "id": 1,
  "username": "pietje",
  "email": "pietje@example.com",
  "avatar_url": "https://res.cloudinary.com/...",
  "bio": "Dit ben ik",
  "created_at": "2024-01-01T00:00:00Z"
}
```

---

#### Profiel updaten
```
PUT /users/me
```
Vereist token.

Body:
```json
{
  "bio": "Nieuwe bio",
  "avatar_url": "https://res.cloudinary.com/..."
}
```

Response `200`:
```json
{
  "message": "Profiel bijgewerkt"
}
```

---

#### Posts van een gebruiker ophalen
```
GET /users/:username/posts
```
Geen token nodig.

Voorbeeld: `GET /users/pietje/posts`

Response `200`:
```json
{
  "posts": [
    {
      "id": 1,
      "user_id": 1,
      "caption": "Mijn eerste post",
      "image_url": "https://res.cloudinary.com/...",
      "created_at": "2024-01-01T00:00:00Z"
    }
  ]
}
```

---

### Posts

#### Feed ophalen
```
GET /posts/feed
```
Geen token nodig. Ondersteunt paginering.

Query parameters:
| Parameter | Standaard | Maximaal |
|-----------|-----------|---------|
| `page` | 1 | — |
| `limit` | 20 | 50 |

Voorbeeld: `GET /posts/feed?page=2&limit=10`

Response `200`:
```json
{
  "posts": [
    {
      "id": 3,
      "user_id": 1,
      "caption": "Mooie foto",
      "image_url": "https://res.cloudinary.com/...",
      "created_at": "2024-01-02T00:00:00Z",
      "author": {
        "id": 1,
        "username": "pietje",
        "avatar_url": ""
      }
    }
  ],
  "page": 1,
  "limit": 20
}
```

---

#### Één post ophalen
```
GET /posts/:id
```
Geen token nodig.

Voorbeeld: `GET /posts/3`

Response `200` — zelfde structuur als feed maar één post.

---

#### Post aanmaken
```
POST /posts
```
Vereist token. Stuur als **multipart/form-data** (niet JSON) want je stuurt een bestand mee.

Velden:
| Veld | Type | Verplicht |
|------|------|-----------|
| `image` | bestand | Ja |
| `caption` | tekst | Nee |

Maximale bestandsgrootte: **10MB**

Response `201`:
```json
{
  "id": 4,
  "user_id": 1,
  "caption": "Mijn post",
  "image_url": "https://res.cloudinary.com/...",
  "created_at": "2024-01-03T00:00:00Z"
}
```

---

#### Post verwijderen
```
DELETE /posts/:id
```
Vereist token. Alleen de eigenaar van de post kan hem verwijderen.

Response `200`:
```json
{
  "message": "Post verwijderd"
}
```

---

## Foutmeldingen

Alle foutmeldingen hebben dit formaat:

```json
{
  "error": "Beschrijving van wat er mis ging"
}
```

| Status code | Betekenis |
|-------------|-----------|
| `400` | Verkeerde input, validatie mislukt |
| `401` | Niet ingelogd of token verlopen |
| `403` | Geen rechten voor deze actie |
| `404` | Niet gevonden |
| `409` | Conflict, bijv. email al in gebruik |
| `429` | Te veel requests gestuurd |
| `500` | Server fout |

---

## Veilig gebruiken in de frontend

### Token opslaan

Sla het token op in `localStorage`:

```javascript
// Na inloggen
localStorage.setItem('token', data.token)

// Bij uitloggen
localStorage.removeItem('token')

// Ophalen
const token = localStorage.getItem('token')
```

> Gebruik geen cookies tenzij je CSRF bescherming toevoegt aan de backend. `localStorage` is prima voor een webapp.

---

### Herbruikbare fetch functie

Maak één centrale functie zodat je niet overal het token handmatig hoeft toe te voegen:

```javascript
const API_URL = 'https://jouw-project.railway.app/api/v1'

async function apiFetch(endpoint, options = {}) {
  const token = localStorage.getItem('token')

  const headers = {
    ...options.headers,
  }

  // Voeg token toe als het er is
  if (token) {
    headers['Authorization'] = `Bearer ${token}`
  }

  // Alleen Content-Type toevoegen als het geen FormData is
  if (!(options.body instanceof FormData)) {
    headers['Content-Type'] = 'application/json'
  }

  const response = await fetch(`${API_URL}${endpoint}`, {
    ...options,
    headers,
  })

  // Token verlopen of ongeldig
  if (response.status === 401) {
    localStorage.removeItem('token')
    window.location.href = '/login' // Stuur naar loginpagina
    return
  }

  const data = await response.json()

  if (!response.ok) {
    throw new Error(data.error || 'Er ging iets mis')
  }

  return data
}
```

---

### Voorbeelden

**Inloggen:**
```javascript
async function login(email, password) {
  try {
    const data = await apiFetch('/auth/login', {
      method: 'POST',
      body: JSON.stringify({ email, password }),
    })
    localStorage.setItem('token', data.token)
    return data.user
  } catch (err) {
    console.error('Login mislukt:', err.message)
  }
}
```

**Feed ophalen:**
```javascript
async function getFeed(page = 1) {
  const data = await apiFetch(`/posts/feed?page=${page}&limit=20`)
  return data.posts
}
```

**Post aanmaken met foto:**
```javascript
async function createPost(imageFile, caption) {
  const formData = new FormData()
  formData.append('image', imageFile)
  formData.append('caption', caption)

  const data = await apiFetch('/posts', {
    method: 'POST',
    body: formData,
    // Geen Content-Type header — browser doet dit automatisch voor FormData
  })
  return data
}
```

**Controleren of gebruiker ingelogd is:**
```javascript
function isLoggedIn() {
  return !!localStorage.getItem('token')
}
```

---

### Wat nooit te doen

- Sla **nooit** wachtwoorden op in de frontend, ook niet tijdelijk
- Zet **nooit** je `JWT_SECRET` of Cloudinary API secret in je frontend code
- Stuur **nooit** het token mee in de URL als query parameter
- Valideer input **altijd** ook in de frontend, ook al doet de backend het ook

---

## Rate limiting

De API staat maximaal **60 requests per minuut** per IP toe. Daarna krijg je een `429` terug. Bouw dit in je frontend af door foutmeldingen netjes te tonen en niet automatisch te blijven retrien.
