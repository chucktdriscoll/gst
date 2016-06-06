package gst

/*
#include <stdlib.h>
#include <gst/gst.h>
static GstPad* get_active_switch_pad(GstElement *switcher) {
	GstPad* active_pad;
	g_object_get(G_OBJECT(switcher), "active-pad", &active_pad, NULL);
	return active_pad;
}
*/
import "C"

import (
	"errors"
	"github.com/lidouf/glib"
	"time"
	"unsafe"
)

type State C.GstState

const (
	STATE_VOID_PENDING = State(C.GST_STATE_VOID_PENDING)
	STATE_NULL         = State(C.GST_STATE_NULL)
	STATE_READY        = State(C.GST_STATE_READY)
	STATE_PAUSED       = State(C.GST_STATE_PAUSED)
	STATE_PLAYING      = State(C.GST_STATE_PLAYING)
)

func (s *State) g() *C.GstState {
	return (*C.GstState)(s)
}

func (s State) String() string {
	switch s {
	case STATE_VOID_PENDING:
		return "STATE_VOID_PENDING"
	case STATE_NULL:
		return "STATE_NULL"
	case STATE_READY:
		return "STATE_READY"
	case STATE_PAUSED:
		return "STATE_PAUSED"
	case STATE_PLAYING:
		return "STATE_PLAYING"
	}
	panic("Unknown state")
}

type StateChangeReturn C.GstStateChangeReturn

const (
	STATE_CHANGE_FAILURE    = StateChangeReturn(C.GST_STATE_CHANGE_FAILURE)
	STATE_CHANGE_SUCCESS    = StateChangeReturn(C.GST_STATE_CHANGE_SUCCESS)
	STATE_CHANGE_ASYNC      = StateChangeReturn(C.GST_STATE_CHANGE_ASYNC)
	STATE_CHANGE_NO_PREROLL = StateChangeReturn(C.GST_STATE_CHANGE_NO_PREROLL)
)

type Element struct {
	GstObj
}

func (e *Element) g() *C.GstElement {
	return (*C.GstElement)(e.GetPtr())
}

func (e *Element) AsElement() *Element {
	return e
}

func (e *Element) Link(next ...*Element) bool {
	for _, dst := range next {
		if C.gst_element_link(e.g(), dst.g()) == 0 {
			return false
		}
		e = dst
	}
	return true
}

func (e *Element) Unlink(next ...*Element) {
	for _, dst := range next {
		C.gst_element_unlink(e.g(), dst.g())
		e = dst
	}
}

func (e *Element) LinkFiltered(dst *Element, filter *Caps) bool {
	return C.gst_element_link_filtered(e.g(), dst.g(), filter.g()) != 0
}

func (e *Element) LinkPads(pad_name string, dst *Element, dst_pad_name string) bool {
	src_pname := (*C.gchar)(C.CString(pad_name))
	defer C.free(unsafe.Pointer(src_pname))
	dst_pname := (*C.gchar)(C.CString(dst_pad_name))
	defer C.free(unsafe.Pointer(dst_pname))
	return C.gst_element_link_pads(e.g(), src_pname, dst.g(), dst_pname) != 0
}

func (e *Element) UnlinkPads(pad_name string, dst *Element, dst_pad_name string) {
	src_pname := (*C.gchar)(C.CString(pad_name))
	defer C.free(unsafe.Pointer(src_pname))
	dst_pname := (*C.gchar)(C.CString(dst_pad_name))
	defer C.free(unsafe.Pointer(dst_pname))
	C.gst_element_unlink_pads(e.g(), src_pname, dst.g(), dst_pname)
}

func (e *Element) SetState(state State) StateChangeReturn {
	return StateChangeReturn(C.gst_element_set_state(e.g(), C.GstState(state)))
}

func (e *Element) GetState(timeout_ns int64) (state, pending State,
	ret StateChangeReturn) {
	ret = StateChangeReturn(C.gst_element_get_state(
		e.g(), state.g(), pending.g(), C.GstClockTime(timeout_ns),
	))
	return
}

func (e *Element) AddPad(p *Pad) bool {
	return C.gst_element_add_pad(e.g(), p.g()) != 0
}

func (e *Element) GetRequestPad(name string) *Pad {
	s := (*C.gchar)(C.CString(name))
	defer C.free(unsafe.Pointer(s))
	cp := C.gst_element_get_request_pad(e.g(), s)
	if cp == nil {
		return nil
	}
	p := new(Pad)
	p.SetPtr(glib.Pointer(cp))
	return p
}

func (e *Element) GetStaticPad(name string) *Pad {
	s := (*C.gchar)(C.CString(name))
	defer C.free(unsafe.Pointer(s))
	cp := C.gst_element_get_static_pad(e.g(), s)
	if cp == nil {
		return nil
	}
	p := new(Pad)
	p.SetPtr(glib.Pointer(cp))
	return p
}

//LiD: add GetSwitchActivePad for input-selector active-pad property
func (e *Element) GetSwitchActivePad() *Pad {
	s := (*C.gchar)(C.CString("active-pad"))
	defer C.free(unsafe.Pointer(s))

	var r1 *C.GstPad
	r1 = C.get_active_switch_pad(e.g())
	if r1 == nil {
		return nil
	}
	r := new(Pad)
	r.SetPtr(glib.Pointer(r1))

	return r
}

func (e *Element) GetBus() *Bus {
	bus := C.gst_element_get_bus(e.g())
	if bus == nil {
		return nil
	}
	b := new(Bus)
	b.SetPtr(glib.Pointer(bus))
	return b
}

func (e *Element) QueryPosition(format Format) (time.Duration, error) {
	var pos C.gint64
	ret := C.gst_element_query_position(e.g(), *(format.g()), &pos)
	if ret == 0 {
		return -1, errors.New("Query position from element failed")
	} else {
		return time.Duration(pos), nil
	}
}

func (e *Element) QueryDuration(format Format) (time.Duration, error) {
	var duration C.gint64
	ret := C.gst_element_query_duration(e.g(), *(format.g()), &duration)
	if ret == 0 {
		return -1, errors.New("Query duration from element failed")
	} else {
		return time.Duration(duration), nil
	}
}

func (e *Element) Query(q *Query) bool {
	return C.gst_element_query(e.g(), q.g()) == 1
}

func (e *Element) SeekSimple(format Format, flags SeekFlags, pos int64) bool {
	return C.gst_element_seek_simple(e.g(), *(format.g()), flags.g(), (C.gint64)(pos)) == 1
}

// TODO: Move ElementFactoryMake to element_factory.go
func ElementFactoryMake(factory_name, name string) *Element {
	fn := (*C.gchar)(C.CString(factory_name))
	defer C.free(unsafe.Pointer(fn))
	n := (*C.gchar)(C.CString(name))
	defer C.free(unsafe.Pointer(n))
	ge := C.gst_element_factory_make(fn, n)
	if ge == nil {
		return nil
	}
	e := new(Element)
	e.SetPtr(glib.Pointer(ge))
	return e
}
