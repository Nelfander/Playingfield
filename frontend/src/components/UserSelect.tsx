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
                const allUsersRaw: any[] = Array.isArray(data) ? data : (data.value || []);

                // Map to consistent User interface to handle potential case differences (id vs ID)
                const normalizedUsers: User[] = allUsersRaw.map(u => ({
                    ID: u.ID || u.id,
                    Email: u.Email || u.email
                }));

                // Filter logic
                const filtered = normalizedUsers.filter(user => {
                    if (!user || !user.ID) return false;

                    // Convert both to Numbers to ensure comparison works regardless of type
                    return !excludeIds.map(Number).includes(Number(user.ID));
                });

                setUsers(filtered);
            } catch (err) {
                console.error("Failed to load users", err);
            }
        };
        fetchUsers();
    }, [excludeIds]);

    return (
        <select
            onChange={(e: ChangeEvent<HTMLSelectElement>) => onUserChange(e.target.value)}
            value="" // Force reset to placeholder after selection
        >
            <option value="" disabled>-- Select a User --</option>
            {users.map(u => (
                <option key={u.ID.toString()} value={u.ID.toString()}>
                    {u.Email || "Deleted User"}
                </option>
            ))}
            {users.length === 0 && <option disabled>No other users to add</option>}
        </select>
    );
};

export default UserSelect;