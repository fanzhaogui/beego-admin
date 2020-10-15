package services

import (
	"beego-admin/form_validate"
	"beego-admin/global"
	"beego-admin/models"
	"beego-admin/utils"
	beego_pagination "beego-admin/utils/beego-pagination"
	"encoding/base64"
	"errors"
	"github.com/astaxie/beego/context"
	"github.com/astaxie/beego/orm"
	"net/url"
	"strconv"
)

type AdminUserService struct {
	BaseService
}

//根据id获取一条admin_user数据
func (*AdminUserService) GetAdminUserById(id int) *models.AdminUser {
	o := orm.NewOrm()
	adminUser := models.AdminUser{Id: id}
	err := o.Read(&adminUser)
	if err != nil {
		return nil
	}
	return &adminUser
}

//权限检测
func (*AdminUserService) AuthCheck(url string, authExcept map[string]interface{}, loginUser *models.AdminUser) bool {
	authUrl := loginUser.GetAuthUrl()
	if utils.KeyInMap(url, authExcept) || utils.KeyInMap(url, authUrl) {
		return true
	} else {
		return false
	}
}

//用户登录验证
func (*AdminUserService) CheckLogin(loginForm form_validate.LoginForm, ctx *context.Context) (*models.AdminUser, error) {
	var adminUser models.AdminUser
	o := orm.NewOrm()
	err := o.QueryTable(new(models.AdminUser)).Filter("username", loginForm.Username).Limit(1).One(&adminUser)
	if err != nil {
		return nil, errors.New("用户不存在")
	}

	decodePasswdStr, err := base64.StdEncoding.DecodeString(adminUser.Password)

	if err != nil || !utils.PasswordVerify(loginForm.Password, string(decodePasswdStr)) {
		return nil, errors.New("密码错误.")
	}

	if adminUser.Status != 1 {
		return nil, errors.New("用户被冻结.")
	}

	ctx.Output.Session(global.LOGIN_USER, adminUser)

	if loginForm.Remember != "" {
		ctx.SetCookie(global.LOGIN_USER_ID, strconv.Itoa(adminUser.Id), 7200)
		ctx.SetCookie(global.LOGIN_USER_ID_SIGN, adminUser.GetSignStrByAdminUser(ctx), 7200)
	} else {
		ctx.SetCookie(global.LOGIN_USER_ID, ctx.GetCookie(global.LOGIN_USER_ID), -1)
		ctx.SetCookie(global.LOGIN_USER_ID_SIGN, ctx.GetCookie(global.LOGIN_USER_ID_SIGN), -1)
	}

	return &adminUser, nil

}

//获取admin_user 总数
func (*AdminUserService) GetCount() int {
	count, err := orm.NewOrm().QueryTable(new(models.AdminUser)).Count()
	if err != nil {
		return 0
	}
	return int(count)
}

//获取所有adminuser
func (*AdminUserService) GetAllAdminUser() []*models.AdminUser {
	var adminUser []*models.AdminUser
	o := orm.NewOrm().QueryTable(new(models.AdminUser))
	_, err := o.All(&adminUser)
	if err != nil {
		return nil
	} else {
		return adminUser
	}
}

//系统管理-个人资料-修改昵称
func (*AdminUserService) UpdateNickName(id int, nickname string) int {
	num, err := orm.NewOrm().QueryTable(new(models.AdminUser)).Filter("id", id).Update(orm.Params{
		"nickname": nickname,
	})
	if err != nil || num <= 0 {
		return 0
	}
	return int(num)
}

func (*AdminUserService) UpdatePassword(id int, newPassword string) int {
	newPasswordForHash, err := utils.PasswordHash(newPassword)

	if err != nil {
		return 0
	}

	num, err := orm.NewOrm().QueryTable(new(models.AdminUser)).Filter("id", id).Update(orm.Params{
		"password": base64.StdEncoding.EncodeToString([]byte(newPasswordForHash)),
	})

	if err != nil || num <= 0 {
		return 0
	}

	return int(num)
}

//系统管理-个人资料-修改头像
func (*AdminUserService) UpdateAvatar(id int, avatar string) int {
	num, err := orm.NewOrm().QueryTable(new(models.AdminUser)).Filter("id", id).Update(orm.Params{
		"avatar": avatar,
	})
	if err != nil || num <= 0 {
		return 0
	}
	return int(num)
}

//通过分页获取adminuser
func (this *AdminUserService) GetPaginateData(listRows int, params url.Values) ([]*models.AdminUser, beego_pagination.Pagination) {
	//搜索、查询字段赋值
	this.SearchField = append(this.SearchField, new(models.AdminUser).SearchField()...)

	var adminUser []*models.AdminUser
	o := orm.NewOrm().QueryTable(new(models.AdminUser))
	_, err := this.PaginateAndScopeWhere(o, listRows, params).All(&adminUser)
	if err != nil {
		return nil, this.Pagination
	} else {
		return adminUser, this.Pagination
	}
}