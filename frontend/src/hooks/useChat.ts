import { useState, useEffect, useCallback, useRef } from 'react';

export interface Message {
    id: number;
    sender_id: number;
    project_id?: number;
    receiver_id?: number;
    content: string;
    created_at: string;
}

interface ChatResponse {
    type: "new_project_message" | "new_direct_message" | "error";
    data?: Message;
    error?: string;
}

export const useChat = (token: string | null, projectId?: number) => {
    const [messages, setMessages] = useState<Message[]>([]);
    const [isConnected, setIsConnected] = useState(false);
    const socket = useRef<WebSocket | null>(null);

    useEffect(() => {
        // Only connect if we have BOTH a token and a projectId
        if (!token || !projectId) return;

        const wsUrl = `ws://localhost:880/ws?token=${token}&projectId=${projectId}`;

        // Close existing socket if it somehow exists before opening a new one
        if (socket.current) {
            socket.current.close();
        }

        const ws = new WebSocket(wsUrl);
        socket.current = ws;

        ws.onopen = () => {
            console.log(`✅ Connected to Project Room: ${projectId}`);
            setIsConnected(true);
        };

        ws.onmessage = (event) => {
            if (!event.data.startsWith('{')) {
                console.log("ℹ️ System:", event.data);
                return;
            }

            try {
                const response: ChatResponse = JSON.parse(event.data);
                if (response.type === "new_project_message" && response.data) {
                    // Only add message if it belongs to the current project
                    if (response.data.project_id === projectId) {
                        setMessages((prev) => [...prev, response.data!]);
                    }
                }
                // Handle errors or DMs here if needed
            } catch (err) {
                console.error("⚠️ Message Parse Error:", err);
            }
        };

        ws.onclose = () => {
            console.log("🔌 Disconnected from Hub");
            setIsConnected(false);
        };

        //  Cleanup Function
        return () => {
            console.log("🧹 Cleaning up WebSocket...");
            ws.close();
            socket.current = null;
        };
    }, [token, projectId]); // Effect reruns if project switches

    const sendMessage = useCallback((content: string, type: 'project_chat' | 'direct_message', targetId: number) => {
        if (socket.current?.readyState === WebSocket.OPEN) {
            const payload = {
                type: type,
                content: content,
                project_id: type === 'project_chat' ? targetId : undefined,
                receiver_id: type === 'direct_message' ? targetId : undefined,
            };
            socket.current.send(JSON.stringify(payload));
        } else {
            console.error("🚫 Socket not open. State:", socket.current?.readyState);
        }
    }, []);

    return { messages, setMessages, sendMessage, isConnected };
};