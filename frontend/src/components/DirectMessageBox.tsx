import React, { useState, useEffect, useRef } from 'react';
import { useDirectChat } from '../hooks/useDirectChat';

interface DirectMessageBoxProps {
    otherUserId: number;
    otherUserEmail: string;
    token: string;
}

export const DirectMessageBox: React.FC<DirectMessageBoxProps> = ({
    otherUserId,
    otherUserEmail,
    token
}) => {
    const {
        messages,
        setMessages,
        sendMessage,
        sendTypingStatus,
        sendReadReceipt,
        isConnected,
        isTyping,
        typingUserId
    } = useDirectChat(token, otherUserId);

    const [inputValue, setInputValue] = useState("");
    const messageListRef = useRef<HTMLDivElement>(null);
    const typingTimeoutRef = useRef<any>(null);
    const currentUserId = Number(localStorage.getItem("userId"));

    useEffect(() => {
        const fetchHistory = async () => {
            try {
                const response = await fetch(`http://localhost:880/messages/direct/${otherUserId}`, {
                    headers: { 'Authorization': `Bearer ${token}` }
                });
                if (response.ok) {
                    const history = await response.json();
                    setMessages(history || []);
                }
            } catch (err) {
                console.error("Failed to load DM history:", err);
                setMessages([]);
            }
        };
        if (otherUserId && token) fetchHistory();
    }, [otherUserId, token, setMessages]);

    // Scroll only when NEW MESSAGES arrive, not when typing happens
    useEffect(() => {
        if (messageListRef.current) {
            messageListRef.current.scrollTop = messageListRef.current.scrollHeight;
        }
    }, [messages]);

    const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
        setInputValue(e.target.value);
        if (!typingTimeoutRef.current) {
            sendTypingStatus(true, otherUserId);
        }
        if (typingTimeoutRef.current) clearTimeout(typingTimeoutRef.current);
        typingTimeoutRef.current = setTimeout(() => {
            sendTypingStatus(false, otherUserId);
            typingTimeoutRef.current = null;
        }, 2000);
    };

    const handleSend = () => {
        if (!inputValue.trim()) return;
        sendMessage(inputValue, otherUserId);
        if (typingTimeoutRef.current) {
            clearTimeout(typingTimeoutRef.current);
            typingTimeoutRef.current = null;
        }
        sendTypingStatus(false, otherUserId);
        setInputValue("");
    };

    const formatTime = (dateStr?: string) => {
        if (!dateStr) return "";
        const date = new Date(dateStr);
        return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
    };

    const lastSentMessageId = [...messages]
        .reverse()
        .find(m => Number(m.sender_id) === currentUserId)?.id;

    return (
        <div style={styles.container}>
            <div style={styles.header}>
                <span>Chat with {otherUserEmail}</span>
                <span style={{ color: isConnected ? '#4caf50' : '#f44336', fontSize: '0.8rem' }}>
                    {isConnected ? ' ● Online' : ' ● Offline'}
                </span>
            </div>

            {/* Scrollable Area */}
            <div ref={messageListRef} style={styles.messageList}>
                {(messages || []).map((m, i) => {
                    const isMe = Number(m.sender_id) === currentUserId;
                    const time = formatTime(m.created_at);
                    const showSeen = isMe && m.id === lastSentMessageId && (m as any).read_at;

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
                                <small style={{ ...styles.sender, color: isMe ? '#e0e0e0' : '#888' }}>
                                    {isMe ? "Me" : otherUserEmail}
                                </small>
                                {time && <small style={{ fontSize: '0.6rem', color: isMe ? '#ccc' : '#999' }}>{time}</small>}
                            </div>
                            <div style={{ marginTop: '2px' }}>{m.content}</div>
                            {showSeen && (
                                <div style={{ textAlign: 'right', fontSize: '0.6rem', marginTop: '2px', color: '#e0e0e0', fontWeight: 'bold' }}>
                                    ✓ Seen
                                </div>
                            )}
                        </div>
                    );
                })}
            </div>

            {/* STICKY TYPING SHELF (Outside the scroll area) */}
            <div style={styles.typingShelf}>
                {isTyping && Number(typingUserId) === Number(otherUserId) && (
                    <span style={styles.typingIndicator}>
                        {otherUserEmail} is typing...
                    </span>
                )}
            </div>

            <div style={styles.inputArea}>
                <input
                    style={styles.input}
                    value={inputValue}
                    onChange={handleInputChange}
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
    messageItem: { padding: '8px 12px', borderRadius: '12px', maxWidth: '80%', wordBreak: 'break-word', boxShadow: '0 1px 2px rgba(0,0,0,0.05)' },
    sender: { fontSize: '0.7rem', fontWeight: 'bold' },
    typingShelf: { height: '22px', paddingLeft: '12px', display: 'flex', alignItems: 'center', background: '#fff' },
    typingIndicator: { fontSize: '0.75rem', color: '#007bff', fontStyle: 'italic', fontWeight: '500' },
    inputArea: { padding: '10px', borderTop: '1px solid #eee', display: 'flex', gap: '5px' },
    input: { flex: 1, padding: '8px', borderRadius: '4px', border: '1px solid #ddd' },
    button: { padding: '8px 15px', background: '#007bff', color: '#fff', border: 'none', borderRadius: '4px', cursor: 'pointer' }
};