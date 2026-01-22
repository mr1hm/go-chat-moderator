import { useState } from 'react';

interface Props {
    onSend: (content: string) => void;
    disabled: boolean;
}

export function MessageInput({ onSend, disabled }: Props) {
    const [content, setContent] = useState('');

    const handleSubmit = (e: React.FormEvent) => {
        e.preventDefault();

        if (!content.trim() || disabled) return;
        onSend(content);
        setContent('');
    };

    return (
        <form className="message-input" onSubmit={handleSubmit}>
            <input
                type="text"
                value={content}
                onChange={(e) => setContent(e.target.value)}
                placeholder="Type a message..."
                disabled={disabled}
            />
            <button type="submit" disabled={disabled || !content.trim()}>
                Send
            </button>
      </form>
    )
};

