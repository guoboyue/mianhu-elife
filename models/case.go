package models

// CaseProgress 案件进度主表
type CaseProgress struct {
	Model

	ServiceCode    string `json:"service_code"`
	ServiceVersion int    `json:"service_version"`
	LawyerCode     string `json:"lawyer_code"`
	OpenID         string `json:"open_id"`
	CaseType       int    `json:"case_type"`
	CurrentStage   int    `json:"current_stage"`
	CaseStatus     int    `json:"case_status"`
}

// TableName 设置表名
func (CaseProgress) TableName() string {
	return "law_case_progress"
}

// CaseProgressDetail 案件进度明细表
type CaseProgressDetail struct {
	Model

	ProgressID  int    `json:"progress_id"`
	Stage       int    `json:"stage"`
	Content     string `json:"content"`
	Attachments string `json:"attachments"`
}

// TableName 设置表名
func (CaseProgressDetail) TableName() string {
	return "law_case_progress_detail"
}

// GetCaseProgressByID 获取单个案件进度
func GetCaseProgressByID(id int) (*CaseProgress, error) {
	var progress CaseProgress
	err := Db.Where("id = ? AND deleted_on = 0", id).First(&progress).Error
	if err != nil {
		return nil, err
	}
	return &progress, nil
}

// GetCaseProgressList 获取案件进度列表
func GetCaseProgressList(maps map[string]interface{}, page, pageSize int) ([]*CaseProgress, error) {
	var progress []*CaseProgress
	maps["deleted_on"] = 0

	err := Db.Where(maps).Offset(page).Limit(pageSize).Find(&progress).Error
	if err != nil {
		return nil, err
	}
	return progress, nil
}

// GetCaseProgressTotal 获取案件进度总数
func GetCaseProgressTotal(maps map[string]interface{}) (int64, error) {
	var count int64
	maps["deleted_on"] = 0

	err := Db.Model(&CaseProgress{}).Where(maps).Count(&count).Error
	if err != nil {
		return 0, err
	}
	return count, nil
}

// AddCaseProgress 新增案件进度
func AddCaseProgress(data map[string]interface{}) error {
	progress := CaseProgress{
		ServiceCode:    data["service_code"].(string),
		ServiceVersion: data["service_version"].(int),
		LawyerCode:     data["lawyer_code"].(string),
		OpenID:         data["open_id"].(string),
		CaseType:       data["case_type"].(int),
		CurrentStage:   data["current_stage"].(int),
		CaseStatus:     data["case_status"].(int),
	}
	if err := Db.Create(&progress).Error; err != nil {
		return err
	}

	// 如果有内容，则添加进度详情
	if content, ok := data["content"]; ok && content != "" {
		detail := CaseProgressDetail{
			ProgressID: progress.ID,
			Stage:      progress.CurrentStage,
			Content:    content.(string),
		}
		if attachments, ok := data["attachments"]; ok {
			detail.Attachments = attachments.(string)
		}
		if err := Db.Create(&detail).Error; err != nil {
			return err
		}
	}

	return nil
}

// EditCaseProgress 修改案件进度
func EditCaseProgress(id int, data map[string]interface{}) error {
	if err := Db.Model(&CaseProgress{}).Where("id = ? AND deleted_on = 0", id).Updates(data).Error; err != nil {
		return err
	}
	return nil
}

// DeleteCaseProgress 删除案件进度
func DeleteCaseProgress(id int) error {
	if err := Db.Where("id = ? AND deleted_on = 0", id).Delete(&CaseProgress{}).Error; err != nil {
		return err
	}
	// 同时删除相关的进度详情
	if err := Db.Where("progress_id = ? AND deleted_on = 0", id).Delete(&CaseProgressDetail{}).Error; err != nil {
		return err
	}
	return nil
}

// GetCaseProgressDetails 获取案件进度详情列表
func GetCaseProgressDetails(progressID int) ([]*CaseProgressDetail, error) {
	var details []*CaseProgressDetail
	err := Db.Where("progress_id = ? AND deleted_on = 0", progressID).
		Order("created_on DESC").
		Find(&details).Error
	if err != nil {
		return nil, err
	}
	return details, nil
}

// AddCaseProgressDetail 新增案件进度详情
func AddCaseProgressDetail(data map[string]interface{}) error {
	detail := CaseProgressDetail{
		ProgressID: data["progress_id"].(int),
		Stage:      data["stage"].(int),
		Content:    data["content"].(string),
	}
	if attachments, ok := data["attachments"]; ok {
		detail.Attachments = attachments.(string)
	}

	if err := Db.Create(&detail).Error; err != nil {
		return err
	}
	return nil
}

// GetLatestProgressDetail 获取最新的进度详情
func GetLatestProgressDetail(progressID int) (*CaseProgressDetail, error) {
	var detail CaseProgressDetail
	err := Db.Where("progress_id = ? AND deleted_on = 0", progressID).
		Order("created_on DESC").
		First(&detail).Error
	if err != nil {
		return nil, err
	}
	return &detail, nil
}
