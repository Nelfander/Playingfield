import { useState } from "react";
import ProjectList from "./components/ProjectList";
import LoginForm from "./components/LoginForm";
import CreateProjectModal from "./components/CreateProjectModal";
import { ChatBox } from "./components/ChatBox";
import { DirectMessageBox } from "./components/DirectMessageBox";
import { type UserInProject } from "./components/ProjectUsers";
import "./App.css";

type Project = {
  id: number;
  name: string;
  description: string;
  owner_id: number;
  owner_name?: string;
};

function App() {
  const [message, setMessage] = useState("");
  const [projects, setProjects] = useState<Project[]>([]);
  const [showProjects, setShowProjects] = useState(false);
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [projectUsersMap, setProjectUsersMap] = useState<Record<number, UserInProject[]>>({});
  const [showTasksMap, setShowTasksMap] = useState<Record<number, boolean>>({});

  // Track which project chat is open
  const [selectedProjectId, setSelectedProjectId] = useState<number | null>(null);

  // Track which direct message conversation is open
  const [selectedDMUserId, setSelectedDMUserId] = useState<number | null>(null);
  const [selectedDMUserEmail, setSelectedDMUserEmail] = useState<string>("");

  const token = localStorage.getItem("token");
  const currentUserId = Number(localStorage.getItem("userId")) || 0;

  async function fetchProjects() {
    if (!token) return;
    try {
      // Note: Ensure your port is correct (8080 vs 880)
      const res = await fetch("http://localhost:880/projects", {
        headers: { Authorization: `Bearer ${token}` },
      });
      const data = await res.json();
      setProjects(data || []);
    } catch (err) {
      console.error("Fetch projects error:", err);
    }
  }

  async function handleProjectToggle() {
    if (showProjects) {
      setShowProjects(false);
      setProjects([]);
      setSelectedProjectId(null); // Close chat if hiding projects
      setSelectedDMUserId(null); // Close DM if hiding projects
    } else {
      await fetchProjects();
      setShowProjects(true);
    }
  }

  async function handleLiveProjectCreated() {
    console.log("WS Signal: New project created. Updating state...");
    await fetchProjects();
  }

  function handleDeleteProjectState(projectId: number) {
    setProjects(prev => prev.filter(p => p.id !== projectId));
    if (selectedProjectId === projectId) setSelectedProjectId(null);
    setProjectUsersMap(prev => {
      const updated = { ...prev };
      delete updated[projectId];
      return updated;
    });
  }

  function handleLiveUserAdded(projectId: number, userId: number, role: string) {
    const isMe = userId === currentUserId;
    if (isMe) {
      fetchProjects();
      setShowProjects(true);
    } else {
      if (!token) return;
      fetch(`http://localhost:880/projects/users?project_id=${projectId}`, {
        headers: { Authorization: `Bearer ${token}` },
      })
        .then(res => res.json())
        .then(data => {
          setProjectUsersMap(prev => ({
            ...prev,
            [projectId]: data || []
          }));
        })
        .catch(err => console.error("Live sync fetch error:", err));
    }
  }

  function handleLiveUserRemoved(projectId: number, userId: number) {
    const isMe = userId === currentUserId;
    if (isMe) {
      fetchProjects();
      if (selectedProjectId === projectId) setSelectedProjectId(null);
    } else {
      setProjectUsersMap(prev => {
        if (!prev[projectId]) return prev;
        return {
          ...prev,
          [projectId]: prev[projectId].filter(u => u.id !== userId)
        };
      });
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

  // Handler to start a direct message conversation
  function handleStartDM(userId: number, userEmail: string) {
    setSelectedDMUserId(userId);
    setSelectedDMUserEmail(userEmail);
    setSelectedProjectId(null); // Close project chat if open
  }

  return (
    <div className="app-container">
      {!token ? (
        <LoginForm message={message} setMessage={setMessage} />
      ) : (
        <div className="main-layout" style={{ display: 'flex', gap: '20px', padding: '20px' }}>

          {/* Left Side: Project List */}
          <div className="project-list-container" style={{ flex: 1 }}>
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
              onDeleteProject={handleDeleteProjectState}
              onUserAdded={handleLiveUserAdded}
              onProjectCreated={handleLiveProjectCreated}
              onUserRemoved={handleLiveUserRemoved}
              // Pass a way to select a project for chat
              onSelectProject={(id) => setSelectedProjectId(id)}
              // Pass a way to start a direct message
              onStartDM={handleStartDM}
            />
          </div>

          {/* Right Side: Chat (Project Chat or Direct Message) */}
          {selectedProjectId && (
            <div className="chat-sidebar">
              <button onClick={() => setSelectedProjectId(null)} style={{ marginBottom: '10px' }}>Close Chat</button>
              <ChatBox projectId={selectedProjectId} token={token} />
            </div>
          )}

          {selectedDMUserId && (
            <div className="chat-sidebar">
              <button onClick={() => setSelectedDMUserId(null)} style={{ marginBottom: '10px' }}>Close Chat</button>
              <DirectMessageBox
                otherUserId={selectedDMUserId}
                otherUserEmail={selectedDMUserEmail}
                token={token!}
              />
            </div>
          )}

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