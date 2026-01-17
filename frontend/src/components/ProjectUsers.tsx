import React from 'react';

export interface UserInProject {
    id: number;
    email: string;
    role: string;
}

interface ProjectUsersProps {
    users: UserInProject[];
    // The '?' makes it optional, fixing the ts(2322) error
    onRemove?: (userId: number) => void;
    onMessage?: (userId: number, userEmail: string) => void;
}

const ProjectUsers: React.FC<ProjectUsersProps> = ({ users, onRemove, onMessage }) => {
    return (
        <div className="member-list">
            {users.map((user) => (
                <div key={user.id} className="member-item">
                    <div className="member-info">
                        <span className="member-email">{user.email}</span>
                        <span className="member-role" style={{ marginLeft: '5px', opacity: 0.7 }}>({user.role})</span>
                    </div>

                    <div style={{ display: 'flex', gap: '5px' }}>
                        {/* Message button - shows for all users */}
                        {onMessage && (
                            <button
                                className="btn-primary-sm"
                                onClick={() => onMessage(user.id, user.email)}
                            >
                                Message
                            </button>
                        )}

                        {/* Only show the remove button if onRemove was passed (Owner check) */}
                        {onRemove && (
                            <button
                                className="btn-danger-sm"
                                onClick={() => onRemove(user.id)}
                            >
                                Remove
                            </button>
                        )}
                    </div>
                </div>
            ))}
        </div>
    );
};

export default ProjectUsers;