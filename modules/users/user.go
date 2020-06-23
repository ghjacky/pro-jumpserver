package users

import "zeus/models"

func IsValid(user *models.User) bool {
	if err := user.GetInfo(nil); err != nil {
		return false
	}
	if user.Valid == models.UserValidYes {
		return true
	}
	return false
}
