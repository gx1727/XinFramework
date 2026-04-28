package flag

// nullStr 将空字符串转换为 nil，用于数据库 NULL 值处理
func nullStr(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}

// nilIfZero 将零值转换为 nil，用于数据库 NULL 值处理
func nilIfZero(v uint) interface{} {
	if v == 0 {
		return nil
	}
	return v
}
