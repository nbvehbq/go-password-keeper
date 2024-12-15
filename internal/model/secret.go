package model

type ResourceType uint8

const (
	LoginPassword ResourceType = iota + 1
	File
	BankCard
)

type Secret struct {
	ID      int64        `db:"id" json:"id"`
	UserID  int64        `db:"user_id" json:"user_id"`
	Type    ResourceType `db:"type" json:"type"`
	Payload []byte       `db:"payload" json:"payload"`
	Meta    []byte       `db:"meta" json:"meta"`
}

func ValidateParam(type_ string) (ResourceType, bool) {
	switch type_ {
	case "1":
		return LoginPassword, true
	case "2":
		return File, true
	case "3":
		return BankCard, true
	default:
		return 0, false
	}
}
