import React, { useState, useEffect, type ChangeEvent } from 'react';

interface User {
    ID: number;
    Email: string;
}

interface UserSelectProps {
    onUserChange: (userId: string) => void;
    excludeIds: number[]; // IDs to hide from the list
}

const UserSelect: React.FC<UserSelectProps> = ({ onUserChange, excludeIds }) => {
    const [users, setUsers] = useState<User[]>([]);

    useEffect(() => {
        const fetchUsers = async () => {
            const token = localStorage.getItem('token');
            try {
                const response = await fetch('http://localhost:880/users', {
                    headers: { 'Authorization': `Bearer ${token}` }
                });
                const data = await response.json();

                // Normalize data format
                const allUsers: User[] = Array.isArray(data) ? data : (data.value || []);

                // Filter out people already in the project AND the creator
                const filtered = allUsers.filter(user => !excludeIds.includes(user.ID));

                setUsers(filtered);
            } catch (err) {
                console.error("Failed to load users", err);
            }
        };
        fetchUsers();
    }, [excludeIds]); // Re-filter if the member list changes

    return (
        <select
            onChange={(e: ChangeEvent<HTMLSelectElement>) => onUserChange(e.target.value)}
            defaultValue=""
        >
            <option value="" disabled>-- Select a User --</option>
            {users.map(u => (
                <option key={u.ID} value={u.ID.toString()}>{u.Email}</option>
            ))}
            {users.length === 0 && <option disabled>No other users to add</option>}
        </select>
    );
};

export default UserSelect;