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
}

const ProjectUsers: React.FC<ProjectUsersProps> = ({ users, onRemove }) => {
    return (
        <div className="member-list">
            {users.map((user) => (
                <div key={user.id} className="member-item">
                    <div className="member-info">
                        <span className="member-email">{user.email}</span>
                        <span className="member-role" style={{ marginLeft: '5px', opacity: 0.7 }}>({user.role})</span>
                    </div>

                    {/* Only show the button if onRemove was passed (Owner check) */}
                    {onRemove && (
                        <button
                            className="btn-danger-sm"
                            onClick={() => onRemove(user.id)}
                        >
                            Remove
                        </button>
                    )}
                </div>
            ))}
        </div>
    );
};

export default ProjectUsers;