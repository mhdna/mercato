package api

import (
	"database/sql"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
	db "github.com/mhdna/kashi/db/sqlc"
	"github.com/mhdna/kashi/token"
	"github.com/mhdna/kashi/util"
)

// PIN codes are the only credential in this POS -- 4 to 6 digits, no email.
const (
	pinMinLen = 4
	pinMaxLen = 6
)

var errPINFormat = errors.New("pin must be 4 to 6 digits")

func isPIN(s string) bool {
	if len(s) < pinMinLen || len(s) > pinMaxLen {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

type createUserRequest struct {
	Name      string `json:"name" binding:"required,alphanum"`
	Password  string `json:"password" binding:"required"`
	Activated bool   `json:"activated"`
}

type userResponse struct {
	ID                int64     `json:"id"`
	Name              string    `json:"name"`
	Activated         bool      `json:"activated"`
	PasswordChangedAt time.Time `json:"password_changed_at"`
	CreatedAt         time.Time `json:"created_at"`
}

func newUserResponse(user db.User) userResponse {
	return userResponse{
		ID:                user.ID,
		Name:              user.Name,
		Activated:         user.Activated,
		PasswordChangedAt: user.PasswordChangedAt,
		CreatedAt:         user.CreatedAt,
	}
}

func (server *Server) createUser(ctx *gin.Context) {
	var req createUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	if !isPIN(req.Password) {
		server.writeError(ctx, http.StatusBadRequest, errPINFormat)
		return
	}

	hashPassword, err := util.HashPassword(req.Password)
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	arg := db.CreateUserParams{
		Name: req.Name,
		// email is unused in this POS but the column is NOT NULL and unique;
		// fill it with a throwaway value so it never has to be entered.
		Email:        util.RandomEmail(),
		PasswordHash: hashPassword,
		Activated:    req.Activated,
	}

	user, err := server.store.CreateUser(ctx, arg)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code.Name() {
			case "foreign_key_violation", "unique_violation":
				server.writeError(ctx, http.StatusForbidden, err)
			}
		}
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}
	rsp := newUserResponse(user)

	ctx.JSON(http.StatusOK, rsp)
}

type getUserRequest struct {
	ID int64 `uri:"id" binding:"required,min=1"`
}

func (server *Server) getUser(ctx *gin.Context) {
	var req getUserRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	user, err := server.store.GetUser(ctx, req.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			server.writeError(ctx, http.StatusNotFound, err)
			return
		}

		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, newUserResponse(user))
}

type listUsersRequest struct {
	PageSize int32 `form:"page_size,default=10" binding:"min=5,max=10"`
	PageID   int32 `form:"page_id,default=0" binding:"min=0"`
}

func (server *Server) listUsers(ctx *gin.Context) {
	var req listUsersRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	arg := db.ListUsersParams{
		Limit:  req.PageSize,
		Offset: req.PageID,
	}
	users, err := server.store.ListUsers(ctx, arg)
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	rsp := make([]userResponse, len(users))
	for i, u := range users {
		rsp[i] = newUserResponse(u)
	}
	ctx.JSON(http.StatusOK, rsp)
}

type updateUserRequest struct {
	ID        int64  `json:"id" binding:"required,min=1"`
	Name      string `json:"name" binding:"required,alphanum"`
	Password  string `json:"password"` // optional -- blank keeps the current PIN
	Activated bool   `json:"activated"`
}

func (server *Server) updateUser(ctx *gin.Context) {
	var req updateUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	current, err := server.store.GetUser(ctx, req.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			server.writeError(ctx, http.StatusNotFound, err)
			return
		}
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	// Keep the existing hash unless a new PIN was supplied.
	hashPassword := current.PasswordHash
	if req.Password != "" {
		if !isPIN(req.Password) {
			server.writeError(ctx, http.StatusBadRequest, errPINFormat)
			return
		}
		hashPassword, err = util.HashPassword(req.Password)
		if err != nil {
			server.writeError(ctx, http.StatusInternalServerError, err)
			return
		}
	}

	arg := db.UpdateUserParams{
		ID:           req.ID,
		Name:         req.Name,
		Email:        current.Email,
		PasswordHash: hashPassword,
		Activated:    req.Activated,
	}

	err = server.store.UpdateUser(ctx, arg)
	if err != nil {
		if err == sql.ErrNoRows {
			server.writeError(ctx, http.StatusBadRequest, err)
			return
		}
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, userResponse{
		ID:                req.ID,
		Name:              req.Name,
		Activated:         req.Activated,
		PasswordChangedAt: current.PasswordChangedAt,
		CreatedAt:         current.CreatedAt,
	})
}

