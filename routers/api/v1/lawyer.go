// routers/api/v1/lawyer.go
package v1

import (
	"github.com/astaxie/beego/validation"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"lvsuobao/models"
	"lvsuobao/pkg/app"
	"lvsuobao/pkg/e"
	"lvsuobao/pkg/logging"
	"net/http"
	"strconv"
	"time"
)

// GetLawyer 获取律师信息(公开)
func GetLawyer(c *gin.Context) {
	appG := app.Gin{C: c}

	// 获取查询参数
	id := c.Query("id")
	code := c.Query("code")

	if id == "" && code == "" {
		appG.Response(http.StatusBadRequest, e.INVALID_PARAMS, "缺少必要的查询参数(id或code)")
		return
	}

	var lawyer *models.LawyerDetail
	var err error

	// 优先使用ID查询,其次使用律师编号查询
	if id != "" {
		lawyerID, err := strconv.ParseInt(id, 10, 64)
		if err != nil {
			appG.Response(http.StatusBadRequest, e.INVALID_PARAMS, "无效的ID格式")
			return
		}
		lawyer, err = models.GetPublicLawyer(lawyerID)
	} else {
		lawyer, err = models.GetPublicLawyerByCode(code)
	}

	if err != nil {
		logging.Error("获取律师信息失败:", err)
		appG.Response(http.StatusInternalServerError, e.ERROR_GET_LAWYER_FAIL, nil)
		return
	}

	if lawyer == nil {
		appG.Response(http.StatusNotFound, e.ERROR_NOT_EXIST_LAWYER, nil)
		return
	}

	// 异步更新浏览量
	go func() {
		if err := models.Db.Model(&models.Lawyer{}).
			Where("id = ?", lawyer.ID).
			UpdateColumn("view_count", gorm.Expr("view_count + ?", 1)).Error; err != nil {
			logging.Error("更新律师浏览量失败:", err)
		}
	}()

	appG.Response(http.StatusOK, e.SUCCESS, lawyer)
}

// GetSelfLawyer 获取自己的律师信息(需要认证)
func GetSelfLawyer(c *gin.Context) {
	appG := app.Gin{C: c}

	// 获取当前登录用户
	user, exists := c.Get("user")
	if !exists {
		appG.Response(http.StatusUnauthorized, e.ERROR_AUTH, nil)
		return
	}
	currentUser := user.(*models.User)

	// 通过openID获取完整的律师信息
	lawyer, err := models.GetPrivateLawyerByOpenID(currentUser.OpenID)
	if err != nil {
		logging.Error("获取律师信息失败:", err)
		appG.Response(http.StatusInternalServerError, e.ERROR_GET_LAWYER_FAIL, nil)
		return
	}

	if lawyer == nil {
		appG.Response(http.StatusNotFound, e.ERROR_NOT_EXIST_LAWYER, nil)
		return
	}

	appG.Response(http.StatusOK, e.SUCCESS, lawyer)
}

