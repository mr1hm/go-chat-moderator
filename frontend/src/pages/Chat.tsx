import { useEffect, useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { useAuth } from '../hooks/useAuth';
import { useWebSocket } from '../hooks/useWebSocket';
import { api } from '../api/client';
import { MessageList } from '../components/MessageList';
import { MessageInput } from '../components/MessageInput';
import { type Message } from '../types';

export function Chat() {
    const { roomId } = useParams<{ roomId: string }>();
    const navigate = useNavigate();
    const { token, user } = useAuth();
    const { messages, sendMessage, status, setMessages } = useWebSocket(roomId!, token);
    const [faqMessages, setFaqMessages] = useState<Message[]>([]);

    const handleMessages = async(content: string) => {
        if (content.startsWith('/')) {
            const question = content.slice(1).trim();
            if (!question) return;

            // Add user's question as local message
            const userMsg: Message = {
                id: `faq-${Date.now()}`,
                content: `[FAQ] ${question}`,
                user_id: 'you',
                username: 'You',
                room_id: roomId!,
                created_at: new Date().toISOString(),
                moderation_status: 'approved',
            }
            setFaqMessages(prev => [...prev, userMsg]);

            try {
                const response = await api.askFAQ(question);
                const aiMsg: Message = {
                    id: `faq-ai-${Date.now()}`,
                    content: `[AI] ${response.answer}`,
                    user_id: 'ai',
                    username: 'FAQ Bot',
                    room_id: roomId!,
                    created_at: new Date().toISOString(),
                    moderation_status: 'approved',
                };
                setFaqMessages(prev => [...prev, aiMsg]);
            } catch (err) {
                console.error('FAQ error:', err);
            }
        } else {
            sendMessage(content);
        }
    }

    useEffect(() => {
        if (roomId) {
            api.getMessages(roomId).then(data => setMessages(data || [])).catch(console.error);
        }
    }, [roomId, setMessages]);

    if (!roomId) return null;

    return (
        <div className="chat-container">
            <header>
                <button onClick={() => navigate('/rooms')}>← Back</button>
                <span className={`status ${status}`}>{status}</span>
            </header>
            <MessageList messages={[...messages, ...faqMessages].sort((a, b) =>
                new Date(a.created_at).getTime() - new Date(b.created_at).getTime()
            )} currentUserId={user?.id} />
            <MessageInput onSend={handleMessages} disabled={status !== 'connected'} />
        </div>
    )
}