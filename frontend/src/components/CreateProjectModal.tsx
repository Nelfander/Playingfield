import React, { useState } from 'react';

interface CreateProjectModalProps {
    isOpen: boolean;
    onClose: () => void;
    onCreate: (name: string, description: string) => void;
}

const CreateProjectModal: React.FC<CreateProjectModalProps> = ({ isOpen, onClose, onCreate }) => {
    const [name, setName] = useState("");
    const [description, setDescription] = useState("");

    if (!isOpen) return null;

    return (
        <div className="modal-overlay">
            <div className="modal-content">
                <h2>Create New Project</h2>

                <label>Project Name</label>
                <input value={name} onChange={(e) => setName(e.target.value)} placeholder="Enter name" />

                <label>Description</label>
                <input value={description} onChange={(e) => setDescription(e.target.value)} placeholder="Enter description" />

                <div className="modal-actions">
                    <button className="btn-confirm" onClick={() => onCreate(name, description)}>Create</button>
                    <button className="btn-cancel" onClick={onClose}>Cancel</button>
                </div>
            </div>
        </div>
    );
};

export default CreateProjectModal;