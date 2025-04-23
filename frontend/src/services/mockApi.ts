import {
    ApiResponse,
    GlobalStatsResponse,
} from "../types/api.types";

interface BaseQuizQuestion {
    id: number;
    question: string;
    options: string[];
    correctOptionIndex: number;
}

export interface QuizQuestion extends BaseQuizQuestion {
    originalIndices?: number[];
}

interface QuizResponse {
    questions: QuizQuestion[];
}

const DB_QUIZ_QUESTIONS: QuizQuestion[] = [
    {
        id: 1,
        question: "SQL에서 데이터를 선택하는 명령어는?",
        options: ["INSERT", "SELECT", "UPDATE", "DELETE"],
        correctOptionIndex: 1, // SELECT
    },
    {
        id: 2,
        question: "데이터베이스에서 정규화(Normalization)의 주요 목적은?",
        options: [
            "데이터 중복성 감소",
            "데이터 접근 속도 증가",
            "파일 크기 증가",
            "쿼리 복잡성 증가",
        ],
        correctOptionIndex: 0, // 데이터 중복성 감소
    },
    {
        id: 3,
        question: "NoSQL 데이터베이스의 특징이 아닌 것은?",
        options: [
            "스키마 없는 설계",
            "수평적 확장성",
            "ACID 트랜잭션 기본 지원",
            "비관계형 데이터 모델",
        ],
        correctOptionIndex: 2, // ACID 트랜잭션 기본 지원
    },
    {
        id: 4,
        question: "인덱스(Index)를 사용하는 주된 이유는?",
        options: [
            "데이터 저장 공간 절약",
            "데이터 검색 속도 향상",
            "데이터 일관성 유지",
            "데이터 보안 강화",
        ],
        correctOptionIndex: 1, // 데이터 검색 속도 향상
    },
    {
        id: 5,
        question: "관계형 데이터베이스에서 기본 키(Primary Key)의 특징은?",
        options: [
            "중복 값 허용",
            "NULL 값 허용",
            "테이블 내 고유 식별자",
            "여러 열의 조합 불가능",
        ],
        correctOptionIndex: 2, // 테이블 내 고유 식별자
    },
    {
        id: 6,
        question: "조인(JOIN)의 주요 목적은?",
        options: [
            "데이터베이스 크기 줄이기",
            "여러 테이블의 데이터 결합",
            "데이터 삭제 용이성",
            "데이터 복제",
        ],
        correctOptionIndex: 1, // 여러 테이블의 데이터 결합
    },
    {
        id: 7,
        question: "트랜잭션의 ACID 속성 중 'I'는 무엇을 의미하는가?",
        options: [
            "식별성(Identity)",
            "무결성(Integrity)",
            "고립성(Isolation)",
            "불변성(Immutability)",
        ],
        correctOptionIndex: 2, // 고립성(Isolation)
    },
    {
        id: 8,
        question: "데이터베이스 복제(Replication)의 주요 이점은?",
        options: [
            "데이터 보안 강화",
            "가용성 및 내결함성 향상",
            "저장 공간 절약",
            "트랜잭션 속도 증가",
        ],
        correctOptionIndex: 1, // 가용성 및 내결함성 향상
    },
    {
        id: 9,
        question: "SQL 인젝션 공격을 방지하는 가장 좋은 방법은?",
        options: [
            "강력한 비밀번호 사용",
            "데이터베이스 암호화",
            "매개변수화된 쿼리 사용",
            "데이터베이스 백업",
        ],
        correctOptionIndex: 2, // 매개변수화된 쿼리 사용
    },
    {
        id: 10,
        question: "데이터베이스 샤딩(Sharding)의 주요 목적은?",
        options: [
            "데이터 보안 강화",
            "데이터베이스 수평적 확장",
            "데이터 중복 제거",
            "데이터 일관성 유지",
        ],
        correctOptionIndex: 1, // 데이터베이스 수평적 확장
    },
    {
        id: 11,
        question: "데이터베이스 트리거(Trigger)의 주요 용도는?",
        options: [
            "데이터 접근 제한",
            "데이터 변경 자동화",
            "데이터 백업",
            "사용자 인증",
        ],
        correctOptionIndex: 1, // 데이터 변경 자동화
    },
    {
        id: 12,
        question: "데이터베이스 뷰(View)의 주요 장점은?",
        options: [
            "데이터 저장 공간 절약",
            "데이터 보안 향상",
            "데이터 모델링 간소화",
            "데이터 백업 자동화",
        ],
        correctOptionIndex: 1, // 데이터 보안 향상
    },
    {
        id: 13,
        question: "데이터베이스에서 '외래 키(Foreign Key)'의 역할은?",
        options: [
            "데이터 암호화",
            "테이블 간 관계 정의",
            "인덱싱 성능 향상",
            "사용자 권한 제어",
        ],
        correctOptionIndex: 1, // 테이블 간 관계 정의
    },
    {
        id: 14,
        question: "SQL의 GROUP BY 절은 무엇을 위해 사용되나요?",
        options: ["데이터 정렬", "데이터 필터링", "데이터 집계", "데이터 조인"],
        correctOptionIndex: 2, // 데이터 집계
    },
    {
        id: 15,
        question: "데이터베이스에서 'ACID'는 무엇을 나타내나요?",
        options: [
            "보안 프로토콜",
            "트랜잭션 속성",
            "데이터 백업 방식",
            "쿼리 최적화 방법",
        ],
        correctOptionIndex: 1, // 트랜잭션 속성
    },
];


export const mockApi = {
    globalStats: async (): Promise<ApiResponse<GlobalStatsResponse>> => {
        return {
            success: true,
            data: {totalHarvested: 23, totalDbInstances: 6, activeUsers: 2},
        };
    },

    availableLemons: async (): Promise<ApiResponse<{ lemons: number[] }>> => {
        return {
            success: true,
            data: {
                lemons: [0, 1, 2, 3, 4, 5, 6, 7, 8, 9],
            },
        };
    },

    // TODO:
    harvestLemon: async (id: number): Promise<ApiResponse<boolean>> => {
        return {
            success: true,
            data: true,
        };
    },

    /**
     * DB 퀴즈 문항
     * @param count 요청할 문항 수 (기본값: 5)
     */
    getQuizQuestions: async (
        count: number = 5
    ): Promise<ApiResponse<QuizResponse>> => {
        await new Promise((resolve) => setTimeout(resolve, 200));

        const shuffled = [...DB_QUIZ_QUESTIONS].sort(() => 0.5 - Math.random());
        const selectedQuestions = shuffled.slice(
            0,
            Math.min(count, shuffled.length)
        );

        return {
            success: true,
            data: {
                questions: selectedQuestions,
            },
        };
    },

    /**
     * 퀴즈 정답 제출
     */
    submitQuizAnswer: async (
        questionId: number,
        selectedAnswer: number
    ): Promise<ApiResponse<{ correct: boolean }>> => {
        await new Promise((resolve) => setTimeout(resolve, 150));

        const question = DB_QUIZ_QUESTIONS.find((q) => q.id === questionId);

        if (!question) {
            return {
                success: false,
                error: "404",
            };
        }

        const isCorrect = question.correctOptionIndex === selectedAnswer;

        return {
            success: true,
            data: {
                correct: isCorrect,
            },
        };
    },
};

