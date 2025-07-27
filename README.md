[English README](./README.en.md)
### 🌳 dBtree: 🍋 레몬나무에서 무료 데이터베이스를 수확하세요!

[서비스 미리보기](https://dbtree.vercel.app/)

**개발 기간**: 2025-02 ~ (현재)

#### 🛠 기술 스택
- **Backend**: Go (표준 라이브러리 중심), PostgreSQL, Redis
- **Infrastructure**: Kubernetes (Custom Resource Definition)
- **Frontend**: React

#### ✨ 서비스 특징
- 🎮 게이미피케이션: DB 퀴즈를 풀어 레몬(크레딧) 획득
- 🍋 획득한 레몬으로 Redis/MongoDB 인스턴스 프로비저닝
- 📚 DB 학습과 실습을 동시에
- 🎨 귀여운 UI로 즐거운 개발 경험
  

<img width="600" height="300" alt="image" src="https://github.com/user-attachments/assets/35037c91-c8bc-413b-aa58-90212e461c35" />

<hr>

<img width="338" height="551" alt="image" src="https://github.com/user-attachments/assets/478afbd5-48ac-4875-aeee-15636f1fcc75" />


#### Architecture: Hexagonal Architecture


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
