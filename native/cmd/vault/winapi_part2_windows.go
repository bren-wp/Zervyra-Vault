//go:build windows

package main

import (
	"time"
	"zervyra-vault-native/internal/core"
)

const (
	WM_CREATE          = 0x0001
	WM_PAINT           = 0x000F
	WM_ERASEBKGND      = 0x0014
	WM_LBUTTONDOWN     = 0x0201
	WM_LBUTTONUP       = 0x0202
	WM_MOUSEMOVE       = 0x0200
	WM_MOUSELEAVE      = 0x02A3
	WM_KEYDOWN         = 0x0100
	WM_KEYUP           = 0x0101
	WM_ENABLE          = 0x000A
	WM_SETFOCUS        = 0x0007
	WM_KILLFOCUS       = 0x0008
	WM_CTLCOLOREDIT    = 0x0133
	WM_CTLCOLORLISTBOX = 0x0134
	WM_CTLCOLORBTN     = 0x0135
	WM_CTLCOLORSTATIC  = 0x0138
	WM_DESTROY         = 0x0002
	WM_SIZE            = 0x0005
	WM_CLOSE           = 0x0010
	WM_QUERYENDSESSION = 0x0011
	WM_ENDSESSION      = 0x0016
	WM_POWERBROADCAST  = 0x0218
	WM_COMMAND         = 0x0111
	WM_TIMER           = 0x0113
	WM_SETFONT         = 0x0030
	WM_SETICON         = 0x0080

	WS_VISIBLE          = 0x10000000
	WS_CHILD            = 0x40000000
	WS_OVERLAPPEDWINDOW = 0x00CF0000
	WS_CLIPCHILDREN     = 0x02000000
	WS_TABSTOP          = 0x00010000
	WS_BORDER           = 0x00800000
	WS_VSCROLL          = 0x00200000
	WS_EX_CLIENTEDGE    = 0x00000200

	ES_AUTOHSCROLL     = 0x0080
	ES_PASSWORD        = 0x0020
	ES_MULTILINE       = 0x0004
	ES_AUTOVSCROLL     = 0x0040
	ES_WANTRETURN      = 0x1000
	ES_READONLY        = 0x0800
	EM_SETPASSWORDCHAR = 0x00CC
	EM_SETLIMITTEXT    = 0x00C5
	EM_SETCUEBANNER    = 0x1501

	LBS_NOTIFY           = 0x0001
	LBS_NOINTEGRALHEIGHT = 0x0100
	LB_ADDSTRING         = 0x0180
	LB_RESETCONTENT      = 0x0184
	LB_GETCURSEL         = 0x0188
	LB_SETCURSEL         = 0x0186
	LBN_SELCHANGE        = 1

	BS_AUTOCHECKBOX = 0x0003
	BM_GETCHECK     = 0x00F0
	BM_SETCHECK     = 0x00F1
	BST_CHECKED     = 1

	BN_CLICKED = 0
	EN_CHANGE  = 0x0300

	SW_HIDE        = 0
	SW_SHOW        = 5
	SIZE_MINIMIZED = 1
	PBT_APMSUSPEND = 0x0004

	CF_UNICODETEXT = 13
	GMEM_MOVEABLE  = 0x0002

	MB_ICONINFORMATION = 0x40
	MB_ICONWARNING     = 0x30
	MB_ICONERROR       = 0x10
	MB_YESNOCANCEL     = 0x3
	MB_YESNO           = 0x4
	IDYES              = 6
	IDNO               = 7
	IDCANCEL           = 2

	OFN_OVERWRITEPROMPT = 0x00000002
	OFN_PATHMUSTEXIST   = 0x00000800
	OFN_FILEMUSTEXIST   = 0x00001000
	OFN_EXPLORER        = 0x00080000

	DEFAULT_GUI_FONT = 17
	NULL_PEN         = 8
	TRANSPARENT      = 1
	DT_CENTER        = 0x00000001
	DT_VCENTER       = 0x00000004
	DT_SINGLELINE    = 0x00000020
	DT_END_ELLIPSIS  = 0x00008000
	TME_LEAVE        = 0x00000002
	VK_SPACE         = 0x20
	VK_RETURN        = 0x0D
	IMAGE_ICON       = 1
	LR_LOADFROMFILE  = 0x0010
	ICON_SMALL       = 0
	ICON_BIG         = 1

	ID_PATH       = 100
	ID_BROWSE     = 101
	ID_MASTER     = 102
	ID_SHOWMASTER = 103
	ID_OPEN       = 104
	ID_CREATE     = 105
	ID_CONFIRM    = 106
	ID_SEARCH     = 110
	ID_LIST       = 111
	ID_TRASHMODE  = 112
	ID_NEW        = 120
	ID_TITLE      = 121
	ID_USERNAME   = 122
	ID_PASSWORD   = 123
	ID_SHOWPASS   = 124
	ID_URL        = 125
	ID_TOTP       = 126
	ID_TAGS       = 127
	ID_NOTES      = 128
	ID_FAVORITE   = 129
	ID_SAVEENTRY  = 130
	ID_DELETE     = 131
	ID_RESTORE    = 132
	ID_COPYUSER   = 133
	ID_COPYPASS   = 134
	ID_GENERATE   = 135
	ID_COPYTOTP   = 136
	ID_OPENURL    = 137
	ID_EMAIL      = 138
	ID_COPYEMAIL  = 139
	ID_SECURITY   = 140
	ID_SAVEVAULT  = 141
	ID_LOCK       = 142
	ID_BACKUP     = 143
	ID_EXPORT     = 144
)

