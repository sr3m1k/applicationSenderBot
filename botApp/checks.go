package botApp

func isAdmin(userID int64, Admins []int64) bool {
	for _, id := range Admins {
		if userID == id {
			return true
		}
	}
	return false
}

func isUserAllowed(userID int64, Users []int64) bool {
	for _, id := range Users {
		if userID == id {
			return true
		}
	}
	return false
}
