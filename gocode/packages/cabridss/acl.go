package cabridss

type Rights struct {
	Read    bool `json:"read"`
	Write   bool `json:"write"`
	Execute bool `json:"execute"`
}

type ACLEntry struct {
	// on unix-like fsy DSS: x-uid:<uid> or x-gid:<gid> will be honored
	User   string `json:"user"` // on encrypted DSS, any alias for an IdentityConfig whose secret is owned by the user will be honored
	Rights Rights `json:"rights"`
}

func (ace ACLEntry) GetUser() string {
	return ace.User
}

func (ace ACLEntry) GetRights() Rights {
	return ace.Rights
}

func Users(aes []ACLEntry) (users []string) {
	for _, ae := range aes {
		users = append(users, ae.User)
	}
	return
}

func GetUserRights(aes []ACLEntry, user string) Rights {
	for _, ae := range aes {
		if ae.User == user {
			return ae.Rights
		}
	}
	return Rights{}
}
