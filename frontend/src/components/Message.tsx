import type { Message as MessageType } from '../types';

interface Props {
    message: MessageType,
    isOwn: boolean,
}

export function Message({ message, isOwn }: Props) {
    const statusClass = `message-status-${message.moderation_status}`;

    return (
        <div className={`message ${isOwn ? 'own' : ''} ${statusClass}`}>
            <div className="message-header">
            <span className="username">{message.username}</span>
            {message.moderation_status === 'pending' && (
                <span className="pending-indicator">⏳</span>
            )}
            {message.moderation_status === 'flagged' && (
                <span className="flagged-indicator">⚠️ Flagged</span>
            )}
            </div>
            <div className="message-content">
            {message.moderation_status === 'flagged'
                ? '[This message was flagged by moderation]'
                : message.content
            }
            </div>
        </div>
    )
}