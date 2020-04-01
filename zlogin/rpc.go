package zlogin

type REQVerifySession struct {
	Account string
	Session string
}
type RESVerifySession struct {
	ErrorCode int
}
