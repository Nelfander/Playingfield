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
        typingUserId // <-- Now matched with the hook's new state
    } = useDirectChat(token, otherUserId);

    const [inputValue, setInputValue] = useState("");
    const messageListRef = useRef<HTMLDivElement>(null);
    const typingTimeoutRef = useRef<any>(null);
    const currentUserId = Number(localStorage.getItem("userId"));

    // 1. Fetch History
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

    // 2. Scroll to Bottom
    useEffect(() => {
        if (messageListRef.current) {
            messageListRef.current.scrollTop = messageListRef.current.scrollHeight;
        }
    }, [messages, isTyping]);

    // 3. Send Read Receipt
    useEffect(() => {
        if (messages.length > 0) {
            const lastMsg = messages[messages.length - 1];
            if (Number(lastMsg.sender_id) === otherUserId) {
                sendReadReceipt(lastMsg.id, otherUserId);
            }
        }
    }, [messages.length, otherUserId, sendReadReceipt]);

    const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
        setInputValue(e.target.value);

        // Match the group chat logic: Only send typing start once
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

        // Clean up typing on send
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

                {/* Updated Typing logic to match ChatBox pattern */}
                {isTyping && (
                    <div style={styles.typingIndicator}>
                        {otherUserEmail} is typing...
                    </div>
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
    typingIndicator: { fontSize: '0.75rem', color: '#888', fontStyle: 'italic', margin: '5px 0', alignSelf: 'flex-start' },
    inputArea: { padding: '10px', borderTop: '1px solid #eee', display: 'flex', gap: '5px' },
    input: { flex: 1, padding: '8px', borderRadius: '4px', border: '1px solid #ddd' },
    button: { padding: '8px 15px', background: '#007bff', color: '#fff', border: 'none', borderRadius: '4px', cursor: 'pointer' }
};