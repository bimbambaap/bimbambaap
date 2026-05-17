# API

REST API gebouwd in Go met Gin. Accounts, posts met foto's, JWT authenticatie.

## Routes

### Auth
| Method | Route | Beschrijving | Auth nodig |
|--------|-------|-------------|------------|
| POST | `/api/v1/auth/register` | Account aanmaken | Nee |
| POST | `/api/v1/auth/login` | Inloggen | Nee |

### Users
| Method | Route | Beschrijving | Auth nodig |
|--------|-------|-------------|------------|
| GET | `/api/v1/users/me` | Eigen profiel | Ja |
| PUT | `/api/v1/users/me` | Profiel updaten | Ja |
| GET | `/api/v1/users/:username/posts` | Posts van gebruiker | Nee |

### Posts
| Method | Route | Beschrijving | Auth nodig |
|--------|-------|-------------|------------|
| GET | `/api/v1/posts/feed` | Feed ophalen | Nee |
| GET | `/api/v1/posts/:id` | Één post | Nee |
| POST | `/api/v1/posts` | Post aanmaken | Ja |
| DELETE | `/api/v1/posts/:id` | Post verwijderen | Ja |

## Lokaal draaien

```bash
# 1. Kopieer env bestand
cp .env.example .env
# Vul .env in met jouw waarden

# 2. Dependencies installeren
go mod tidy

# 3. Starten
go run main.go
```

## Op Railway deployen

```bash
# 1. Install Railway CLI
npm install -g @railway/cli

# 2. Login
railway login

# 3. Nieuw project
railway init

# 4. Voeg Postgres toe via Railway dashboard
#    Ga naar je project -> New -> Database -> PostgreSQL
#    Railway zet DATABASE_URL automatisch als environment variable

# 5. Zet de rest van je environment variables in Railway dashboard:
#    JWT_SECRET
#    CLOUDINARY_CLOUD_NAME
#    CLOUDINARY_API_KEY
#    CLOUDINARY_API_SECRET

# 6. Deploy
railway up
```

## Post aanmaken (multipart form)

```bash
curl -X POST https://jouw-api.railway.app/api/v1/posts \
  -H "Authorization: Bearer jouw_token" \
  -F "image=@foto.jpg" \
  -F "caption=Mijn eerste post"
```

## Inloggen

```bash
curl -X POST https://jouw-api.railway.app/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@test.com","password":"wachtwoord123"}'
```

Response:
```json
{
  "token": "eyJ...",
  "user": {
    "id": 1,
    "username": "pietje",
    "email": "test@test.com"
  }
}
```

Gebruik dit token als `Authorization: Bearer <token>` header bij beveiligde routes.

## Beveiliging

- Wachtwoorden worden gehasht met bcrypt
- JWT tokens verlopen na 30 dagen
- Rate limiting: max 60 requests per minuut per IP
- Gebruikers kunnen alleen hun eigen posts verwijderen
- Foto upload max 10MB
