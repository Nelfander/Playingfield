import { useEffect, useRef } from 'react';
import { WS_BASE_URL } from '../config/env';

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

// Global User Types
type OnUserScrubbed = (userId: number) => void;
type OnUserStatusUpdated = (userId: number, newStatus: string) => void;
type OnUserCreated = (userId: number) => void;

// --- NEW CHAT TYPES ---
type OnMessageCreated = (projectId: number, senderId: number) => void;
type OnDMCreated = (senderId: number) => void;

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
    onUserCreated: OnUserCreated,
    onMessageCreated: OnMessageCreated,
    onDMCreated: OnDMCreated
) => {
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
        onUserCreated,
        onMessageCreated,
        onDMCreated
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
            onUserCreated,
            onMessageCreated,
            onDMCreated
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
        onUserCreated,
        onMessageCreated,
        onDMCreated
    ]);

    useEffect(() => {
        if (!token) return;

        const socket = new WebSocket(`${WS_BASE_URL}/ws?token=${token}`);

        socket.onmessage = (event: MessageEvent) => {
            const data: string = event.data;
            console.log("raw WS data received:", data);

            // 1. IMPROVED JSON PARSING
            try {
                const jsonData = JSON.parse(data);

                // Extract from the 'data' sub-object if it exists (per your screenshot)
                const payload = jsonData.data || jsonData;
                const type = jsonData.type || payload.type;

                // Group Message Logic
                // We check payload.project_id because it might be inside the 'data' wrapper
                if (payload.project_id && Number(payload.project_id) > 0) {
                    callbacks.current.onMessageCreated(
                        Number(payload.project_id),
                        Number(payload.sender_id || payload.user_id)
                    );
                    return;
                }

                // DM Logic
                if (type === "new_direct_message") {
                    callbacks.current.onDMCreated(Number(payload.sender_id));
                    return;
                }

                // Ignore typing indicators
                if (type === "user_typing") return;

            } catch (e) {
                // Not JSON data, proceed to fallback logic
            }

            // 2. FALLBACK TO STRING PARSING (e.g., "DM_CREATED:2")
            const parts = data.split(":");
            const action = parts[0];

            switch (action) {
                case "MESSAGE_CREATED": {
                    const projectId = parseInt(parts[1], 10);
                    const senderId = parseInt(parts[2], 10);
                    if (!isNaN(projectId)) callbacks.current.onMessageCreated(projectId, senderId);
                    break;
                }
                case "DM_CREATED": {
                    const senderId = parseInt(parts[1], 10);
                    if (!isNaN(senderId)) callbacks.current.onDMCreated(senderId);
                    break;
                }
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
                case "ATTACHMENT_CREATED":
                case "ATTACHMENT_DELETED": {
                    const projectId = parseInt(parts[1], 10);
                    const taskId = parseInt(parts[2], 10);
                    if (!isNaN(projectId) && !isNaN(taskId)) {
                        callbacks.current.onTaskUpdated(projectId, taskId);
                    }
                    break;
                }
                default:
                    if (data.trim() && !data.startsWith("{")) {
                        console.warn("Unknown WebSocket action received:", action);
                    }
            }
        };

        socket.onopen = () => console.log("🚀 WebSocket Connected");
        socket.onclose = () => console.log("🔌 WebSocket Disconnected");
        socket.onerror = (error) => console.error("❌ WebSocket Error:", error);

        return () => {
            if (socket.readyState === WebSocket.OPEN || socket.readyState === WebSocket.CONNECTING) {
                socket.close();
            }
        };
    }, [token]);
};