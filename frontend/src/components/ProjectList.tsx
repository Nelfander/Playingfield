import React, { useState } from 'react';
import ProjectUsers, { type UserInProject } from './ProjectUsers';
import AddMemberSection from './AddMemberSection';
import TaskBoard from './TaskBoard';


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
    showUsersMap: Record<number, boolean>;
    showTasksMap: Record<number, boolean>;
    // CHANGED: Simple number instead of Record
    taskRefreshTick: number;
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
    onProjectUpdated: () => void;
}

const ProjectList: React.FC<ProjectListProps> = ({
    projects,
    currentUserId,
    showProjects,
    projectUsersMap,
    showUsersMap,
    showTasksMap,
    taskRefreshTick,
    toggleProjectUsers,
    toggleTasks,
    removeUser,
    handleAddMember,
    onDeleteProject,
    onSelectProject,
    onStartDM,
    onProjectUpdated
}) => {
    const [showInfoMap, setShowInfoMap] = useState<Record<number, boolean>>({});
    const [editingProjectId, setEditingProjectId] = useState<number | null>(null);
    const [editForm, setEditForm] = useState({ name: '', description: '' });

    const token = localStorage.getItem('token');

    if (!showProjects) return null;

    if (!projects || !Array.isArray(projects)) {
        return <div className="projects-container"><p>Loading projects...</p></div>;
    }

    const handleUpdate = async (projectId: number) => {
        try {
            const response = await fetch(`http://localhost:880/projects/${projectId}`, {
                method: 'PUT',
                headers: {
                    'Authorization': `Bearer ${token}`,
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify(editForm)
            });

            if (response.ok) {
                setEditingProjectId(null);
                onProjectUpdated();
            } else {
                const errorData = await response.json();
                alert(`Update failed: ${errorData.error || 'Unknown error'}`);
            }
        } catch (err) {
            console.error("Error updating project:", err);
        }
    };

    const handleDeleteClick = async (projectId: number, projectName: string) => {
        if (window.confirm(`Are you sure you want to delete "${projectName}"?`)) {
            try {
                const response = await fetch(`http://localhost:880/projects/${projectId}`, {
                    method: 'DELETE',
                    headers: { 'Authorization': `Bearer ${token}` }
                });

                if (response.ok) {
                    onDeleteProject(projectId);
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
                const isEditing = editingProjectId === project.id;

                return (
                    <div key={project.id} className="project-card">
                        <div className="project-header">
                            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
                                <div style={{ flex: 1 }}>
                                    {isEditing ? (
                                        <input
                                            value={editForm.name}
                                            onChange={(e) => setEditForm({ ...editForm, name: e.target.value })}
                                            style={{ fontSize: '1.2rem', fontWeight: 'bold', width: '90%' }}
                                        />
                                    ) : (
                                        <h2>{project.name}</h2>
                                    )}
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

                                    {isOwner && !isEditing && (
                                        <button
                                            onClick={() => {
                                                setEditingProjectId(project.id);
                                                setEditForm({ name: project.name, description: project.description });
                                            }}
                                            style={{ backgroundColor: '#faad14', color: 'white', padding: '5px 10px', fontSize: '0.8rem', border: 'none', borderRadius: '4px', cursor: 'pointer' }}
                                        >
                                            Edit
                                        </button>
                                    )}

                                    {isEditing && (
                                        <button
                                            onClick={() => handleUpdate(project.id)}
                                            style={{ backgroundColor: '#52c41a', color: 'white', padding: '5px 10px', fontSize: '0.8rem', border: 'none', borderRadius: '4px', cursor: 'pointer' }}
                                        >
                                            Save
                                        </button>
                                    )}

                                    {isOwner && (
                                        <button
                                            className="btn-danger"
                                            onClick={() => isEditing ? setEditingProjectId(null) : handleDeleteClick(project.id, project.name)}
                                            style={{ backgroundColor: '#ff4d4f', color: 'white', padding: '5px 10px', fontSize: '0.8rem', border: 'none', borderRadius: '4px', cursor: 'pointer' }}
                                        >
                                            {isEditing ? 'Cancel' : 'Delete'}
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
                                {isEditing ? (
                                    <textarea
                                        value={editForm.description}
                                        onChange={(e) => setEditForm({ ...editForm, description: e.target.value })}
                                        style={{ width: '100%', minHeight: '60px' }}
                                    />
                                ) : (
                                    project.description
                                )}
                            </div>
                        )}

                        <div className="accordion-section">
                            <button className="full-width" onClick={() => toggleProjectUsers(project.id)}>
                                {showUsersMap[project.id] ? "Hide Members" : "Show Members"}
                            </button>

                            {showUsersMap[project.id] && (
                                <div className="accordion-content">
                                    <ProjectUsers
                                        users={projectUsersMap[project.id] || []}
                                        onRemove={isOwner ? (uId) => removeUser(project.id, uId) : undefined}
                                        onMessage={onStartDM}
                                    />
                                    {isOwner && (
                                        <AddMemberSection
                                            projectId={project.id}
                                            onAdd={handleAddMember}
                                            excludeIds={(projectUsersMap[project.id] || []).map(u => u.id)}
                                        />
                                    )}
                                </div>
                            )}

                            <button className="full-width" onClick={() => toggleTasks(project.id)} style={{ marginTop: '5px', backgroundColor: '#52c41a', color: 'white' }}>
                                {showTasksMap[project.id] ? "Hide Tasks" : "Show Tasks"}
                            </button>

                            {showTasksMap[project.id] && (
                                <div className="accordion-content">
                                    <TaskBoard
                                        projectId={project.id}
                                        // CHANGED: Use the simple number directly
                                        refreshTick={taskRefreshTick}
                                        isOwner={isOwner}
                                        members={projectUsersMap[project.id] || []}
                                    />
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