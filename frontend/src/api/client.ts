import type { AuthResponse, Room, Message } from '../types'

const API_URL = '/api'

function getToken(): string | null {
    return localStorage.getItem('token');
}

async function request<T>(path: string, options: RequestInit = {}): Promise<T> {
    const token = getToken();
    const headers: HeadersInit = {
        'Content-Type': 'application/json',
        ...(token && { Authorization: `Bearer ${token}` }),
        ...options.headers,
    };

    const res = await fetch(`${API_URL}${path}`, { ...options, headers });
    if (!res.ok) {
        const error = await res.json()
        throw new Error(error.error || 'Request failed');
    }

    return res.json()
}

export const api = {
    register: (email: string, password: string, username: string) =>
        request<AuthResponse>('/register', {
            method: 'POST',
            body: JSON.stringify({ email, password, username }),
        }),

    login: (email: string, password: string) =>
        request<AuthResponse>('/login', {
            method: 'POST',
            body: JSON.stringify({ email, password }),
        }),

    getRooms: () => request<Room[]>('/rooms'),

    createRoom: (name: string) =>
        request<Room>('/rooms', {
            method: 'POST',
            body: JSON.stringify({ name })
        }),

    getMessages: (roomId: string, limit = 50) =>
        request<Message[]>(`/rooms/${roomId}/messages?limit=${limit}`),
}