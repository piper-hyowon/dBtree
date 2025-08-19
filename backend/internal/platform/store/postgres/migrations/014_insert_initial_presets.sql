-- 기존 프리셋 데이터 전체 삭제 (프리셋 스펙 재조정)
TRUNCATE TABLE db_presets RESTART IDENTITY;

-- 조정된 스펙으로 프리셋 재입력
INSERT INTO db_presets (id, type, size, mode, name, icon, description,
                        friendly_description, technical_terms, use_cases,
                        cpu, memory, disk, creation_cost, hourly_cost,
                        default_config, sort_order, available, unavailable_reason)
VALUES
-- MongoDB Tiny (0.1 vCPU, 512MB, 5GB)
('mongodb-standalone-tiny', 'mongodb', 'tiny', 'standalone',
 'MongoDB Tiny', '🔬',
 'Tiny MongoDB 7.0 인스턴스 - 0.1 vCPU, 512MB RAM, 5GB SSD',
 '가장 작은 MongoDB 인스턴스예요. 개발이나 학습용으로 적합하며, 간단한 테스트나 프로토타입 개발에 충분해요.',
 '{
   "MongoDB": "문서 지향 NoSQL 데이터베이스",
   "Standalone": "단일 노드 인스턴스",
   "WiredTiger": "기본 스토리지 엔진"
 }'::jsonb,
 ARRAY ['개발 환경', '학습용', '프로토타입', 'CI/CD 테스트'],
 0.1, 512, 5, 5, 1,
 '{
   "version": "7.0",
   "wiredTigerCacheSizeGB": 0.1,
   "journalEnabled": true
 }'::jsonb,
 5, true, NULL),

-- MongoDB Small (0.25 vCPU, 768, 10GB)
('mongodb-standalone-small', 'mongodb', 'small', 'standalone',
 'MongoDB Small', '📦',
 'Small MongoDB 7.0 인스턴스 - 0.25 vCPU, 768MB RAM, 10GB SSD',
 '소규모 애플리케이션에 적합한 MongoDB 인스턴스예요. 개인 프로젝트나 스타트업 MVP에 충분한 성능을 제공해요.',
 '{
   "MongoDB": "문서 지향 NoSQL 데이터베이스",
   "Standalone": "단일 노드 인스턴스",
   "WiredTiger": "문서 레벨 잠금과 압축을 지원하는 스토리지 엔진"
 }'::jsonb,
 ARRAY ['소규모 프로덕션', 'MVP', '개인 프로젝트', '블로그'],
 0.25, 768, 10, 10, 2,
 '{
   "version": "7.0",
   "wiredTigerCacheSizeGB": 0.25,
   "journalEnabled": true
 }'::jsonb,
 10, true, NULL),

-- MongoDB Medium (0.5 vCPU, 1GB, 20GB) - 현재 비활성화
('mongodb-standalone-medium', 'mongodb', 'medium', 'standalone',
 'MongoDB Medium', '📈',
 'Medium MongoDB 7.0 인스턴스 - 0.5 vCPU, 1GB RAM, 20GB SSD',
 '중간 규모 서비스를 위한 MongoDB 인스턴스예요. 일반적인 웹 애플리케이션에 적합해요.',
 '{
   "MongoDB": "문서 지향 NoSQL 데이터베이스",
   "Collection": "관련 문서들의 그룹",
   "Index": "쿼리 성능 향상을 위한 데이터 구조"
 }'::jsonb,
 ARRAY ['중간 규모 서비스', '웹 애플리케이션', 'API 서버'],
 0.5, 1024, 20, 20, 3,
 '{
   "version": "7.0",
   "wiredTigerCacheSizeGB": 0.5,
   "journalEnabled": true
 }'::jsonb,
 20, false, '서버 리소스 확장 후 지원 예정입니다'),

-- MongoDB Large (0.75 vCPU, 1.5GB, 30GB) - 현재 비활성화
('mongodb-standalone-large', 'mongodb', 'large', 'standalone',
 'MongoDB Large', '🚀',
 'Large MongoDB 7.0 인스턴스 - 0.75 vCPU, 1.5GB RAM, 30GB SSD',
 '대규모 서비스를 위한 MongoDB 인스턴스예요. 높은 트래픽과 대용량 데이터를 처리할 수 있어요.',
 '{
   "MongoDB": "문서 지향 NoSQL 데이터베이스",
   "Aggregation Pipeline": "복잡한 데이터 처리를 위한 파이프라인"
 }'::jsonb,
 ARRAY ['대규모 프로덕션', '데이터 분석', '실시간 처리'],
 0.75, 1536, 30, 30, 5,
 '{
   "version": "7.0",
   "wiredTigerCacheSizeGB": 0.75,
   "journalEnabled": true
 }'::jsonb,
 30, false, '서버 리소스 확장 후 지원 예정입니다');