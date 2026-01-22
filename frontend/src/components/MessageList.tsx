import { useEffect, useRef } from 'react';
import type { Message as MessageType } from '../types';
import { Message } from './Message';

interface Props {
    messages: MessageType[];
    currentUserId?: string;
}

export function MessageList({ messages, currentUserId }: Props) {
    const bottomRef = useRef<HTMLDivElement>(null);

    useEffect(() => {
        bottomRef.current?.scrollIntoView({ behavior: 'smooth' });
    }, [messages]);

    return (
        <div className="message-list">
            {(messages || []).map((msg) => (
                <Message key={msg.id} message={msg} isOwn={msg.user_id === currentUserId} />
            ))}
            <div ref={bottomRef} />
        </div>
    );
}