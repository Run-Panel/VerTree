package services

import (
	"fmt"
	"net"
	"time"

	"github.com/Run-Panel/VerTree/internal/database"
	"github.com/Run-Panel/VerTree/internal/models"
	"gorm.io/gorm"
)

// StatsService handles statistics-related business logic
type StatsService struct {
	db *gorm.DB
}

// NewStatsService creates a new stats service instance
func NewStatsService() *StatsService {
	return &StatsService{
		db: database.DB,
	}
}

// RecordUpdateStat records an update statistic
func (s *StatsService) RecordUpdateStat(req *models.UpdateStatRequest, clientIP string) error {
	// Parse IP address
	var ipAddr net.IP
	if clientIP != "" {
		ipAddr = net.ParseIP(clientIP)
	}

	stat := &models.UpdateStat{
		Version:       req.Version,
		ClientID:      req.ClientID,
		ClientVersion: req.ClientVersion,
		Region:        req.Region,
		IPAddress:     ipAddr,
		UserAgent:     req.UserAgent,
		Action:        req.Action,
		ErrorMessage:  req.ErrorMessage,
	}

	if err := s.db.Create(stat).Error; err != nil {
		return fmt.Errorf("failed to record update stat: %w", err)
	}

	return nil
}

// GetStats retrieves statistics based on the request
func (s *StatsService) GetStats(req *models.StatsRequest) (*models.StatsResponse, error) {
	// Calculate time range
	var startTime time.Time
	now := time.Now()

	switch req.Period {
	case "1d":
		startTime = now.AddDate(0, 0, -1)
	case "7d":
		startTime = now.AddDate(0, 0, -7)
	case "30d":
		startTime = now.AddDate(0, 0, -30)
	case "90d":
		startTime = now.AddDate(0, 0, -90)
	default:
		startTime = now.AddDate(0, 0, -7) // Default to 7 days
	}

	query := s.db.Model(&models.UpdateStat{}).Where("created_at >= ?", startTime)

	// Apply action filter if specified
	if req.Action != "all" {
		query = query.Where("action = ?", req.Action)
	}

	// Get total unique users
	var totalUsers int64
	if err := s.db.Model(&models.UpdateStat{}).
		Where("created_at >= ?", startTime).
		Distinct("client_id").
		Count(&totalUsers).Error; err != nil {
		return nil, fmt.Errorf("failed to count total users: %w", err)
	}

	// Get total downloads
	var totalDownloads int64
	if err := s.db.Model(&models.UpdateStat{}).
		Where("created_at >= ? AND action = ?", startTime, "download").
		Count(&totalDownloads).Error; err != nil {
		return nil, fmt.Errorf("failed to count total downloads: %w", err)
	}

	// Calculate success rate
	var successCount int64
	var totalAttempts int64

	if err := s.db.Model(&models.UpdateStat{}).
		Where("created_at >= ? AND action = ?", startTime, "success").
		Count(&successCount).Error; err != nil {
		return nil, fmt.Errorf("failed to count successful updates: %w", err)
	}

	if err := s.db.Model(&models.UpdateStat{}).
		Where("created_at >= ? AND action IN (?)", startTime, []string{"success", "failed"}).
		Count(&totalAttempts).Error; err != nil {
		return nil, fmt.Errorf("failed to count total attempts: %w", err)
	}

	var successRate float64
	if totalAttempts > 0 {
		successRate = float64(successCount) / float64(totalAttempts) * 100
	}

	// Get version distribution
	versionDistribution, err := s.getVersionDistribution(startTime)
	if err != nil {
		return nil, fmt.Errorf("failed to get version distribution: %w", err)
	}

	// Get region distribution
	regionDistribution, err := s.getRegionDistribution(startTime)
	if err != nil {
		return nil, fmt.Errorf("failed to get region distribution: %w", err)
	}

	// Get daily stats
	dailyStats, err := s.getDailyStats(startTime)
	if err != nil {
		return nil, fmt.Errorf("failed to get daily stats: %w", err)
	}

	return &models.StatsResponse{
		TotalUsers:          totalUsers,
		TotalDownloads:      totalDownloads,
		SuccessRate:         successRate,
		VersionDistribution: versionDistribution,
		RegionDistribution:  regionDistribution,
		DailyStats:          dailyStats,
	}, nil
}

