package users

import "zeus/models"

func IsValid(username string) bool {
	user := models.User{Username: username}
	if err := user.GetInfo(nil); err != nil {
		return false
	}
	if user.Valid == models.UserValidYes {
		return true
	}
	return false
}
