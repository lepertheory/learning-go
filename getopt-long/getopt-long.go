package getopt_long

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"unsafe"
)

// #include <getopt.h>
// #include <stdlib.h>
import "C"

type ArgRequirement int
const (
	ArgNotAllowed ArgRequirement = iota
	ArgOptional   ArgRequirement = iota
	ArgRequired   ArgRequirement = iota
)

type Option struct {
	Name *string
	Short *string
	Required bool
	Arg ArgRequirement
}

type OptionResult struct {
	SetCount  int
	Arguments []string
}

type GetOpt struct {
	Options   []Option
	Arguments []string
	Results   map[Option]OptionResult
}

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

func getOptstring(options []Option) string {
	var retval strings.Builder
	for _, option := range options {
		if option.Short != nil {
			retval.WriteString(*option.Short)
			retval.WriteString(option.Arg.toOptstring())
		}
	}
	return retval.String()
}

func (o ArgRequirement) toOptstring() string {
	switch o {
	case ArgNotAllowed: return ""
	case ArgOptional:   return "::"
	case ArgRequired:   return ":"
	default:
		// TODO: Is this the right way to raise a programming error?
		panic(errors.New(fmt.Sprintf("Unexpected ArgRequirement value: %d", o)))
	}
}

func (o ArgRequirement) toHasArg() C.int {
	switch o {
	case ArgNotAllowed: return C.no_argument
	case ArgOptional:   return C.optional_argument
	case ArgRequired:   return C.required_argument
	default:
		// TODO: Is this the right way to raise a programming error?
		panic(errors.New(fmt.Sprintf("Unexpected ArgRequirement value: %d", o)))
	}
}

func (o *GetOpt) Process() {

	argvSlice, argvSliceCloser := getCSlice(
		o.Arguments,
		func(from string)(*C.char, func()) { return getCString(from) },
	)
	defer argvSliceCloser()
	argv := &(argvSlice[0])
	argc := C.int(len(argvSlice))

	optstringBacking := getOptstring(o.Options)
	optstring, optstringCloser := getCString(optstringBacking)
	defer optstringCloser()

	optionsSlice, optionsSliceCloser := getCSlice(
		o.Options,
		func(from Option)(C.struct_option, func()) {
			name, nameCloser := getCString(*from.Name)

			// FIXME: What happens if .toHasArg() panics and is recovered? I'm missing something,
			//        this shouldn't be so painful.
			return C.struct_option{
				name: name,
				has_arg: from.Arg.toHasArg(),
				flag: (*C.int)(C.NULL),
				val: C.int(0),
			}, nameCloser
		},
	)
	defer optionsSliceCloser()

	longOpts := make(map[string]*Option)
	shortOpts := make(map[string]*Option)
	for _, opt := range(o.Options) {
		if opt.Name	!= nil {
			longOpts[*opt.Name] = &opt
		}
		if opt.Short != nil {
			shortOpts[*opt.Short] = &opt
		}
	}

	// FIXME: Is this the right place to initialize this map?
	o.Results = map[Option]OptionResult{}
	optind := C.int(0)
	for {
		// FIXME: Handle option arguments.
		result := C.getopt_long(argc, argv, optstring, &(optionsSlice[0]), &optind)
		if result == -1 {
			break
		}

		var opt *Option
		if result == 0 {
			opt = longOpts[C.GoString(optionsSlice[optind].name)]
		} else {
			opt = shortOpts[string(rune(result))]
		}
		// TODO: There has to be a reason this is such a pain, and it has to be that I'm doing
		//       something wrong.
		// FIXME: Initialize at the right time.
		optResult := o.Results[*opt]
		optResult.SetCount++
		o.Results[*opt] = optResult
	}

}

func Fart() {
	fmt.Println("fart")

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
		defer C.free(unsafe.Pointer(opts[i].name))
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
