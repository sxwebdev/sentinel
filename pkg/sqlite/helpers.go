package sqlite

func GetDSN(dbPath string) string {
	dsn := "file:" + dbPath +
		"?_busy_timeout=30000" +
		"&_pragma=journal_mode(WAL)" +
		"&_pragma=synchronous(NORMAL)" +
		"&_pragma=foreign_keys(ON)" +
		"&_pragma=cache_size(10000)"
	return dsn
}
