import { useState, useEffect, useCallback, useRef } from 'react';
import type { Message, WSMessage, ModerationUpdate } from '../types';

const WS_URL = 'ws://localhost:8080';

export function useWebSocket(roomId: string, token: string | null) {
    const [messages, setMessages] = useState<Message[]>([]);
    const [status, setStatus] = useState<'connecting' | 'connected' | 'disconnected'>('disconnected');
    const wsRef = useRef<WebSocket | null>(null);

    useEffect(() => {
        if (!token || !roomId) return;

        setStatus('connecting');
        const ws = new WebSocket(`${WS_URL}/ws/${roomId}?token=${token}`)
        wsRef.current = ws;

        ws.onopen = () => setStatus('connected');
        ws.onclose = () => setStatus('disconnected');
        ws.onerror = () => setStatus('disconnected');

        ws.onmessage = (event) => {
            const data: WSMessage = JSON.parse(event.data);

            if (data.type === 'message') {
                setMessages((prev) => [...(prev || []), data.payload as Message]);
            } else if (data.type === 'moderation_update') {
                const update = data.payload as ModerationUpdate;
                setMessages((prev) =>
                    (prev || []).map((msg) =>
                        msg.id === update.message_id
                            ? { ...msg, moderation_status: update.status }
                            : msg
                    )
                );
            }
        };

        return () => {
            ws.close();
            wsRef.current = null;
        };
    }, [roomId, token]);

    const sendMessage = useCallback((content: string) => {
        if (wsRef.current?.readyState === WebSocket.OPEN) {
            wsRef.current.send(JSON.stringify({ content }));
        }
    }, []);

    return { messages, sendMessage, status, setMessages };
} 