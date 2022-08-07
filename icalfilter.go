package icalfilter

/*
#cgo CFLAGS: -I${SRCDIR}/libical/include
#cgo LDFLAGS: -L${SRCDIR}/libical/lib -L${SRCDIR}/libical/lib64 -lical

#include <libical/ical.h>
#include <stdio.h>
#include <stdlib.h>
*/
import "C"

import (
	"fmt"
	"time"
	"unsafe"
)

type Calendar struct {
	component *C.icalcomponent
}

func parse(f *C.FILE) (*Calendar, error) {
	parser := C.icalparser_new()
	defer C.icalparser_free(parser)

	C.icalparser_set_gen_data(parser, unsafe.Pointer(f))

	// Because `line_gen_func` is same sigunature as `fgets(3)` (I think it is intentional,)
	// we can use `fgets(3)` as a `line_get_func` for `icalparser_parse`.
	// The 3rd argument `FILE *` is geven as gen data `void *` by `icalparser_set_gen_data`,
	// which can be `fopen` file or `stdin`.
	c := C.icalparser_parse(parser, C.icalparser_line_gen_func(C.fgets))
	if c == nil {
		return nil, fmt.Errorf("fail to parse string")
	}

	if C.icalcomponent_isa(c) != C.ICAL_VCALENDAR_COMPONENT {
		return nil, fmt.Errorf("component is not VCALENDER")
	}
	return &Calendar{c}, nil
}

func Open(path string) (*Calendar, error) {
	p := C.CString(path)
	m := C.CString("r")
	f, err := C.fopen(p, m)
	defer C.free(unsafe.Pointer(p))
	defer C.free(unsafe.Pointer(m))
	if err != nil {
		return nil, err
	}
	defer C.fclose(f)

	return parse(f)
}

func Parse(source string) (*Calendar, error) {
	s := C.CString(source)
	defer C.free(unsafe.Pointer(s))
	c := C.icalparser_parse_string(s)
	if c == nil {
		return nil, fmt.Errorf("fail to parse string")
	}
	if C.icalcomponent_isa(c) != C.ICAL_VCALENDAR_COMPONENT {
		return nil, fmt.Errorf("component is not VCALENDER")
	}
	return &Calendar{c}, nil
}

func (c *Calendar) Close() {
	C.icalcomponent_free(c.component)
}

func icaltimeFromTime(t time.Time) C.icaltimetype {
	timet := C.time_t(t.UTC().Unix())
	return C.icaltime_from_timet_with_zone(timet, 0, C.icaltimezone_get_utc_timezone())
}

func repeatAfter(c *C.icalcomponent, t time.Time) bool {
	// See `icalcomponent_foreach_recurrence` implementation.
	// Currently only sipports `RRULE`.
	// TODO: Support `RDATE`.

	dtstart := C.icalcomponent_get_dtstart(c)
	after := icaltimeFromTime(t)

	rrules := []*C.icalproperty{}
	for rrule := C.icalcomponent_get_first_property(c, C.ICAL_RRULE_PROPERTY); rrule != nil; rrule = C.icalcomponent_get_next_property(c, C.ICAL_RRULE_PROPERTY) {
		rrules = append(rrules, rrule)
	}
	for _, rrule := range rrules {
		recur := C.icalproperty_get_rrule(rrule)
		it := C.icalrecur_iterator_new(recur, dtstart)
		if it == nil {
			continue
		}
		defer C.icalrecur_iterator_free(it)

		if recur.count == 0 {
			C.icalrecur_iterator_set_start(it, after)
		}
		for rt := C.icalrecur_iterator_next(it); C.icaltime_is_null_time(rt) == 0; rt = C.icalrecur_iterator_next(it) {
			if C.icaltime_compare(rt, after) >= 0 && C.icalproperty_recurrence_is_excluded(c, &dtstart, &rt) == 0 {
				return true
			}
		}
	}

	return false
}

func (c *Calendar) FilterBefore(t time.Time) error {
	removing := []*C.icalcomponent{}
	for event := C.icalcomponent_get_first_component(c.component, C.ICAL_VEVENT_COMPONENT); event != nil; event = C.icalcomponent_get_next_component(c.component, C.ICAL_VEVENT_COMPONENT) {
		// summary := C.icalcomponent_get_summary(event)
		span := C.icalcomponent_get_span(event)
		// startTime := time.Unix(int64(span.start), 0)
		endTime := time.Unix(int64(span.end), 0)

		if endTime.Before(t) && !repeatAfter(event, t) {
			removing = append(removing, event)
			// layout := "2006/01/02 03:04:05"
			// fmt.Printf("Removing: %s - %s: %s\n", startTime.Format(layout), endTime.Format(layout), C.GoString(summary))
		} else {
			// Remove all `X-LIC-ERROR` properties.
			removing := []*C.icalproperty{}
			for p := C.icalcomponent_get_first_property(event, C.ICAL_ANY_PROPERTY); p != nil; p = C.icalcomponent_get_next_property(event, C.ICAL_ANY_PROPERTY) {
				if C.icalproperty_isa(p) == C.ICAL_XLICERROR_PROPERTY {
					removing = append(removing, p)
				} else {
					// Remove all `X-...` parameters from all the other properties.
					C.icalproperty_remove_parameter_by_kind(p, C.ICAL_X_PARAMETER)
				}
			}
			for _, p := range removing {
				C.icalcomponent_remove_property(event, p)
				C.icalproperty_free(p)
			}
		}
	}
	for _, event := range removing {
		C.icalcomponent_remove_component(c.component, event)
		C.icalcomponent_free(event)
	}

	return nil
}

func (c *Calendar) String() string {
	return C.GoString(C.icalcomponent_as_ical_string(c.component))
}
