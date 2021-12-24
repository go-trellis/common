/*
Copyright © 2019 Henry Huang <hhh@rutcode.com>

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

package txorm

import (
	"trellis.tech/trellis/common.v0.1/errcode"
)

// define connector errors
var (
	ErrNotFoundDefaultDatabase     = errcode.New("not found default database")
	ErrAtLeastOneRepo              = errcode.New("input one repo at least")
	ErrNotFoundTransactionFunction = errcode.New("not found transaction function")
	ErrStructCombineWithRepo       = errcode.New("your repository struct should combine repo")
	ErrFailToCreateRepo            = errcode.New("fail to create an new repo")
	ErrFailToConvertTXToNonTX      = errcode.New("could not convert TX to NON-TX")
	ErrTransactionIsAlreadyBegin   = errcode.New("transaction is already begin")
	ErrNonTransactionCantCommit    = errcode.New("non-transaction can't commit")
	ErrTransactionSessionIsNil     = errcode.New("transaction session is nil")
	ErrNotFoundXormEngine          = errcode.New("not found xorm engine")
)
