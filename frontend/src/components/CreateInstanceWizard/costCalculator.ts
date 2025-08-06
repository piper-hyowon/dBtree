import {DBType, ResourceSpec, CostResponse} from '../../types/database.types';

/**
 * 백엔드 CalculateCustomCost 로직과 동일
 */
export function calculateCustomCost(dbType: DBType, resources: ResourceSpec): CostResponse {
    let base = 0;

    // 메모리 기반 비용
    switch (dbType) {
        case 'redis':
            base = Math.floor(resources.memory / 512); // 512MB당 1레몬
            break;
        case 'mongodb':
            base = Math.floor(resources.memory / 1024) * 3; // 1GB당 3레몬
            break;
    }

    // CPU 추가 비용 (1 vCPU 초과분에 대해)
    if (resources.cpu > 1) {
        base += (resources.cpu - 1) * 2;
    }

    // 디스크 추가 비용 (10GB 초과분에 대해)
    if (resources.disk > 10) {
        base += Math.floor((resources.disk - 10) / 10);
    }

    // 최소값 보장
    if (base < 1) {
        base = 1;
    }

    const hourlyLemons = base;
    const creationCost = base * 10;

    return {
        creationCost,
        hourlyLemons,
        dailyLemons: hourlyLemons * 24,
        monthlyLemons: hourlyLemons * 24 * 30
    };
}

/**
 * 리소스 사이즈 계산
 */
export function calculateSize(resources: ResourceSpec): 'tiny' | 'small' | 'medium' | 'large' | 'xlarge' {
    const totalPoints = resources.cpu + (resources.memory / 1024) + (resources.disk / 50);

    if (totalPoints <= 3) return 'tiny';
    if (totalPoints <= 6) return 'small';
    if (totalPoints <= 12) return 'medium';
    if (totalPoints <= 24) return 'large';
    return 'xlarge';
}

/**
 * 인스턴스 이름 유효성 검사
 */
export function validateInstanceName(name: string): { valid: boolean; error?: string } {
    if (!name || name.trim().length === 0) {
        return {valid: false, error: '이름을 입력해주세요'};
    }

    if (name.length < 3) {
        return {valid: false, error: '이름은 최소 3자 이상이어야 합니다'};
    }

    if (name.length > 63) {
        return {valid: false, error: '이름은 최대 63자까지 가능합니다'};
    }

    // 소문자, 숫자, 하이픈만 허용 (시작과 끝은 영숫자)
    const nameRegex = /^[a-z0-9]([-a-z0-9]*[a-z0-9])?$/;
    if (!nameRegex.test(name)) {
        return {valid: false, error: '소문자, 숫자, 하이픈(-)만 사용 가능하며, 시작과 끝은 영숫자여야 합니다'};
    }

    return {valid: true};
}

/**
 * 백업 스케줄 Cron 표현식 프리셋
 */
export const BACKUP_SCHEDULES = {
    DAILY_2AM: '0 2 * * *',        // 매일 새벽 2시
    DAILY_4AM: '0 4 * * *',        // 매일 새벽 4시
    WEEKLY_SUN: '0 2 * * 0',       // 매주 일요일 새벽 2시
    WEEKLY_SAT: '0 2 * * 6',       // 매주 토요일 새벽 2시
    TWICE_DAILY: '0 2,14 * * *',   // 매일 새벽 2시, 오후 2시
} as const;

export const BACKUP_SCHEDULE_LABELS = {
    [BACKUP_SCHEDULES.DAILY_2AM]: '매일 새벽 2시',
    [BACKUP_SCHEDULES.DAILY_4AM]: '매일 새벽 4시',
    [BACKUP_SCHEDULES.WEEKLY_SUN]: '매주 일요일 새벽 2시',
    [BACKUP_SCHEDULES.WEEKLY_SAT]: '매주 토요일 새벽 2시',
    [BACKUP_SCHEDULES.TWICE_DAILY]: '매일 새벽 2시, 오후 2시',
};

/**
 * 생성 가능 여부 체크
 */
export function canCreateInstance(
    lemonBalance: number,
    cost: CostResponse,
    currentInstanceCount: number
): { canCreate: boolean; reason?: string } {
    // 인스턴스 개수 제한
    if (currentInstanceCount >= 2) {
        return {
            canCreate: false,
            reason: '최대 2개의 인스턴스만 생성할 수 있습니다'
        };
    }

    // 레몬 잔액 체크 (생성 비용 + 최소 1시간 운영 비용)
    const requiredLemons = cost.creationCost + cost.hourlyLemons;
    if (lemonBalance < requiredLemons) {
        return {
            canCreate: false,
            reason: `레몬이 부족합니다 (필요: ${requiredLemons}🍋, 현재: ${lemonBalance}🍋)`
        };
    }

    return {canCreate: true};
}