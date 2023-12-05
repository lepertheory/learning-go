package getopt_long

import (
	"errors"
	"fmt"
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

func getGoString(value *C.char) string {
	return C.GoString(value)
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
		// FIXME: This should probably be an error.
		if opt != nil {
			// TODO: There has to be a reason this is such a pain, and it has to be that I'm doing
			//       something wrong.
			// FIXME: Initialize at the right time.
			optResult := o.Results[*opt]
			optResult.SetCount++
			o.Results[*opt] = optResult
		}
	}

}
