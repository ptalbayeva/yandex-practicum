package model

// generate:reset
type ResetableStruct struct {
	I     int
	STR   string
	STRP  *string
	S     []int
	M     map[string]string
	CHILD *ResetableStruct
}

func (rs *ResetableStruct) Reset() {
	if rs == nil {
		return
	}

	rs.I = 0
	rs.STR = ""
	if rs.STRP != nil {
		*rs.STRP = ""
	}
	rs.S = rs.S[:0]
	clear(rs.M)

	if rs.CHILD != nil {
		var anyChild any = rs.CHILD
		if resetter, ok := anyChild.(interface{ Reset() }); ok {
			resetter.Reset()
		}
	}
}