// getVersionDistribution gets version distribution statistics
func (s *StatsService) getVersionDistribution(startTime time.Time) (map[string]int64, error) {
	type VersionCount struct {
		Version string
		Count   int64
	}

	var results []VersionCount
	if err := s.db.Model(&models.UpdateStat{}).
		Select("client_version as version, COUNT(DISTINCT client_id) as count").
		Where("created_at >= ? AND client_version != ''", startTime).
		Group("client_version").
		Order("count DESC").
		Scan(&results).Error; err != nil {
		return nil, err
	}

	distribution := make(map[string]int64)
	for _, result := range results {
		if result.Version != "" {
			distribution[result.Version] = result.Count
		}
	}

	return distribution, nil
}

// getRegionDistribution gets region distribution statistics
func (s *StatsService) getRegionDistribution(startTime time.Time) (map[string]int64, error) {
	type RegionCount struct {
		Region string
		Count  int64
	}

	var results []RegionCount
	if err := s.db.Model(&models.UpdateStat{}).
		Select("region, COUNT(DISTINCT client_id) as count").
		Where("created_at >= ? AND region != ''", startTime).
		Group("region").
		Order("count DESC").
		Scan(&results).Error; err != nil {
		return nil, err
	}

	distribution := make(map[string]int64)
	for _, result := range results {
		if result.Region == "" {
			distribution["unknown"] = result.Count
		} else {
			distribution[result.Region] = result.Count
		}
	}

	return distribution, nil
}

// getDailyStats gets daily statistics
func (s *StatsService) getDailyStats(startTime time.Time) ([]models.DailyStat, error) {
	type DailyCount struct {
		Date   string
		Action string
		Count  int64
	}

	var results []DailyCount
	if err := s.db.Model(&models.UpdateStat{}).
		Select("DATE(created_at) as date, action, COUNT(*) as count").
		Where("created_at >= ?", startTime).
		Group("DATE(created_at), action").
		Order("date").
		Scan(&results).Error; err != nil {
		return nil, err
	}

	// Organize data by date
	dailyMap := make(map[string]*models.DailyStat)
	for _, result := range results {
		if dailyMap[result.Date] == nil {
			dailyMap[result.Date] = &models.DailyStat{
				Date: result.Date,
			}
		}

		switch result.Action {
		case "download":
			dailyMap[result.Date].Downloads = result.Count
		case "success":
			dailyMap[result.Date].Installs = result.Count
		case "failed":
			dailyMap[result.Date].Failures = result.Count
		}
	}

	// Convert map to slice and sort by date
	var dailyStats []models.DailyStat
	for _, stat := range dailyMap {
		dailyStats = append(dailyStats, *stat)
	}

	return dailyStats, nil
}

// GetVersionDistribution gets version distribution for a specific period
func (s *StatsService) GetVersionDistribution(period string) (map[string]int64, error) {
	// Calculate time range
	var startTime time.Time
	now := time.Now()

	switch period {
	case "1d":
		startTime = now.AddDate(0, 0, -1)
	case "7d":
		startTime = now.AddDate(0, 0, -7)
	case "30d":
		startTime = now.AddDate(0, 0, -30)
	case "90d":
		startTime = now.AddDate(0, 0, -90)
	default:
		startTime = now.AddDate(0, 0, -7) // Default to 7 days
	}

	return s.getVersionDistribution(startTime)
}

// GetRegionDistribution gets region distribution for a specific period
func (s *StatsService) GetRegionDistribution(period string) (map[string]int64, error) {
	// Calculate time range
	var startTime time.Time
	now := time.Now()

	switch period {
	case "1d":
		startTime = now.AddDate(0, 0, -1)
	case "7d":
		startTime = now.AddDate(0, 0, -7)
	case "30d":
		startTime = now.AddDate(0, 0, -30)
	case "90d":
		startTime = now.AddDate(0, 0, -90)
	default:
		startTime = now.AddDate(0, 0, -7) // Default to 7 days
	}

	return s.getRegionDistribution(startTime)
}
