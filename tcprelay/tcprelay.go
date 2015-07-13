package tcprelay

import (
	"fmt"
	"usbmux"
)

type SocketRelay struct {
	a      int
	b      int
	atob   string
	btoa   string
	maxbuf int
}

func NewSocketRelay(a, b, maxbuf int) *SocketRelay {
	return &SocketRelay{a, b, maxbuf: 65535}
}

// not done
func (s *SocketRelay) handle() {
	for true {
		var rlist, wlist []int
		xlist := []int{s.a, s.b}

		if s.atob {
			wlist.append(s.b)
		}
		if s.btoa {
			wlist.append(s.a)
		}
		if len(s.atob) < s.maxbuf {
			rlist.append(s.a)
		}
		if len(s.btoa) < s.maxbuf {
			rlist.append(s.a)
		}
	}
}

type TCPRelay struct {
}
