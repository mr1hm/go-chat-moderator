export interface User {
    id: string;
    email: string;
    username: string;
}

export interface AuthResponse {
    token: string;
    user: User;
}

export interface Room {
    id: string;
    name: string;
    created_by: string;
    created_at: string;
}

export interface Message {
    id: string;
    room_id: string;
    user_id: string;
    username: string;
    content: string;
    moderation_status: 'pending' | 'approved' | 'flagged';
    created_at: string;
}

export interface WSMessage {
    type: 'message' | 'moderation_update'
    payload: Message | ModerationUpdate;
}

export interface ModerationUpdate {
    message_id: string;
    status: 'approved' | 'flagged';
}