package main

/*
typedef struct _IO_FILE FILE;
#include <stdlib.h>
*/
import "C"

import (
	"fmt"
	"github.com/rainycape/dl"
	"log"
	"os"
	"strings"
	"syscall"
	"unsafe"
)

const From string = "RWRT_FROM"
const To string = "RWRT_TO"
const DebugString string = "LOADER_DEBUG"

func logIt(message string) {
	debugFile := os.Getenv(DebugString)
	if debugFile == "" {
		return
	}
	fp, err := os.OpenFile(debugFile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0600)
	if err != nil {
		log.Fatalln(err)
	}
	defer fp.Close()
	fmt.Fprintln(fp, message)
}

func getPath(inPath *C.char) string {
	outPath := C.GoString(inPath)
	from := os.Getenv(From)
	to := os.Getenv(To)
	if from == "" || to == "" {
		return outPath
	}
	if outPath == from {
		logIt(fmt.Sprintf("Matched '%s', changing to '%s'", outPath, to))
		return to
	}
	return outPath
}

func myOpen(caller string, path *C.char, flags C.int, mode C.mode_t) int {
	filePath := C.GoString(path)
	logIt(fmt.Sprintf("%s called on %s", caller, filePath))
	flagInt := int(flags)
	modeInt := uint32(mode)
	if !(flagInt&(os.O_WRONLY|os.O_CREATE|os.O_RDWR) > 0) {
		filePath = getPath(path)
	}
	fd, err := syscall.Open(filePath, flagInt, modeInt)
	if err != nil {
		return -1
	}
	return fd
}

func myFopen(caller string, path *C.char, mode *C.char) *C.FILE {
	filePath := C.GoString(path)
	modeString := C.GoString(mode)
	logIt(fmt.Sprintf("%s called on %s", caller, filePath))
	if !strings.Contains(modeString, "w") {
		filePath = getPath(path)
	}
	var returnPath *C.char = C.CString(filePath)
	defer C.free(unsafe.Pointer(returnPath))
	lib, err := dl.Open("libc", 0)
	if err != nil {
		log.Fatal("Unable to dlopen libc:", err)
	}
	var origFopen func(p, m *C.char) *C.FILE
	err = lib.Sym("fopen", &origFopen)
	if err != nil {
		log.Fatal("Unable to load fopen:", err)
	}
	return origFopen(returnPath, mode)
}

//export open
func open(path *C.char, flags C.int, mode C.mode_t) int {
	return myOpen("open", path, flags, mode)
}

//export fopen
func fopen(path *C.char, mode *C.char) *C.FILE {
	return myFopen("fopen", path, mode)
}

//export fopen64
func fopen64(path *C.char, mode *C.char) *C.FILE {
	return myFopen("fopen64", path, mode)
}

func main() {}
