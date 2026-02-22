import React, { useState, useEffect, useRef } from 'react';
import { useChat } from '../hooks/useChat';

interface ChatBoxProps {
    projectId: number;
    token: string;
    onClose: () => void; // New prop to handle closing the chat
}

export const ChatBox: React.FC<ChatBoxProps> = ({ projectId, token, onClose }) => {
    const {
        messages,
        setMessages,
        sendMessage,
        sendTypingStatus,
        isConnected,
        isTyping,
        typingUserId,
        typingUserEmail
    } = useChat(token, projectId);

    const [inputValue, setInputValue] = useState("");
    const [projectName, setProjectName] = useState<string>("");

    const messageListRef = useRef<HTMLDivElement>(null);
    const typingTimeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null);
    const currentUserId = Number(localStorage.getItem("userId"));

    // 1. Fetch Project Details
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

    // 2. Fetch History
    useEffect(() => {
        const fetchHistory = async () => {
            try {
                const res = await fetch(`http://localhost:880/projects/${projectId}/messages`, {
                    headers: { 'Authorization': `Bearer ${token}` }
                });
                if (res.ok) {
                    const history = await res.json();
                    setMessages(history || []);
                }
            } catch (err) {
                console.error("Failed to load history:", err);
                setMessages([]);
            }
        };
        if (projectId && token) fetchHistory();
    }, [projectId, token, setMessages]);

    // 3. Scroll to bottom
    useEffect(() => {
        if (messageListRef.current) {
            messageListRef.current.scrollTop = messageListRef.current.scrollHeight;
        }
    }, [messages]);

    const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
        setInputValue(e.target.value);
        if (!typingTimeoutRef.current) sendTypingStatus(true);

        if (typingTimeoutRef.current) clearTimeout(typingTimeoutRef.current);
        typingTimeoutRef.current = setTimeout(() => {
            sendTypingStatus(false);
            typingTimeoutRef.current = null;
        }, 2000);
    };

    const handleSend = () => {
        if (!inputValue.trim()) return;
        sendMessage(inputValue, 'project_chat', projectId);
        setInputValue("");
        if (typingTimeoutRef.current) {
            clearTimeout(typingTimeoutRef.current);
            sendTypingStatus(false);
            typingTimeoutRef.current = null;
        }
    };

    const formatTime = (dateStr?: string) => {
        if (!dateStr) return "";
        const date = new Date(dateStr);
        return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
    };

    return (
        <div style={styles.container}>
            {/* Header with Status Indicator */}
            <div style={styles.header}>
                <div style={{ display: 'flex', flexDirection: 'column' }}>
                    <span style={styles.headerTitle}>{projectName || `Project ${projectId}`}</span>
                    <div style={{ display: 'flex', alignItems: 'center', gap: '4px' }}>
                        <div style={{
                            width: '8px', height: '8px', borderRadius: '50%',
                            background: isConnected ? '#4caf50' : '#f44336',
                            boxShadow: isConnected ? '0 0 8px #4caf50' : 'none'
                        }} />
                        <span style={styles.statusText}>{isConnected ? 'Active' : 'Disconnected'}</span>
                    </div>
                </div>
                {/* Functional Close Button */}
                <button
                    style={styles.closeBtn}
                    onClick={onClose}
                    title="Close Chat"
                >
                    ✕
                </button>
            </div>

            {/* Message Area */}
            <div ref={messageListRef} style={styles.messageList}>
                {(messages || []).map((m, i) => {
                    const isMe = Number(m.sender_id) === currentUserId;
                    const time = formatTime(m.created_at);

                    return (
                        <div
                            key={m.id || i}
                            style={{
                                ...styles.messageWrapper,
                                alignItems: isMe ? 'flex-end' : 'flex-start',
                            }}
                        >
                            {!isMe && <span style={styles.senderLabel}>{m.sender_email?.split('@')[0] || `User ${m.sender_id}`}</span>}
                            <div
                                style={{
                                    ...styles.bubble,
                                    backgroundColor: isMe ? '#007bff' : '#ffffff',
                                    color: isMe ? '#fff' : '#333',
                                    borderRadius: isMe ? '16px 16px 2px 16px' : '16px 16px 16px 2px',
                                    border: isMe ? 'none' : '1px solid #e0e0e0',
                                }}
                            >
                                <div style={styles.messageContent}>{m.content}</div>
                                <div style={{
                                    ...styles.timeLabel,
                                    color: isMe ? 'rgba(255,255,255,0.7)' : '#999',
                                    textAlign: isMe ? 'right' : 'left'
                                }}>
                                    {time}
                                </div>
                            </div>
                        </div>
                    );
                })}
            </div>

            {/* Typing Indicator Overlay */}
            <div style={styles.typingShelf}>
                {isTyping && typingUserId !== currentUserId && (
                    <div style={styles.typingIndicator}>
                        <span className="typing-dots">{typingUserEmail || `User ${typingUserId}`} is typing</span>
                    </div>
                )}
            </div>

            {/* Input Bar */}
            <div style={styles.inputArea}>
                <div style={styles.inputWrapper}>
                    <input
                        style={styles.input}
                        value={inputValue}
                        onChange={handleInputChange}
                        onKeyDown={(e) => e.key === 'Enter' && handleSend()}
                        placeholder="Write a message..."
                    />
                    <button style={styles.sendButton} onClick={handleSend}>
                        <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                            <line x1="22" y1="2" x2="11" y2="13"></line>
                            <polygon points="22 2 15 22 11 13 2 9 22 2"></polygon>
                        </svg>
                    </button>
                </div>
            </div>
        </div>
    );
};

