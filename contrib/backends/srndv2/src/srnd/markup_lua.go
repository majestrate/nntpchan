// +build lua

package srnd

// #cgo pkg-config: lua5.2
// #include <lua.h>
// #include <lauxlib.h>
// #include <lualib.h>
// #include <stdlib.h>
// #include <string.h>
import "C"

import (
	"errors"
	"log"
	"sync"
	"unsafe"
)

const luaFuncName = "memeposting"

const markupErrorText = "[failed to render ur meme D:]"

// lua interpreter
type Lua struct {
	mtx   sync.RWMutex
	state *C.lua_State
}

// do full GC cycle
func (l *Lua) GC() {
	C.lua_gc(l.state, C.LUA_GCCOLLECT, 0)
}

// close the interpreter
// all resources are expunged and no operations can be done after this
func (l *Lua) Close() {
	if l.state != nil {
		C.lua_close(l.state)
	}
	l.state = nil
}

func (l *Lua) LoadFile(fname string) (err error) {
	cfname := C.CString(fname)
	//defer C.free(unsafe.Pointer(cfname))
	res := C.luaL_loadfilex(l.state, cfname, nil)
	if res == 0 {
		res = C.lua_pcallk(l.state, 0, C.LUA_MULTRET, 0, 0, nil)
		if res != 0 {
			err = errors.New(C.GoString(C.lua_tolstring(l.state, -1, nil)))
		}
	} else {
		// failed to load file
		err = errors.New("failed to load file " + fname)
	}
	return
}

func (l *Lua) MEMEPosting(prefix, body string) (meme string) {
	l.mtx.Lock()
	defer l.mtx.Unlock()
	cf := C.CString(luaFuncName)
	C.lua_getglobal(l.state, cf)
	cp := C.CString(prefix)
	C.lua_pushstring(l.state, cp)
	cb := C.CString(body)
	C.lua_pushstring(l.state, cb)
	res := C.lua_pcallk(l.state, 2, 1, 0, 0, nil)
	var sz C.size_t
	cret := C.lua_tolstring(l.state, -1, &sz)

	if sz > 0 {
		meme = C.GoStringN(cret, C.int(sz))
	}

	C.lua_settop(l.state, -(1)-1)

	if res != C.LUA_OK {
		// error
		log.Println("lua error:", meme)
		meme = markupErrorText
	}

	// free buffers
	C.free(unsafe.Pointer(cb))
	C.free(unsafe.Pointer(cp))
	return
}

// create a new lua interpreter
func createLua() (l *Lua) {
	l = new(Lua)
	l.state = C.luaL_newstate()
	if l.state == nil {
		l = nil
	} else {
		// open stdlib
		C.luaL_openlibs(l.state)
	}
	return
}

var luaInt *Lua

func extraMemePosting(src, prefix string) string {
	if luaInt == nil {
		return src
	}
	defer luaInt.GC()
	return luaInt.MEMEPosting(src, prefix)
}

func SetMarkupScriptFile(fname string) error {
	if luaInt != nil {
		luaInt.Close()
		luaInt = nil
	}
	luaInt = createLua()
	err := luaInt.LoadFile(fname)
	if err == nil {
		return nil
	}
	luaInt.Close()
	luaInt = nil
	return err
}
