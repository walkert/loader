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

func myOpen(path *C.char, flags C.int) int {
	filePath := C.GoString(path)
	flagInt := int(flags)
	fp, err := os.OpenFile("/tmp/dee", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0600)
	if err != nil {
		log.Fatalln(err)
	}
	fp.WriteString(fmt.Sprintf("flags for %s are %d\n", filePath, flagInt))
	defer fp.Close()
	if flagInt&os.O_RDONLY != 0 {
		fp.WriteString(fmt.Sprintf("This is read-only! %s\n", filePath))
		filePath = getPath(path)
	}
	fd, err := syscall.Open(filePath, flagInt, 0600)
	if err != nil {
		return -1
	}
	return fd
}

//export open
func open(path *C.char, flags C.int) int {
	return myOpen(path, flags)
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
