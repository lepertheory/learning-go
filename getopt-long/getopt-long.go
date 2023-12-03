package getopt_long

import (
	"fmt"
	"os"
	"unsafe"
)

// #include <getopt.h>
// #include <stdlib.h>
import "C"

func getCString(value string) (*C.char, func()) {
	retString := C.CString(value)
	return retString, func() { C.free(unsafe.Pointer(retString)) }
}

func getCSlice[TFrom any, TTo any](values []TFrom, converter func(TFrom)(TTo, func())) ([]TTo, func()) {
	var retArray []TTo
	var retFrees []func()
	for _, value := range values {
		retValue, retFree := converter(value)
		retArray = append(retArray, retValue)
		retFrees = append([]func(){retFree}, retFrees...)
	}
	return retArray, func() {
		for _, freeFunc := range retFrees {
			freeFunc()
		}
	}
}

func Fart() {
	fmt.Println("fart")
	//argvArr, freeArgvArr := getCStringArray(os.Args)

	argvArr, freeArgvArr := getCSlice(os.Args, func(from string)(*C.char, func()) { return getCString(from); })
	defer freeArgvArr()

	argv := &(argvArr[0])
	argc := C.int(len(argvArr))
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
		argvArr[i] = C.CString(s)
		defer C.free(unsafe.Pointer(argvArr[i]))
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
