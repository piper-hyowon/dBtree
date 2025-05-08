package redis

import (
	"fmt"
)

const (
	prefixQuiz = "quiz"
)

// InProgressQuiz (SetNX 사용)
// Value: Hash{ quizID, timestamp(퀴즈 시작 시간)}
// 퀴즈 시작 요청시 진행중인지 확인
// 정답 제출시 timestamp 랑 제출 시간 확인
func InProgressQuiz(userEmail string) string {
	return fmt.Sprintf("%s:in_progress:%s", prefixQuiz, userEmail)
}

// UserPass 해당 positionID 레몬 퀴즈 통과 확인용
// Value: 패스한 시각 timestamp
// 레몬 수확시 패스한 시각에서 5초 이내인지 확인
func UserPass(userEmail string, positionID int) string {
	return fmt.Sprintf("%s:passed:%d:%s", prefixQuiz, positionID, userEmail)
}
