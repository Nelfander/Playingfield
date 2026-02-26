import React from 'react';

export interface UserInProject {
    id: number;
    email: string;
    role: string;
}

interface ProjectUsersProps {
    users: UserInProject[];
    // NEW: Added unreadDMs to the interface to fix ts(2322)
    unreadDMs: Record<number, boolean>;
    onRemove?: (userId: number) => void;
    onMessage?: (userId: number, userEmail: string) => void;
}

const ProjectUsers: React.FC<ProjectUsersProps> = ({ users, unreadDMs, onRemove, onMessage }) => {
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
                            <div style={{ position: 'relative', display: 'inline-block' }}>
                                <button
                                    className="btn-primary-sm"
                                    onClick={() => onMessage(user.id, user.email)}
                                >
                                    Message
                                </button>

                                {/* THE RED DOT FOR DMS */}
                                {unreadDMs[user.id] && (
                                    <span style={{
                                        position: 'absolute',
                                        top: '-4px',
                                        right: '-4px',
                                        width: '10px',
                                        height: '10px',
                                        backgroundColor: '#ff4d4f',
                                        borderRadius: '50%',
                                        border: '2px solid white',
                                        boxShadow: '0 2px 4px rgba(0,0,0,0.2)',
                                        zIndex: 10
                                    }} />
                                )}
                            </div>
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