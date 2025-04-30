package context

type contextKey string

const UserKey contextKey = "user"
const TokenKey contextKey = "token"
const RequestIDKey contextKey = "requestID"
