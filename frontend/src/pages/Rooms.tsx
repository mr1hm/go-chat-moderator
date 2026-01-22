import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { api } from '../api/client';
import { useAuth } from '../hooks/useAuth';
import type { Room } from '../types';

export function Rooms() {
    const [rooms, setRooms] = useState<Room[]>([]);
    const [newRoomName, setNewRoomName] = useState('');
    const [error, setError] = useState('');
    const navigate = useNavigate();
    const { logout, user } = useAuth();

    useEffect(() => {
        api.getRooms().then(setRooms).catch(console.error);
    }, []);

    const handleCreateRoom = async (e: React.FormEvent) => {
        e.preventDefault();
        
        if (!newRoomName.trim()) return;
        try {
            const room = await api.createRoom(newRoomName);
            setRooms([...rooms, room]);
            setNewRoomName('');
        } catch (err) {
            setError(err instanceof Error ? err.message : 'Failed to create room');
        }
    };

    return (
        <div className="rooms-container">
            <header>
                <h1>Chat Rooms</h1>
                <div>
                    <span>Welcome, {user?.username}</span>
                    <button onClick={logout}>Logout</button>
                </div>
            </header>

            {error && <div className="error">{error}</div>}

            <form onSubmit={handleCreateRoom} className="create-room">
                <input
                    type="text"
                    placeholder="New room name"
                    value={newRoomName}
                    onChange={(e) => setNewRoomName(e.target.value)}
                />
                <button type="submit">Create Room</button>
            </form>

            <div className="room-list">
            {rooms.map((room) => (
                <div
                key={room.id}
                className="room-card"
                onClick={() => navigate(`/chat/${room.id}`)}
                >
                    <h3>{room.name}</h3>
                    <small>Created {new Date(room.created_at).toLocaleDateString()}</small>
                </div>
            ))}
            </div>
        </div>
    );
}