package services

import (
	"context"
	"sort"
	"time"

	"github.com/splitbill/backend/internal/models"
	"github.com/splitbill/backend/internal/repository"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type StatsService struct {
	billRepo        *repository.BillRepository
	transactionRepo *repository.TransactionRepository
	groupRepo       *repository.GroupRepository
	userRepo        *repository.UserRepository
}

func NewStatsService(
	billRepo *repository.BillRepository,
	transactionRepo *repository.TransactionRepository,
	groupRepo *repository.GroupRepository,
	userRepo *repository.UserRepository,
) *StatsService {
	return &StatsService{
		billRepo:        billRepo,
		transactionRepo: transactionRepo,
		groupRepo:       groupRepo,
		userRepo:        userRepo,
	}
}

// GroupStats represents statistics for a group
type GroupStats struct {
	GroupID         string              `json:"group_id"`
	GroupName       string              `json:"group_name"`
	TotalSpent      float64             `json:"total_spent"`
	TotalBills      int                 `json:"total_bills"`
	TotalMembers    int                 `json:"total_members"`
	AverageBill     float64             `json:"average_bill"`
	LargestBill     *BillSummary        `json:"largest_bill,omitempty"`
	SmallestBill    *BillSummary        `json:"smallest_bill,omitempty"`
	MemberStats     []MemberSpendStats  `json:"member_stats"`
	CategoryStats   []CategoryStat      `json:"category_stats"`
	MonthlyTrend    []MonthlySpend      `json:"monthly_trend"`
	RecentBills     []BillSummary       `json:"recent_bills"`
}

// BillSummary is a lightweight bill info for stats
type BillSummary struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Amount      float64   `json:"amount"`
	Category    string    `json:"category"`
	PaidByName  string    `json:"paid_by_name"`
	CreatedAt   time.Time `json:"created_at"`
}

// MemberSpendStats tracks how much each member spent and owes
type MemberSpendStats struct {
	UserID      string  `json:"user_id"`
	DisplayName string  `json:"display_name"`
	AvatarURL   string  `json:"avatar_url"`
	TotalPaid   float64 `json:"total_paid"`
	TotalOwed   float64 `json:"total_owed"`
	NetBalance  float64 `json:"net_balance"`
	BillCount   int     `json:"bill_count"`
	Percentage  float64 `json:"percentage"`
}

// CategoryStat tracks spending by category
type CategoryStat struct {
	Category   string  `json:"category"`
	Total      float64 `json:"total"`
	Count      int     `json:"count"`
	Percentage float64 `json:"percentage"`
	Icon       string  `json:"icon"`
	Color      string  `json:"color"`
}

// MonthlySpend tracks spending per month
type MonthlySpend struct {
	Month     string  `json:"month"`
	Year      int     `json:"year"`
	MonthNum  int     `json:"month_num"`
	Total     float64 `json:"total"`
	BillCount int     `json:"bill_count"`
}

// UserOverallStats represents user-level statistics across all groups
type UserOverallStats struct {
	TotalGroups     int              `json:"total_groups"`
	TotalSpent      float64          `json:"total_spent"`
	TotalOwed       float64          `json:"total_owed"`
	TotalBills      int              `json:"total_bills"`
	TopGroups       []GroupSpendInfo `json:"top_groups"`
	CategoryStats   []CategoryStat   `json:"category_stats"`
	MonthlyTrend    []MonthlySpend   `json:"monthly_trend"`
}

// GroupSpendInfo is a lightweight group spend summary
type GroupSpendInfo struct {
	GroupID   string  `json:"group_id"`
	GroupName string  `json:"group_name"`
	Total     float64 `json:"total"`
	BillCount int     `json:"bill_count"`
}

var categoryMeta = map[string]struct {
	Icon  string
	Color string
}{
	"food":          {Icon: "restaurant", Color: "#FF6B6B"},
	"drinks":        {Icon: "beer", Color: "#FFA502"},
	"groceries":     {Icon: "cart", Color: "#2ED573"},
	"transport":     {Icon: "car", Color: "#1E90FF"},
	"accommodation": {Icon: "bed", Color: "#A29BFE"},
	"entertainment": {Icon: "game-controller", Color: "#FD79A8"},
	"shopping":      {Icon: "bag-handle", Color: "#E17055"},
	"utilities":     {Icon: "flash", Color: "#FDCB6E"},
	"health":        {Icon: "medkit", Color: "#00B894"},
	"travel":        {Icon: "airplane", Color: "#74B9FF"},
	"other":         {Icon: "ellipsis-horizontal", Color: "#636E72"},
}

