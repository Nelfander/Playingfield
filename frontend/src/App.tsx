import { useState, useEffect } from "react";
import ProjectList from "./components/ProjectList";
import LoginForm from "./components/LoginForm";
import CreateProjectModal from "./components/CreateProjectModal";
import { ChatBox } from "./components/ChatBox";
import { DirectMessageBox } from "./components/DirectMessageBox";
import { type UserInProject } from "./components/ProjectUsers";
import { useWebSockets } from "./hooks/useWebSockets";
import AdminDashboard from "./components/AdminDashboard";
import { apiUrl } from './config/env';
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
  const [isAdminView, setIsAdminView] = useState(false);
  const [projectUsersMap, setProjectUsersMap] = useState<Record<number, UserInProject[]>>({});
  const [showUsersMap, setShowUsersMap] = useState<Record<number, boolean>>({});
  const [showTasksMap, setShowTasksMap] = useState<Record<number, boolean>>({});
  const [selectedProjectId, setSelectedProjectId] = useState<number | null>(null);
  const [selectedDMUserId, setSelectedDMUserId] = useState<number | null>(null);
  const [selectedDMUserEmail, setSelectedDMUserEmail] = useState<string>("");

  // --- ONBOARDING STATE ---
  const [showInstructions, setShowInstructions] = useState(true);

  const token = localStorage.getItem("token");
  const currentUserId = Number(localStorage.getItem("userId")) || 0;
  const userRole = localStorage.getItem("role");

  // --- PERSISTENT UNREAD MESSAGE STATES ---
  const [unreadProjects, setUnreadProjects] = useState<Record<number, boolean>>(() => {
    const saved = localStorage.getItem("unreadProjects");
    return saved ? JSON.parse(saved) : {};
  });

  const [unreadDMs, setUnreadDMs] = useState<Record<number, boolean>>(() => {
    const saved = localStorage.getItem("unreadDMs");
    return saved ? JSON.parse(saved) : {};
  });

  useEffect(() => {
    localStorage.setItem("unreadProjects", JSON.stringify(unreadProjects));
  }, [unreadProjects]);

  useEffect(() => {
    localStorage.setItem("unreadDMs", JSON.stringify(unreadDMs));
  }, [unreadDMs]);

  const [taskRefreshTick, setTaskRefreshTick] = useState(0);
  const [adminRefreshTick, setAdminRefreshTick] = useState(0);

  const handleLogout = () => {
    localStorage.removeItem("token");
    localStorage.removeItem("userId");
    localStorage.removeItem("role");
    window.location.reload();
  };

  const forceLogout = (reason: string) => {
    alert(reason);
    handleLogout();
  };

  // --- WebSocket Handlers ---
  const handleTaskSignal = (projectId: number) => {
    console.log(`WS Signal: Task change in project ${projectId}`);
    setTaskRefreshTick(prev => prev + 1);
  };

  const handleUserScrubbed = (userId: number) => {
    if (userId === currentUserId) {
      forceLogout("Your account has been deleted by an administrator.");
    }
    setAdminRefreshTick(prev => prev + 1);
    fetchProjects();
  };

  const handleUserStatusUpdated = (userId: number, newStatus: string) => {
    if (userId === currentUserId && newStatus === "inactive") {
      forceLogout("Your account has been deactivated.");
    }
    setAdminRefreshTick(prev => prev + 1);
  };

  const handleNewChatMessage = (projectId: number, senderId: number) => {
    const pId = Number(projectId);
    const sId = Number(senderId);
    const myId = Number(currentUserId);
    const activeP = selectedProjectId !== null ? Number(selectedProjectId) : null;

    if (sId !== myId && activeP !== pId) {
      setUnreadProjects(prev => ({ ...prev, [pId]: true }));
    }
  };

  const handleNewDM = (senderId: number) => {
    const sId = Number(senderId);
    const myId = Number(currentUserId);
    const activeDM = selectedDMUserId !== null ? Number(selectedDMUserId) : null;

    if (sId !== myId && activeDM !== sId) {
      setUnreadDMs(prev => ({ ...prev, [sId]: true }));
    }
  };

  useWebSockets(
    token,
    (id) => handleDeleteProjectState(id),
    (pId, uId, role) => handleLiveUserAdded(pId, uId, role),
    () => handleLiveProjectCreated(),
    (pId, uId) => handleLiveUserRemoved(pId, uId),
    () => fetchProjects(),
    (pId) => handleTaskSignal(pId),
    (pId) => handleTaskSignal(pId),
    (pId) => handleTaskSignal(pId),
    (uId) => handleUserScrubbed(uId),
    (uId, status) => handleUserStatusUpdated(uId, status),
    (uId) => {
      console.log(`New user detected: ${uId}`);
      setAdminRefreshTick(prev => prev + 1);
    },
    handleNewChatMessage,
    handleNewDM
  );

  async function fetchUsersData(projectId: number) {
    if (!token) return;
    try {
      const res = await fetch(apiUrl(`projects/users?project_id=${projectId}`), {
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
      const res = await fetch(apiUrl("projects"), {
        headers: { Authorization: `Bearer ${token}` },
      });
      const data = await res.json();
      const projectData: Project[] = data || [];
      setProjects(projectData);
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
      setShowInstructions(false); // Hide instructions when user takes action
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

  function handleLiveUserAdded(projectId: number, userId: number, _role: string) {
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
      const res = await fetch(apiUrl("projects/users"), {
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
      await fetch(apiUrl("projects/users"), {
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
    setUnreadDMs(prev => {
      const updated = { ...prev };
      delete updated[userId];
      return updated;
    });
    setSelectedDMUserId(userId);
    setSelectedDMUserEmail(userEmail);
    setSelectedProjectId(null);
  }

  return (
    <div className="app-container">
      {!token ? (
        <LoginForm message={message} setMessage={setMessage} />
      ) : isAdminView && userRole === "admin" ? (
        <div style={adminLayoutStyles.overlay}>
          <div style={adminLayoutStyles.navContainer}>
            <button
              onClick={() => setIsAdminView(false)}
              style={adminLayoutStyles.backBtn}
            >
              <span>←</span> Return to User Dashboard
            </button>
          </div>
          <AdminDashboard
            token={token}
            currentUserId={currentUserId}
            refreshTick={adminRefreshTick}
          />
        </div>
      ) : (
        <div className="main-layout">
          {/* LEFT SIDE: Project List Container */}
          <div className="project-list-container">
            {/* --- HELP ICON --- */}
            <button
              onClick={() => setShowInstructions(!showInstructions)}
              className="help-icon-btn"
              title="Need help?"
            >
              ?
            </button>

            <h1>My Projects</h1>

            {/* --- ONBOARDING GUIDE --- */}
            {showInstructions && (
              <div className="onboarding-guide">
                <h2>Welcome to the Team Space 🏔️</h2>
                <div className="guide-steps">
                  <div className="step">
                    <span>🚀</span>
                    <p><strong>Create:</strong> Use the green button below to start a new project board.</p>
                  </div>
                  <div className="step">
                    <span>👥</span>
                    <p><strong>Invite:</strong> Inside a project, click the <strong>Add Member arrow</strong> to add other users to the project.</p>
                  </div>
                  <div className="step">
                    <span>💬</span>
                    <p><strong>Chat:</strong> Each project features a dedicated real-time chat room. Click the <strong>Chat icon</strong> to open the sticky chat window.</p>
                  </div>
                  <div className="step">
                    <span>📋</span>
                    <p><strong>Tasks:</strong> Inside a project, click the <strong>Add New Task</strong> to create a new task.</p>
                  </div>
                  <div className="step">
                    <span>📤</span>
                    <p><strong>Attach Files:</strong> Attach files to tasks and upload them for everybody to download.</p>
                  </div>
                  <div className="step">
                    <span>📖</span>
                    <p><strong>History:</strong> Read the history of tasks and see who has done what.</p>
                  </div>
                </div>
                <button className="dismiss-btn" onClick={() => setShowInstructions(false)}>
                  Got it, thanks!
                </button>
              </div>
            )}

            <div className="button-group">
              <button onClick={handleProjectToggle}>
                {showProjects ? "Hide Projects" : "Load Projects"}
              </button>
              <button onClick={() => setIsModalOpen(true)} className="btn-success">
                Create Project
              </button>

              {userRole === "admin" && (
                <button
                  onClick={() => setIsAdminView(true)}
                  style={{ backgroundColor: '#ff4d4d', color: 'white', fontWeight: 'bold' }}
                >
                  🛡️ Admin Panel
                </button>
              )}

              <button onClick={handleLogout}>
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
              onSelectProject={(id) => {
                const pId = Number(id);
                setUnreadProjects(prev => {
                  const updated = { ...prev };
                  delete updated[pId];
                  return updated;
                });
                setSelectedProjectId(pId);
                setSelectedDMUserId(null);
              }}
              onStartDM={handleStartDM}
              onProjectUpdated={fetchProjects}
              taskRefreshTick={taskRefreshTick}
              unreadProjects={unreadProjects}
              unreadDMs={unreadDMs}
            />
          </div>

          {/* RIGHT SIDE: Sticky Chat Sidebar Container */}
          {(selectedProjectId || selectedDMUserId) && (
            <div className="chat-sidebar">
              {selectedProjectId ? (
                <ChatBox
                  projectId={selectedProjectId}
                  token={token!}
                  onClose={() => setSelectedProjectId(null)}
                />
              ) : (
                <DirectMessageBox
                  otherUserId={selectedDMUserId!}
                  otherUserEmail={selectedDMUserEmail}
                  token={token!}
                  onClose={() => setSelectedDMUserId(null)}
                />
              )}
            </div>
          )}

          <CreateProjectModal
            isOpen={isModalOpen}
            onClose={() => setIsModalOpen(false)}
            onCreate={async (name, desc) => {
              try {
                const res = await fetch(apiUrl("projects"), {
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
                setShowInstructions(false);
                fetchUsersData(newP.id);
              } catch (err) { console.error(err); }
            }}
          />
        </div>
      )}
    </div>
  );
}

const adminLayoutStyles: Record<string, React.CSSProperties> = {
  overlay: {
    minHeight: '100vh',
    background: 'rgba(15, 23, 42, 0.4)',
    backdropFilter: 'blur(8px)',
    WebkitBackdropFilter: 'blur(8px)',
    paddingTop: '20px',
    transition: 'all 0.5s ease'
  },
  navContainer: {
    maxWidth: '1000px',
    margin: '0 auto',
    padding: '0 40px',
    display: 'flex',
    justifyContent: 'flex-start'
  },
  backBtn: {
    background: 'rgba(255, 255, 255, 0.1)',
    backdropFilter: 'blur(10px)',
    WebkitBackdropFilter: 'blur(10px)',
    border: '1px solid rgba(255, 255, 255, 0.2)',
    color: '#fff',
    padding: '10px 20px',
    borderRadius: '12px',
    cursor: 'pointer',
    fontSize: '14px',
    fontWeight: '500',
    transition: 'all 0.3s ease',
    display: 'flex',
    alignItems: 'center'
  }
};

export default App;