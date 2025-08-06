import React, {useState, useEffect} from 'react';
import './CreateInstanceWizard.css';
import {
    DBType,
    DBMode,
    ResourceSpec,
    PresetResponse,
    CreateInstanceRequest,
    CostResponse,
    InstanceResponse
} from '../../types/database.types';
import {createInstance, getPresets} from '../../services/api/database.api';
import {useAuth} from '../../contexts/AuthContext';
import {useToast} from '../../hooks/useToast';
import {
    calculateCustomCost,
    validateInstanceName,
    canCreateInstance,
    BACKUP_SCHEDULES,
    BACKUP_SCHEDULE_LABELS
} from './costCalculator';
import PresetCard from './PresetCard';
import ResourceSelector from './ResourceSelector';
import BackupOptions from './BackupOptions';
import CostSummary from './CostSummary';
import CredentialsModal from './CredentialsModal';

interface CreateInstanceWizardProps {
    currentInstanceCount: number;
    onSuccess: () => void;
    onCancel: () => void;
}

const CreateInstanceWizard: React.FC<CreateInstanceWizardProps> = ({
                                                                       currentInstanceCount,
                                                                       onSuccess,
                                                                       onCancel
                                                                   }) => {
    const {user} = useAuth();
    const {showToast} = useToast();

    // 기본 상태
    const [isLoading, setIsLoading] = useState(false);
    const [presets, setPresets] = useState<PresetResponse[]>([]);
    const [selectedTab, setSelectedTab] = useState<'preset' | 'custom'>('preset');

    // 폼 데이터
    const [instanceName, setInstanceName] = useState('');
    const [nameError, setNameError] = useState<string | null>(null);

    // 프리셋 선택
    const [selectedPreset, setSelectedPreset] = useState<PresetResponse | null>(null);

    // 커스텀 설정
    const [customType, setCustomType] = useState<DBType>('mongodb');
    const [customMode, setCustomMode] = useState<DBMode>('standalone');
    const [customResources, setCustomResources] = useState<ResourceSpec>({
        cpu: 2,
        memory: 2048,
        disk: 20
    });

    // 백업 설정
    const [backupEnabled, setBackupEnabled] = useState(false);
    const [backupSchedule, setBackupSchedule] = useState<string>(BACKUP_SCHEDULES.DAILY_2AM);
    const [backupRetentionDays, setBackupRetentionDays] = useState(7);

    // 비용 계산
    const [estimatedCost, setEstimatedCost] = useState<CostResponse>({
        creationCost: 0,
        hourlyLemons: 0,
        dailyLemons: 0,
        monthlyLemons: 0
    });

    // 생성 완료 모달
    const [showCredentials, setShowCredentials] = useState(false);
    const [createdInstance, setCreatedInstance] = useState<InstanceResponse | null>(null);

    // 프리셋 로드
    useEffect(() => {
        loadPresets();
    }, []);

    // 비용 계산
    useEffect(() => {
        if (selectedTab === 'preset' && selectedPreset) {
            setEstimatedCost(selectedPreset.cost);
        } else if (selectedTab === 'custom') {
            const cost = calculateCustomCost(customType, customResources);
            setEstimatedCost(cost);
        }
    }, [selectedTab, selectedPreset, customType, customResources]);

    // 이름 유효성 검사
    useEffect(() => {
        if (instanceName) {
            const validation = validateInstanceName(instanceName);
            setNameError(validation.valid ? null : validation.error || null);
        } else {
            setNameError(null);
        }
    }, [instanceName]);

    const loadPresets = async () => {
        try {
            const data = await getPresets();
            // MongoDB 프리셋만 필터링 (현재 Redis는 미지원)
            const mongoPresets = data.filter(p => p.type === 'mongodb');
            setPresets(mongoPresets);

            // 첫 번째 프리셋 자동 선택
            if (mongoPresets.length > 0) {
                setSelectedPreset(mongoPresets[0]);
            }
        } catch (error) {
            showToast('프리셋을 불러오는데 실패했습니다', 'error');
        }
    };

    const handleTypeChange = (type: DBType) => {
        if (type === 'redis') {
            showToast('Redis는 현재 준비 중입니다', 'info');
            return;
        }
        setCustomType(type);

        // 타입에 따른 기본 모드 설정
        if (type === 'mongodb') {
            setCustomMode('standalone');
        }
    };

    const handleModeChange = (mode: DBMode) => {
        setCustomMode(mode);
    };

    const handleResourceChange = (resources: ResourceSpec) => {
        setCustomResources(resources);
    };

    const handleCreate = async () => {
        // 유효성 검사
        const nameValidation = validateInstanceName(instanceName);
        if (!nameValidation.valid) {
            setNameError(nameValidation.error || '유효하지 않은 이름입니다');
            return;
        }

        // 생성 가능 여부 체크
        const lemonBalance = user?.lemonBalance || 0;
        const createCheck = canCreateInstance(lemonBalance, estimatedCost, currentInstanceCount);
        if (!createCheck.canCreate) {
            showToast(createCheck.reason || '인스턴스를 생성할 수 없습니다', 'error');
            return;
        }

        setIsLoading(true);

        try {
            let request: CreateInstanceRequest;

            if (selectedTab === 'preset' && selectedPreset) {
                // 프리셋 기반 생성
                request = {
                    name: instanceName,
                    presetId: selectedPreset.id,
                    backupEnabled: backupEnabled,
                    backupSchedule: backupEnabled ? backupSchedule : undefined,
                    backupRetentionDays: backupEnabled ? backupRetentionDays : undefined
                };
            } else {
                // 커스텀 생성
                request = {
                    name: instanceName,
                    type: customType,
                    mode: customMode,
                    resources: customResources,
                    backupEnabled: backupEnabled,
                    backupSchedule: backupEnabled ? backupSchedule : undefined,
                    backupRetentionDays: backupEnabled ? backupRetentionDays : undefined
                };
            }

            const response = await createInstance(request);

            // 생성 성공
            setCreatedInstance(response);
            setShowCredentials(true);
            showToast('인스턴스가 생성되었습니다', 'success');

        } catch (error: any) {
            const errorMessage = error?.response?.data?.error || '인스턴스 생성에 실패했습니다';
            showToast(errorMessage, 'error');
        } finally {
            setIsLoading(false);
        }
    };

    const handleCredentialsClose = () => {
        setShowCredentials(false);
        onSuccess();
    };

    const isCreateDisabled = () => {
        if (!instanceName || nameError) return true;
        if (isLoading) return true;

        const lemonBalance = user?.lemonBalance || 0;
        const check = canCreateInstance(lemonBalance, estimatedCost, currentInstanceCount);

        return !check.canCreate;
    };

    const getCreateButtonText = () => {
        if (isLoading) return '생성 중...';

        const lemonBalance = user?.lemonBalance || 0;
        const check = canCreateInstance(lemonBalance, estimatedCost, currentInstanceCount);

        if (!check.canCreate) {
            if (currentInstanceCount >= 2) return '인스턴스 개수 초과';
            if (lemonBalance < estimatedCost.creationCost + estimatedCost.hourlyLemons) {
                return `레몬 부족 (${estimatedCost.creationCost + estimatedCost.hourlyLemons}🍋 필요)`;
            }
        }

        return '인스턴스 생성';
    };

    return (
        <div className="create-instance-wizard">
            <div className="wizard-header">
                <h2>새 인스턴스 생성</h2>
                <p>데이터베이스 인스턴스를 생성합니다</p>
            </div>

            <div className="wizard-content">
                {/* 기본 정보 */}
                <div className="wizard-section">
                    <h3>기본 정보</h3>
                    <div className="form-group">
                        <label htmlFor="instance-name">
                            인스턴스 이름 <span className="required">*</span>
                        </label>
                        <input
                            id="instance-name"
                            type="text"
                            value={instanceName}
                            onChange={(e) => setInstanceName(e.target.value.toLowerCase())}
                            placeholder="my-instance"
                            className={nameError ? 'error' : ''}
                            maxLength={63}
                        />
                        {nameError && (
                            <div className="error-message">{nameError}</div>
                        )}
                        <div className="help-text">
                            소문자, 숫자, 하이픈(-)만 사용 가능 (3-63자)
                        </div>
                    </div>
                </div>

                {/* 구성 선택 */}
                <div className="wizard-section">
                    <h3>구성</h3>
                    <div className="config-tabs">
                        <button
                            className={`tab-button ${selectedTab === 'preset' ? 'active' : ''}`}
                            onClick={() => setSelectedTab('preset')}
                        >
                            프리셋
                        </button>
                        <button
                            className={`tab-button ${selectedTab === 'custom' ? 'active' : ''}`}
                            onClick={() => setSelectedTab('custom')}
                        >
                            ⚙️   커스텀
                        </button>
                    </div>

                    <div className="tab-content">
                        {selectedTab === 'preset' ? (
                            <div className="preset-grid">
                                {presets.map(preset => (
                                    <PresetCard
                                        key={preset.id}
                                        preset={preset}
                                        selected={selectedPreset?.id === preset.id}
                                        onSelect={() => setSelectedPreset(preset)}
                                    />
                                ))}
                            </div>
                        ) : (
                            <div className="custom-config">
                                <div className="form-group">
                                    <label>데이터베이스 타입</label>
                                    <div className="type-selector">
                                        <button
                                            className={`type-button ${customType === 'mongodb' ? 'active' : ''}`}
                                            onClick={() => handleTypeChange('mongodb')}
                                        >
                                            🍃 MongoDB
                                        </button>
                                        <button
                                            className="type-button disabled"
                                            onClick={() => handleTypeChange('redis')}
                                        >
                                            🔴 Redis (준비 중)
                                        </button>
                                    </div>
                                </div>

                                {customType === 'mongodb' && (
                                    <div className="form-group">
                                        <label>모드</label>
                                        <div className="mode-selector">
                                            <button
                                                className={`mode-button ${customMode === 'standalone' ? 'active' : ''}`}
                                                onClick={() => handleModeChange('standalone')}
                                            >
                                                Standalone
                                            </button>
                                            <button
                                                className={`mode-button ${customMode === 'replica_set' ? 'active' : ''}`}
                                                onClick={() => handleModeChange('replica_set')}
                                            >
                                                Replica Set
                                            </button>
                                            <button
                                                className={`mode-button ${customMode === 'sharded' ? 'active' : ''}`}
                                                onClick={() => handleModeChange('sharded')}
                                            >
                                                Sharded
                                            </button>
                                        </div>
                                    </div>
                                )}

                                <ResourceSelector
                                    resources={customResources}
                                    onChange={handleResourceChange}
                                />
                            </div>
                        )}
                    </div>
                </div>

                {/* 백업 옵션 */}
                <BackupOptions
                    enabled={backupEnabled}
                    schedule={backupSchedule}
                    retentionDays={backupRetentionDays}
                    onEnabledChange={setBackupEnabled}
                    onScheduleChange={setBackupSchedule}
                    onRetentionChange={setBackupRetentionDays}
                />

                {/* 비용 요약 */}
                <CostSummary
                    cost={estimatedCost}
                    lemonBalance={user?.lemonBalance || 0}
                />

                {/* 액션 버튼 */}
                <div className="wizard-actions">
                    <button
                        className="btn-cancel"
                        onClick={onCancel}
                        disabled={isLoading}
                    >
                        취소
                    </button>
                    <button
                        className="btn-create"
                        onClick={handleCreate}
                        disabled={isCreateDisabled()}
                    >
                        {getCreateButtonText()}
                    </button>
                </div>
            </div>

            {/* 인증 정보 모달 */}
            {showCredentials && createdInstance && (
                <CredentialsModal
                    instance={createdInstance}
                    onClose={handleCredentialsClose}
                />
            )}
        </div>
    );
};

export default CreateInstanceWizard;