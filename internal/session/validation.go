package session

// ReservedNames contains names that cannot be used by participants
var ReservedNames = map[string]bool{
	"Moderator": true,
}

// IsReservedName checks if a name is reserved
func IsReservedName(name string) bool {
	return ReservedNames[name]
}
