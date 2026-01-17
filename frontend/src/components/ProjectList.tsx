import React, { useState } from 'react';
import ProjectUsers, { type UserInProject } from './ProjectUsers';
import AddMemberSection from './AddMemberSection';
import { useWebSockets } from '../hooks/useWebSockets';

interface Project {
    id: number;
    name: string;
    description: string;
    owner_id: number;
    owner_name?: string;
}

interface ProjectListProps {
    projects: Project[];
    currentUserId: number;
    showProjects: boolean;
    projectUsersMap: Record<number, UserInProject[]>;
    showTasksMap: Record<number, boolean>;
    toggleProjectUsers: (id: number) => void;
    toggleTasks: (id: number) => void;
    removeUser: (projectId: number, userId: number) => void;
    handleAddMember: (projectId: number, userId: number) => void;
    onDeleteProject: (projectId: number) => void;
    onUserAdded: (projectId: number, userId: number, role: string) => void;
    onProjectCreated: () => void;
    onUserRemoved: (projectId: number, userId: number) => void;
    onSelectProject: (projectId: number) => void;
    onStartDM: (userId: number, userEmail: string) => void;
}

const ProjectList: React.FC<ProjectListProps> = ({
    projects,
    currentUserId,
    showProjects,
    projectUsersMap,
    showTasksMap,
    toggleProjectUsers,
    toggleTasks,
    removeUser,
    handleAddMember,
    onDeleteProject,
    onUserAdded,
    onProjectCreated,
    onUserRemoved,
    onSelectProject,
    onStartDM
}) => {
    const [showInfoMap, setShowInfoMap] = useState<Record<number, boolean>>({});
    const token = localStorage.getItem('token');

    useWebSockets(
        token,
        (id) => {
            console.log(`Live Update: Project ${id} was deleted.`);
            onDeleteProject(id);
        },
        (projectId, userId, role) => {
            console.log(`Live Update: User ${userId} added to project ${projectId}.`);
            onUserAdded(projectId, userId, role);
        },
        () => {
            console.log("Live Update: A new project was created.");
            onProjectCreated();
        },
        (projectId, userId) => {
            console.log(`Live Update: User ${userId} removed from project ${projectId}.`);
            onUserRemoved(projectId, userId);
        }
    );

    if (!showProjects) return null;

    // --- GUARD CLAUSE: Fixes "projects.map is not a function" ---
    if (!projects || !Array.isArray(projects)) {
        return (
            <div className="projects-container">
                <p>Loading projects...</p>
            </div>
        );
    }
    // -----------------------------------------------------------

    const handleDeleteClick = async (projectId: number, projectName: string) => {
        if (window.confirm(`Are you sure you want to delete "${projectName}"?`)) {
            try {
                const response = await fetch(`http://localhost:880/projects/${projectId}`, {
                    method: 'DELETE',
                    headers: { 'Authorization': `Bearer ${token}` }
                });

                if (response.ok) {
                    onDeleteProject(projectId);
                } else {
                    const errorData = await response.json();
                    alert(`Failed to delete: ${errorData.error || 'Unknown error'}`);
                }
            } catch (err) {
                console.error("Error deleting project:", err);
            }
        }
    };

    return (
        <div className="projects-container">
            {projects.map((project) => {
                const isOwner = Number(project.owner_id) === Number(currentUserId);

                return (
                    <div key={project.id} className="project-card">
                        <div className="project-header">
                            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
                                <div>
                                    <h2>{project.name}</h2>
                                    <p className="project-owner">
                                        Owner: <span>{project.owner_name || `User #${project.owner_id}`}</span>
                                    </p>
                                </div>

                                <div style={{ display: 'flex', gap: '8px' }}>
                                    <button
                                        className="btn-chat"
                                        onClick={() => onSelectProject(project.id)}
                                        style={{ backgroundColor: '#1890ff', color: 'white', padding: '5px 10px', fontSize: '0.8rem', border: 'none', borderRadius: '4px', cursor: 'pointer' }}
                                    >
                                        💬 Chat
                                    </button>

                                    {isOwner && (
                                        <button
                                            className="btn-danger"
                                            onClick={() => handleDeleteClick(project.id, project.name)}
                                            style={{ backgroundColor: '#ff4d4f', color: 'white', padding: '5px 10px', fontSize: '0.8rem', border: 'none', borderRadius: '4px', cursor: 'pointer' }}
                                        >
                                            Delete
                                        </button>
                                    )}
                                </div>
                            </div>
                        </div>

                        <button
                            className="btn-info"
                            onClick={() => setShowInfoMap(p => ({ ...p, [project.id]: !p[project.id] }))}
                        >
                            {showInfoMap[project.id] ? 'Hide Info' : 'Show Info'}
                        </button>

                        {showInfoMap[project.id] && (
                            <div className="info-content" style={{ padding: '10px', backgroundColor: '#f9f9f9', marginTop: '5px' }}>
                                {project.description}
                            </div>
                        )}

                        <div className="accordion-section">
                            <button className="full-width" onClick={() => toggleProjectUsers(project.id)}>
                                {projectUsersMap[project.id] ? "Hide Members" : "Show Members"}
                            </button>

                            {projectUsersMap[project.id] && (
                                <div className="accordion-content">
                                    <ProjectUsers
                                        users={projectUsersMap[project.id]}
                                        onRemove={isOwner ? (uId) => removeUser(project.id, uId) : undefined}
                                        onMessage={onStartDM}
                                    />

                                    {isOwner && (
                                        <AddMemberSection
                                            projectId={project.id}
                                            onAdd={handleAddMember}
                                            excludeIds={projectUsersMap[project.id].map(u => u.id)}
                                        />
                                    )}
                                </div>
                            )}
                        </div>
                    </div>
                );
            })}
        </div>
    );
};

export default ProjectList;