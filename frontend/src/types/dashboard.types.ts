// TODO: 임시

export type DatabaseType = 'mongodb' | 'redis';
export type DatabaseSize = 'small' | 'medium' | 'large';
export type DatabaseMode = 'standalone' | 'replica' | 'cluster';
export type DatabaseStatus = 'provisioning' | 'running' | 'stopped' | 'error' | 'maintenance';

export interface DatabaseResources {
    cpu: number;
    memory: number;
    disk: number;
}

export interface DatabaseCost {
    creationCost: number;
    hourlyLemons: number;
    dailyLemons: number;
    monthlyLemons: number;
}

export interface DatabaseConfig {
    version: string;
    wiredTigerCache?: number;
    [key: string]: any;
}

export interface Database {
    id: string;
    name: string;
    type: DatabaseType;
    size: DatabaseSize;
    mode: DatabaseMode;
    status: DatabaseStatus;
    resources: DatabaseResources;
    cost: DatabaseCost;
    backupEnabled: boolean;
    config: DatabaseConfig;
    createdAt: string;
    updatedAt: string;
    createdFromPreset: string;
}

export interface DatabaseDetail extends Database {
    externalHost: string;
    externalPort: number;
    externalUriTemplate: string;
}