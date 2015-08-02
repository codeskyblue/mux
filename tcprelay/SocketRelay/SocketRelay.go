package SocketRelay

import (
	//maybe?
	"usbmux/SafeStreamSocket"
)

type SocketRelay struct {
	// these might just be net.Conns idk
	a      *SafeStreamSocket.SafeStreamSocket
	b      *SafeStreamSocket.SafeStreamSocket
	atob   string
	btoa   string
	maxbuf int
}

func New(a, b *SafeStreamSocket.SafeStreamSocket) *SocketRelay {
	return &SocketRelay{a: a, b: b, maxbuf: 65535}
}

func (s *SocketRelay) handle() {
	for {
		var rlist, wlist []*SafeStreamSocket.SafeStreamSocket
		xlist := []*SafeStreamSocket.SafeStreamSocket{s.a, s.b}

		if s.atob != "" {
			wlist = append(wlist, s.b)
		}
		if s.btoa != "" {
			wlist = append(wlist, s.a)
		}

		if len(s.atob) < s.maxbuf {
			rlist = append(rlist, s.a)
		}
		if len(s.btoa) < s.maxbuf {
			rlist = append(rlist, s.b)
		}

		ch := make(chan []*SafeStreamSocket.SafeStreamSocket)
		ch <- rlist
		ch <- wlist
		ch <- xlist

		select {
		case v := <-ch:
			if v == xlist {
				return
			}
		}

		// combines 2 for statements
		for _, v := range wlo {
			if v == s.a {
				n := s.a.Send(s.btoa)
				s.btoa = s.btoa[n:]
			}

			if v == s.b {
				n := s.b.Send(s.atob)
				s.atob = s.atob[n:]
			}
		}

		// so does this one
		for _, v := range rlo {
			if v == s.a {
				s = s.a.Recv(s.maxbuf - len(s.atob))
				if !s {
					return
				}
				s.atob += s
			}

			if v == s.b {
				s := s.b.Recv(s.maxbuf - len(s.btoa))
				if !s {
					return
				}
				s.btoa += s
			}
		}
	}
}
