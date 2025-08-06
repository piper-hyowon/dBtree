import React from 'react';
import {ResourceSpec} from '../../types/database.types';
import './ResourceSelector.css';

interface ResourceSelectorProps {
    resources: ResourceSpec;
    onChange: (resources: ResourceSpec) => void;
}

const ResourceSelector: React.FC<ResourceSelectorProps> = ({resources, onChange}) => {
    const cpuOptions = [1, 2, 4, 8, 16];
    const memoryOptions = [512, 1024, 2048, 4096, 8192, 16384, 32768]; // MB
    const diskOptions = [10, 20, 50, 100, 200, 500, 1000]; // GB

    const handleCpuChange = (cpu: number) => {
        onChange({...resources, cpu});
    };

    const handleMemoryChange = (memory: number) => {
        onChange({...resources, memory});
    };

    const handleDiskChange = (disk: number) => {
        onChange({...resources, disk});
    };

    const formatMemory = (mb: number): string => {
        if (mb >= 1024) {
            return `${(mb / 1024).toFixed(mb % 1024 === 0 ? 0 : 1)} GB`;
        }
        return `${mb} MB`;
    };

    const formatDisk = (gb: number): string => {
        if (gb >= 1000) {
            return `${(gb / 1000).toFixed(gb % 1000 === 0 ? 0 : 1)} TB`;
        }
        return `${gb} GB`;
    };

    return (
        <div className="resource-selector">
            <div className="resource-group">
                <label>CPU</label>
                <div className="resource-options">
                    {cpuOptions.map(cpu => (
                        <button
                            key={cpu}
                            className={`resource-option ${resources.cpu === cpu ? 'selected' : ''}`}
                            onClick={() => handleCpuChange(cpu)}
                        >
                            {cpu} vCPU
                        </button>
                    ))}
                </div>
            </div>

            <div className="resource-group">
                <label>메모리</label>
                <div className="resource-options">
                    {memoryOptions.map(memory => (
                        <button
                            key={memory}
                            className={`resource-option ${resources.memory === memory ? 'selected' : ''}`}
                            onClick={() => handleMemoryChange(memory)}
                        >
                            {formatMemory(memory)}
                        </button>
                    ))}
                </div>
            </div>

            <div className="resource-group">
                <label>디스크</label>
                <div className="resource-options">
                    {diskOptions.map(disk => (
                        <button
                            key={disk}
                            className={`resource-option ${resources.disk === disk ? 'selected' : ''}`}
                            onClick={() => handleDiskChange(disk)}
                        >
                            {formatDisk(disk)}
                        </button>
                    ))}
                </div>
            </div>

            <div className="resource-summary">
                <div className="summary-item">
                    <span className="summary-label">선택된 리소스:</span>
                    <span className="summary-value">
                        {resources.cpu} vCPU • {formatMemory(resources.memory)} • {formatDisk(resources.disk)} SSD
                    </span>
                </div>
            </div>
        </div>
    );
};

export default ResourceSelector;