package controllers

import (
	"github.com/cloudreve/Cloudreve/v4/application/dependency"
	"github.com/cloudreve/Cloudreve/v4/ent"
	"github.com/cloudreve/Cloudreve/v4/inventory"
	"github.com/cloudreve/Cloudreve/v4/pkg/hashid"
	"github.com/cloudreve/Cloudreve/v4/pkg/serializer"
	"github.com/cloudreve/Cloudreve/v4/pkg/util"
	"github.com/cloudreve/Cloudreve/v4/service/share"
	"github.com/cloudreve/Cloudreve/v4/service/user"
	"github.com/gin-gonic/gin"
	"github.com/samber/lo"
)

// StartLoginAuthn 开始注册WebAuthn登录
func StartLoginAuthn(c *gin.Context) {
	res, err := user.PreparePasskeyLogin(c)
	if err != nil {
		c.JSON(200, serializer.Err(c, err))
		return
	}

	c.JSON(200, serializer.Response{Data: res})
}

// FinishLoginAuthn 完成注册WebAuthn登录
func FinishLoginAuthn(c *gin.Context) {
	service := ParametersFromContext[*user.FinishPasskeyLoginService](c, user.FinishPasskeyLoginParameterCtx{})
	u, err := service.FinishPasskeyLogin(c)
	if err != nil {
		c.JSON(200, serializer.Err(c, err))
		c.Abort()
		return
	}

	util.WithValue(c, inventory.UserCtx{}, u)
}

// StartRegAuthn 开始注册WebAuthn信息
func StartRegAuthn(c *gin.Context) {
	res, err := user.PreparePasskeyRegister(c)
	if err != nil {
		c.JSON(200, serializer.Err(c, err))
		return
	}

	c.JSON(200, serializer.Response{Data: res})
}

// FinishRegAuthn 完成注册WebAuthn信息
func FinishRegAuthn(c *gin.Context) {
	service := ParametersFromContext[*user.FinishPasskeyRegisterService](c, user.FinishPasskeyRegisterParameterCtx{})
	res, err := service.FinishPasskeyRegister(c)
	if err != nil {
		c.JSON(200, serializer.Err(c, err))
		return
	}

	c.JSON(200, serializer.Response{Data: res})
}

// UserDeletePasskey deletes user passkey
func UserDeletePasskey(c *gin.Context) {
	service := ParametersFromContext[*user.DeletePasskeyService](c, user.DeletePasskeyParameterCtx{})
	err := service.DeletePasskey(c)
	if err != nil {
		c.JSON(200, serializer.Err(c, err))
		return
	}

	c.JSON(200, serializer.Response{})
}

// UserLoginValidation validates user login request
func UserLoginValidation(c *gin.Context) {
	service := ParametersFromContext[*user.UserLoginService](c, user.LoginParameterCtx{})
	expectedUser, twoFaSession, err := service.Login(c)
	if err != nil {
		c.JSON(200, serializer.Err(c, err))
		c.Abort()
		return
	}

	if twoFaSession == "" {
		// No 2FA required, proceed
		util.WithValue(c, inventory.UserCtx{}, expectedUser)
		c.Next()
		return
	}

	c.JSON(200, serializer.Response{Code: serializer.CodeNotFullySuccess, Data: twoFaSession})
	c.Abort()
}

// UserLogin2FAValidation validates user OTP code
func UserLogin2FAValidation(c *gin.Context) {
	service := ParametersFromContext[*user.OtpValidationService](c, user.OtpValidationParameterCtx{})
	expectedUser, err := service.Verify2FA(c)
	if err != nil {
		c.JSON(200, serializer.Err(c, err))
		c.Abort()
		return
	}

	util.WithValue(c, inventory.UserCtx{}, expectedUser)
	c.Next()
}

// UserIssueToken generates new token pair for user
func UserIssueToken(c *gin.Context) {
	resp, err := user.IssueToken(c)
	if err != nil {
		c.JSON(200, serializer.Err(c, err))
		c.Abort()
		return
	}

	c.JSON(200, serializer.Response{
		Data: resp,
	})
}

