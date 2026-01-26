import { useState } from "react";
import ProjectList from "./components/ProjectList";
import LoginForm from "./components/LoginForm";
import CreateProjectModal from "./components/CreateProjectModal";
import { ChatBox } from "./components/ChatBox";
import { DirectMessageBox } from "./components/DirectMessageBox";
import { type UserInProject } from "./components/ProjectUsers";
import { useWebSockets } from "./hooks/useWebSockets";
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

  // projectUsersMap holds the ACTUAL DATA (emails/ids)
  const [projectUsersMap, setProjectUsersMap] = useState<Record<number, UserInProject[]>>({});

  // UI state for accordions
  const [showUsersMap, setShowUsersMap] = useState<Record<number, boolean>>({});
  const [showTasksMap, setShowTasksMap] = useState<Record<number, boolean>>({});

  const [selectedProjectId, setSelectedProjectId] = useState<number | null>(null);
  const [selectedDMUserId, setSelectedDMUserId] = useState<number | null>(null);
  const [selectedDMUserEmail, setSelectedDMUserEmail] = useState<string>("");

  const token = localStorage.getItem("token");
  const currentUserId = Number(localStorage.getItem("userId")) || 0;

  // This simple number triggers refreshes across all task boards
  const [taskRefreshTick, setTaskRefreshTick] = useState(0);

  // --- WebSocket Handlers ---
  const handleTaskSignal = (projectId: number) => {
    console.log(`WS Signal: Task change in project ${projectId}`);
    // Incrementing the number forces a re-render/re-fetch in children
    setTaskRefreshTick(prev => prev + 1);
  };

  useWebSockets(
    token,
    (id) => handleDeleteProjectState(id),
    (pId, uId, role) => handleLiveUserAdded(pId, uId, role),
    () => handleLiveProjectCreated(),
    (pId, uId) => handleLiveUserRemoved(pId, uId),
    () => fetchProjects(),
    (pId) => handleTaskSignal(pId), // Created
    (pId) => handleTaskSignal(pId), // Updated
    (pId) => handleTaskSignal(pId)  // Deleted
  );

  async function fetchUsersData(projectId: number) {
    if (!token) return;
    try {
      const res = await fetch(`http://localhost:880/projects/users?project_id=${projectId}`, {
        headers: { Authorization: `Bearer ${token}` },
      });
      const data = await res.json();
      setProjectUsersMap(prev => ({ ...prev, [projectId]: data || [] }));
    } catch (err) {
      console.error(`Error fetching users for project ${projectId}:`, err);
    }
  }

  async function fetchProjects() {
    if (!token) return;
    try {
      const res = await fetch("http://localhost:880/projects", {
        headers: { Authorization: `Bearer ${token}` },
      });
      const data = await res.json();
      const projectData: Project[] = data || [];
      setProjects(projectData);

      // Fetch members for ALL projects immediately so TaskBoard always has them
      projectData.forEach(p => fetchUsersData(p.id));

    } catch (err) {
      console.error("Fetch projects error:", err);
    }
  }

  async function handleProjectToggle() {
    if (showProjects) {
      setShowProjects(false);
      setProjects([]);
      setSelectedProjectId(null);
      setSelectedDMUserId(null);
    } else {
      await fetchProjects();
      setShowProjects(true);
    }
  }

  async function handleLiveProjectCreated() {
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
      fetchUsersData(projectId);
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

  async function addMemberToMap(projectId: number, userId: number) {
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
        fetchUsersData(projectId);
      } else {
        const err = await res.json();
        alert(err.error || "Failed to add member");
      }
    } catch (err) { console.error(err); }
  }

  async function toggleProjectUsers(projectId: number) {
    if (!projectUsersMap[projectId]) {
      await fetchUsersData(projectId);
    }
    setShowUsersMap(prev => ({
      ...prev,
      [projectId]: !prev[projectId]
    }));
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

  function handleStartDM(userId: number, userEmail: string) {
    setSelectedDMUserId(userId);
    setSelectedDMUserEmail(userEmail);
    setSelectedProjectId(null);
  }

  return (
    <div className="app-container">
      {!token ? (
        <LoginForm message={message} setMessage={setMessage} />
      ) : (
        <div className="main-layout" style={{ display: 'flex', gap: '20px', padding: '20px' }}>
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
              showUsersMap={showUsersMap}
              showTasksMap={showTasksMap}
              toggleProjectUsers={toggleProjectUsers}
              toggleTasks={(id) => setShowTasksMap(p => ({ ...p, [id]: !p[id] }))}
              removeUser={removeUser}
              handleAddMember={addMemberToMap}
              onDeleteProject={handleDeleteProjectState}
              onUserAdded={handleLiveUserAdded}
              onProjectCreated={handleLiveProjectCreated}
              onUserRemoved={handleLiveUserRemoved}
              onSelectProject={(id) => setSelectedProjectId(id)}
              onStartDM={handleStartDM}
              onProjectUpdated={fetchProjects}
              taskRefreshTick={taskRefreshTick} // Corrected syntax here
            />
          </div>

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
                fetchUsersData(newP.id);
              } catch (err) { console.error(err); }
            }}
          />
        </div>
      )}
    </div>
  );
}

export default App;