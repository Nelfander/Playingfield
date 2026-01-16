import React, { useState, useEffect, useRef } from 'react';
import { useChat } from '../hooks/useChat';

interface Message {
    id: number;
    sender_id: number;
    sender_email?: string;
    content: string;
    created_at?: string;
}

interface ChatBoxProps {
    projectId: number;
    token: string;
}

export const ChatBox: React.FC<ChatBoxProps> = ({ projectId, token }) => {
    const { messages, setMessages, sendMessage, isConnected } = useChat(token, projectId);
    const [inputValue, setInputValue] = useState("");
    // NEW STEP: State to hold the project name
    const [projectName, setProjectName] = useState<string>("");

    const messagesEndRef = useRef<HTMLDivElement>(null);
    const currentUserId = Number(localStorage.getItem("userId"));

    // NEW STEP: Fetch project details (Name) on load
    useEffect(() => {
        const fetchProjectDetails = async () => {
            try {
                const response = await fetch(`http://localhost:880/projects/${projectId}`, {
                    headers: { 'Authorization': `Bearer ${token}` }
                });
                if (response.ok) {
                    const project = await response.json();
                    setProjectName(project.name);
                }
            } catch (err) {
                console.error("Failed to fetch project details:", err);
            }
        };

        if (projectId && token) fetchProjectDetails();
    }, [projectId, token]);

    useEffect(() => {
        const fetchHistory = async () => {
            try {
                const response = await fetch(`http://localhost:880/projects/${projectId}/messages`, {
                    headers: { 'Authorization': `Bearer ${token}` }
                });
                if (response.ok) {
                    const history = await response.json();
                    setMessages(history || []);
                }
            } catch (err) {
                console.error("Failed to load history:", err);
                setMessages([]);
            }
        };

        if (projectId && token) fetchHistory();
    }, [projectId, token, setMessages]);

    useEffect(() => {
        messagesEndRef.current?.scrollIntoView({ behavior: "smooth" });
    }, [messages]);

    const handleSend = () => {
        if (!inputValue.trim()) return;
        sendMessage(inputValue, 'project_chat', projectId);
        setInputValue("");
    };

    const formatTime = (dateStr?: string) => {
        if (!dateStr) return "";
        const date = new Date(dateStr);
        return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
    };

    return (
        <div style={styles.container}>
            <div style={styles.header}>

                <span>{projectName ? projectName : `Project ${projectId}`} Chat</span>
                <span style={{ color: isConnected ? '#4caf50' : '#f44336' }}>
                    {isConnected ? ' ● Online' : ' ● Offline'}
                </span>
            </div>

            <div style={styles.messageList}>
                {(messages || []).map((m, i) => {
                    const isMe = Number(m.sender_id) === currentUserId;
                    const time = formatTime(m.created_at);

                    return (
                        <div
                            key={m.id || i}
                            style={{
                                ...styles.messageItem,
                                alignSelf: isMe ? 'flex-end' : 'flex-start',
                                backgroundColor: isMe ? '#007bff' : '#f1f1f1',
                                color: isMe ? '#fff' : '#000',
                            }}
                        >
                            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'baseline', gap: '10px' }}>
                                <small style={{
                                    ...styles.sender,
                                    color: isMe ? '#e0e0e0' : '#888'
                                }}>
                                    {isMe ? "Me" : m.sender_email || `User ${m.sender_id}`}
                                </small>
                                {time && (
                                    <small style={{ fontSize: '0.6rem', color: isMe ? '#ccc' : '#999' }}>
                                        {time}
                                    </small>
                                )}
                            </div>
                            <div style={{ marginTop: '2px' }}>{m.content}</div>
                        </div>
                    );
                })}
                <div ref={messagesEndRef} />
            </div>

            <div style={styles.inputArea}>
                <input
                    style={styles.input}
                    value={inputValue}
                    onChange={(e) => setInputValue(e.target.value)}
                    onKeyDown={(e) => e.key === 'Enter' && handleSend()}
                    placeholder="Type a message..."
                />
                <button style={styles.button} onClick={handleSend}>Send</button>
            </div>
        </div>
    );
};

const styles: { [key: string]: React.CSSProperties } = {
    container: { border: '1px solid #ccc', borderRadius: '8px', width: '350px', display: 'flex', flexDirection: 'column', height: '450px', background: '#fff' },
    header: { padding: '10px', borderBottom: '1px solid #eee', display: 'flex', justifyContent: 'space-between', fontWeight: 'bold' },
    messageList: { flex: 1, overflowY: 'auto', padding: '10px', display: 'flex', flexDirection: 'column', gap: '12px' },
    messageItem: {
        padding: '8px 12px',
        borderRadius: '12px',
        maxWidth: '80%',
        wordBreak: 'break-word',
        boxShadow: '0 1px 2px rgba(0,0,0,0.05)'
    },
    sender: { fontSize: '0.7rem', fontWeight: 'bold' },
    inputArea: { padding: '10px', borderTop: '1px solid #eee', display: 'flex', gap: '5px' },
    input: { flex: 1, padding: '8px', borderRadius: '4px', border: '1px solid #ddd' },
    button: { padding: '8px 15px', background: '#007bff', color: '#fff', border: 'none', borderRadius: '4px', cursor: 'pointer' }
};