export type DBType = 'mongodb' | 'redis';
export type DBSize = 'tiny' | 'small' | 'medium' | 'large' | 'xlarge';

// MongoDB modes
export type MongoDBMode = 'standalone' | 'replica_set' | 'sharded';
// Redis modes
export type RedisMode = 'basic' | 'sentinel' | 'cluster';
export type DBMode = MongoDBMode | RedisMode;

export type InstanceStatus = 'provisioning' | 'running' | 'stopped' | 'error' | 'maintenance';

export interface ResourceSpec {
    cpu: number;      // vCPU 개수
    memory: number;   // MB 단위
    disk: number;     // GB 단위
}

export interface CostResponse {
    creationCost: number;   // 프로비저닝 비용
    hourlyLemons: number;   // 시간당 레몬
    dailyLemons: number;    // 일일 레몬
    monthlyLemons: number;  // 월간 레몬
}

export interface InstanceResponse {
    id: string;                      // ExternalID
    name: string;
    type: DBType;
    size: DBSize;
    mode: DBMode;
    status: InstanceStatus;
    statusReason?: string;
    resources: ResourceSpec;
    cost: CostResponse;
    endpoint?: string;               // 내부 엔드포인트
    port?: number;                   // 내부 포트
    externalHost?: string;           // 외부 접속 호스트
    externalPort?: number;           // 외부 접속 포트
    externalUriTemplate?: string;    // 연결 문자열 템플릿
    backupEnabled: boolean;
    config: Record<string, any>;
    createdAt: string;               // ISO 8601
    updatedAt: string;               // ISO 8601
    createdFromPreset?: string;      // 프리셋 ID
    pausedAt?: string;               // ISO 8601
}

export interface CreateInstanceRequest {
    name: string;
    presetId?: string;

    // 커스텀 옵션 (presetId 없을 때만)
    type?: DBType;
    mode?: DBMode;
    resources?: ResourceSpec;
    config?: Record<string, any>;

    // 백업 옵션
    backupEnabled: boolean;
    backupSchedule?: string;         // cron 표현식
    backupRetentionDays?: number;    // 0-365
}

export interface UpdateInstanceRequest {
    resources?: ResourceSpec;
    config?: Record<string, any>;
    backupEnabled?: boolean;
    backupSchedule?: string;
    backupRetentionDays?: number;
}

export interface PresetResponse {
    id: string;
    type: DBType;
    size: DBSize;
    mode: DBMode;
    name: string;
    icon: string;
    description: string;
    friendlyDescription: string;
    technicalTerms?: Record<string, any>;
    useCases: string[];
    resources: ResourceSpec;
    cost: CostResponse;
    defaultConfig?: Record<string, any>;
    sortOrder: number;
}

export interface EstimateCostRequest {
    type: DBType;
    resources: ResourceSpec;
    mode: DBMode;
    config?: Record<string, any>;
}

export interface EstimateCostResponse {
    cost: CostResponse;
    warnings?: string[];
    suggestions?: string[];
}

