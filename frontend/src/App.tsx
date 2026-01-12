import { useState } from "react";
import Layout from "./components/Layout";

type Project = {
  id: number;
  name: string;
  description: string;
};

function App() {
  // --- auth state ---
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [message, setMessage] = useState("");

  // --- projects state ---
  const [projects, setProjects] = useState<Project[]>([]);
  const [loadingProjects, setLoadingProjects] = useState(false);

  const token = localStorage.getItem("token");

  async function handleLogin(e: React.FormEvent) {
    e.preventDefault();
    setMessage("");

    try {
      const res = await fetch("http://localhost:880/login", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ email, password }),
      });

      if (!res.ok) {
        setMessage("Login failed");
        return;
      }

      const data = await res.json();
      localStorage.setItem("token", data.token);
      setMessage("Logged in successfully");
      window.location.reload();
    } catch {
      setMessage("Network error");
    }
  }

  async function loadProjects() {
    if (!token) return;

    setLoadingProjects(true);

    try {
      const res = await fetch("http://localhost:880/projects", {
        headers: {
          Authorization: `Bearer ${token}`,
        },
      });

      if (!res.ok) return;

      const data = await res.json();
      setProjects(data);
    } finally {
      setLoadingProjects(false);
    }
  }

  // =========================
  // SINGLE RETURN — THIS IS KEY
  // =========================
  return (
    <Layout>
      {!token ? (
        <>
          <h1>Playingfield</h1>

          <form onSubmit={handleLogin}>
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

            <button type="submit">Login</button>
          </form>

          <p>{message}</p>
        </>
      ) : (
        <>
          <h1>My Projects</h1>

          <button onClick={loadProjects}>Load projects</button>
          <button
            onClick={() => {
              localStorage.removeItem("token");
              window.location.reload();
            }}
          >
            Logout
          </button>

          {loadingProjects && <p>Loading…</p>}

          <ul>
            {projects.map((p) => (
              <li key={p.id}>
                <strong>{p.name}</strong> — {p.description}
              </li>
            ))}
          </ul>
        </>
      )}
    </Layout>
  );
}

export default App;