// UserRefreshToken refreshes token pair for user
func UserRefreshToken(c *gin.Context) {
	service := ParametersFromContext[*user.RefreshTokenService](c, user.RefreshTokenParameterCtx{})
	resp, err := service.Refresh(c)
	if err != nil {
		c.JSON(200, serializer.Err(c, err))
		c.Abort()
		return
	}

	c.JSON(200, serializer.Response{
		Data: resp,
	})
}

// UserRegister 用户注册
func UserRegister(c *gin.Context) {
	service := ParametersFromContext[*user.UserRegisterService](c, user.RegisterParameterCtx{})
	c.JSON(200, service.Register(c))
}

// UserSendReset 发送密码重设邮件
func UserSendReset(c *gin.Context) {
	service := ParametersFromContext[*user.UserResetEmailService](c, user.UserResetEmailParameterCtx{})
	if err := service.Reset(c); err != nil {
		c.JSON(200, serializer.Err(c, err))
		c.Abort()
		return
	}
	c.JSON(200, serializer.Response{})
}

// UserReset 重设密码
func UserReset(c *gin.Context) {
	service := ParametersFromContext[*user.UserResetService](c, user.UserResetParameterCtx{})
	res, err := service.Reset(c)
	if err != nil {
		c.JSON(200, serializer.Err(c, err))
		c.Abort()
		return
	}
	c.JSON(200, serializer.Response{Data: res})
}

// UserActivate 用户激活
func UserActivate(c *gin.Context) {
	c.JSON(200, user.ActivateUser(c))
}

// UserSignOut 用户退出登录
func UserSignOut(c *gin.Context) {
	service := ParametersFromContext[*user.RefreshTokenService](c, user.RefreshTokenParameterCtx{})
	res, err := service.Delete(c)
	if err != nil {
		c.JSON(200, serializer.Err(c, err))
		c.Abort()
		return
	}

	c.JSON(200, serializer.Response{
		Data: res,
	})
}

// UserMe 获取当前登录的用户
func UserMe(c *gin.Context) {
	dep := dependency.FromContext(c)
	c.JSON(200, serializer.Response{
		Data: user.BuildUser(inventory.UserFromContext(c), dep.HashIDEncoder()),
	})
}

// UserGet 获取用户信息
func UserGet(c *gin.Context) {
	u, err := user.GetUser(c)
	if err != nil {
		c.JSON(200, serializer.Err(c, err))
		c.Abort()
		return
	}

	isAnonymous := inventory.IsAnonymousUser(inventory.UserFromContext(c))
	redactLevel := user.RedactLevelUser
	if isAnonymous {
		redactLevel = user.RedactLevelAnonymous
	}
	c.JSON(200, serializer.Response{
		Data: user.BuildUserRedacted(u, redactLevel, dependency.FromContext(c).HashIDEncoder()),
	})
}

// UserStorage 获取用户的存储信息
func UserStorage(c *gin.Context) {
	res, err := user.GetUserCapacity(c)
	if err != nil {
		c.JSON(200, serializer.Err(c, err))
		c.Abort()
		return
	}

	c.JSON(200, serializer.Response{
		Data: res,
	})
}

// UserSetting 获取用户设定
func UserSetting(c *gin.Context) {
	res, err := user.GetUserSettings(c)
	if err != nil {
		c.JSON(200, serializer.Err(c, err))
		c.Abort()
		return
	}

	c.JSON(200, serializer.Response{
		Data: res,
	})
}

// UploadAvatar 从文件上传头像
func UploadAvatar(c *gin.Context) {
	if err := user.UpdateUserAvatar(c); err != nil {
		c.JSON(200, serializer.Err(c, err))
		return
	}

	c.JSON(200, serializer.Response{})
}

// GetUserAvatar 获取用户头像
func GetUserAvatar(c *gin.Context) {
	service := ParametersFromContext[*user.GetAvatarService](c, user.GetAvatarServiceParamsCtx{})
	err := service.Get(c)
	if err != nil {
		c.JSON(200, serializer.Err(c, err))
		c.Abort()
		return
	}
}

