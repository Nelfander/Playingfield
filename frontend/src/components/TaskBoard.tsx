import React, { useState, useEffect } from 'react';

interface Task {
    id: number;
    project_id: number;
    title: string;
    description: string;
    status: 'TODO' | 'IN_PROGRESS' | 'DONE';
    assigned_to: number | null;
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
    refreshTick: number; // This is now a simple number that increments on any WS signal
    isOwner: boolean;
    members: User[];
}

const TaskBoard: React.FC<TaskBoardProps> = ({ projectId, refreshTick, isOwner, members }) => {
    const [tasks, setTasks] = useState<Task[]>([]);
    const [updatingTaskId, setUpdatingTaskId] = useState<number | null>(null);
    const [commitMessage, setCommitMessage] = useState("");
    const [newStatus, setNewStatus] = useState<'TODO' | 'IN_PROGRESS' | 'DONE'>('TODO');

    const [isAddingTask, setIsAddingTask] = useState(false);
    const [newTaskForm, setNewTaskForm] = useState({ title: '', description: '', assigned_to: '' });

    const [history, setHistory] = useState<TaskActivity[]>([]);
    const [showHistoryId, setShowHistoryId] = useState<number | null>(null);

    const token = localStorage.getItem('token');
    const currentUserId = token ? JSON.parse(atob(token.split('.')[1])).user_id : null;

    // --- LIVE SYNC LOGIC ---
    useEffect(() => {
        // We add a tiny 150ms delay to allow the Database (Neon) to finish 
        // the write operation before we fetch the new data.
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
            if (Array.isArray(data)) setTasks(data); else setTasks([]);
        } catch (err) {
            console.error("Fetch tasks error:", err);
            setTasks([]);
        }
    };

    // HELPER: Replaces "user 5" with the actual email from the members list
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
                // Note: The WebSocket will trigger the global refresh for others
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

    const renderTaskCard = (task: Task) => {
        const canUpdate = isOwner || (task.assigned_to !== null && task.assigned_to === currentUserId);

        return (
            <div key={task.id} className="task-card" style={{ border: '1px solid #ddd', padding: '12px', marginBottom: '10px', borderRadius: '6px', backgroundColor: 'white', boxShadow: '0 2px 4px rgba(0,0,0,0.05)' }}>
                <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                    <h4 style={{ margin: '0 0 5px 0', fontSize: '1rem' }}>{task.title}</h4>
                    {isOwner && (
                        <button onClick={() => handleDeleteTask(task.id)} style={{ border: 'none', background: 'none', color: '#ff4d4f', cursor: 'pointer', fontWeight: 'bold' }}>&times;</button>
                    )}
                </div>
                <p style={{ fontSize: '0.8rem', color: '#555', marginBottom: '8px' }}>{task.description}</p>
                <div style={{ fontSize: '0.7rem', color: '#888' }}>
                    👤 {members.find(m => m.id === task.assigned_to)?.email || "Unassigned"}
                </div>

                {updatingTaskId === task.id ? (
                    <div style={{ marginTop: '10px', borderTop: '1px solid #eee', paddingTop: '10px' }}>
                        <textarea
                            value={commitMessage}
                            onChange={e => setCommitMessage(e.target.value)}
                            placeholder="Commit message..."
                            style={{ width: '100%', fontSize: '0.8rem', minHeight: '50px', marginBottom: '5px' }}
                        />

                        <label style={{ fontSize: '0.7rem', color: '#666' }}>Status:</label>
                        <select value={newStatus} onChange={e => setNewStatus(e.target.value as any)} style={{ width: '100%', marginBottom: '10px' }}>
                            <option value="TODO">To Do</option>
                            <option value="IN_PROGRESS">In Progress</option>
                            <option value="DONE">Done</option>
                        </select>

                        {isOwner && (
                            <>
                                <label style={{ fontSize: '0.7rem', color: '#666' }}>Assignee:</label>
                                <select
                                    value={task.assigned_to || ""}
                                    onChange={e => {
                                        const val = e.target.value ? Number(e.target.value) : null;
                                        setTasks(tasks.map(t => t.id === task.id ? { ...t, assigned_to: val } : t));
                                    }}
                                    style={{ width: '100%', marginBottom: '10px' }}
                                >
                                    <option value="">Unassigned</option>
                                    {members.map(m => <option key={m.id} value={m.id}>{m.email}</option>)}
                                </select>
                            </>
                        )}

                        <div style={{ display: 'flex', gap: '5px' }}>
                            <button onClick={() => handleUpdateTask(task.id)} style={{ flex: 1, backgroundColor: '#52c41a', color: 'white', border: 'none', padding: '5px', borderRadius: '4px', cursor: 'pointer' }}>Save</button>
                            <button onClick={() => setUpdatingTaskId(null)} style={{ flex: 1, backgroundColor: '#ff4d4f', color: 'white', border: 'none', padding: '5px', borderRadius: '4px', cursor: 'pointer' }}>Cancel</button>
                        </div>
                    </div>
                ) : (
                    <div style={{ display: 'flex', gap: '5px', marginTop: '10px' }}>
                        {canUpdate && (
                            <button onClick={() => { setUpdatingTaskId(task.id); setNewStatus(task.status); }} style={{ flex: 1, fontSize: '0.75rem', padding: '4px', cursor: 'pointer', backgroundColor: '#1890ff', color: 'white', border: 'none', borderRadius: '4px' }}>Update</button>
                        )}
                        <button onClick={() => fetchHistory(task.id)} style={{ flex: 1, fontSize: '0.75rem', padding: '4px', cursor: 'pointer' }}>History</button>
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
                        <button onClick={() => setIsAddingTask(true)} style={{ backgroundColor: '#1890ff', color: 'white', border: 'none', padding: '8px 16px', borderRadius: '4px', cursor: 'pointer' }}>
                            + Add New Task
                        </button>
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
                        <h3 style={{ fontSize: '0.9rem', color: '#444', textTransform: 'uppercase', marginBottom: '10px', textAlign: 'center' }}>
                            {status.replace('_', ' ')}
                        </h3>
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
                        {history.length === 0 ? (
                            <p style={{ fontSize: '0.85rem', color: '#666' }}>No activity recorded yet.</p>
                        ) : (
                            history.map((act) => (
                                <div key={act.id} style={{ fontSize: '0.85rem', padding: '10px 0', borderBottom: '1px solid #eee' }}>
                                    <div style={{ marginBottom: '4px' }}>
                                        <strong style={{ color: '#1890ff' }}>{act.user_email || "System"}</strong>: {formatActivityDetails(act.details)}
                                    </div>
                                    <div style={{ color: '#999', fontSize: '0.7rem' }}>
                                        {act.created_at ? new Date(act.created_at).toLocaleString() : "Invalid Date"}
                                    </div>
                                </div>
                            ))
                        )}
                        <button onClick={() => setShowHistoryId(null)} style={{ marginTop: '15px', width: '100%', padding: '8px', cursor: 'pointer', borderRadius: '4px', border: '1px solid #ccc' }}>Close</button>
                    </div>
                </div>
            )}
        </div>
    );
};

export default TaskBoard;