type deleteUserRequest struct {
	ID int64 `uri:"id" binding:"required,min=1"`
}

func (server *Server) deleteUser(ctx *gin.Context) {
	var req deleteUserRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	err := server.store.DeleteUser(ctx, req.ID)
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "user deleted"})
}

type loginUserRequest struct {
	Name     string `json:"name" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type loginUserResponse struct {
	SessionID             string       `json:"session_id"`
	AccessToken           string       `json:"access_token"`
	AccessTokenExpiresAt  time.Time    `json:"access_token_expires_at"`
	RefreshToken          string       `json:"refresh_token"`
	RefreshTokenExpiresAt time.Time    `json:"refresh_token_expires_at"`
	User                  userResponse `json:"user"`
}

func (server *Server) loginUser(ctx *gin.Context) {
	var req loginUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}
	user, err := server.store.GetUserByUsername(ctx, req.Name)
	if err != nil {
		if err == sql.ErrNoRows {
			server.writeError(ctx, http.StatusNotFound, err)
			return
		}
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}
	err = util.CheckPassword(req.Password, user.PasswordHash)
	if err != nil {
		server.writeError(ctx, http.StatusUnauthorized, err)
		return
	}
	accessToken, accessPayload, err := server.tokenMaker.CreateToken(
		user.Name,
		server.config.AccessTokenDuration,
	)
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}
	refreshToken, refreshPayload, err := server.tokenMaker.CreateToken(
		user.Name,
		server.config.RefreshTokenDuration,
	)
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	session, err := server.store.CreateSession(ctx, db.CreateSessionParams{
		ID:           refreshPayload.ID,
		Username:     user.Name,
		RefreshToken: refreshToken,
		UserAgent:    ctx.Request.UserAgent(),
		ClientIp:     ctx.ClientIP(),
		IsBlocked:    false,
		ExpiresAt:    refreshPayload.ExpiredAt,
	})
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	rsp := loginUserResponse{
		SessionID:             session.ID.String(),
		AccessToken:           accessToken,
		AccessTokenExpiresAt:  accessPayload.ExpiredAt,
		RefreshToken:          refreshToken,
		RefreshTokenExpiresAt: refreshPayload.ExpiredAt,
		User:                  newUserResponse(user),
	}

	ctx.JSON(http.StatusOK, rsp)
}

type renewAccessTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type renewAccessTokenResponse struct {
	AccessToken          string    `json:"access_token"`
	AccessTokenExpiresAt time.Time `json:"access_token_expires_at"`
}

func (server *Server) renewAccessToken(ctx *gin.Context) {
	var req renewAccessTokenRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	refreshPayload, err := server.tokenMaker.VerifyToken(req.RefreshToken)
	if err != nil {
		server.writeError(ctx, http.StatusUnauthorized, err)
		return
	}

	session, err := server.store.GetSession(ctx, refreshPayload.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			server.writeError(ctx, http.StatusNotFound, err)
			return
		}
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	if session.IsBlocked {
		server.writeError(ctx, http.StatusUnauthorized, token.ErrInvalidToken)
		return
	}

	if session.Username != refreshPayload.Username {
		server.writeError(ctx, http.StatusUnauthorized, token.ErrInvalidToken)
		return
	}

	if session.RefreshToken != req.RefreshToken {
		server.writeError(ctx, http.StatusUnauthorized, token.ErrInvalidToken)
		return
	}

	accessToken, accessPayload, err := server.tokenMaker.CreateToken(
		refreshPayload.Username,
		server.config.AccessTokenDuration,
	)
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	rsp := renewAccessTokenResponse{
		AccessToken:          accessToken,
		AccessTokenExpiresAt: accessPayload.ExpiredAt,
	}

	ctx.JSON(http.StatusOK, rsp)
}