const styles: { [key: string]: React.CSSProperties } = {
    container: {
        width: '400px',
        height: '600px',
        display: 'flex',
        flexDirection: 'column',
        background: '#f7f9fc',
        borderRadius: '20px',
        boxShadow: '0 12px 28px rgba(0,0,0,0.12)',
        overflow: 'hidden',
        fontFamily: '-apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Helvetica, Arial, sans-serif'
    },
    header: {
        padding: '16px 20px',
        background: '#ffffff',
        borderBottom: '1px solid #eef2f7',
        display: 'flex',
        justifyContent: 'space-between',
        alignItems: 'center'
    },
    headerTitle: { fontWeight: 700, fontSize: '1.1rem', color: '#1a1d21' },
    statusText: { fontSize: '0.75rem', color: '#71767c', fontWeight: 500 },
    closeBtn: {
        background: 'none',
        border: 'none',
        color: '#b0b7c3',
        cursor: 'pointer',
        fontSize: '1.2rem',
        padding: '4px',
        lineHeight: 1,
        transition: 'color 0.2s ease'
    },
    messageList: {
        flex: 1,
        overflowY: 'auto',
        padding: '20px',
        display: 'flex',
        flexDirection: 'column',
        gap: '16px'
    },
    messageWrapper: { display: 'flex', flexDirection: 'column', maxWidth: '85%' },
    senderLabel: { fontSize: '0.7rem', fontWeight: 600, color: '#8a94a6', marginBottom: '4px', marginLeft: '4px' },
    bubble: {
        padding: '10px 14px',
        boxShadow: '0 2px 4px rgba(0,0,0,0.02)',
        position: 'relative'
    },
    messageContent: { fontSize: '0.95rem', lineHeight: '1.4' },
    timeLabel: { fontSize: '0.65rem', marginTop: '4px', fontWeight: 500 },
    typingShelf: { height: '28px', padding: '0 20px', display: 'flex', alignItems: 'center' },
    typingIndicator: { fontSize: '0.8rem', color: '#007bff', fontWeight: 500 },
    inputArea: { padding: '15px 20px 25px 20px', background: '#ffffff' },
    inputWrapper: {
        display: 'flex',
        alignItems: 'center',
        background: '#f0f2f5',
        borderRadius: '25px',
        padding: '5px 5px 5px 15px'
    },
    input: {
        flex: 1,
        background: 'transparent',
        border: 'none',
        padding: '10px 0',
        outline: 'none',
        fontSize: '0.95rem',
        color: '#333'
    },
    sendButton: {
        background: '#007bff',
        color: '#fff',
        border: 'none',
        borderRadius: '50%',
        width: '38px',
        height: '38px',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        cursor: 'pointer',
        transition: 'transform 0.2s ease',
        marginLeft: '10px'
    }
};