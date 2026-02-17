import { useEffect, useState } from "react";

const AdminDashboard = ({ token }: { token: string }) => {
    const [users, setUsers] = useState<any[]>([]);

    const fetchAdminData = async () => {
        try {
            const userRes = await fetch("http://localhost:880/admin/users", {
                headers: { Authorization: `Bearer ${token}` },
            });
            const userData = await userRes.json();

            // Safety check: ensure we got an array
            setUsers(Array.isArray(userData) ? userData : []);
        } catch (err) {
            console.error("Admin fetch error:", err);
            setUsers([]);
        }
    };

    useEffect(() => {
        fetchAdminData();
    }, []);

    const handleScrub = async (userId: number, email: string) => {
        if (!window.confirm(`PERMANENT: Scrub all data for ${email}?`)) return;

        try {
            const res = await fetch(`http://localhost:880/admin/users/${userId}`, {
                method: "DELETE",
                headers: { Authorization: `Bearer ${token}` },
            });

            if (res.ok) {
                fetchAdminData();
            } else {
                alert("Failed to scrub user.");
            }
        } catch (err) {
            console.error("Scrub error:", err);
        }
    };

    return (
        <div className="admin-container" style={{ padding: '40px', maxWidth: '900px', margin: '0 auto' }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                <h1 style={{ color: '#ff4d4d' }}>🛡️ ADMIN: USER MANAGEMENT</h1>
                <button
                    className="btn-logout"
                    onClick={() => { localStorage.clear(); window.location.reload(); }}
                    style={{ padding: '8px 16px', cursor: 'pointer' }}
                >
                    Logout
                </button>
            </div>

            <hr style={{ margin: '20px 0' }} />

            <section>
                <p>Manage all registered users and moderate the platform.</p>
                <table className="admin-table" style={{ width: '100%', borderCollapse: 'collapse', marginTop: '20px', backgroundColor: 'white', color: 'black' }}>
                    <thead>
                        <tr style={{ textAlign: 'left', borderBottom: '2px solid #ddd' }}>
                            <th style={{ padding: '12px' }}>ID</th>
                            <th style={{ padding: '12px' }}>Email</th>
                            <th style={{ padding: '12px' }}>Status</th>
                            <th style={{ padding: '12px' }}>Actions</th>
                        </tr>
                    </thead>
                    <tbody>
                        {users.length > 0 ? (
                            users.map(u => {
                                // Extract values checking both lowercase and uppercase keys
                                const uId = u.id ?? u.ID;
                                const uEmail = u.email ?? u.Email;

                                return (
                                    <tr key={uId} style={{ borderBottom: '1px solid #eee' }}>
                                        <td style={{ padding: '12px' }}>{uId}</td>
                                        <td style={{ padding: '12px' }}>{uEmail}</td>
                                        <td style={{ padding: '12px' }}>
                                            <span style={{
                                                padding: '4px 8px',
                                                borderRadius: '4px',
                                                fontSize: '12px',
                                                backgroundColor: '#e0f0ff',
                                                color: '#007bff'
                                            }}>
                                                Active
                                            </span>
                                        </td>
                                        <td style={{ padding: '12px' }}>
                                            <button
                                                onClick={() => handleScrub(uId, uEmail)}
                                                className="btn-danger"
                                                style={{
                                                    backgroundColor: '#ff4d4d',
                                                    color: 'white',
                                                    border: 'none',
                                                    padding: '6px 12px',
                                                    borderRadius: '4px',
                                                    cursor: 'pointer'
                                                }}
                                            >
                                                Scrub Identity
                                            </button>
                                        </td>
                                    </tr>
                                );
                            })
                        ) : (
                            <tr>
                                <td colSpan={4} style={{ padding: '20px', textAlign: 'center' }}>
                                    No users found.
                                </td>
                            </tr>
                        )}
                    </tbody>
                </table>
            </section>
        </div>
    );
};

export default AdminDashboard;