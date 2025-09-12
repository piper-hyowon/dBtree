### 🌳 dBtree: 🍋 레몬나무에서 무료 데이터베이스를 수확하세요!

[Live](https://www.dbtree.cloud)

**개발 기간**: 2025-02 ~ (현재)

#### 기술 스택
- **Backend**: Go 1.24.0, PostgreSQL, Redis, Kubebuilder 4.7.1
- **Infra**: AWS EC2, SES, Kubernetes, K3s

#### 주요 기능
- 공유 레몬 나무 시스템: 고정된 10개 위치 랜덤 퀴즈, 수확 후 1시간 후 재생성
- 수확 프로세스: 퀴즈 정답 → 5초 내 움직이는 원 클릭 →  수확 완료
- DB 인스턴스 관리: CPU/Memory/Disk 커스터마이징, NodePort 자동 할당(30000~31999)
- 보안: 패스워드 1회 제공 후 서버 미저장, K8s Secret 관리
  

<img width="600" height="300" alt="image" src="https://github.com/user-attachments/assets/35037c91-c8bc-413b-aa58-90212e461c35" />


#### 기술적 특징
- 외부 웹 프레임워크, 라이브러리 의존성을 최소화하고 net/http, database/sql 등 Go 표준 라이브러리를 적극 활용
- HTTP 라우터, 미들웨어 체인, 에러 핸들링  직접 구현
- Hexagonal Architecture로 도메인/인프라 레이어 분리
- main.go에서 모든 의존성 수동 주입


#### 백엔드와 오퍼레이터 관심사 분리
<img width="700"  alt="image" src="https://github.com/user-attachments/assets/22e36627-c083-48a8-885a-5d2da4c61fd0" />


## 배포 
<img width="700" alt="image" src="https://github.com/user-attachments/assets/abfcf27f-933f-4b41-b12c-a9c92a1c15c7" />

배포 방법은 [manifests/Deployment.md](./manifests/deployment.md)를 참조


#### Architecture: Hexagonal Architecture

**프로젝트 구조**
```
internal/
├── core/              # 순수 도메인 (인터페이스, 타입, 비즈니스 규칙)
│   ├── auth/         
│   ├── lemon/        
│   ├── quiz/         
│   └── dbservice/    
├── {domain}/          # 도메인별 구현
│   ├── rest/          # Primary Adapter (HTTP)
│   ├── service.go     # Application Service
│   └── store.go       # Repository Factory
└── platform/          # Infrastructure
    ├── rest/          # HTTP 서버, 라우터, 미들웨어
    └── store/         # Secondary Adapter (DB)
```


- 모든 인터페이스를 `core` 패키지에 집중
- 컴파일 타임 인터페이스 검증
- 명시적 의존성 주입(main.go)

### 레몬 수확 미리보기
<img width="620" height="257" alt="image" src="https://github.com/user-attachments/assets/8fda58a0-a658-45e9-b942-c4056447418b" />

- Redis SETNX로 유저별 퀴즈 중복 시작 방지
- PostgreSQL FOR UPDATE로 레몬 행 레벨 락 획득하여 동시 클릭 경쟁 해결
- 트랜잭션 내에서 레몬 상태 변경 + 유저 잔액 업데이트 원자적 처리
- 정답 후 5초 내 가장 빠른 1명만 수확 성공, 나머지는 실패 처리


### DB 인스턴스 생성
<img width="354" height="202" alt="image" src="https://github.com/user-attachments/assets/799c75db-fe9f-438b-b2cc-ad10863cdbbd" />

- 유저별 최대 2개 인스턴스 생성 제한
- PostgreSQL UNIQUE 제약으로 NodePort(30000~31999) 충돌 방지
- K8s 리소스 생성 실패 시 레몬/포트 롤백


```go
defer func() {
	if err != nil {
		if lemonDeducted {
			s.lemonService.AddLemons(ctx, userID, cost, ActionInstanceCreateRefund)
		}
		if portAllocated {
			_ = s.portStore.ReleasePort(ctx, instance.ExternalID)
		}
	}
}()

```

### 리소스 모니터링
<img width="321" height="171" alt="image" src="https://github.com/user-attachments/assets/a907862f-8d7a-42fc-b386-e20e1790f2e8" />

- 전체 가용 자원 및 사용량 실시간 표시
- 생성 가능한 크기(Tiny/Small/Medium) 체크
- 생성 전 가용 자원 검증

### 레몬 트랜잭션 내역
<img width="548" height="292" alt="image" src="https://github.com/user-attachments/assets/bd3e4738-be51-4a68-a253-1a2136abbc60" />







