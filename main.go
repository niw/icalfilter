package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"
	"unsafe"
)

/*
#cgo CFLAGS: -I${SRCDIR}/libical/include
#cgo !linux,amd64 LDFLAGS: -L${SRCDIR}/libical/lib -lical
#cgo linux,amd64 LDFLAGS: -L${SRCDIR}/libical/lib64 -lical

#include <libical/ical.h>
#include <stdio.h>
#include <stdlib.h>
*/
import "C"

func shrink(months int, component *C.icalcomponent) error {
	if C.icalcomponent_isa(component) != C.ICAL_VCALENDAR_COMPONENT {
		return fmt.Errorf("component is not VCALENDER")
	}

	// Remove all `VEVENT` that end time is before given months ago.
	since := time.Now().AddDate(0, -months, 0)
	removing := []*C.icalcomponent{}
	for c := C.icalcomponent_get_first_component(component, C.ICAL_VEVENT_COMPONENT); c != nil; c = C.icalcomponent_get_next_component(component, C.ICAL_VEVENT_COMPONENT) {
		// summary := C.icalcomponent_get_summary(c)
		span := C.icalcomponent_get_span(c)
		// startTime := time.Unix(int64(span.start), 0)
		endTime := time.Unix(int64(span.end), 0)

		if endTime.Before(since) {
			removing = append(removing, c)
			// layout := "2006/01/02 03:04:05"
			// fmt.Printf("Removing: %s - %s: %s\n", startTime.Format(layout), endTime.Format(layout), C.GoString(summary))
		} else {
			// Remove all `X-LIC-ERROR` properties.
			removing := []*C.icalproperty{}
			for p := C.icalcomponent_get_first_property(c, C.ICAL_ANY_PROPERTY); p != nil; p = C.icalcomponent_get_next_property(c, C.ICAL_ANY_PROPERTY) {
				if C.icalproperty_isa(p) == C.ICAL_XLICERROR_PROPERTY {
					removing = append(removing, p)
				} else {
					// Remove all `X-...` parameters from all the other properties.
					C.icalproperty_remove_parameter_by_kind(p, C.ICAL_X_PARAMETER)
				}
			}
			for _, p := range removing {
				C.icalcomponent_remove_property(c, p)
			}
		}
	}
	for _, c := range removing {
		C.icalcomponent_remove_component(component, c)
	}

	return nil
}

func main() {
	input := flag.String("i", "", "input file")
	output := flag.String("o", "", "output file")
	flag.Parse()

	parser := C.icalparser_new()
	defer C.icalparser_free(parser)

	if input != nil && len(*input) > 0 {
		p := C.CString(*input)
		m := C.CString("r")
		f, err := C.fopen(p, m)
		defer C.free(unsafe.Pointer(p))
		defer C.free(unsafe.Pointer(m))
		if err != nil {
			log.Fatalln(err)
		}
		defer C.fclose(f)
		C.icalparser_set_gen_data(parser, unsafe.Pointer(f))
	} else {
		C.icalparser_set_gen_data(parser, unsafe.Pointer(C.stdin))
	}

	// Because `line_gen_func` is same sigunature as `fgets(3)` (I think it is intentional,)
	// we can use `fgets(3)` as a `line_get_func` for `icalparser_parse`.
	// The 3rd argument `FILE *` is geven as gen data `void *` by `icalparser_set_gen_data`,
	// which is either `fopen` file or `stdin`.
	component := C.icalparser_parse(parser, (*C.icalparser_line_gen_func)(C.fgets))
	if component == nil {
		log.Fatalln("Fail to parse component.")
	}
	defer C.icalcomponent_free(component)

	err := shrink(3, component)
	if err != nil {
		log.Fatalln(err)
	}

	result := C.GoString(C.icalcomponent_as_ical_string(component))

	if output != nil && len(*output) > 0 {
		f, err := os.Create(*output)
		if err != nil {
			log.Fatalln(err)
		}
		defer f.Close()
		fmt.Fprint(f, result)
	} else {
		fmt.Print(result)
	}
}
