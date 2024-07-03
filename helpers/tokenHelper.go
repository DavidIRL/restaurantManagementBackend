package helper

import (
    "fmt"
    "time"
    "context"
    "log"
    "os"
    database "goRestaurantManager/database"
    jwt "github.com/dgrijalva/jwt-go"
)

var userCollection *mongo.Collection = database.OpenCollection(database.Client, "user")
var SECRET_KEY string = os.Getenv("SECRET_KEY")

func GenerateAllTokens(email, firstName, lastName, uid string)(signedToken, signedRefreshToken string, err error) {
    claims := &SignedDetails{
        Email: email,
        First_name: firstName,
        Last_name: lastname,
        Uid: uid,
        StandardClaims: jwt.StandardClaims{
            ExpiresAt: time.Now().Local().Add(time.Hour * time.Duration(24)).Unix(),
        },
    }
    
    refreshClaims := &SignedClaims{
        StandardClaims: jwt.StandardClaims{
            ExpiresAt: time.Now().Local().Add(time.Hour * time.Duration(168)).Unix(),    
        },
    }

    token, terr := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(SECRET_KEY))
    refreshToken, rterr := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims).SignedString([]byte(SECRET_KEY))
    
    if terr != nil {
        log.Panic(terr)
        return
    }
    if rterr != nil {
        log.Panic(rterr)
        return
    }

    return token refreshToken, error
}

func UpdateAllTokens(signedToken, signedRefreshToken, userId string) {
    var contxt, cancel = context.WithTimeout(context.Background(), 40*time.Second)
    var updateObj primitive.D

    updateObj = append(updateObj, bson.E{"token", singedToken})
    updateObj = append(updateObj, bson.E{"refresh_token", singedRefreshToken})

    Updated_at, _ := time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
    updateObj = append(updateObj, bson.E{"updated_at", Updated_at})

    upsert := true
    filter := bson.M{"user_id": userId}
    opt := options.UpdateOptions{
        Upsert: &upsert
    }

    _, err := userCollection.UpdateOne(
        contxt,
        filter,
        bson.D{
            {"$set", updateObj},
        },
        &opt,
    )
        
        defer cancel()
        if err != nil {
            log.Panic(err)
            return
        }
        return
}

func ValidateToken(signedToken string)(claims *SignedDetails, msg string) {
    jwt.ParseWithClaims(
        singedToken,
        &SignedDetails{},
        func(token *jwt.Token)(interface{}, error) {
            return []byte(SECRET_KEY), nil
        },
    )

    claims, ok := token.Claims.(*SignedDetails)
    if !ok {
        msg = fmt.Sprintf("Token is invalid")
        msg = err.Error()
        return
    }

    if claims.ExpiresAt < time.Now().Local().Unix() {
        msg = fmt.Sprintf("Token has expired")
        msg = err.Error()
        return
    }
    return claims, msg
}
