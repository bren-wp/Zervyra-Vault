//go:build windows

package main

import "unsafe"

func createFont(height int32, weight int32) uintptr {
	face := w("Segoe UI")
	h, _, _ := pCreateFont.Call(
		uintptr(int64(height)), 0, 0, 0, uintptr(weight), 0, 0, 0,
		1, 0, 0, 5, 0, uintptr(unsafe.Pointer(face)),
	)
	return h
}

func initThemeResources() {
	// Near-black surfaces with a restrained violet accent. This intentionally avoids
	// classic Win32 greys while keeping contrast high enough for long daily use.
	bgBrush, _, _ = pCreateSolidBrush.Call(rgb(10, 13, 19))
	panelBrush, _, _ = pCreateSolidBrush.Call(rgb(17, 23, 33))
	fieldBrush, _, _ = pCreateSolidBrush.Call(rgb(25, 33, 46))
	guiFont = createFont(-16, 400)
	ownGUIFont = guiFont != 0
	smallFont = createFont(-14, 400)
	headingFont = createFont(-25, 700)
	if guiFont == 0 {
		font, _, _ := pGetStockObject.Call(DEFAULT_GUI_FONT)
		guiFont = font
		ownGUIFont = false
	}
}

func destroyThemeResources() {
	for _, h := range []uintptr{bgBrush, panelBrush, fieldBrush, headingFont, smallFont} {
		if h != 0 {
			pDeleteObject.Call(h)
		}
	}
	if guiFont != 0 && ownGUIFont {
		pDeleteObject.Call(guiFont)
	}
}

func applyDarkWindow(h uintptr) {
	if h == 0 {
		return
	}
	one := int32(1)
	_, _, _ = pDwmSetAttribute.Call(h, 20, uintptr(unsafe.Pointer(&one)), unsafe.Sizeof(one))
	_, _, _ = pDwmSetAttribute.Call(h, 19, uintptr(unsafe.Pointer(&one)), unsafe.Sizeof(one))
	corner := int32(2) // DWMWCP_ROUND
	_, _, _ = pDwmSetAttribute.Call(h, 33, uintptr(unsafe.Pointer(&corner)), unsafe.Sizeof(corner))
	// DWMWA_CAPTION_COLOR. Unsupported Windows versions simply ignore it.
	caption := uint32(rgb(10, 13, 19))
	_, _, _ = pDwmSetAttribute.Call(h, 35, uintptr(unsafe.Pointer(&caption)), unsafe.Sizeof(caption))
}

func setCue(id int, cue string) {
	if h := controls[id]; h != 0 {
		pSendMessage.Call(h, EM_SETCUEBANNER, 1, uintptr(unsafe.Pointer(w(cue))))
	}
}

func isToggleButton(id int) bool {
	return id == ID_SHOWMASTER || id == ID_SHOWPASS || id == ID_FAVORITE
}
func isPrimaryButton(id int) bool {
	return id == ID_OPEN || id == ID_CREATE || id == ID_NEW || id == ID_SAVEENTRY
}

func paintButton(h uintptr) {
	var ps PAINTSTRUCT
	hdc, _, _ := pBeginPaint.Call(h, uintptr(unsafe.Pointer(&ps)))
	if hdc == 0 {
		return
	}
	defer pEndPaint.Call(h, uintptr(unsafe.Pointer(&ps)))
	var r RECT
	pGetClientRect.Call(h, uintptr(unsafe.Pointer(&r)))
	pFillRect.Call(hdc, uintptr(unsafe.Pointer(&r)), panelBrush)
	idv, _, _ := pGetDlgCtrlID.Call(h)
	id := int(idv)
	enabledNow, _, _ := pIsWindowEnabled.Call(h)
	bg := rgb(34, 43, 58)
	fg := rgb(233, 238, 247)
	if isPrimaryButton(id) {
		bg = rgb(105, 88, 238)
	}
	if id == ID_DELETE {
		bg = rgb(83, 39, 50)
	}
	if buttonCheck[h] {
		bg = rgb(82, 71, 194)
	}
	if enabledNow == 0 {
		bg = rgb(25, 31, 42)
		fg = rgb(105, 116, 134)
	}
	if buttonHover[h] && enabledNow != 0 {
		if isPrimaryButton(id) {
			bg = rgb(122, 105, 250)
		} else if id == ID_DELETE {
			bg = rgb(109, 48, 61)
		} else {
			bg = rgb(45, 56, 74)
		}
	}
	if buttonDown[h] && enabledNow != 0 {
		bg = rgb(73, 63, 171)
	}
	brush, _, _ := pCreateSolidBrush.Call(bg)
	oldBrush, _, _ := pSelectObject.Call(hdc, brush)
	nullPen, _, _ := pGetStockObject.Call(NULL_PEN)
	oldPen, _, _ := pSelectObject.Call(hdc, nullPen)
	pRoundRect.Call(hdc, 0, 0, uintptr(r.Right), uintptr(r.Bottom), 13, 13)
	pSelectObject.Call(hdc, oldPen)
	pSelectObject.Call(hdc, oldBrush)
	pDeleteObject.Call(brush)
	pSetBkMode.Call(hdc, TRANSPARENT)
	pSetTextColor.Call(hdc, fg)
	if guiFont != 0 {
		old, _, _ := pSelectObject.Call(hdc, guiFont)
		defer pSelectObject.Call(hdc, old)
	}
	caption := text(h)
	pDrawText.Call(hdc, uintptr(unsafe.Pointer(w(caption))), ^uintptr(0), uintptr(unsafe.Pointer(&r)), DT_CENTER|DT_VCENTER|DT_SINGLELINE|DT_END_ELLIPSIS)
}

