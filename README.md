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


**1. 모든 인터페이스를 `core` 패키지에 집중**
- 순환 참조를 구조적으로 방지
- 의존성 방향의 단일화
```
  adapter ──▶ core/{domain} ◀── implementation
  ▲
  │
  infrastructure
```


```go
// 도메인 간 직접 참조 (X)
import "github.com/piper-hyowon/dBtree/internal/lemon"

// core 인터페이스만 참조(O)
import "github.com/piper-hyowon/dBtree/internal/core/lemon"
```

**2. 컴파일 타임 인터페이스 검증**
```go
var _ quiz.Store = (*QuizStore)(nil)
```


**3. 도메인별 저장소 전략**
```go
// 단일 저장소: 영속성만 필요한 경우
userStore := user.NewStore(db)  // PostgreSQL only

// 복합 저장소: 상태와 영속성 분리가 필요한 경우
quizStore := quiz.NewStore(cache, db)  // Redis + PostgreSQL
```

**4. 도메인 객체의 비즈니스 규칙 캡슐화**
```go
func (o *OTP) Verify(code string) bool {
    if time.Now().UTC().After(o.ExpiresAt) {
        return false
    }
}
```


**5. 명시적 의존성 주입**
```go
// main.go 에서 명시적 의존성 주입
sessionStore := auth.NewSessionStore(appConfig.UseLocalMemoryStore, pgClient.DB())
userStore := user.NewStore(appConfig.UseLocalMemoryStore, pgClient.DB())
lemonStore := lemon.NewLemonStore(appConfig.UseLocalMemoryStore, pgClient.DB())
quizStore := quiz.NewStore(redisClient.Redis(), pgClient.DB())

authService := auth.NewService(
sessionStore,
emailService,
userStore,
logger,
)

authHandler := authRest.NewHandler(authService, logger)
```


