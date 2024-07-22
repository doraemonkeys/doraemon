package doraemon

type Ternary int8

const (
	Unknown Ternary = -1
	True    Ternary = 1
	False   Ternary = 0
)

func (t Ternary) String() string {
	switch t {
	case True:
		return "True"
	case False:
		return "False"
	default:
		return "Unknown"
	}
}

func (t Ternary) IsTrue() bool {
	return t == True
}

func (t Ternary) IsFalse() bool {
	return t == False
}

func (t Ternary) IsUnknown() bool {
	return t == Unknown
}