var monthNames = []string{
	"", "Tháng 1", "Tháng 2", "Tháng 3", "Tháng 4", "Tháng 5", "Tháng 6",
	"Tháng 7", "Tháng 8", "Tháng 9", "Tháng 10", "Tháng 11", "Tháng 12",
}

// GetGroupStats computes statistics for a group
func (s *StatsService) GetGroupStats(ctx context.Context, groupID string) (*GroupStats, error) {
	objID, err := primitive.ObjectIDFromHex(groupID)
	if err != nil {
		return nil, err
	}

	// Get group info
	group, err := s.groupRepo.FindByID(ctx, objID)
	if err != nil {
		return nil, err
	}

	// Get all active bills
	bills, err := s.billRepo.FindActiveByGroupID(ctx, objID)
	if err != nil {
		return nil, err
	}

	// Build user name cache
	userNames := make(map[string]string)
	userAvatars := make(map[string]string)
	for _, m := range group.Members {
		user, err := s.userRepo.FindByID(ctx, m.UserID)
		if err == nil {
			userNames[m.UserID.Hex()] = user.DisplayName
			userAvatars[m.UserID.Hex()] = user.AvatarURL
		}
	}

	stats := &GroupStats{
		GroupID:      groupID,
		GroupName:    group.Name,
		TotalBills:   len(bills),
		TotalMembers: len(group.Members),
	}

	if len(bills) == 0 {
		stats.MemberStats = []MemberSpendStats{}
		stats.CategoryStats = []CategoryStat{}
		stats.MonthlyTrend = []MonthlySpend{}
		stats.RecentBills = []BillSummary{}
		return stats, nil
	}

	// Calculate totals
	memberPaid := make(map[string]float64)
	memberOwed := make(map[string]float64)
	memberBillCount := make(map[string]int)
	categoryTotals := make(map[string]float64)
	categoryCounts := make(map[string]int)
	monthlyMap := make(map[string]*MonthlySpend)

	var totalSpent float64
	var largestBill, smallestBill *models.Bill

	for i := range bills {
		bill := &bills[i]
		totalSpent += bill.TotalAmount

		// Track largest/smallest
		if largestBill == nil || bill.TotalAmount > largestBill.TotalAmount {
			largestBill = bill
		}
		if smallestBill == nil || bill.TotalAmount < smallestBill.TotalAmount {
			smallestBill = bill
		}

		// Track who paid
		paidByID := bill.PaidBy.Hex()
		memberPaid[paidByID] += bill.TotalAmount
		memberBillCount[paidByID]++

		// Track splits (who owes)
		for _, split := range bill.Splits {
			memberOwed[split.UserID.Hex()] += split.Amount
		}

		// Track categories
		cat := bill.Category
		if cat == "" {
			cat = "other"
		}
		categoryTotals[cat] += bill.TotalAmount
		categoryCounts[cat]++

		// Track monthly
		monthKey := bill.CreatedAt.Format("2006-01")
		if _, ok := monthlyMap[monthKey]; !ok {
			monthlyMap[monthKey] = &MonthlySpend{
				Month:    monthNames[int(bill.CreatedAt.Month())],
				Year:     bill.CreatedAt.Year(),
				MonthNum: int(bill.CreatedAt.Month()),
			}
		}
		monthlyMap[monthKey].Total += bill.TotalAmount
		monthlyMap[monthKey].BillCount++
	}

	stats.TotalSpent = totalSpent
	if len(bills) > 0 {
		stats.AverageBill = totalSpent / float64(len(bills))
	}

	// Largest/smallest bill
	if largestBill != nil {
		stats.LargestBill = &BillSummary{
			ID:         largestBill.ID.Hex(),
			Title:      largestBill.Title,
			Amount:     largestBill.TotalAmount,
			Category:   largestBill.Category,
			PaidByName: userNames[largestBill.PaidBy.Hex()],
			CreatedAt:  largestBill.CreatedAt,
		}
	}
	if smallestBill != nil {
		stats.SmallestBill = &BillSummary{
			ID:         smallestBill.ID.Hex(),
			Title:      smallestBill.Title,
			Amount:     smallestBill.TotalAmount,
			Category:   smallestBill.Category,
			PaidByName: userNames[smallestBill.PaidBy.Hex()],
			CreatedAt:  smallestBill.CreatedAt,
		}
	}

	// Member stats
	allMembers := make(map[string]bool)
	for k := range memberPaid {
		allMembers[k] = true
	}
	for k := range memberOwed {
		allMembers[k] = true
	}

	memberStats := make([]MemberSpendStats, 0)
	for uid := range allMembers {
		paid := memberPaid[uid]
		owed := memberOwed[uid]
		pct := 0.0
		if totalSpent > 0 {
			pct = (paid / totalSpent) * 100
		}
		memberStats = append(memberStats, MemberSpendStats{
			UserID:      uid,
			DisplayName: userNames[uid],
			AvatarURL:   userAvatars[uid],
			TotalPaid:   paid,
			TotalOwed:   owed,
			NetBalance:  paid - owed,
			BillCount:   memberBillCount[uid],
			Percentage:  pct,
		})
	}
	sort.Slice(memberStats, func(i, j int) bool {
		return memberStats[i].TotalPaid > memberStats[j].TotalPaid
	})
	stats.MemberStats = memberStats

	// Category stats
	catStats := make([]CategoryStat, 0)
	for cat, total := range categoryTotals {
		meta := categoryMeta[cat]
		if meta.Icon == "" {
			meta = categoryMeta["other"]
		}
		pct := 0.0
		if totalSpent > 0 {
			pct = (total / totalSpent) * 100
		}
		catStats = append(catStats, CategoryStat{
			Category:   cat,
			Total:      total,
			Count:      categoryCounts[cat],
			Percentage: pct,
			Icon:       meta.Icon,
			Color:      meta.Color,
		})
	}
	sort.Slice(catStats, func(i, j int) bool {
		return catStats[i].Total > catStats[j].Total
	})
	stats.CategoryStats = catStats

	// Monthly trend (last 6 months)
	monthlyTrend := make([]MonthlySpend, 0)
	for _, ms := range monthlyMap {
		monthlyTrend = append(monthlyTrend, *ms)
	}
	sort.Slice(monthlyTrend, func(i, j int) bool {
		if monthlyTrend[i].Year != monthlyTrend[j].Year {
			return monthlyTrend[i].Year < monthlyTrend[j].Year
		}
		return monthlyTrend[i].MonthNum < monthlyTrend[j].MonthNum
	})
	if len(monthlyTrend) > 6 {
		monthlyTrend = monthlyTrend[len(monthlyTrend)-6:]
	}
	stats.MonthlyTrend = monthlyTrend

	// Recent bills (top 5)
	limit := 5
	if len(bills) < limit {
		limit = len(bills)
	}
	recentBills := make([]BillSummary, limit)
	for i := 0; i < limit; i++ {
		recentBills[i] = BillSummary{
			ID:         bills[i].ID.Hex(),
			Title:      bills[i].Title,
			Amount:     bills[i].TotalAmount,
			Category:   bills[i].Category,
			PaidByName: userNames[bills[i].PaidBy.Hex()],
			CreatedAt:  bills[i].CreatedAt,
		}
	}
	stats.RecentBills = recentBills

	return stats, nil
}

