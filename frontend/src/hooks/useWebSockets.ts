import { useEffect, useRef } from 'react';

// Project Types
type OnProjectDeleted = (id: number) => void;
type OnUserAdded = (projectId: number, userId: number, role: string) => void;
type OnProjectCreated = () => void;
type OnUserRemoved = (projectId: number, userId: number) => void;
type OnProjectUpdated = (id: number) => void;

// Task Types
type OnTaskCreated = (projectId: number) => void;
type OnTaskUpdated = (projectId: number, taskId: number) => void;
type OnTaskDeleted = (projectId: number, taskId: number) => void;

// --- GLOBAL USER TYPES ---
type OnUserScrubbed = (userId: number) => void;
type OnUserStatusUpdated = (userId: number, newStatus: string) => void;
type OnUserCreated = (userId: number) => void; // Semantically correct for registration

export const useWebSockets = (
    token: string | null,
    onProjectDeleted: OnProjectDeleted,
    onUserAdded: OnUserAdded,
    onProjectCreated: OnProjectCreated,
    onUserRemoved: OnUserRemoved,
    onProjectUpdated: OnProjectUpdated,
    onTaskCreated: OnTaskCreated,
    onTaskUpdated: OnTaskUpdated,
    onTaskDeleted: OnTaskDeleted,
    onUserScrubbed: OnUserScrubbed,
    onUserStatusUpdated: OnUserStatusUpdated,
    onUserCreated: OnUserCreated
) => {
    // We use a Ref for callbacks so the WebSocket effect doesn't 
    // restart every time a function reference changes in the parent.
    const callbacks = useRef({
        onProjectDeleted,
        onUserAdded,
        onProjectCreated,
        onUserRemoved,
        onProjectUpdated,
        onTaskCreated,
        onTaskUpdated,
        onTaskDeleted,
        onUserScrubbed,
        onUserStatusUpdated,
        onUserCreated
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
            onTaskDeleted,
            onUserScrubbed,
            onUserStatusUpdated,
            onUserCreated
        };
    }, [
        onProjectDeleted,
        onUserAdded,
        onProjectCreated,
        onUserRemoved,
        onProjectUpdated,
        onTaskCreated,
        onTaskUpdated,
        onTaskDeleted,
        onUserScrubbed,
        onUserStatusUpdated,
        onUserCreated
    ]);

    useEffect(() => {
        if (!token) return;

        const socket = new WebSocket(`ws://localhost:880/ws?token=${token}`);

        socket.onmessage = (event: MessageEvent) => {
            const data: string = event.data;
            const parts = data.split(":");
            const action = parts[0];

            switch (action) {
                // --- Global User Logic ---
                case "USER_CREATED": {
                    const userId = parseInt(parts[1], 10);
                    if (!isNaN(userId)) callbacks.current.onUserCreated(userId);
                    break;
                }
                case "USER_SCRUBBED": {
                    const userId = parseInt(parts[1], 10);
                    if (!isNaN(userId)) callbacks.current.onUserScrubbed(userId);
                    break;
                }
                case "USER_UPDATED": {
                    const userId = parseInt(parts[1], 10);
                    const status = parts[2];
                    if (!isNaN(userId)) callbacks.current.onUserStatusUpdated(userId, status);
                    break;
                }

                // --- Project Logic ---
                case "PROJECT_CREATED":
                    callbacks.current.onProjectCreated();
                    break;

                case "PROJECT_UPDATED": {
                    const id = parseInt(parts[1], 10);
                    if (!isNaN(id)) callbacks.current.onProjectUpdated(id);
                    break;
                }

                case "PROJECT_DELETED": {
                    const id = parseInt(parts[1], 10);
                    if (!isNaN(id)) callbacks.current.onProjectDeleted(id);
                    break;
                }

                case "USER_ADDED": {
                    const projectId = parseInt(parts[1], 10);
                    const userId = parseInt(parts[2], 10);
                    const role = parts[3];
                    if (!isNaN(projectId) && !isNaN(userId)) {
                        callbacks.current.onUserAdded(projectId, userId, role);
                    }
                    break;
                }

                case "USER_REMOVED": {
                    const projectId = parseInt(parts[1], 10);
                    const userId = parseInt(parts[2], 10);
                    if (!isNaN(projectId) && !isNaN(userId)) {
                        callbacks.current.onUserRemoved(projectId, userId);
                    }
                    break;
                }

                // --- Task Logic ---
                case "TASK_CREATED": {
                    const projectId = parseInt(parts[1], 10);
                    if (!isNaN(projectId)) callbacks.current.onTaskCreated(projectId);
                    break;
                }

                case "TASK_UPDATED": {
                    const projectId = parseInt(parts[1], 10);
                    const taskId = parseInt(parts[2], 10);
                    if (!isNaN(projectId) && !isNaN(taskId)) {
                        callbacks.current.onTaskUpdated(projectId, taskId);
                    }
                    break;
                }

                case "TASK_DELETED": {
                    const projectId = parseInt(parts[1], 10);
                    const taskId = parseInt(parts[2], 10);
                    if (!isNaN(projectId) && !isNaN(taskId)) {
                        callbacks.current.onTaskDeleted(projectId, taskId);
                    }
                    break;
                }

                default:
                    console.warn("Unknown WebSocket action received:", action);
            }
        };

        socket.onopen = () => console.log("🚀 WebSocket: Connected to Playingfield Hub");
        socket.onclose = () => console.log("🔌 WebSocket: Global Disconnected");
        socket.onerror = (error) => console.error("❌ WebSocket Error:", error);

        return () => {
            if (socket.readyState === WebSocket.OPEN || socket.readyState === WebSocket.CONNECTING) {
                socket.close();
            }
        };
    }, [token]);
};