package analytics

import (
	"time"

	"gorm.io/gorm"
)

type AnalyticsService struct {
	db *gorm.DB
}

type DailyConversationCount struct {
	Date  string `json:"date"`
	Count int64  `json:"count"`
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

	var rows []row
	err := as.db.Table("conversations").
		Select("DATE(created_at) as date, COUNT(*) as count").
		Where("created_at >= ? AND created_at < ?", start, endExclusive).
		Group("DATE(created_at)").
		Order("DATE(created_at) ASC").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	countByDay := make(map[string]int64, len(rows))
	for _, r := range rows {
		countByDay[r.Date] = r.Count
	}

	result := make([]DailyConversationCount, 0, days)
	for i := 0; i < days; i++ {
		day := start.AddDate(0, 0, i).Format("2006-01-02")
		result = append(result, DailyConversationCount{
			Date:  day,
			Count: countByDay[day],
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
