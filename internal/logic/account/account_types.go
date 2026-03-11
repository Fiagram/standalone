package logic_account

type Role uint8

const (
	None Role = iota
	Admin
	Member
)

type AccountInfo struct {
	Username    string
	Fullname    string
	Email       string
	PhoneNumber string
	Role        Role
}

type CreateAccountParams struct {
	AccountInfo AccountInfo
	Password    string
}

type CreateAccountOutput struct {
	AccountId uint64
}

type DeleteAccountParams struct {
	AccountId uint64
}

type DeleteAccountByUsernameParams struct {
	Username string
}

type CheckAccountValidParams struct {
	Username string
	Password string
}

type CheckAccountValidOutput struct {
	AccountId uint64
}

type IsUsernameTakenParams struct {
	Username string
}

type IsUsernameTakenOutput struct {
	IsTaken bool
}

type GetAccountParams struct {
	AccountId uint64
}

type GetAccountOutput struct {
	AccountId   uint64
	AccountInfo AccountInfo
}

type GetAccountAllParams struct{}

type GetAccountAllOutput struct {
	AccountIds   []uint64
	AccountInfos []AccountInfo
}

type GetAccountListParams struct {
	AccountIds []uint64
}

type GetAccountListOutput struct {
	AccountIds   []uint64
	AccountInfos []AccountInfo
}

type UpdateAccountInfoParams struct {
	AccountId          uint64
	UpdatedAccountInfo AccountInfo
}

type UpdateAccountInfoOutput struct {
	AccountId uint64
}

type UpdateAccountPasswordParams struct {
	AccountId uint64
	Password  string
}

type UpdateAccountPasswordOutput struct {
	AccountId uint64
}
