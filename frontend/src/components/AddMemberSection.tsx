import React, { useState } from 'react';
import UserSelect from './UserSelect';

interface AddMemberSectionProps {
    projectId: number;
    onAdd: (projectId: number, userId: number) => void;
    excludeIds: number[];
}

const AddMemberSection: React.FC<AddMemberSectionProps> = ({ projectId, onAdd, excludeIds }) => {
    const [selectedId, setSelectedId] = useState("");

    const handleAdd = () => {
        if (!selectedId) return;
        onAdd(projectId, parseInt(selectedId));
        setSelectedId("");
    };

    return (
        <div className="inner-add-section">
            <p>Add Member to Project</p>
            <div className="add-member-controls">
                <UserSelect onUserChange={setSelectedId} excludeIds={excludeIds} />
                <button onClick={handleAdd} className="btn-success">Add</button>
            </div>
        </div>
    );
};

export default AddMemberSection;