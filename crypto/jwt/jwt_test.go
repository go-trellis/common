/*
Copyright © 2025 Henry Huang <hhh@rutcode.com>

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program. If not, see <http://www.gnu.org/licenses/>.
*/

package jwt_test

import (
	"testing"
	"time"

	"github.com/go-trellis/common/crypto/jwt"
	"github.com/go-trellis/common/utils/testutils"

	go_jwt "github.com/golang-jwt/jwt/v5"
)

func TestNewJWT(t *testing.T) {
	jwtInstance := jwt.NewJWT([]byte("trellis"))

	claims := jwt.NewTrellisClaims(
		jwt.UserID("userid1"),
		jwt.Username("userid1_name"),
		jwt.MetaData(map[string]any{"scope": []string{"admin", "user"}, "age": "1"}),
		jwt.TokenExpiredDuration(time.Hour*12),
		jwt.Issuer("trellis"),
		jwt.Audience([]string{"trellis_audience"}),
		jwt.Subject("trellis_jwt"),
	)

	s, err := jwtInstance.GenToken(claims)
	testutils.Ok(t, err)

	tc := &jwt.TrellisClaims{}

	token, err := jwtInstance.ParseJWTWithClaims(s, tc)
	testutils.Ok(t, err)

	testutils.Assert(t, token.Valid, "token not valid")
	testutils.Equals(t, "userid1", tc.UserID)
	testutils.Equals(t, "userid1_name", tc.Username)
	testutils.Equals(t, []any{"admin", "user"}, tc.MetaData["scope"])
	testutils.Equals(t, "1", tc.MetaData["age"])
	testutils.Equals(t, "trellis", tc.Issuer)
	testutils.Equals(t, "trellis_jwt", tc.Subject)
	testutils.Equals(t, tc.Audience, go_jwt.ClaimStrings{"trellis_audience"})
}
