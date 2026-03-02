package service

import (
	"minilend/dao"
)

type PositionService struct{}

// GetUserPosition 获取用户仓位信息
func (s *PositionService) GetUserPosition(userAddress string) (*dao.Position, error) {
	return dao.GetPositionByUserAddress(userAddress)
}
