import React, { useState } from "react";

const LoginForm = ({ message, setMessage }: any) => {
    const [email, setEmail] = useState("");
    const [password, setPassword] = useState("");
    const [confirmPassword, setConfirmPassword] = useState("");
    const [isRegistering, setIsRegistering] = useState(false);
    const [status, setStatus] = useState<"error" | "success" | "">("");

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        setMessage("");
        setStatus("");

        // 1. Client-side check: Passwords must match for registration
        if (isRegistering && password !== confirmPassword) {
            setStatus("error");
            setMessage("Passwords do not match");
            return;
        }

        const endpoint = isRegistering ? "/users" : "/login";

        try {
            const res = await fetch(`http://localhost:880${endpoint}`, {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({ email, password }),
            });

            const data = await res.json();

            if (res.ok) {
                setStatus("success");
                if (isRegistering) {
                    // Registration Success
                    setMessage("Account created! You can now login.");
                    setIsRegistering(false); // Switch UI back to login mode
                    setPassword("");
                    setConfirmPassword("");
                } else {
                    // --- Login Success ---
                    localStorage.setItem("token", data.token);
                    localStorage.setItem("userId", data.userId);

                    // Store the role so the App knows whether to show User or Admin UI
                    // We check if data.user exists, otherwise fallback to "user"
                    const userRole = data.user?.role || "user";
                    localStorage.setItem("role", userRole);

                    // Reload triggers the App.tsx logic to pick up the new role
                    window.location.reload();
                }
            } else {
                // Backend Error: Uses "error" key from your Go Translator
                setStatus("error");
                setMessage(data.error || "An unexpected error occurred");
            }
        } catch (err) {
            setStatus("error");
            setMessage("Cannot connect to server");
        }
    };

    return (
        <div className="login-wrapper">
            <div className="login-card">
                <h1>Playingfield</h1>

                <form onSubmit={handleSubmit} className="login-form">
                    <input
                        type="email"
                        placeholder="Email"
                        value={email}
                        onChange={(e) => setEmail(e.target.value)}
                        required
                    />
                    <input
                        type="password"
                        placeholder="Password"
                        value={password}
                        onChange={(e) => setPassword(e.target.value)}
                        required
                    />

                    {/* Only visible when 'Register' is clicked */}
                    {isRegistering && (
                        <input
                            type="password"
                            placeholder="Confirm Password"
                            value={confirmPassword}
                            onChange={(e) => setConfirmPassword(e.target.value)}
                            required
                        />
                    )}

                    <div style={{ display: 'flex', flexDirection: 'column', gap: '10px', marginTop: '10px' }}>
                        <button type="submit" className="btn-login">
                            {isRegistering ? "Create Account" : "Login"}
                        </button>

                        <button
                            type="button"
                            onClick={() => {
                                setIsRegistering(!isRegistering);
                                setMessage("");
                                setStatus("");
                                setConfirmPassword("");
                            }}
                            style={{
                                background: 'none',
                                border: 'none',
                                color: '#007bff',
                                cursor: 'pointer',
                                textDecoration: 'underline',
                                fontSize: '0.9rem'
                            }}
                        >
                            {isRegistering ? "Back to Login" : "Need an account? Register here"}
                        </button>
                    </div>
                </form>

                {message && (
                    <p style={{
                        color: status === "success" ? "#28a745" : "#dc3545",
                        marginTop: "15px",
                        fontWeight: "bold",
                        fontSize: "0.9rem"
                    }}>
                        {message}
                    </p>
                )}
            </div>
        </div>
    );
};

export default LoginForm;