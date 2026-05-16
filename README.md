# Genzite-Backend

## API Requests

Base URL: http://localhost:8080

### Health
- URL: http://localhost:8080/api/v1/health
- Request header: none
- Request body: none

### Auth
- URL: http://localhost:8080/api/v1/auth/register
- Request header: Content-Type: application/json
- Request body:
```json
{
	"email": "user@example.com",
	"password": "password123",
	"name": "User Name"
}
```

- URL: http://localhost:8080/api/v1/auth/login
- Request header: Content-Type: application/json
- Request body:
```json
{
	"email": "user@example.com",
	"password": "password123"
}
```

- URL: http://localhost:8080/api/v1/auth/me
- Request header: Authorization: Bearer <jwt>
- Request body: none

- URL: http://localhost:8080/api/v1/auth/google/login
- Request header: none
- Request body: none

- URL: http://localhost:8080/api/v1/auth/google/callback?state=<state>&code=<code>
- Request header: none
- Request body: none

### Migrations (admin + migrations:* permission required)
- URL: http://localhost:8080/api/v1/migrations/up
- Request header: Authorization: Bearer <jwt>
- Request body: none
- Optional query: steps=<positive integer>

- URL: http://localhost:8080/api/v1/migrations/down?steps=<positive integer>
- Request header: Authorization: Bearer <jwt>
- Request body: none

- URL: http://localhost:8080/api/v1/migrations/seed
- Request header: Authorization: Bearer <jwt>
- Request body: none

- URL: http://localhost:8080/api/v1/migrations/version
- Request header: Authorization: Bearer <jwt>
- Request body: none

### Web Builder
- URL: http://localhost:8080/api/v1/api/v1/web-builder/sites
- Request header: Authorization: Bearer <jwt>
- Request body:
```json
{
	"slug": "my-site",
	"config": {
		"title": "My Site",
		"name": "My Name",
		"headline": "Short headline",
		"bio": "Short bio",
		"avatar_url": "https://example.com/avatar.png",
		"links": [
			{"label": "GitHub", "url": "https://github.com/my"}
		]
	}
}
```

- URL: http://localhost:8080/api/v1/api/v1/web-builder/sites/:id/publish
- Request header: Authorization: Bearer <jwt>
- Request body: none

- URL (auto serve only): http://localhost:8080/api/v1/:slug
- Request header: none
- Request body: none

### Chatbot
- URL: http://localhost:8080/api/v1/api/v1/chatbot/generate
- Request header: Authorization: Bearer <jwt>
- Request body:
```json
{
	"slug": "my-site",
	"prompt": "I am a designer focused on minimal portfolios",
	"name": "My Name",
	"headline": "Short headline",
	"bio": "Short bio",
	"avatar_url": "https://example.com/avatar.png",
	"links": [
		{"label": "GitHub", "url": "https://github.com/my"}
	]
}
```

### Payment
- URL: http://localhost:8080/api/v1/api/v1/payment/transactions
- Request header: Authorization: Bearer <jwt>
- Request body:
```json
{
	"site_id": 1
}
```

- URL: http://localhost:8080/api/v1/api/v1/payment/webhook
- Request header: Content-Type: application/json
- Request body:
```json
{
	"order_id": "wb-1-1-20240502120000-abc123",
	"status_code": "200",
	"gross_amount": "99000",
	"signature_key": "<midtrans-signature>",
	"transaction_status": "settlement",
	"fraud_status": "accept"
}
```

### Template
- URL: http://localhost:8080/api/v1/api/v1/template/ping
- Request header: Authorization: Bearer <jwt>
- Request body: none