import React from "react";

export type UserInProject = {
    id: number;
    email: string;
    role: string;
};

export type ProjectUsersProps = {
    users: UserInProject[];
    removeUser: (userID: number) => Promise<void>;
};

const ProjectUsers: React.FC<ProjectUsersProps> = ({ users, removeUser }) => {
    return (
        <ul>
            {users.map((u) => (
                <li key={u.id}>
                    {u.email} — {u.role}{" "}
                    <button onClick={() => removeUser(u.id)}>Remove</button>
                </li>
            ))}
        </ul>
    );
};

export default ProjectUsers;
