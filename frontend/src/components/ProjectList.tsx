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
    // NEW: Prop to handle live user removals
    onUserRemoved: (projectId: number, userId: number) => void;
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
    onUserRemoved // Destructure new prop
}) => {
    const [showInfoMap, setShowInfoMap] = useState<Record<number, boolean>>({});

    const token = localStorage.getItem('token');

    // Updated hook call with the 5th argument (onUserRemoved)
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

    const handleDeleteClick = async (projectId: number, projectName: string) => {
        if (window.confirm(`Are you sure you want to delete "${projectName}"? This action cannot be undone.`)) {
            try {
                const token = localStorage.getItem('token');
                const response = await fetch(`http://localhost:880/projects/${projectId}`, {
                    method: 'DELETE',
                    headers: {
                        'Authorization': `Bearer ${token}`
                    }
                });

                if (response.ok) {
                    onDeleteProject(projectId);
                } else {
                    const errorData = await response.json();
                    alert(`Failed to delete: ${errorData.error || 'Unknown error'}`);
                }
            } catch (err) {
                console.error("Error deleting project:", err);
                alert("An error occurred while trying to delete the project.");
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
                                        Project owner: <span>{project.owner_name || `User #${project.owner_id}`}</span>
                                    </p>
                                </div>

                                {isOwner && (
                                    <button
                                        className="btn-danger"
                                        onClick={() => handleDeleteClick(project.id, project.name)}
                                        style={{
                                            backgroundColor: '#ff4d4f',
                                            color: 'white',
                                            padding: '5px 10px',
                                            fontSize: '0.8rem'
                                        }}
                                    >
                                        Delete Project
                                    </button>
                                )}
                            </div>
                        </div>

                        <button
                            className="btn-info"
                            onClick={() => setShowInfoMap(p => ({ ...p, [project.id]: !p[project.id] }))}
                        >
                            {showInfoMap[project.id] ? 'Hide Info' : 'Show Info'}
                        </button>

                        {showInfoMap[project.id] && (
                            <div className="info-content">{project.description}</div>
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
                                    />

                                    {isOwner && (
                                        <AddMemberSection
                                            projectId={project.id}
                                            onAdd={handleAddMember}
                                            excludeIds={projectUsersMap[project.id].map(u => u.id)}
                                        />
                                    )}

                                    {!isOwner && (
                                        <p style={{ fontSize: '0.8rem', color: '#64748b', textAlign: 'center', marginTop: '10px' }}>
                                            Only the owner can manage members.
                                        </p>
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