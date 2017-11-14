package main

/*
typedef struct _IO_FILE FILE;
#include <stdlib.h>
#include <libio.h>
*/
import "C"

import (
	"fmt"
	"github.com/rainycape/dl"
	"io/ioutil"
	"log"
	"os"
	"reflect"
	"syscall"
	"testing"
	"unsafe"
)

const From string = "RWRT_FROM"
const To string = "RWRT_TO"
const DebugString string = "LOADER_DEBUG"

var fatal = log.Fatalf

func logIt(message string) {
	debugFile := os.Getenv(DebugString)
	if debugFile == "" {
		return
	}
	fp, err := os.OpenFile(debugFile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0600)
	if err != nil {
		fatal(err.Error())
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
	if modeString == "r" {
		filePath = getPath(path)
	}
	var returnPath *C.char = C.CString(filePath)
	defer C.free(unsafe.Pointer(returnPath))
	lib, err := dl.Open("libc", 0)
	if err != nil {
		fatal(fmt.Sprintf("Unable to dlopen libc: %s\n", err))
	}
	var origFopen func(p, m *C.char) *C.FILE
	err = lib.Sym("fopen", &origFopen)
	if err != nil {
		fatal(fmt.Sprintf("Unable to load fopen: %s\n", err))
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

// Tests
const debugFile string = "testdata/__debug"

func checkData(t *testing.T, fd int, want []byte) {
	got := make([]byte, len(want))
	_, err := syscall.Read(fd, got)
	if err != nil {
		t.Fatalf("Error readying bytes from fixture: %s\n", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Got: %s, want: %s\n", string(got), string(want))
	}
}

func checkLog(t *testing.T, want []byte) {
	// Check the log
	got, err := ioutil.ReadFile(debugFile)
	defer os.Remove(debugFile)
	if err != nil {
		t.Fatalf("Unable to read from debugfile: %s\n", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Got: %s, want: %s\n", string(got), string(want))
	}
}

func getFileno(t *testing.T, file *C.FILE) int {
	// We need to load fileno from libc to get the file descriptor from C.FILE.
	// We can't do this direct because importing stdio breaks our local fopen!
	lib, err := dl.Open("libc", 0)
	if err != nil {
		t.Fatal("Unable to dlopen libc:", err)
	}
	var fileno func(f *C.FILE) int32
	err = lib.Sym("fileno", &fileno)
	if err != nil {
		t.Fatal("Unable to load fileno:", err)
	}
	fd := fileno(file)
	if fd == -1 {
		t.Fatalf("Error opening fixture!")
	}
	return int(fd)
}

func testOpen(t *testing.T) {
	// Expect that reading from results in "from\n"
	os.Setenv(From, "")
	os.Setenv(To, "")
	os.Setenv(DebugString, debugFile)
	var tPath *C.char = C.CString("testdata/from")
	defer C.free(unsafe.Pointer(tPath))
	fd := open(tPath, C.int(os.O_RDONLY), C.mode_t(0600))
	if fd == -1 {
		t.Fatalf("Error opening fixture!")
	}
	defer syscall.Close(fd)
	checkData(t, fd, []byte("from\n"))
	checkLog(t, []byte("open called on testdata/from\n"))
}

func testOpenWithVars(t *testing.T) {
	// Expect that reading from results in "to\n"
	os.Setenv(From, "testdata/from")
	os.Setenv(To, "testdata/to")
	os.Setenv(DebugString, debugFile)
	var tPath *C.char = C.CString("testdata/from")
	defer C.free(unsafe.Pointer(tPath))
	fd := open(tPath, C.int(os.O_RDONLY), C.mode_t(0600))
	if fd == -1 {
		t.Fatalf("Error opening fixture!")
	}
	defer syscall.Close(fd)
	checkData(t, fd, []byte("to\n"))
	// Check the log
	checkLog(t, []byte("open called on testdata/from\nMatched 'testdata/from', changing to 'testdata/to'\n"))
}

func testOpenWithVarsRW(t *testing.T) {
	// Expect that reading from results in "from\n" due to opening RD/RW
	os.Setenv(From, "testdata/from")
	os.Setenv(To, "testdata/to")
	os.Setenv(DebugString, "")
	var tPath *C.char = C.CString("testdata/from")
	defer C.free(unsafe.Pointer(tPath))
	fd := open(tPath, C.int(os.O_RDWR), C.mode_t(0600))
	if fd == -1 {
		t.Fatalf("Error opening fixture!")
	}
	defer syscall.Close(fd)
	checkData(t, fd, []byte("from\n"))
}

func testFOpen(t *testing.T) {
	// Expect that reading from results in "from\n"
	os.Setenv(From, "")
	os.Setenv(To, "")
	os.Setenv(DebugString, debugFile)
	var tPath *C.char = C.CString("testdata/from")
	var tMode *C.char = C.CString("r")
	defer C.free(unsafe.Pointer(tPath))
	defer C.free(unsafe.Pointer(tMode))
	file := fopen(tPath, tMode)
	if file == nil {
		t.Fatalf("Error opening fixture!")
	}
	fd := getFileno(t, file)
	checkData(t, int(fd), []byte("from\n"))
	checkLog(t, []byte("fopen called on testdata/from\n"))
}

func testFOpenWithVars(t *testing.T) {
	// Expect that reading from results in "to\n"
	os.Setenv(From, "testdata/from")
	os.Setenv(To, "testdata/to")
	debugFile := "testdata/__debug"
	os.Setenv(DebugString, debugFile)
	var tPath *C.char = C.CString("testdata/from")
	var tMode *C.char = C.CString("r")
	defer C.free(unsafe.Pointer(tPath))
	defer C.free(unsafe.Pointer(tMode))
	file := fopen(tPath, tMode)
	if file == nil {
		t.Fatalf("Error opening fixture!")
	}
	fd := getFileno(t, file)
	checkData(t, fd, []byte("to\n"))
	checkLog(t, []byte("fopen called on testdata/from\nMatched 'testdata/from', changing to 'testdata/to'\n"))
}

func testFOpenWithVarsRW(t *testing.T) {
	// Expect that reading from results in "from\n" due to opening RD/RW
	os.Setenv(From, "testdata/from")
	os.Setenv(To, "testdata/to")
	os.Setenv(DebugString, "")
	var tPath *C.char = C.CString("testdata/from")
	var tMode *C.char = C.CString("r+")
	defer C.free(unsafe.Pointer(tPath))
	defer C.free(unsafe.Pointer(tMode))
	file := fopen(tPath, tMode)
	if file == nil {
		t.Fatalf("Error opening fixture!")
	}
	fd := getFileno(t, file)
	checkData(t, fd, []byte("from\n"))
}

func testDebugFailure(t *testing.T) {
	// Expect that reading from results in "from\n" due to opening RD/RW
	os.Setenv(DebugString, "/root/bad")
	var tPath *C.char = C.CString("testdata/from")
	var tMode *C.char = C.CString("r+")
	defer C.free(unsafe.Pointer(tPath))
	defer C.free(unsafe.Pointer(tMode))
	var errors []string
	fatal = func(msg string, i ...interface{}) { errors = append(errors, msg) }
	fopen(tPath, tMode)
	if len(errors) != 1 {
		t.Fatalf("Expected to receive a debug error but found none\n")
	}
	want := "open /root/bad: permission denied"
	if errors[0] != want {
		t.Fatalf("Got error: %s, wanted: %s\n", errors[0], want)
	}
}

func main() {}
