import React, { useState } from "react";

const LoginForm = ({ message, setMessage }: any) => {
    const [email, setEmail] = useState("");
    const [password, setPassword] = useState("");

    const handleLogin = async (e: React.FormEvent) => {
        e.preventDefault();
        try {
            const res = await fetch("http://localhost:880/login", {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({ email, password }),
            });

            const data = await res.json();

            if (res.ok) {
                // Simple, direct storage
                localStorage.setItem("token", data.token);
                localStorage.setItem("userId", data.userId);
                window.location.reload();
            } else {
                setMessage(data.error || "Login failed");
            }
        } catch (err) {
            setMessage("Login Error");
        }
    };

    return (
        <div className="login-wrapper">
            <div className="login-card">
                <h1>Playingfield</h1>
                <form onSubmit={handleLogin} className="login-form">
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
                    <button type="submit" className="btn-login">Login</button>
                </form>
                {message && <p style={{ color: "red", marginTop: "15px" }}>{message}</p>}
            </div>
        </div>
    );
};

export default LoginForm;