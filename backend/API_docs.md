# Real-Time Task Management System API Documentation

## Base URL
`api`

## CORS Configuration
The API allows cross-origin requests with the following configuration:

- **Allowed Origins:** `*` (all origins)
- **Allowed Methods:** `GET`, `POST`, `PUT`, `DELETE`, `OPTIONS`
- **Allowed Headers:** `Origin`, `Authorization`, `Content-Type`

## Authentication
All protected endpoints require a JWT token passed in the `Authorization` header:

```
Authorization: Bearer <token>
```

---

## Auth Endpoints

### Register User
**POST** `/auth/register`

**Content-Type:** `application/json`

```json
{
  "email": "user@example.com",
  "password": "password123" // minimum 8 characters with at least 1 number
}
```

**Response 201:**
```json
{
  "token": "jwt_token_here",
  "user": {
    "id": "uuid",
    "email": "user@example.com",
    "created_at": "2024-03-10T15:04:05Z"
  }
}
```

### Login
**POST** `/auth/login`

**Content-Type:** `application/json`

```json
{
  "email": "user@example.com",
  "password": "password123"
}
```

**Response 200:**
```json
{
  "token": "jwt_token_here",
  "user": {
    "id": "uuid",
    "email": "user@example.com",
    "created_at": "2024-03-10T15:04:05Z"
  }
}
```

---

## Task Management

### Create Task

**POST** `/tasks`

**Authorization:** `Bearer <token>`

**Content-Type:** `application/json`

```json
{
  "title": "Task Title",
  "description": "Task description",
  "priority": "low|medium|high",
  "assigned_to": "user_uuid",
  "due_date": "2024-03-20T15:00:00Z"
}
```

**Response 201:**
```json
{
  "task": {
    "id": "uuid",
    "title": "Task Title",
    "description": "Task description",
    "status": "pending",
    "priority": "low",
    "assigned_to": "user_uuid",
    "created_by": "user_uuid",
    "created_at": "2024-03-10T15:04:05Z",
    "updated_at": "2024-03-10T15:04:05Z",
    "due_date": "2024-03-20T15:00:00Z"
  }
}
```

### List Tasks

**GET** `/tasks?status=pending&assigned_to=user_uuid&page=1&page_size=10&sort_by=created_at&sort_order=desc`

**Response 200:**
```json
{
  "tasks": [
    {
      "id": "uuid",
      "title": "Task Title",
      "description": "Task description",
      "status": "pending",
      "priority": "low",
      "assigned_to": "user_uuid",
      "created_by": "user_uuid",
      "created_at": "2024-03-10T15:04:05Z",
      "updated_at": "2024-03-10T15:04:05Z",
      "due_date": "2024-03-20T15:00:00Z"
    }
  ],
  "pagination": {
    "current_page": 1,
    "page_size": 10,
    "total_items": 50,
    "total_pages": 5
  }
}
```

### Update Task

**PUT** `/tasks/:id`

**Authorization:** `Bearer <token>`

**Content-Type:** `application/json`

```json
{
  "title": "Updated Title",
  "description": "Updated description",
  "status": "in_progress",
  "priority": "high",
  "assigned_to": "user_uuid",
  "due_date": "2024-03-25T15:00:00Z"
}
```

**Response 200:**
```json
{
  "task": {
    "id": "uuid",
    "title": "Updated Title",
    "description": "Updated description",
    "status": "in_progress",
    "priority": "high",
    "assigned_to": "user_uuid",
    "created_by": "user_uuid",
    "created_at": "2024-03-10T15:04:05Z",
    "updated_at": "2024-03-10T15:04:05Z",
    "due_date": "2024-03-25T15:00:00Z"
  }
}
```

### Delete Task

**DELETE** `/tasks/:id`

**Authorization:** `Bearer <token>`

**Response 200:**
```json
{
  "message": "task deleted successfully"
}
```

---

## WebSocket Connection

### Connect to WebSocket

```javascript
const ws = new WebSocket('wss://yourdomain.com/api/tasks/ws');

ws.onmessage = (event) => {
  const message = JSON.parse(event.data);
  console.log('Received:', message);
};

ws.onclose = () => {
  console.log('WebSocket connection closed');
};

setInterval(() => {
  if (ws.readyState === WebSocket.OPEN) {
    ws.send('ping');
  }
}, 30000);
```

---

## Error Responses

### Common Errors
```json
{ "error": "validation error message" }
{ "error": "unauthorized" }
{ "error": "task not found" }
{ "error": "rate limit exceeded", "retry_after": "60s" }
{ "error": "internal server error" }
```

---

## Rate Limiting
- API requests: `10 requests per second per client`
- WebSocket messages: `60 messages per minute per client`

## Data Validation
- **Task title:** Required, max 255 characters
- **Task description:** Optional, max 1000 characters
- **Task priority:** Required, one of: `low`, `medium`, `high`
- **Task status:** One of: `pending`, `in_progress`, `completed`
- **Due date:** Required, must be in the future
- **Page size:** Maximum 100 items per page
