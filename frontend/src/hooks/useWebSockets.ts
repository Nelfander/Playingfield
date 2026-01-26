import { useEffect, useRef } from 'react';

type OnProjectDeleted = (id: number) => void;
type OnUserAdded = (projectId: number, userId: number, role: string) => void;
type OnProjectCreated = () => void;
type OnUserRemoved = (projectId: number, userId: number) => void;
type OnProjectUpdated = (id: number) => void;

// Task Types
type OnTaskCreated = (projectId: number) => void;
type OnTaskUpdated = (projectId: number, taskId: number) => void;
type OnTaskDeleted = (projectId: number, taskId: number) => void;

export const useWebSockets = (
    token: string | null,
    onProjectDeleted: OnProjectDeleted,
    onUserAdded: OnUserAdded,
    onProjectCreated: OnProjectCreated,
    onUserRemoved: OnUserRemoved,
    onProjectUpdated: OnProjectUpdated,
    // Task Parameters
    onTaskCreated: OnTaskCreated,
    onTaskUpdated: OnTaskUpdated,
    onTaskDeleted: OnTaskDeleted
) => {
    const callbacks = useRef({
        onProjectDeleted,
        onUserAdded,
        onProjectCreated,
        onUserRemoved,
        onProjectUpdated,
        onTaskCreated,
        onTaskUpdated,
        onTaskDeleted
    });

    useEffect(() => {
        callbacks.current = {
            onProjectDeleted,
            onUserAdded,
            onProjectCreated,
            onUserRemoved,
            onProjectUpdated,
            onTaskCreated,
            onTaskUpdated,
            onTaskDeleted
        };
    }, [
        onProjectDeleted,
        onUserAdded,
        onProjectCreated,
        onUserRemoved,
        onProjectUpdated,
        onTaskCreated,
        onTaskUpdated,
        onTaskDeleted
    ]);

    useEffect(() => {
        if (!token) return;

        const socket = new WebSocket(`ws://localhost:880/ws?token=${token}`);

        socket.onmessage = (event: MessageEvent) => {
            const data: string = event.data;

            // --- Project & User Logic ---
            if (data === "PROJECT_CREATED") {
                callbacks.current.onProjectCreated();
            }
            else if (data.startsWith("PROJECT_UPDATED:")) {
                const id = parseInt(data.split(":")[1], 10);
                if (!isNaN(id)) {
                    callbacks.current.onProjectUpdated(id);
                }
            }
            else if (data.startsWith("PROJECT_DELETED:")) {
                const id = parseInt(data.split(":")[1], 10);
                callbacks.current.onProjectDeleted(id);
            }
            else if (data.startsWith("USER_ADDED:")) {
                const parts = data.split(":");
                const projectId = parseInt(parts[1], 10);
                const userId = parseInt(parts[2], 10);
                const role = parts[3];

                if (!isNaN(projectId) && !isNaN(userId)) {
                    callbacks.current.onUserAdded(projectId, userId, role);
                }
            }
            else if (data.startsWith("USER_REMOVED:")) {
                const parts = data.split(":");
                const projectId = parseInt(parts[1], 10);
                const userId = parseInt(parts[2], 10);

                if (!isNaN(projectId) && !isNaN(userId)) {
                    callbacks.current.onUserRemoved(projectId, userId);
                }
            }
            // --- Task Logic ---
            else if (data.startsWith("TASK_CREATED:")) {
                const projectId = parseInt(data.split(":")[1], 10);
                if (!isNaN(projectId)) {
                    callbacks.current.onTaskCreated(projectId);
                }
            }
            else if (data.startsWith("TASK_UPDATED:")) {
                const parts = data.split(":");
                const projectId = parseInt(parts[1], 10);
                const taskId = parseInt(parts[2], 10);
                if (!isNaN(projectId) && !isNaN(taskId)) {
                    callbacks.current.onTaskUpdated(projectId, taskId);
                }
            }
            else if (data.startsWith("TASK_DELETED:")) {
                const parts = data.split(":");
                const projectId = parseInt(parts[1], 10);
                const taskId = parseInt(parts[2], 10);
                if (!isNaN(projectId) && !isNaN(taskId)) {
                    callbacks.current.onTaskDeleted(projectId, taskId);
                }
            }
        };

        socket.onclose = () => console.log("🔌 WebSocket: Global Disconnected");

        return () => {
            if (socket.readyState === WebSocket.OPEN || socket.readyState === WebSocket.CONNECTING) {
                socket.close();
            }
        };
    }, [token]);
};