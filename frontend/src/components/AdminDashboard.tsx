import { useEffect, useState, useCallback } from "react";
import { apiUrl } from '../config/env';

// Props updated to include refreshTick for real-time UI sync
interface AdminDashboardProps {
    token: string;
    currentUserId: number;
    refreshTick?: number;
}

const AdminDashboard = ({ token, currentUserId, refreshTick }: AdminDashboardProps) => {
    const [users, setUsers] = useState<any[]>([]);

    // Using useCallback so we can safely include it in the useEffect dependency array
    const fetchAdminData = useCallback(async () => {
        try {
            const userRes = await fetch(apiUrl("admin/users"), {
                headers: { Authorization: `Bearer ${token}` },
            });
            const userData = await userRes.json();
            setUsers(Array.isArray(userData) ? userData : []);
        } catch (err) {
            console.error("Admin fetch error:", err);
            setUsers([]);
        }
    }, [token]);

    // This effect runs on mount, when token changes, AND when refreshTick changes
    useEffect(() => {
        fetchAdminData();
    }, [fetchAdminData, refreshTick]);

    const handleToggleStatus = async (userId: number) => {
        try {
            const res = await fetch(apiUrl(`admin/users/${userId}/toggle`), {
                method: "POST",
                headers: { Authorization: `Bearer ${token}` },
            });
            if (res.ok) {
                // We don't necessarily need to call fetchAdminData() here 
                // because the WebSocket will trigger the refreshTick anyway!
                // But keeping it for instant local feedback is fine.
                fetchAdminData();
            } else {
                const errorData = await res.json();
                alert(`Action Failed: ${errorData.message || 'Check backend logs'}`);
            }
        } catch (err) {
            console.error("Toggle error:", err);
        }
    };

    const handleScrub = async (userId: number, email: string) => {
        if (!window.confirm(`⚠️ PERMANENT DESTRUCTION: Scrub all data for ${email}?`)) return;
        try {
            const res = await fetch(apiUrl(`admin/users/${userId}`), {
                method: "DELETE",
                headers: { Authorization: `Bearer ${token}` },
            });
            if (res.ok) fetchAdminData();
        } catch (err) {
            console.error("Scrub error:", err);
        }
    };

    return (
        <div style={styles.adminOverlay}>
            <div style={styles.glassCard}>
                <header style={styles.header}>
                    <div>
                        <h1 style={styles.title}>🛡️ CONTROL CENTER</h1>
                        <p style={styles.subtitle}>System-wide User Management & Moderation</p>
                    </div>
                    <div style={styles.badge}>ADMIN MODE</div>
                </header>

                <div style={styles.tableWrapper}>
                    <table style={styles.table}>
                        <thead>
                            <tr style={styles.theadRow}>
                                <th style={styles.th}>ID</th>
                                <th style={styles.th}>IDENTIFIER (EMAIL)</th>
                                <th style={styles.th}>STATUS</th>
                                <th style={{ ...styles.th, textAlign: 'right' }}>ACTIONS</th>
                            </tr>
                        </thead>
                        <tbody>
                            {users.length > 0 ? (
                                users
                                    .filter(u => (u.id ?? u.ID) !== currentUserId)
                                    .map(u => {
                                        const uId = u.id ?? u.ID;
                                        const uEmail = u.email ?? u.Email;
                                        const uStatus = (u.status ?? u.Status ?? 'active').toLowerCase();

                                        return (
                                            <tr key={uId} style={styles.tr}>
                                                <td style={styles.td}>#{uId}</td>
                                                <td style={styles.td}><strong>{uEmail}</strong></td>
                                                <td style={styles.td}>
                                                    <span style={uStatus === 'active' ? styles.statusActive : styles.statusInactive}>
                                                        ● {uStatus.toUpperCase()}
                                                    </span>
                                                </td>
                                                <td style={{ ...styles.td, textAlign: 'right' }}>
                                                    <button
                                                        onClick={() => handleToggleStatus(uId)}
                                                        style={uStatus === 'active' ? styles.deactivateBtn : styles.activateBtn}
                                                    >
                                                        {uStatus === 'active' ? "DEACTIVATE" : "ACTIVATE"}
                                                    </button>

                                                    <button
                                                        onClick={() => handleScrub(uId, uEmail)}
                                                        style={styles.scrubBtn}
                                                    >
                                                        SCRUB DATA
                                                    </button>
                                                </td>
                                            </tr>
                                        );
                                    })
                            ) : (
                                <tr>
                                    <td colSpan={4} style={styles.noData}>No system users found.</td>
                                </tr>
                            )}
                        </tbody>
                    </table>
                </div>
            </div>
        </div>
    );
};

