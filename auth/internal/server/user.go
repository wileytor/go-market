package server

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/lahnasti/go-market/auth/internal/server/responses"
	"github.com/lahnasti/go-market/common/models"
	"golang.org/x/crypto/bcrypt"
)

// GetUserProfileHandler получает профиль пользователя по ID
// @Summary Получение профиля пользователя
// @Description Возвращает профиль пользователя по указанному ID
// @Tags Пользователи
// @Produce json
// @Param id path int true "ID пользователя"
// @Success 200 {object} responses.Success
// @Failure 400 {object} responses.Error
// @Failure 500 {object} responses.Error
// @Router /users/{id} [get]
func (s *Server) GetUserProfileHandler(ctx *gin.Context) {
	id := ctx.Param("id")
	userID, err := strconv.Atoi(id)
	if err != nil || userID <= 0 {
		responses.SendError(ctx, http.StatusBadRequest, "Invalid user id", err)
		return
	}
	user, err := s.Db.GetUserProfile(userID)
	if err != nil {
		responses.SendError(ctx, http.StatusInternalServerError, "error", err)
		return
	}
	responses.SendSuccess(ctx, http.StatusOK, "User profile found", user)
}

// RegisterUserHandler регистрирует нового пользователя
// @Summary Регистрация нового пользователя
// @Description Регистрирует нового пользователя с предоставленными данными
// @Tags Пользователи
// @Accept json
// @Produce json
// @Param user body models.User true "User information"
// @Success 201 {object} responses.Success
// @Failure 400 {object} responses.Error
// @Failure 409 {object} responses.Error
// @Failure 500 {object} responses.Error
// @Router /users/register [post]
func (s *Server) RegisterUserHandler(ctx *gin.Context) {
	var user models.User
	if err := ctx.ShouldBindJSON(&user); err != nil {
		responses.SendError(ctx, http.StatusBadRequest, "Invalid request data", err)
		return
	}
	if err := s.Valid.Struct(user); err != nil {
		responses.SendError(ctx, http.StatusBadRequest, "Invalid user data", err)
		return
	}

	if !isValidUsername(user.Username) {
		responses.SendError(ctx, http.StatusBadRequest, "Not a valid username", nil)
		return
	}

	if !isValidPass(user.Password) {
		responses.SendError(ctx, http.StatusBadRequest, "Password must contain at least 8 characters", nil)
		return
	}

	exists, err := s.Db.IsUsernameUnique(user.Username)
	if err != nil {
		responses.SendError(ctx, http.StatusInternalServerError, "error", err)
		return
	}
	if !exists {
		responses.SendError(ctx, http.StatusBadRequest, "Username already exists", nil)
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		s.log.Error().Err(err).Msg("Error generating password hash")
		responses.SendError(ctx, http.StatusInternalServerError, "Failed to hash password", err)
		return
	}
	user.Password = string(hash)
	id, err := s.Db.RegisterUser(user)
	if err != nil {
		responses.SendError(ctx, http.StatusInternalServerError, "Failed to register user", err)
		return
	}
	responses.SendSuccess(ctx, http.StatusCreated, "User registered successfully", id)

}

// LoginUserHandler обрабатывает вход пользователя
// @Summary Вход пользователя
// @Description Выполняет вход пользователя с указанными учетными данными
// @Tags Пользователи
// @Accept json
// @Produce json
// @Param credentials body models.Credentials true "Учетные данные пользователя"
// @Success 200 {object} responses.Success
// @Failure 400 {object} responses.Error
// @Failure 401 {object} responses.Error
// @Failure 500 {object} responses.Error
// @Router /users/login [post]
func (s *Server) LoginUserHandler(ctx *gin.Context) {
	var credentials models.Credentials
	if err := ctx.ShouldBindJSON(&credentials); err != nil {
		responses.SendError(ctx, http.StatusBadRequest, "Invalid request data", err)
		return
	}
	if err := s.Valid.Struct(credentials); err != nil {
		responses.SendError(ctx, http.StatusBadRequest, "Not a valid user", err)
		return
	}
	userID, err := s.Db.LoginUser(credentials.Username, credentials.Password)
	if err != nil {
		if err.Error() == "user not found" {
			responses.SendError(ctx, http.StatusUnauthorized, "Invalid username or password", err)
		} else if err.Error() == "invalid password" {
			responses.SendError(ctx, http.StatusUnauthorized, "Invalid password", err)
		} else {
			responses.SendError(ctx, http.StatusInternalServerError, "error", err)
		}
		return
	}
	token, err := CreateJWTToken(userID)
	if err != nil {
		responses.SendError(ctx, http.StatusInternalServerError, "error", err)
		return
	}
	ctx.Header("Authorization", token)
	responses.SendSuccess(ctx, http.StatusOK, "User was login successfully", token)
}
