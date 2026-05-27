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

package jwt

import (
	"time"

	"github.com/go-trellis/common/errors/errcode"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const (
	defaultTokenExpired = time.Hour * 24
)

type Config struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`

	MetaData map[string]any `json:"metadata"`
}

type TrellisClaims struct {
	Config               `json:",inline"`
	jwt.RegisteredClaims `json:",inline"`

	issureAt     time.Time
	tokenExpired time.Duration
}

type Option func(*TrellisClaims)

func UserID(userid string) Option {
	return func(o *TrellisClaims) {
		o.UserID = userid
	}
}

func Username(username string) Option {
	return func(o *TrellisClaims) {
		o.Username = username
	}
}

func MetaData(metadata map[string]any) Option {
	return func(o *TrellisClaims) {
		o.MetaData = metadata
	}
}

func TokenExpiredDuration(tokenExpired time.Duration) Option {
	return func(o *TrellisClaims) {
		o.tokenExpired = tokenExpired
	}
}

func Issuer(issuer string) Option {
	return func(o *TrellisClaims) {
		o.Issuer = issuer
	}
}

func IssureAt(t time.Time) Option {
	return func(o *TrellisClaims) {
		o.issureAt = t
	}
}

func Subject(subject string) Option {
	return func(o *TrellisClaims) {
		o.Subject = subject
	}
}

func Audience(audience []string) Option {
	return func(o *TrellisClaims) {
		o.Audience = audience
	}
}

func NewTrellisClaims(options ...Option) jwt.Claims {
	tc := &TrellisClaims{}
	tc.initOptions(options...)

	if tc.issureAt.IsZero() {
		tc.issureAt = time.Now()
	}

	tc.RegisteredClaims.ID = uuid.NewString()
	tc.RegisteredClaims.ExpiresAt = jwt.NewNumericDate(tc.issureAt.Add(tc.tokenExpired))
	tc.RegisteredClaims.IssuedAt = jwt.NewNumericDate(tc.issureAt)
	tc.RegisteredClaims.NotBefore = jwt.NewNumericDate(tc.issureAt)

	return tc
}

func (p *TrellisClaims) initOptions(options ...Option) {
	// Note: p cannot be nil when called as a method receiver
	for _, o := range options {
		o(p)
	}

	if p.tokenExpired <= 0 {
		p.tokenExpired = defaultTokenExpired
	}
}

func checkClaims(claims jwt.Claims) error {
	switch claims.(type) {
	case *TrellisClaims:
	case jwt.MapClaims:
	case *jwt.RegisteredClaims:
	default:
		return errcode.New("unsupported claims type")
	}
	return nil
}
