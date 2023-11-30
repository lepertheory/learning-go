package getopt_long

import (
	"fmt"
	"os"
	"unsafe"
)

/*
#include <getopt.h>

// Start From: https://groups.google.com/g/golang-nuts/c/pQueMFdY0mk/m/OAX5-Fqus0UJ
#include <stdlib.h>

static char** makeCharArray(int size) {
	// This can return C NULL on failure. As far as I know, this is not handled. Not that there is
	// much of a sane handling of this. Well, generating an error message should be pre-allocated,
	// also this is most likely to fail when trying to allocate huge amounts of memory or a
	// segfault, which is either that (is that in fact a possible segfault reason?) or not actual
	// memory (things like DMA or mmap())... neither of which can be handled in a regular-ass
	// program just trying to allocate memory to call getopt_long(), but both of which should be
	// handled properly in a library function.
	//
	// Is this call atomic? Could it "successfully" reserve the memory, then get a trap when
	// initializing? Almost certainly not, but best to check.
	return calloc(sizeof(char*), size);
}

static void setArrayString(char** a, char* s, int n) {
	// Cannot fail unless inputs are invalid.
	a[n] = s;
}

static void freeCharArray(char** a, int size) {
	int i;
	for (i = 0; i < size; i++) {
		// We are again not looking at the error code. See makeCharArray() for elaboration.
		free(a[i]);
	}
	free(a);
}
// End From: https://groups.google.com/g/golang-nuts/c/pQueMFdY0mk/m/OAX5-Fqus0UJ
*/
import "C"

// Is Go low-level enough that this might actually be something you can get a C pointer to?
type fart struct {
	name   string
	hasArg int
	flag   *int
	val    int
}

func getCString(value string) (*C.char, func()) {
	retString := C.CString(value)
	return retString, func() { C.free(unsafe.Pointer(retString)) }
}

func getCStringArray(values []string) ([]*C.char, func()) {
	var retStringArray []*C.char
	var retFreeArray []func()
	for _, value := range values {
		retString, retFree := getCString(value)
		retStringArray = append(retStringArray, retString)
		retFreeArray = append([]func(){retFree}, retFreeArray...)
	}
	return retStringArray, func() {
		for _, freeFunc := range retFreeArray {
			freeFunc()
		}
	}
}

func Fart() {
	fmt.Println("fart")
	//argc := C.int(len(os.Args))
	//argv := C.makeCharArray(argc)
	argvArr, freeArgvArr := getCStringArray(os.Args)
	defer freeArgvArr()
	argv := &(argvArr[0])
	argc := C.int(len(argvArr))
	// Is there no way for failures to happen here? A lot is happening. I'm just learning Go. Know
	// C++, Java, Python, etc., but not familiar with the walrus operator other than recognizing its
	// existence, and (that it apparently is a declaration + definition/initial assignment)?.
	//
	// We are (to my understanding):
	// Using example `./Fart 3
	// 1. Declaring a variable called `argv`, of the type that is the declared return value of
	//    whatever the return value of char** maps to in `cgo`.
	// 2. Calling the C function `makeCharArray`, with pass-by-value argument that is the `int`
	//    value of `len(os.Args)`, which is `1`.
	// 3.
	//defer C.freeCharArray(argv, argc)
	//goland:noinspection SpellCheckingInspection
	optstring, freeOptstring := getCString("h")
	defer freeOptstring()
	opts := [1]C.struct_option{{
		name: C.CString("help"),
		has_arg: C.no_argument,
		flag: (*C.int)(C.NULL),
		val: C.int(0),
	}}
	for i := range opts {
		defer func(idx int) {
			C.free(unsafe.Pointer(opts[idx].name))
			fmt.Printf("freed %d\n", idx)
		}(i)
	}
	for i, s := range os.Args {
		C.setArrayString(argv, C.CString(s), C.int(i))
	}
	optind := C.int(0)
	for {
		retval := C.getopt_long(argc, argv, optstring, &(opts[0]), &optind)
		if retval == -1 {
			break
		}
		fmt.Printf("optstring: %s\n", C.GoString(optstring))
		fmt.Printf("opts[0].name: %s\n", C.GoString(opts[0].name))
		fmt.Printf("opts[0].has_arg: %d\n", int(opts[0].has_arg))
		fmt.Printf("opts[0].flag: %d\n", opts[0].flag)
		fmt.Printf("opts[0].val: %d\n", opts[0].val)
		fmt.Printf("retval: %d\n", retval)
		fmt.Printf("optind: %d\n", optind)

		if retval == 0 {
			fmt.Printf("Long option: %s\n", C.GoString(opts[optind].name))
		} else {
			fmt.Printf("Short option: %s\n", string(rune(retval)))
		}
	}

	for int(optind) < len(os.Args) {
		//fmt.Printf("argument: %s\n", C.GoString(argv[optind]))
		optind++
	}
}

//// C helper functions:
//
//static char**makeCharArray(int size) {
//return calloc(sizeof(char*), size);
//}
//
//static void setArrayString(char **a, char *s, int n) {
//a[n] = s;
//}
//
//static void freeCharArray(char **a, int size) {
//int i;
//for (i = 0; i < size; i++)
//free(a[i]);
//free(a);
//}
//
//// Build C array in Go from sargs []string
//
//cargs := C.makeCharArray(C.int(len(sargs)))
//defer C.freeCharArray(cargs, C.int(len(sargs)))
//for i, s := range sargs {
//C.setArrayString(cargs, C.CString(s), C.int(i))
//}
