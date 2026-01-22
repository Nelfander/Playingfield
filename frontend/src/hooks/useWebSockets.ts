import { useEffect, useRef } from 'react';

type OnProjectDeleted = (id: number) => void;
type OnUserAdded = (projectId: number, userId: number, role: string) => void;
type OnProjectCreated = () => void;
type OnUserRemoved = (projectId: number, userId: number) => void;
// 1. Add the new type
type OnProjectUpdated = (id: number) => void;

export const useWebSockets = (
    token: string | null,
    onProjectDeleted: OnProjectDeleted,
    onUserAdded: OnUserAdded,
    onProjectCreated: OnProjectCreated,
    onUserRemoved: OnUserRemoved,
    // 2. Add to parameters
    onProjectUpdated: OnProjectUpdated
) => {
    const callbacks = useRef({
        onProjectDeleted,
        onUserAdded,
        onProjectCreated,
        onUserRemoved,
        onProjectUpdated
    });

    useEffect(() => {
        callbacks.current = {
            onProjectDeleted,
            onUserAdded,
            onProjectCreated,
            onUserRemoved,
            onProjectUpdated
        };
    }, [onProjectDeleted, onUserAdded, onProjectCreated, onUserRemoved, onProjectUpdated]);

    useEffect(() => {
        if (!token) return;

        const socket = new WebSocket(`ws://localhost:880/ws?token=${token}`);

        socket.onmessage = (event: MessageEvent) => {
            const data: string = event.data;

            if (data === "PROJECT_CREATED") {
                callbacks.current.onProjectCreated();
            }
            // 3. Logic to handle the update signal
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
        };

        socket.onclose = () => console.log("🔌 WebSocket: Global Disconnected");

        return () => {
            if (socket.readyState === WebSocket.OPEN || socket.readyState === WebSocket.CONNECTING) {
                socket.close();
            }
        };
    }, [token]);
};