package v1

import (
	"github.com/gin-gonic/gin"
	"github.com/unknwon/com"
	"lvsuobao/models"
	"lvsuobao/pkg/app"
	"lvsuobao/pkg/e"
	"lvsuobao/pkg/util"
	"net/http"
)

// GetCaseProgress 获取案件进度详情
func GetCaseProgress(c *gin.Context) {
	appG := app.Gin{C: c}
	id := com.StrTo(c.Param("id")).MustInt()

	progress, err := models.GetCaseProgressByID(id)
	if err != nil {
		appG.Response(http.StatusInternalServerError, e.ERROR_GET_PROGRESS_FAIL, nil)
		return
	}

	details, err := models.GetCaseProgressDetails(id)
	if err != nil {
		appG.Response(http.StatusInternalServerError, e.ERROR_GET_PROGRESS_FAIL, nil)
		return
	}

	data := make(map[string]interface{})
	data["progress"] = progress
	data["details"] = details

	appG.Response(http.StatusOK, e.SUCCESS, data)
}

// GetCaseProgressList 获取案件进度列表
func GetCaseProgressList(c *gin.Context) {
	appG := app.Gin{C: c}

	maps := make(map[string]interface{})

	// 获取查询参数
	if lawyerCode := c.Query("lawyer_code"); lawyerCode != "" {
		maps["lawyer_code"] = lawyerCode
	}
	if openID := c.Query("open_id"); openID != "" {
		maps["open_id"] = openID
	}
	if serviceCode := c.Query("service_code"); serviceCode != "" {
		maps["service_code"] = serviceCode
	}

	// 处理分页
	page := util.GetPage(c)
	pageSize := 10 // 每页显示数量

	list, err := models.GetCaseProgressList(maps, page, pageSize)
	if err != nil {
		appG.Response(http.StatusInternalServerError, e.ERROR_GET_PROGRESSES_FAIL, nil)
		return
	}

	total, err := models.GetCaseProgressTotal(maps)
	if err != nil {
		appG.Response(http.StatusInternalServerError, e.ERROR_COUNT_PROGRESS_FAIL, nil)
		return
	}

	data := make(map[string]interface{})
	data["list"] = list
	data["total"] = total

	appG.Response(http.StatusOK, e.SUCCESS, data)
}

// AddCaseProgress 添加案件进度
func AddCaseProgress(c *gin.Context) {
	appG := app.Gin{C: c}

	// 绑定请求参数
	type AddProgressForm struct {
		ServiceCode    string `json:"service_code" valid:"Required;MaxSize(50)"`
		ServiceVersion int    `json:"service_version" valid:"Required"`
		LawyerCode     string `json:"lawyer_code" valid:"Required;MaxSize(50)"`
		OpenID         string `json:"open_id" valid:"Required;MaxSize(50)"`
		CaseType       int    `json:"case_type" valid:"Required;Range(1,5)"`
		CurrentStage   int    `json:"current_stage" valid:"Required;Range(1,5)"`
		CaseStatus     int    `json:"case_status" valid:"Required;Range(1,3)"`
		Content        string `json:"content" valid:"Required;MaxSize(1000)"`
	}

	var form AddProgressForm
	if err := c.ShouldBindJSON(&form); err != nil {
		appG.Response(http.StatusBadRequest, e.INVALID_PARAMS, nil)
		return
	}

	// 构建数据
	data := make(map[string]interface{})
	data["service_code"] = form.ServiceCode
	data["service_version"] = form.ServiceVersion
	data["lawyer_code"] = form.LawyerCode
	data["open_id"] = form.OpenID
	data["case_type"] = form.CaseType
	data["current_stage"] = form.CurrentStage
	data["case_status"] = form.CaseStatus

	// 创建案件进度
	err := models.AddCaseProgress(data)
	if err != nil {
		appG.Response(http.StatusInternalServerError, e.ERROR_ADD_PROGRESS_FAIL, nil)
		return
	}

	appG.Response(http.StatusOK, e.SUCCESS, nil)
}

// UpdateCaseProgress 更新案件进度
func UpdateCaseProgress(c *gin.Context) {
	appG := app.Gin{C: c}
	id := com.StrTo(c.Param("id")).MustInt()

	type UpdateProgressForm struct {
		CurrentStage int    `json:"current_stage" valid:"Required;Range(1,5)"`
		CaseStatus   int    `json:"case_status" valid:"Required;Range(1,3)"`
		Content      string `json:"content" valid:"Required;MaxSize(1000)"`
	}

	var form UpdateProgressForm
	if err := c.ShouldBindJSON(&form); err != nil {
		appG.Response(http.StatusBadRequest, e.INVALID_PARAMS, nil)
		return
	}

	// 更新数据
	data := make(map[string]interface{})
	data["current_stage"] = form.CurrentStage
	data["case_status"] = form.CaseStatus

	err := models.EditCaseProgress(id, data)
	if err != nil {
		appG.Response(http.StatusInternalServerError, e.ERROR_UPDATE_PROGRESS_FAIL, nil)
		return
	}

	// 添加进度详情
	detailData := make(map[string]interface{})
	detailData["progress_id"] = id
	detailData["stage"] = form.CurrentStage
	detailData["content"] = form.Content

	err = models.AddCaseProgressDetail(detailData)
	if err != nil {
		appG.Response(http.StatusInternalServerError, e.ERROR_ADD_PROGRESS_FAIL, nil)
		return
	}

	appG.Response(http.StatusOK, e.SUCCESS, nil)
}

// DeleteCaseProgress 删除案件进度
func DeleteCaseProgress(c *gin.Context) {
	appG := app.Gin{C: c}
	id := com.StrTo(c.Param("id")).MustInt()

	err := models.DeleteCaseProgress(id)
	if err != nil {
		appG.Response(http.StatusInternalServerError, e.ERROR_DELETE_PROGRESS_FAIL, nil)
		return
	}

	appG.Response(http.StatusOK, e.SUCCESS, nil)
}
