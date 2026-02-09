package middleware

import (
	"context"
	"log"
	"strings"
	"time"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	"github.com/gin-gonic/gin"
	"github.com/splitbill/backend/internal/utils"
	"google.golang.org/api/option"
)

type AuthMiddleware struct {
	authClient *auth.Client
}

func NewAuthMiddleware(credentialsFile string) *AuthMiddleware {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var app *firebase.App
	var err error

	if credentialsFile != "" {
		opt := option.WithCredentialsFile(credentialsFile)
		app, err = firebase.NewApp(ctx, nil, opt)
	} else {
		// Use default credentials (e.g., GOOGLE_APPLICATION_CREDENTIALS env var)
		app, err = firebase.NewApp(ctx, nil)
	}

	if err != nil {
		log.Printf("⚠️  Warning: Failed to initialize Firebase: %v", err)
		log.Println("⚠️  Auth middleware will use development mode (no token verification)")
		return &AuthMiddleware{authClient: nil}
	}

	authClient, err := app.Auth(ctx)
	if err != nil {
		log.Printf("⚠️  Warning: Failed to get Firebase Auth client: %v", err)
		return &AuthMiddleware{authClient: nil}
	}

	log.Println("✅ Firebase Auth initialized successfully")
	return &AuthMiddleware{authClient: authClient}
}

// Authenticate verifies the Firebase ID token in the Authorization header
func (am *AuthMiddleware) Authenticate() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			utils.RespondUnauthorized(c, "Authorization header is required")
			c.Abort()
			return
		}

		// Extract token from "Bearer <token>"
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			utils.RespondUnauthorized(c, "Invalid authorization header format")
			c.Abort()
			return
		}

		idToken := parts[1]

		// Development mode: skip verification if Firebase is not configured
		if am.authClient == nil {
			// In dev mode, use the token as the UID directly
			c.Set("firebase_uid", idToken)
			c.Set("user_phone", "")
			c.Next()
			return
		}

		// Verify the Firebase ID token
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		token, err := am.authClient.VerifyIDToken(ctx, idToken)
		if err != nil {
			utils.RespondUnauthorized(c, "Invalid or expired token")
			c.Abort()
			return
		}

		// Set user info in context
		c.Set("firebase_uid", token.UID)
		if phone, ok := token.Claims["phone_number"].(string); ok {
			c.Set("user_phone", phone)
		}

		c.Next()
	}
}
