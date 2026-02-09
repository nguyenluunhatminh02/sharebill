package services

import (
	"context"
	"math"
	"sort"

	"github.com/splitbill/backend/internal/models"
	"github.com/splitbill/backend/internal/repository"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type DebtService struct {
	billRepo        *repository.BillRepository
	transactionRepo *repository.TransactionRepository
	userRepo        *repository.UserRepository
}

func NewDebtService(
	billRepo *repository.BillRepository,
	transactionRepo *repository.TransactionRepository,
	userRepo *repository.UserRepository,
) *DebtService {
	return &DebtService{
		billRepo:        billRepo,
		transactionRepo: transactionRepo,
		userRepo:        userRepo,
	}
}

// GetGroupBalances calculates the net balance for each member in a group
func (s *DebtService) GetGroupBalances(ctx context.Context, groupID string) ([]models.BalanceResponse, error) {
	objID, err := primitive.ObjectIDFromHex(groupID)
	if err != nil {
		return nil, err
	}

	// Get all active bills in the group
	bills, err := s.billRepo.FindActiveByGroupID(ctx, objID)
	if err != nil {
		return nil, err
	}

	// Get all confirmed transactions (settlements)
	transactions, err := s.transactionRepo.FindConfirmedByGroupID(ctx, objID)
	if err != nil {
		return nil, err
	}

	// Calculate net balances
	// Positive balance = person is owed money (paid more than their share)
	// Negative balance = person owes money (paid less than their share)
	balances := make(map[string]float64)

	for _, bill := range bills {
		if bill.Status == models.BillCancelled {
			continue
		}

		paidByID := bill.PaidBy.Hex()

		for _, split := range bill.Splits {
			splitUserID := split.UserID.Hex()

			if splitUserID == paidByID {
				// The payer is owed (total - their share) by others
				// Their balance increases by what others owe them
				balances[paidByID] += (bill.TotalAmount + bill.ExtraCharges.Tax +
					bill.ExtraCharges.ServiceCharge + bill.ExtraCharges.Tip -
					bill.ExtraCharges.Discount) - split.Amount
			} else if !split.IsPaid {
				// Non-payer owes their split amount
				balances[splitUserID] -= split.Amount
			}
		}
	}

	// Factor in confirmed settlements
	for _, tx := range transactions {
		fromID := tx.FromUser.Hex()
		toID := tx.ToUser.Hex()

		balances[fromID] += tx.Amount // Payer reduces debt
		balances[toID] -= tx.Amount   // Receiver's credit decreases
	}

	// Build response with user details
	var result []models.BalanceResponse
	for userIDStr, balance := range balances {
		if math.Abs(balance) < 0.01 {
			balance = 0 // Treat tiny amounts as zero
		}

		userObjID, err := primitive.ObjectIDFromHex(userIDStr)
		if err != nil {
			continue
		}

		user, err := s.userRepo.FindByID(ctx, userObjID)
		displayName := userIDStr
		if err == nil {
			displayName = user.DisplayName
		}

		result = append(result, models.BalanceResponse{
			UserID:      userIDStr,
			DisplayName: displayName,
			Balance:     math.Round(balance*100) / 100,
		})
	}

	return result, nil
}

// GetOptimalSettlements returns the minimum number of transactions to settle all debts
func (s *DebtService) GetOptimalSettlements(ctx context.Context, groupID string) ([]models.Settlement, error) {
	balances, err := s.GetGroupBalances(ctx, groupID)
	if err != nil {
		return nil, err
	}

	// Build net amounts map
	netAmounts := make(map[string]float64)
	nameMap := make(map[string]string)

	for _, b := range balances {
		if math.Abs(b.Balance) >= 0.01 {
			netAmounts[b.UserID] = b.Balance
			nameMap[b.UserID] = b.DisplayName
		}
	}

	return optimizeSettlements(netAmounts, nameMap), nil
}

// optimizeSettlements implements the greedy min-cash flow algorithm
// to find the minimum number of transactions to settle all debts
func optimizeSettlements(netAmounts map[string]float64, nameMap map[string]string) []models.Settlement {
	type userBalance struct {
		userID  string
		balance float64
	}

	// Separate into debtors (negative balance = owes money) and creditors (positive balance = owed money)
	var debtors []userBalance  // People who owe money
	var creditors []userBalance // People who are owed money

	for uid, amount := range netAmounts {
		if amount < -0.01 {
			debtors = append(debtors, userBalance{uid, amount})
		} else if amount > 0.01 {
			creditors = append(creditors, userBalance{uid, amount})
		}
	}

	// Sort: largest debtor first, largest creditor first
	sort.Slice(debtors, func(i, j int) bool {
		return debtors[i].balance < debtors[j].balance // Most negative first
	})
	sort.Slice(creditors, func(i, j int) bool {
		return creditors[i].balance > creditors[j].balance // Most positive first
	})

	var settlements []models.Settlement

	i, j := 0, 0
	for i < len(debtors) && j < len(creditors) {
		debtAmount := math.Abs(debtors[i].balance)
		creditAmount := creditors[j].balance

		// Transfer the smaller of the two amounts
		transferAmount := math.Min(debtAmount, creditAmount)
		transferAmount = math.Round(transferAmount*100) / 100

		if transferAmount >= 0.01 {
			settlements = append(settlements, models.Settlement{
				FromUserID:   debtors[i].userID,
				FromUserName: nameMap[debtors[i].userID],
				ToUserID:     creditors[j].userID,
				ToUserName:   nameMap[creditors[j].userID],
				Amount:       transferAmount,
			})
		}

		debtors[i].balance += transferAmount
		creditors[j].balance -= transferAmount

		// Move to next if settled
		if math.Abs(debtors[i].balance) < 0.01 {
			i++
		}
		if creditors[j].balance < 0.01 {
			j++
		}
	}

	return settlements
}
