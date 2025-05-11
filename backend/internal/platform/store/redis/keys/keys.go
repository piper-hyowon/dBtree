package keys

import (
	"fmt"
)

const (
	prefixQuiz = "quiz"
)

// InProgressKey (SetNX 사용)
// Value: Hash{ quizID, timestamp(퀴즈 시작 시간)}
// 퀴즈 시작 요청시 진행중인지 확인
// 정답 제출시 timestamp 랑 제출 시간 확인
func InProgressKey(userEmail string) string {
	return fmt.Sprintf("%s:in_progress:%s", prefixQuiz, userEmail)
}
