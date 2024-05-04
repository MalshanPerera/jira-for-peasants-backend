package middlewares

import (
	errpkg "jira-for-peasants/errors"
	"jira-for-peasants/services"
	"jira-for-peasants/utils"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

func IsAuthenticated(sessionService *services.SessionService) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Check the Authorization header
			headerStr := c.Request().Header.Get("Authorization")

			// If the token is not in the Authorization header, check the cookies
			if headerStr == "" {
				cookie, err := c.Cookie("token")
				if err != nil {
					return errpkg.NewApiError(http.StatusUnauthorized, "Unauthorized")
				}
				headerStr = cookie.Value
			}

			// Parse and validate the token
			tokenString := strings.TrimPrefix(headerStr, "Bearer ")
			token, err := utils.VerifyToken(tokenString)
			if err != nil {
				return errpkg.BadRequest("Invalid token")
			}

			// Extract the user information from the token
			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				return errpkg.BadRequest("Invalid token claims")
			}

			userId, ok := claims["user_id"].(string)
			if !ok {
				return errpkg.BadRequest("Invalid user id")
			}

			// Check if the session exists
			_, err = sessionService.ValidateUserSession(c.Request().Context(), userId)
			if err != nil {
				return err
			}

			// Check if the token is expired
			if float64(time.Now().Unix()) < claims["expired_at"].(float64) {
				return errpkg.NewApiError(http.StatusUnauthorized, "Token expired")
			}

			// Add the user information to the context
			utils.SetUser(c, userId)

			// If the user is authenticated, call the next handler with the new context
			return next(c)
		}
	}
}
