package middleware

import (
	"context"

	"github.com/gin-gonic/gin"
)

const (
	LanguageKey = "language"
)

func LanguageMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		lang := c.Request.Header.Get("Accept-Language") // Get language from header
		if lang == "" {
			lang = "en" // default language
		}

		// Set language in context
		ctx := context.WithValue(c.Request.Context(), LanguageKey, lang)
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}
