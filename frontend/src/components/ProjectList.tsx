import React, { useState } from 'react';
import ProjectUsers, { type UserInProject } from './ProjectUsers';
import AddMemberSection from './AddMemberSection';

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
    handleAddMember
}) => {
    const [showInfoMap, setShowInfoMap] = useState<Record<number, boolean>>({});

    if (!showProjects) return null;

    return (
        <div className="projects-container">
            {projects.map((project) => {
                // DEBUG: Look at your browser console (F12) to see these values!
                console.log(`Project: ${project.name}`, {
                    projectOwnerID: project.owner_id,
                    loggedInUserID: currentUserId
                });

                // Strict comparison forcing both to Numbers
                const isOwner = Number(project.owner_id) === Number(currentUserId);

                return (
                    <div key={project.id} className="project-card">
                        <div className="project-header">
                            <h2>{project.name}</h2>
                            <p className="project-owner">
                                Project owner: <span>{project.owner_name || `User #${project.owner_id}`}</span>
                            </p>
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
                                        /* Pass function only if isOwner is true */
                                        onRemove={isOwner ? (uId) => removeUser(project.id, uId) : undefined}
                                    />

                                    {/* Only show the Add Member UI if the user is the owner */}
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