const styles: Record<string, React.CSSProperties> = {
    adminOverlay: {
        padding: '40px',
        display: 'flex',
        justifyContent: 'center',
    },
    glassCard: {
        width: '100%',
        maxWidth: '1000px',
        background: 'rgba(255, 255, 255, 0.08)',
        backdropFilter: 'blur(20px)',
        WebkitBackdropFilter: 'blur(20px)',
        borderRadius: '24px',
        border: '1px solid rgba(255, 255, 255, 0.15)',
        boxShadow: '0 25px 50px -12px rgba(0, 0, 0, 0.5)',
        padding: '40px',
        color: '#fff',
    },
    header: {
        display: 'flex',
        justifyContent: 'space-between',
        alignItems: 'flex-start',
        marginBottom: '40px',
        borderBottom: '1px solid rgba(255,255,255,0.1)',
        paddingBottom: '20px',
    },
    title: {
        margin: 0,
        fontSize: '2rem',
        letterSpacing: '2px',
        background: 'linear-gradient(to right, #fff, #ff4d4d)',
        WebkitBackgroundClip: 'text',
        WebkitTextFillColor: 'transparent',
    },
    subtitle: {
        margin: '5px 0 0 0',
        opacity: 0.6,
        fontSize: '0.9rem',
    },
    badge: {
        background: '#ff4d4d',
        padding: '5px 15px',
        borderRadius: '50px',
        fontSize: '12px',
        fontWeight: 'bold',
        letterSpacing: '1px',
    },
    tableWrapper: {
        overflowX: 'auto',
    },
    table: {
        width: '100%',
        borderCollapse: 'collapse',
    },
    theadRow: {
        borderBottom: '2px solid rgba(255,255,255,0.1)',
    },
    th: {
        padding: '15px',
        textAlign: 'left',
        fontSize: '12px',
        opacity: 0.5,
        letterSpacing: '1px',
        textTransform: 'uppercase',
    },
    tr: {
        borderBottom: '1px solid rgba(255,255,255,0.05)',
        transition: 'background 0.3s',
    },
    td: {
        padding: '20px 15px',
        fontSize: '14px',
    },
    statusActive: {
        color: '#4CAF50',
        fontSize: '12px',
        fontWeight: 'bold',
        letterSpacing: '0.5px',
    },
    statusInactive: {
        color: '#feb019',
        fontSize: '12px',
        fontWeight: 'bold',
        letterSpacing: '0.5px',
    },
    activateBtn: {
        background: 'rgba(76, 175, 80, 0.1)',
        border: '1px solid #4CAF50',
        color: '#4CAF50',
        padding: '8px 16px',
        borderRadius: '8px',
        cursor: 'pointer',
        fontSize: '12px',
        fontWeight: 'bold',
        marginRight: '10px',
        transition: 'all 0.2s',
    },
    deactivateBtn: {
        background: 'rgba(254, 176, 25, 0.1)',
        border: '1px solid #feb019',
        color: '#feb019',
        padding: '8px 16px',
        borderRadius: '8px',
        cursor: 'pointer',
        fontSize: '12px',
        fontWeight: 'bold',
        marginRight: '10px',
        transition: 'all 0.2s',
    },
    scrubBtn: {
        background: 'transparent',
        border: '1px solid #ff4d4d',
        color: '#ff4d4d',
        padding: '8px 16px',
        borderRadius: '8px',
        cursor: 'pointer',
        fontSize: '12px',
        fontWeight: 'bold',
        transition: 'all 0.2s',
    },
    noData: {
        padding: '40px',
        textAlign: 'center',
        opacity: 0.5,
    },
};

export default AdminDashboard;