// GetUserStats computes statistics for a user across all groups
func (s *StatsService) GetUserStats(ctx context.Context, userID primitive.ObjectID) (*UserOverallStats, error) {
	groups, err := s.groupRepo.FindByMemberUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	stats := &UserOverallStats{
		TotalGroups: len(groups),
	}

	if len(groups) == 0 {
		stats.TopGroups = []GroupSpendInfo{}
		stats.CategoryStats = []CategoryStat{}
		stats.MonthlyTrend = []MonthlySpend{}
		return stats, nil
	}

	categoryTotals := make(map[string]float64)
	categoryCounts := make(map[string]int)
	monthlyMap := make(map[string]*MonthlySpend)
	groupSpends := make([]GroupSpendInfo, 0)

	userIDStr := userID.Hex()

	for _, group := range groups {
		bills, err := s.billRepo.FindActiveByGroupID(ctx, group.ID)
		if err != nil {
			continue
		}

		groupTotal := 0.0
		groupBillCount := 0

		for _, bill := range bills {
			// Check if user is involved
			isPaidBy := bill.PaidBy.Hex() == userIDStr
			userSplitAmount := 0.0
			for _, split := range bill.Splits {
				if split.UserID.Hex() == userIDStr {
					userSplitAmount = split.Amount
					break
				}
			}

			if !isPaidBy && userSplitAmount == 0 {
				continue
			}

			if isPaidBy {
				stats.TotalSpent += bill.TotalAmount
			}
			stats.TotalOwed += userSplitAmount
			stats.TotalBills++
			groupTotal += userSplitAmount
			groupBillCount++

			// Category tracking
			cat := bill.Category
			if cat == "" {
				cat = "other"
			}
			categoryTotals[cat] += userSplitAmount
			categoryCounts[cat]++

			// Monthly tracking
			monthKey := bill.CreatedAt.Format("2006-01")
			if _, ok := monthlyMap[monthKey]; !ok {
				monthlyMap[monthKey] = &MonthlySpend{
					Month:    monthNames[int(bill.CreatedAt.Month())],
					Year:     bill.CreatedAt.Year(),
					MonthNum: int(bill.CreatedAt.Month()),
				}
			}
			monthlyMap[monthKey].Total += userSplitAmount
			monthlyMap[monthKey].BillCount++
		}

		if groupBillCount > 0 {
			groupSpends = append(groupSpends, GroupSpendInfo{
				GroupID:   group.ID.Hex(),
				GroupName: group.Name,
				Total:     groupTotal,
				BillCount: groupBillCount,
			})
		}
	}

	// Sort top groups by spending
	sort.Slice(groupSpends, func(i, j int) bool {
		return groupSpends[i].Total > groupSpends[j].Total
	})
	if len(groupSpends) > 5 {
		groupSpends = groupSpends[:5]
	}
	stats.TopGroups = groupSpends

	// Category stats
	totalForPct := 0.0
	for _, v := range categoryTotals {
		totalForPct += v
	}
	catStats := make([]CategoryStat, 0)
	for cat, total := range categoryTotals {
		meta := categoryMeta[cat]
		if meta.Icon == "" {
			meta = categoryMeta["other"]
		}
		pct := 0.0
		if totalForPct > 0 {
			pct = (total / totalForPct) * 100
		}
		catStats = append(catStats, CategoryStat{
			Category:   cat,
			Total:      total,
			Count:      categoryCounts[cat],
			Percentage: pct,
			Icon:       meta.Icon,
			Color:      meta.Color,
		})
	}
	sort.Slice(catStats, func(i, j int) bool {
		return catStats[i].Total > catStats[j].Total
	})
	stats.CategoryStats = catStats

	// Monthly trend
	monthlyTrend := make([]MonthlySpend, 0)
	for _, ms := range monthlyMap {
		monthlyTrend = append(monthlyTrend, *ms)
	}
	sort.Slice(monthlyTrend, func(i, j int) bool {
		if monthlyTrend[i].Year != monthlyTrend[j].Year {
			return monthlyTrend[i].Year < monthlyTrend[j].Year
		}
		return monthlyTrend[i].MonthNum < monthlyTrend[j].MonthNum
	})
	if len(monthlyTrend) > 6 {
		monthlyTrend = monthlyTrend[len(monthlyTrend)-6:]
	}
	stats.MonthlyTrend = monthlyTrend

	return stats, nil
}

