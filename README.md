# API for Chirpy 

## .env file requirements
db address
secret string for jwt token
apikey (polka key)

## dependencies 
github.com/google/uuid
github.com/joho/godotenv
github.com/lib/pq
github.com/golang-jwt/jwt/v5
golang.org/x/crypto/bcrypt

## how to use the api / endpoints

# Users

#### create a user account
request: POST /api/users
request body:

```json
{
  "email": "test@email.com",
  "password": "12345"
}
```

**Password will be hashed, then stored inside db.**

response body:

```json
{
  "id": "valid-uuid-here",
  "created_at": "2025-01-01T00:00:00Z",
  "updated_at": "2025-01-01T00:00:00Z",
  "email": "testing@email.com",
  "is_chirpy_red": false
}
```

#### login user
request: POST /api/login
request body:

```json
{
  "email": "test@email.com",
  "password": "12345"
}
```

response body:

```json
{
  "id": "valid-uuid-here",
  "created_at": "a split second ago",
  "updated_at": "a split second ago",
  "email": "tester@email.com",
  "is_chirpy_red": false,
  "token": "jwt-here",
  "refresh_token": "refresh-token"
}
```	

#### update user information
request: PUT /api/users
request body:

```json
{
  "email": "aaa@email.com",
  "password": "345aa"
}
```
**limitation: email and password must both be fresh. Will not work unless the requesting client has a (valid) jwt that links to the existing user attempting to change its login info.**

response body:

```json
{
  "id": "valid-uuid-here",
  "created_at": "2025-01-01T00:00:00Z",
  "updated_at": "2025-01-01T00:00:00Z",
  "email": "aaa@email.com",
  "is_chirpy_red": true
}
```

#### refresh jwt
request: POST /api/refresh
**requires authorization header in this form: 'Authorization: Bearer TOKEN_STRING'**

response body:

```json
{
    "token": "new-jwt-token"
}
```

#### revoke refresh token
request: POST /api/revoke
**requires authorization header in this form: 'Authorization: Bearer TOKEN_STRING'**

response: 204 code if all goes well

#### polka webhook
request: POST /api/polka/webhooks
request body:

```json
{
  "event": "user.upgraded",
  "data": {
    "user_id": "3311741c-680c-4546-99f3-fc9efac2036c"
  }
}
```

requires polka api key to be present in environmnet file

response: 204 in case event is anything other than "user.upgraded" and in case everything went well and the user was successfully upgraded (idempotent handler).

# Chirps 

#### create chirp
request: POST /api/chirps
request body:

```json
{
  "body": "Hello, world!",
  "user_id": "123e4567-e89b-12d3-a456-426614174000"
}
```

**limitation: chirp length (body) cannot exceed 140 tokens.**

response request:

```json
{
  "valid": true,
  "id": "94b7e44c-3604-42e3-bef7-ebfcc3efff8f",
  "body": "Hello, world!",
  "created_at": "2025-01-01T00:00:00Z",
  "updated_at": "2025-01-01T00:00:00Z",
  "user_id": "123e4567-e89b-12d3-a456-426614174000"
}
```		

#### load posted chirps
request: GET /api/chirps

**optional: author id query and sorting asc/desc**
examples:
+ GET /api/chirps?sort=asc&author_id=2
+ GET /api/chirps?sort=asc
+ GET /api/chirps?sort=desc

response body:

```json
[
  {
    "id": "94b7e44c-3604-42e3-bef7-ebfcc3efff8f",
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z",
    "body": "very cool chirp",
    "user_id": "123e4567-e89b-12d3-a456-426614174000"
  },
  {
    "id": "f0f87ec2-a8b5-48cc-b66a-a85ce7c7b862",
    "created_at": "2025-01-01T00:00:00Z",
    "updated_at": "2025-01-01T00:00:00Z",
    "body": "another really cool chirp",
    "user_id": "123e4567-e89b-12d3-a456-426614174000"
  }
]
```

#### load specific chirp (by id)
request: GET /api/chirps/{chirpID}

response body:
```json
  {
    "id": "chirpID",
    "created_at": "2023-01-01T00:00:00Z",
    "updated_at": "2023-01-01T00:00:00Z",
    "body": "was looking for this exact chirp",
    "user_id": "123e4567-e89b-12d3-a456-426614174000"
  }
```

#### delete chirp
request: DELETE /api/chirps/{chirpID}
response: 204 code upon successful deletion

# misc

#### check api status
request: GET /api/healthz
response: 200 code + "OK" message

#### admin metrics: hits counter
request: GET /admin/metrics

**requires PLATFORM="dev" setting from environment**

response: html template -> "Chirpy has been visited %d times!"

#### admin metrics: DELETE users and reset hits counter
request POST /admin/reset

**requires PLATFORM="dev" setting from environment**

response: "Hits reset to 0, users deleted"





	mux.HandleFunc("POST /admin/reset", apiCfg.resetHandler)
	// PUT





