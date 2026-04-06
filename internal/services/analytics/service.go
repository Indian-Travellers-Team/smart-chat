package analytics

import (
	"time"

	"gorm.io/gorm"
)

type AnalyticsService struct {
	db *gorm.DB
}

type DailyConversationCount struct {
	Date            string `json:"date"`
	Count           int64  `json:"count"`
	ConversationIDs []uint `json:"conversationIds"`
}

type DashboardConversationSummary struct {
	TotalConversations        int64   `json:"totalConversations"`
	CurrentWeekConversations  int64   `json:"currentWeekConversations"`
	PreviousWeekConversations int64   `json:"previousWeekConversations"`
	WeeklyChangePercentage    float64 `json:"weeklyChangePercentage"`
	Trend                     string  `json:"trend"`
	PeriodLabel               string  `json:"periodLabel"`
}

func NewAnalyticsService(db *gorm.DB) *AnalyticsService {
	return &AnalyticsService{db: db}
}

func (as *AnalyticsService) GetConversationCountsByDateRange(startDate, endDate time.Time) ([]DailyConversationCount, error) {
	start := time.Date(startDate.UTC().Year(), startDate.UTC().Month(), startDate.UTC().Day(), 0, 0, 0, 0, time.UTC)
	end := time.Date(endDate.UTC().Year(), endDate.UTC().Month(), endDate.UTC().Day(), 0, 0, 0, 0, time.UTC)

	if end.Before(start) {
		return []DailyConversationCount{}, nil
	}

	days := int(end.Sub(start).Hours()/24) + 1
	endExclusive := end.AddDate(0, 0, 1)

	type row struct {
		Date  string
		Count int64
	}

	type conversationRow struct {
		Date           string
		ConversationID uint
	}

	var rows []row
	err := as.db.Table("conversations").
		Select("CAST(DATE(created_at) AS TEXT) as date, COUNT(*) as count").
		Where("created_at >= ? AND created_at < ?", start, endExclusive).
		Group("DATE(created_at)").
		Order("DATE(created_at) ASC").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	var conversationRows []conversationRow
	err = as.db.Table("conversations").
		Select("CAST(DATE(created_at) AS TEXT) as date, id as conversation_id").
		Where("created_at >= ? AND created_at < ?", start, endExclusive).
		Order("created_at ASC").
		Scan(&conversationRows).Error
	if err != nil {
		return nil, err
	}

	countByDay := make(map[string]int64, len(rows))
	for _, r := range rows {
		countByDay[r.Date] = r.Count
	}

	conversationIDsByDay := make(map[string][]uint)
	for _, row := range conversationRows {
		conversationIDsByDay[row.Date] = append(conversationIDsByDay[row.Date], row.ConversationID)
	}

	result := make([]DailyConversationCount, 0, days)
	for i := 0; i < days; i++ {
		day := start.AddDate(0, 0, i).Format("2006-01-02")
		conversationIDs := conversationIDsByDay[day]
		if conversationIDs == nil {
			conversationIDs = []uint{}
		}

		result = append(result, DailyConversationCount{
			Date:            day,
			Count:           countByDay[day],
			ConversationIDs: conversationIDs,
		})
	}

	return result, nil
}

func (as *AnalyticsService) GetConversationCountsLastNDays(days int, now time.Time) ([]DailyConversationCount, error) {
	if days <= 0 {
		return []DailyConversationCount{}, nil
	}

	end := time.Date(now.UTC().Year(), now.UTC().Month(), now.UTC().Day(), 0, 0, 0, 0, time.UTC)
	start := end.AddDate(0, 0, -(days - 1))

	return as.GetConversationCountsByDateRange(start, end)
}

func (as *AnalyticsService) GetDashboardConversationSummary(now time.Time) (DashboardConversationSummary, error) {
	totalConversations, err := as.countAllConversations()
	if err != nil {
		return DashboardConversationSummary{}, err
	}

	currentWeekStart := startOfWeek(now)
	currentWeekEnd := currentWeekStart.AddDate(0, 0, 7)
	previousWeekStart := currentWeekStart.AddDate(0, 0, -7)

	currentWeekConversations, err := as.countConversationsBetween(previousWeekStart.AddDate(0, 0, 7), currentWeekEnd)
	if err != nil {
		return DashboardConversationSummary{}, err
	}

	previousWeekConversations, err := as.countConversationsBetween(previousWeekStart, currentWeekStart)
	if err != nil {
		return DashboardConversationSummary{}, err
	}

	weeklyChangePercentage := calculatePercentageChange(previousWeekConversations, currentWeekConversations)
	trend := "flat"
	if currentWeekConversations > previousWeekConversations {
		trend = "up"
	} else if currentWeekConversations < previousWeekConversations {
		trend = "down"
	}

	return DashboardConversationSummary{
		TotalConversations:        totalConversations,
		CurrentWeekConversations:  currentWeekConversations,
		PreviousWeekConversations: previousWeekConversations,
		WeeklyChangePercentage:    weeklyChangePercentage,
		Trend:                     trend,
		PeriodLabel:               "this week",
	}, nil
}

func (as *AnalyticsService) countAllConversations() (int64, error) {
	var total int64
	err := as.db.Table("conversations").Count(&total).Error
	return total, err
}

func (as *AnalyticsService) countConversationsBetween(start, end time.Time) (int64, error) {
	var total int64
	err := as.db.Table("conversations").
		Where("created_at >= ? AND created_at < ?", start, end).
		Count(&total).Error
	return total, err
}

func startOfWeek(now time.Time) time.Time {
	weekday := int(now.Weekday())
	if weekday == 0 {
		weekday = 7
	}

	start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	return start.AddDate(0, 0, -(weekday - 1))
}

func calculatePercentageChange(previousValue, currentValue int64) float64 {
	if previousValue == 0 {
		if currentValue == 0 {
			return 0
		}
		return 100
	}

	return (float64(currentValue-previousValue) / float64(previousValue)) * 100
}
