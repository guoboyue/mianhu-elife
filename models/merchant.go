// models/merchant.go
package models

import (
	"github.com/jinzhu/gorm"
	"mianhuelife/models"
)

type Merchant struct {
	models.Model

	Name            string  `json:"name"`
	Logo            string  `json:"logo"`
	CoverImage      string  `json:"cover_image"`
	BusinessLicense string  `json:"business_license"`
	ContactName     string  `json:"contact_name"`
	ContactPhone    string  `json:"contact_phone"`
	Province        string  `json:"province"`
	City            string  `json:"city"`
	District        string  `json:"district"`
	Address         string  `json:"address"`
	Longitude       float64 `json:"longitude"`
	Latitude        float64 `json:"latitude"`
	Rating          float64 `json:"rating" gorm:"default:5.0"`
	MonthSales      int     `json:"month_sales" gorm:"default:0"`
	BusinessHours   string  `json:"business_hours"`
	TableCount      int     `json:"table_count" gorm:"default:0"`
	AdminID         int     `json:"admin_id"`
	Status          int     `json:"status" gorm:"default:1"` // 状态：0-休息 1-营业中 2-待审核 3-已禁用
	Notice          string  `json:"notice"`
}

// ExistMerchantByID 检查商户是否存在
func ExistMerchantByID(id int) (bool, error) {
	var merchant Merchant
	err := models.Db.Select("id").Where("id = ? AND deleted_on = 0", id).First(&merchant).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return false, err
	}
	return merchant.ID > 0, nil
}

// GetMerchantTotal 获取商户总数
func GetMerchantTotal(maps interface{}) (int64, error) {
	var count int64
	if err := Db.Model(&Merchant{}).Where(maps).Where("deleted_on = 0").Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

// GetMerchants 获取商户列表
func GetMerchants(pageNum int, pageSize int, maps interface{}) ([]*Merchant, error) {
	var merchants []*Merchant
	err := Db.Where(maps).
		Where("deleted_on = 0").
		Offset(pageNum).
		Limit(pageSize).
		Find(&merchants).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	return merchants, nil
}

// GetMerchant 获取单个商户信息
func GetMerchant(id int) (*Merchant, error) {
	var merchant Merchant
	err := Db.Where("id = ? AND deleted_on = 0", id).First(&merchant).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	return &merchant, nil
}

// GetNearbyMerchants 获取附近商户
func GetNearbyMerchants(longitude, latitude float64, distance float64) ([]*Merchant, error) {
	var merchants []*Merchant
	// 使用MySQL空间查询，这里的6371是地球半径(km)
	err := Db.Where("deleted_on = 0").
		Where("6371 * acos(cos(radians(?)) * cos(radians(latitude)) * cos(radians(longitude) - radians(?)) + sin(radians(?)) * sin(radians(latitude))) <= ?",
			latitude, longitude, latitude, distance/1000). // distance转换为km
		Find(&merchants).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	return merchants, nil
}

// AddMerchant 添加商户
func AddMerchant(merchant *Merchant) error {
	return Db.Create(merchant).Error
}

// UpdateMerchant 更新商户
func UpdateMerchant(id int, data interface{}) error {
	return Db.Model(&Merchant{}).Where("id = ? AND deleted_on = 0", id).Updates(data).Error
}

// DeleteMerchant 删除商户
func DeleteMerchant(id int) error {
	return Db.Where("id = ?", id).Delete(&Merchant{}).Error
}

// UpdateMerchantStatus 更新商户状态
func UpdateMerchantStatus(id int, status int) error {
	return Db.Model(&Merchant{}).Where("id = ? AND deleted_on = 0", id).Update("status", status).Error
}
