import React, { useState, useEffect } from 'react';

interface Attachment {
    id: number;
    file_name: string;
    file_url: string;
    file_size: number;
    created_at: string;
}

interface Task {
    id: number;
    project_id: number;
    title: string;
    description: string;
    status: 'TODO' | 'IN_PROGRESS' | 'DONE';
    assigned_to: number | null;
    attachments?: Attachment[];
}

interface User {
    id: number;
    email: string;
}

interface TaskActivity {
    id: number;
    task_id: number;
    user_id: number;
    user_email: string;
    action: string;
    details: string;
    created_at: string;
}

interface TaskBoardProps {
    projectId: number;
    refreshTick: number;
    isOwner: boolean;
    members: User[];
}

const formatFileSize = (bytes: number) => {
    if (!bytes || bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i];
};

const TaskBoard: React.FC<TaskBoardProps> = ({ projectId, refreshTick, isOwner, members }) => {
    const [tasks, setTasks] = useState<Task[]>([]);
    const [updatingTaskId, setUpdatingTaskId] = useState<number | null>(null);
    const [hoveredUploadId, setHoveredUploadId] = useState<number | null>(null); // New state for hover effect
    const [commitMessage, setCommitMessage] = useState("");
    const [newStatus, setNewStatus] = useState<'TODO' | 'IN_PROGRESS' | 'DONE'>('TODO');
    const [isAddingTask, setIsAddingTask] = useState(false);
    const [newTaskForm, setNewTaskForm] = useState({ title: '', description: '', assigned_to: '' });
    const [history, setHistory] = useState<TaskActivity[]>([]);
    const [showHistoryId, setShowHistoryId] = useState<number | null>(null);

    const token = localStorage.getItem('token');
    const currentUserId = token ? JSON.parse(atob(token.split('.')[1])).user_id : null;

    useEffect(() => {
        const timer = setTimeout(() => {
            fetchTasks();
        }, 150);
        return () => clearTimeout(timer);
    }, [projectId, refreshTick]);

    const fetchTasks = async () => {
        try {
            const res = await fetch(`http://localhost:880/projects/${projectId}/tasks`, {
                headers: { Authorization: `Bearer ${token}` }
            });
            const data = await res.json();
            if (Array.isArray(data)) {
                const tasksWithFiles = await Promise.all(data.map(async (task: Task) => {
                    try {
                        const fileRes = await fetch(`http://localhost:880/tasks/${task.id}/attachments`, {
                            headers: { Authorization: `Bearer ${token}` }
                        });
                        const attachments = await fileRes.json();
                        return { ...task, attachments: Array.isArray(attachments) ? attachments : [] };
                    } catch {
                        return { ...task, attachments: [] };
                    }
                }));
                setTasks(tasksWithFiles);
            } else {
                setTasks([]);
            }
        } catch (err) {
            console.error("Fetch tasks error:", err);
            setTasks([]);
        }
    };

    const formatActivityDetails = (details: string) => {
        if (!details) return "";
        return details.replace(/user (\d+)/g, (match, id) => {
            const member = members.find(m => m.id === parseInt(id));
            return member ? member.email : `User ${id}`;
        });
    };

    const handleCreateTask = async (e: React.FormEvent) => {
        e.preventDefault();
        const payload = {
            project_id: projectId,
            title: newTaskForm.title,
            description: newTaskForm.description,
            status: 'TODO',
            assigned_to: newTaskForm.assigned_to ? Number(newTaskForm.assigned_to) : null
        };
        try {
            const res = await fetch(`http://localhost:880/tasks`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${token}` },
                body: JSON.stringify(payload),
            });
            if (res.ok) {
                setIsAddingTask(false);
                setNewTaskForm({ title: '', description: '', assigned_to: '' });
                fetchTasks();
            }
        } catch (err) { console.error(err); }
    };

    const handleDeleteTask = async (taskId: number) => {
        if (!window.confirm("Are you sure you want to delete this task?")) return;
        try {
            const res = await fetch(`http://localhost:880/tasks/${taskId}`, {
                method: 'DELETE',
                headers: { Authorization: `Bearer ${token}` }
            });
            if (res.ok) fetchTasks();
        } catch (err) { console.error("Delete error:", err); }
    };

    const fetchHistory = async (taskId: number) => {
        try {
            const res = await fetch(`http://localhost:880/tasks/${taskId}/history`, {
                headers: { Authorization: `Bearer ${token}` }
            });
            const data = await res.json();
            setHistory(Array.isArray(data) ? data : []);
            setShowHistoryId(taskId);
        } catch (err) { console.error(err); }
    };

    const handleUpdateTask = async (taskId: number) => {
        if (!commitMessage.trim()) { alert("Please enter an update message."); return; }
        const currentTask = tasks.find(t => t.id === taskId);
        if (!currentTask) return;
        const payload = { ...currentTask, status: newStatus, message: commitMessage };
        try {
            const res = await fetch(`http://localhost:880/tasks/${taskId}`, {
                method: 'PUT',
                headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${token}` },
                body: JSON.stringify(payload),
            });
            if (res.ok) {
                setUpdatingTaskId(null);
                setCommitMessage("");
                fetchTasks();
            }
        } catch (err) { console.error(err); }
    };

    const handleFileUpload = async (taskId: number, e: React.ChangeEvent<HTMLInputElement>) => {
        const file = e.target.files?.[0];
        if (!file) return;
        const formData = new FormData();
        formData.append('file', file);
        try {
            const res = await fetch(`http://localhost:880/tasks/${taskId}/attachments`, {
                method: 'POST',
                headers: { Authorization: `Bearer ${token}` },
                body: formData,
            });
            if (res.ok) {
                fetchTasks();
            } else {
                const err = await res.json();
                alert(err.error || "Upload failed");
            }
        } catch (err) { console.error("Upload error:", err); }
    };

    const handleDeleteAttachment = async (attachmentId: number) => {
        if (!window.confirm("Delete this attachment forever?")) return;
        try {
            const res = await fetch(`http://localhost:880/tasks/attachments/${attachmentId}`, {
                method: 'DELETE',
                headers: { Authorization: `Bearer ${token}` }
            });
            if (res.ok) fetchTasks();
            else {
                const err = await res.json();
                alert(err.error || "Delete failed");
            }
        } catch (err) { console.error("Delete attachment error:", err); }
    };

    const handleDownload = async (attachmentId: number, fileName: string) => {
        try {
            const response = await fetch(`http://localhost:880/tasks/attachments/${attachmentId}/file`, {
                headers: { 'Authorization': `Bearer ${token}` }
            });
            const blob = await response.blob();
            const url = window.URL.createObjectURL(blob);
            const link = document.createElement('a');
            link.href = url;
            link.setAttribute('download', fileName);
            document.body.appendChild(link);
            link.click();
            link.parentNode?.removeChild(link);
            window.URL.revokeObjectURL(url);
        } catch (err) { console.error("Download error:", err); }
    };

    const renderTaskCard = (task: Task) => {
        const canManageAssets = isOwner || (task.assigned_to !== null && task.assigned_to === currentUserId);
        const canUpdateStatus = canManageAssets;

        return (
            <div key={task.id} className="task-card" style={{ border: '1px solid #ddd', padding: '12px', marginBottom: '10px', borderRadius: '6px', backgroundColor: 'white', boxShadow: '0 2px 4px rgba(0,0,0,0.05)' }}>
                <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                    <h4 style={{ margin: '0 0 5px 0', fontSize: '1rem' }}>{task.title}</h4>
                    {isOwner && (
                        <button onClick={() => handleDeleteTask(task.id)} style={{ border: 'none', background: 'none', color: '#ff4d4f', cursor: 'pointer', fontWeight: 'bold' }}>&times;</button>
                    )}
                </div>

                <p style={{ fontSize: '0.8rem', color: '#555', marginBottom: '8px' }}>{task.description}</p>

                <div style={{ fontSize: '0.7rem', color: '#888', marginBottom: '8px' }}>
                    👤 {members.find(m => m.id === task.assigned_to)?.email || "Unassigned"}
                </div>

                {task.attachments && task.attachments.length > 0 && (
                    <div style={{ marginTop: '5px', display: 'flex', flexWrap: 'wrap', gap: '5px', borderTop: '1px solid #f0f0f0', paddingTop: '8px' }}>
                        {task.attachments.map(file => (
                            <div key={file.id} style={{
                                display: 'flex',
                                alignItems: 'center',
                                gap: '4px',
                                background: '#e6f7ff',
                                padding: '2px 8px',
                                borderRadius: '12px',
                                border: '1px solid #91d5ff'
                            }}>
                                <button
                                    onClick={() => handleDownload(file.id, file.file_name)}
                                    style={{
                                        all: 'unset',
                                        fontSize: '0.65rem',
                                        cursor: 'pointer',
                                        color: '#1890ff',
                                        display: 'inline-flex',
                                        alignItems: 'center',
                                        fontWeight: 500
                                    }}
                                >
                                    📎 {file.file_name.length > 12 ? file.file_name.substring(0, 10) + "..." : file.file_name}
                                    <span style={{ marginLeft: '5px', fontSize: '0.6rem', color: '#8c8c8c', background: 'rgba(255,255,255,0.6)', padding: '0 4px', borderRadius: '4px' }}>
                                        {formatFileSize(file.file_size)}
                                    </span>
                                </button>
                                {canManageAssets && (
                                    <button onClick={() => handleDeleteAttachment(file.id)} style={{ border: 'none', background: 'none', color: '#ff4d4f', cursor: 'pointer', fontSize: '0.8rem', padding: '0 2px', marginLeft: '2px', display: 'flex', alignItems: 'center', fontWeight: 'bold' }}>&times;</button>
                                )}
                            </div>
                        ))}
                    </div>
                )}

                {updatingTaskId === task.id ? (
                    <div style={{ marginTop: '10px', borderTop: '1px solid #eee', paddingTop: '10px' }}>
                        <textarea value={commitMessage} onChange={e => setCommitMessage(e.target.value)} placeholder="Commit message..." style={{ width: '100%', fontSize: '0.8rem', minHeight: '50px', marginBottom: '5px' }} />
                        <label style={{ fontSize: '0.7rem', color: '#666' }}>Status:</label>
                        <select value={newStatus} onChange={e => setNewStatus(e.target.value as any)} style={{ width: '100%', marginBottom: '10px' }}>
                            <option value="TODO">To Do</option>
                            <option value="IN_PROGRESS">In Progress</option>
                            <option value="DONE">Done</option>
                        </select>
                        <div style={{ display: 'flex', gap: '5px' }}>
                            <button onClick={() => handleUpdateTask(task.id)} style={{ flex: 1, backgroundColor: '#52c41a', color: 'white', border: 'none', padding: '5px', borderRadius: '4px', cursor: 'pointer' }}>Save</button>
                            <button onClick={() => setUpdatingTaskId(null)} style={{ flex: 1, backgroundColor: '#ff4d4f', color: 'white', border: 'none', padding: '5px', borderRadius: '4px', cursor: 'pointer' }}>Cancel</button>
                        </div>
                    </div>
                ) : (
                    <div style={{ display: 'flex', gap: '5px', marginTop: '10px' }}>
                        <button
                            onClick={() => { setUpdatingTaskId(task.id); setNewStatus(task.status); }}
                            style={{ flex: 2, fontSize: '0.75rem', padding: '4px', cursor: 'pointer', backgroundColor: '#1890ff', color: 'white', border: 'none', borderRadius: '4px', display: canUpdateStatus ? 'block' : 'none' }}
                        >
                            Update
                        </button>

                        {canManageAssets && (
                            <label
                                onMouseEnter={() => setHoveredUploadId(task.id)}
                                onMouseLeave={() => setHoveredUploadId(null)}
                                style={{
                                    flex: 2,
                                    fontSize: '0.75rem',
                                    padding: '4px',
                                    textAlign: 'center',
                                    backgroundColor: hoveredUploadId === task.id ? '#e6e6e6' : '#f5f5f5',
                                    border: '1px solid #ccc',
                                    borderRadius: '4px',
                                    cursor: 'pointer',
                                    display: 'flex',
                                    alignItems: 'center',
                                    justifyContent: 'center',
                                    gap: '4px',
                                    transition: 'all 0.2s ease',
                                    transform: hoveredUploadId === task.id ? 'scale(1.05)' : 'scale(1)'
                                }}
                            >
                                📎 Upload File
                                <input type="file" hidden onChange={(e) => handleFileUpload(task.id, e)} />
                            </label>
                        )}

                        <button onClick={() => fetchHistory(task.id)} style={{ flex: 2, fontSize: '0.75rem', padding: '4px', cursor: 'pointer', borderRadius: '4px', border: '1px solid #ccc', backgroundColor: 'white' }}>History</button>
                    </div>
                )}
            </div>
        );
    };

    return (
        <div style={{ padding: '10px', background: '#f8f9fa', borderRadius: '8px' }}>
            {isOwner && (
                <div style={{ marginBottom: '15px' }}>
                    {!isAddingTask ? (
                        <button onClick={() => setIsAddingTask(true)} style={{ backgroundColor: '#1890ff', color: 'white', border: 'none', padding: '8px 16px', borderRadius: '4px', cursor: 'pointer' }}>+ Add New Task</button>
                    ) : (
                        <form onSubmit={handleCreateTask} style={{ background: 'white', padding: '15px', borderRadius: '6px', border: '1px solid #ddd' }}>
                            <input required placeholder="Task Title" value={newTaskForm.title} onChange={e => setNewTaskForm({ ...newTaskForm, title: e.target.value })} style={{ width: '100%', marginBottom: '10px' }} />
                            <textarea placeholder="Description" value={newTaskForm.description} onChange={e => setNewTaskForm({ ...newTaskForm, description: e.target.value })} style={{ width: '100%', marginBottom: '10px' }} />
                            <select value={newTaskForm.assigned_to} onChange={e => setNewTaskForm({ ...newTaskForm, assigned_to: e.target.value })} style={{ width: '100%', marginBottom: '10px' }}>
                                <option value="">Assign to...</option>
                                {members.map(m => <option key={m.id} value={m.id}>{m.email}</option>)}
                            </select>
                            <div style={{ display: 'flex', gap: '10px' }}>
                                <button type="submit" style={{ backgroundColor: '#52c41a', color: 'white', border: 'none', padding: '6px 12px', borderRadius: '4px' }}>Create</button>
                                <button type="button" onClick={() => setIsAddingTask(false)} style={{ padding: '6px 12px' }}>Cancel</button>
                            </div>
                        </form>
                    )}
                </div>
            )}

            <div style={{ display: 'flex', gap: '15px', overflowX: 'auto' }}>
                {['TODO', 'IN_PROGRESS', 'DONE'].map(status => (
                    <div key={status} style={{ flex: 1, minWidth: '250px', background: '#ebecf0', padding: '10px', borderRadius: '8px', minHeight: '400px' }}>
                        <h3 style={{ fontSize: '0.9rem', color: '#444', textTransform: 'uppercase', marginBottom: '10px', textAlign: 'center' }}>{status.replace('_', ' ')}</h3>
                        {tasks.filter(t => t.status === status).map(renderTaskCard)}
                    </div>
                ))}
            </div>

            {showHistoryId && (
                <div className="modal-overlay" style={{ position: 'fixed', top: 0, left: 0, right: 0, bottom: 0, backgroundColor: 'rgba(0,0,0,0.6)', display: 'flex', justifyContent: 'center', alignItems: 'center', zIndex: 1000 }}>
                    <div style={{ backgroundColor: 'white', padding: '20px', borderRadius: '8px', width: '400px', maxHeight: '70vh', overflowY: 'auto' }}>
                        <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '10px' }}>
                            <h3 style={{ margin: 0 }}>Task History</h3>
                            <button onClick={() => setShowHistoryId(null)} style={{ border: 'none', background: 'none', cursor: 'pointer', fontSize: '1.2rem' }}>&times;</button>
                        </div>
                        {history.length === 0 ? <p style={{ fontSize: '0.85rem', color: '#666' }}>No activity recorded yet.</p> : history.map((act) => (
                            <div key={act.id} style={{ fontSize: '0.85rem', padding: '10px 0', borderBottom: '1px solid #eee' }}>
                                <div style={{ marginBottom: '4px' }}><strong style={{ color: '#1890ff' }}>{act.user_email || "System"}</strong>: {formatActivityDetails(act.details)}</div>
                                <div style={{ color: '#999', fontSize: '0.7rem' }}>{act.created_at ? new Date(act.created_at).toLocaleString() : "Invalid Date"}</div>
                            </div>
                        ))}
                        <button onClick={() => setShowHistoryId(null)} style={{ marginTop: '15px', width: '100%', padding: '8px', cursor: 'pointer', borderRadius: '4px', border: '1px solid #ccc' }}>Close</button>
                    </div>
                </div>
            )}
        </div>
    );
};

export default TaskBoard;