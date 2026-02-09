import { useState, useEffect, useCallback, useRef } from 'react';

export interface DirectMessage {
    id: number;
    sender_id: number;
    receiver_id?: number;
    content: string;
    created_at: string;
    read_at?: string; // Important for "Seen" status
    sender_email?: string;
}

interface DirectChatResponse {
    type: "new_direct_message" | "user_typing" | "message_read" | "error";
    data?: DirectMessage;
    user_id?: number;
    sender_id?: number;
    receiver_id?: number;
    message_id?: number;
    is_typing?: boolean;
    project_id?: number;
}

export const useDirectChat = (token: string | null, otherUserId?: number) => {
    const [messages, setMessages] = useState<DirectMessage[]>([]);
    const [isConnected, setIsConnected] = useState(false);
    const [isTyping, setIsTyping] = useState(false);
    const [typingUserId, setTypingUserId] = useState<number | null>(null);
    const socket = useRef<WebSocket | null>(null);

    useEffect(() => {
        if (!token || !otherUserId) return;

        // Using projectId=0 to signify Direct Messaging mode to the server
        const wsUrl = `ws://localhost:880/ws?token=${token}&projectId=0`;

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
            try {
                const response: DirectChatResponse = JSON.parse(event.data);

                switch (response.type) {
                    case "new_direct_message":
                        if (response.data) {
                            const msg = response.data;
                            const currentUserId = Number(localStorage.getItem("userId"));

                            // Only add if message belongs to this specific 1-on-1 chat
                            if ((msg.sender_id === currentUserId && msg.receiver_id === otherUserId) ||
                                (msg.sender_id === otherUserId && msg.receiver_id === currentUserId)) {

                                setMessages((prev) => [...prev, msg]);

                                // Reset typing state if they just sent a message
                                if (Number(msg.sender_id) === Number(otherUserId)) {
                                    setIsTyping(false);
                                    setTypingUserId(null);
                                }
                            }
                        }
                        break;

                    case "message_read":
                        setMessages((prev) => prev.map(m =>
                            m.id === response.message_id ? { ...m, read_at: new Date().toISOString() } : m
                        ));
                        break;

                    case "user_typing":
                        // 1. Extract the user who is typing from the Go backend payload
                        const sid = response.sender_id || response.user_id;
                        const typingId = Number(sid);

                        // 2. Only show it if the typing person is the one we are chatting with
                        if (typingId === Number(otherUserId)) {
                            console.log("✅ Match! User", typingId, "is typing.");
                            setIsTyping(response.is_typing === true);
                            setTypingUserId(response.is_typing ? typingId : null);
                        }
                        break;
                }
            } catch (err) {
                console.error("WS Parse Error:", err);
            }
        };

        ws.onclose = () => {
            setIsConnected(false);
        };

        return () => {
            if (socket.current) socket.current.close();
        };
    }, [token, otherUserId]);

    const sendMessage = useCallback((content: string, receiverId: number) => {
        if (socket.current?.readyState === WebSocket.OPEN) {
            socket.current.send(JSON.stringify({
                type: "direct_message",
                content: content,
                receiver_id: Number(receiverId),
            }));
        }
    }, []);

    const sendTypingStatus = useCallback((typing: boolean, receiverId: number) => {
        console.log("1. sendTypingStatus called. Typing:", typing, "To:", receiverId);

        if (!socket.current) {
            console.error("❌ No socket object found!");
            return;
        }

        if (socket.current.readyState !== WebSocket.OPEN) {
            console.error("❌ WebSocket is NOT open. Current state:", socket.current.readyState);
            return;
        }

        const payload = {
            type: "typing",
            is_typing: typing,
            receiver_id: Number(receiverId),
            project_id: 0,
            content: ""
        };

        console.log("2. 📤 Sending Payload to Server:", JSON.stringify(payload));
        socket.current.send(JSON.stringify(payload));
    }, []);

    const sendReadReceipt = useCallback((messageId: number, senderId: number) => {
        if (socket.current?.readyState === WebSocket.OPEN) {
            socket.current.send(JSON.stringify({
                type: "read_receipt",
                message_id: messageId,
                receiver_id: Number(senderId),
                project_id: 0
            }));
        }
    }, []);

    return {
        messages,
        setMessages,
        sendMessage,
        sendTypingStatus,
        sendReadReceipt,
        isConnected,
        isTyping,
        typingUserId
    };
};