package xerr

import "testing"

func TestNew(t *testing.T) {
	e := New(CodeInternalError)
	println(e.Error())
}

func TestStackDepth(t *testing.T) {
	var e *XErr
	f1 := func() func() func() func() func() func() {
		return func() func() func() func() func() {
			return func() func() func() func() {
				return func() func() func() {
					return func() func() {
						return func() {
							e = New(CodeInternalError)
						}
					}
				}
			}
		}
	}
	f1()()()()()()

	println(e.FormatStack())
}
