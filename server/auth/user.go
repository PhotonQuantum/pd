// Copyright 2020 TiKV Project Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// See the License for the specific language governing permissionKeys and
// limitations under the License.
//

package auth

import (
	"encoding/json"
	"sort"
)

// User records user info.
// Read-Only once created.
type User struct {
	Username string
	Hash     string
	RoleKeys map[string]struct{}
}

// jsonUser is used as an intermediate model when marshaling/unmarshaling json data
// because we need to convert map[Permission]struct{} from/to []Permission first.
type jsonUser struct {
	Username string   `json:"username"`
	Hash     string   `json:"hash"`
	RoleKeys []string `json:"roles"`
}

// SafeUser records user info without password hash, so it's safe to serialize it and send it as API responses.
// Read-Only once created.
type SafeUser struct {
	Username string   `json:"username"`
	RoleKeys []string `json:"roles"`
}

// MarshalJSON implements Marshaler interface.
func (u *User) MarshalJSON() ([]byte, error) {
	roleKeys := make([]string, 0, len(u.RoleKeys))
	for k := range u.RoleKeys {
		roleKeys = append(roleKeys, k)
	}
	sort.Strings(roleKeys)

	_u := jsonUser{Username: u.Username, Hash: u.Hash, RoleKeys: roleKeys}
	return json.Marshal(_u)
}

// UnmarshalJSON implements Unmarshaler interface.
func (u *User) UnmarshalJSON(bytes []byte) error {
	var _u jsonUser

	err := json.Unmarshal(bytes, &_u)
	if err != nil {
		return err
	}

	u.Username = _u.Username
	u.Hash = _u.Hash
	for _, v := range _u.RoleKeys {
		u.RoleKeys[v] = struct{}{}
	}

	return nil
}

// NewUser safely creates a new user instance.
func NewUser(username string, hash string) (*User, error) {
	err := validateName(username)
	if err != nil {
		return nil, err
	}

	return &User{Username: username, Hash: hash, RoleKeys: make(map[string]struct{})}, nil
}

// NewUserFromJSON safely deserialize a json string to a user instance.
func NewUserFromJSON(j string) (*User, error) {
	user := User{RoleKeys: make(map[string]struct{})}
	err := json.Unmarshal([]byte(j), &user)
	if err != nil {
		return nil, err
	}

	err = validateName(user.Username)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

// Clone creates a deep copy of user instance.
func (u *User) Clone() *User {
	return &User{Username: u.Username, Hash: u.Hash, RoleKeys: u.RoleKeys}
}

// GetSafeUser returns a SafeUser instance. More details is available in the comment of SafeUser struct.
func (u *User) GetSafeUser() SafeUser {
	roleKeys := make([]string, 0, len(u.RoleKeys))
	for k := range u.RoleKeys {
		roleKeys = append(roleKeys, k)
	}
	sort.Strings(roleKeys)

	return SafeUser{Username: u.Username, RoleKeys: roleKeys}
}

// GetUsername returns username of this user.
func (u *User) GetUsername() string {
	return u.Username
}

// GetRoleKeys returns role keys of this user.
func (u *User) GetRoleKeys() map[string]struct{} {
	return u.RoleKeys
}

// HasRole checks whether this user has a specific role.
func (u *User) HasRole(name string) bool {
	for k := range u.RoleKeys {
		if k == name {
			return true
		}
	}

	return false
}

// ComparePassword checks whether given string matches the password of this user.
func (u *User) ComparePassword(candidate string) error {
	return compareHashAndPassword(u.Hash, candidate)
}