// ExportGroupSummary generates a text summary of a group
func (s *StatsService) ExportGroupSummary(ctx context.Context, groupID string) (string, error) {
	groupStats, err := s.GetGroupStats(ctx, groupID)
	if err != nil {
		return "", err
	}

	objID, _ := primitive.ObjectIDFromHex(groupID)
	settlements, err := s.getSettlements(ctx, objID)
	if err != nil {
		settlements = []string{}
	}

	summary := "📊 SPLIT BILL - TỔNG KẾT NHÓM\n"
	summary += "═══════════════════════════════\n"
	summary += "Nhóm: " + groupStats.GroupName + "\n"
	summary += "───────────────────────────────\n\n"

	summary += "💰 TỔNG QUAN\n"
	summary += formatAmount("  Tổng chi tiêu", groupStats.TotalSpent)
	summary += formatCount("  Số hóa đơn", groupStats.TotalBills)
	summary += formatCount("  Số thành viên", groupStats.TotalMembers)
	summary += formatAmount("  Trung bình/hóa đơn", groupStats.AverageBill)
	summary += "\n"

	if len(groupStats.MemberStats) > 0 {
		summary += "👥 CHI TIÊU THEO THÀNH VIÊN\n"
		for _, m := range groupStats.MemberStats {
			summary += "  " + m.DisplayName + ":\n"
			summary += formatAmount("    Đã trả", m.TotalPaid)
			summary += formatAmount("    Phần phải trả", m.TotalOwed)
			summary += formatAmount("    Số dư", m.NetBalance)
			summary += "\n"
		}
	}

	if len(groupStats.CategoryStats) > 0 {
		summary += "📁 CHI TIÊU THEO DANH MỤC\n"
		for _, c := range groupStats.CategoryStats {
			summary += formatCategorySummary("  "+c.Category, c.Total, c.Percentage)
		}
		summary += "\n"
	}

	if len(settlements) > 0 {
		summary += "🔄 GỢI Ý THANH TOÁN\n"
		for _, s := range settlements {
			summary += "  " + s + "\n"
		}
		summary += "\n"
	}

	summary += "───────────────────────────────\n"
	summary += "🕐 Xuất lúc: " + time.Now().Format("15:04 02/01/2006") + "\n"
	summary += "📱 Split Bill App\n"

	return summary, nil
}