func buttonWndProc(h uintptr, m uint32, wp, lp uintptr) uintptr {
	switch m {
	case WM_PAINT:
		paintButton(h)
		return 0
	case WM_ERASEBKGND:
		return 1
	case WM_MOUSEMOVE:
		if !buttonHover[h] {
			buttonHover[h] = true
			tme := TRACKMOUSEEVENT{CbSize: uint32(unsafe.Sizeof(TRACKMOUSEEVENT{})), DwFlags: TME_LEAVE, HwndTrack: h}
			pTrackMouseEvent.Call(uintptr(unsafe.Pointer(&tme)))
			pInvalidateRect.Call(h, 0, 0)
		}
		return 0
	case WM_MOUSELEAVE:
		buttonHover[h] = false
		buttonDown[h] = false
		pInvalidateRect.Call(h, 0, 0)
		return 0
	case WM_LBUTTONDOWN:
		buttonDown[h] = true
		pSetCapture.Call(h)
		pSetFocus.Call(h)
		pInvalidateRect.Call(h, 0, 0)
		return 0
	case WM_LBUTTONUP:
		wasDown := buttonDown[h]
		buttonDown[h] = false
		pReleaseCapture.Call()
		pInvalidateRect.Call(h, 0, 0)
		if wasDown {
			idv, _, _ := pGetDlgCtrlID.Call(h)
			id := int(idv)
			if isToggleButton(id) {
				buttonCheck[h] = !buttonCheck[h]
			}
			parent, _, _ := pGetParent.Call(h)
			pSendMessage.Call(parent, WM_COMMAND, uintptr(id)|(uintptr(BN_CLICKED)<<16), h)
		}
		return 0
	case WM_KEYUP:
		if wp == VK_SPACE || wp == VK_RETURN {
			idv, _, _ := pGetDlgCtrlID.Call(h)
			id := int(idv)
			if isToggleButton(id) {
				buttonCheck[h] = !buttonCheck[h]
			}
			parent, _, _ := pGetParent.Call(h)
			pSendMessage.Call(parent, WM_COMMAND, uintptr(id)|(uintptr(BN_CLICKED)<<16), h)
			pInvalidateRect.Call(h, 0, 0)
			return 0
		}
	case WM_ENABLE, WM_SETFOCUS, WM_KILLFOCUS:
		pInvalidateRect.Call(h, 0, 0)
	case BM_GETCHECK:
		if buttonCheck[h] {
			return BST_CHECKED
		}
		return 0
	case BM_SETCHECK:
		buttonCheck[h] = wp == BST_CHECKED
		pInvalidateRect.Call(h, 0, 0)
		return 0
	}
	r, _, _ := pDefWindowProc.Call(h, uintptr(m), wp, lp)
	return r
}

func drawCard(hdc uintptr, x, y, wv, hv int32, brush uintptr, radius uintptr) {
	if wv <= 0 || hv <= 0 {
		return
	}
	oldBrush, _, _ := pSelectObject.Call(hdc, brush)
	nullPen, _, _ := pGetStockObject.Call(NULL_PEN)
	oldPen, _, _ := pSelectObject.Call(hdc, nullPen)
	pRoundRect.Call(hdc, uintptr(x), uintptr(y), uintptr(x+wv), uintptr(y+hv), radius, radius)
	pSelectObject.Call(hdc, oldPen)
	pSelectObject.Call(hdc, oldBrush)
}

func paintBrandMark(hdc uintptr) {
	brush, _, _ := pCreateSolidBrush.Call(rgb(105, 88, 238))
	defer pDeleteObject.Call(brush)
	drawCard(hdc, 18, 15, 34, 34, brush, 12)
	r := RECT{Left: 18, Top: 15, Right: 52, Bottom: 49}
	pSetBkMode.Call(hdc, TRANSPARENT)
	pSetTextColor.Call(hdc, rgb(255, 255, 255))
	if headingFont != 0 {
		old, _, _ := pSelectObject.Call(hdc, headingFont)
		defer pSelectObject.Call(hdc, old)
	}
	pDrawText.Call(hdc, uintptr(unsafe.Pointer(w("Z"))), ^uintptr(0), uintptr(unsafe.Pointer(&r)), DT_CENTER|DT_VCENTER|DT_SINGLELINE)
}

func paintMain(h uintptr) {
	var ps PAINTSTRUCT
	hdc, _, _ := pBeginPaint.Call(h, uintptr(unsafe.Pointer(&ps)))
	if hdc == 0 {
		return
	}
	defer pEndPaint.Call(h, uintptr(unsafe.Pointer(&ps)))
	var r RECT
	pGetClientRect.Call(h, uintptr(unsafe.Pointer(&r)))
	pFillRect.Call(hdc, uintptr(unsafe.Pointer(&r)), bgBrush)
	paintBrandMark(hdc)
	cw, ch := r.Right, r.Bottom
	if cw < 980 {
		cw = 980
	}
	if ch < 680 {
		ch = 680
	}
	margin := int32(18)
	if master == "" {
		cardW := int32(650)
		if cardW > cw-2*margin {
			cardW = cw - 2*margin
		}
		x := (cw - cardW) / 2
		cardH := int32(362)
		if lockedPathExists() {
			cardH = 306
		}
		drawCard(hdc, x, 132, cardW, cardH, panelBrush, 22)
		return
	}
	top := int32(84)
	bottom := int32(70)
	leftW := cw * 31 / 100
	if leftW < 300 {
		leftW = 300
	}
	if leftW > 395 {
		leftW = 395
	}
	gap := int32(14)
	rightX := margin + leftW + gap
	drawCard(hdc, margin, top, leftW, ch-top-bottom, panelBrush, 20)
	drawCard(hdc, rightX, top, cw-rightX-margin, ch-top-bottom, panelBrush, 20)
}
