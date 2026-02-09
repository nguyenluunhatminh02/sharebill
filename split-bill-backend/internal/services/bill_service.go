package services

import (
	"context"
	"errors"
	"math"
	"time"

	"github.com/splitbill/backend/internal/models"
	"github.com/splitbill/backend/internal/repository"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type BillService struct {
	billRepo  *repository.BillRepository
	groupRepo *repository.GroupRepository
	userRepo  *repository.UserRepository
}

func NewBillService(billRepo *repository.BillRepository, groupRepo *repository.GroupRepository, userRepo *repository.UserRepository) *BillService {
	return &BillService{
		billRepo:  billRepo,
		groupRepo: groupRepo,
		userRepo:  userRepo,
	}
}

// CreateBill creates a new bill in a group
func (s *BillService) CreateBill(ctx context.Context, groupID string, firebaseUID string, req models.CreateBillRequest) (*models.Bill, error) {
	groupObjID, err := primitive.ObjectIDFromHex(groupID)
	if err != nil {
		return nil, errors.New("invalid group ID")
	}

	user, err := s.userRepo.FindByFirebaseUID(ctx, firebaseUID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	// Verify user is member of the group
	isMember, err := s.groupRepo.IsMember(ctx, groupObjID, user.ID)
	if err != nil || !isMember {
		return nil, errors.New("you are not a member of this group")
	}

	paidByID, err := primitive.ObjectIDFromHex(req.PaidBy)
	if err != nil {
		return nil, errors.New("invalid paid_by user ID")
	}

	// Build bill items
	items := make([]models.BillItem, len(req.Items))
	for i, item := range req.Items {
		assignedTo := make([]primitive.ObjectID, len(item.AssignedTo))
		for j, uid := range item.AssignedTo {
			id, err := primitive.ObjectIDFromHex(uid)
			if err != nil {
				return nil, errors.New("invalid user ID in assigned_to")
			}
			assignedTo[j] = id
		}

		totalPrice := item.TotalPrice
		if totalPrice == 0 {
			totalPrice = float64(item.Quantity) * item.UnitPrice
		}

		items[i] = models.BillItem{
			ID:         primitive.NewObjectID(),
			Name:       item.Name,
			Quantity:   item.Quantity,
			UnitPrice:  item.UnitPrice,
			TotalPrice: totalPrice,
			AssignedTo: assignedTo,
		}
	}

	bill := &models.Bill{
		GroupID:         groupObjID,
		Title:           req.Title,
		Description:     req.Description,
		ReceiptImageURL: req.ReceiptImageURL,
		TotalAmount:     req.TotalAmount,
		Currency:        req.Currency,
		PaidBy:          paidByID,
		SplitType:       req.SplitType,
		Items:           items,
		ExtraCharges:    req.ExtraCharges,
		Status:          models.BillPending,
	}

	// Calculate splits based on split type
	splits, err := s.calculateSplits(ctx, bill, req.SplitAmong)
	if err != nil {
		return nil, err
	}
	bill.Splits = splits

	if err := s.billRepo.Create(ctx, bill); err != nil {
		return nil, err
	}

	return bill, nil
}

// calculateSplits calculates how much each person owes
func (s *BillService) calculateSplits(ctx context.Context, bill *models.Bill, splitAmong []string) ([]models.BillSplit, error) {
	switch bill.SplitType {
	case models.SplitEqual:
		return s.calculateEqualSplit(ctx, bill, splitAmong)
	case models.SplitByItem:
		return s.calculateByItemSplit(bill)
	case models.SplitByAmount:
		// Custom amounts are already set in the request
		return nil, errors.New("by_amount split should provide splits directly")
	default:
		return s.calculateEqualSplit(ctx, bill, splitAmong)
	}
}

// calculateEqualSplit splits the total equally among specified users
func (s *BillService) calculateEqualSplit(ctx context.Context, bill *models.Bill, splitAmong []string) ([]models.BillSplit, error) {
	if len(splitAmong) == 0 {
		// If no users specified, get all group members
		group, err := s.groupRepo.FindByID(ctx, bill.GroupID)
		if err != nil {
			return nil, err
		}
		for _, m := range group.Members {
			splitAmong = append(splitAmong, m.UserID.Hex())
		}
	}

	// Calculate total including extra charges
	total := bill.TotalAmount + bill.ExtraCharges.Tax + bill.ExtraCharges.ServiceCharge + bill.ExtraCharges.Tip - bill.ExtraCharges.Discount

	perPerson := roundToTwo(total / float64(len(splitAmong)))

	splits := make([]models.BillSplit, len(splitAmong))
	for i, uid := range splitAmong {
		userID, err := primitive.ObjectIDFromHex(uid)
		if err != nil {
			return nil, errors.New("invalid user ID in split_among")
		}

		amount := perPerson
		// The payer's split represents what they owe themselves (they already paid)
		isPaid := userID == bill.PaidBy

		splits[i] = models.BillSplit{
			UserID: userID,
			Amount: amount,
			IsPaid: isPaid,
		}

		if isPaid {
			now := time.Now()
			splits[i].PaidAt = &now
		}
	}

	return splits, nil
}

// calculateByItemSplit splits based on item assignments
func (s *BillService) calculateByItemSplit(bill *models.Bill) ([]models.BillSplit, error) {
	if len(bill.Items) == 0 {
		return nil, errors.New("items are required for by_item split")
	}

	// Calculate each person's total from assigned items
	userTotals := make(map[primitive.ObjectID]float64)

	for _, item := range bill.Items {
		if len(item.AssignedTo) == 0 {
			continue
		}
		perPerson := roundToTwo(item.TotalPrice / float64(len(item.AssignedTo)))
		for _, uid := range item.AssignedTo {
			userTotals[uid] += perPerson
		}
	}

	// Distribute extra charges proportionally
	itemsTotal := 0.0
	for _, t := range userTotals {
		itemsTotal += t
	}

	extraTotal := bill.ExtraCharges.Tax + bill.ExtraCharges.ServiceCharge + bill.ExtraCharges.Tip - bill.ExtraCharges.Discount

	splits := make([]models.BillSplit, 0, len(userTotals))
	for uid, amount := range userTotals {
		// Add proportional extra charges
		if itemsTotal > 0 && extraTotal != 0 {
			proportion := amount / itemsTotal
			amount += roundToTwo(extraTotal * proportion)
		}

		isPaid := uid == bill.PaidBy
		split := models.BillSplit{
			UserID: uid,
			Amount: roundToTwo(amount),
			IsPaid: isPaid,
		}
		if isPaid {
			now := time.Now()
			split.PaidAt = &now
		}
		splits = append(splits, split)
	}

	return splits, nil
}

// GetBill gets a bill by ID
func (s *BillService) GetBill(ctx context.Context, billID string) (*models.Bill, error) {
	objID, err := primitive.ObjectIDFromHex(billID)
	if err != nil {
		return nil, errors.New("invalid bill ID")
	}
	return s.billRepo.FindByID(ctx, objID)
}

// ListBills lists all bills in a group
func (s *BillService) ListBills(ctx context.Context, groupID string) ([]models.Bill, error) {
	objID, err := primitive.ObjectIDFromHex(groupID)
	if err != nil {
		return nil, errors.New("invalid group ID")
	}
	return s.billRepo.FindByGroupID(ctx, objID)
}

// UpdateBill updates a bill
func (s *BillService) UpdateBill(ctx context.Context, billID string, req models.UpdateBillRequest) (*models.Bill, error) {
	bill, err := s.GetBill(ctx, billID)
	if err != nil {
		return nil, err
	}

	if req.Title != "" {
		bill.Title = req.Title
	}
	if req.Description != "" {
		bill.Description = req.Description
	}
	if req.TotalAmount != nil {
		bill.TotalAmount = *req.TotalAmount
	}
	if req.Status != "" {
		bill.Status = req.Status
	}

	if err := s.billRepo.Update(ctx, bill); err != nil {
		return nil, err
	}

	return bill, nil
}

// DeleteBill soft-deletes a bill
func (s *BillService) DeleteBill(ctx context.Context, billID string) error {
	objID, err := primitive.ObjectIDFromHex(billID)
	if err != nil {
		return errors.New("invalid bill ID")
	}
	return s.billRepo.Delete(ctx, objID)
}

// roundToTwo rounds a float to 2 decimal places
func roundToTwo(val float64) float64 {
	return math.Round(val*100) / 100
}
