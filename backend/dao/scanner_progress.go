package dao

import (
	"time"

	"gorm.io/gorm"
)

// ScannerProgress 记录扫链进度
type ScannerProgress struct {
	ID                   int64     `gorm:"primaryKey;autoIncrement"`
	ContractAddress      string    `gorm:"type:varchar(42);not null;uniqueIndex"` // 合约地址
	LastScannedBlock     uint64    `gorm:"not null;default:0"`                    // 最后扫描的区块高度
	LastScanTime         time.Time `gorm:"not null"`                              // 最后扫描时间
	TotalEventsProcessed int64     `gorm:"not null;default:0"`                    // 总共处理的事件数
	CreatedAt            time.Time `gorm:"not null"`
	UpdatedAt            time.Time `gorm:"not null"`
}

// GetScannerProgress 获取指定合约的扫描进度
func GetScannerProgress(contractAddress string) (*ScannerProgress, error) {
	var progress ScannerProgress
	err := Mysql.Where("contract_address = ?", contractAddress).First(&progress).Error
	return &progress, err
}

// GetOrCreateScannerProgress 获取或创建扫描进度
func GetOrCreateScannerProgress(contractAddress string) (*ScannerProgress, error) {
	progress, err := GetScannerProgress(contractAddress)
	if err != nil {
		// 创建新的进度记录
		progress = &ScannerProgress{
			ContractAddress:      contractAddress,
			LastScannedBlock:     0,
			LastScanTime:         time.Now(),
			TotalEventsProcessed: 0,
			CreatedAt:            time.Now(),
			UpdatedAt:            time.Now(),
		}
		if err := Mysql.Create(progress).Error; err != nil {
			return nil, err
		}
		return progress, nil
	}
	return progress, nil
}

// UpdateScannerProgress 更新扫描进度
func UpdateScannerProgress(progress *ScannerProgress) error {
	progress.UpdatedAt = time.Now()
	return Mysql.Save(progress).Error
}

// UpdateLastScannedBlock 更新最后扫描的区块高度
func UpdateLastScannedBlock(contractAddress string, blockNumber uint64, eventsCount int64) error {
	return Mysql.Model(&ScannerProgress{}).
		Where("contract_address = ?", contractAddress).
		Updates(map[string]interface{}{
			"last_scanned_block":     blockNumber,
			"last_scan_time":         time.Now(),
			"total_events_processed": gorm.Expr("total_events_processed + ?", eventsCount),
		}).Error
}
