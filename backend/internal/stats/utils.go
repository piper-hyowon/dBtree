package stats

import stringutil "github.com/piper-hyowon/dBtree/internal/utils/string"

const (
	EmailMaskKeepChars = 3 // 앞에서 보여줄 글자 수
	EmailMaskString    = "***"
)

func maskEmailForLeaderboard(email string) string {
	return stringutil.MaskEmail(email, EmailMaskKeepChars, EmailMaskString)
}