type WNDCLASSEX struct {
	CbSize                          uint32
	Style                           uint32
	WndProc                         uintptr
	ClsExtra, WndExtra              int32
	Instance, Icon, Cursor          uintptr
	Background, MenuName, ClassName uintptr
	IconSm                          uintptr
}

type MSG struct {
	Hwnd           uintptr
	Message        uint32
	WParam, LParam uintptr
	Time           uint32
	Pt             struct{ X, Y int32 }
}

type RECT struct{ Left, Top, Right, Bottom int32 }

type PAINTSTRUCT struct {
	Hdc       uintptr
	Erase     int32
	RcPaint   RECT
	Restore   int32
	IncUpdate int32
	Reserved  [32]byte
}

type TRACKMOUSEEVENT struct {
	CbSize      uint32
	DwFlags     uint32
	HwndTrack   uintptr
	DwHoverTime uint32
}

type OPENFILENAME struct {
	LStructSize       uint32
	HwndOwner         uintptr
	HInstance         uintptr
	LpstrFilter       *uint16
	LpstrCustomFilter *uint16
	NMaxCustFilter    uint32
	NFilterIndex      uint32
	LpstrFile         *uint16
	NMaxFile          uint32
	LpstrFileTitle    *uint16
	NMaxFileTitle     uint32
	LpstrInitialDir   *uint16
	LpstrTitle        *uint16
	Flags             uint32
	NFileOffset       uint16
	NFileExtension    uint16
	LpstrDefExt       *uint16
	LCustData         uintptr
	LpfnHook          uintptr
	LpTemplateName    *uint16
	PvReserved        uintptr
	DwReserved        uint32
	FlagsEx           uint32
}

var (
	hwnd            uintptr
	controls        = map[int]uintptr{}
	guiFont         uintptr
	ownGUIFont      bool
	headingFont     uintptr
	smallFont       uintptr
	bgBrush         uintptr
	panelBrush      uintptr
	fieldBrush      uintptr
	buttonClassName = "ZervyraVaultButtonV110"
	buttonHover     = map[uintptr]bool{}
	buttonDown      = map[uintptr]bool{}
	buttonCheck     = map[uintptr]bool{}

	vault            = core.NewVault()
	currentPath      string
	master           string
	vaultLock        *core.VaultLock
	visibleIndices   []int
	selectedID       string
	editingNew       bool
	showTrash        bool
	dirty            bool
	suppressChange   bool
	lastActivity     = time.Now()
	lastLockTouch    = time.Now()
	clipboardSeq     uintptr
	clipboardClearAt time.Time
	lastEditAt       time.Time
	lastSavedAt      time.Time
	autosaveError    string
	revisionCaptured bool
)
