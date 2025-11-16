package middleware

import (
	"github.com/gofiber/fiber/v2/middleware/cors"
)

// CORSConfig holds CORS configuration
type CORSConfig struct {
	AllowOrigins     []string
	AllowMethods     []string
	AllowHeaders     []string
	ExposeHeaders    []string
	AllowCredentials bool
	MaxAge           int
}

// CORS returns a middleware that handles CORS requests
func CORS(config CORSConfig) Handler {
	corsConfig := cors.Config{
		AllowOrigins:     "*",
		AllowMethods:     "GET,POST,HEAD,PUT,DELETE,PATCH",
		AllowHeaders:     "*",
		AllowCredentials: false,
		MaxAge:           0,
	}

	if len(config.AllowOrigins) > 0 {
		corsConfig.AllowOrigins = config.AllowOrigins[0]
		if len(config.AllowOrigins) > 1 {
			corsConfig.AllowOrigins = ""
			corsConfig.AllowOriginsFunc = func(origin string) bool {
				for _, allowed := range config.AllowOrigins {
					if allowed == "*" || allowed == origin {
						return true
					}
				}
				return false
			}
		}
	}

	if len(config.AllowMethods) > 0 {
		methods := ""
		for i, method := range config.AllowMethods {
			if i > 0 {
				methods += ","
			}
			methods += method
		}
		corsConfig.AllowMethods = methods
	}

	if len(config.AllowHeaders) > 0 {
		headers := ""
		for i, header := range config.AllowHeaders {
			if i > 0 {
				headers += ","
			}
			headers += header
		}
		corsConfig.AllowHeaders = headers
	}

	if len(config.ExposeHeaders) > 0 {
		headers := ""
		for i, header := range config.ExposeHeaders {
			if i > 0 {
				headers += ","
			}
			headers += header
		}
		corsConfig.ExposeHeaders = headers
	}

	if config.AllowCredentials {
		corsConfig.AllowCredentials = true
	}

	if config.MaxAge > 0 {
		corsConfig.MaxAge = config.MaxAge
	}

	corsMiddleware := cors.New(corsConfig)
	return ToFiber(corsMiddleware)
}
