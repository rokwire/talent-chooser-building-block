/*
 *   Copyright (c) 2020 Board of Trustees of the University of Illinois.
 *   All rights reserved.

 *   Licensed under the Apache License, Version 2.0 (the "License");
 *   you may not use this file except in compliance with the License.
 *   You may obtain a copy of the License at

 *   http://www.apache.org/licenses/LICENSE-2.0

 *   Unless required by applicable law or agreed to in writing, software
 *   distributed under the License is distributed on an "AS IS" BASIS,
 *   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *   See the License for the specific language governing permissions and
 *   limitations under the License.
 */

package model

import (
	"fmt"
	"strings"
	"talent-chooser/utils"
)

//User represents the user entity
type User struct {
	UUID            string
	PrivacySettings PrivacySettings
	Roles           *[]string
}

func (user *User) String() string {
	rolesData := ""
	if user.Roles != nil {
		rolesData = strings.Join(*user.Roles, ",")
	}
	return fmt.Sprintf("UUID:%s, PrivacySettings:%d, Roles:%s", user.UUID, user.PrivacySettings.Level, rolesData)
}

//PrivacySettings represents the user privacy settings entity
type PrivacySettings struct {
	Level int
}

//IlliniCash represents the illini cash entity
type IlliniCash struct {
	HousingResidentStatus bool
}

func (illiniCash *IlliniCash) String() string {
	return fmt.Sprintf("HousingResidentStatus:%t", illiniCash.HousingResidentStatus)
}

//Auth represents the shiboleth auth entity
type Auth struct {
	UIuceduUIN *string `json:"uiucedu_uin"`
}

func (auth *Auth) String() string {
	return fmt.Sprintf("UIuceduUIN:%s", utils.Value(auth.UIuceduUIN))
}

//Auth V2

//AuthV2 represents the auth v2 entity
type AuthV2 struct {
	IDToken      *string   `json:"id_token"`
	AccessToken  *string   `json:"access_token"`
	RefreshToken *string   `json:"refresh_token"`
	PhoneNumber  *string   `json:"phone_number"`
	UserInfo     *AuthInfo `json:"user_info"`
}

func (auth *AuthV2) String() string {
	userInfo := "nil"
	if auth.UserInfo != nil {
		userInfo = auth.UserInfo.String()
	}
	return fmt.Sprintf("IDToken:%s AccessToken:%s RefreshToken:%s PhoneNumber:%s UserInfo:[%s]",
		utils.Value(auth.IDToken), utils.Value(auth.AccessToken), utils.Value(auth.RefreshToken),
		utils.Value(auth.PhoneNumber), userInfo)
}

//AuthInfo represents auth info entity
type AuthInfo struct {
	Name              *string   `json:"name"`
	GivenName         *string   `json:"given_name"`
	FamilyName        *string   `json:"family_name"`
	UIuceduUIN        *string   `json:"uiucedu_uin"`
	PreferredUsername *string   `json:"preferred_username"`
	Sub               *string   `json:"sub"`
	Email             *string   `json:"email"`
	UIuceduIsMemberOf *[]string `json:"uiucedu_is_member_of"`
}

func (auth *AuthInfo) String() string {
	memberOfList := "nil"
	if auth.UIuceduIsMemberOf != nil {
		memberOfList = strings.Join(*auth.UIuceduIsMemberOf, ",")
	}
	return fmt.Sprintf("Name:%s GivenName:%s FamilyName:%s UIuceduUIN:%s PreferredUsername:%s Sub:%s Email:%s UIuceduIsMemberOf:%s",
		utils.Value(auth.Name), utils.Value(auth.GivenName), utils.Value(auth.FamilyName), utils.Value(auth.UIuceduUIN),
		utils.Value(auth.PreferredUsername), utils.Value(auth.Sub), utils.Value(auth.Email), memberOfList)
}

///////////////////////

//Auth V3

//AuthV3 represents the auth v2 entity
type AuthV3 struct {
	Token *AuthToken
	User  *AuthUser
	Card  *AuthCard
	Pii   *Pii
}

//AuthCard represents the auth card entity
type AuthCard struct {
	CardNumber    *string
	LibraryNumber *string
}

//AuthToken represents the auth token entity
type AuthToken struct {
	IDToken      *string
	AccessToken  *string
	RefreshToken *string
	PhoneNumber  *string
}

//AuthUser represents the auth user entity
type AuthUser struct {
	Name              *string
	GivenName         *string
	FamilyName        *string
	UIuceduUIN        *string
	PreferredUsername *string
	Sub               *string
	Email             *string
	UIuceduIsMemberOf *[]string
}

//Pii represents the pii entity
type Pii struct {
	DocumentType *string
}

////////////////////////

//Platform represents the platform entity
type Platform struct {
	OS *string
}
