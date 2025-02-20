## Authentication

### Token Management

- Implement JWT-based authentication
- Refresh tokens every 15 minutes
- Store access tokens securely

### CORS
- Backend allows all origins in development.
- For production, configure specific origins.

## Rate Limiting

### API
- 10 requests/second

### WebSocket
- 60 messages/minute

## Authentication

- JWT token required for all protected routes
- Store token in `localStorage`
- Include token in Authorization header

## Real-time Updates

- WebSocket connection for real-time task updates
- Implement reconnection logic
- Handle different message types

## Data Validation

- **Task title**: Required, max 255 chars
- **Description**: Optional, max 1000 chars
- **Priority**: low/medium/high
- **Status**: pending/in_progress/completed
- **Due date**: Must be in the future

## Error Handling

- Implement global error handling
- Handle rate limiting
- Handle authentication errors
- Show appropriate user feedback

---

## Environment Configuration

```ini
# .env.local
NEXT_PUBLIC_API_URL=http://localhost:8080/api
NEXT_PUBLIC_WS_URL=ws://localhost:8080/api/tasks/ws
```

---

## Type Definitions

```typescript
// types/task.ts
interface Task {
  id: string;
  title: string;
  description: string;
  status: 'pending' | 'in_progress' | 'completed';
  priority: 'low' | 'medium' | 'high';
  assigned_to: string;
  created_by: string;
  created_at: string;
  updated_at: string;
  due_date: string;
}

interface User {
  id: string;
  email: string;
  created_at: string;
}

interface PaginationParams {
  page: number;
  page_size: number;
  total_items: number;
  total_pages: number;
}

interface WebSocketMessage {
  type: 'task_created' | 'task_updated' | 'task_deleted' | 'task_assigned';
  payload: Task;
  timestamp: string;
}
```

---

## API Client Setup

```typescript
// lib/api.ts
import axios from 'axios';

const api = axios.create({
  baseURL: process.env.NEXT_PUBLIC_API_URL,
  headers: {
    'Content-Type': 'application/json',
  },
});

api.interceptors.request.use((config) => {
  const token = localStorage.getItem('token');
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});
```

---

## Authentication Hook

```typescript
// hooks/useAuth.ts
import { create } from 'zustand';

interface AuthStore {
  user: User | null;
  token: string | null;
  login: (email: string, password: string) => Promise<void>;
  logout: () => void;
  register: (email: string, password: string) => Promise<void>;
}

export const useAuth = create<AuthStore>((set) => ({
  user: null,
  token: null,
  login: async (email, password) => {
    const response = await api.post('/auth/login', { email, password });
    localStorage.setItem('token', response.data.token);
    set({ user: response.data.user, token: response.data.token });
  },
  logout: () => {
    localStorage.removeItem('token');
    set({ user: null, token: null });
  },
  register: async (email, password) => {
    const response = await api.post('/auth/register', { email, password });
    localStorage.setItem('token', response.data.token);
    set({ user: response.data.user, token: response.data.token });
  },
}));
```

---

## WebSocket Connection Hook

```typescript
// hooks/useWebSocket.ts
import { useEffect, useRef } from 'react';

export const useWebSocket = (onMessage: (message: WebSocketMessage) => void) => {
  const ws = useRef<WebSocket | null>(null);

  useEffect(() => {
    const connect = () => {
      ws.current = new WebSocket(process.env.NEXT_PUBLIC_WS_URL!);

      ws.current.onmessage = (event) => {
        const message: WebSocketMessage = JSON.parse(event.data);
        onMessage(message);
      };

      ws.current.onclose = () => {
        setTimeout(connect, 5000); // Reconnect after 5 seconds
      };
    };

    connect();

    const heartbeat = setInterval(() => {
      if (ws.current?.readyState === WebSocket.OPEN) {
        ws.current.send('ping');
      }
    }, 30000);

    return () => {
      clearInterval(heartbeat);
      ws.current?.close();
    };
  }, [onMessage]);

  return ws;
};
```

---

## Required Dependencies

```json
{
  "dependencies": {
    "@hookform/resolvers": "^3.3.4",
    "axios": "^1.6.7",
    "date-fns": "^3.3.1",
    "next": "14.1.0",
    "react": "^18.2.0",
    "react-dom": "^18.2.0",
    "react-hook-form": "^7.50.1",
    "react-query": "^3.39.3",
    "tailwindcss": "^3.4.1",
    "zod": "^3.22.4",
    "zustand": "^4.5.1"
  }
}
```

---

## Validation Schema (Using Zod)

```typescript
// lib/validation.ts
import { z } from 'zod';

export const taskSchema = z.object({
  title: z.string().min(1).max(255),
  description: z.string().max(1000).optional(),
  priority: z.enum(['low', 'medium', 'high']),
  due_date: z.string().datetime(),
  assigned_to: z.string().uuid(),
});

export const loginSchema = z.object({
  email: z.string().email(),
  password: z.string().min(8).regex(/\d/, 'Password must contain at least one number'),
});
```

---

## Error Handling

```typescript
// lib/error-handling.ts
interface ApiError {
  error: string;
  details?: string;
  retry_after?: string;
}

export const handleApiError = (error: unknown) => {
  if (axios.isAxiosError(error) && error.response) {
    const data = error.response.data as ApiError;
    
    switch (error.response.status) {
      case 401:
        useAuth.getState().logout();
        return 'Session expired. Please login again.';
      case 429:
        return `Rate limit exceeded. Please try again in ${data.retry_after}.`;
      default:
        return data.error || 'An unexpected error occurred';
    }
  }
  return 'Network error occurred';
};
```
