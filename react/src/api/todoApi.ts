import axios from 'axios';

// Base URL points to the Go backend.
// In production, Nginx proxies /api/* to the backend. In dev, falls back to localhost:8080.
const api = axios.create({
  baseURL: import.meta.env.VITE_API_URL || (import.meta.env.PROD ? '/api' : 'http://localhost:8080'),
});

// Todo type matches the Go backend model
export interface Todo {
  id: number;
  title: string;
  status: number;
}

export interface Facet {
  [status: number]: number;
}

export interface PaginatedTodos {
  data: Todo[];
  total: number;
  page: number;
  limit: number;
}

let useCQRS = localStorage.getItem('useCQRS') === 'true';

export const getCQRSMode = () => useCQRS;

export const setCQRSMode = (mode: boolean) => {
  useCQRS = mode;
  localStorage.setItem('useCQRS', mode.toString());
};

const getPrefix = () => useCQRS ? '/cqrs/todos' : '/todos';

export const fetchFacets = async (searchQuery: string = ''): Promise<Facet> => {
  const { data } = await api.get(`${getPrefix()}/facets`, { params: { search: searchQuery } });
  return data;
};

export const fetchTodos = async (searchQuery: string = '', status?: number, page: number = 1, limit: number = 10): Promise<PaginatedTodos> => {
  const params: any = { search: searchQuery, page, limit };
  if (status) {
    params.status = status;
  }
  const { data } = await api.get(getPrefix(), { params });
  return data;
};

export const fetchTodoById = async (id: number): Promise<Todo> => {
  const { data } = await api.get(`/todos/${id}`);
  return data;
};

export const createTodo = async (title: string, passkey: string): Promise<Todo> => {
  const { data } = await api.post('/todos', { title }, { headers: { 'X-Passkey': passkey } });
  return data;
};

export const updateTodo = async (todo: Todo, passkey: string): Promise<Todo> => {
  const { data } = await api.put(`/todos/${todo.id}`, {
    title: todo.title,
    status: todo.status,
  }, { headers: { 'X-Passkey': passkey } });
  return data;
};

export const deleteTodo = async (id: number, passkey: string): Promise<void> => {
  await api.delete(`/todos/${id}`, { headers: { 'X-Passkey': passkey } });
};
