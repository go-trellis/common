/*
Copyright © 2018 Henry Huang <hhh@rutcode.com>

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

package main

import (
	"fmt"
	"time"

	"trellis.tech/trellis/common.v1/cache"
)

var table1 = "tab1"

func evict(key, value interface{}) {
	fmt.Println("evict kv: ", key, value)
}

func main() {
	c := cache.New()
	if err := c.New(table1,
		cache.OptionValueMode(cache.ValueModeBag),
		cache.OptionKeySize(3),
		cache.OptionEvict(evict),
	); err != nil {
		panic(err)
	}

	tab1, ok := c.GetTableCache(table1)
	if !ok {
		panic("table1 should exists")
	}
	ok = tab1.Insert("key4", "value4")
	if !ok {
		panic("set key4 value4 failed")
	}

	ok = tab1.Insert("key1", "value1")
	if !ok {
		panic("set key1 value1 failed")
	}

	ok = tab1.Insert("key1", "value111")
	if !ok {
		panic("set key1 value111 failed")
	}

	ok = tab1.Insert("key1", "value111")
	if !ok {
		panic("set key1 value111 again failed")
	}

	vs, ok := tab1.Lookup("key1")
	if !ok {
		panic("key1 should exists")
	}

	if vs[0] != "value1" || vs[1] != "value111" || len(vs) != 2 {
		panic("key1 should be value1 and value111")
	}

	_, ok = tab1.Lookup("key2")
	if ok {
		panic("key2 should not exists")
	}

	ok = tab1.InsertExpire("key2", "value2", time.Second)
	if !ok {
		panic("set key2 failed")
	}

	_, ok = tab1.Lookup("key2")
	if !ok {
		panic("key2 should exists")
	}

	ok = tab1.Insert("key3", "value3")
	if !ok {
		panic("set key3 failed")
	}

	_, ok = tab1.Lookup("key4")
	if ok {
		panic("key2 should evicted")
	}

	ok = tab1.DeleteObject("key1")
	if !ok {
		panic("del key1 failed")
	}

	_, ok = tab1.Lookup("key1")
	if ok {
		panic("key1 should not exists")
	}

	tab1.SetExpire("key1", time.Second)

	_, ok = tab1.Lookup("key3")
	if !ok {
		panic("key1 should exists")
	}

	time.Sleep(time.Second)

	_, ok = tab1.Lookup("key2")
	if ok {
		panic("key2 should not exists")
	}

	tab1.DeleteObjects()

	_, ok = tab1.Lookup("key1")
	if ok {
		panic("key1 should not exists")
	}
}
