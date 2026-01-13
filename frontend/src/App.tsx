import { useState } from "react";
import ProjectList from "./components/ProjectList";
import LoginForm from "./components/LoginForm";
import CreateProjectModal from "./components/CreateProjectModal";
import { type UserInProject } from "./components/ProjectUsers";
import "./App.css";

type Project = {
  id: number;
  name: string;
  description: string;
  owner_id: number;
};

function App() {
  const [message, setMessage] = useState("");
  const [projects, setProjects] = useState<Project[]>([]);
  const [showProjects, setShowProjects] = useState(false);
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [projectUsersMap, setProjectUsersMap] = useState<Record<number, UserInProject[]>>({});
  const [showTasksMap, setShowTasksMap] = useState<Record<number, boolean>>({});

  const token = localStorage.getItem("token");

  // LOGIC FIX: Ensure we get a valid number or 0
  const currentUserId = Number(localStorage.getItem("userId")) || 0;

  async function handleProjectToggle() {
    if (!token) return;

    if (showProjects) {
      setShowProjects(false);
      setProjects([]);
      return;
    }

    try {
      const res = await fetch("http://localhost:880/projects", {
        headers: { Authorization: `Bearer ${token}` },
      });
      const data = await res.json();
      setProjects(data || []);
      setShowProjects(true);
    } catch (err) {
      console.error("Fetch projects error:", err);
      setMessage("Could not load projects.");
    }
  }

  async function handleAddMember(projectId: number, userId: number) {
    if (!token) return;

    const confirmAdd = window.confirm("Are you sure you want to add this member?");
    if (!confirmAdd) return;

    try {
      const res = await fetch("http://localhost:880/projects/users", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          Authorization: `Bearer ${token}`
        },
        body: JSON.stringify({ project_id: projectId, user_id: userId, role: "member" }),
      });

      if (res.ok) {
        // Refresh the user list for this project
        setProjectUsersMap(prev => {
          const n = { ...prev };
          delete n[projectId];
          return n;
        });
        toggleProjectUsers(projectId);
      } else {
        const err = await res.json();
        alert(err.error || "Failed to add member");
      }
    } catch (err) { console.error(err); }
  }

  async function toggleProjectUsers(projectID: number) {
    if (projectUsersMap[projectID]) {
      setProjectUsersMap(prev => {
        const n = { ...prev };
        delete n[projectID];
        return n;
      });
      return;
    }
    try {
      const res = await fetch(`http://localhost:880/projects/users?project_id=${projectID}`, {
        headers: { Authorization: `Bearer ${token}` },
      });
      const data = await res.json();
      setProjectUsersMap(prev => ({ ...prev, [projectID]: data || [] }));
    } catch (err) { console.error(err); }
  }

  async function removeUser(projectID: number, userID: number) {
    if (!window.confirm("Are you sure you want to remove this member?")) return;
    try {
      await fetch(`http://localhost:880/projects/users`, {
        method: "DELETE",
        headers: {
          "Authorization": `Bearer ${token}`,
          "Content-Type": "application/json"
        },
        body: JSON.stringify({ project_id: projectID, user_id: userID })
      });
      setProjectUsersMap(prev => ({
        ...prev,
        [projectID]: prev[projectID]?.filter(u => u.id !== userID) || []
      }));
    } catch (err) { console.error(err); }
  }

  return (
    <div className="app-container">
      {!token ? (
        <LoginForm message={message} setMessage={setMessage} />
      ) : (
        <div className="project-list-container">
          <h1>My Projects</h1>
          <div className="button-group">
            <button onClick={handleProjectToggle}>
              {showProjects ? "Hide Projects" : "Load Projects"}
            </button>
            <button onClick={() => setIsModalOpen(true)} className="btn-success">
              Create Project
            </button>
            <button onClick={() => {
              localStorage.removeItem("token");
              localStorage.removeItem("userId");
              window.location.reload();
            }}>
              Logout
            </button>
          </div>

          <ProjectList
            projects={projects}
            currentUserId={currentUserId}
            showProjects={showProjects}
            projectUsersMap={projectUsersMap}
            showTasksMap={showTasksMap}
            toggleProjectUsers={toggleProjectUsers}
            toggleTasks={(id) => setShowTasksMap(p => ({ ...p, [id]: !p[id] }))}
            removeUser={removeUser}
            handleAddMember={handleAddMember}
          />

          <CreateProjectModal
            isOpen={isModalOpen}
            onClose={() => setIsModalOpen(false)}
            onCreate={async (name, desc) => {
              try {
                const res = await fetch("http://localhost:880/projects", {
                  method: "POST",
                  headers: {
                    "Content-Type": "application/json",
                    Authorization: `Bearer ${token}`
                  },
                  body: JSON.stringify({ name, description: desc }),
                });
                const newP = await res.json();
                setProjects(p => [...p, newP]);
                setIsModalOpen(false);
                setShowProjects(true);
              } catch (err) { console.error(err); }
            }}
          />
        </div>
      )}
    </div>
  );
}

export default App;