func (s *StatsService) getSettlements(ctx context.Context, groupID primitive.ObjectID) ([]string, error) {
	// Simple settlement text - reuse debt optimizer logic
	bills, err := s.billRepo.FindActiveByGroupID(ctx, groupID)
	if err != nil {
		return nil, err
	}

	netAmounts := make(map[string]float64)
	nameMap := make(map[string]string)

	for _, bill := range bills {
		paidByID := bill.PaidBy.Hex()
		netAmounts[paidByID] += bill.TotalAmount

		for _, split := range bill.Splits {
			uid := split.UserID.Hex()
			netAmounts[uid] -= split.Amount
		}
	}

	// Get names
	for uid := range netAmounts {
		objID, err := primitive.ObjectIDFromHex(uid)
		if err != nil {
			continue
		}
		user, err := s.userRepo.FindByID(ctx, objID)
		if err == nil {
			nameMap[uid] = user.DisplayName
		}
	}

	// Simple greedy settlement
	type entry struct {
		id     string
		amount float64
	}

	var creditors, debtors []entry
	for uid, amount := range netAmounts {
		if amount > 0.01 {
			creditors = append(creditors, entry{uid, amount})
		} else if amount < -0.01 {
			debtors = append(debtors, entry{uid, -amount})
		}
	}

	sort.Slice(creditors, func(i, j int) bool {
		return creditors[i].amount > creditors[j].amount
	})
	sort.Slice(debtors, func(i, j int) bool {
		return debtors[i].amount > debtors[j].amount
	})

	var results []string
	ci, di := 0, 0
	for ci < len(creditors) && di < len(debtors) {
		transfer := creditors[ci].amount
		if debtors[di].amount < transfer {
			transfer = debtors[di].amount
		}

		fromName := nameMap[debtors[di].id]
		if fromName == "" {
			fromName = debtors[di].id
		}
		toName := nameMap[creditors[ci].id]
		if toName == "" {
			toName = creditors[ci].id
		}

		results = append(results, fromName+" → "+toName+": "+statsFormatVND(transfer))

		creditors[ci].amount -= transfer
		debtors[di].amount -= transfer

		if creditors[ci].amount < 0.01 {
			ci++
		}
		if debtors[di].amount < 0.01 {
			di++
		}
	}

	return results, nil
}

func formatAmount(label string, amount float64) string {
	return label + ": " + statsFormatVND(amount) + "\n"
}

func formatCount(label string, count int) string {
	return label + ": " + intToStr(count) + "\n"
}

func formatCategorySummary(label string, amount float64, pct float64) string {
	return label + ": " + statsFormatVND(amount) + " (" + floatToStr(pct) + "%)\n"
}

func statsFormatVND(amount float64) string {
	if amount >= 1000000 {
		return floatToStr(amount/1000000) + "tr₫"
	}
	if amount >= 1000 {
		return floatToStr(amount/1000) + "k₫"
	}
	return floatToStr(amount) + "₫"
}

func intToStr(n int) string {
	if n == 0 {
		return "0"
	}
	result := ""
	neg := false
	if n < 0 {
		neg = true
		n = -n
	}
	for n > 0 {
		result = string(rune('0'+n%10)) + result
		n /= 10
	}
	if neg {
		result = "-" + result
	}
	return result
}

func floatToStr(f float64) string {
	// Simple float formatting
	intPart := int(f)
	fracPart := int((f - float64(intPart)) * 10)
	if fracPart == 0 {
		return intToStr(intPart)
	}
	if fracPart < 0 {
		fracPart = -fracPart
	}
	return intToStr(intPart) + "." + intToStr(fracPart)
}
