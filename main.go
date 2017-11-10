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
	"syscall"
	"unsafe"
)

const RWRT_FROM string = "RWRT_FROM"
const RWRT_TO string = "RWRT_TO"

func getPath(inPath *C.char) string {
	outPath := C.GoString(inPath)
	from := os.Getenv(RWRT_FROM)
	to := os.Getenv(RWRT_TO)
	if from == "" || to == "" {
		return outPath
	}
	if outPath == from {
		return to
	}
	return outPath
}

//export open
func open(path *C.char, flags C.int, mode C.mode_t) int {
	var modeInt uint32
	filePath := getPath(path)
	flagInt := int(flags)
	if flagInt&os.O_CREATE == os.O_CREATE {
		modeInt = uint32(mode)
	} else {
		modeInt = 0666
	}
	fd, err := syscall.Open(filePath, flagInt, modeInt)
	if err != nil {
		return -1
	}
	return fd
}

//export fopen
func fopen(path *C.char, mode *C.char) *C.FILE {
	filePath := getPath(path)
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

//export fopen64
func fopen64(path *C.char, mode *C.char) *C.FILE {
	filePath := getPath(path)
	var returnPath *C.char = C.CString(filePath)
	defer C.free(unsafe.Pointer(returnPath))
	lib, err := dl.Open("libc", 0)
	if err != nil {
		log.Fatal("Unable to dlopen libc:", err)
	}
	var origFopen64 func(p, m *C.char) *C.FILE
	err = lib.Sym("fopen64", &origFopen64)
	if err != nil {
		log.Fatal("Unable to load fopen64:", err)
	}
	return origFopen64(returnPath, mode)
}

func main() {}
