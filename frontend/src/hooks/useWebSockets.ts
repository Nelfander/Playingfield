import { useEffect, useRef } from 'react';

type OnProjectDeleted = (id: number) => void;
type OnUserAdded = (projectId: number, userId: number, role: string) => void;
type OnProjectCreated = () => void;
type OnUserRemoved = (projectId: number, userId: number) => void;

export const useWebSockets = (
    token: string | null,
    onProjectDeleted: OnProjectDeleted,
    onUserAdded: OnUserAdded,
    onProjectCreated: OnProjectCreated,
    onUserRemoved: OnUserRemoved
) => {
    // Use refs for callbacks to prevent the effect from re-running 
    // every time the parent component re-renders.
    const callbacks = useRef({
        onProjectDeleted,
        onUserAdded,
        onProjectCreated,
        onUserRemoved
    });

    // Update refs whenever the props change
    useEffect(() => {
        callbacks.current = {
            onProjectDeleted,
            onUserAdded,
            onProjectCreated,
            onUserRemoved
        };
    }, [onProjectDeleted, onUserAdded, onProjectCreated, onUserRemoved]);

    useEffect(() => {
        if (!token) return;

        console.log("WebSocket: Attempting global connection...");
        const socket = new WebSocket(`ws://localhost:880/ws?token=${token}`);

        socket.onopen = () => console.log("✅ WebSocket: Global Connection Established");

        socket.onmessage = (event: MessageEvent) => {
            const data: string = event.data;

            if (data === "PROJECT_CREATED") {
                callbacks.current.onProjectCreated();
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

        // CLEANUP: This is the most important part to stop the spam.
        return () => {
            if (socket.readyState === WebSocket.OPEN || socket.readyState === WebSocket.CONNECTING) {
                socket.close();
            }
        };
    }, [token]); // Only re-run if token changes (login/logout)
};