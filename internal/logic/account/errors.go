package logic_account

import "fmt"

var (
	ErrTxCommitFailed = fmt.Errorf("failed to commit transaction")
	ErrTxBeginFailed  = fmt.Errorf("failed to begin transaction")
)
