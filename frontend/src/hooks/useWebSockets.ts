import { useEffect } from 'react';

type OnProjectDeleted = (id: number) => void;
type OnUserAdded = (projectId: number, userId: number, role: string) => void;
type OnProjectCreated = () => void;
type OnUserRemoved = (projectId: number, userId: number) => void;

export const useWebSockets = (
    token: string | null,
    onProjectDeleted: OnProjectDeleted,
    onUserAdded: OnUserAdded,
    onProjectCreated: OnProjectCreated,
    onUserRemoved: OnUserRemoved // 
) => {
    useEffect(() => {
        if (!token) return;

        const socket = new WebSocket(`ws://localhost:880/ws?token=${token}`);

        socket.onopen = () => console.log("WebSocket: Connected");

        socket.onmessage = (event: MessageEvent) => {
            const data: string = event.data;

            // Handle Project Creation
            if (data === "PROJECT_CREATED") {
                onProjectCreated();
            }
            // Handle Project Deletion
            else if (data.startsWith("PROJECT_DELETED:")) {
                const id = parseInt(data.split(":")[1], 10);
                onProjectDeleted(id);
            }
            // Handle User Addition
            else if (data.startsWith("USER_ADDED:")) {
                const parts = data.split(":");
                const projectId = parseInt(parts[1], 10);
                const userId = parseInt(parts[2], 10);
                const role = parts[3];

                if (!isNaN(projectId) && !isNaN(userId)) {
                    onUserAdded(projectId, userId, role);
                }
            }
            //  Handle User Removal
            else if (data.startsWith("USER_REMOVED:")) {
                const parts = data.split(":");
                const projectId = parseInt(parts[1], 10);
                const userId = parseInt(parts[2], 10);

                if (!isNaN(projectId) && !isNaN(userId)) {
                    onUserRemoved(projectId, userId);
                }
            }
        };

        socket.onclose = () => console.log("WebSocket: Disconnected");

        return () => socket.close();
    }, [token, onProjectDeleted, onUserAdded, onProjectCreated, onUserRemoved]); // Added onUserRemoved to dependencies
};