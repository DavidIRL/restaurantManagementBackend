package middleware

import (
    "github.com/gin-gonic/gin"
)

func Authentication() gin.HandlerFunc {
    return func(c *gin.Context) {
        clientToken := c.Request.Header.Get("token")
        if clientToken == "" {
            msg := fmt.Sprintf("No Auth header provided")
            c.JSON(http.StatusInernalServerError, gin.H{"error": msg})
            c.Abort()
            return
        }

        claims, err := helper.ValidateToken(clientToken)
        if err != "" {
            c.JSON(http.StatusInernalServerError, gin.H{"error": err})
            c.Abort()
            return
        }

        c.Set("email", claims.Email)
        c.Set("first_name", claims.First_name)
        c.Set("last_name", claims.last_name)
        c.Set("uid", claims.Uid)

        c.Next()
    }
}
