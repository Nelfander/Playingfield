import { useState, useEffect, useCallback, useRef } from 'react';
import { WS_BASE_URL } from '../config/env';

export interface Message {
    id: number;
    sender_id: number;
    project_id?: number;
    receiver_id?: number;
    content: string;
    created_at: string;
    read_at?: string;
    sender_email?: string;
}

interface ChatResponse {
    type: "new_project_message" | "message_read" | "user_typing" | "error";
    data?: Message;
    message_id?: number;
    is_typing?: boolean;
    user_id?: number;
    email?: string; // <--- 1. Capture the email from Go
    project_id?: number;
    error?: string;
}

export const useChat = (token: string | null, projectId?: number) => {
    const [messages, setMessages] = useState<Message[]>([]);
    const [isConnected, setIsConnected] = useState(false);
    const [isTyping, setIsTyping] = useState(false);
    const [typingUserId, setTypingUserId] = useState<number | null>(null);
    const [typingUserEmail, setTypingUserEmail] = useState<string | null>(null); // 2. New State
    const socket = useRef<WebSocket | null>(null);

    useEffect(() => {
        if (!token || !projectId) return;

        const wsUrl = `${WS_BASE_URL}/ws?token=${token}&projectId=${projectId}`;
        if (socket.current) socket.current.close();

        const ws = new WebSocket(wsUrl);
        socket.current = ws;

        ws.onopen = () => setIsConnected(true);

        ws.onmessage = (event) => {
            try {
                const response: ChatResponse = JSON.parse(event.data);

                switch (response.type) {
                    case "new_project_message":
                        if (response.data?.project_id === projectId) {
                            setMessages((prev) => [...prev, response.data!]);
                        }
                        break;
                    case "message_read":
                        setMessages((prev) => prev.map(m =>
                            m.id === response.message_id ? { ...m, read_at: new Date().toISOString() } : m
                        ));
                        break;
                    case "user_typing":
                        if (response.project_id === projectId && response.project_id !== 0) {
                            setIsTyping(!!response.is_typing);
                            setTypingUserId(response.is_typing ? (response.user_id ?? null) : null);
                            // 3. Set the email directly from the server signal
                            setTypingUserEmail(response.is_typing ? (response.email ?? null) : null);
                        }
                        break;
                }
            } catch (err) {
                console.error("WS Parse Error:", err);
            }
        };

        ws.onclose = () => setIsConnected(false);

        return () => {
            if (socket.current) socket.current.close();
        };
    }, [token, projectId]);

    const sendMessage = useCallback((content: string, type: 'project_chat' | 'direct_message', targetId: number) => {
        if (socket.current?.readyState === WebSocket.OPEN) {
            socket.current.send(JSON.stringify({
                type,
                content,
                project_id: type === 'project_chat' ? targetId : undefined,
                receiver_id: type === 'direct_message' ? targetId : undefined,
            }));
        }
    }, []);

    const sendReadReceipt = useCallback((messageId: number) => {
        if (socket.current?.readyState === WebSocket.OPEN) {
            socket.current.send(JSON.stringify({
                type: "read_receipt",
                message_id: messageId,
                project_id: projectId
            }));
        }
    }, [projectId]);

    const sendTypingStatus = useCallback((typing: boolean) => {
        if (socket.current?.readyState === WebSocket.OPEN) {
            socket.current.send(JSON.stringify({
                type: "typing",
                project_id: projectId,
                is_typing: typing
            }));
        }
    }, [projectId]);

    return {
        messages,
        setMessages,
        sendMessage,
        sendReadReceipt,
        sendTypingStatus,
        isConnected,
        isTyping,
        typingUserId,
        typingUserEmail // 4. Export the email
    };
};