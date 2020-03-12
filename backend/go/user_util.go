package macedonio

import (
	"encoding/base64"
	"fmt"
	"time"
)

func TokenToUser(token64 string) (DBUser, error) {
	token, err := base64.StdEncoding.DecodeString(token64)
	var dbToken DBUserToken

	if err != nil {
		return DBUser{}, fmt.Errorf("failed to parse token err: %v", err)
	}
	if len(token) == 0 {
		return DBUser{}, fmt.Errorf("empty token")
	}

	db := GetDBHandle()
	if db.Where("token = ?", token).First(&dbToken).RecordNotFound() {
		return DBUser{}, fmt.Errorf("recorder not found")
	}

	if dbToken.ExpiresAt.Before(time.Now()) {
		return DBUser{}, fmt.Errorf("expired token")
	}

	if db.Model(&dbToken).Related(&dbToken.DBUser).RecordNotFound() {
		return DBUser{}, db.Error
	}

	return dbToken.DBUser, nil

}

func TokenToDBToken(token64 string) (DBUserToken, error) {
	token, err := base64.StdEncoding.DecodeString(token64)
	var dbToken DBUserToken
	if err != nil {
		return DBUserToken{}, fmt.Errorf("failed to parse token err: %v", err)
	}

	db := GetDBHandle()
	if db.Where("token = ?", token).First(&dbToken).RecordNotFound() {
		return DBUserToken{}, db.Error
	}
	if dbToken.ExpiresAt.Before(time.Now()) {
		return DBUserToken{}, fmt.Errorf("expired token")
	}

	return dbToken, nil
}
