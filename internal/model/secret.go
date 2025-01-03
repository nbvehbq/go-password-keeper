package model

type ResourceType uint8

const (
	LoginPasswordType ResourceType = iota + 1
	TextType
	BinaryType
	BankCardType
)

type BankCard struct {
	Number   string `json:"number"`
	ExpireAt string `json:"expireAt"`
	Name     string `json:"name,omitempty"`
	Surname  string `json:"surname,omitempty"`
}

type LoginPassword struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type Text struct {
	Value string `json:"value"`
}

type Binary struct {
	Name  string `json:"name"`
	Value []byte `json:"value"`
}

type Secret struct {
	ID      int64        `db:"id" json:"id"`
	Name    string       `db:"name" json:"name"`
	UserID  int64        `db:"user_id" json:"user_id"`
	Type    ResourceType `db:"type" json:"type"`
	Payload []byte       `db:"payload" json:"payload"`
	Meta    []byte       `db:"meta" json:"meta"`
}

func ValidateParam(type_ string) (ResourceType, bool) {
	switch type_ {
	case "1":
		return LoginPasswordType, true
	case "2":
		return TextType, true
	case "3":
		return BinaryType, true
	case "4":
		return BankCardType, true
	default:
		return 0, false
	}
}
