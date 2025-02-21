import axios from 'axios';

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api';
const WS_URL = process.env.NEXT_PUBLIC_WS_URL || 'ws://localhost:8080/api/tasks/ws';

const api = axios.create({
  baseURL: API_BASE_URL,
  headers: {
    'Content-Type': 'application/json',
  },
});

// Add request interceptor for authentication
api.interceptors.request.use((config) => {
  const token = localStorage.getItem('token');
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

// Add response interceptor for handling errors
api.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      localStorage.removeItem('token');
      localStorage.removeItem('user');
      window.location.href = '/login';
    }
    return Promise.reject(error);
  }
);

export interface User {
  id: string;
  email: string;
  created_at: string;
}

export interface Task {
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

export interface AuthResponse {
  token: string;
  user: User;
}

export interface TasksResponse {
  tasks: Task[];
  pagination: {
    current_page: number;
    page_size: number;
    total_items: number;
    total_pages: number;
  };
}

export interface CreateTaskRequest {
  title: string;
  description: string;
  priority: Task['priority'];
  assigned_to: string;
  due_date: string;
}

export interface UpdateTaskRequest {
  title?: string;
  description?: string;
  status?: Task['status'];
  priority?: Task['priority'];
  assigned_to?: string;
  due_date?: string;
}

export const auth = {
  register: (email: string, password: string) =>
    api.post<AuthResponse>('/auth/register', { email, password }),
  
  login: (email: string, password: string) =>
    api.post<AuthResponse>('/auth/login', { email, password }),
};

export const tasks = {
  create: (task: CreateTaskRequest) =>
    api.post<{ task: Task }>('/tasks', task),
  
  list: (params?: {
    status?: Task['status'];
    assigned_to?: string;
    page?: number;
    page_size?: number;
    sort_by?: keyof Task;
    sort_order?: 'asc' | 'desc';
  }) =>
    api.get<TasksResponse>('/tasks', { params }),
  
  update: (id: string, task: UpdateTaskRequest) =>
    api.put<{ task: Task }>(`/tasks/${id}`, task),
  
  delete: (id: string) =>
    api.delete<{ message: string }>(`/tasks/${id}`),
};

export interface WebSocketMessage {
  type: 'task_created' | 'task_updated' | 'task_deleted';
  task?: Task;
  task_id?: string;
}

export const createWebSocket = () => {
  const ws = new WebSocket(WS_URL);
  
  ws.onopen = () => {
    const token = localStorage.getItem('token');
    if (token) {
      ws.send(JSON.stringify({ type: 'auth', token }));
    }
  };

  return ws;
};

export default api;