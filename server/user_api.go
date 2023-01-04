package server

import (
	"encoding/json"
	"fmt"
	"sao-datastore-storage/model"
	"sao-datastore-storage/util"
	"sao-datastore-storage/util/api"
	"strconv"

	"github.com/gin-gonic/gin"
)

func (s *Server) UpdateUserProfile(ctx *gin.Context) {
	ethAddress, _ := ctx.Get("User")
	if ethAddress.(string) == "" {
		api.Unauthorized(ctx, "invalid.signature", "invalid signature")
		return
	}
	ethAddr := ethAddress.(string)

	var profileToUpdate model.UserProfile
	decoder := json.NewDecoder(ctx.Request.Body)
	err := decoder.Decode(&profileToUpdate)
	if err != nil {
		log.Error("update user profile request body invalid: ", err)
		api.BadRequest(ctx, "invalid.body", "update user profile request body invalid")
		return
	}

	profile, err := s.Model.UpsertUserProfile(ethAddr, profileToUpdate)
	if err != nil {
		log.Error("update user failed:", err)
		api.ServerError(ctx, "error.update.user", "database error")
		return
	}
	if profile.Username == "" {
		defaultUsername := fmt.Sprintf("%s_%s", "Storverse", ethAddr[len(ethAddr)-4:])
		err = s.Model.UpdateUsername(ethAddr, defaultUsername)
		if err != nil {
			log.Error(err)
			api.ServerError(ctx, "error.update.username", "failed to assign username.")
			return
		}
	}
	api.Success(ctx, profile)
}

func (s *Server) GetUserProfile(ctx *gin.Context) {
	ethAddress := ctx.GetHeader("address")
	util.VerifySignature(ctx)
	owner, _ := ctx.Get("User")
	if ethAddress != "" && owner.(string) == "" {
		api.Unauthorized(ctx, "invalid.signature", "invalid signature")
		return
	}

	userAddress,got := ctx.GetQuery("address")
	if !got {
		userAddress = ethAddress
	}
	if userAddress == "" {
		api.BadRequest(ctx, "invalid.param", "")
		return
	}
	profile, err := s.Model.GetUserProfile(userAddress, ethAddress)
	if err != nil {
		log.Error(err)
		api.ServerError(ctx, "error.get.userprofile", "database error")
		return
	}
	api.Success(ctx, profile)
}

func (s *Server) FollowUser(ctx *gin.Context) {
	ethAddress, _ := ctx.Get("User")
	if ethAddress.(string) == "" {
		api.Unauthorized(ctx, "invalid.signature", "invalid signature")
		return
	}
	following := ctx.Param("address")
	err := s.Model.FollowUser(ethAddress.(string), following)
	if err != nil {
		log.Error(err)
		api.ServerError(ctx, "error.followuser", "database error")
		return
	}
	api.Success(ctx, true)
}

func (s *Server) GetUserFollowers(ctx *gin.Context) {
	ethAddress := ctx.GetHeader("address")
	util.VerifySignature(ctx)
	owner, _ := ctx.Get("User")
	if ethAddress != "" && owner.(string) == "" {
		api.Unauthorized(ctx, "invalid.signature", "invalid signature")
		return
	}

	userAddress,got := ctx.GetQuery("address")
	if !got {
		userAddress = ethAddress
	}
	if userAddress == "" {
		api.BadRequest(ctx, "invalid.param", "")
		return
	}
	followers, err := s.Model.GetUserFollowers(userAddress)
	if err != nil {
		log.Error(err)
		api.ServerError(ctx, "error.get.userfollowers", "database error")
		return
	}
	api.Success(ctx, followers)
}

func (s *Server) GetUserFollowings(ctx *gin.Context) {
	ethAddress := ctx.GetHeader("address")
	util.VerifySignature(ctx)
	owner, _ := ctx.Get("User")
	if ethAddress != "" && owner.(string) == "" {
		api.Unauthorized(ctx, "invalid.signature", "invalid signature")
		return
	}

	userAddress,got := ctx.GetQuery("address")
	if !got {
		userAddress = ethAddress
	}
	if userAddress == "" {
		api.BadRequest(ctx, "invalid.param", "")
		return
	}

	followings, err := s.Model.GetUserFollowings(userAddress)
	if err != nil {
		log.Error(err)
		api.ServerError(ctx, "error.get.userfollowers", "database error")
		return
	}
	api.Success(ctx, followings)
}

func (s *Server) GetUserDashboard(ctx *gin.Context) {
	ethAddress, _ := ctx.Get("User")
	if ethAddress.(string) == "" {
		api.Unauthorized(ctx, "invalid.signature", "invalid signature")
		return
	}

	offset, got := ctx.GetQuery("offset")
	if !got {
		offset = "0"
	}
	o, err := strconv.Atoi(offset)
	if err != nil {
		log.Info(err)
		o = 0
	}
	limit, got := ctx.GetQuery("limit")
	if !got {
		limit = "10"
	}
	l, err := strconv.Atoi(limit)
	if err != nil {
		log.Info(err)
		l = 10
	}

	dashboard, err := s.Model.GetUserDashboard(l, o, ethAddress.(string), func(preview string) string {
		return fmt.Sprintf("%s/previews/%s", s.Config.Host, preview)
	})
	if err != nil {
		log.Error(err)
		api.ServerError(ctx, "error.get.userdashboard", "database error")
		return
	}
	api.Success(ctx, dashboard)
}

func (s *Server) GetUserPurchases(ctx *gin.Context) {
	ethAddress, _ := ctx.Get("User")
	if ethAddress.(string) == "" {
		api.Unauthorized(ctx, "invalid.signature", "invalid signature")
		return
	}

	offset, got := ctx.GetQuery("offset")
	if !got {
		offset = "0"
	}
	o, err := strconv.Atoi(offset)
	if err != nil {
		log.Info(err)
		o = 0
	}
	limit, got := ctx.GetQuery("limit")
	if !got {
		limit = "10"
	}
	l, err := strconv.Atoi(limit)
	if err != nil {
		log.Info(err)
		l = 10
	}

	dashboard, err := s.Model.GetUserPurchases(l, o, ethAddress.(string), func(preview string) string {
		return fmt.Sprintf("%s/previews/%s", s.Config.Host, preview)
	})
	if err != nil {
		log.Error(err)
		api.ServerError(ctx, "error.get.userpurchases", "database error")
		return
	}
	api.Success(ctx, dashboard)
}

func (s *Server) GetUserSummary(ctx *gin.Context) {
	ethAddress, _ := ctx.Get("User")
	if ethAddress.(string) == "" {
		api.Unauthorized(ctx, "invalid.signature", "invalid signature")
		return
	}

	summary, err := s.Model.GetUserSummary(ethAddress.(string))
	if err != nil {
		log.Error(err)
		api.ServerError(ctx, "error.get.usersummary", "database error")
		return
	}
	api.Success(ctx, summary)
}
