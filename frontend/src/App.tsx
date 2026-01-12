import { useState } from "react";
import Layout from "./components/Layout";
import ProjectUsers, { type UserInProject } from "./components/ProjectUsers";
import "./App.css";

type Project = {
  id: number;
  name: string;
  description: string;
};

function App() {
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [message, setMessage] = useState("");

  const [projects, setProjects] = useState<Project[]>([]); // Initialized as empty array
  const [loadingProjects, setLoadingProjects] = useState(false);
  const [showProjects, setShowProjects] = useState(false);

  const [projectUsersMap, setProjectUsersMap] = useState<Record<number, UserInProject[]>>({});
  const [loadingUsersMap, setLoadingUsersMap] = useState<Record<number, boolean>>({});
  const [showTasksMap, setShowTasksMap] = useState<Record<number, boolean>>({});

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

    if (showProjects) {
      setShowProjects(false);
      return;
    }

    setLoadingProjects(true);
    try {
      const res = await fetch("http://localhost:880/projects", {
        headers: { Authorization: `Bearer ${token}` },
      });
      if (!res.ok) return;

      const data = await res.json();
      // FIX: If data is null/undefined, set projects to an empty array
      setProjects(data || []);
      setShowProjects(true);
    } catch (err) {
      console.error("Failed to load projects:", err);
      setProjects([]); // Fallback on error
    } finally {
      setLoadingProjects(false);
    }
  }

  async function createProject() {
    if (!token) return;
    const name = prompt("Project name:");
    if (!name) return;
    const description = prompt("Project description:") || "";
    try {
      const res = await fetch("http://localhost:880/projects", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          Authorization: `Bearer ${token}`,
        },
        body: JSON.stringify({ name, description }),
      });
      if (!res.ok) return;
      const newProject = await res.json();
      // FIX: Spread safely using current projects or empty array
      setProjects((prev) => [...(prev || []), newProject]);
      setShowProjects(true);
    } catch (err) {
      console.error(err);
    }
  }

  async function toggleProjectUsers(projectID: number) {
    if (!token) return;
    if (projectUsersMap[projectID]) {
      setProjectUsersMap((prev) => {
        const next = { ...prev };
        delete next[projectID];
        return next;
      });
      return;
    }

    setLoadingUsersMap((prev) => ({ ...prev, [projectID]: true }));
    try {
      const res = await fetch(`http://localhost:880/projects/users?project_id=${projectID}`, {
        headers: { Authorization: `Bearer ${token}` },
      });
      if (!res.ok) {
        setProjectUsersMap((prev) => ({ ...prev, [projectID]: [] }));
        return;
      }
      const data = await res.json();
      setProjectUsersMap((prev) => ({ ...prev, [projectID]: data || [] }));
    } finally {
      setLoadingUsersMap((prev) => ({ ...prev, [projectID]: false }));
    }
  }

  const toggleTasks = (projectID: number) => {
    setShowTasksMap((prev) => ({ ...prev, [projectID]: !prev[projectID] }));
  };

  async function removeUser(projectID: number, userID: number) {
    if (!token) return;
    try {
      const res = await fetch(
        `http://localhost:880/projects/users?project_id=${projectID}&user_id=${userID}`,
        { method: "DELETE", headers: { Authorization: `Bearer ${token}` } }
      );
      if (!res.ok) return;
      setProjectUsersMap((prev) => ({
        ...prev,
        [projectID]: prev[projectID]?.filter((u) => u.id !== userID) || [],
      }));
    } catch (err) {
      console.error(err);
    }
  }

  return (
    <Layout>
      <div className="app-container">
        {!token ? (
          <div className="card">
            <h1>Playingfield</h1>
            <form onSubmit={handleLogin} className="login-form">
              <input type="email" placeholder="Email" value={email} onChange={(e) => setEmail(e.target.value)} required />
              <input type="password" placeholder="Password" value={password} onChange={(e) => setPassword(e.target.value)} required />
              <button type="submit" className="btn-primary">Login</button>
            </form>
            <p className="error-message">{message}</p>
          </div>
        ) : (
          <div className="project-list-container">
            <h1>My Projects</h1>
            <div className="button-group">
              <button onClick={loadProjects} className="btn-secondary">
                {showProjects ? "Hide Projects" : "Load Projects"}
              </button>
              <button onClick={createProject} className="btn-success">Create Project</button>
              <button onClick={() => { localStorage.removeItem("token"); window.location.reload(); }} className="btn-secondary">Logout</button>
            </div>

            {loadingProjects && <p>Loading…</p>}

            <div className={`expandable-section main-list ${showProjects ? "open" : ""}`}>
              <div className="inner-content-wrapper">
                {/* SAFE CHECK: Ensure projects isn't null before mapping */}
                {(projects && projects.length > 0) ? (
                  <ul className="project-list">
                    {projects.map((p) => (
                      <li key={p.id} className="project-card">
                        <div className="project-header">
                          <strong>{p.name}</strong> — {p.description}
                        </div>
                        <div className="project-actions">
                          <button onClick={() => toggleProjectUsers(p.id)} className="toggle-btn">
                            {projectUsersMap[p.id] ? "Hide Members" : "Show Members"}
                          </button>
                          <div className={`expandable-section ${projectUsersMap[p.id] ? "open" : ""}`}>
                            <div className="inner-content">
                              {loadingUsersMap[p.id] && <p>Loading members…</p>}
                              {projectUsersMap[p.id]?.length > 0 ? (
                                <ProjectUsers users={projectUsersMap[p.id]} removeUser={(uID) => removeUser(p.id, uID)} />
                              ) : (
                                !loadingUsersMap[p.id] && <p>No members yet.</p>
                              )}
                            </div>
                          </div>
                          <button onClick={() => toggleTasks(p.id)} className="toggle-btn">
                            {showTasksMap[p.id] ? "Hide Tasks" : "Show Tasks"}
                          </button>
                          <div className={`expandable-section ${showTasksMap[p.id] ? "open" : ""}`}>
                            <div className="inner-content">
                              <p>No tasks yet.</p>
                            </div>
                          </div>
                        </div>
                      </li>
                    ))}
                  </ul>
                ) : (
                  <p style={{ marginTop: "1rem" }}>You don't have any projects yet.</p>
                )}
              </div>
            </div>
          </div>
        )}
      </div>
    </Layout>
  );
}

export default App;