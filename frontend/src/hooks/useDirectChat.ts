import { useState, useEffect, useCallback, useRef } from 'react';

export interface DirectMessage {
    id: number;
    sender_id: number;
    receiver_id?: number;
    content: string;
    created_at: string;
    sender_email?: string;
}

interface DirectChatResponse {
    type: "new_direct_message" | "error";
    data?: DirectMessage;
    error?: string;
}

export const useDirectChat = (token: string | null, otherUserId?: number) => {
    const [messages, setMessages] = useState<DirectMessage[]>([]);
    const [isConnected, setIsConnected] = useState(false);
    const socket = useRef<WebSocket | null>(null);

    useEffect(() => {
        if (!token || !otherUserId) return;

        // For DMs, we still connect to a project room (or could be a user room)
        // The backend sends to users directly via SendToUser, so any connection works
        const wsUrl = `ws://localhost:880/ws?token=${token}&projectId=0`; // projectId not needed for DMs

        if (socket.current) {
            socket.current.close();
        }

        const ws = new WebSocket(wsUrl);
        socket.current = ws;

        ws.onopen = () => {
            console.log(`✅ Connected for DM with user: ${otherUserId}`);
            setIsConnected(true);
        };

        ws.onmessage = (event) => {
            if (!event.data.startsWith('{')) {
                console.log("ℹ️ System:", event.data);
                return;
            }

            try {
                const response: DirectChatResponse = JSON.parse(event.data);
                if (response.type === "new_direct_message" && response.data) {
                    const msg = response.data;
                    // Only add if it's between current user and otherUserId
                    const currentUserId = Number(localStorage.getItem("userId"));
                    if ((msg.sender_id === currentUserId && msg.receiver_id === otherUserId) ||
                        (msg.sender_id === otherUserId && msg.receiver_id === currentUserId)) {
                        setMessages((prev) => [...prev, msg]);
                    }
                }
            } catch (err) {
                console.error("⚠️ Message Parse Error:", err);
            }
        };

        ws.onclose = () => {
            console.log("🔌 Disconnected from DM Hub");
            setIsConnected(false);
        };

        return () => {
            console.log("🧹 Cleaning up Direct Chat WebSocket...");
            ws.close();
            socket.current = null;
        };
    }, [token, otherUserId]);

    const sendMessage = useCallback((content: string, receiverId: number) => {
        if (socket.current?.readyState === WebSocket.OPEN) {
            const payload = {
                type: "direct_message",
                content: content,
                receiver_id: receiverId,
            };
            socket.current.send(JSON.stringify(payload));
        } else {
            console.error("🚫 Socket not open. State:", socket.current?.readyState);
        }
    }, []);

    return { messages, setMessages, sendMessage, isConnected };
};