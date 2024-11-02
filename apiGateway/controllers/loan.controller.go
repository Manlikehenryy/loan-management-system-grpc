package controllers

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/manlikehenryy/loan-management-system-grpc/apiGateway/configs"
	"github.com/manlikehenryy/loan-management-system-grpc/apiGateway/dto"
	"github.com/manlikehenryy/loan-management-system-grpc/apiGateway/grpcclient"
	"github.com/manlikehenryy/loan-management-system-grpc/apiGateway/helpers"
	loanPb "github.com/manlikehenryy/loan-management-system-grpc/apiGateway/loan"
)

func ApplyLoan(c *gin.Context) {
	userId:= c.MustGet("userId").(string)
	
	var applyLoanDto dto.ApplyLoanDto

	if err := c.ShouldBindJSON(&applyLoanDto); err != nil {
		log.Println("Unable to parse body:", err)
		helpers.SendError(c, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Initialize the gRPC client
    loanServiceClient, cleanup, err := grpcclient.NewLoanServiceClient(c)
    if err != nil {
        log.Fatalf("Failed to connect to LoanService: %v", err)
        helpers.SendError(c, http.StatusInternalServerError, "Internal service error")
        return
    }
    defer cleanup()

    // Set up the context with authorization metadata
    ctx := grpcclient.NewAuthContext(c, configs.Env.TOKEN)

	applyLoanReq := &loanPb.ApplyLoanRequest{
		UserId: userId,
		Amount: applyLoanDto.Amount,
		Duration: applyLoanDto.Duration,
	}

	applyLoanResp, err_ := loanServiceClient.ApplyLoan(ctx, applyLoanReq)

	if applyLoanResp == nil{
        helpers.SendError(c, http.StatusInternalServerError, "Unexpected service response")
        return
    }

	if err_ != nil || !applyLoanResp.Status {
		helpers.SendError(c, int(applyLoanResp.StatusCode), applyLoanResp.Message)
		return
	}


	helpers.SendJSON(c, http.StatusCreated, gin.H{
		"message": applyLoanResp.Message,
		"data": gin.H{
          "loanId": applyLoanResp.LoandId,
		},
	})
}

func ApproveLoan(c *gin.Context) {
	userId:= c.MustGet("userId").(string)
	
	var approveLoanDto dto.ApproveLoanDto

	if err := c.ShouldBindJSON(&approveLoanDto); err != nil {
		log.Println("Unable to parse body:", err)
		helpers.SendError(c, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Initialize the gRPC client
    loanServiceClient, cleanup, err := grpcclient.NewLoanServiceClient(c)
    if err != nil {
        log.Fatalf("Failed to connect to LoanService: %v", err)
        helpers.SendError(c, http.StatusInternalServerError, "Internal service error")
        return
    }
    defer cleanup()

    // Set up the context with authorization metadata
    ctx := grpcclient.NewAuthContext(c, configs.Env.TOKEN)

	approveLoanReq := &loanPb.ApproveLoanRequest{
		UserId: userId,
		LoanId: approveLoanDto.LoanId,
		ApprovedAmount: approveLoanDto.ApprovedAmount,
		Tenure: approveLoanDto.Tenure,
		MonthlyRepayment: approveLoanDto.MonthlyRepayment,
		EffectiveDate: approveLoanDto.EffectiveDate,
		ExpiryDate: approveLoanDto.ExpiryDate,
	}

	approveLoanResp, err_ := loanServiceClient.ApproveLoan(ctx, approveLoanReq)

	if approveLoanResp == nil{
        helpers.SendError(c, http.StatusInternalServerError, "Unexpected service response")
        return
    }
	
	if err_ != nil || !approveLoanResp.Status {
		helpers.SendError(c, int(approveLoanResp.StatusCode), approveLoanResp.Message)
		return
	}


	helpers.SendJSON(c, http.StatusCreated, gin.H{
		"message": approveLoanResp.Message,
	})
}

func RejectLoan(c *gin.Context) {
	userId:= c.MustGet("userId").(string)
	
	var rejectLoanDto dto.RejectLoanDto

	if err := c.ShouldBindJSON(&rejectLoanDto); err != nil {
		log.Println("Unable to parse body:", err)
		helpers.SendError(c, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Initialize the gRPC client
    loanServiceClient, cleanup, err := grpcclient.NewLoanServiceClient(c)
    if err != nil {
        log.Fatalf("Failed to connect to LoanService: %v", err)
        helpers.SendError(c, http.StatusInternalServerError, "Internal service error")
        return
    }
    defer cleanup()

    // Set up the context with authorization metadata
    ctx := grpcclient.NewAuthContext(c, configs.Env.TOKEN)

	rejectLoanReq := &loanPb.RejectLoanRequest{
		UserId: userId,
		LoanId: rejectLoanDto.LoanId,
	}

	rejectLoanResp, err_ := loanServiceClient.RejectLoan(ctx, rejectLoanReq)

	if rejectLoanResp == nil{
        helpers.SendError(c, http.StatusInternalServerError, "Unexpected service response")
        return
    }

	if err_ != nil || !rejectLoanResp.Status {
		helpers.SendError(c, int(rejectLoanResp.StatusCode), rejectLoanResp.Message)
		return
	}


	helpers.SendJSON(c, http.StatusCreated, gin.H{
		"message": rejectLoanResp.Message,
	})
}