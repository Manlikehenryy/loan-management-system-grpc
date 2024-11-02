package dto

type ApplyLoanDto struct {
	Amount   float32 `json:"amount"`
	Duration int32   `json:"duration"`
}

type ApproveLoanDto struct {
	LoanId           string  `json:"loanId"`
	ApprovedAmount   float32 `json:"approvedAmount"`
	Tenure           int32   `json:"tenure"`
	MonthlyRepayment float32 `json:"monthlyRepayment"`
	EffectiveDate    string  `json:"effectiveDate"`
	ExpiryDate       string  `json:"expiryDate"`
}

type RejectLoanDto struct {
	LoanId string `json:"loanId"`
}
