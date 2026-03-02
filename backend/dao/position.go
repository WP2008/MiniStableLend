package dao

import (
	"time"
)

type Position struct {
	ID                 int64     `gorm:"primaryKey;autoIncrement"`
	UserAddress        string    `gorm:"type:varchar(42);not null;index:idx_user_address"` // 用户钱包地址
	CollateralAmount   float64   `gorm:"type:decimal(30,18);not null;default:0"`           // 质押的ETH/stETH数量（18位小数）
	DebtAmount         float64   `gorm:"type:decimal(30,18);not null;default:0"`           // 当前债务mUSD数量（18位小数）
	InitialDebt        float64   `gorm:"type:decimal(30,18);not null;default:0"`           // 初始借款本金（18位小数）
	HealthFactor       float64   `gorm:"type:decimal(10,4);not null;default:0"`            // 健康因子
	InterestRate       float64   `gorm:"type:decimal(10,8);not null;default:0"`            // 借款利率
	LastInterestUpdate time.Time `gorm:"not null"`                                         // 最后利息更新时间
	IsActive           bool      `gorm:"not null;default:true"`                            // 仓位是否活跃
	LiquidationPrice   float64   `gorm:"type:decimal(30,18);not null;default:0"`           // 清算价格（18位小数）
	CreatedAt          time.Time `gorm:"not null"`
	UpdatedAt          time.Time `gorm:"not null"`
}

// CreatePosition 创建新的质押仓位
func CreatePosition(position *Position) error {
	position.CreatedAt = time.Now()
	position.UpdatedAt = time.Now()
	return Mysql.Create(position).Error
}

// GetPositionByID 根据ID获取仓位
func GetPositionByID(id int64) (*Position, error) {
	var position Position
	err := Mysql.Where("id = ?", id).First(&position).Error
	return &position, err
}

// GetPositionByUserAddress 根据用户地址获取仓位
func GetPositionByUserAddress(userAddress string) (*Position, error) {
	var position Position
	err := Mysql.Where("user_address = ?", userAddress).First(&position).Error
	return &position, err
}

// UpdatePosition 更新仓位信息
func UpdatePosition(position *Position) error {
	position.UpdatedAt = time.Now()
	return Mysql.Save(position).Error
}