// GetLawyers 获取律师列表
func GetLawyers(c *gin.Context) {
	appG := app.Gin{C: c}
	name := c.Query("name")
	legalField := c.Query("legal_field")

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	lawyers, total, err := models.GetLawyers(name, legalField, page, pageSize)
	if err != nil {
		logging.Error("获取律师列表失败:", err)
		appG.Response(http.StatusInternalServerError, e.ERROR_GET_LAWYERS_FAIL, nil)
		return
	}

	appG.Response(http.StatusOK, e.SUCCESS, map[string]interface{}{
		"lists":     lawyers,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// CreateLawyer 创建律师
func CreateLawyer(c *gin.Context) {
	appG := app.Gin{C: c}
	var lawyer models.Lawyer

	// 获取当前用户
	user, exists := c.Get("user")
	if !exists {
		appG.Response(http.StatusUnauthorized, e.ERROR_AUTH, nil)
		return
	}
	currentUser := user.(*models.User)

	// 绑定请求数据
	if err := c.ShouldBindJSON(&lawyer); err != nil {
		appG.Response(http.StatusBadRequest, e.INVALID_PARAMS, nil)
		return
	}

	// 验证必填字段
	valid := validation.Validation{}
	valid.Required(lawyer.Name, "name").Message("名称不能为空")
	valid.Required(lawyer.LawyerCode, "lawyer_code").Message("律师编号不能为空")
	valid.Required(lawyer.Phone, "phone").Message("手机号不能为空")

	if valid.HasErrors() {
		appG.Response(http.StatusBadRequest, e.INVALID_PARAMS, valid.Errors)
		return
	}

	// 检查律师编号是否已存在
	exists, err := models.ExistLawyerByCode(lawyer.LawyerCode)
	if err != nil {
		logging.Error("检查律师编号失败:", err)
		appG.Response(http.StatusInternalServerError, e.ERROR_ADD_LAWYER_FAIL, nil)
		return
	}
	if exists {
		appG.Response(http.StatusBadRequest, e.ERROR_EXIST_LAWYER_CODE, nil)
		return
	}

	// 设置创建人信息
	lawyer.OpenID = currentUser.OpenID
	lawyer.CreateBy = currentUser.OpenID
	lawyer.CreateTime = time.Now().Unix()
	lawyer.Status = 1
	lawyer.Version = 1
	lawyer.AuditStatus = 0 // 默认待审核

	// 创建律师记录
	if err := models.AddLawyer(&lawyer); err != nil {
		logging.Error("创建律师失败:", err)
		appG.Response(http.StatusInternalServerError, e.ERROR_ADD_LAWYER_FAIL, nil)
		return
	}

	appG.Response(http.StatusOK, e.SUCCESS, map[string]interface{}{
		"id": lawyer.ID,
	})
}

// UpdateLawyer 更新律师信息
func UpdateLawyer(c *gin.Context) {
	appG := app.Gin{C: c}

	// 获取当前用户
	user, exists := c.Get("user")
	if !exists {
		appG.Response(http.StatusUnauthorized, e.ERROR_AUTH, nil)
		return
	}
	currentUser := user.(*models.User)

	// 获取要更新的律师信息
	lawyer, err := models.GetLawyerByOpenID(currentUser.OpenID)
	if err != nil {
		logging.Error("获取律师信息失败:", err)
		appG.Response(http.StatusInternalServerError, e.ERROR_GET_LAWYER_FAIL, nil)
		return
	}
	if lawyer == nil {
		appG.Response(http.StatusNotFound, e.ERROR_NOT_EXIST_LAWYER, nil)
		return
	}

	var updateData map[string]interface{}
	if err := c.ShouldBindJSON(&updateData); err != nil {
		appG.Response(http.StatusBadRequest, e.INVALID_PARAMS, nil)
		return
	}

	// 不允许更新的字段
	delete(updateData, "id")
	delete(updateData, "open_id")
	delete(updateData, "lawyer_code")
	delete(updateData, "create_time")
	delete(updateData, "create_by")
	delete(updateData, "delete_time")
	delete(updateData, "delete_by")
	delete(updateData, "version")
	delete(updateData, "status")
	delete(updateData, "audit_status")

	// 添加更新信息
	updateData["update_by"] = currentUser.OpenID
	updateData["update_time"] = time.Now().Unix()

	if err := models.UpdateLawyer(lawyer.ID, updateData); err != nil {
		logging.Error("更新律师信息失败:", err)
		appG.Response(http.StatusInternalServerError, e.ERROR_EDIT_LAWYER_FAIL, nil)
		return
	}

	appG.Response(http.StatusOK, e.SUCCESS, nil)
}

// DeleteLawyer 删除律师信息
func DeleteLawyer(c *gin.Context) {
	appG := app.Gin{C: c}

	// 获取当前用户
	user, exists := c.Get("user")
	if !exists {
		appG.Response(http.StatusUnauthorized, e.ERROR_AUTH, nil)
		return
	}
	currentUser := user.(*models.User)

	// 获取要删除的律师信息
	lawyer, err := models.GetLawyerByOpenID(currentUser.OpenID)
	if err != nil {
		logging.Error("获取律师信息失败:", err)
		appG.Response(http.StatusInternalServerError, e.ERROR_GET_LAWYER_FAIL, nil)
		return
	}
	if lawyer == nil {
		appG.Response(http.StatusNotFound, e.ERROR_NOT_EXIST_LAWYER, nil)
		return
	}

	// 执行软删除
	if err := models.DeleteLawyer(lawyer.ID, currentUser.OpenID); err != nil {
		logging.Error("删除律师失败:", err)
		appG.Response(http.StatusInternalServerError, e.ERROR_DELETE_LAWYER_FAIL, nil)
		return
	}

	appG.Response(http.StatusOK, e.SUCCESS, nil)
}

// UpdateLikeCount 更新点赞数
func UpdateLikeCount(c *gin.Context) {
	appG := app.Gin{C: c}
	id := c.Query("id")

	lawyerID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		appG.Response(http.StatusBadRequest, e.INVALID_PARAMS, "无效的ID格式")
		return
	}

	// 检查律师是否存在
	exists, err := models.ExistLawyerByID(lawyerID)
	if err != nil {
		logging.Error("检查律师是否存在失败:", err)
		appG.Response(http.StatusInternalServerError, e.ERROR_CHECK_LAWYER_EXIST_FAIL, nil)
		return
	}
	if !exists {
		appG.Response(http.StatusNotFound, e.ERROR_NOT_EXIST_LAWYER, nil)
		return
	}

	// 更新点赞数
	if err := models.UpdateLikeCount(lawyerID); err != nil {
		logging.Error("更新律师点赞数失败:", err)
		appG.Response(http.StatusInternalServerError, e.ERROR_UPDATE_LAWYER_FAIL, nil)
		return
	}

	appG.Response(http.StatusOK, e.SUCCESS, nil)
}
