import { useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { useAuth } from '../hooks/useAuth';
import { useWebSocket } from '../hooks/useWebSocket';
import { api } from '../api/client';
import { MessageList } from '../components/MessageList';
import { MessageInput } from '../components/MessageInput';

export function Chat() {
    const { roomId } = useParams<{ roomId: string }>();
    const navigate = useNavigate();
    const { token, user } = useAuth();
    const { messages, sendMessage, status, setMessages } = useWebSocket(roomId!, token);

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
            <MessageList messages={messages} currentUserId={user?.id} />
            <MessageInput onSend={sendMessage} disabled={status !== 'connected'} />
        </div>
    )
}