// UpdateOption 更改用户设定
func UpdateOption(c *gin.Context) {
	service := ParametersFromContext[*user.PatchUserSetting](c, user.PatchUserSettingParamsCtx{})
	err := service.Patch(c)
	if err != nil {
		c.JSON(200, serializer.Err(c, err))
		c.Abort()
		return
	}

	c.JSON(200, serializer.Response{})

	//var service user.SettingUpdateService
	//if err := c.ShouldBindUri(&service); err == nil {
	//	var (
	//		subService user.OptionsChangeHandler
	//		subErr     error
	//	)
	//
	//	switch service.Option {
	//	case "nick":
	//		subService = &user.ChangerNick{}
	//	case "vip":
	//		subService = &user.VIPUnsubscribe{}
	//	case "qq":
	//		subService = &user.QQBind{}
	//	case "policy":
	//		subService = &user.PolicyChange{}
	//	case "homepage":
	//		subService = &user.HomePage{}
	//	case "password":
	//		subService = &user.PasswordChange{}
	//	case "2fa":
	//		subService = &user.Enable2FA{}
	//	case "authn":
	//		subService = &user.DeleteWebAuthn{}
	//	case "theme":
	//		subService = &user.ThemeChose{}
	//	default:
	//		subService = &user.ChangerNick{}
	//	}
	//
	//	subErr = c.ShouldBindJSON(subService)
	//	if subErr != nil {
	//		c.JSON(200, ErrorResponse(subErr))
	//		return
	//	}
	//
	//	res := subService.Update(c, CurrentUser(c))
	//	c.JSON(200, res)
	//
	//} else {
	//	c.JSON(200, ErrorResponse(err))
	//}
}

// UserInit2FA 初始化二步验证
func UserInit2FA(c *gin.Context) {
	secret, err := user.Init2FA(c)
	if err != nil {
		c.JSON(200, serializer.Err(c, err))
		c.Abort()
		return
	}

	c.JSON(200, serializer.Response{
		Data: secret,
	})
}

// UserPerformCopySession copy to create new session or refresh current session
func UserPerformCopySession(c *gin.Context) {
	//var service user.CopySessionService
	//if err := c.ShouldBindUri(&service); err == nil {
	//	res := service.Copy(c)
	//	c.JSON(200, res)
	//} else {
	//	c.JSON(200, ErrorResponse(err))
	//}
}

// UserPrepareLogin validates precondition for login
func UserPrepareLogin(c *gin.Context) {
	service := ParametersFromContext[*user.PrepareLoginService](c, user.PrepareLoginParameterCtx{})
	res, err := service.Prepare(c)
	if err != nil {
		c.JSON(200, serializer.Err(c, err))
		c.Abort()
		return
	}

	c.JSON(200, serializer.Response{Data: res})
}

// UserSearch Search user by keyword
func UserSearch(c *gin.Context) {
	service := ParametersFromContext[*user.SearchUserService](c, user.SearchUserParamCtx{})
	u, err := service.Search(c)
	if err != nil {
		c.JSON(200, serializer.Err(c, err))
		c.Abort()
		return
	}

	hasher := dependency.FromContext(c).HashIDEncoder()
	c.JSON(200, serializer.Response{
		Data: lo.Map(u, func(item *ent.User, index int) user.User {
			return user.BuildUserRedacted(item, user.RedactLevelUser, hasher)
		}),
	})
}

// ListPublicShare lists all public shares for given user
func ListPublicShare(c *gin.Context) {
	service := ParametersFromContext[*share.ListShareService](c, share.ListShareParamCtx{})
	resp, err := service.ListInUserProfile(c, hashid.FromContext(c))
	if err != nil {
		c.JSON(200, serializer.Err(c, err))
		c.Abort()
		return
	}

	if resp != nil {
		c.JSON(200, serializer.Response{
			Data: resp,
		})
	}
}
